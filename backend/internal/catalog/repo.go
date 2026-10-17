package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// mapDBErr memetakan error database ke sentinel slice, setara mapDBErr
// lama di repo/base.go: record kosong -> ErrNotFound; pelanggaran unik
// (Postgres 23505 / SQLite) -> ErrDuplicate.
func mapDBErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "23505") ||
			strings.Contains(msg, "duplicate key") ||
			strings.Contains(msg, "UNIQUE constraint failed") ||
			strings.Contains(msg, "unique constraint") {
			return ErrDuplicate
		}
	}
	return err
}

func isNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

type CategoryRepo struct {
	database.BaseRepo
}

func NewCategoryRepo(db *gorm.DB) *CategoryRepo {
	return &CategoryRepo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *CategoryRepo) ListByStore(ctx context.Context, storeID uint) ([]*Category, error) {
	out := make([]*Category, 0)
	err := r.DB(ctx).Where("store_id = ?", storeID).Order("created_at ASC").Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *CategoryRepo) GetByID(ctx context.Context, storeID, id uint) (*Category, error) {
	var c Category
	err := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&c).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	return &c, nil
}

func (r *CategoryRepo) Create(ctx context.Context, storeID uint, name string) (*Category, error) {
	c := Category{
		StoreID: storeID,
		Name:    name,
		Active:  true,
	}
	err := r.DB(ctx).Create(&c).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	return &c, nil
}

func (r *CategoryRepo) Delete(ctx context.Context, storeID, id uint) (softDeleted bool, err error) {
	var c Category
	err = r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("id = ? AND store_id = ?", id, storeID).First(&c).Error; err != nil {
			if isNotFound(err) {
				return ErrNotFound
			}
			return err
		}

		var prodCount int64
		if err := tx.Model(&Product{}).Where("category_id = ? AND store_id = ?", id, storeID).Count(&prodCount).Error; err != nil {
			return err
		}

		if prodCount > 0 {
			softDeleted = true
			return tx.Model(&c).Update("active", false).Error
		}

		// gorm.Model soft-deletes by default; empty categories were hard
		// deleted before, so bypass the soft delete here.
		return tx.Unscoped().Delete(&c).Error
	})
	return softDeleted, mapDBErr(err)
}

type ProductFilter struct {
	Q          string
	CategoryID uint
	Active     *bool
	Page       int
	Limit      int
}

