package report

import (
	"context"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/sales"
)

// TransactionExporter adalah kontrak transaksi yang dipakai
// ExportService — hanya ListForExport dari TrxRepository lama yang
// benar-benar dipakai (tipe sales.Trx, BE-F8).
type TransactionExporter interface {
	ListForExport(ctx context.Context, storeID uint, period, timezone, date, from, to string) ([]*sales.Trx, error)
}

// Products adalah kontrak katalog yang dipakai ExportService - hanya
// ListForExportAtOutlet dari ProductRepository lama yang benar-benar
// dipakai (tipe catalog.Product).
type Products interface {
	ListForExportAtOutlet(ctx context.Context, storeID, outletID uint) ([]*catalog.Product, error)
}

// ExportService adalah ExportService lama di service/export.go
// (BE-F10), dipindah verbatim.
type ExportService struct {
	stores   Stores
	trxs     TransactionExporter
	products Products
}

func NewExportService(stores Stores, trxs TransactionExporter, products Products) *ExportService {
	return &ExportService{stores: stores, trxs: trxs, products: products}
}

func (s *ExportService) Transactions(ctx context.Context, storeID uint, period, date, from, to string) ([]*sales.Trx, string, error) {
	if !validPeriods[period] {
		return nil, "", ErrBadPeriod
	}
	timezone := s.timezone(ctx, storeID)
	transactions, err := s.trxs.ListForExport(ctx, storeID, period, timezone, date, from, to)
	if err != nil {
		return nil, "", err
	}
	return transactions, timezone, nil
}

func (s *ExportService) Products(ctx context.Context, storeID, outletID uint) ([]*catalog.Product, string, error) {
	products, err := s.products.ListForExportAtOutlet(ctx, storeID, outletID)
	if err != nil {
		return nil, "", err
	}
	return products, s.timezone(ctx, storeID), nil
}

func (s *ExportService) timezone(ctx context.Context, storeID uint) string {
	timezone, err := s.stores.GetTimezone(ctx, storeID)
	if err != nil || timezone == "" {
		return "Asia/Makassar"
	}
	return timezone
}
