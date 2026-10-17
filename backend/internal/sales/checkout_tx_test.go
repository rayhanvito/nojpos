package sales

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/customer"
	"github.com/rayhanvito/nojpos/backend/internal/finance"
	"github.com/rayhanvito/nojpos/backend/internal/inventory"
	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
	"github.com/rayhanvito/nojpos/backend/internal/promo"
	"github.com/rayhanvito/nojpos/backend/internal/shift"
	"github.com/rayhanvito/nojpos/backend/internal/staff"
)

type checkoutTxSeed struct {
	db         *gorm.DB
	svc        *Service
	tx         *database.Transactor
	storeID    uint
	outletID   uint
	cashierID  uint
	productID  uint
	customerID uint
	initialQty int
	buyQty     int
	redeem     int64
	paid       int64
}

// newCheckoutTxSeed menyiapkan skema penuh (database.Migrate dengan
// daftar model eksplisit yang disentuh alur checkout) dan seed:
// store dengan loyalty aktif (earn per Rp10.000, nilai poin Rp100),
// outlet, kasir, SHIFT AKTIF kasir itu di outlet itu, produk harga
// jual Rp5.000 dengan stok outlet 10, dan pelanggan berpoin 50.
func newCheckoutTxSeed(t *testing.T) checkoutTxSeed {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "checkout-tx.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db,
		&outlet.Store{}, &outlet.Outlet{}, &staff.Cashier{}, &shift.CashierShift{},
		&catalog.Product{}, &catalog.ProductVariant{}, &catalog.PriceTier{},
		&inventory.ProductStock{}, &inventory.Movement{},
		&customer.Customer{}, &promo.Promo{},
		&Trx{}, &TransactionItem{}, &Refund{}, &finance.Receivable{},
	); err != nil {
		t.Fatal(err)
	}

	store := &outlet.Store{
		Name: "Toko Uji", Timezone: "Asia/Jakarta",
		LoyaltyEnabled: true, LoyaltyEarnPer: 10000, LoyaltyPointValue: 100,
	}
	if err := db.Create(store).Error; err != nil {
		t.Fatal(err)
	}
	outletRec := &outlet.Outlet{StoreID: store.ID, Name: "Outlet Utama", IsDefault: true}
	if err := db.Create(outletRec).Error; err != nil {
		t.Fatal(err)
	}
	cashier := &staff.Cashier{StoreID: store.ID, OutletID: &outletRec.ID, Name: "Kasir Uji", Active: true}
	if err := db.Create(cashier).Error; err != nil {
		t.Fatal(err)
	}
	activeShift := &shift.CashierShift{
		StoreID: store.ID, OutletID: &outletRec.ID, CashierID: cashier.ID,
		CashierName: cashier.Name, StartedAt: time.Now().UTC(),
	}
	if err := db.Create(activeShift).Error; err != nil {
		t.Fatal(err)
	}
	product := &catalog.Product{
		StoreID: store.ID, OutletID: &outletRec.ID, Name: "Produk Uji", SKU: "SKU-UJI-1",
		BuyPrice: 2000, SellPrice: 5000, Stock: 10, Active: true,
	}
	if err := db.Create(product).Error; err != nil {
		t.Fatal(err)
	}
	stock := &inventory.ProductStock{ProductID: product.ID, OutletID: outletRec.ID, Qty: 10}
	if err := db.Create(stock).Error; err != nil {
		t.Fatal(err)
	}
	cust := &customer.Customer{StoreID: store.ID, Name: "Pelanggan Uji", Phone: "0812000001", Points: 50}
	if err := db.Create(cust).Error; err != nil {
		t.Fatal(err)
	}

	transactor := database.NewTransactor(db)
	repo := NewRepo(db, transactor)
	svc := NewService(repo, staff.NewRepo(db), catalog.NewProductRepo(db), shift.NewRepo(db), nil, transactor)

	return checkoutTxSeed{
		db: db, svc: svc, tx: transactor,
		storeID: store.ID, outletID: outletRec.ID, cashierID: cashier.ID,
		productID: product.ID, customerID: cust.ID,
		initialQty: 10, buyQty: 2, redeem: 10, paid: 10000,
	}
}

func (s checkoutTxSeed) checkoutCmd() CheckoutCmd {
	return CheckoutCmd{
		Items:        []CheckoutItemCmd{{ProductID: s.productID, Qty: s.buyQty}},
		Method:       "Cash",
		Paid:         s.paid,
		Customer:     "Pelanggan Uji",
		CustomerID:   &s.customerID,
		PointsRedeem: s.redeem,
	}
}

func (s checkoutTxSeed) checkout(ctx context.Context) (*Trx, bool, error) {
	return s.svc.Checkout(ctx, s.storeID, s.outletID, &s.cashierID, "Kasir Uji", s.checkoutCmd())
}

func (s checkoutTxSeed) checkoutWithinOuterTx(forceErr bool) error {
	return s.tx.WithinTx(context.Background(), func(ctx context.Context) error {
		_, _, err := s.checkout(ctx)
		if err != nil {
			return err
		}
		if forceErr {
			return errors.New("paksa gagal di langkah terakhir")
		}
		return nil
	})
}

