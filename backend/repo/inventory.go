package repo

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/model"
)

func (r *ProductRepo) attachInventory(ctx context.Context, storeID uint, products []*model.Product) error {
	if len(products) == 0 {
		return nil
	}
	productIDs := make([]uint, len(products))
	for i, product := range products {
		productIDs[i] = product.ID
		product.Variants = []model.ProductVariant{}
		product.PriceTiers = []model.PriceTier{}
	}

	var productTiers []model.PriceTier
	if err := r.db.WithContext(ctx).Where("store_id = ? AND product_id IN ? AND variant_id IS NULL", storeID, productIDs).Order("min_qty ASC, id ASC").Find(&productTiers).Error; err != nil {
		return err
	}
	var variants []model.ProductVariant
	if err := r.db.WithContext(ctx).Where("store_id = ? AND product_id IN ?", storeID, productIDs).Order("id ASC").Find(&variants).Error; err != nil {
		return err
	}
	variantIDs := make([]uint, len(variants))
	variantByID := make(map[uint]*model.ProductVariant, len(variants))
	for i := range variants {
		variants[i].PriceTiers = []model.PriceTier{}
		variantIDs[i] = variants[i].ID
		variantByID[variants[i].ID] = &variants[i]
	}
	var variantTiers []model.PriceTier
	if len(variantIDs) > 0 {
		if err := r.db.WithContext(ctx).Where("store_id = ? AND variant_id IN ?", storeID, variantIDs).Order("min_qty ASC, id ASC").Find(&variantTiers).Error; err != nil {
			return err
		}
	}
	for _, tier := range productTiers {
		for _, product := range products {
			if product.ID == tier.ProductID {
				product.PriceTiers = append(product.PriceTiers, tier)
				break
			}
		}
	}
	for _, tier := range variantTiers {
		if variant := variantByID[*tier.VariantID]; variant != nil {
			variant.PriceTiers = append(variant.PriceTiers, tier)
		}
	}
	for _, variant := range variants {
		for _, product := range products {
			if product.ID == variant.ProductID {
				product.Variants = append(product.Variants, variant)
				break
			}
		}
	}
	return nil
}

func (r *ProductRepo) attachInventoryAtOutlet(ctx context.Context, storeID uint, outletID *uint, products []*model.Product) error {
	if err := r.attachInventory(ctx, storeID, products); err != nil || outletID == nil || len(products) == 0 {
		return err
	}
	return r.attachOutletStocks(ctx, *outletID, products)
}

func (r *ProductRepo) attachOutletStocks(ctx context.Context, outletID uint, products []*model.Product) error {
	if len(products) == 0 {
		return nil
	}
	productIDs := make([]uint, 0, len(products))
	productByID := make(map[uint]*model.Product, len(products))
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
		productByID[product.ID] = product
	}
	var stocks []model.ProductStock
	if err := r.db.WithContext(ctx).Where("outlet_id = ? AND product_id IN ?", outletID, productIDs).Find(&stocks).Error; err != nil {
		return err
	}
	variantByID := make(map[uint]*model.ProductVariant)
	for _, product := range products {
		product.Stock = 0
		for i := range product.Variants {
			product.Variants[i].Stock = 0
			variantByID[product.Variants[i].ID] = &product.Variants[i]
		}
	}
	for _, stock := range stocks {
		product := productByID[stock.ProductID]
		if product == nil {
			continue
		}
		if stock.VariantID == nil {
			product.Stock = stock.Qty
			continue
		}
		if variant := variantByID[*stock.VariantID]; variant != nil {
			variant.Stock = stock.Qty
		}
	}
	return nil
}

func (r *ProductRepo) ListVariants(ctx context.Context, storeID, productID uint) ([]model.ProductVariant, error) {
	product, err := r.GetByID(ctx, storeID, productID)
	if err != nil {
		return nil, err
	}
	return product.Variants, nil
}

func (r *ProductRepo) ListVariantsAtOutlet(ctx context.Context, storeID, outletID, productID uint) ([]model.ProductVariant, error) {
	product, err := r.GetByIDAtOutlet(ctx, storeID, outletID, productID)
	if err != nil {
		return nil, err
	}
	return product.Variants, nil
}

func (r *ProductRepo) GetVariant(ctx context.Context, storeID, productID, variantID uint) (*model.ProductVariant, error) {
	product, err := r.GetByID(ctx, storeID, productID)
	if err != nil {
		return nil, err
	}
	for i := range product.Variants {
		if product.Variants[i].ID == variantID {
			return &product.Variants[i], nil
		}
	}
	return nil, ErrNotFound
}

func (r *ProductRepo) GetVariantAtOutlet(ctx context.Context, storeID, outletID, productID, variantID uint) (*model.ProductVariant, error) {
	product, err := r.GetByIDAtOutlet(ctx, storeID, outletID, productID)
	if err != nil {
		return nil, err
	}
	for i := range product.Variants {
		if product.Variants[i].ID == variantID {
			return &product.Variants[i], nil
		}
	}
	return nil, ErrNotFound
}

