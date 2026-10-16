package service

import (
	"context"
	"errors"
	"time"

	"github.com/adrr-dev/openPOS/backend/model"
	"github.com/adrr-dev/openPOS/backend/repo"
)

var (
	ErrDebtStatusInvalid = errors.New("status hutang-piutang tidak valid")
	ErrDebtPaymentInput  = errors.New("pembayaran hutang-piutang tidak valid")
)

type DebtService struct {
	debts  *repo.DebtRepo
	stores StoreRepository
}

func NewDebtService(debts *repo.DebtRepo, stores StoreRepository) *DebtService {
	return &DebtService{debts: debts, stores: stores}
}

func (s *DebtService) ListReceivables(ctx context.Context, storeID uint, status string, overdue bool) ([]*model.Receivable, error) {
	if !validDebtStatus(status) {
		return nil, ErrDebtStatusInvalid
	}
	var overdueBefore *time.Time
	if overdue {
		timezone, err := s.stores.GetTimezone(ctx, storeID)
		if err != nil {
			timezone = "Asia/Makassar"
		}
		location, err := time.LoadLocation(timezone)
		if err != nil {
			location = time.UTC
		}
		now := time.Now().In(location)
		startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
		overdueBefore = &startOfToday
	}
	return s.debts.ListReceivables(ctx, storeID, status, overdueBefore)
}

func (s *DebtService) GetReceivable(ctx context.Context, storeID, id uint) (*model.Receivable, error) {
	return s.debts.GetReceivable(ctx, storeID, id)
}

func (s *DebtService) PayReceivable(ctx context.Context, storeID, id uint, amount int64, method, actor string) (*model.Receivable, error) {
	if amount <= 0 || !model.ValidStandardPayMethod(method) {
		return nil, ErrDebtPaymentInput
	}
	return s.debts.PayReceivable(ctx, storeID, id, amount, method, actor)
}

func (s *DebtService) ListPayables(ctx context.Context, storeID uint, status string) ([]*model.Payable, error) {
	if !validDebtStatus(status) {
		return nil, ErrDebtStatusInvalid
	}
	return s.debts.ListPayables(ctx, storeID, status)
}

func (s *DebtService) GetPayable(ctx context.Context, storeID, id uint) (*model.Payable, error) {
	return s.debts.GetPayable(ctx, storeID, id)
}

func (s *DebtService) PayPayable(ctx context.Context, storeID, id uint, amount int64, method, actor string) (*model.Payable, error) {
	if amount <= 0 || !model.ValidStandardPayMethod(method) {
		return nil, ErrDebtPaymentInput
	}
	return s.debts.PayPayable(ctx, storeID, id, amount, method, actor)
}

func validDebtStatus(status string) bool {
	switch status {
	case "", string(model.DebtUnpaid), string(model.DebtPartiallyPaid), string(model.DebtPaid):
		return true
	default:
		return false
	}
}
