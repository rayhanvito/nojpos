package catalog

import (
	"context"
	"testing"
)

func TestProductCountsByStoreScopesToStore(t *testing.T) {
	s := newTenantCatalogSeed(t)
	ctx := context.Background()
	cats := NewCategoryRepo(s.db)
	foreignCategory, err := cats.Create(ctx, s.storeB, "Kategori B")
	if err != nil {
		t.Fatal(err)
	}
	foreignProduct := &Product{
		StoreID: s.storeB, CategoryID: &foreignCategory.ID,
		Name: "Produk B", SKU: "SKU-B-COUNT", Active: true,
	}
	if err := s.db.Create(foreignProduct).Error; err != nil {
		t.Fatal(err)
	}

	repo := NewCategoryRepo(s.db)
	countsA, err := repo.ProductCountsByStore(ctx, s.storeA)
	if err != nil {
		t.Fatal(err)
	}
	if countsA[s.categoryID] != 1 {
		t.Fatalf("store A product count = %d, mau 1", countsA[s.categoryID])
	}
	if _, ok := countsA[foreignCategory.ID]; ok {
		t.Fatalf("ProductCountsByStore store A memuat kategori store B: %+v", countsA)
	}
	countsB, err := repo.ProductCountsByStore(ctx, s.storeB)
	if err != nil {
		t.Fatal(err)
	}
	if countsB[foreignCategory.ID] != 1 {
		t.Fatalf("store B product count = %d, mau 1", countsB[foreignCategory.ID])
	}

	countAForForeignCategory, err := repo.ProductCountByCategory(ctx, s.storeA, foreignCategory.ID)
	if err != nil {
		t.Fatal(err)
	}
	if countAForForeignCategory != 0 {
		t.Fatalf("store A count for store B category = %d, mau 0", countAForForeignCategory)
	}
	countB, err := repo.ProductCountByCategory(ctx, s.storeB, foreignCategory.ID)
	if err != nil {
		t.Fatal(err)
	}
	if countB != 1 {
		t.Fatalf("store B category count = %d, mau 1", countB)
	}
}

func TestListCategoriesCountsInactiveProductsAndExcludesDeleted(t *testing.T) {
	s := newTenantCatalogSeed(t)
	ctx := context.Background()
	cats := NewCategoryRepo(s.db)
	prods := NewProductRepo(s.db)
	emptyCategory, err := cats.Create(ctx, s.storeA, "Kosong")
	if err != nil {
		t.Fatal(err)
	}

	inactive := &Product{
		StoreID: s.storeA, CategoryID: &s.categoryID,
		Name: "Produk Nonaktif", SKU: "SKU-A-INACTIVE", Active: true,
	}
	if err := s.db.Create(inactive).Error; err != nil {
		t.Fatal(err)
	}
	if err := prods.SetActive(ctx, s.storeA, inactive.ID, false); err != nil {
		t.Fatal(err)
	}
	var storedInactive Product
	if err := s.db.First(&storedInactive, inactive.ID).Error; err != nil {
		t.Fatal(err)
	}
	if storedInactive.Active {
		t.Fatal("fixture product must be inactive")
	}

	deleted := &Product{
		StoreID: s.storeA, CategoryID: &s.categoryID,
		Name: "Produk Dihapus", SKU: "SKU-A-DELETED", Active: true,
	}
	if err := s.db.Create(deleted).Error; err != nil {
		t.Fatal(err)
	}
	service := NewService(cats, prods, nil)
	if err := service.DeleteProduct(ctx, s.storeA, deleted.ID); err != nil {
		t.Fatal(err)
	}

	uncategorized := &Product{StoreID: s.storeA, Name: "Produk Tanpa Kategori", SKU: "SKU-A-UNCAT", Active: true}
	if err := s.db.Create(uncategorized).Error; err != nil {
		t.Fatal(err)
	}

	items, err := service.ListCategories(ctx, s.storeA)
	if err != nil {
		t.Fatal(err)
	}
	productCounts, err := cats.ProductCountsByStore(ctx, s.storeA)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := productCounts[0]; ok {
		t.Fatalf("produk tanpa kategori tidak boleh masuk hitungan: %+v", productCounts)
	}
	counts := make(map[uint]*Category, len(items))
	for _, category := range items {
		counts[category.ID] = category
	}
	if counts[s.categoryID] == nil {
		t.Fatalf("kategori seed tidak ada: %+v", items)
	}
	assertCategoryProductCount(t, counts[s.categoryID], 2)
	if counts[emptyCategory.ID] == nil {
		t.Fatalf("kategori kosong tidak ada: %+v", items)
	}
	assertCategoryProductCount(t, counts[emptyCategory.ID], 0)
}

func TestRenameCategoryReturnsProductCount(t *testing.T) {
	s := newTenantCatalogSeed(t)
	service := NewService(NewCategoryRepo(s.db), NewProductRepo(s.db), nil)

	category, err := service.RenameCategory(context.Background(), s.storeA, s.categoryID, "Kategori Baru")
	if err != nil {
		t.Fatal(err)
	}
	assertCategoryProductCount(t, category, 1)
}
