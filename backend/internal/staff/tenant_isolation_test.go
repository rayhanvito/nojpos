package staff

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Seed isolasi lintas-tenant staff: dua store, satu kasir milik storeA.
// Repo kasir TIDAK punya get-by-id yang di-scope store — satu-satunya get
// adalah GetByIDUnscoped yang secara eksplisit terdokumentasi TANPA scope
// (pemanggil wajib memverifikasi StoreID hasil) — jadi permukaan yang
// diuji di sini adalah ListByStore (list store-scoped) dan UpdateProfile
// (satu-satunya jalur baca-untuk-tulis yang mengunci id+store_id).
type tenantStaffSeed struct {
	db        *gorm.DB
	repo      *Repo
	storeA    uint
	storeB    uint
	cashierID uint
}

func newTenantStaffSeed(t *testing.T) tenantStaffSeed {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tenant-staff.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db, &outlet.Store{}, &outlet.Outlet{}, &Cashier{}); err != nil {
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
	cashier := &Cashier{StoreID: storeA.ID, Name: "Kasir A", Active: true}
	if err := db.Create(cashier).Error; err != nil {
		t.Fatal(err)
	}
	return tenantStaffSeed{db: db, repo: NewRepo(db), storeA: storeA.ID, storeB: storeB.ID, cashierID: cashier.ID}
}

func TestTenantIsolationCashierListByStore(t *testing.T) {
	s := newTenantStaffSeed(t)
	ctx := context.Background()

	items, err := s.repo.ListByStore(ctx, s.storeB)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("ListByStore storeB harus kosong, dapat %d item", len(items))
	}

	items, err = s.repo.ListByStore(ctx, s.storeA)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != s.cashierID {
		t.Fatalf("kontrol: ListByStore storeA harus berisi kasir seed, dapat %+v", items)
	}
}

func TestTenantIsolationCashierUpdateProfile(t *testing.T) {
	s := newTenantStaffSeed(t)
	ctx := context.Background()

	name := "Nama Lintas Tenant"
	if err := s.repo.UpdateProfile(ctx, s.storeB, s.cashierID, &name, nil); !errors.Is(err, ErrNotFound) {
		t.Fatalf("UpdateProfile storeB atas kasir storeA harus ErrNotFound, dapat %v", err)
	}

	var after Cashier
	if err := s.db.First(&after, s.cashierID).Error; err != nil {
		t.Fatal(err)
	}
	if after.Name != "Kasir A" {
		t.Fatalf("kontrol negatif: nama kasir tidak boleh berubah, dapat %q", after.Name)
	}

	if err := s.repo.UpdateProfile(ctx, s.storeA, s.cashierID, &name, nil); err != nil {
		t.Fatal(err)
	}
	if err := s.db.First(&after, s.cashierID).Error; err != nil {
		t.Fatal(err)
	}
	if after.Name != name {
		t.Fatalf("kontrol: UpdateProfile storeA harus mengubah nama, dapat %q", after.Name)
	}
}
