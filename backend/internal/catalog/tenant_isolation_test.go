package catalog

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

// Seed isolasi lintas-tenant catalog: dua store, satu kategori dan satu
// produk milik storeA dibuat langsung via gorm.
type tenantCatalogSeed struct {
	db         *gorm.DB
	storeA     uint
	storeB     uint
	categoryID uint
	productID  uint
}

func newTenantCatalogSeed(t *testing.T) tenantCatalogSeed {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tenant-catalog.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db, &outlet.Store{}, &outlet.Outlet{}, &Category{}, &Product{}, &ProductVariant{}, &PriceTier{}); err != nil {
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
	cat := &Category{StoreID: storeA.ID, Name: "Kategori A", Active: true}
	if err := db.Create(cat).Error; err != nil {
		t.Fatal(err)
	}
	prod := &Product{StoreID: storeA.ID, CategoryID: &cat.ID, Name: "Produk A", SKU: "SKU-A-1", BuyPrice: 2000, SellPrice: 5000, Active: true}
	if err := db.Create(prod).Error; err != nil {
		t.Fatal(err)
	}
	return tenantCatalogSeed{db: db, storeA: storeA.ID, storeB: storeB.ID, categoryID: cat.ID, productID: prod.ID}
}

func TestTenantIsolationCategoryListByStore(t *testing.T) {
	s := newTenantCatalogSeed(t)
	repo := NewCategoryRepo(s.db)
	ctx := context.Background()

	items, err := repo.ListByStore(ctx, s.storeB)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("ListByStore storeB harus kosong, dapat %d item", len(items))
	}

	items, err = repo.ListByStore(ctx, s.storeA)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != s.categoryID {
		t.Fatalf("kontrol: ListByStore storeA harus berisi kategori seed, dapat %+v", items)
	}
}

func TestTenantIsolationCategoryGetByID(t *testing.T) {
	s := newTenantCatalogSeed(t)
	repo := NewCategoryRepo(s.db)
	ctx := context.Background()

	if _, err := repo.GetByID(ctx, s.storeB, s.categoryID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID storeB atas kategori storeA harus ErrNotFound, dapat %v", err)
	}

	got, err := repo.GetByID(ctx, s.storeA, s.categoryID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != s.categoryID || got.StoreID != s.storeA {
		t.Fatalf("kontrol: GetByID storeA salah, dapat %+v", got)
	}
}

func TestTenantIsolationProductListForStore(t *testing.T) {
	s := newTenantCatalogSeed(t)
	repo := NewProductRepo(s.db)
	ctx := context.Background()

	page, err := repo.ListForStore(ctx, s.storeB, ProductFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 0 || len(page.Items) != 0 {
		t.Fatalf("ListForStore storeB harus kosong, dapat total=%d items=%d", page.Total, len(page.Items))
	}

	page, err = repo.ListForStore(ctx, s.storeA, ProductFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != s.productID {
		t.Fatalf("kontrol: ListForStore storeA harus berisi produk seed, dapat total=%d items=%+v", page.Total, page.Items)
	}
}

func TestTenantIsolationProductGetByIDForStore(t *testing.T) {
	s := newTenantCatalogSeed(t)
	repo := NewProductRepo(s.db)
	ctx := context.Background()

	if _, err := repo.GetByIDForStore(ctx, s.storeB, s.productID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByIDForStore storeB atas produk storeA harus ErrNotFound, dapat %v", err)
	}

	got, err := repo.GetByIDForStore(ctx, s.storeA, s.productID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != s.productID || got.StoreID != s.storeA {
		t.Fatalf("kontrol: GetByIDForStore storeA salah, dapat %+v", got)
	}
}
