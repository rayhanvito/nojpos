package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/rayhanvito/nojpos/backend/model"
	"github.com/rayhanvito/nojpos/backend/repo"
)

var (
	ErrExpenseNotFound = errors.New("pengeluaran tidak ditemukan")
	ErrExpenseInvalid  = errors.New("data pengeluaran tidak valid")
)

type ExpenseInput struct {
	Date     string
	Category string
	Note     string
	Amount   int64
	Method   string
	Source   string
}

type ExpenseService struct {
	expenses ExpenseRepository
	cashiers CashierRepository
	shifts   ShiftRepository
}

func NewExpenseService(expenses ExpenseRepository, cashiers CashierRepository, shifts ShiftRepository) *ExpenseService {
	return &ExpenseService{expenses: expenses, cashiers: cashiers, shifts: shifts}
}

func (s *ExpenseService) List(ctx context.Context, storeID uint, f repo.ExpenseFilter) (*repo.ExpensePage, error) {
	return s.expenses.List(ctx, storeID, f)
}

func (s *ExpenseService) Create(ctx context.Context, storeID, outletID uint, actingAsCashierID *uint, fallbackName string, in ExpenseInput) (*model.Expense, error) {
	expense, err := normalizeExpense(in)
	if err != nil {
		return nil, err
	}

	// Identitas pencatat selalu dari JWT, mengikuti pola Checkout: admin
	// tanpa switch tercatat sebagai dirinya (id 0), kasir memakai record-nya.
	var cashierID uint
	cashierName := fallbackName
	if actingAsCashierID != nil {
		cashierID = *actingAsCashierID
		if c, err := s.cashiers.GetByID(ctx, cashierID); err == nil {
			cashierName = c.Name
		}
	}

	expense.StoreID = storeID
	outlet := outletID
	expense.OutletID = &outlet
	expense.CreatedByID = cashierID
	expense.CreatedByName = cashierName
	s.linkShift(ctx, storeID, outletID, expense)
	return s.expenses.Create(ctx, expense)
}

func (s *ExpenseService) Update(ctx context.Context, storeID, id uint, in ExpenseInput) (*model.Expense, error) {
	existing, err := s.expenses.GetByID(ctx, storeID, id)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrExpenseNotFound
	}
	if err != nil {
		return nil, err
	}
	normalized, err := normalizeExpense(in)
	if err != nil {
		return nil, err
	}
	existing.Date = normalized.Date
	existing.Category = normalized.Category
	existing.Note = normalized.Note
	existing.Amount = normalized.Amount
	existing.Method = normalized.Method
	existing.Source = normalized.Source
	if existing.Source != model.ExpenseSourceDrawer {
		existing.ShiftID = nil
	} else if existing.ShiftID == nil {
		// Tautan shift hanya dibuat saat pencatatan (atau saat koreksi
		// sumber menjadi laci) — tidak pernah menautkan ulang yang sudah ada.
		outletID := uint(0)
		if existing.OutletID != nil {
			outletID = *existing.OutletID
		}
		probe := &model.Expense{CreatedByID: existing.CreatedByID, Source: existing.Source, Method: existing.Method}
		s.linkShift(ctx, storeID, outletID, probe)
		existing.ShiftID = probe.ShiftID
	}
	updated, err := s.expenses.Update(ctx, storeID, id, existing)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrExpenseNotFound
	}
	return updated, err
}

func (s *ExpenseService) Delete(ctx context.Context, storeID, id uint) error {
	err := s.expenses.Delete(ctx, storeID, id)
	if errors.Is(err, repo.ErrNotFound) {
		return ErrExpenseNotFound
	}
	return err
}

// linkShift menautkan pengeluaran laci ke shift terbuka pencatatnya, bila
// ada. Tanpa shift terbuka pengeluaran tetap tercatat, hanya tak tertaut.
func (s *ExpenseService) linkShift(ctx context.Context, storeID, outletID uint, expense *model.Expense) {
	if expense.Source != model.ExpenseSourceDrawer || expense.Method != string(model.PayCash) {
		return
	}
	shift, err := s.shifts.GetActiveShift(ctx, storeID, outletID, expense.CreatedByID)
	if err == nil && shift != nil {
		expense.ShiftID = &shift.ID
	}
}

func normalizeExpense(in ExpenseInput) (*model.Expense, error) {
	category := strings.TrimSpace(in.Category)
	if category == "" || in.Amount <= 0 {
		return nil, ErrExpenseInvalid
	}
	date, err := time.Parse("2006-01-02", strings.TrimSpace(in.Date))
	if err != nil {
		return nil, ErrExpenseInvalid
	}
	if !model.ValidStandardPayMethod(in.Method) {
		return nil, ErrExpenseInvalid
	}
	source := in.Source
	if source == "" {
		source = model.ExpenseSourceOutside
	}
	if source != model.ExpenseSourceDrawer && source != model.ExpenseSourceOutside {
		return nil, ErrExpenseInvalid
	}
	if in.Method != string(model.PayCash) {
		// Hanya tunai yang bisa keluar dari laci.
		source = model.ExpenseSourceOutside
	}
	return &model.Expense{
		Date:     date,
		Category: category,
		Note:     strings.TrimSpace(in.Note),
		Amount:   in.Amount,
		Method:   in.Method,
		Source:   source,
	}, nil
}
