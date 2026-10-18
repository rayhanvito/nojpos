package inventory

import (
	"context"
	"errors"
	"fmt"
	"github.com/rayhanvito/nojpos/backend/internal/user"
	"strings"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/webhook"
)

// ProductRepository adalah kontrak baca produk yang dipakai Service —
// sejak Unit B penyesuaian stok tidak lagi didelegasikan ke repo
// katalog: Service mengorkestrasi mesin stok paket ini sendiri (Engine +
// Transactor), repo katalog hanya dipakai untuk membaca produk/varian.
type ProductRepository interface {
	GetByIDForOutlet(ctx context.Context, storeID, outletID, id uint) (*catalog.Product, error)
	GetVariantAtOutlet(ctx context.Context, storeID, outletID, productID, variantID uint) (*catalog.ProductVariant, error)
}

// MovementRepository adalah kontrak penyimpanan pergerakan stok yang
// dipakai Service (tanda tangan sama dengan interface lama di
// service/interfaces.go, tipe kini milik slice ini).
type MovementRepository interface {
	ListForStore(ctx context.Context, storeID uint, f MovementFilter) (*MovementPage, error)
	ListForOutlet(ctx context.Context, storeID, outletID uint, f MovementFilter) (*MovementPage, error)
}

// Notifier adalah potongan NotificationService lama yang dipakai
// Service: pemberitahuan stok rendah.
type Notifier interface {
	CheckAndNotifyLowStock(ctx context.Context, storeID, itemID uint, productName, variantName string, currentStock, minStock int, unit, referenceType string) error
}

// Transactor menjalankan unit kerja dalam satu transaksi (semantik
// JOIN database.Transactor, yang mengimplementasikannya).
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// WebhookEmitter adalah potongan webhook.Dispatcher yang dipakai
// paket ini (Unit F): event stock.updated untuk sistem luar, dipakai
// Service (penyesuaian + transfer) dan OpnameService (posting opname).
// Pengiriman asinkron di sisi dispatcher.
type WebhookEmitter interface {
	EmitStockUpdated(ctx context.Context, storeID uint, data webhook.StockData)
}

// Service memegang penyesuaian stok dan riwayat pergerakan stok.
// Terpisah dari CatalogService sejak BE-B3; metode-metodenya dipindah persis
// dari CatalogService tanpa mengubah perilaku.
type Service struct {
	prods     ProductRepository
	movs      MovementRepository
	notif     Notifier
	stock     *Engine
	tx        Transactor
	transfers *TransferRepo
	webhooks  WebhookEmitter
}

func NewService(prods ProductRepository, movs MovementRepository, notif Notifier) *Service {
	return &Service{prods: prods, movs: movs, notif: notif}
}

// SetTransferRepo memasang repo transfer stok antar outlet (Unit D),
// pola setter yang sama dengan SetStockEngine.
func (s *Service) SetTransferRepo(r *TransferRepo) {
	s.transfers = r
}

// SetStockEngine memasang Engine (mesin stok paket ini) + Transactor
// yang dipakai AdjustStock (pola setter yang sama dengan
// catalog.Service.SetInventory).
func (s *Service) SetStockEngine(stock *Engine, tx Transactor) {
	s.stock = stock
	s.tx = tx
}

// SetWebhookEmitter memasang dispatcher webhook (Unit F). Tanpa
// emitter, penyesuaian stok dan transfer tidak memancarkan event.
func (s *Service) SetWebhookEmitter(e WebhookEmitter) {
	s.webhooks = e
}

