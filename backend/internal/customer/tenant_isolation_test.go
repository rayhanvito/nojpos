package customer

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

// Seed isolasi lintas-tenant customer: dua store, satu pelanggan milik
// storeA dibuat langsung via gorm.
type tenantCustomerSeed struct {
	db         *gorm.DB
	repo       *Repo
	storeA     uint
	storeB     uint
	customerID uint
}

func newTenantCustomerSeed(t *testing.T) tenantCustomerSeed {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tenant-customer.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db, &outlet.Store{}, &Customer{}); err != nil {
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
	cust := &Customer{StoreID: storeA.ID, Name: "Pelanggan A", Phone: "0812000001"}
	if err := db.Create(cust).Error; err != nil {
		t.Fatal(err)
	}
	return tenantCustomerSeed{db: db, repo: NewRepo(db), storeA: storeA.ID, storeB: storeB.ID, customerID: cust.ID}
}

func TestTenantIsolationCustomerList(t *testing.T) {
	s := newTenantCustomerSeed(t)
	ctx := context.Background()

	page, err := s.repo.List(ctx, s.storeB, CustomerFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 0 || len(page.Items) != 0 {
		t.Fatalf("List storeB harus kosong, dapat total=%d items=%d", page.Total, len(page.Items))
	}

	page, err = s.repo.List(ctx, s.storeA, CustomerFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != s.customerID {
		t.Fatalf("kontrol: List storeA harus berisi pelanggan seed, dapat total=%d items=%+v", page.Total, page.Items)
	}
}

func TestTenantIsolationCustomerGetByID(t *testing.T) {
	s := newTenantCustomerSeed(t)
	ctx := context.Background()

	if _, err := s.repo.GetByID(ctx, s.storeB, s.customerID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID storeB atas pelanggan storeA harus ErrNotFound, dapat %v", err)
	}

	got, err := s.repo.GetByID(ctx, s.storeA, s.customerID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != s.customerID || got.StoreID != s.storeA {
		t.Fatalf("kontrol: GetByID storeA salah, dapat %+v", got)
	}
}
