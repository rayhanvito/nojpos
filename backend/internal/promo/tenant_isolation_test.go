package promo

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Seed isolasi lintas-tenant promo: dua store, satu promo milik storeA
// dibuat langsung via gorm.
type tenantPromoSeed struct {
	db      *gorm.DB
	repo    *Repo
	storeA  uint
	storeB  uint
	promoID uint
}

func newTenantPromoSeed(t *testing.T) tenantPromoSeed {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tenant-promo.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db, &outlet.Store{}, &Promo{}); err != nil {
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
	now := time.Now()
	p := &Promo{
		StoreID: storeA.ID, Name: "Promo A", Type: "percent", Value: 10, Scope: "all",
		StartsAt: now.Add(-time.Hour), EndsAt: now.Add(24 * time.Hour), Active: true,
	}
	if err := db.Create(p).Error; err != nil {
		t.Fatal(err)
	}
	return tenantPromoSeed{db: db, repo: NewRepo(db), storeA: storeA.ID, storeB: storeB.ID, promoID: p.ID}
}

func TestTenantIsolationPromoList(t *testing.T) {
	s := newTenantPromoSeed(t)
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
	if len(items) != 1 || items[0].ID != s.promoID {
		t.Fatalf("kontrol: List storeA harus berisi promo seed, dapat %+v", items)
	}
}

func TestTenantIsolationPromoGetByID(t *testing.T) {
	s := newTenantPromoSeed(t)
	ctx := context.Background()

	if _, err := s.repo.GetByID(ctx, s.storeB, s.promoID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID storeB atas promo storeA harus ErrNotFound, dapat %v", err)
	}

	got, err := s.repo.GetByID(ctx, s.storeA, s.promoID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != s.promoID || got.StoreID != s.storeA {
		t.Fatalf("kontrol: GetByID storeA salah, dapat %+v", got)
	}
}
