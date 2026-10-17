package auth

import (
	"context"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Seed isolasi lintas-tenant auth: dua store, satu user milik storeA.
// UserRepo TIDAK punya get-by-id yang di-scope store — GetByID dan
// GetByEmail memang global (email unik global; pemanggil menurunkan store
// dari record hasil) — jadi permukaan store-scoped yang diuji adalah
// ListByStore saja.
func TestTenantIsolationUserListByStore(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tenant-auth.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db, &outlet.Store{}, &outlet.Outlet{}, &User{}); err != nil {
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
	user := &User{StoreID: storeA.ID, Email: "pemilik-a@example.com", Name: "Pemilik A", PasswordHash: "hash", Active: true}
	if err := db.Create(user).Error; err != nil {
		t.Fatal(err)
	}

	repo := NewUserRepo(db)
	ctx := context.Background()

	items, err := repo.ListByStore(ctx, storeB.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("ListByStore storeB harus kosong, dapat %d item", len(items))
	}

	items, err = repo.ListByStore(ctx, storeA.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != user.ID {
		t.Fatalf("kontrol: ListByStore storeA harus berisi user seed, dapat %+v", items)
	}
}
