package sales

import (
	"context"
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

// Tes Unit C: checkout memakai tarif pajak EFEKTIF outlet transaksi —
// override outlet_settings.tax_pct menimpa tarif toko lewat
// SettingsResolver (outlet.SettingsService asli di atas sqlite).
// Toko: pajak aktif 11%. Produk Rp10.000 × 1.
func TestCheckoutUsesOutletTaxOverride(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "checkout-outlet-tax.db")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.Migrate(db,
		&outlet.Store{}, &outlet.Outlet{}, &outlet.OutletSettings{}, &staff.Cashier{}, &shift.CashierShift{},
		&catalog.Product{}, &catalog.ProductVariant{}, &catalog.PriceTier{},
		&inventory.ProductStock{}, &inventory.Movement{},
		&customer.Customer{}, &promo.Promo{},
		&Trx{}, &TransactionItem{}, &Refund{}, &finance.Receivable{},
	); err != nil {
		t.Fatal(err)
	}

	store := &outlet.Store{
		Name: "Toko Uji", Timezone: "Asia/Jakarta",
		TaxEnabled: true, TaxPct: 11, TaxRounding: "none",
		LoyaltyEarnPer: 10000, LoyaltyPointValue: 100,
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
		StoreID: store.ID, OutletID: &outletRec.ID, Name: "Produk Uji", SKU: "SKU-PAJAK-1",
		BuyPrice: 2000, SellPrice: 10000, Stock: 10, Active: true,
	}
	if err := db.Create(product).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&inventory.ProductStock{ProductID: product.ID, OutletID: outletRec.ID, Qty: 10}).Error; err != nil {
		t.Fatal(err)
	}

	transactor := database.NewTransactor(db)
	svc := NewService(NewRepo(db, transactor), staff.NewRepo(db), catalog.NewProductRepo(db), shift.NewRepo(db), nil, transactor)
	svc.SetSettingsResolver(outlet.NewSettingsService(outlet.NewStoreRepo(db)))

	checkout := func(paid int64) *Trx {
		trx, _, err := svc.Checkout(context.Background(), store.ID, outletRec.ID, &cashier.ID, "Kasir Uji", CheckoutCmd{
			Items:  []CheckoutItemCmd{{ProductID: product.ID, Qty: 1}},
			Method: "Cash",
			Paid:   paid,
		})
		if err != nil {
			t.Fatalf("checkout: %v", err)
		}
		return trx
	}

	// Tanpa override: tarif toko 11% → pajak 1.100, total 11.100.
	trx := checkout(11100)
	if trx.Tax != 1100 || trx.Total != 11100 {
		t.Fatalf("tanpa override: tax=%d total=%d, want 1100/11100", trx.Tax, trx.Total)
	}

	// Override outlet 0%: pajak 0, total 10.000.
	zero := 0.0
	if err := db.Create(&outlet.OutletSettings{StoreID: store.ID, OutletID: outletRec.ID, TaxPct: &zero}).Error; err != nil {
		t.Fatal(err)
	}
	trx = checkout(10000)
	if trx.Tax != 0 || trx.Total != 10000 {
		t.Fatalf("override 0%%: tax=%d total=%d, want 0/10000", trx.Tax, trx.Total)
	}

	// Override diganti 5%: pajak 500, total 10.500.
	if err := db.Model(&outlet.OutletSettings{}).Where("store_id = ? AND outlet_id = ?", store.ID, outletRec.ID).Update("tax_pct", 5.0).Error; err != nil {
		t.Fatal(err)
	}
	trx = checkout(10500)
	if trx.Tax != 500 || trx.Total != 10500 {
		t.Fatalf("override 5%%: tax=%d total=%d, want 500/10500", trx.Tax, trx.Total)
	}
}
