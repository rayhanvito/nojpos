package service

import (
	"context"

	"github.com/rayhanvito/nojpos/backend/model"
)

type ExportService struct {
	stores   StoreRepository
	trxs     TrxRepository
	products ProductRepository
}

func NewExportService(stores StoreRepository, trxs TrxRepository, products ProductRepository) *ExportService {
	return &ExportService{stores: stores, trxs: trxs, products: products}
}

func (s *ExportService) Transactions(ctx context.Context, storeID uint, period, date, from, to string) ([]*model.Trx, string, error) {
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

func (s *ExportService) Products(ctx context.Context, storeID, outletID uint) ([]*model.Product, string, error) {
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
