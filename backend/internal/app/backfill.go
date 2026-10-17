package app

import (
	"github.com/rayhanvito/nojpos/backend/internal/inventory"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/rayhanvito/nojpos/backend/internal/auth"
	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	"github.com/rayhanvito/nojpos/backend/internal/purchasing"
	"github.com/rayhanvito/nojpos/backend/internal/sales"
	"github.com/rayhanvito/nojpos/backend/internal/shift"
	"github.com/rayhanvito/nojpos/backend/internal/staff"
)

// backfillDefaultOutlets adalah langkah migrasi data warisan multi-outlet:
// sesudah AutoMigrate (jalur SQLite), setiap toko yang sudah ada diberi
// tepat satu outlet default dan baris-baris lamanya dihubungkan ke outlet
// itu. Idempoten — aman dijalankan berulang. Kode ini verbatim dari paket
// db lama; ia tinggal di akar komposisi karena mengimpor model domain,
// sedangkan internal/platform/database tidak boleh mengimpornya (siklus).

func backfillDefaultOutlets(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var stores []outlet.Store
		if err := tx.Find(&stores).Error; err != nil {
			return err
		}
		for _, store := range stores {
			outlet, err := defaultOutlet(tx, store)
			if err != nil {
				return err
			}
			for _, entity := range []any{
				&catalog.Product{}, &catalog.ProductVariant{}, &auth.User{}, &staff.Cashier{},
				&sales.Trx{}, &shift.CashierShift{}, &purchasing.PurchaseOrder{},
				&inventory.StockOpnameSession{}, &inventory.Movement{},
			} {
				if err := tx.Unscoped().Model(entity).
					Where("store_id = ? AND outlet_id IS NULL", store.ID).
					Update("outlet_id", outlet.ID).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(&auth.RefreshToken{}).
				Where("user_id IN (SELECT id FROM users WHERE store_id = ?) AND outlet_id IS NULL", store.ID).
				Update("outlet_id", outlet.ID).Error; err != nil {
				return err
			}
			if err := seedDefaultOutletStocks(tx, store.ID, outlet.ID); err != nil {
				return err
			}
		}
		return nil
	})
}

func defaultOutlet(tx *gorm.DB, store outlet.Store) (*outlet.Outlet, error) {
	var outlets []outlet.Outlet
	if err := tx.Unscoped().Where("store_id = ?", store.ID).Order("id ASC").Find(&outlets).Error; err != nil {
		return nil, err
	}
	var selected *outlet.Outlet
	for i := range outlets {
		if outlets[i].IsDefault {
			selected = &outlets[i]
			break
		}
	}
	if selected == nil && len(outlets) > 0 {
		selected = &outlets[0]
	}
	if selected == nil {
		name := strings.TrimSpace(store.Name)
		if name == "" {
			name = "Pusat"
		}
		created := outlet.Outlet{StoreID: store.ID, Name: name, Address: store.Address, IsDefault: true}
		if err := tx.Create(&created).Error; err != nil {
			return nil, err
		}
		return &created, nil
	}
	if err := tx.Unscoped().Model(&outlet.Outlet{}).Where("store_id = ?", store.ID).Update("is_default", false).Error; err != nil {
		return nil, err
	}
	if err := tx.Unscoped().Model(&outlet.Outlet{}).Where("id = ?", selected.ID).Update("is_default", true).Error; err != nil {
		return nil, err
	}
	selected.IsDefault = true
	return selected, nil
}

func seedDefaultOutletStocks(tx *gorm.DB, storeID, outletID uint) error {
	var products []catalog.Product
	if err := tx.Unscoped().Where("store_id = ?", storeID).Order("id ASC").Find(&products).Error; err != nil {
		return err
	}
	for _, product := range products {
		stock := inventory.ProductStock{ProductID: product.ID, OutletID: outletID, Qty: product.Stock}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&stock).Error; err != nil {
			return err
		}
	}
	var variants []catalog.ProductVariant
	if err := tx.Unscoped().Where("store_id = ?", storeID).Order("id ASC").Find(&variants).Error; err != nil {
		return err
	}
	for _, variant := range variants {
		variantID := variant.ID
		stock := inventory.ProductStock{ProductID: variant.ProductID, VariantID: &variantID, OutletID: outletID, Qty: variant.Stock}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&stock).Error; err != nil {
			return err
		}
	}
	return nil
}

// BackfillDefaultOutlets adalah alias ekspor dari backfillDefaultOutlets
// untuk suite integrasi di test/integration; ia memanggil fungsi yang sama.
func BackfillDefaultOutlets(db *gorm.DB) error { return backfillDefaultOutlets(db) }
