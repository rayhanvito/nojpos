package shift

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Seed isolasi lintas-tenant shift: dua store + satu outlet milik
// storeA, satu shift aktif milik storeA dibuat langsung via gorm.
type tenantShiftSeed struct {
	db        *gorm.DB
	repo      *Repo
	storeA    uint
	storeB    uint
	outletID  uint
	cashierID uint
	shiftID   uint
}

func newTenantShiftSeed(t *testing.T) tenantShiftSeed {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tenant-shift.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db, &outlet.Store{}, &outlet.Outlet{}, &CashierShift{}); err != nil {
		t.Fatal(err)
	}
	storeA := &outlet.Store{Name: "Toko A"}
	if err := db.Create(storeA).Error; err != nil {
		t.Fatal(err)
	}
	storeB := &outlet.Store{Name: "Toko B"}
	if err := db.Create(storeB).Error; err != nil {
		t.Fatal(err)
	}
	out := &outlet.Outlet{StoreID: storeA.ID, Name: "Outlet A", IsDefault: true}
	if err := db.Create(out).Error; err != nil {
		t.Fatal(err)
	}
	sh := &CashierShift{
		StoreID: storeA.ID, OutletID: &out.ID, CashierID: 7, CashierName: "Kasir A",
		StartedAt: time.Now().UTC(), OpeningCash: 100000,
	}
	if err := db.Create(sh).Error; err != nil {
		t.Fatal(err)
	}
	return tenantShiftSeed{db: db, repo: NewRepo(db), storeA: storeA.ID, storeB: storeB.ID, outletID: out.ID, cashierID: 7, shiftID: sh.ID}
}

func TestTenantIsolationShiftListStoreShifts(t *testing.T) {
	s := newTenantShiftSeed(t)
	ctx := context.Background()

	items, err := s.repo.ListStoreShifts(ctx, s.storeB, nil, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("ListStoreShifts storeB harus kosong, dapat %d item", len(items))
	}

	items, err = s.repo.ListStoreShifts(ctx, s.storeA, nil, "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != s.shiftID {
		t.Fatalf("kontrol: ListStoreShifts storeA harus berisi shift seed, dapat %+v", items)
	}
}

func TestTenantIsolationShiftGetShiftByID(t *testing.T) {
	s := newTenantShiftSeed(t)
	ctx := context.Background()

	if _, err := s.repo.GetShiftByID(ctx, s.storeB, s.shiftID, nil); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetShiftByID storeB atas shift storeA harus ErrNotFound, dapat %v", err)
	}

	got, err := s.repo.GetShiftByID(ctx, s.storeA, s.shiftID, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != s.shiftID || got.StoreID != s.storeA {
		t.Fatalf("kontrol: GetShiftByID storeA salah, dapat %+v", got)
	}
}

func TestTenantIsolationShiftGetActiveShift(t *testing.T) {
	s := newTenantShiftSeed(t)
	ctx := context.Background()

	if _, err := s.repo.GetActiveShift(ctx, s.storeB, s.outletID, s.cashierID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetActiveShift storeB harus ErrNotFound, dapat %v", err)
	}

	got, err := s.repo.GetActiveShift(ctx, s.storeA, s.outletID, s.cashierID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != s.shiftID {
		t.Fatalf("kontrol: GetActiveShift storeA salah, dapat %+v", got)
	}
}
