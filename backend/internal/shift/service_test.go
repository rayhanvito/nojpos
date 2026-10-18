package shift

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/rayhanvito/nojpos/backend/internal/staff"
)

// fakeShifts adalah Shifts palsu: perilaku per metode dikonfigurasi
// tes; StartShift mencatat argumen kasir yang diterimanya supaya
// resolveCashier bisa diperiksa.
type fakeShifts struct {
	active    *CashierShift
	activeErr error
	startErr  error
	closeErr  error
	forceErr  error
	listed    []*CashierShift

	gotCashierID   uint
	gotCashierName string
}

func (f *fakeShifts) GetActiveShift(ctx context.Context, storeID, outletID, cashierID uint) (*CashierShift, error) {
	if f.activeErr != nil {
		return nil, f.activeErr
	}
	return f.active, nil
}

func (f *fakeShifts) StartShift(ctx context.Context, storeID, outletID, cashierID uint, cashierName string, openingCash int64) (*CashierShift, error) {
	f.gotCashierID = cashierID
	f.gotCashierName = cashierName
	if f.startErr != nil {
		return nil, f.startErr
	}
	return &CashierShift{StoreID: storeID, CashierID: cashierID, CashierName: cashierName}, nil
}

func (f *fakeShifts) CloseShift(ctx context.Context, storeID, outletID, cashierID uint, closingCash int64, note *string, closedByName string) (*CashierShift, *ShiftStats, error) {
	if f.closeErr != nil {
		return nil, nil, f.closeErr
	}
	return &CashierShift{}, emptyShiftStats(), nil
}

func (f *fakeShifts) ForceCloseShift(ctx context.Context, storeID, shiftID uint, closingCash *int64, note *string, closedByName string) (*CashierShift, *ShiftStats, error) {
	if f.forceErr != nil {
		return nil, nil, f.forceErr
	}
	return &CashierShift{}, emptyShiftStats(), nil
}

func (f *fakeShifts) ListStoreShifts(ctx context.Context, storeID uint, outletID *uint, from string, to string) ([]*CashierShift, error) {
	if f.listed == nil {
		return []*CashierShift{}, nil
	}
	return f.listed, nil
}

func (f *fakeShifts) GetShiftByID(ctx context.Context, storeID, shiftID uint, outletID *uint) (*CashierShift, error) {
	return nil, ErrNotFound
}

func (f *fakeShifts) GetShiftStats(ctx context.Context, shift CashierShift, endTime time.Time) (*ShiftStats, error) {
	return emptyShiftStats(), nil
}

func (f *fakeShifts) ListShiftTransactions(ctx context.Context, storeID uint, shift CashierShift, page, limit int) ([]*trx, int64, error) {
	return []*trx{}, 0, nil
}

type fakeCashiers struct{}

func (fakeCashiers) GetByIDUnscoped(ctx context.Context, id uint) (*staff.Cashier, error) {
	return &staff.Cashier{Name: "Kasir Uji"}, nil
}

type fakeStores struct {
	tz  string
	err error
}

func (f fakeStores) GetTimezone(ctx context.Context, storeID uint) (string, error) {
	return f.tz, f.err
}

func TestStartShiftOutletChangedPassThrough(t *testing.T) {
	shifts := &fakeShifts{startErr: fmt.Errorf("repo: %w", ErrCashierOutletChanged)}
	svc := NewService(shifts, fakeCashiers{}, fakeStores{})
	if _, err := svc.StartShift(context.Background(), 1, 2, nil, "Admin", 0); !errors.Is(err, ErrCashierOutletChanged) {
		t.Fatalf("ErrCashierOutletChanged harus diteruskan apa adanya, dapat %v", err)
	}
}

func TestStartShiftAlreadyOpenPassThrough(t *testing.T) {
	shifts := &fakeShifts{startErr: fmt.Errorf("repo: %w", ErrShiftAlreadyOpen)}
	svc := NewService(shifts, fakeCashiers{}, fakeStores{})
	if _, err := svc.StartShift(context.Background(), 1, 2, nil, "Admin", 0); !errors.Is(err, ErrShiftAlreadyOpen) {
		t.Fatalf("ErrShiftAlreadyOpen harus diteruskan apa adanya, dapat %v", err)
	}
}

