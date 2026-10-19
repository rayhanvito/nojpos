package catalog

import (
	"context"
	"reflect"
	"testing"
)

// DistinctUnits: unik, urut, non-kosong, dan terikat toko — daftar
// satuan form produk diturunkan dari pemakaian nyata per toko.
func TestDistinctUnitsScopesToStore(t *testing.T) {
	s := newTenantCatalogSeed(t)
	ctx := context.Background()
	db := s.db

	mk := func(storeID uint, sku, unit string) *Product {
		p := &Product{StoreID: storeID, Name: "Produk " + sku, SKU: sku, Active: true, Unit: unit}
		if err := db.Create(p).Error; err != nil {
			t.Fatal(err)
		}
		return p
	}
	mk(s.storeA, "SKU-U1", "box")
	mk(s.storeA, "SKU-U2", "box") // duplikat: harus terlipat
	blank := mk(s.storeA, "SKU-U3", "sementara")
	if err := db.Model(blank).Update("unit", "").Error; err != nil {
		t.Fatal(err)
	}
	mk(s.storeB, "SKU-U4", "renteng")

	repo := NewProductRepo(db)
	unitsA, err := repo.DistinctUnits(ctx, s.storeA)
	if err != nil {
		t.Fatal(err)
	}
	// Produk seed toko A berunit default 'pcs'.
	if want := []string{"box", "pcs"}; !reflect.DeepEqual(unitsA, want) {
		t.Fatalf("DistinctUnits store A = %v, mau %v", unitsA, want)
	}
	unitsB, err := repo.DistinctUnits(ctx, s.storeB)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"renteng"}; !reflect.DeepEqual(unitsB, want) {
		t.Fatalf("DistinctUnits store B = %v, mau %v", unitsB, want)
	}
}
