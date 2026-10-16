package repo

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/rayhanvito/nojpos/backend/model"
)

var ErrOutletDefault = errors.New("outlet default tidak dapat dinonaktifkan")
var ErrOutletNotFound = errors.New("outlet tidak ditemukan di toko")

type OutletRepo struct {
	db *gorm.DB
}

func NewOutletRepo(db *gorm.DB) *OutletRepo { return &OutletRepo{db: db} }

func (r *OutletRepo) List(ctx context.Context, storeID uint) ([]model.Outlet, error) {
	items := make([]model.Outlet, 0)
	err := r.db.WithContext(ctx).Where("store_id = ?", storeID).Order("is_default DESC, id ASC").Find(&items).Error
	return items, err
}

func (r *OutletRepo) GetByID(ctx context.Context, storeID, id uint) (*model.Outlet, error) {
	var outlet model.Outlet
	if err := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&outlet).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &outlet, nil
}

func (r *OutletRepo) GetDefault(ctx context.Context, storeID uint) (*model.Outlet, error) {
	var outlet model.Outlet
	if err := r.db.WithContext(ctx).Where("store_id = ? AND is_default = ?", storeID, true).First(&outlet).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &outlet, nil
}

func (r *OutletRepo) Create(ctx context.Context, storeID uint, outlet *model.Outlet) (*model.Outlet, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var store model.Store
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&store, storeID).Error; err != nil {
			return mapDBErr(err)
		}
		var defaults int64
		if err := tx.Model(&model.Outlet{}).Where("store_id = ? AND is_default = ?", storeID, true).Count(&defaults).Error; err != nil {
			return err
		}
		if outlet.IsDefault || defaults == 0 {
			if err := tx.Model(&model.Outlet{}).Where("store_id = ?", storeID).Update("is_default", false).Error; err != nil {
				return err
			}
			outlet.IsDefault = true
		}
		outlet.StoreID = storeID
		if err := tx.Create(outlet).Error; err != nil {
			return err
		}
		if outlet.IsDefault {
			return mirrorOutletStocksToLegacy(tx, storeID, outlet.ID)
		}
		return nil
	})
	if err != nil {
		return nil, mapDBErr(err)
	}
	return r.GetByID(ctx, storeID, outlet.ID)
}

func (r *OutletRepo) Update(ctx context.Context, storeID, id uint, updated model.Outlet, isDefault *bool) (*model.Outlet, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var store model.Store
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&store, storeID).Error; err != nil {
			return mapDBErr(err)
		}
		var current model.Outlet
		if err := tx.Where("id = ? AND store_id = ?", id, storeID).First(&current).Error; err != nil {
			return mapDBErr(err)
		}
		if isDefault != nil && current.IsDefault && !*isDefault {
			return ErrOutletDefault
		}
		if isDefault != nil && *isDefault {
			if err := tx.Model(&model.Outlet{}).Where("store_id = ?", storeID).Update("is_default", false).Error; err != nil {
				return err
			}
		}
		updates := map[string]any{"name": updated.Name, "address": updated.Address}
		if isDefault != nil {
			updates["is_default"] = *isDefault
		}
		result := tx.Model(&model.Outlet{}).Where("id = ? AND store_id = ?", id, storeID).Updates(updates)
		if result.Error != nil {
			return mapDBErr(result.Error)
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		if isDefault != nil && *isDefault {
			return mirrorOutletStocksToLegacy(tx, storeID, id)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, storeID, id)
}

func mirrorOutletStocksToLegacy(tx *gorm.DB, storeID, outletID uint) error {
	if err := tx.Unscoped().Model(&model.Product{}).
		Where("store_id = ?", storeID).
		Update("stock", gorm.Expr("COALESCE((SELECT qty FROM product_stocks WHERE product_stocks.product_id = products.id AND product_stocks.outlet_id = ? AND product_stocks.variant_id IS NULL AND product_stocks.deleted_at IS NULL LIMIT 1), 0)", outletID)).Error; err != nil {
		return err
	}
	return tx.Unscoped().Model(&model.ProductVariant{}).
		Where("store_id = ?", storeID).
		Update("stock", gorm.Expr("COALESCE((SELECT qty FROM product_stocks WHERE product_stocks.product_id = product_variants.product_id AND product_stocks.variant_id = product_variants.id AND product_stocks.outlet_id = ? AND product_stocks.deleted_at IS NULL LIMIT 1), 0)", outletID)).Error
}
