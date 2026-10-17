package purchasing

import (
	"context"
	"errors"
	"github.com/rayhanvito/nojpos/backend/internal/inventory"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/finance"
	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

type purchaseOrderTxSeed struct {
	db         *gorm.DB
	repo       *PurchaseOrderRepo
	storeID    uint
	outletID   uint
	productID  uint
	poID       uint
	itemID     uint
	buyPrice   int64
	initialQty int
	receiveQty int
}

// newPurchaseOrderTxSeed menyiapkan skema penuh (database.Migrate dengan
// daftar model eksplisit yang disentuh alur PO; paket database tidak
// mengimpor model sehingga tidak ada siklus impor dari paket repo lama) dan seed:
// store, outlet, supplier, produk tanpa varian, stok awal 10, dan PO ordered
// berisi 1 item qty_ordered 5.
func newPurchaseOrderTxSeed(t *testing.T) purchaseOrderTxSeed {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "purchase-order-tx.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db,
		&outlet.Store{}, &outlet.Outlet{}, &Supplier{},
		&catalog.Product{}, &catalog.ProductVariant{}, &inventory.ProductStock{},
		&inventory.Movement{}, &PurchaseOrder{}, &PurchaseOrderItem{},
		&finance.Payable{},
	); err != nil {
		t.Fatal(err)
	}

	store := &outlet.Store{Name: "Toko Uji"}
	if err := db.Create(store).Error; err != nil {
		t.Fatal(err)
	}
	outlet := &outlet.Outlet{StoreID: store.ID, Name: "Outlet Utama", IsDefault: true}
	if err := db.Create(outlet).Error; err != nil {
		t.Fatal(err)
	}
	supplier := &Supplier{StoreID: store.ID, Name: "Supplier Uji"}
	if err := db.Create(supplier).Error; err != nil {
		t.Fatal(err)
	}
	product := &catalog.Product{
		StoreID: store.ID, OutletID: &outlet.ID, Name: "Produk Uji", SKU: "SKU-UJI-1",
		BuyPrice: 2000, SellPrice: 5000, Stock: 10, Active: true,
	}
	if err := db.Create(product).Error; err != nil {
		t.Fatal(err)
	}
	stock := &inventory.ProductStock{ProductID: product.ID, OutletID: outlet.ID, Qty: 10}
	if err := db.Create(stock).Error; err != nil {
		t.Fatal(err)
	}
	po := &PurchaseOrder{
		StoreID: store.ID, OutletID: &outlet.ID, SupplierID: supplier.ID,
		Number: "PO-UJI-0001", Status: PurchaseOrderOrdered,
	}
	if err := db.Create(po).Error; err != nil {
		t.Fatal(err)
	}
	item := &PurchaseOrderItem{
		PurchaseOrderID: po.ID, ProductID: product.ID,
		QtyOrdered: 5, QtyReceived: 0, BuyPrice: 2000,
	}
	if err := db.Create(item).Error; err != nil {
		t.Fatal(err)
	}

	return purchaseOrderTxSeed{
		db: db, repo: NewPurchaseOrderRepo(db),
		storeID: store.ID, outletID: outlet.ID, productID: product.ID,
		poID: po.ID, itemID: item.ID,
		buyPrice: 2000, initialQty: 10, receiveQty: 5,
	}
}

func (s purchaseOrderTxSeed) receiveWithinOuterTx(forceErr bool) error {
	return database.NewTransactor(s.db).WithinTx(context.Background(), func(ctx context.Context) error {
		_, err := s.repo.Receive(ctx, s.storeID, s.outletID, s.poID,
			[]PurchaseOrderReceiptItem{{ItemID: s.itemID, Qty: s.receiveQty}}, "tester")
		if err != nil {
			return err
		}
		if forceErr {
			return errors.New("paksa gagal di langkah terakhir")
		}
		return nil
	})
}

