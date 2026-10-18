package inventory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rayhanvito/nojpos/backend/internal/webhook"
)

// StockOpnameRepository adalah kontrak penyimpanan stock opname yang
// dipakai OpnameService (tanda tangan sama dengan interface lama di
// service/interfaces.go, tipe kini milik slice ini).
type StockOpnameRepository interface {
	List(ctx context.Context, storeID, outletID uint) ([]StockOpnameSession, error)
	GetByID(ctx context.Context, storeID, outletID, id uint) (*StockOpnameSession, error)
	Create(ctx context.Context, storeID, outletID uint, note, createdBy string) (*StockOpnameSession, error)
	UpdateItems(ctx context.Context, storeID, outletID, id uint, items []StockOpnameItemCount) (*StockOpnameSession, error)
	Post(ctx context.Context, storeID, outletID, id uint, actor string) (*StockOpnameSession, error)
	Cancel(ctx context.Context, storeID, outletID, id uint) (*StockOpnameSession, error)
}

type OpnameService struct {
	opnames  StockOpnameRepository
	webhooks WebhookEmitter
}

func NewOpnameService(opnames StockOpnameRepository) *OpnameService {
	return &OpnameService{opnames: opnames}
}

// SetWebhookEmitter memasang dispatcher webhook (Unit F). Tanpa
// emitter, posting opname tidak memancarkan event stock.updated.
func (s *OpnameService) SetWebhookEmitter(e WebhookEmitter) {
	s.webhooks = e
}

func (s *OpnameService) List(ctx context.Context, storeID, outletID uint) ([]StockOpnameSession, error) {
	return s.opnames.List(ctx, storeID, outletID)
}

func (s *OpnameService) Get(ctx context.Context, storeID, outletID, id uint) (*StockOpnameSession, error) {
	session, err := s.opnames.GetByID(ctx, storeID, outletID, id)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrStockOpnameNotFound
	}
	return session, err
}

func (s *OpnameService) Create(ctx context.Context, storeID, outletID uint, note, createdBy string) (*StockOpnameSession, error) {
	session, err := s.opnames.Create(ctx, storeID, outletID, strings.TrimSpace(note), strings.TrimSpace(createdBy))
	return session, s.mapOpnameError(err)
}

func (s *OpnameService) UpdateItems(ctx context.Context, storeID, outletID, id uint, items []StockOpnameItemCount) (*StockOpnameSession, error) {
	if len(items) == 0 {
		return nil, invalidStockOpname("Minimal satu item hitung wajib diisi.")
	}
	seen := make(map[uint]bool, len(items))
	for _, item := range items {
		if item.ItemID == 0 || item.CountedStock < 0 || seen[item.ItemID] {
			return nil, invalidStockOpname("Item hitung harus unik dan countedStock tidak boleh negatif.")
		}
		seen[item.ItemID] = true
	}
	session, err := s.opnames.UpdateItems(ctx, storeID, outletID, id, items)
	return session, s.mapOpnameError(err)
}

func (s *OpnameService) Post(ctx context.Context, storeID, outletID, id uint, actor string) (*StockOpnameSession, error) {
	session, err := s.opnames.Post(ctx, storeID, outletID, id, actor)
	if err != nil {
		return session, s.mapOpnameError(err)
	}
	// Webhook stock.updated per item yang stoknya benar-benar berubah
	// (selisih hitung != 0), sesudah posting berhasil.
	if s.webhooks != nil && session != nil {
		for _, item := range session.Items {
			if item.CountedStock == nil {
				continue
			}
			delta := *item.CountedStock - item.SystemStock
			if item.Difference != nil {
				delta = *item.Difference
			}
			if delta == 0 {
				continue
			}
			s.webhooks.EmitStockUpdated(ctx, storeID, webhook.StockData{
				Source: "opname", ProductID: item.ProductID, VariantID: item.VariantID,
				QtyDelta: delta, OutletID: outletID, OpnameSessionID: session.ID,
			})
		}
	}
	return session, nil
}

func (s *OpnameService) Cancel(ctx context.Context, storeID, outletID, id uint) (*StockOpnameSession, error) {
	session, err := s.opnames.Cancel(ctx, storeID, outletID, id)
	return session, s.mapOpnameError(err)
}

func (s *OpnameService) mapOpnameError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return ErrStockOpnameNotFound
	case errors.Is(err, ErrStockOpnameActive):
		return ErrStockOpnameActive
	case errors.Is(err, ErrStockOpnameState):
		return ErrStockOpnameState
	case errors.Is(err, ErrStockOpnameItem):
		return invalidStockOpname("Item tidak ditemukan dalam sesi stock opname ini.")
	case errors.Is(err, ErrStockOpnameTarget):
		return invalidStockOpname("Produk atau varian tujuan stok tidak tersedia.")
	default:
		return err
	}
}

func invalidStockOpname(message string) error {
	return fmt.Errorf("%w: %s", ErrStockOpnameInvalid, message)
}
