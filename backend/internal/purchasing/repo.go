package purchasing

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/finance"
	"github.com/rayhanvito/nojpos/backend/internal/inventory"
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

func uintPointer(value uint) *uint { return &value }

// SupplierRepo adalah repositori supplier (SupplierRepo lama di
// repo/supplier.go); metode-metodenya verbatim, akses DB lewat DB(ctx).
type SupplierRepo struct {
	database.BaseRepo
}

func NewSupplierRepo(db *gorm.DB) *SupplierRepo {
	return &SupplierRepo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *SupplierRepo) List(ctx context.Context, storeID uint) ([]*Supplier, error) {
	items := make([]*Supplier, 0)
	err := r.DB(ctx).Where("store_id = ?", storeID).Order("name ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *SupplierRepo) GetByID(ctx context.Context, storeID, id uint) (*Supplier, error) {
	var supplier Supplier
	if err := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&supplier).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &supplier, nil
}

func (r *SupplierRepo) Create(ctx context.Context, storeID uint, supplier *Supplier) (*Supplier, error) {
	supplier.StoreID = storeID
	if err := r.DB(ctx).Create(supplier).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return supplier, nil
}

func (r *SupplierRepo) Update(ctx context.Context, storeID, id uint, supplier *Supplier) (*Supplier, error) {
	res := r.DB(ctx).Model(&Supplier{}).Where("id = ? AND store_id = ?", id, storeID).Updates(map[string]any{
		"name": supplier.Name, "contact": supplier.Contact, "address": supplier.Address,
	})
	if res.Error != nil {
		return nil, mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, storeID, id)
}

func (r *SupplierRepo) Delete(ctx context.Context, storeID, id uint) error {
	res := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).Delete(&Supplier{})
	if res.Error != nil {
		return mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// PurchaseOrderFilter dan PurchaseOrderReceiptItem dipindah verbatim
// dari repo/purchase_order.go (BE-F7).
type PurchaseOrderFilter struct {
	Status     string
	SupplierID uint
}

type PurchaseOrderReceiptItem struct {
	ItemID uint
	Qty    int
}

// PurchaseOrderRepo adalah repositori purchase order (PurchaseOrderRepo
// lama di repo/purchase_order.go): sudah menyematkan BaseRepo dan pola
// WithinTx sejak BE-C, dipertahankan apa adanya; metode-metodenya
// verbatim, akses DB lewat DB(ctx).
type PurchaseOrderRepo struct {
	database.BaseRepo
}

func NewPurchaseOrderRepo(db *gorm.DB) *PurchaseOrderRepo {
	return &PurchaseOrderRepo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *PurchaseOrderRepo) List(ctx context.Context, storeID, outletID uint, f PurchaseOrderFilter) ([]*PurchaseOrder, error) {
	query := r.DB(ctx).Model(&PurchaseOrder{}).Where("store_id = ?", storeID)
	if outletID != 0 {
		query = query.Where("outlet_id = ?", outletID)
	}
	if f.Status != "" {
		query = query.Where("status = ?", f.Status)
	}
	if f.SupplierID != 0 {
		query = query.Where("supplier_id = ?", f.SupplierID)
	}
	items := make([]*PurchaseOrder, 0)
	err := query.Order("created_at DESC, id DESC").Find(&items).Error
	return items, err
}

func (r *PurchaseOrderRepo) GetByID(ctx context.Context, storeID, outletID, id uint) (*PurchaseOrder, error) {
	var po PurchaseOrder
	query := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID)
	if outletID != 0 {
		query = query.Where("outlet_id = ?", outletID)
	}
	if err := query.First(&po).Error; err != nil {
		return nil, mapDBErr(err)
	}
	var supplier Supplier
	if err := r.DB(ctx).Unscoped().Where("id = ? AND store_id = ?", po.SupplierID, storeID).First(&supplier).Error; err == nil {
		po.Supplier = &supplier
	} else if !isNotFound(err) {
		return nil, err
	}
	po.Items = make([]PurchaseOrderItem, 0)
	if err := r.DB(ctx).Where("purchase_order_id = ?", po.ID).Order("id ASC").Find(&po.Items).Error; err != nil {
		return nil, err
	}
	return &po, nil
}