func (r *ProductRepo) CreateVariant(ctx context.Context, storeID, productID uint, variant *model.ProductVariant, actor string) (*model.ProductVariant, error) {
	return r.CreateVariantAtOutlet(ctx, storeID, 0, productID, variant, actor)
}

func (r *ProductRepo) CreateVariantAtOutlet(ctx context.Context, storeID, outletID, productID uint, variant *model.ProductVariant, actor string) (*model.ProductVariant, error) {
	if outletID == 0 {
		outlet, err := r.defaultOutlet(ctx, storeID)
		if err != nil {
			return nil, err
		}
		outletID = outlet.ID
	}
	if _, err := r.GetByIDAtOutlet(ctx, storeID, outletID, productID); err != nil {
		return nil, err
	}
	variant.StoreID = storeID
	variant.ProductID = productID
	variant.OutletID = &outletID
	initialStock := variant.Stock
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		isDefault, err := isDefaultOutlet(tx, storeID, outletID)
		if err != nil {
			return err
		}
		if !isDefault {
			variant.Stock = 0
		}
		if err := tx.Create(variant).Error; err != nil {
			return mapDBErr(err)
		}
		variantID := variant.ID
		if err := setOutletStock(tx, storeID, outletID, productID, &variantID, initialStock); err != nil {
			return err
		}
		if initialStock > 0 {
			movement := model.Movement{StoreID: storeID, OutletID: &outletID, ProductID: productID, VariantID: &variant.ID, Type: model.MovementInitial, Qty: initialStock, Reason: "Stok awal varian", Actor: actor}
			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetVariantAtOutlet(ctx, storeID, outletID, productID, variant.ID)
}

func (r *ProductRepo) IsVariantSkuTaken(ctx context.Context, storeID uint, sku string, exceptID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.ProductVariant{}).Where("store_id = ? AND lower(sku) = lower(?) AND id <> ?", storeID, strings.TrimSpace(sku), exceptID).Count(&count).Error
	if err != nil || count > 0 {
		return count > 0, err
	}
	err = r.db.WithContext(ctx).Model(&model.Product{}).Where("store_id = ? AND lower(sku) = lower(?)", storeID, strings.TrimSpace(sku)).Count(&count).Error
	return count > 0, err
}

func (r *ProductRepo) UpdateVariant(ctx context.Context, storeID, productID, variantID uint, variant *model.ProductVariant, actor string) (*model.ProductVariant, error) {
	return r.UpdateVariantAtOutlet(ctx, storeID, 0, productID, variantID, variant, actor)
}

