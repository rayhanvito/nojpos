package integration

import (
	"github.com/rayhanvito/nojpos/backend/internal/app"
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

func TestPriceTierOwnerAndQuantityAreUnique(t *testing.T) {
	gormDB, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tiers.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := gormDB.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := database.Migrate(gormDB, app.AutoMigrateModels...); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	productTier := catalog.PriceTier{StoreID: 1, ProductID: 10, MinQty: 5, SellPrice: 900}
	if err := gormDB.Create(&productTier).Error; err != nil {
		t.Fatal(err)
	}
	if err := gormDB.Create(&catalog.PriceTier{StoreID: 1, ProductID: 10, MinQty: 5, SellPrice: 800}).Error; err == nil {
		t.Fatal("duplicate product tier quantity was accepted")
	}
	if err := gormDB.Delete(&productTier).Error; err != nil {
		t.Fatal(err)
	}
	if err := gormDB.Create(&catalog.PriceTier{StoreID: 1, ProductID: 10, MinQty: 5, SellPrice: 800}).Error; err != nil {
		t.Fatalf("a deleted product tier quantity can be reused: %v", err)
	}

	variantID := uint(20)
	variantTier := catalog.PriceTier{StoreID: 1, ProductID: 10, VariantID: &variantID, MinQty: 5, SellPrice: 850}
	if err := gormDB.Create(&variantTier).Error; err != nil {
		t.Fatal(err)
	}
	if err := gormDB.Create(&catalog.PriceTier{StoreID: 1, ProductID: 10, VariantID: &variantID, MinQty: 5, SellPrice: 750}).Error; err == nil {
		t.Fatal("duplicate variant tier quantity was accepted")
	}

	otherVariantID := uint(21)
	if err := gormDB.Create(&catalog.PriceTier{StoreID: 1, ProductID: 10, VariantID: &otherVariantID, MinQty: 5, SellPrice: 700}).Error; err != nil {
		t.Fatalf("same quantity is allowed for a different variant: %v", err)
	}
}
