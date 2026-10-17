package inventory

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
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

func uintPointer(value uint) *uint { return &value }

type MovementFilter struct {
	Type      string
	ProductID uint
	Q         string
	Page      int
	Limit     int
}

type MovementPage struct {
	Items []*Movement `json:"items"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

type MovementRepo struct {
	database.BaseRepo
}

func NewMovementRepo(db *gorm.DB) *MovementRepo {
	return &MovementRepo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *MovementRepo) InsertTx(ctx context.Context, tx *gorm.DB, m *Movement) error {
	return tx.WithContext(ctx).Create(m).Error
}

func (r *MovementRepo) ListForStore(ctx context.Context, storeID uint, f MovementFilter) (*MovementPage, error) {
	return r.list(ctx, storeID, nil, f)
}

func (r *MovementRepo) ListForOutlet(ctx context.Context, storeID, outletID uint, f MovementFilter) (*MovementPage, error) {
	return r.list(ctx, storeID, &outletID, f)
}

func (r *MovementRepo) list(ctx context.Context, storeID uint, outletID *uint, f MovementFilter) (*MovementPage, error) {
	page, limit := f.Page, f.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 25
	}
	if limit > 200 {
		limit = 200
	}

	query := database.ScopeStore(r.DB(ctx).Model(&Movement{}), storeID)
	if outletID != nil {
		query = database.ScopeOutlet(query, *outletID)
	}
	if f.Type != "" {
		query = query.Where("type = ?", f.Type)
	}
	if f.ProductID != 0 {
		query = query.Where("product_id = ?", f.ProductID)
	}
	if f.Q != "" {
		// Cari berdasarkan nama produk; subquery mentah supaya produk yang
		// sudah dihapus (soft-delete) tetap bisa ditemukan riwayatnya.
		query = query.Where("product_id IN (SELECT id FROM products WHERE store_id = ? AND name ILIKE ?)", storeID, "%"+f.Q+"%")
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	items := make([]*Movement, 0)
	offset := (page - 1) * limit
	err := query.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&items).Error
	if err != nil {
		return nil, err
	}

	for _, m := range items {
		var prod catalog.Product
		if err := r.DB(ctx).First(&prod, "id = ?", m.ProductID).Error; err == nil {
			m.ProductName = &prod.Name
		}
	}

	return &MovementPage{Items: items, Total: int(total), Page: page, Limit: limit}, nil
}

type StockOpnameItemCount struct {
	ItemID       uint
	CountedStock int
}

type StockOpnameRepo struct {
	database.BaseRepo
}

func NewStockOpnameRepo(db *gorm.DB) *StockOpnameRepo {
	return &StockOpnameRepo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *StockOpnameRepo) List(ctx context.Context, storeID, outletID uint) ([]StockOpnameSession, error) {
	items := make([]StockOpnameSession, 0)
	err := r.DB(ctx).Where("store_id = ? AND outlet_id = ?", storeID, outletID).Order("created_at DESC, id DESC").Find(&items).Error
	return items, err
}

func (r *StockOpnameRepo) GetByID(ctx context.Context, storeID, outletID, id uint) (*StockOpnameSession, error) {
	var session StockOpnameSession
	if err := r.DB(ctx).Where("id = ? AND store_id = ? AND outlet_id = ?", id, storeID, outletID).First(&session).Error; err != nil {
		return nil, mapDBErr(err)
	}
	if err := r.loadItems(ctx, storeID, &session); err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *StockOpnameRepo) Create(ctx context.Context, storeID, outletID uint, note, createdBy string) (*StockOpnameSession, error) {
	var sessionID uint
	err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var store outlet.Store
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&store, storeID).Error; err != nil {
			return mapDBErr(err)
		}
		var activeCount int64
		if err := tx.Model(&StockOpnameSession{}).
			Where("store_id = ? AND status = ?", storeID, StockOpnameCounting).Count(&activeCount).Error; err != nil {
			return err
		}
		if activeCount > 0 {
			return ErrStockOpnameActive
		}

		session := StockOpnameSession{
			StoreID: storeID, OutletID: uintPointer(outletID), Status: StockOpnameCounting,
			Note: note, CreatedBy: createdBy,
		}
		if err := tx.Create(&session).Error; err != nil {
			return err
		}
		sessionID = session.ID

		products := make([]catalog.Product, 0)
		if err := tx.Where("store_id = ? AND active = ?", storeID, true).Order("id ASC").Find(&products).Error; err != nil {
			return err
		}
		productIDs := make([]uint, len(products))
		for i := range products {
			productIDs[i] = products[i].ID
		}
		variants := make([]catalog.ProductVariant, 0)
		if len(productIDs) > 0 {
			if err := tx.Where("store_id = ? AND product_id IN ?", storeID, productIDs).Order("id ASC").Find(&variants).Error; err != nil {
				return err
			}
		}
		variantsByProduct := make(map[uint][]catalog.ProductVariant, len(products))
		for _, variant := range variants {
			variantsByProduct[variant.ProductID] = append(variantsByProduct[variant.ProductID], variant)
		}
		items := make([]StockOpnameItem, 0, len(products)+len(variants))
		for _, product := range products {
			productVariants := variantsByProduct[product.ID]
			if len(productVariants) == 0 {
				stock, err := GetOutletStock(tx, outletID, product.ID, nil)
				if err != nil {
					return err
				}
				items = append(items, StockOpnameItem{
					SessionID: session.ID, ProductID: product.ID, SystemStock: stock,
				})
				continue
			}
			for _, variant := range productVariants {
				variantID := variant.ID
				stock, err := GetOutletStock(tx, outletID, product.ID, &variantID)
				if err != nil {
					return err
				}
				items = append(items, StockOpnameItem{
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

func (r *StockOpnameRepo) UpdateItems(ctx context.Context, storeID, outletID, id uint, items []StockOpnameItemCount) (*StockOpnameSession, error) {
	err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		session, err := lockStockOpname(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if session.Status != StockOpnameCounting {
			return ErrStockOpnameState
		}
		itemIDs := make([]uint, len(items))
		for i := range items {
			itemIDs[i] = items[i].ItemID
		}
		var storedItems []StockOpnameItem
		if err := tx.Where("session_id = ? AND id IN ?", session.ID, itemIDs).Find(&storedItems).Error; err != nil {
			return err
		}
		if len(storedItems) != len(items) {
			return ErrStockOpnameItem
		}
		for _, item := range items {
			result := tx.Model(&StockOpnameItem{}).
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

func (r *StockOpnameRepo) Post(ctx context.Context, storeID, outletID, id uint, actor string) (*StockOpnameSession, error) {
	err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		session, err := lockStockOpname(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if session.Status != StockOpnameCounting {
			return ErrStockOpnameState
		}

		var items []StockOpnameItem
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
			if err := tx.Model(&catalog.Product{}).Where("id = ? AND store_id = ?", item.ProductID, storeID).Count(&productCount).Error; err != nil {
				return err
			}
			if productCount == 0 {
				return ErrStockOpnameTarget
			}

			if item.VariantID != nil {
				var variantCount int64
				if err := tx.Model(&catalog.ProductVariant{}).Where("id = ? AND product_id = ? AND store_id = ?", *item.VariantID, item.ProductID, storeID).Count(&variantCount).Error; err != nil {
					return err
				}
				if variantCount == 0 {
					return ErrStockOpnameTarget
				}
				if err := SetOutletStock(tx, storeID, outletID, item.ProductID, item.VariantID, *item.CountedStock); err != nil {
					if errors.Is(err, ErrNotFound) {
						return ErrStockOpnameTarget
					}
					return err
				}
			} else {
				var variantCount int64
				if err := tx.Model(&catalog.ProductVariant{}).
					Where("store_id = ? AND product_id = ?", storeID, item.ProductID).
					Count(&variantCount).Error; err != nil {
					return err
				}
				if variantCount > 0 {
					return ErrStockOpnameTarget
				}
				if err := SetOutletStock(tx, storeID, outletID, item.ProductID, nil, *item.CountedStock); err != nil {
					if errors.Is(err, ErrNotFound) {
						return ErrStockOpnameTarget
					}
					return err
				}
			}

			movement := Movement{
				StoreID: storeID, OutletID: uintPointer(outletID), ProductID: item.ProductID, VariantID: item.VariantID,
				Type: MovementOpname, Qty: difference,
				Reason: fmt.Sprintf("Opname #%d", session.ID), Actor: actor,
			}
			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
		}
		result := tx.Model(session).Where("status = ?", StockOpnameCounting).Update("status", StockOpnamePosted)
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

func (r *StockOpnameRepo) Cancel(ctx context.Context, storeID, outletID, id uint) (*StockOpnameSession, error) {
	err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		session, err := lockStockOpname(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if session.Status != StockOpnameCounting {
			return ErrStockOpnameState
		}
		result := tx.Model(session).Where("status = ?", StockOpnameCounting).Update("status", StockOpnameCancelled)
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

func (r *StockOpnameRepo) loadItems(ctx context.Context, storeID uint, session *StockOpnameSession) error {
	session.Items = make([]StockOpnameItem, 0)
	if err := r.DB(ctx).Where("session_id = ?", session.ID).Order("id ASC").Find(&session.Items).Error; err != nil {
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
	products := make([]catalog.Product, 0, len(productIDs))
	if len(productIDs) > 0 {
		if err := r.DB(ctx).Unscoped().Where("store_id = ? AND id IN ?", storeID, productIDs).Find(&products).Error; err != nil {
			return err
		}
	}
	productNames := make(map[uint]string, len(products))
	for _, product := range products {
		productNames[product.ID] = product.Name
	}
	variants := make([]catalog.ProductVariant, 0, len(variantIDs))
	if len(variantIDs) > 0 {
		if err := r.DB(ctx).Unscoped().Where("store_id = ? AND id IN ?", storeID, variantIDs).Find(&variants).Error; err != nil {
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

func lockStockOpname(tx *gorm.DB, storeID, outletID, id uint) (*StockOpnameSession, error) {
	var session StockOpnameSession
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND store_id = ? AND outlet_id = ?", id, storeID, outletID).First(&session).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &session, nil
}