func (s checkoutTxSeed) stockQty(t *testing.T) int {
	t.Helper()
	var stock inventory.ProductStock
	if err := s.db.Where("product_id = ? AND outlet_id = ? AND variant_id IS NULL", s.productID, s.outletID).First(&stock).Error; err != nil {
		t.Fatal(err)
	}
	return stock.Qty
}

func (s checkoutTxSeed) count(t *testing.T, m any) int64 {
	t.Helper()
	var n int64
	if err := s.db.Model(m).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func (s checkoutTxSeed) customerPoints(t *testing.T) int64 {
	t.Helper()
	var cust customer.Customer
	if err := s.db.First(&cust, s.customerID).Error; err != nil {
		t.Fatal(err)
	}
	return cust.Points
}

// Tes rollback checkout (§6 dokumen arsitektur): checkout lewat
// Service (gerbang shift + repo checkout) yang dipanggil di dalam
// transaksi luar HARUS ikut transaksi itu (join), sehingga error
// paksa di langkah terakhir membatalkan SELURUH checkout — stok,
// movement, transaksi, item transaksi, dan poin pelanggan.
func TestSalesCheckoutRollsBackWithOuterTx(t *testing.T) {
	seed := newCheckoutTxSeed(t)

	if err := seed.checkoutWithinOuterTx(true); err == nil {
		t.Fatal("WithinTx luar harus mengembalikan error paksa")
	}

	if got := seed.stockQty(t); got != seed.initialQty {
		t.Fatalf("stok harus tetap %d sesudah rollback, dapat %d", seed.initialQty, got)
	}
	if got := seed.count(t, &inventory.Movement{}); got != 0 {
		t.Fatalf("movement harus 0 sesudah rollback, dapat %d", got)
	}
	if got := seed.count(t, &Trx{}); got != 0 {
		t.Fatalf("transactions harus 0 sesudah rollback, dapat %d", got)
	}
	if got := seed.count(t, &TransactionItem{}); got != 0 {
		t.Fatalf("transaction_items harus 0 sesudah rollback, dapat %d", got)
	}
	if got := seed.customerPoints(t); got != 50 {
		t.Fatalf("poin pelanggan harus tetap 50 sesudah rollback, dapat %d", got)
	}
}

// Kontrol: checkout yang sama tanpa transaksi luar harus commit
// seluruhnya — stok berkurang tepat, movement penjualan tercatat,
// trx + items tersimpan, dan poin berubah sesuai aturan kode:
// subtotal 2 × Rp5.000 = Rp10.000; tukar 10 poin × Rp100 = diskon
// poin Rp1.000 (dibayar tunai Rp10.000, kembali Rp1.000); poin
// diperoleh = total Rp10.000 / earnPer Rp10.000 = 1, sehingga saldo
// akhir 50 − 10 + 1 = 41.
func TestSalesCheckoutCommitsWithoutOuterTx(t *testing.T) {
	seed := newCheckoutTxSeed(t)

	trx, created, err := seed.checkout(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !created {
		t.Fatal("checkout pertama harus created=true")
	}
	if trx.Total != 10000 || trx.PointsRedeemed != 10 || trx.PointsDiscount != 1000 || trx.PointsEarned != 1 || trx.Paid != 10000 || trx.Change != 1000 {
		t.Fatalf("ringkasan trx tak sesuai aturan kode: total=%d redeemed=%d pointsDiscount=%d earned=%d paid=%d change=%d",
			trx.Total, trx.PointsRedeemed, trx.PointsDiscount, trx.PointsEarned, trx.Paid, trx.Change)
	}
	if len(trx.FinalStocks) != 1 || trx.FinalStocks[0].Stock != seed.initialQty-seed.buyQty {
		t.Fatalf("final_stocks harus stok %d sesudah commit, dapat %+v", seed.initialQty-seed.buyQty, trx.FinalStocks)
	}

	if got, want := seed.stockQty(t), seed.initialQty-seed.buyQty; got != want {
		t.Fatalf("stok harus %d sesudah commit, dapat %d", want, got)
	}
	if got := seed.count(t, &inventory.Movement{}); got != 1 {
		t.Fatalf("movement harus 1 sesudah commit, dapat %d", got)
	}
	var movement inventory.Movement
	if err := seed.db.First(&movement).Error; err != nil {
		t.Fatal(err)
	}
	if movement.Type != inventory.MovementSale || movement.Qty != -seed.buyQty {
		t.Fatalf("movement harus sale qty %d, dapat %s qty %d", -seed.buyQty, movement.Type, movement.Qty)
	}
	if got := seed.count(t, &Trx{}); got != 1 {
		t.Fatalf("transactions harus 1 sesudah commit, dapat %d", got)
	}
	if got := seed.count(t, &TransactionItem{}); got != 1 {
		t.Fatalf("transaction_items harus 1 sesudah commit, dapat %d", got)
	}
	if got := seed.customerPoints(t); got != 41 {
		t.Fatalf("poin pelanggan harus 41 sesudah commit, dapat %d", got)
	}
}