func (s purchaseOrderTxSeed) stockQty(t *testing.T) int {
	t.Helper()
	var stock inventory.ProductStock
	if err := s.db.Where("product_id = ? AND outlet_id = ? AND variant_id IS NULL", s.productID, s.outletID).First(&stock).Error; err != nil {
		t.Fatal(err)
	}
	return stock.Qty
}

func (s purchaseOrderTxSeed) count(t *testing.T, m any) int64 {
	t.Helper()
	var n int64
	if err := s.db.Model(m).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func (s purchaseOrderTxSeed) reloadPO(t *testing.T) (PurchaseOrder, PurchaseOrderItem) {
	t.Helper()
	var po PurchaseOrder
	if err := s.db.First(&po, s.poID).Error; err != nil {
		t.Fatal(err)
	}
	var item PurchaseOrderItem
	if err := s.db.First(&item, s.itemID).Error; err != nil {
		t.Fatal(err)
	}
	return po, item
}

// Pilot Transactor: Receive yang dipanggil di dalam transaksi luar HARUS ikut
// transaksi itu (join), sehingga error paksa di langkah terakhir membatalkan
// SELURUH penerimaan — stok, movement, payable, status PO, dan qty_received.
func TestPurchaseOrderReceiveRollsBackWithOuterTx(t *testing.T) {
	seed := newPurchaseOrderTxSeed(t)

	if err := seed.receiveWithinOuterTx(true); err == nil {
		t.Fatal("WithinTx luar harus mengembalikan error paksa")
	}

	if got := seed.stockQty(t); got != seed.initialQty {
		t.Fatalf("stok harus tetap %d sesudah rollback, dapat %d", seed.initialQty, got)
	}
	if got := seed.count(t, &inventory.Movement{}); got != 0 {
		t.Fatalf("movement harus 0 sesudah rollback, dapat %d", got)
	}
	if got := seed.count(t, &finance.Payable{}); got != 0 {
		t.Fatalf("payable harus 0 sesudah rollback, dapat %d", got)
	}
	po, item := seed.reloadPO(t)
	if po.Status != PurchaseOrderOrdered {
		t.Fatalf("status PO harus tetap ordered, dapat %s", po.Status)
	}
	if item.QtyReceived != 0 {
		t.Fatalf("qty_received harus tetap 0, dapat %d", item.QtyReceived)
	}
}

// Kontrol: alur yang sama tanpa error paksa harus commit seluruhnya.
func TestPurchaseOrderReceiveCommitsWithOuterTx(t *testing.T) {
	seed := newPurchaseOrderTxSeed(t)

	if err := seed.receiveWithinOuterTx(false); err != nil {
		t.Fatal(err)
	}

	if got, want := seed.stockQty(t), seed.initialQty+seed.receiveQty; got != want {
		t.Fatalf("stok harus %d sesudah commit, dapat %d", want, got)
	}
	if got := seed.count(t, &inventory.Movement{}); got != 1 {
		t.Fatalf("movement harus 1 sesudah commit, dapat %d", got)
	}
	var movement inventory.Movement
	if err := seed.db.First(&movement).Error; err != nil {
		t.Fatal(err)
	}
	if movement.Type != inventory.MovementPurchase {
		t.Fatalf("tipe movement harus purchase, dapat %s", movement.Type)
	}
	if got := seed.count(t, &finance.Payable{}); got != 1 {
		t.Fatalf("payable harus 1 sesudah commit, dapat %d", got)
	}
	var payable finance.Payable
	if err := seed.db.First(&payable).Error; err != nil {
		t.Fatal(err)
	}
	if want := int64(seed.receiveQty) * seed.buyPrice; payable.Amount != want {
		t.Fatalf("amount payable harus %d, dapat %d", want, payable.Amount)
	}
	po, item := seed.reloadPO(t)
	if po.Status != PurchaseOrderFullyReceived {
		t.Fatalf("status PO harus fully_received, dapat %s", po.Status)
	}
	if item.QtyReceived != seed.receiveQty {
		t.Fatalf("qty_received harus %d, dapat %d", seed.receiveQty, item.QtyReceived)
	}
}
