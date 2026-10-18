package inventory

import (
	"context"
	"fmt"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/webhook"
)

// TransferItemCmd adalah satu baris barang pada perintah transfer stok.
type TransferItemCmd struct {
	ProductID uint
	VariantID *uint
	Qty       int
}

// CreateTransferCmd adalah perintah pembuatan transfer stok antar outlet.
type CreateTransferCmd struct {
	FromOutletID uint
	ToOutletID   uint
	Note         string
	Items        []TransferItemCmd
}

// CreateTransfer memindahkan stok antar dua outlet milik toko yang sama
// secara atomik penuh: header + items tersimpan, stok asal dikurangi dan
// stok tujuan ditambah lewat Engine (pemilik tunggal mutasi stok), dan
// pergerakan tercatat di kedua sisi (transfer_out / transfer_in) dengan
// alasan merujuk nomor transfer. Kegagalan di titik mana pun — termasuk
// stok asal tidak cukup — membatalkan seluruhnya (rollback Transactor).
func (s *Service) CreateTransfer(ctx context.Context, storeID, createdByID uint, createdByName string, cmd CreateTransferCmd) (*StockTransfer, error) {
	if len(cmd.Items) == 0 {
		return nil, ErrTransferItemsEmpty
	}
	if cmd.FromOutletID == cmd.ToOutletID {
		return nil, ErrTransferSameOutlet
	}
	for _, item := range cmd.Items {
		if item.ProductID == 0 {
			return nil, ErrTransferInvalidItem
		}
		if item.Qty <= 0 {
			return nil, ErrTransferInvalidQty
		}
	}
	if s.transfers == nil || s.stock == nil || s.tx == nil {
		return nil, fmt.Errorf("layanan transfer stok belum terpasang")
	}

	var created *StockTransfer
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		for _, outletID := range []uint{cmd.FromOutletID, cmd.ToOutletID} {
			belongs, err := s.transfers.OutletBelongsToStore(ctx, storeID, outletID)
			if err != nil {
				return err
			}
			if !belongs {
				return ErrNotFound
			}
		}

		transfer := &StockTransfer{
			StoreID:       storeID,
			FromOutletID:  cmd.FromOutletID,
			ToOutletID:    cmd.ToOutletID,
			Note:          cmd.Note,
			CreatedBy:     createdByID,
			CreatedByName: createdByName,
		}
		if err := s.transfers.CreateHeader(ctx, transfer); err != nil {
			return err
		}

		reason := fmt.Sprintf("Transfer #%d", transfer.ID)
		if cmd.Note != "" {
			reason += ": " + cmd.Note
		}

		for _, item := range cmd.Items {
			if err := s.validateTransferItem(ctx, storeID, cmd.FromOutletID, item); err != nil {
				return err
			}
			if _, _, _, err := s.stock.ChangeStock(ctx, storeID, cmd.FromOutletID, item.ProductID, item.VariantID, -item.Qty, false); err != nil {
				return err
			}
			if _, _, _, err := s.stock.ChangeStock(ctx, storeID, cmd.ToOutletID, item.ProductID, item.VariantID, item.Qty, false); err != nil {
				return err
			}
			if err := s.stock.InsertMovement(ctx, &Movement{
				StoreID: storeID, OutletID: &cmd.FromOutletID, ProductID: item.ProductID, VariantID: item.VariantID,
				Type: MovementTransferOut, Qty: -item.Qty, Reason: reason, Actor: createdByName,
			}); err != nil {
				return err
			}
			if err := s.stock.InsertMovement(ctx, &Movement{
				StoreID: storeID, OutletID: &cmd.ToOutletID, ProductID: item.ProductID, VariantID: item.VariantID,
				Type: MovementTransferIn, Qty: item.Qty, Reason: reason, Actor: createdByName,
			}); err != nil {
				return err
			}
			row := StockTransferItem{
				TransferID: transfer.ID,
				ProductID:  item.ProductID,
				VariantID:  item.VariantID,
				Qty:        item.Qty,
			}
			if err := s.transfers.CreateItem(ctx, &row); err != nil {
				return err
			}
			transfer.Items = append(transfer.Items, row)
		}

		created = transfer
		return nil
	})
	if err != nil {
		return nil, err
	}
	// Webhook stock.updated per item untuk kedua sisi transfer:
	// pengurangan di outlet asal dan penambahan di outlet tujuan.
	if s.webhooks != nil {
		for _, item := range created.Items {
			base := webhook.StockData{
				Source: "transfer", ProductID: item.ProductID, VariantID: item.VariantID,
				TransferID: created.ID, FromOutletID: created.FromOutletID, ToOutletID: created.ToOutletID,
			}
			out := base
			out.OutletID = created.FromOutletID
			out.QtyDelta = -item.Qty
			s.webhooks.EmitStockUpdated(ctx, storeID, out)
			in := base
			in.OutletID = created.ToOutletID
			in.QtyDelta = item.Qty
			s.webhooks.EmitStockUpdated(ctx, storeID, in)
		}
	}
	return created, nil
}

// validateTransferItem memastikan produk (dan variannya, bila disebut)
// ada di toko ini, mengikuti pemeriksaan penyesuaian stok: produk
// bervarian wajib ditransfer per varian.
func (s *Service) validateTransferItem(ctx context.Context, storeID, outletID uint, item TransferItemCmd) error {
	if item.VariantID != nil {
		_, err := s.prods.GetVariantAtOutlet(ctx, storeID, outletID, item.ProductID, *item.VariantID)
		return err
	}
	p, err := s.prods.GetByIDForOutlet(ctx, storeID, outletID, item.ProductID)
	if err != nil {
		return err
	}
	if len(p.Variants) > 0 {
		return catalog.ErrVariantRequired
	}
	return nil
}

// ListTransfers mengembalikan daftar transfer toko (terbaru dulu)
// sesuai filter; paginasi diterapkan handler lewat httpx.PaginateList.
func (s *Service) ListTransfers(ctx context.Context, storeID uint, f TransferListFilter) ([]StockTransfer, error) {
	if s.transfers == nil {
		return nil, fmt.Errorf("layanan transfer stok belum terpasang")
	}
	items, err := s.transfers.List(ctx, storeID, f)
	if err != nil {
		return nil, err
	}
	return items, nil
}
