package repo

import (
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/adrr-dev/openPOS/backend/model"
)

func stockQuery(tx *gorm.DB, outletID, productID uint, variantID *uint) *gorm.DB {
	query := tx.Model(&model.ProductStock{}).Where("outlet_id = ? AND product_id = ?", outletID, productID)
	if variantID == nil {
		return query.Where("variant_id IS NULL")
	}
	return query.Where("variant_id = ?", *variantID)
}

func getOutletStock(tx *gorm.DB, outletID, productID uint, variantID *uint) (int, error) {
	var stock model.ProductStock
	err := stockQuery(tx, outletID, productID, variantID).First(&stock).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return stock.Qty, err
}

func setOutletStock(tx *gorm.DB, storeID, outletID, productID uint, variantID *uint, qty int) error {
	if qty < 0 {
		return ErrStockInsufficient
	}
	if err := ensureOutletStock(tx, outletID, productID, variantID); err != nil {
		return err
	}
	result := stockQuery(tx, outletID, productID, variantID).Update("qty", qty)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return mirrorDefaultStock(tx, storeID, outletID, productID, variantID, qty)
}

func changeOutletStock(tx *gorm.DB, storeID, outletID, productID uint, variantID *uint, delta int, allowShortage bool) (before, after, actualDelta int, err error) {
	if err = ensureOutletStock(tx, outletID, productID, variantID); err != nil {
		return
	}
	query := stockQuery(tx, outletID, productID, variantID)
	if tx.Dialector.Name() != "sqlite" {
		query = query.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	var stock model.ProductStock
	if err = query.First(&stock).Error; err != nil {
		return
	}
	before = stock.Qty
	after = before + delta
	actualDelta = delta
	if after < 0 {
		if !allowShortage {
			err = ErrStockInsufficient
			return
		}
		after = 0
		actualDelta = -before
	}
	result := stockQuery(tx, outletID, productID, variantID).Update("qty", after)
	if result.Error != nil {
		err = result.Error
		return
	}
	if result.RowsAffected == 0 {
		err = ErrNotFound
		return
	}
	err = mirrorDefaultStock(tx, storeID, outletID, productID, variantID, after)
	return
}

func ensureOutletStock(tx *gorm.DB, outletID, productID uint, variantID *uint) error {
	var count int64
	if err := stockQuery(tx, outletID, productID, variantID).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	stock := model.ProductStock{ProductID: productID, VariantID: variantID, OutletID: outletID}
	return tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&stock).Error
}

func isDefaultOutlet(tx *gorm.DB, storeID, outletID uint) (bool, error) {
	var count int64
	err := tx.Model(&model.Outlet{}).
		Where("id = ? AND store_id = ? AND is_default = ?", outletID, storeID, true).
		Count(&count).Error
	return count > 0, err
}

func mirrorDefaultStock(tx *gorm.DB, storeID, outletID, productID uint, variantID *uint, qty int) error {
	var count int64
	if err := tx.Model(&model.Outlet{}).Where("id = ? AND store_id = ? AND is_default = ?", outletID, storeID, true).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	if variantID == nil {
		result := tx.Unscoped().Model(&model.Product{}).Where("id = ? AND store_id = ?", productID, storeID).Update("stock", qty)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	}
	result := tx.Unscoped().Model(&model.ProductVariant{}).
		Where("id = ? AND product_id = ? AND store_id = ?", *variantID, productID, storeID).
		Update("stock", qty)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
