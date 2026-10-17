package outlet

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Seed isolasi lintas-tenant outlet: dua store, satu outlet milik storeA
// dibuat langsung via gorm. Query repo dengan storeB tidak boleh melihat
// outlet storeA sama sekali; kontrol dengan storeA wajib menemukannya.
type tenantOutletSeed struct {
	db       *gorm.DB
	repo     *Repo
	storeA   uint
	storeB   uint
	outletID uint
}

func newTenantOutletSeed(t *testing.T) tenantOutletSeed {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tenant-outlet.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db, &Store{}, &Outlet{}); err != nil {
		t.Fatal(err)
	}
	storeA := &Store{Name: "Toko A"}
	if err := db.Create(storeA).Error; err != nil {
		t.Fatal(err)
	}
	storeB := &Store{Name: "Toko B"}
	if err := db.Create(storeB).Error; err != nil {
		t.Fatal(err)
	}
	out := &Outlet{StoreID: storeA.ID, Name: "Outlet A", IsDefault: true}
	if err := db.Create(out).Error; err != nil {
		t.Fatal(err)
	}
	return tenantOutletSeed{db: db, repo: NewOutletRepo(db), storeA: storeA.ID, storeB: storeB.ID, outletID: out.ID}
}

func TestTenantIsolationOutletList(t *testing.T) {
	s := newTenantOutletSeed(t)
	ctx := context.Background()

	items, err := s.repo.List(ctx, s.storeB)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("List storeB harus kosong, dapat %d item", len(items))
	}

	items, err = s.repo.List(ctx, s.storeA)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != s.outletID {
		t.Fatalf("kontrol: List storeA harus berisi outlet seed, dapat %+v", items)
	}
}

func TestTenantIsolationOutletGetByID(t *testing.T) {
	s := newTenantOutletSeed(t)
	ctx := context.Background()

	if _, err := s.repo.GetByID(ctx, s.storeB, s.outletID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID storeB atas outlet storeA harus ErrNotFound, dapat %v", err)
	}

	got, err := s.repo.GetByID(ctx, s.storeA, s.outletID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != s.outletID || got.StoreID != s.storeA {
		t.Fatalf("kontrol: GetByID storeA salah, dapat %+v", got)
	}
}
