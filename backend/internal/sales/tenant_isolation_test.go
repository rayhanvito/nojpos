package sales

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

// Seed isolasi lintas-tenant sales: dua store, satu transaksi milik
// storeA dibuat langsung via gorm.
type tenantSalesSeed struct {
	db     *gorm.DB
	repo   *Repo
	storeA uint
	storeB uint
	trxID  uint
}

func newTenantSalesSeed(t *testing.T) tenantSalesSeed {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tenant-sales.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db, &outlet.Store{}, &outlet.Outlet{}, &Trx{}, &TransactionItem{}, &Refund{}); err != nil {
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
	trx := &Trx{
		StoreID: storeA.ID, CashierID: 1, CashierName: "Kasir A",
		Items:    TrxItemList{{ProductID: 1, Name: "Produk A", BuyPrice: 2000, Price: 5000, Qty: 1}},
		Subtotal: 5000, Total: 5000, Method: "Cash", Paid: 5000, Status: TrxCompleted,
	}
	if err := db.Create(trx).Error; err != nil {
		t.Fatal(err)
	}
	return tenantSalesSeed{db: db, repo: NewRepo(db, database.NewTransactor(db)), storeA: storeA.ID, storeB: storeB.ID, trxID: trx.ID}
}

func TestTenantIsolationSalesList(t *testing.T) {
	s := newTenantSalesSeed(t)
	ctx := context.Background()

	items, total, err := s.repo.List(ctx, s.storeB, nil, 0, "", "", "", "", "", "", nil, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 || len(items) != 0 {
		t.Fatalf("List storeB harus kosong, dapat total=%d items=%d", total, len(items))
	}

	items, total, err = s.repo.List(ctx, s.storeA, nil, 0, "", "", "", "", "", "", nil, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].ID != s.trxID {
		t.Fatalf("kontrol: List storeA harus berisi transaksi seed, dapat total=%d items=%+v", total, items)
	}
}

func TestTenantIsolationSalesGetByID(t *testing.T) {
	s := newTenantSalesSeed(t)
	ctx := context.Background()

	if _, err := s.repo.GetByID(ctx, s.storeB, s.trxID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID storeB atas transaksi storeA harus ErrNotFound, dapat %v", err)
	}

	got, err := s.repo.GetByID(ctx, s.storeA, s.trxID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != s.trxID || got.StoreID != s.storeA {
		t.Fatalf("kontrol: GetByID storeA salah, dapat %+v", got)
	}
}