type ProductPage struct {
	Items []*Product `json:"items"`
	Total int        `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}

type ProductRepo struct {
	database.BaseRepo
}

func NewProductRepo(db *gorm.DB) *ProductRepo {
	return &ProductRepo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *ProductRepo) GetByIDForStore(ctx context.Context, storeID, id uint) (*Product, error) {
	return r.getByID(ctx, storeID, nil, id)
}

func (r *ProductRepo) GetByIDForOutlet(ctx context.Context, storeID, outletID, id uint) (*Product, error) {
	return r.getByID(ctx, storeID, &outletID, id)
}

func (r *ProductRepo) getByID(ctx context.Context, storeID uint, outletID *uint, id uint) (*Product, error) {
	var p Product
	var cat Category

	err := database.ScopeStore(r.DB(ctx), storeID).Where("id = ?", id).First(&p).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	if p.CategoryID != nil {
		if err := r.DB(ctx).First(&cat, "id = ?", *p.CategoryID).Error; err == nil {
			p.CategoryName = &cat.Name
		}
	}
	if err := r.attachInventoryAtOutlet(ctx, storeID, outletID, []*Product{&p}); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *ProductRepo) ListForStore(ctx context.Context, storeID uint, f ProductFilter) (*ProductPage, error) {
	return r.list(ctx, storeID, nil, f)
}

func (r *ProductRepo) ListForOutlet(ctx context.Context, storeID, outletID uint, f ProductFilter) (*ProductPage, error) {
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

	query := database.ScopeStore(r.DB(ctx).Model(&Product{}), storeID)

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

	products := make([]*Product, 0)
	offset := (page - 1) * limit
	err := query.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&products).Error
	if err != nil {
		return nil, err
	}

	// Attach category names
	for _, p := range products {
		if p.CategoryID != nil {
			var cat Category
			if err := r.DB(ctx).First(&cat, "id = ?", *p.CategoryID).Error; err == nil {
				p.CategoryName = &cat.Name
			}
		}
	}
	if err := r.attachInventoryAtOutlet(ctx, storeID, outletID, products); err != nil {
		return nil, err
	}

	return &ProductPage{Items: products, Total: int(total), Page: page, Limit: limit}, nil
}

func (r *ProductRepo) ListForExport(ctx context.Context, storeID uint) ([]*Product, error) {
	return r.listForExport(ctx, storeID, nil)
}

func (r *ProductRepo) ListForExportAtOutlet(ctx context.Context, storeID, outletID uint) ([]*Product, error) {
	return r.listForExport(ctx, storeID, &outletID)
}

func (r *ProductRepo) listForExport(ctx context.Context, storeID uint, outletID *uint) ([]*Product, error) {
	products := make([]*Product, 0)
	if err := r.DB(ctx).Where("store_id = ?", storeID).Order("created_at DESC, id DESC").Find(&products).Error; err != nil {
		return nil, err
	}
	if len(products) == 0 {
		return products, nil
	}
	var categories []Category
	if err := r.DB(ctx).Where("store_id = ?", storeID).Find(&categories).Error; err != nil {
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

// InsertProduct hanya menyisipkan baris produk (tanpa stok/movement):
// orkestrasi stok awal ada di Service lewat mesin stok inventory (Unit B).
func (r *ProductRepo) InsertProduct(ctx context.Context, storeID, outletID uint, p *Product) error {
	p.StoreID = storeID
	p.OutletID = &outletID
	if err := r.DB(ctx).Create(p).Error; err != nil {
		return mapDBErr(err)
	}
	return nil
}

func (r *ProductRepo) Update(ctx context.Context, storeID, id uint, p *Product) (*Product, error) {
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
	res := r.DB(ctx).Model(&Product{}).Where("id = ? AND store_id = ?", id, storeID).Updates(updates)
	if res.Error != nil {
		return nil, mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByIDForStore(ctx, storeID, id)
}

func (r *ProductRepo) SetActive(ctx context.Context, storeID, id uint, active bool) error {
	res := r.DB(ctx).Model(&Product{}).Where("id = ? AND store_id = ?", id, storeID).Update("active", active)
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
	err := r.DB(ctx).Model(&Product{}).Where("store_id = ? AND lower(sku) = lower(?) AND id <> ?", storeID, strings.TrimSpace(sku), exceptProductID).Count(&count).Error
	if err != nil || count > 0 {
		return count > 0, err
	}
	err = r.DB(ctx).Model(&ProductVariant{}).Where("store_id = ? AND lower(sku) = lower(?)", storeID, strings.TrimSpace(sku)).Count(&count).Error
	return count > 0, err
}

// DefaultOutlet mengembalikan outlet default toko (ErrNotFound bila tidak ada).
func (r *ProductRepo) DefaultOutlet(ctx context.Context, storeID uint) (*outlet.Outlet, error) {
	return r.defaultOutlet(ctx, storeID)
}

func (r *ProductRepo) Delete(ctx context.Context, storeID, id uint) error {
	res := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).Delete(&Product{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProductRepo) attachInventory(ctx context.Context, storeID uint, products []*Product) error {
	if len(products) == 0 {
		return nil
	}
	productIDs := make([]uint, len(products))
	for i, product := range products {
		productIDs[i] = product.ID
		product.Variants = []ProductVariant{}
		product.PriceTiers = []PriceTier{}
	}

	var productTiers []PriceTier
	if err := r.DB(ctx).Where("store_id = ? AND product_id IN ? AND variant_id IS NULL", storeID, productIDs).Order("min_qty ASC, id ASC").Find(&productTiers).Error; err != nil {
		return err
	}
	var variants []ProductVariant
	if err := r.DB(ctx).Where("store_id = ? AND product_id IN ?", storeID, productIDs).Order("id ASC").Find(&variants).Error; err != nil {
		return err
	}
	variantIDs := make([]uint, len(variants))
	variantByID := make(map[uint]*ProductVariant, len(variants))
	for i := range variants {
		variants[i].PriceTiers = []PriceTier{}
		variantIDs[i] = variants[i].ID
		variantByID[variants[i].ID] = &variants[i]
	}
	var variantTiers []PriceTier
	if len(variantIDs) > 0 {
		if err := r.DB(ctx).Where("store_id = ? AND variant_id IN ?", storeID, variantIDs).Order("min_qty ASC, id ASC").Find(&variantTiers).Error; err != nil {
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

func (r *ProductRepo) attachInventoryAtOutlet(ctx context.Context, storeID uint, outletID *uint, products []*Product) error {
	if err := r.attachInventory(ctx, storeID, products); err != nil || outletID == nil || len(products) == 0 {
		return err
	}
	return r.attachOutletStocks(ctx, *outletID, products)
}

func (r *ProductRepo) attachOutletStocks(ctx context.Context, outletID uint, products []*Product) error {
	if len(products) == 0 {
		return nil
	}
	productIDs := make([]uint, 0, len(products))
	productByID := make(map[uint]*Product, len(products))
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
		productByID[product.ID] = product
	}
	// Proyeksi baca-saja atas tabel product_stocks milik inventory
	// (pola Table() yang sama dengan report/repo.go): katalog tidak
	// menyimpan tipe baris stok dan tidak pernah menulis tabel ini.
	type outletStockRow struct {
		ProductID uint  `gorm:"column:product_id"`
		VariantID *uint `gorm:"column:variant_id"`
		Qty       int   `gorm:"column:qty"`
	}
	var stocks []outletStockRow
	if err := r.DB(ctx).Table("product_stocks").Select("product_id, variant_id, qty").Where("outlet_id = ? AND product_id IN ? AND deleted_at IS NULL", outletID, productIDs).Find(&stocks).Error; err != nil {
		return err
	}
	variantByID := make(map[uint]*ProductVariant)
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

func (r *ProductRepo) ListVariantsForStore(ctx context.Context, storeID, productID uint) ([]ProductVariant, error) {
	product, err := r.GetByIDForStore(ctx, storeID, productID)
	if err != nil {
		return nil, err
	}
	return product.Variants, nil
}

func (r *ProductRepo) ListVariantsForOutlet(ctx context.Context, storeID, outletID, productID uint) ([]ProductVariant, error) {
	product, err := r.GetByIDForOutlet(ctx, storeID, outletID, productID)
	if err != nil {
		return nil, err
	}
	return product.Variants, nil
}

func (r *ProductRepo) GetVariant(ctx context.Context, storeID, productID, variantID uint) (*ProductVariant, error) {
	product, err := r.GetByIDForStore(ctx, storeID, productID)
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

func (r *ProductRepo) GetVariantAtOutlet(ctx context.Context, storeID, outletID, productID, variantID uint) (*ProductVariant, error) {
	product, err := r.GetByIDForOutlet(ctx, storeID, outletID, productID)
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

// InsertVariant hanya menyisipkan baris varian (tanpa stok/movement):
// orkestrasi stok awal ada di Service lewat mesin stok inventory (Unit B).
func (r *ProductRepo) InsertVariant(ctx context.Context, storeID, outletID, productID uint, variant *ProductVariant) error {
	variant.StoreID = storeID
	variant.ProductID = productID
	variant.OutletID = &outletID
	if err := r.DB(ctx).Create(variant).Error; err != nil {
		return mapDBErr(err)
	}
	return nil
}

func (r *ProductRepo) IsVariantSkuTaken(ctx context.Context, storeID uint, sku string, exceptID uint) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&ProductVariant{}).Where("store_id = ? AND lower(sku) = lower(?) AND id <> ?", storeID, strings.TrimSpace(sku), exceptID).Count(&count).Error
	if err != nil || count > 0 {
		return count > 0, err
	}
	err = r.DB(ctx).Model(&Product{}).Where("store_id = ? AND lower(sku) = lower(?)", storeID, strings.TrimSpace(sku)).Count(&count).Error
	return count > 0, err
}

// UpdateVariantFields hanya memperbarui kolom non-stok varian
// (nama/SKU/harga/min_stock): penyesuaian stok + movement-nya
// diorkestrasi Service lewat mesin stok inventory (Unit B).
func (r *ProductRepo) UpdateVariantFields(ctx context.Context, storeID, productID, variantID uint, variant *ProductVariant) error {
	var current ProductVariant
	if err := r.DB(ctx).Where("id = ? AND store_id = ? AND product_id = ?", variantID, storeID, productID).First(&current).Error; err != nil {
		return mapDBErr(err)
	}
	if err := r.DB(ctx).Model(&current).Updates(map[string]any{
		"name": variant.Name, "sku": variant.SKU, "buy_price": variant.BuyPrice,
		"sell_price": variant.SellPrice, "min_stock": variant.MinStock,
	}).Error; err != nil {
		return mapDBErr(err)
	}
	return nil
}

func (r *ProductRepo) defaultOutlet(ctx context.Context, storeID uint) (*outlet.Outlet, error) {
	var outlet outlet.Outlet
	if err := r.DB(ctx).Where("store_id = ? AND is_default = ?", storeID, true).First(&outlet).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &outlet, nil
}

func (r *ProductRepo) DeleteVariant(ctx context.Context, storeID, productID, variantID uint) error {
	res := r.DB(ctx).Where("id = ? AND store_id = ? AND product_id = ?", variantID, storeID, productID).Delete(&ProductVariant{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ProductRepo) ListPriceTiers(ctx context.Context, storeID, productID uint, variantID *uint) ([]PriceTier, error) {
	if variantID == nil {
		if _, err := r.GetByIDForStore(ctx, storeID, productID); err != nil {
			return nil, err
		}
	} else if _, err := r.GetVariant(ctx, storeID, productID, *variantID); err != nil {
		return nil, err
	}
	query := r.DB(ctx).Where("store_id = ? AND product_id = ?", storeID, productID)
	if variantID == nil {
		query = query.Where("variant_id IS NULL")
	} else {
		query = query.Where("variant_id = ?", *variantID)
	}
	var tiers []PriceTier
	if err := query.Order("min_qty ASC, id ASC").Find(&tiers).Error; err != nil {
		return nil, err
	}
	if tiers == nil {
		tiers = []PriceTier{}
	}
	return tiers, nil
}

func (r *ProductRepo) CreatePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tier *PriceTier) (*PriceTier, error) {
	if variantID == nil {
		if _, err := r.GetByIDForStore(ctx, storeID, productID); err != nil {
			return nil, err
		}
	} else if _, err := r.GetVariant(ctx, storeID, productID, *variantID); err != nil {
		return nil, err
	}
	query := r.DB(ctx).Model(&PriceTier{}).Where("store_id = ? AND product_id = ? AND min_qty = ?", storeID, productID, tier.MinQty)
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
	if err := r.DB(ctx).Create(tier).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return tier, nil
}

func (r *ProductRepo) UpdatePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tierID uint, tier *PriceTier) (*PriceTier, error) {
	query := r.DB(ctx).Model(&PriceTier{}).Where("id = ? AND store_id = ? AND product_id = ?", tierID, storeID, productID)
	if variantID == nil {
		query = query.Where("variant_id IS NULL")
	} else {
		query = query.Where("variant_id = ?", *variantID)
	}
	var current PriceTier
	if err := query.First(&current).Error; err != nil {
		return nil, mapDBErr(err)
	}
	dup := r.DB(ctx).Model(&PriceTier{}).Where("store_id = ? AND product_id = ? AND min_qty = ? AND id <> ?", storeID, productID, tier.MinQty, tierID)
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
	update := r.DB(ctx).Model(&PriceTier{}).Where("id = ? AND store_id = ? AND product_id = ?", tierID, storeID, productID)
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
	query := r.DB(ctx).Where("id = ? AND store_id = ? AND product_id = ?", tierID, storeID, productID)
	if variantID == nil {
		query = query.Where("variant_id IS NULL")
	} else {
		query = query.Where("variant_id = ?", *variantID)
	}
	res := query.Delete(&PriceTier{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