func TestStartShiftResolvesCashierName(t *testing.T) {
	shifts := &fakeShifts{}
	svc := NewService(shifts, fakeCashiers{}, fakeStores{})
	cashierID := uint(7)
	if _, err := svc.StartShift(context.Background(), 1, 2, &cashierID, "Fallback", 1000); err != nil {
		t.Fatalf("StartShift: %v", err)
	}
	if shifts.gotCashierID != 7 || shifts.gotCashierName != "Kasir Uji" {
		t.Fatalf("kasir pelaku harus ter-resolve dari repo kasir, dapat id=%d nama=%q", shifts.gotCashierID, shifts.gotCashierName)
	}

	shifts = &fakeShifts{}
	svc = NewService(shifts, fakeCashiers{}, fakeStores{})
	if _, err := svc.StartShift(context.Background(), 1, 2, nil, "Admin Sendiri", 0); err != nil {
		t.Fatalf("StartShift admin: %v", err)
	}
	if shifts.gotCashierID != 0 || shifts.gotCashierName != "Admin Sendiri" {
		t.Fatalf("admin tanpa switch harus cashier_id=0 dengan nama fallback, dapat id=%d nama=%q", shifts.gotCashierID, shifts.gotCashierName)
	}
}

func TestGetCashierShiftNoActive(t *testing.T) {
	svc := NewService(&fakeShifts{activeErr: ErrNotFound}, fakeCashiers{}, fakeStores{})
	if _, err := svc.GetCashierShift(context.Background(), 1, 2, nil, "Admin"); !errors.Is(err, ErrNoActiveShift) {
		t.Fatalf("tanpa shift aktif harus ErrNoActiveShift, dapat %v", err)
	}
}

func TestCloseShiftNotFoundMapsToNoActive(t *testing.T) {
	svc := NewService(&fakeShifts{closeErr: ErrNotFound}, fakeCashiers{}, fakeStores{})
	if _, err := svc.CloseShift(context.Background(), 1, 2, nil, "Admin", 5000, nil); !errors.Is(err, ErrNoActiveShift) {
		t.Fatalf("ErrNotFound dari repo harus menjadi ErrNoActiveShift, dapat %v", err)
	}
}

func TestForceCloseShiftErrorMapping(t *testing.T) {
	svc := NewService(&fakeShifts{forceErr: ErrShiftClosed}, fakeCashiers{}, fakeStores{})
	if _, err := svc.ForceCloseShift(context.Background(), 1, 9, nil, nil, "Admin"); !errors.Is(err, ErrShiftAlreadyClosed) {
		t.Fatalf("ErrShiftClosed dari repo harus menjadi ErrShiftAlreadyClosed, dapat %v", err)
	}
	svc = NewService(&fakeShifts{forceErr: ErrNotFound}, fakeCashiers{}, fakeStores{})
	if _, err := svc.ForceCloseShift(context.Background(), 1, 9, nil, nil, "Admin"); !errors.Is(err, ErrShiftNotFound) {
		t.Fatalf("ErrNotFound dari repo harus menjadi ErrShiftNotFound, dapat %v", err)
	}
}

func TestGetShiftNotFoundMapping(t *testing.T) {
	svc := NewService(&fakeShifts{}, fakeCashiers{}, fakeStores{})
	if _, err := svc.GetShift(context.Background(), 1, 9, nil, 1, 20); !errors.Is(err, ErrShiftNotFound) {
		t.Fatalf("ErrNotFound dari repo harus menjadi ErrShiftNotFound, dapat %v", err)
	}
}

func TestExportShiftsTimezone(t *testing.T) {
	svc := NewService(&fakeShifts{}, fakeCashiers{}, fakeStores{tz: "Asia/Jakarta"})
	_, tz, err := svc.ExportShifts(context.Background(), 1, nil, "", "")
	if err != nil {
		t.Fatalf("ExportShifts: %v", err)
	}
	if tz != "Asia/Jakarta" {
		t.Fatalf("zona waktu toko harus dipakai, dapat %q", tz)
	}

	svc = NewService(&fakeShifts{}, fakeCashiers{}, fakeStores{err: errors.New("db mati")})
	_, tz, err = svc.ExportShifts(context.Background(), 1, nil, "", "")
	if err != nil {
		t.Fatalf("ExportShifts fallback: %v", err)
	}
	if tz != "Asia/Makassar" {
		t.Fatalf("zona waktu fallback harus Asia/Makassar, dapat %q", tz)
	}
}
