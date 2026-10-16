package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/adrr-dev/openPOS/backend/model"
)

var (
	ErrStockOpnameActive = errors.New("sesi stock opname aktif sudah ada")
	ErrStockOpnameState  = errors.New("status sesi stock opname tidak mengizinkan aksi ini")
	ErrStockOpnameItem   = errors.New("item tidak sesuai dengan sesi stock opname")
	ErrStockOpnameTarget = errors.New("produk atau varian tujuan stok tidak tersedia")
)

type StockOpnameItemCount struct {
	ItemID       uint
	CountedStock int
}

type StockOpnameRepo struct {
	db *gorm.DB
}

func NewStockOpnameRepo(db *gorm.DB) *StockOpnameRepo { return &StockOpnameRepo{db: db} }

func (r *StockOpnameRepo) List(ctx context.Context, storeID, outletID uint) ([]model.StockOpnameSession, error) {
	items := make([]model.StockOpnameSession, 0)
	err := r.db.WithContext(ctx).Where("store_id = ? AND outlet_id = ?", storeID, outletID).Order("created_at DESC, id DESC").Find(&items).Error
	return items, err
}

func (r *StockOpnameRepo) GetByID(ctx context.Context, storeID, outletID, id uint) (*model.StockOpnameSession, error) {
	var session model.StockOpnameSession
	if err := r.db.WithContext(ctx).Where("id = ? AND store_id = ? AND outlet_id = ?", id, storeID, outletID).First(&session).Error; err != nil {
		return nil, mapDBErr(err)
	}
	if err := r.loadItems(ctx, storeID, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *StockOpnameRepo) Create(ctx context.Context, storeID, outletID uint, note, createdBy string) (*model.StockOpnameSession, error) {
	var sessionID uint
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var store model.Store
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&store, storeID).Error; err != nil {
			return mapDBErr(err)
		}
		var activeCount int64
		if err := tx.Model(&model.StockOpnameSession{}).
			Where("store_id = ? AND status = ?", storeID, model.StockOpnameCounting).Count(&activeCount).Error; err != nil {
			return err
		}
		if activeCount > 0 {
			return ErrStockOpnameActive
		}

		session := model.StockOpnameSession{
			StoreID: storeID, OutletID: uintPointer(outletID), Status: model.StockOpnameCounting,
			Note: note, CreatedBy: createdBy,
		}
		if err := tx.Create(&session).Error; err != nil {
			return err
		}
		sessionID = session.ID

		products := make([]model.Product, 0)
		if err := tx.Where("store_id = ? AND active = ?", storeID, true).Order("id ASC").Find(&products).Error; err != nil {
			return err
		}
		productIDs := make([]uint, len(products))
		for i := range products {
			productIDs[i] = products[i].ID
		}
		variants := make([]model.ProductVariant, 0)
		if len(productIDs) > 0 {
			if err := tx.Where("store_id = ? AND product_id IN ?", storeID, productIDs).Order("id ASC").Find(&variants).Error; err != nil {
				return err
			}
		}
		variantsByProduct := make(map[uint][]model.ProductVariant, len(products))
		for _, variant := range variants {
			variantsByProduct[variant.ProductID] = append(variantsByProduct[variant.ProductID], variant)
		}
		items := make([]model.StockOpnameItem, 0, len(products)+len(variants))
		for _, product := range products {
			productVariants := variantsByProduct[product.ID]
			if len(productVariants) == 0 {
				stock, err := getOutletStock(tx, outletID, product.ID, nil)
				if err != nil {
					return err
				}
				items = append(items, model.StockOpnameItem{
					SessionID: session.ID, ProductID: product.ID, SystemStock: stock,
				})
				continue
			}
			for _, variant := range productVariants {
				variantID := variant.ID
				stock, err := getOutletStock(tx, outletID, product.ID, &variantID)
				if err != nil {
					return err
				}
				items = append(items, model.StockOpnameItem{
					SessionID: session.ID, ProductID: product.ID, VariantID: &variantID, SystemStock: stock,
				})
			}
		}
		if len(items) > 0 {
			return tx.Create(&items).Error
		}
		return nil
	})
	if err != nil {
		err = mapDBErr(err)
		if errors.Is(err, ErrDuplicate) {
			return nil, ErrStockOpnameActive
		}
		return nil, err
	}
	return r.GetByID(ctx, storeID, outletID, sessionID)
}

