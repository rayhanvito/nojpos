package finance

import (
	"context"
	"errors"
	"testing"

	"github.com/rayhanvito/nojpos/backend/internal/shift"
	"github.com/rayhanvito/nojpos/backend/internal/staff"
)

// Tes unit logika validasi service finance yang benar-benar ada:
// validDebtStatus + gerbang pembayaran di Service (utang-piutang) dan
// normalizeExpense + linkShift di ExpenseService.

func TestValidDebtStatus(t *testing.T) {
	for _, status := range []string{"", "unpaid", "partially_paid", "paid"} {
		if !validDebtStatus(status) {
			t.Fatalf("status %q harus valid", status)
		}
	}
	for _, status := range []string{"lunas", "PAID", "overdue"} {
		if validDebtStatus(status) {
			t.Fatalf("status %q harus tidak valid", status)
		}
	}
}

func TestListReceivablesInvalidStatus(t *testing.T) {
	svc := NewService(nil, nil)
	if _, err := svc.ListReceivables(context.Background(), 1, "lunas", false); !errors.Is(err, ErrDebtStatusInvalid) {
		t.Fatalf("status tak valid harus ErrDebtStatusInvalid, dapat %v", err)
	}
	if _, err := svc.ListPayables(context.Background(), 1, "lunas"); !errors.Is(err, ErrDebtStatusInvalid) {
		t.Fatalf("status tak valid harus ErrDebtStatusInvalid, dapat %v", err)
	}
}

func TestPayInvalidInputRejectedBeforeRepo(t *testing.T) {
	svc := NewService(nil, nil)
	if _, err := svc.PayReceivable(context.Background(), 1, 1, 0, "Cash", "kasir"); !errors.Is(err, ErrDebtPaymentInput) {
		t.Fatalf("nominal 0 harus ErrDebtPaymentInput, dapat %v", err)
	}
	if _, err := svc.PayReceivable(context.Background(), 1, 1, 1000, "E-Wallet", "kasir"); !errors.Is(err, ErrDebtPaymentInput) {
		t.Fatalf("metode non-standar harus ErrDebtPaymentInput, dapat %v", err)
	}
	if _, err := svc.PayPayable(context.Background(), 1, 1, -5, "Cash", "kasir"); !errors.Is(err, ErrDebtPaymentInput) {
		t.Fatalf("nominal negatif harus ErrDebtPaymentInput, dapat %v", err)
	}
	if _, err := svc.PayPayable(context.Background(), 1, 1, 1000, "Credit", "kasir"); !errors.Is(err, ErrDebtPaymentInput) {
		t.Fatalf("metode Credit harus ErrDebtPaymentInput, dapat %v", err)
	}
}

func TestNormalizeExpenseValidation(t *testing.T) {
	base := ExpenseInput{Date: "2026-10-16", Category: "Listrik & Air", Amount: 50000, Method: "Cash", Source: "outside"}

	invalid := []ExpenseInput{
		{Date: "2026-10-16", Category: "  ", Amount: 50000, Method: "Cash"},
		{Date: "2026-10-16", Category: "Sewa", Amount: 0, Method: "Cash"},
		{Date: "16-10-2026", Category: "Sewa", Amount: 50000, Method: "Cash"},
		{Date: "2026-10-16", Category: "Sewa", Amount: 50000, Method: "Credit"},
		{Date: "2026-10-16", Category: "Sewa", Amount: 50000, Method: "Cash", Source: "kantong"},
	}
	for i, in := range invalid {
		if _, err := normalizeExpense(in); !errors.Is(err, ErrExpenseInvalid) {
			t.Fatalf("kasus %d harus ErrExpenseInvalid, dapat %v", i, err)
		}
	}

	expense, err := normalizeExpense(base)
	if err != nil {
		t.Fatalf("input valid tidak boleh gagal: %v", err)
	}
	if expense.Category != "Listrik & Air" || expense.Amount != 50000 || expense.Source != ExpenseSourceOutside {
		t.Fatalf("hasil normalisasi salah: %+v", expense)
	}
}

