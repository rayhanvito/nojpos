package db_test

import (
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/rayhanvito/nojpos/backend/db"
	"github.com/rayhanvito/nojpos/backend/model"
)

func TestMigrateBackfillsDefaultOutletIdempotently(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatalf("initial migrate: %v", err)
	}

	store := model.Store{Name: "Toko Lama"}
	if err := database.Create(&store).Error; err != nil {
		t.Fatal(err)
	}
	emptyStore := model.Store{Name: ""}
	if err := database.Create(&emptyStore).Error; err != nil {
		t.Fatal(err)
	}
	product := model.Product{StoreID: store.ID, Name: "Teh", SKU: "TEH", Stock: 10}
	if err := database.Create(&product).Error; err != nil {
		t.Fatal(err)
	}
	variant := model.ProductVariant{StoreID: store.ID, ProductID: product.ID, Name: "Kotak", SKU: "TEH-K", Stock: 4}
	if err := database.Create(&variant).Error; err != nil {
		t.Fatal(err)
	}
	user := model.User{StoreID: store.ID, Email: "admin@example.com", Name: "Admin", PasswordHash: "hash", Role: model.RoleAdmin, Active: true}
	if err := database.Create(&user).Error; err != nil {
		t.Fatal(err)
	}
	cashier := model.Cashier{StoreID: store.ID, Name: "Kasir", Active: true}
	if err := database.Create(&cashier).Error; err != nil {
		t.Fatal(err)
	}
	transaction := model.Trx{StoreID: store.ID, CashierID: cashier.ID, CashierName: cashier.Name, Items: model.TrxItemList{}, Method: string(model.PayCash)}
	if err := database.Create(&transaction).Error; err != nil {
		t.Fatal(err)
	}
	shift := model.CashierShift{StoreID: store.ID, CashierID: cashier.ID, CashierName: cashier.Name, StartedAt: time.Now().UTC()}
	if err := database.Create(&shift).Error; err != nil {
		t.Fatal(err)
	}

	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate run 1: %v", err)
	}
	if err := database.Create(&model.Outlet{StoreID: store.ID, Name: "Cabang", IsDefault: true}).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Model(&model.Outlet{}).Where("store_id = ?", emptyStore.ID).Update("is_default", false).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate run 2: %v", err)
	}

	var defaults []struct {
		ID        uint
		StoreID   uint
		Name      string
		IsDefault bool
	}
	if err := database.Table("outlets").Select("id, store_id, name, is_default").Where("is_default = ?", true).Order("store_id ASC").Find(&defaults).Error; err != nil {
		t.Fatalf("read default outlets: %v", err)
	}
	if len(defaults) != 2 || defaults[0].StoreID != store.ID || defaults[0].Name != store.Name || !defaults[0].IsDefault || defaults[1].StoreID != emptyStore.ID || defaults[1].Name != "Pusat" {
		t.Fatalf("one named default per existing store: %+v", defaults)
	}

	var linked struct{ OutletID *uint }
	for _, tableAndID := range []struct {
		table string
		id    uint
	}{{"products", product.ID}, {"product_variants", variant.ID}, {"users", user.ID}, {"cashiers", cashier.ID}, {"transactions", transaction.ID}, {"cashier_shifts", shift.ID}} {
		if err := database.Table(tableAndID.table).Select("outlet_id").Where("id = ?", tableAndID.id).Scan(&linked).Error; err != nil {
			t.Fatalf("read %s outlet: %v", tableAndID.table, err)
		}
		if linked.OutletID == nil || *linked.OutletID != defaults[0].ID {
			t.Fatalf("%s id=%d outlet: got %v want store default %d", tableAndID.table, tableAndID.id, linked.OutletID, defaults[0].ID)
		}
	}
	var openShift struct {
		ClosedAt *time.Time
	}
	if err := database.Table("cashier_shifts").Select("closed_at").Where("id = ?", shift.ID).Scan(&openShift).Error; err != nil {
		t.Fatal(err)
	}
	if openShift.ClosedAt != nil {
		t.Fatal("migration closed the active shift")
	}

	var seeded []struct {
		ProductID uint
		VariantID *uint
		OutletID  uint
		Qty       int
	}
	if err := database.Table("product_stocks").Order("variant_id ASC").Find(&seeded).Error; err != nil {
		t.Fatalf("read product stocks: %v", err)
	}
	if len(seeded) != 2 || seeded[0].ProductID != product.ID || seeded[0].VariantID != nil || seeded[0].Qty != 10 || seeded[1].ProductID != product.ID || seeded[1].VariantID == nil || *seeded[1].VariantID != variant.ID || seeded[1].Qty != 4 {
		t.Fatalf("default outlet stock backfill: %+v", seeded)
	}
	if err := database.Table("product_stocks").Where("product_id = ? AND variant_id IS NULL AND outlet_id = ?", product.ID, defaults[0].ID).Update("qty", 77).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(database); err != nil {
		t.Fatalf("third migrate: %v", err)
	}
	var qty int
	if err := database.Table("product_stocks").Select("qty").Where("product_id = ? AND variant_id IS NULL AND outlet_id = ?", product.ID, defaults[0].ID).Scan(&qty).Error; err != nil {
		t.Fatal(err)
	}
	if qty != 77 {
		t.Fatalf("repeat backfill overwrote outlet stock: got %d want 77", qty)
	}
}
