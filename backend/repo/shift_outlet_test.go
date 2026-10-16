package repo

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/model"
)

func newShiftOutletRepos(t *testing.T) (*gorm.DB, *CashierRepo, *ShiftRepo, *model.Cashier) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "shift-outlet.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Cashier{}, &model.CashierShift{}); err != nil {
		t.Fatal(err)
	}

	outletAID := uint(10)
	cashier := &model.Cashier{StoreID: 1, OutletID: &outletAID, Name: "Kasir", Active: true}
	if err := db.Create(cashier).Error; err != nil {
		t.Fatal(err)
	}
	return db, NewCashierRepo(db), NewShiftRepo(db), cashier
}

func TestCashierOutletUpdateRejectsOpenShift(t *testing.T) {
	ctx := context.Background()
	_, cashiers, shifts, cashier := newShiftOutletRepos(t)
	if _, err := shifts.StartShift(ctx, cashier.StoreID, *cashier.OutletID, cashier.ID, cashier.Name, 0); err != nil {
		t.Fatal(err)
	}

	outletBID := uint(20)
	if err := cashiers.UpdateProfile(ctx, cashier.StoreID, cashier.ID, nil, &outletBID); !errors.Is(err, ErrCashierOpenShift) {
		t.Fatalf("moving a cashier with an open shift should fail with ErrCashierOpenShift, got %v", err)
	}
	updated, err := cashiers.GetByID(ctx, cashier.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.OutletID == nil || *updated.OutletID != *cashier.OutletID {
		t.Fatalf("cashier outlet changed despite open shift: got %v", updated.OutletID)
	}
}

func TestShiftCannotStartAtCashiersPreviousOutlet(t *testing.T) {
	ctx := context.Background()
	db, cashiers, shifts, cashier := newShiftOutletRepos(t)
	outletAID := *cashier.OutletID
	outletBID := uint(20)
	if err := cashiers.UpdateProfile(ctx, cashier.StoreID, cashier.ID, nil, &outletBID); err != nil {
		t.Fatal(err)
	}

	if _, err := shifts.StartShift(ctx, cashier.StoreID, outletAID, cashier.ID, cashier.Name, 0); !errors.Is(err, ErrCashierOutletChanged) {
		t.Fatalf("starting a shift at the previous outlet should fail with ErrCashierOutletChanged, got %v", err)
	}
	var count int64
	if err := db.Model(&model.CashierShift{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatalf("stale outlet shift should not be created, got %d shifts", count)
	}
}
