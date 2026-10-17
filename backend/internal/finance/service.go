package finance

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/rayhanvito/nojpos/backend/internal/shift"
	"github.com/rayhanvito/nojpos/backend/internal/staff"
)

// Stores adalah kontrak toko yang dipakai Service utang-piutang —
// hanya GetTimezone dari StoreRepository lama yang benar-benar dipakai
// (diimplementasikan outlet.StoreRepo).
type Stores interface {
	GetTimezone(ctx context.Context, storeID uint) (string, error)
}

// Service adalah DebtService lama di service/debt.go, dipindah
// verbatim; konstruktornya kini menerima *DebtRepo slice (tipe konkret,
// sama seperti sebelumnya).
type Service struct {
	debts  *DebtRepo
	stores Stores
}

func NewService(debts *DebtRepo, stores Stores) *Service {
	return &Service{debts: debts, stores: stores}
}

func (s *Service) ListReceivables(ctx context.Context, storeID uint, status string, overdue bool) ([]*Receivable, error) {
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

func (s *Service) GetReceivable(ctx context.Context, storeID, id uint) (*Receivable, error) {
	return s.debts.GetReceivable(ctx, storeID, id)
}

func (s *Service) PayReceivable(ctx context.Context, storeID, id uint, amount int64, method, actor string) (*Receivable, error) {
	if amount <= 0 || !validStandardPayMethod(method) {
		return nil, ErrDebtPaymentInput
	}
	return s.debts.PayReceivable(ctx, storeID, id, amount, method, actor)
}

func (s *Service) ListPayables(ctx context.Context, storeID uint, status string) ([]*Payable, error) {
	if !validDebtStatus(status) {
		return nil, ErrDebtStatusInvalid
	}
	return s.debts.ListPayables(ctx, storeID, status)
}

func (s *Service) GetPayable(ctx context.Context, storeID, id uint) (*Payable, error) {
	return s.debts.GetPayable(ctx, storeID, id)
}

func (s *Service) PayPayable(ctx context.Context, storeID, id uint, amount int64, method, actor string) (*Payable, error) {
	if amount <= 0 || !validStandardPayMethod(method) {
		return nil, ErrDebtPaymentInput
	}
	return s.debts.PayPayable(ctx, storeID, id, amount, method, actor)
}

func validDebtStatus(status string) bool {
	switch status {
	case "", string(DebtUnpaid), string(DebtPartiallyPaid), string(DebtPaid):
		return true
	default:
		return false
	}
}

// Expenses adalah kontrak penyimpanan pengeluaran yang dipakai
// ExpenseService (ExpenseRepository lama di service/interfaces.go,
// bertipe slice).
type Expenses interface {
	List(ctx context.Context, storeID uint, f ExpenseFilter) (*ExpensePage, error)
	GetByID(ctx context.Context, storeID, id uint) (*Expense, error)
	Create(ctx context.Context, expense *Expense) (*Expense, error)
	Update(ctx context.Context, storeID, id uint, expense *Expense) (*Expense, error)
	Delete(ctx context.Context, storeID, id uint) error
}

// Cashiers adalah kontrak kasir yang dipakai ExpenseService — hanya
// GetByIDUnscoped dari CashierRepository lama yang benar-benar dipakai
// (tipe staff.Cashier).
type Cashiers interface {
	GetByIDUnscoped(ctx context.Context, id uint) (*staff.Cashier, error)
}

// Shifts adalah kontrak shift yang dipakai ExpenseService — hanya
// GetActiveShift dari ShiftRepository lama yang benar-benar dipakai
// (tipe shift.CashierShift).
type Shifts interface {
	GetActiveShift(ctx context.Context, storeID, outletID, cashierID uint) (*shift.CashierShift, error)
}

// ExpenseService adalah ExpenseService lama di service/expense.go,
// dipindah verbatim.
type ExpenseService struct {
	expenses Expenses
	cashiers Cashiers
	shifts   Shifts
}

func NewExpenseService(expenses Expenses, cashiers Cashiers, shifts Shifts) *ExpenseService {
	return &ExpenseService{expenses: expenses, cashiers: cashiers, shifts: shifts}
}

func (s *ExpenseService) List(ctx context.Context, storeID uint, f ExpenseFilter) (*ExpensePage, error) {
	return s.expenses.List(ctx, storeID, f)
}

func (s *ExpenseService) Create(ctx context.Context, storeID, outletID uint, actingAsCashierID *uint, fallbackName string, in ExpenseInput) (*Expense, error) {
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
		if c, err := s.cashiers.GetByIDUnscoped(ctx, cashierID); err == nil {
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

func (s *ExpenseService) Update(ctx context.Context, storeID, id uint, in ExpenseInput) (*Expense, error) {
	existing, err := s.expenses.GetByID(ctx, storeID, id)
	if errors.Is(err, ErrNotFound) {
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
	if existing.Source != ExpenseSourceDrawer {
		existing.ShiftID = nil
	} else if existing.ShiftID == nil {
		// Tautan shift hanya dibuat saat pencatatan (atau saat koreksi
		// sumber menjadi laci) — tidak pernah menautkan ulang yang sudah ada.
		outletID := uint(0)
		if existing.OutletID != nil {
			outletID = *existing.OutletID
		}
		probe := &Expense{CreatedByID: existing.CreatedByID, Source: existing.Source, Method: existing.Method}
		s.linkShift(ctx, storeID, outletID, probe)
		existing.ShiftID = probe.ShiftID
	}
	updated, err := s.expenses.Update(ctx, storeID, id, existing)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrExpenseNotFound
	}
	return updated, err
}

func (s *ExpenseService) Delete(ctx context.Context, storeID, id uint) error {
	err := s.expenses.Delete(ctx, storeID, id)
	if errors.Is(err, ErrNotFound) {
		return ErrExpenseNotFound
	}
	return err
}

// linkShift menautkan pengeluaran laci ke shift terbuka pencatatnya, bila
// ada. Tanpa shift terbuka pengeluaran tetap tercatat, hanya tak tertaut.
func (s *ExpenseService) linkShift(ctx context.Context, storeID, outletID uint, expense *Expense) {
	if expense.Source != ExpenseSourceDrawer || expense.Method != payCash {
		return
	}
	shift, err := s.shifts.GetActiveShift(ctx, storeID, outletID, expense.CreatedByID)
	if err == nil && shift != nil {
		expense.ShiftID = &shift.ID
	}
}

func normalizeExpense(in ExpenseInput) (*Expense, error) {
	category := strings.TrimSpace(in.Category)
	if category == "" || in.Amount <= 0 {
		return nil, ErrExpenseInvalid
	}
	date, err := time.Parse("2006-01-02", strings.TrimSpace(in.Date))
	if err != nil {
		return nil, ErrExpenseInvalid
	}
	if !validStandardPayMethod(in.Method) {
		return nil, ErrExpenseInvalid
	}
	source := in.Source
	if source == "" {
		source = ExpenseSourceOutside
	}
	if source != ExpenseSourceDrawer && source != ExpenseSourceOutside {
		return nil, ErrExpenseInvalid
	}
	if in.Method != payCash {
		// Hanya tunai yang bisa keluar dari laci.
		source = ExpenseSourceOutside
	}
	return &Expense{
		Date:     date,
		Category: category,
		Note:     strings.TrimSpace(in.Note),
		Amount:   in.Amount,
		Method:   in.Method,
		Source:   source,
	}, nil
}