func (r *StockOpnameRepo) UpdateItems(ctx context.Context, storeID, outletID, id uint, items []StockOpnameItemCount) (*model.StockOpnameSession, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		session, err := lockStockOpname(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if session.Status != model.StockOpnameCounting {
			return ErrStockOpnameState
		}
		itemIDs := make([]uint, len(items))
		for i := range items {
			itemIDs[i] = items[i].ItemID
		}
		var storedItems []model.StockOpnameItem
		if err := tx.Where("session_id = ? AND id IN ?", session.ID, itemIDs).Find(&storedItems).Error; err != nil {
			return err
		}
		if len(storedItems) != len(items) {
			return ErrStockOpnameItem
		}
		for _, item := range items {
			result := tx.Model(&model.StockOpnameItem{}).
				Where("id = ? AND session_id = ?", item.ItemID, session.ID).
				Update("counted_stock", item.CountedStock)
			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return ErrStockOpnameItem
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, storeID, outletID, id)
}

func (r *StockOpnameRepo) Post(ctx context.Context, storeID, outletID, id uint, actor string) (*model.StockOpnameSession, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		session, err := lockStockOpname(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if session.Status != model.StockOpnameCounting {
			return ErrStockOpnameState
		}

		var items []model.StockOpnameItem
		if err := tx.Where("session_id = ?", session.ID).Order("id ASC").Find(&items).Error; err != nil {
			return err
		}
		for _, item := range items {
			if item.CountedStock == nil {
				continue
			}
			difference := *item.CountedStock - item.SystemStock
			if difference == 0 {
				continue
			}

			var productCount int64
			if err := tx.Model(&model.Product{}).Where("id = ? AND store_id = ?", item.ProductID, storeID).Count(&productCount).Error; err != nil {
				return err
			}
			if productCount == 0 {
				return ErrStockOpnameTarget
			}

			if item.VariantID != nil {
				var variantCount int64
				if err := tx.Model(&model.ProductVariant{}).Where("id = ? AND product_id = ? AND store_id = ?", *item.VariantID, item.ProductID, storeID).Count(&variantCount).Error; err != nil {
					return err
				}
				if variantCount == 0 {
					return ErrStockOpnameTarget
				}
				if err := setOutletStock(tx, storeID, outletID, item.ProductID, item.VariantID, *item.CountedStock); err != nil {
					if errors.Is(err, ErrNotFound) {
						return ErrStockOpnameTarget
					}
					return err
				}
			} else {
				var variantCount int64
				if err := tx.Model(&model.ProductVariant{}).
					Where("store_id = ? AND product_id = ?", storeID, item.ProductID).
					Count(&variantCount).Error; err != nil {
					return err
				}
				if variantCount > 0 {
					return ErrStockOpnameTarget
				}
				if err := setOutletStock(tx, storeID, outletID, item.ProductID, nil, *item.CountedStock); err != nil {
					if errors.Is(err, ErrNotFound) {
						return ErrStockOpnameTarget
					}
					return err
				}
			}

			movement := model.Movement{
				StoreID: storeID, OutletID: uintPointer(outletID), ProductID: item.ProductID, VariantID: item.VariantID,
				Type: model.MovementOpname, Qty: difference,
				Reason: fmt.Sprintf("Opname #%d", session.ID), Actor: actor,
			}
			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
		}
		result := tx.Model(session).Where("status = ?", model.StockOpnameCounting).Update("status", model.StockOpnamePosted)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrStockOpnameState
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, storeID, outletID, id)
}

func (r *StockOpnameRepo) Cancel(ctx context.Context, storeID, outletID, id uint) (*model.StockOpnameSession, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		session, err := lockStockOpname(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if session.Status != model.StockOpnameCounting {
			return ErrStockOpnameState
		}
		result := tx.Model(session).Where("status = ?", model.StockOpnameCounting).Update("status", model.StockOpnameCancelled)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrStockOpnameState
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, storeID, outletID, id)
}

func (r *StockOpnameRepo) loadItems(ctx context.Context, storeID uint, session *model.StockOpnameSession) error {
	session.Items = make([]model.StockOpnameItem, 0)
	if err := r.db.WithContext(ctx).Where("session_id = ?", session.ID).Order("id ASC").Find(&session.Items).Error; err != nil {
		return err
	}
	productIDs := make([]uint, 0, len(session.Items))
	variantIDs := make([]uint, 0, len(session.Items))
	productSeen := make(map[uint]bool, len(session.Items))
	for _, item := range session.Items {
		if !productSeen[item.ProductID] {
			productIDs = append(productIDs, item.ProductID)
			productSeen[item.ProductID] = true
		}
		if item.VariantID != nil {
			variantIDs = append(variantIDs, *item.VariantID)
		}
	}
	products := make([]model.Product, 0, len(productIDs))
	if len(productIDs) > 0 {
		if err := r.db.WithContext(ctx).Unscoped().Where("store_id = ? AND id IN ?", storeID, productIDs).Find(&products).Error; err != nil {
			return err
		}
	}
	productNames := make(map[uint]string, len(products))
	for _, product := range products {
		productNames[product.ID] = product.Name
	}
	variants := make([]model.ProductVariant, 0, len(variantIDs))
	if len(variantIDs) > 0 {
		if err := r.db.WithContext(ctx).Unscoped().Where("store_id = ? AND id IN ?", storeID, variantIDs).Find(&variants).Error; err != nil {
			return err
		}
	}
	variantNames := make(map[uint]string, len(variants))
	for _, variant := range variants {
		variantNames[variant.ID] = variant.Name
	}
	for i := range session.Items {
		item := &session.Items[i]
		item.ProductName = productNames[item.ProductID]
		if item.VariantID != nil {
			item.VariantName = variantNames[*item.VariantID]
		}
		if item.CountedStock != nil {
			difference := *item.CountedStock - item.SystemStock
			item.Difference = &difference
		}
	}
	return nil
}

func lockStockOpname(tx *gorm.DB, storeID, outletID, id uint) (*model.StockOpnameSession, error) {
	var session model.StockOpnameSession
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND store_id = ? AND outlet_id = ?", id, storeID, outletID).First(&session).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &session, nil
}