func (s *Service) AdjustStock(ctx context.Context, storeID, outletID, productID uint, variantID *uint, actor, direction string, qty int64, reason string) (*catalog.Product, error) {
	reason = strings.TrimSpace(reason)
	if direction != "plus" && direction != "minus" {
		return nil, ErrBadDirection
	}
	if qty < 1 {
		return nil, fmt.Errorf("jumlah harus minimal 1")
	}
	if reason == "" {
		return nil, fmt.Errorf("alasan penyesuaian wajib diisi")
	}
	delta := int(qty)
	if direction == "minus" {
		delta = -int(qty)
	}
	if outletID == 0 {
		def, err := s.stock.DefaultOutletID(ctx, storeID)
		if err != nil {
			if errors.Is(err, ErrNotFound) {
				return nil, user.ErrStoreMismatch
			}
			return nil, err
		}
		outletID = def
	}
	appliedDelta := delta
	if variantID != nil {
		err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
			if _, err := s.prods.GetVariantAtOutlet(ctx, storeID, outletID, productID, *variantID); err != nil {
				return err
			}
			_, _, actualDelta, err := s.stock.ChangeStock(ctx, storeID, outletID, productID, variantID, delta, false)
			if err != nil {
				return err
			}
			appliedDelta = actualDelta
			return s.stock.InsertMovement(ctx, &Movement{
				StoreID: storeID, OutletID: &outletID, ProductID: productID, VariantID: variantID,
				Type: MovementAdjust, Qty: actualDelta, Reason: reason, Actor: actor,
			})
		})
		if err != nil {
			switch {
			case errors.Is(err, ErrNotFound):
				return nil, user.ErrStoreMismatch
			case errors.Is(err, catalog.ErrStockInsufficient):
				return nil, fmt.Errorf("stok tidak boleh negatif")
			}
			return nil, err
		}
		variant, err := s.prods.GetVariantAtOutlet(ctx, storeID, outletID, productID, *variantID)
		if err != nil {
			return nil, err
		}
		product, err := s.prods.GetByIDForOutlet(ctx, storeID, outletID, productID)
		if err != nil {
			return nil, err
		}
		if s.notif != nil {
			_ = s.notif.CheckAndNotifyLowStock(ctx, storeID, variant.ID, product.Name, variant.Name, variant.Stock, variant.MinStock, product.Unit, "variant")
		}
		if s.webhooks != nil {
			s.webhooks.EmitStockUpdated(ctx, storeID, webhook.StockData{
				Source: "adjustment", ProductID: productID, VariantID: variantID,
				QtyDelta: appliedDelta, OutletID: outletID, Reason: reason,
			})
		}
		return product, nil
	}

	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		p, err := s.prods.GetByIDForOutlet(ctx, storeID, outletID, productID)
		if err != nil {
			return err
		}
		if len(p.Variants) > 0 {
			return catalog.ErrVariantRequired
		}
		_, _, actualDelta, err := s.stock.ChangeStock(ctx, storeID, outletID, productID, nil, delta, false)
		if err != nil {
			return err
		}
		appliedDelta = actualDelta
		return s.stock.InsertMovement(ctx, &Movement{
			StoreID: storeID, OutletID: &outletID, ProductID: productID,
			Type: MovementAdjust, Qty: delta, Reason: reason, Actor: actor,
		})
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return nil, user.ErrStoreMismatch
		case errors.Is(err, catalog.ErrStockInsufficient):
			return nil, fmt.Errorf("stok tidak boleh negatif")
		}
		return nil, err
	}
	p, err := s.prods.GetByIDForOutlet(ctx, storeID, outletID, productID)
	if err != nil {
		return nil, err
	}
	if p != nil && s.notif != nil {
		_ = s.notif.CheckAndNotifyLowStock(ctx, storeID, p.ID, p.Name, "", p.Stock, p.MinStock, p.Unit, "product")
	}
	if s.webhooks != nil {
		s.webhooks.EmitStockUpdated(ctx, storeID, webhook.StockData{
			Source: "adjustment", ProductID: productID,
			QtyDelta: appliedDelta, OutletID: outletID, Reason: reason,
		})
	}
	return p, nil
}

func (s *Service) ListMovements(ctx context.Context, storeID, outletID uint, f MovementListFilter) (*MovementPage, error) {
	return s.movs.ListForOutlet(ctx, storeID, outletID, MovementFilter{
		Type:      f.Type,
		ProductID: f.ProductID,
		Q:         f.Q,
		Page:      f.Page,
		Limit:     f.Limit,
	})
}