func (r *PurchaseOrderRepo) CreateDraft(ctx context.Context, storeID, outletID, supplierID uint, items []PurchaseOrderItem, datePrefix string) (*PurchaseOrder, error) {
	var createdID uint
	err := r.WithinTx(ctx, func(ctx context.Context) error {
		tx := r.DB(ctx)
		var store outlet.Store
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&store, storeID).Error; err != nil {
			return mapDBErr(err)
		}
		var count int64
		if err := tx.Unscoped().Model(&PurchaseOrder{}).
			Where("store_id = ? AND number LIKE ?", storeID, "PO-"+datePrefix+"-%").Count(&count).Error; err != nil {
			return err
		}
		if count >= 9999 {
			return ErrPurchaseOrderNumberLimit
		}

		po := PurchaseOrder{
			StoreID: storeID, OutletID: uintPointer(outletID), SupplierID: supplierID,
			Number: fmt.Sprintf("PO-%s-%04d", datePrefix, count+1),
			Status: PurchaseOrderDraft,
		}
		if err := tx.Create(&po).Error; err != nil {
			return mapDBErr(err)
		}
		createdID = po.ID
		return createPurchaseOrderItems(tx, po.ID, items)
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, storeID, outletID, createdID)
}

func (r *PurchaseOrderRepo) UpdateDraft(ctx context.Context, storeID, outletID, id, supplierID uint, items []PurchaseOrderItem) (*PurchaseOrder, error) {
	err := r.WithinTx(ctx, func(ctx context.Context) error {
		tx := r.DB(ctx)
		po, err := lockPurchaseOrder(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if po.Status != PurchaseOrderDraft {
			return ErrPurchaseOrderItemsLocked
		}
		if err := tx.Model(po).Update("supplier_id", supplierID).Error; err != nil {
			return mapDBErr(err)
		}
		if err := tx.Where("purchase_order_id = ?", id).Delete(&PurchaseOrderItem{}).Error; err != nil {
			return err
		}
		return createPurchaseOrderItems(tx, id, items)
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, storeID, outletID, id)
}

func (r *PurchaseOrderRepo) DeleteDraft(ctx context.Context, storeID, outletID, id uint) error {
	return r.WithinTx(ctx, func(ctx context.Context) error {
		tx := r.DB(ctx)
		po, err := lockPurchaseOrder(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if po.Status != PurchaseOrderDraft {
			return ErrPurchaseOrderState
		}
		if err := tx.Where("purchase_order_id = ?", id).Delete(&PurchaseOrderItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(po).Error
	})
}

func (r *PurchaseOrderRepo) Order(ctx context.Context, storeID, outletID, id uint) (*PurchaseOrder, error) {
	err := r.WithinTx(ctx, func(ctx context.Context) error {
		tx := r.DB(ctx)
		po, err := lockPurchaseOrder(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if po.Status != PurchaseOrderDraft {
			return ErrPurchaseOrderState
		}
		return tx.Model(po).Update("status", PurchaseOrderOrdered).Error
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, storeID, outletID, id)
}

func (r *PurchaseOrderRepo) Receive(ctx context.Context, storeID, outletID, id uint, items []PurchaseOrderReceiptItem, actor string) (*PurchaseOrder, error) {
	err := r.WithinTx(ctx, func(ctx context.Context) error {
		tx := r.DB(ctx)
		po, err := lockPurchaseOrder(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if po.Status != PurchaseOrderOrdered && po.Status != PurchaseOrderPartiallyReceived {
			return ErrPurchaseOrderState
		}
		if len(items) == 0 {
			return ErrPurchaseOrderReceiptItem
		}

		var orderedItems []PurchaseOrderItem
		if err := tx.Where("purchase_order_id = ?", po.ID).Find(&orderedItems).Error; err != nil {
			return err
		}
		byID := make(map[uint]PurchaseOrderItem, len(orderedItems))
		for _, item := range orderedItems {
			byID[item.ID] = item
		}
		seen := make(map[uint]bool, len(items))
		for _, received := range items {
			item, ok := byID[received.ItemID]
			if received.ItemID == 0 || received.Qty <= 0 || !ok || seen[received.ItemID] {
				return ErrPurchaseOrderReceiptItem
			}
			if received.Qty > item.QtyOrdered-item.QtyReceived {
				return ErrPurchaseOrderReceiptOver
			}
			seen[received.ItemID] = true
		}

		for _, received := range items {
			item := byID[received.ItemID]
			var productCount int64
			if err := tx.Model(&catalog.Product{}).
				Where("id = ? AND store_id = ?", item.ProductID, storeID).
				Count(&productCount).Error; err != nil {
				return err
			}
			if productCount == 0 {
				return ErrPurchaseOrderReceiptTarget
			}

			if item.VariantID != nil {
				var variantCount int64
				if err := tx.Model(&catalog.ProductVariant{}).
					Where("id = ? AND store_id = ? AND product_id = ?", *item.VariantID, storeID, item.ProductID).
					Count(&variantCount).Error; err != nil {
					return err
				}
				if variantCount == 0 {
					return ErrPurchaseOrderReceiptTarget
				}
				if _, _, _, err := inventory.ChangeOutletStock(tx, storeID, outletID, item.ProductID, item.VariantID, received.Qty, false); err != nil {
					if errors.Is(err, ErrNotFound) {
						return ErrPurchaseOrderReceiptTarget
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
					return ErrPurchaseOrderReceiptTarget
				}
				if _, _, _, err := inventory.ChangeOutletStock(tx, storeID, outletID, item.ProductID, nil, received.Qty, false); err != nil {
					if errors.Is(err, ErrNotFound) {
						return ErrPurchaseOrderReceiptTarget
					}
					return err
				}
			}

			updated := tx.Model(&PurchaseOrderItem{}).
				Where("id = ? AND purchase_order_id = ? AND qty_received + ? <= qty_ordered", received.ItemID, po.ID, received.Qty).
				Update("qty_received", gorm.Expr("qty_received + ?", received.Qty))
			if updated.Error != nil {
				return updated.Error
			}
			if updated.RowsAffected == 0 {
				return ErrPurchaseOrderReceiptOver
			}

			movement := inventory.Movement{
				StoreID: storeID, OutletID: uintPointer(outletID), ProductID: item.ProductID, VariantID: item.VariantID,
				Type: inventory.MovementPurchase, Qty: received.Qty,
				Reason: fmt.Sprintf("Penerimaan PO %s", po.Number), Actor: actor,
			}
			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
		}

		var payableAmount int64
		for _, received := range items {
			payableAmount += int64(received.Qty) * byID[received.ItemID].BuyPrice
		}

		var remaining int64
		if err := tx.Model(&PurchaseOrderItem{}).
			Where("purchase_order_id = ? AND qty_received < qty_ordered", po.ID).Count(&remaining).Error; err != nil {
			return err
		}
		status := PurchaseOrderPartiallyReceived
		if remaining == 0 {
			status = PurchaseOrderFullyReceived
		}
		if err := tx.Model(po).Update("status", status).Error; err != nil {
			return err
		}
		payableStatus := finance.DebtUnpaid
		if payableAmount == 0 {
			payableStatus = finance.DebtPaid
		}
		return tx.Create(&finance.Payable{
			StoreID: storeID, SupplierID: po.SupplierID, PurchaseOrderID: po.ID,
			Amount: payableAmount, Status: payableStatus,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, storeID, outletID, id)
}

func (r *PurchaseOrderRepo) Cancel(ctx context.Context, storeID, outletID, id uint) (*PurchaseOrder, error) {
	err := r.WithinTx(ctx, func(ctx context.Context) error {
		tx := r.DB(ctx)
		po, err := lockPurchaseOrder(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if po.Status != PurchaseOrderDraft && po.Status != PurchaseOrderOrdered {
			return ErrPurchaseOrderState
		}
		return tx.Model(po).Update("status", PurchaseOrderCancelled).Error
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, storeID, outletID, id)
}

func lockPurchaseOrder(tx *gorm.DB, storeID, outletID, id uint) (*PurchaseOrder, error) {
	var po PurchaseOrder
	query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND store_id = ?", id, storeID)
	if outletID != 0 {
		query = query.Where("outlet_id = ?", outletID)
	}
	if err := query.First(&po).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &po, nil
}

func createPurchaseOrderItems(tx *gorm.DB, poID uint, items []PurchaseOrderItem) error {
	for i := range items {
		items[i].PurchaseOrderID = poID
		if err := tx.Create(&items[i]).Error; err != nil {
			return mapDBErr(err)
		}
	}
	return nil
}
