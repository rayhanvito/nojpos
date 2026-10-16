package repo

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/adrr-dev/openPOS/backend/model"
)

// Agregat per kasir di bundle laporan: neto = bruto - refund kasir itu,
// item terpotong refund, produk terlaris bersih refund, dan total semua
// kasir harus sama persis dengan ringkasan (satu sumber kebenaran angka).
func TestReportCashiersAggregate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "report-cashiers.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&model.Trx{}, &model.TransactionItem{}, &model.Refund{}, &model.Product{}, &model.ProductStock{}, &model.Outlet{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	p1 := &model.Product{StoreID: 1, Name: "Produk A", SKU: "A", BuyPrice: 5000, SellPrice: 10000}
	p2 := &model.Product{StoreID: 1, Name: "Produk B", SKU: "B", BuyPrice: 2000, SellPrice: 5000}
	if err := db.Create(p1).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(p2).Error; err != nil {
		t.Fatal(err)
	}
	mkTrx := func(cashierID uint, name string, total int64, items []model.TransactionItem) *model.Trx {
		trx := &model.Trx{
			StoreID: 1, CashierID: cashierID, CashierName: name,
			Subtotal: total, Total: total, Paid: total, Method: "Cash",
			Status: model.TrxCompleted,
		}
		trx.CreatedAt = now
		if err := db.Create(trx).Error; err != nil {
			t.Fatal(err)
		}
		for i := range items {
			items[i].TrxID = trx.ID
			items[i].CreatedAt = now
			if err := db.Create(&items[i]).Error; err != nil {
				t.Fatal(err)
			}
		}
		return trx
	}
	a1 := mkTrx(11, "Kasir A", 20000, []model.TransactionItem{{ProductID: p1.ID, Name: "Produk A", Price: 10000, BuyPrice: 5000, Qty: 2}})
	_ = mkTrx(11, "Kasir A", 10000, []model.TransactionItem{{ProductID: p1.ID, Name: "Produk A", Price: 10000, BuyPrice: 5000, Qty: 1}})
	_ = mkTrx(12, "Kasir B", 30000, []model.TransactionItem{{ProductID: p1.ID, Name: "Produk A", Price: 10000, BuyPrice: 5000, Qty: 3}})
	_ = mkTrx(0, "Owner", 5000, []model.TransactionItem{{ProductID: p2.ID, Name: "Produk B", Price: 5000, BuyPrice: 2000, Qty: 1}})
	ref := &model.Refund{StoreID: 1, TrxID: a1.ID, Items: model.RefundItemList{p1.ID: 1}, Reason: "tes", ByName: "Owner"}
	ref.CreatedAt = now
	if err := db.Create(ref).Error; err != nil {
		t.Fatal(err)
	}

	bundle, err := NewReportRepo(db).Report(context.Background(), 1, nil, "today", "Asia/Makassar", "", "", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Cashiers) != 3 {
		t.Fatalf("cashiers = %d, mau 3: %+v", len(bundle.Cashiers), bundle.Cashiers)
	}
	byName := map[string]model.CashierReportRow{}
	var sumNet, sumItems int64
	for _, c := range bundle.Cashiers {
		byName[c.Name] = c
		sumNet += c.Net
		sumItems += c.Items
	}
	if sumNet != bundle.Summary.Omzet {
		t.Fatalf("Σ net kasir %d != omzet ringkasan %d", sumNet, bundle.Summary.Omzet)
	}
	if sumItems != bundle.Summary.ItemsSold {
		t.Fatalf("Σ item kasir %d != items_sold ringkasan %d", sumItems, bundle.Summary.ItemsSold)
	}
	a := byName["Kasir A"]
	if a.Gross != 30000 || a.Refund != 10000 || a.Net != 20000 || a.Trx != 2 || a.Items != 2 || a.TopProduct != "Produk A" || a.TopProductQty != 2 {
		t.Fatalf("Kasir A salah: %+v", a)
	}
	b := byName["Kasir B"]
	if b.Net != 30000 || b.Trx != 1 || b.Items != 3 || b.TopProductQty != 3 {
		t.Fatalf("Kasir B salah: %+v", b)
	}
	o := byName["Owner"]
	if o.Net != 5000 || o.TopProduct != "Produk B" {
		t.Fatalf("Owner salah: %+v", o)
	}
	if bundle.Cashiers[0].Name != "Kasir B" {
		t.Fatalf("urutan harus neto terbesar dulu, dapat %s", bundle.Cashiers[0].Name)
	}
}
