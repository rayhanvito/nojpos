package finance

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/customer"
	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Seed isolasi lintas-tenant finance: dua store, satu receivable, satu
// payable, dan satu expense milik storeA dibuat langsung via gorm.
// Tabel suppliers/purchase_orders yang dibaca fillPayables dibuat dari
// baris baca privat slice ini sendiri (payableSupplierRow /
// payableOrderRow) karena purchasing mengimpor finance — mengimpor
// purchasing dari tes paket finance akan menjadi siklus impor.
type tenantFinanceSeed struct {
	db           *gorm.DB
	storeA       uint
	storeB       uint
	receivableID uint
	payableID    uint
	expenseID    uint
}

func newTenantFinanceSeed(t *testing.T) tenantFinanceSeed {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tenant-finance.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db,
		&outlet.Store{}, &customer.Customer{},
		&Receivable{}, &ReceivablePayment{}, &Payable{}, &PayablePayment{}, &Expense{},
		&payableSupplierRow{}, &payableOrderRow{},
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
	cust := &customer.Customer{StoreID: storeA.ID, Name: "Pelanggan A", Phone: "0812000001"}
	if err := db.Create(cust).Error; err != nil {
		t.Fatal(err)
	}
	rec := &Receivable{StoreID: storeA.ID, CustomerID: cust.ID, TransactionID: 101, Amount: 50000, DueDate: time.Now().Add(24 * time.Hour), Status: DebtUnpaid}
	if err := db.Create(rec).Error; err != nil {
		t.Fatal(err)
	}
	pay := &Payable{StoreID: storeA.ID, SupplierID: 1, PurchaseOrderID: 1, Amount: 40000, Status: DebtUnpaid}
	if err := db.Create(pay).Error; err != nil {
		t.Fatal(err)
	}
	exp := &Expense{StoreID: storeA.ID, Date: time.Now(), Category: "Listrik & Air", Amount: 25000, Method: "Cash", Source: ExpenseSourceOutside}
	if err := db.Create(exp).Error; err != nil {
		t.Fatal(err)
	}
	return tenantFinanceSeed{db: db, storeA: storeA.ID, storeB: storeB.ID, receivableID: rec.ID, payableID: pay.ID, expenseID: exp.ID}
}

func TestTenantIsolationReceivableList(t *testing.T) {
	s := newTenantFinanceSeed(t)
	repo := NewDebtRepo(s.db)
	ctx := context.Background()

	items, err := repo.ListReceivables(ctx, s.storeB, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("ListReceivables storeB harus kosong, dapat %d item", len(items))
	}

	items, err = repo.ListReceivables(ctx, s.storeA, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != s.receivableID {
		t.Fatalf("kontrol: ListReceivables storeA harus berisi receivable seed, dapat %+v", items)
	}
}

func TestTenantIsolationReceivableGet(t *testing.T) {
	s := newTenantFinanceSeed(t)
	repo := NewDebtRepo(s.db)
	ctx := context.Background()

	if _, err := repo.GetReceivable(ctx, s.storeB, s.receivableID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetReceivable storeB atas receivable storeA harus ErrNotFound, dapat %v", err)
	}

	got, err := repo.GetReceivable(ctx, s.storeA, s.receivableID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != s.receivableID || got.StoreID != s.storeA {
		t.Fatalf("kontrol: GetReceivable storeA salah, dapat %+v", got)
	}
}

func TestTenantIsolationPayableList(t *testing.T) {
	s := newTenantFinanceSeed(t)
	repo := NewDebtRepo(s.db)
	ctx := context.Background()

	items, err := repo.ListPayables(ctx, s.storeB, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("ListPayables storeB harus kosong, dapat %d item", len(items))
	}

	items, err = repo.ListPayables(ctx, s.storeA, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != s.payableID {
		t.Fatalf("kontrol: ListPayables storeA harus berisi payable seed, dapat %+v", items)
	}
}

func TestTenantIsolationPayableGet(t *testing.T) {
	s := newTenantFinanceSeed(t)
	repo := NewDebtRepo(s.db)
	ctx := context.Background()

	if _, err := repo.GetPayable(ctx, s.storeB, s.payableID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetPayable storeB atas payable storeA harus ErrNotFound, dapat %v", err)
	}

	got, err := repo.GetPayable(ctx, s.storeA, s.payableID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != s.payableID || got.StoreID != s.storeA {
		t.Fatalf("kontrol: GetPayable storeA salah, dapat %+v", got)
	}
}

func TestTenantIsolationExpenseList(t *testing.T) {
	s := newTenantFinanceSeed(t)
	repo := NewExpenseRepo(s.db)
	ctx := context.Background()

	page, err := repo.List(ctx, s.storeB, ExpenseFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 0 || len(page.Items) != 0 {
		t.Fatalf("List expense storeB harus kosong, dapat total=%d items=%d", page.Total, len(page.Items))
	}

	page, err = repo.List(ctx, s.storeA, ExpenseFilter{})
	if err != nil {
		t.Fatal(err)
	}
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0].ID != s.expenseID {
		t.Fatalf("kontrol: List expense storeA harus berisi expense seed, dapat total=%d items=%+v", page.Total, page.Items)
	}
}

func TestTenantIsolationExpenseGetByID(t *testing.T) {
	s := newTenantFinanceSeed(t)
	repo := NewExpenseRepo(s.db)
	ctx := context.Background()

	if _, err := repo.GetByID(ctx, s.storeB, s.expenseID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByID expense storeB atas expense storeA harus ErrNotFound, dapat %v", err)
	}

	got, err := repo.GetByID(ctx, s.storeA, s.expenseID)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != s.expenseID || got.StoreID != s.storeA {
		t.Fatalf("kontrol: GetByID expense storeA salah, dapat %+v", got)
	}
}
