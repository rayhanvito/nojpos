package inventory

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Seed isolasi lintas-tenant inventory: dua store + satu outlet milik
// storeA, satu movement dan satu sesi opname milik storeA dibuat
// langsung via gorm. MovementRepo tidak punya get-by-id, jadi
// permukaannya adalah list; StockOpnameRepo punya list + get.
type tenantInventorySeed struct {
	db        *gorm.DB
	storeA    uint
	storeB    uint
	outletID  uint
	sessionID uint
}

func newTenantInventorySeed(t *testing.T) tenantInventorySeed {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tenant-inventory.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db,
		&outlet.Store{}, &outlet.Outlet{}, &catalog.Product{}, &catalog.ProductVariant{},
		&Movement{}, &StockOpnameSession{}, &StockOpnameItem{},
	); err != nil {
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
	product := &catalog.Product{StoreID: storeA.ID, OutletID: &out.ID, Name: "Produk A", SKU: "SKU-A-1", Active: true}
	if err := db.Create(product).Error; err != nil {
		t.Fatal(err)
	}
	mv := &Movement{StoreID: storeA.ID, OutletID: &out.ID, ProductID: product.ID, Type: MovementAdjust, Qty: 1, Actor: "tester"}
	if err := db.Create(mv).Error; err != nil {
		t.Fatal(err)
	}
	session := &StockOpnameSession{StoreID: storeA.ID, OutletID: &out.ID, Status: StockOpnameCounting, CreatedBy: "tester"}
	if err := db.Create(session).Error; err != nil {
		t.Fatal(err)
	}
	return tenantInventorySeed{db: db, storeA: storeA.ID, storeB: storeB.ID, outletID: out.ID, sessionID: session.ID}
}

func TestTenantIsolationMovementListForStore(t *testing.T) {
	s := newTenantInventorySeed(t)
	repo := NewMovementRepo(s.db)
	ctx := context.Background()

	page, err := repo.ListForStore(ctx, s.storeB, MovementFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 0 || len(page.Items) != 0 {
		t.Fatalf("ListForStore storeB harus kosong, dapat total=%d items=%d", page.Total, len(page.Items))
	}

	page, err = repo.ListForStore(ctx, s.storeA, MovementFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("kontrol: ListForStore storeA harus berisi movement seed, dapat total=%d items=%d", page.Total, len(page.Items))
	}
}

func TestTenantIsolationStockOpnameList(t *testing.T) {
	s := newTenantInventorySeed(t)
	repo := NewStockOpnameRepo(s.db)
	ctx := context.Background()

	items, err := repo.List(ctx, s.storeB, s.outletID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("List opname storeB harus kosong, dapat %d item", len(items))
	}

	items, err = repo.List(ctx, s.storeA, s.outletID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != s.sessionID {
		t.Fatalf("kontrol: List opname storeA harus berisi sesi seed, dapat %+v", items)
	}
}

func TestTenantIsolationStockOpnameGetByID(t *testing.T) {
	s := newTenantInventorySeed(t)
	repo := NewStockOpnameRepo(s.db)
	ctx := context.Background()

	if _, err := repo.GetByID(ctx, s.storeB, s.outletID, s.sessionID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID opname storeB atas sesi storeA harus ErrNotFound, dapat %v", err)
	}

	got, err := repo.GetByID(ctx, s.storeA, s.outletID, s.sessionID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != s.sessionID || got.StoreID != s.storeA {
		t.Fatalf("kontrol: GetByID opname storeA salah, dapat %+v", got)
	}
}
