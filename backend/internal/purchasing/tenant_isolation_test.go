package purchasing

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

// Seed isolasi lintas-tenant purchasing: dua store + satu outlet milik
// storeA, satu supplier dan satu PO milik storeA dibuat langsung via
// gorm (bukan seed purchaseOrderTxSeed — file ini berdiri sendiri).
type tenantPurchasingSeed struct {
	db         *gorm.DB
	storeA     uint
	storeB     uint
	outletID   uint
	supplierID uint
	poID       uint
}

func newTenantPurchasingSeed(t *testing.T) tenantPurchasingSeed {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tenant-purchasing.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db, &outlet.Store{}, &outlet.Outlet{}, &Supplier{}, &PurchaseOrder{}, &PurchaseOrderItem{}); err != nil {
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
	sup := &Supplier{StoreID: storeA.ID, Name: "Supplier A"}
	if err := db.Create(sup).Error; err != nil {
		t.Fatal(err)
	}
	po := &PurchaseOrder{StoreID: storeA.ID, OutletID: &out.ID, SupplierID: sup.ID, Number: "PO-A-0001", Status: PurchaseOrderDraft}
	if err := db.Create(po).Error; err != nil {
		t.Fatal(err)
	}
	return tenantPurchasingSeed{db: db, storeA: storeA.ID, storeB: storeB.ID, outletID: out.ID, supplierID: sup.ID, poID: po.ID}
}

func TestTenantIsolationSupplierList(t *testing.T) {
	s := newTenantPurchasingSeed(t)
	repo := NewSupplierRepo(s.db)
	ctx := context.Background()

	items, err := repo.List(ctx, s.storeB)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("List supplier storeB harus kosong, dapat %d item", len(items))
	}

	items, err = repo.List(ctx, s.storeA)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != s.supplierID {
		t.Fatalf("kontrol: List supplier storeA harus berisi supplier seed, dapat %+v", items)
	}
}

func TestTenantIsolationSupplierGetByID(t *testing.T) {
	s := newTenantPurchasingSeed(t)
	repo := NewSupplierRepo(s.db)
	ctx := context.Background()

	if _, err := repo.GetByID(ctx, s.storeB, s.supplierID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID supplier storeB atas supplier storeA harus ErrNotFound, dapat %v", err)
	}

	got, err := repo.GetByID(ctx, s.storeA, s.supplierID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != s.supplierID || got.StoreID != s.storeA {
		t.Fatalf("kontrol: GetByID supplier storeA salah, dapat %+v", got)
	}
}

func TestTenantIsolationPurchaseOrderList(t *testing.T) {
	s := newTenantPurchasingSeed(t)
	repo := NewPurchaseOrderRepo(s.db)
	ctx := context.Background()

	items, err := repo.List(ctx, s.storeB, s.outletID, PurchaseOrderFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("List PO storeB harus kosong, dapat %d item", len(items))
	}

	items, err = repo.List(ctx, s.storeA, s.outletID, PurchaseOrderFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != s.poID {
		t.Fatalf("kontrol: List PO storeA harus berisi PO seed, dapat %+v", items)
	}
}

func TestTenantIsolationPurchaseOrderGetByID(t *testing.T) {
	s := newTenantPurchasingSeed(t)
	repo := NewPurchaseOrderRepo(s.db)
	ctx := context.Background()

	if _, err := repo.GetByID(ctx, s.storeB, s.outletID, s.poID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID PO storeB atas PO storeA harus ErrNotFound, dapat %v", err)
	}

	got, err := repo.GetByID(ctx, s.storeA, s.outletID, s.poID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != s.poID || got.StoreID != s.storeA {
		t.Fatalf("kontrol: GetByID PO storeA salah, dapat %+v", got)
	}
}