func (r *ProductRepo) UpdateVariantAtOutlet(ctx context.Context, storeID, outletID, productID, variantID uint, variant *model.ProductVariant, actor string) (*model.ProductVariant, error) {
	if outletID == 0 {
		outlet, err := r.defaultOutlet(ctx, storeID)
		if err != nil {
			return nil, err
		}
		outletID = outlet.ID
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current model.ProductVariant
		if err := tx.Where("id = ? AND store_id = ? AND product_id = ?", variantID, storeID, productID).First(&current).Error; err != nil {
			return mapDBErr(err)
		}
		currentStock, err := getOutletStock(tx, outletID, productID, &variantID)
		if err != nil {
			return err
		}
		if err := tx.Model(&current).Updates(map[string]any{
			"name": variant.Name, "sku": variant.SKU, "buy_price": variant.BuyPrice,
			"sell_price": variant.SellPrice, "min_stock": variant.MinStock,
		}).Error; err != nil {
			return mapDBErr(err)
		}
		delta := variant.Stock - currentStock
		if err := setOutletStock(tx, storeID, outletID, productID, &variantID, variant.Stock); err != nil {
			return err
		}
		if delta != 0 {
			movement := model.Movement{StoreID: storeID, OutletID: &outletID, ProductID: productID, VariantID: &variantID, Type: model.MovementAdjust, Qty: delta, Reason: "Perubahan stok varian", Actor: actor}
			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetVariantAtOutlet(ctx, storeID, outletID, productID, variantID)
}

func (r *ProductRepo) AdjustVariantStock(ctx context.Context, storeID, productID, variantID uint, delta int, reason, actor string) (*model.ProductVariant, error) {
	return r.AdjustVariantStockAtOutlet(ctx, storeID, 0, productID, variantID, delta, reason, actor)
}

func (r *ProductRepo) AdjustVariantStockAtOutlet(ctx context.Context, storeID, outletID, productID, variantID uint, delta int, reason, actor string) (*model.ProductVariant, error) {
	if outletID == 0 {
		outlet, err := r.defaultOutlet(ctx, storeID)
		if err != nil {
			return nil, err
		}
		outletID = outlet.ID
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var exists int64
		if err := tx.Model(&model.ProductVariant{}).Where("id = ? AND store_id = ? AND product_id = ?", variantID, storeID, productID).Count(&exists).Error; err != nil {
			return err
		}
		if exists == 0 {
			return ErrNotFound
		}
		if _, _, actualDelta, err := changeOutletStock(tx, storeID, outletID, productID, &variantID, delta, false); err != nil {
			return err
		} else {
			delta = actualDelta
		}
		movement := model.Movement{StoreID: storeID, OutletID: &outletID, ProductID: productID, VariantID: &variantID, Type: model.MovementAdjust, Qty: delta, Reason: reason, Actor: actor}
		return tx.Create(&movement).Error
	})
	if err != nil {
		return nil, err
	}
	return r.GetVariantAtOutlet(ctx, storeID, outletID, productID, variantID)
}

func (r *ProductRepo) defaultOutlet(ctx context.Context, storeID uint) (*model.Outlet, error) {
	var outlet model.Outlet
	if err := r.db.WithContext(ctx).Where("store_id = ? AND is_default = ?", storeID, true).First(&outlet).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &outlet, nil
}

func (r *ProductRepo) DeleteVariant(ctx context.Context, storeID, productID, variantID uint) error {
	res := r.db.WithContext(ctx).Where("id = ? AND store_id = ? AND product_id = ?", variantID, storeID, productID).Delete(&model.ProductVariant{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProductRepo) ListPriceTiers(ctx context.Context, storeID, productID uint, variantID *uint) ([]model.PriceTier, error) {
	if variantID == nil {
		if _, err := r.GetByID(ctx, storeID, productID); err != nil {
			return nil, err
		}
	} else if _, err := r.GetVariant(ctx, storeID, productID, *variantID); err != nil {
		return nil, err
	}
	query := r.db.WithContext(ctx).Where("store_id = ? AND product_id = ?", storeID, productID)
	if variantID == nil {
		query = query.Where("variant_id IS NULL")
	} else {
		query = query.Where("variant_id = ?", *variantID)
	}
	var tiers []model.PriceTier
	if err := query.Order("min_qty ASC, id ASC").Find(&tiers).Error; err != nil {
		return nil, err
	}
	if tiers == nil {
		tiers = []model.PriceTier{}
	}
	return tiers, nil
}

func (r *ProductRepo) CreatePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tier *model.PriceTier) (*model.PriceTier, error) {
	if variantID == nil {
		if _, err := r.GetByID(ctx, storeID, productID); err != nil {
			return nil, err
		}
	} else if _, err := r.GetVariant(ctx, storeID, productID, *variantID); err != nil {
		return nil, err
	}
	query := r.db.WithContext(ctx).Model(&model.PriceTier{}).Where("store_id = ? AND product_id = ? AND min_qty = ?", storeID, productID, tier.MinQty)
	if variantID == nil {
		query = query.Where("variant_id IS NULL")
	} else {
		query = query.Where("variant_id = ?", *variantID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("jumlah minimum harga bertingkat sudah digunakan")
	}
	tier.StoreID, tier.ProductID, tier.VariantID = storeID, productID, variantID
	if err := r.db.WithContext(ctx).Create(tier).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return tier, nil
}

func (r *ProductRepo) UpdatePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tierID uint, tier *model.PriceTier) (*model.PriceTier, error) {
	query := r.db.WithContext(ctx).Model(&model.PriceTier{}).Where("id = ? AND store_id = ? AND product_id = ?", tierID, storeID, productID)
	if variantID == nil {
		query = query.Where("variant_id IS NULL")
	} else {
		query = query.Where("variant_id = ?", *variantID)
	}
	var current model.PriceTier
	if err := query.First(&current).Error; err != nil {
		return nil, mapDBErr(err)
	}
	dup := r.db.WithContext(ctx).Model(&model.PriceTier{}).Where("store_id = ? AND product_id = ? AND min_qty = ? AND id <> ?", storeID, productID, tier.MinQty, tierID)
	if variantID == nil {
		dup = dup.Where("variant_id IS NULL")
	} else {
		dup = dup.Where("variant_id = ?", *variantID)
	}
	var count int64
	if err := dup.Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, fmt.Errorf("jumlah minimum harga bertingkat sudah digunakan")
	}
	update := r.db.WithContext(ctx).Model(&model.PriceTier{}).Where("id = ? AND store_id = ? AND product_id = ?", tierID, storeID, productID)
	if variantID == nil {
		update = update.Where("variant_id IS NULL")
	} else {
		update = update.Where("variant_id = ?", *variantID)
	}
	if err := update.Updates(map[string]any{"min_qty": tier.MinQty, "sell_price": tier.SellPrice}).Error; err != nil {
		return nil, mapDBErr(err)
	}
	current.MinQty, current.SellPrice = tier.MinQty, tier.SellPrice
	return &current, nil
}

func (r *ProductRepo) DeletePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tierID uint) error {
	query := r.db.WithContext(ctx).Where("id = ? AND store_id = ? AND product_id = ?", tierID, storeID, productID)
	if variantID == nil {
		query = query.Where("variant_id IS NULL")
	} else {
		query = query.Where("variant_id = ?", *variantID)
	}
	res := query.Delete(&model.PriceTier{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
