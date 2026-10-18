package catalog

import (
	"context"
	"errors"
	"testing"
)

func TestCategoryRepoRenamePersistsName(t *testing.T) {
	s := newTenantCatalogSeed(t)
	repo := NewCategoryRepo(s.db)

	got, err := repo.Rename(context.Background(), s.storeA, s.categoryID, "Kategori Baru")
	if err != nil {
		t.Fatal(err)
	}
	var stored Category
	if err := s.db.First(&stored, s.categoryID).Error; err != nil {
		t.Fatal(err)
	}
	if got.Name != "Kategori Baru" || stored.Name != "Kategori Baru" {
		t.Fatalf("nama kategori harus tersimpan, hasil=%q database=%q", got.Name, stored.Name)
	}
}

func TestCategoryRepoRenameDuplicateNameInStore(t *testing.T) {
	s := newTenantCatalogSeed(t)
	other := &Category{StoreID: s.storeA, Name: "Kategori Lain", Active: true}
	if err := s.db.Create(other).Error; err != nil {
		t.Fatal(err)
	}

	_, err := NewCategoryRepo(s.db).Rename(context.Background(), s.storeA, s.categoryID, other.Name)
	if !errors.Is(err, ErrDuplicate) {
		t.Fatalf("nama kategori duplikat dalam toko harus ErrDuplicate, dapat %v", err)
	}
}

func TestCategoryRepoRenameForeignStoreIsNotFound(t *testing.T) {
	s := newTenantCatalogSeed(t)
	_, err := NewCategoryRepo(s.db).Rename(context.Background(), s.storeB, s.categoryID, "Kategori Baru")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("kategori milik toko lain harus ErrNotFound, dapat %v", err)
	}
}
