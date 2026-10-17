package catalog

import (
	"context"
)

// StockEngine adalah mesin stok kanonis milik slice inventory (Unit B),
// dinyatakan di sini sebagai interface primitif supaya katalog tidak
// mengimpor inventory (inventory mengimpor katalog — DAG satu arah).
// Semua metode bekerja pada DB(ctx): di dalam Transactor.WithinTx
// mereka bergabung ke transaksi pemanggil.
type StockEngine interface {
	GetStock(ctx context.Context, outletID, productID uint, variantID *uint) (int, error)
	SetStock(ctx context.Context, storeID, outletID, productID uint, variantID *uint, qty int) error
	RecordMovement(ctx context.Context, storeID, outletID, productID uint, variantID *uint, movementType string, qty int, reason, actor string) error
	IsDefaultOutlet(ctx context.Context, storeID, outletID uint) (bool, error)
}

// Transactor menjalankan unit kerja lintas-repo dalam satu transaksi
// (semantik JOIN database.Transactor, yang mengimplementasikannya).
type Transactor interface {
	WithinTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// SetStockEngine memasang mesin stok inventory + Transactor yang dipakai
// orkestrasi stok Service (pola setter yang sama dengan SetInventory).
func (s *Service) SetStockEngine(stock StockEngine, tx Transactor) {
	s.stock = stock
	s.tx = tx
}

// createProductWithStock menggantikan ProductRepo.createWithInitialMovement
// lama: baris produk disisipkan repo katalog, stok awal + movement
// "initial" ditulis SATU-SATUNYA mesin stok (inventory) — atomik dalam
// satu transaksi Transactor. Semantik lama dipertahankan persis: outlet
// 0 = outlet default; kolom stock produk hanya terisi bila outlet tujuan
// adalah outlet default (di luar itu stok hidup di product_stocks).
func (s *Service) createProductWithStock(ctx context.Context, storeID, outletID uint, p *Product, actor string) (*Product, error) {
	if outletID == 0 {
		def, err := s.prods.DefaultOutlet(ctx, storeID)
		if err != nil {
			return nil, err
		}
		outletID = def.ID
	}
	initialStock := p.Stock
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		isDefault, err := s.stock.IsDefaultOutlet(ctx, storeID, outletID)
		if err != nil {
			return err
		}
		if !isDefault {
			p.Stock = 0
		}
		if err := s.prods.InsertProduct(ctx, storeID, outletID, p); err != nil {
			return err
		}
		if err := s.stock.SetStock(ctx, storeID, outletID, p.ID, nil, initialStock); err != nil {
			return err
		}
		if initialStock > 0 {
			return s.stock.RecordMovement(ctx, storeID, outletID, p.ID, nil, "initial", initialStock, "Stok awal", actor)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.prods.GetByIDForOutlet(ctx, storeID, outletID, p.ID)
}

// createVariantWithStock menggantikan ProductRepo.CreateVariantAtOutlet
// lama dengan orkestrasi yang sama seperti createProductWithStock:
// keberadaan produk diverifikasi dulu (ErrNotFound), lalu baris varian +
// stok awal + movement "initial" ditulis atomik dalam satu transaksi.
func (s *Service) createVariantWithStock(ctx context.Context, storeID, outletID, productID uint, variant *ProductVariant, actor string) (*ProductVariant, error) {
	if outletID == 0 {
		def, err := s.prods.DefaultOutlet(ctx, storeID)
		if err != nil {
			return nil, err
		}
		outletID = def.ID
	}
	if _, err := s.prods.GetByIDForOutlet(ctx, storeID, outletID, productID); err != nil {
		return nil, err
	}
	initialStock := variant.Stock
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		isDefault, err := s.stock.IsDefaultOutlet(ctx, storeID, outletID)
		if err != nil {
			return err
		}
		if !isDefault {
			variant.Stock = 0
		}
		if err := s.prods.InsertVariant(ctx, storeID, outletID, productID, variant); err != nil {
			return err
		}
		variantID := variant.ID
		if err := s.stock.SetStock(ctx, storeID, outletID, productID, &variantID, initialStock); err != nil {
			return err
		}
		if initialStock > 0 {
			return s.stock.RecordMovement(ctx, storeID, outletID, productID, &variantID, "initial", initialStock, "Stok awal varian", actor)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.prods.GetVariantAtOutlet(ctx, storeID, outletID, productID, variant.ID)
}

// updateVariantWithStock menggantikan ProductRepo.UpdateVariantAtOutlet
// lama: kolom non-stok diperbarui repo katalog; selisih stok varian
// terhadap stok outlet saat ini diterapkan mesin stok inventory dan
// dicatat sebagai movement "adjust" — atomik dalam satu transaksi.
func (s *Service) updateVariantWithStock(ctx context.Context, storeID, outletID, productID, variantID uint, variant *ProductVariant, actor string) (*ProductVariant, error) {
	if outletID == 0 {
		def, err := s.prods.DefaultOutlet(ctx, storeID)
		if err != nil {
			return nil, err
		}
		outletID = def.ID
	}
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		currentStock, err := s.stock.GetStock(ctx, outletID, productID, &variantID)
		if err != nil {
			return err
		}
		if err := s.prods.UpdateVariantFields(ctx, storeID, productID, variantID, variant); err != nil {
			return err
		}
		if err := s.stock.SetStock(ctx, storeID, outletID, productID, &variantID, variant.Stock); err != nil {
			return err
		}
		if delta := variant.Stock - currentStock; delta != 0 {
			return s.stock.RecordMovement(ctx, storeID, outletID, productID, &variantID, "adjust", delta, "Perubahan stok varian", actor)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return s.prods.GetVariantAtOutlet(ctx, storeID, outletID, productID, variantID)
}
