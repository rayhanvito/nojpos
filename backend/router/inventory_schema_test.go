package router

import (
	"path/filepath"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/rayhanvito/nojpos/backend/db"
	"github.com/rayhanvito/nojpos/backend/model"
)

func TestPriceTierOwnerAndQuantityAreUnique(t *testing.T) {
	database, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "tiers.db")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.Migrate(database); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	productTier := model.PriceTier{StoreID: 1, ProductID: 10, MinQty: 5, SellPrice: 900}
	if err := database.Create(&productTier).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Create(&model.PriceTier{StoreID: 1, ProductID: 10, MinQty: 5, SellPrice: 800}).Error; err == nil {
		t.Fatal("duplicate product tier quantity was accepted")
	}
	if err := database.Delete(&productTier).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Create(&model.PriceTier{StoreID: 1, ProductID: 10, MinQty: 5, SellPrice: 800}).Error; err != nil {
		t.Fatalf("a deleted product tier quantity can be reused: %v", err)
	}

	variantID := uint(20)
	variantTier := model.PriceTier{StoreID: 1, ProductID: 10, VariantID: &variantID, MinQty: 5, SellPrice: 850}
	if err := database.Create(&variantTier).Error; err != nil {
		t.Fatal(err)
	}
	if err := database.Create(&model.PriceTier{StoreID: 1, ProductID: 10, VariantID: &variantID, MinQty: 5, SellPrice: 750}).Error; err == nil {
		t.Fatal("duplicate variant tier quantity was accepted")
	}

	otherVariantID := uint(21)
	if err := database.Create(&model.PriceTier{StoreID: 1, ProductID: 10, VariantID: &otherVariantID, MinQty: 5, SellPrice: 700}).Error; err != nil {
		t.Fatalf("same quantity is allowed for a different variant: %v", err)
	}
}