func TestNormalizeExpenseSourceRules(t *testing.T) {
	// Source kosong menjadi outside.
	expense, err := normalizeExpense(ExpenseInput{Date: "2026-10-16", Category: "Sewa", Amount: 1000, Method: "Cash"})
	if err != nil || expense.Source != ExpenseSourceOutside {
		t.Fatalf("source kosong harus outside, dapat %v %+v", err, expense)
	}
	// Metode non-tunai tidak boleh dari laci.
	expense, err = normalizeExpense(ExpenseInput{Date: "2026-10-16", Category: "Sewa", Amount: 1000, Method: "QRIS", Source: "drawer"})
	if err != nil || expense.Source != ExpenseSourceOutside {
		t.Fatalf("non-tunai harus dipaksa outside, dapat %v %+v", err, expense)
	}
	// Tunai + drawer dipertahankan.
	expense, err = normalizeExpense(ExpenseInput{Date: "2026-10-16", Category: "Sewa", Amount: 1000, Method: "Cash", Source: "drawer"})
	if err != nil || expense.Source != ExpenseSourceDrawer {
		t.Fatalf("tunai drawer harus dipertahankan, dapat %v %+v", err, expense)
	}
}

// fakeExpenses / fakeCashiers / fakeShifts adalah palsu minimal untuk
// jalur Create ExpenseService: hanya perilaku yang dikonfigurasi tes.
type fakeExpenses struct {
	created *Expense
	byID    *Expense
	err     error
}

func (f *fakeExpenses) List(ctx context.Context, storeID uint, filter ExpenseFilter) (*ExpensePage, error) {
	return &ExpensePage{Items: []*Expense{}}, nil
}

func (f *fakeExpenses) GetByID(ctx context.Context, storeID, id uint) (*Expense, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.byID, nil
}

func (f *fakeExpenses) Create(ctx context.Context, expense *Expense) (*Expense, error) {
	f.created = expense
	return expense, nil
}

func (f *fakeExpenses) Update(ctx context.Context, storeID, id uint, expense *Expense) (*Expense, error) {
	if f.err != nil {
		return nil, f.err
	}
	return expense, nil
}

func (f *fakeExpenses) Delete(ctx context.Context, storeID, id uint) error { return f.err }

type fakeCashiers struct {
	name string
	err  error
}

func (f *fakeCashiers) GetByIDUnscoped(ctx context.Context, id uint) (*staff.Cashier, error) {
	if f.err != nil {
		return nil, f.err
	}
	return &staff.Cashier{Name: f.name}, nil
}

type fakeShifts struct {
	active *shift.CashierShift
	err    error
}

func (f *fakeShifts) GetActiveShift(ctx context.Context, storeID, outletID, cashierID uint) (*shift.CashierShift, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.active, nil
}

func TestExpenseCreateLinksOpenShiftForDrawerCash(t *testing.T) {
	expenses := &fakeExpenses{}
	shiftID := uint(77)
	openShift := &shift.CashierShift{}
	openShift.ID = shiftID
	svc := NewExpenseService(expenses, &fakeCashiers{name: "Andi"}, &fakeShifts{active: openShift})

	cashierID := uint(5)
	created, err := svc.Create(context.Background(), 1, 2, &cashierID, "fallback", ExpenseInput{
		Date: "2026-10-16", Category: "Kemasan", Amount: 25000, Method: "Cash", Source: "drawer",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ShiftID == nil || *created.ShiftID != shiftID {
		t.Fatalf("pengeluaran laci tunai harus tertaut shift terbuka, dapat %+v", created.ShiftID)
	}
	if created.CreatedByID != cashierID || created.CreatedByName != "Andi" {
		t.Fatalf("pencatat harus dari record kasir, dapat %d %q", created.CreatedByID, created.CreatedByName)
	}
}

func TestExpenseCreateNoShiftStaysUnlinked(t *testing.T) {
	expenses := &fakeExpenses{}
	svc := NewExpenseService(expenses, &fakeCashiers{}, &fakeShifts{err: errors.New("tidak ada shift")})
	created, err := svc.Create(context.Background(), 1, 2, nil, "Admin", ExpenseInput{
		Date: "2026-10-16", Category: "Kemasan", Amount: 25000, Method: "Cash", Source: "drawer",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.ShiftID != nil {
		t.Fatalf("tanpa shift terbuka pengeluaran tidak tertaut, dapat %+v", created.ShiftID)
	}
}

func TestExpenseUpdateDeleteNotFoundMapping(t *testing.T) {
	svc := NewExpenseService(&fakeExpenses{err: ErrNotFound}, &fakeCashiers{}, &fakeShifts{})
	if _, err := svc.Update(context.Background(), 1, 9, ExpenseInput{Date: "2026-10-16", Category: "Sewa", Amount: 1000, Method: "Cash"}); !errors.Is(err, ErrExpenseNotFound) {
		t.Fatalf("Update harus ErrExpenseNotFound, dapat %v", err)
	}
	if err := svc.Delete(context.Background(), 1, 9); !errors.Is(err, ErrExpenseNotFound) {
		t.Fatalf("Delete harus ErrExpenseNotFound, dapat %v", err)
	}
}
