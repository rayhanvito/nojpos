package repo

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/model"
)

var ErrStockInsufficient = errors.New("stok tidak cukup")

type ProductFilter struct {
	Q          string
	CategoryID uint
	Active     *bool
	Page       int
	Limit      int
}

type ProductPage struct {
	Items []*model.Product `json:"items"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

type ProductRepo struct {
	db *gorm.DB
}

func NewProductRepo(db *gorm.DB) *ProductRepo { return &ProductRepo{db: db} }

func (r *ProductRepo) GetByID(ctx context.Context, storeID, id uint) (*model.Product, error) {
	return r.getByID(ctx, storeID, nil, id)
}

func (r *ProductRepo) GetByIDAtOutlet(ctx context.Context, storeID, outletID, id uint) (*model.Product, error) {
	return r.getByID(ctx, storeID, &outletID, id)
}

func (r *ProductRepo) getByID(ctx context.Context, storeID uint, outletID *uint, id uint) (*model.Product, error) {
	var p model.Product
	var cat model.Category

	err := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&p).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	if p.CategoryID != nil {
		if err := r.db.WithContext(ctx).First(&cat, "id = ?", *p.CategoryID).Error; err == nil {
			p.CategoryName = &cat.Name
		}
	}
	if err := r.attachInventoryAtOutlet(ctx, storeID, outletID, []*model.Product{&p}); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) List(ctx context.Context, storeID uint, f ProductFilter) (*ProductPage, error) {
	return r.list(ctx, storeID, nil, f)
}

func (r *ProductRepo) ListAtOutlet(ctx context.Context, storeID, outletID uint, f ProductFilter) (*ProductPage, error) {
	return r.list(ctx, storeID, &outletID, f)
}

func (r *ProductRepo) list(ctx context.Context, storeID uint, outletID *uint, f ProductFilter) (*ProductPage, error) {
	page, limit := f.Page, f.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	query := r.db.WithContext(ctx).Model(&model.Product{}).Where("store_id = ?", storeID)

	if q := strings.TrimSpace(f.Q); q != "" {
		searchTerm := "%" + strings.ToLower(q) + "%"
		query = query.Where("lower(name) LIKE ? OR lower(sku) LIKE ? OR lower(barcode) LIKE ?", searchTerm, searchTerm, searchTerm)
	}
	if f.CategoryID != 0 {
		query = query.Where("category_id = ?", f.CategoryID)
	}
	if f.Active != nil {
		query = query.Where("active = ?", *f.Active)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	products := make([]*model.Product, 0)
	offset := (page - 1) * limit
	err := query.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&products).Error
	if err != nil {
		return nil, err
	}

	// Attach category names
	for _, p := range products {
		if p.CategoryID != nil {
			var cat model.Category
			if err := r.db.WithContext(ctx).First(&cat, "id = ?", *p.CategoryID).Error; err == nil {
				p.CategoryName = &cat.Name
			}
		}
	}
	if err := r.attachInventoryAtOutlet(ctx, storeID, outletID, products); err != nil {
		return nil, err
	}

	return &ProductPage{Items: products, Total: int(total), Page: page, Limit: limit}, nil
}

func (r *ProductRepo) ListForExport(ctx context.Context, storeID uint) ([]*model.Product, error) {
	return r.listForExport(ctx, storeID, nil)
}

func (r *ProductRepo) ListForExportAtOutlet(ctx context.Context, storeID, outletID uint) ([]*model.Product, error) {
	return r.listForExport(ctx, storeID, &outletID)
}

func (r *ProductRepo) listForExport(ctx context.Context, storeID uint, outletID *uint) ([]*model.Product, error) {
	products := make([]*model.Product, 0)
	if err := r.db.WithContext(ctx).Where("store_id = ?", storeID).Order("created_at DESC, id DESC").Find(&products).Error; err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return products, nil
	}
	var categories []model.Category
	if err := r.db.WithContext(ctx).Where("store_id = ?", storeID).Find(&categories).Error; err != nil {
		return nil, err
	}
	categoryNames := make(map[uint]string, len(categories))
	for _, category := range categories {
		categoryNames[category.ID] = category.Name
	}
	for _, product := range products {
		if product.CategoryID != nil {
			if name, ok := categoryNames[*product.CategoryID]; ok {
				product.CategoryName = &name
			}
		}
	}
	if outletID != nil {
		if err := r.attachOutletStocks(ctx, *outletID, products); err != nil {
			return nil, err
		}
	}
	return products, nil
}

func (r *ProductRepo) Create(ctx context.Context, storeID uint, p *model.Product) (*model.Product, error) {
	p.StoreID = storeID
	initialStock := p.Stock
	var outletID uint
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var outlet model.Outlet
		if err := tx.Where("store_id = ? AND is_default = ?", storeID, true).First(&outlet).Error; err != nil {
			return mapDBErr(err)
		}
		outletID = outlet.ID
		p.OutletID = &outletID
		if err := tx.Create(p).Error; err != nil {
			return mapDBErr(err)
		}
		return setOutletStock(tx, storeID, outletID, p.ID, nil, initialStock)
	})
	if err != nil {
		return nil, err
	}
	return r.GetByIDAtOutlet(ctx, storeID, outletID, p.ID)
}

func (r *ProductRepo) Update(ctx context.Context, storeID, id uint, p *model.Product) (*model.Product, error) {
	updates := map[string]interface{}{
		"category_id": p.CategoryID,
		"name":        p.Name,
		"sku":         p.SKU,
		"barcode":     p.Barcode,
		"buy_price":   p.BuyPrice,
		"sell_price":  p.SellPrice,
		"unit":        p.Unit,
	}
	if p.MinStock >= 0 {
		updates["min_stock"] = p.MinStock
	}
	res := r.db.WithContext(ctx).Model(&model.Product{}).Where("id = ? AND store_id = ?", id, storeID).Updates(updates)
	if res.Error != nil {
		return nil, mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, storeID, id)
}

func (r *ProductRepo) SetActive(ctx context.Context, storeID, id uint, active bool) error {
	res := r.db.WithContext(ctx).Model(&model.Product{}).Where("id = ? AND store_id = ?", id, storeID).Update("active", active)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProductRepo) IsSkuTaken(ctx context.Context, storeID uint, sku string, exceptProductID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Product{}).Where("store_id = ? AND lower(sku) = lower(?) AND id <> ?", storeID, strings.TrimSpace(sku), exceptProductID).Count(&count).Error
	if err != nil || count > 0 {
		return count > 0, err
	}
	err = r.db.WithContext(ctx).Model(&model.ProductVariant{}).Where("store_id = ? AND lower(sku) = lower(?)", storeID, strings.TrimSpace(sku)).Count(&count).Error
	return count > 0, err
}

func (r *ProductRepo) CreateWithInitialMovement(ctx context.Context, storeID uint, p *model.Product, actor string) (*model.Product, error) {
	return r.createWithInitialMovement(ctx, storeID, 0, p, actor)
}

func (r *ProductRepo) CreateWithInitialMovementAtOutlet(ctx context.Context, storeID, outletID uint, p *model.Product, actor string) (*model.Product, error) {
	return r.createWithInitialMovement(ctx, storeID, outletID, p, actor)
}

func (r *ProductRepo) createWithInitialMovement(ctx context.Context, storeID, outletID uint, p *model.Product, actor string) (*model.Product, error) {
	p.StoreID = storeID
	if outletID == 0 {
		var outlet model.Outlet
		if err := r.db.WithContext(ctx).Where("store_id = ? AND is_default = ?", storeID, true).First(&outlet).Error; err != nil {
			return nil, mapDBErr(err)
		}
		outletID = outlet.ID
	}
	initialStock := p.Stock
	p.OutletID = &outletID
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		isDefault, err := isDefaultOutlet(tx, storeID, outletID)
		if err != nil {
			return err
		}
		if !isDefault {
			p.Stock = 0
		}
		if err := tx.Create(p).Error; err != nil {
			return mapDBErr(err)
		}
		if err := setOutletStock(tx, storeID, outletID, p.ID, nil, initialStock); err != nil {
			return err
		}
		if initialStock > 0 {
			mv := model.Movement{
				StoreID:   storeID,
				OutletID:  &outletID,
				ProductID: p.ID,
				Type:      model.MovementInitial,
				Qty:       initialStock,
				Reason:    "Stok awal",
				Actor:     actor,
			}
			if err := tx.Create(&mv).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return r.GetByIDAtOutlet(ctx, storeID, outletID, p.ID)
}

func (r *ProductRepo) AdjustStock(ctx context.Context, storeID, productID uint, delta int, reason, actor string) (*model.Product, error) {
	return r.adjustStock(ctx, storeID, 0, productID, delta, reason, actor)
}

func (r *ProductRepo) AdjustStockAtOutlet(ctx context.Context, storeID, outletID, productID uint, delta int, reason, actor string) (*model.Product, error) {
	return r.adjustStock(ctx, storeID, outletID, productID, delta, reason, actor)
}

func (r *ProductRepo) adjustStock(ctx context.Context, storeID, outletID, productID uint, delta int, reason, actor string) (*model.Product, error) {
	if outletID == 0 {
		var outlet model.Outlet
		if err := r.db.WithContext(ctx).Where("store_id = ? AND is_default = ?", storeID, true).First(&outlet).Error; err != nil {
			return nil, mapDBErr(err)
		}
		outletID = outlet.ID
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var variantCount int64
		if err := tx.Model(&model.ProductVariant{}).Where("store_id = ? AND product_id = ?", storeID, productID).Count(&variantCount).Error; err != nil {
			return err
		}
		if variantCount > 0 {
			return ErrVariantRequired
		}
		var exists int64
		if err := tx.Model(&model.Product{}).Where("id = ? AND store_id = ?", productID, storeID).Count(&exists).Error; err != nil {
			return err
		}
		if exists == 0 {
			return ErrNotFound
		}
		if _, _, _, err := changeOutletStock(tx, storeID, outletID, productID, nil, delta, false); err != nil {
			return err
		}

		mv := model.Movement{
			StoreID:   storeID,
			OutletID:  &outletID,
			ProductID: productID,
			Type:      model.MovementAdjust,
			Qty:       delta,
			Reason:    reason,
			Actor:     actor,
		}
		return tx.Create(&mv).Error
	})
	if err != nil {
		return nil, err
	}
	return r.GetByIDAtOutlet(ctx, storeID, outletID, productID)
}

func (r *ProductRepo) Delete(ctx context.Context, storeID, id uint) error {
	res := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", id, storeID).Delete(&model.Product{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
