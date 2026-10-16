package repo

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/rayhanvito/nojpos/backend/model"
)

var (
	ErrPurchaseOrderItemsLocked   = errors.New("barang pada PO sudah terkunci")
	ErrPurchaseOrderState         = errors.New("status PO tidak mengizinkan aksi ini")
	ErrPurchaseOrderNumberLimit   = errors.New("nomor PO harian sudah mencapai batas")
	ErrPurchaseOrderReceiptItem   = errors.New("barang penerimaan tidak sesuai dengan PO")
	ErrPurchaseOrderReceiptOver   = errors.New("jumlah penerimaan melebihi sisa pesanan")
	ErrPurchaseOrderReceiptTarget = errors.New("target stok PO sudah tidak tersedia")
)

type PurchaseOrderFilter struct {
	Status     string
	SupplierID uint
}

type PurchaseOrderReceiptItem struct {
	ItemID uint
	Qty    int
}

type PurchaseOrderRepo struct {
	db *gorm.DB
}

func NewPurchaseOrderRepo(db *gorm.DB) *PurchaseOrderRepo { return &PurchaseOrderRepo{db: db} }

func (r *PurchaseOrderRepo) List(ctx context.Context, storeID, outletID uint, f PurchaseOrderFilter) ([]*model.PurchaseOrder, error) {
	query := r.db.WithContext(ctx).Model(&model.PurchaseOrder{}).Where("store_id = ?", storeID)
	if outletID != 0 {
		query = query.Where("outlet_id = ?", outletID)
	}
	if f.Status != "" {
		query = query.Where("status = ?", f.Status)
	}
	if f.SupplierID != 0 {
		query = query.Where("supplier_id = ?", f.SupplierID)
	}
	items := make([]*model.PurchaseOrder, 0)
	err := query.Order("created_at DESC, id DESC").Find(&items).Error
	return items, err
}

func (r *PurchaseOrderRepo) GetByID(ctx context.Context, storeID, outletID, id uint) (*model.PurchaseOrder, error) {
	var po model.PurchaseOrder
	query := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", id, storeID)
	if outletID != 0 {
		query = query.Where("outlet_id = ?", outletID)
	}
	if err := query.First(&po).Error; err != nil {
		return nil, mapDBErr(err)
	}
	var supplier model.Supplier
	if err := r.db.WithContext(ctx).Unscoped().Where("id = ? AND store_id = ?", po.SupplierID, storeID).First(&supplier).Error; err == nil {
		po.Supplier = &supplier
	} else if !isNotFound(err) {
		return nil, err
	}
	po.Items = make([]model.PurchaseOrderItem, 0)
	if err := r.db.WithContext(ctx).Where("purchase_order_id = ?", po.ID).Order("id ASC").Find(&po.Items).Error; err != nil {
		return nil, err
	}
	return &po, nil
}

func (r *PurchaseOrderRepo) CreateDraft(ctx context.Context, storeID, outletID, supplierID uint, items []model.PurchaseOrderItem, datePrefix string) (*model.PurchaseOrder, error) {
	var createdID uint
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var store model.Store
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&store, storeID).Error; err != nil {
			return mapDBErr(err)
		}
		var count int64
		if err := tx.Unscoped().Model(&model.PurchaseOrder{}).
			Where("store_id = ? AND number LIKE ?", storeID, "PO-"+datePrefix+"-%").Count(&count).Error; err != nil {
			return err
		}
		if count >= 9999 {
			return ErrPurchaseOrderNumberLimit
		}

		po := model.PurchaseOrder{
			StoreID: storeID, OutletID: uintPointer(outletID), SupplierID: supplierID,
			Number: fmt.Sprintf("PO-%s-%04d", datePrefix, count+1),
			Status: model.PurchaseOrderDraft,
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

func (r *PurchaseOrderRepo) UpdateDraft(ctx context.Context, storeID, outletID, id, supplierID uint, items []model.PurchaseOrderItem) (*model.PurchaseOrder, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		po, err := lockPurchaseOrder(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if po.Status != model.PurchaseOrderDraft {
			return ErrPurchaseOrderItemsLocked
		}
		if err := tx.Model(po).Update("supplier_id", supplierID).Error; err != nil {
			return mapDBErr(err)
		}
		if err := tx.Where("purchase_order_id = ?", id).Delete(&model.PurchaseOrderItem{}).Error; err != nil {
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
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		po, err := lockPurchaseOrder(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if po.Status != model.PurchaseOrderDraft {
			return ErrPurchaseOrderState
		}
		if err := tx.Where("purchase_order_id = ?", id).Delete(&model.PurchaseOrderItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(po).Error
	})
}

func (r *PurchaseOrderRepo) Order(ctx context.Context, storeID, outletID, id uint) (*model.PurchaseOrder, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		po, err := lockPurchaseOrder(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if po.Status != model.PurchaseOrderDraft {
			return ErrPurchaseOrderState
		}
		return tx.Model(po).Update("status", model.PurchaseOrderOrdered).Error
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, storeID, outletID, id)
}

func (r *PurchaseOrderRepo) Receive(ctx context.Context, storeID, outletID, id uint, items []PurchaseOrderReceiptItem, actor string) (*model.PurchaseOrder, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		po, err := lockPurchaseOrder(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if po.Status != model.PurchaseOrderOrdered && po.Status != model.PurchaseOrderPartiallyReceived {
			return ErrPurchaseOrderState
		}
		if len(items) == 0 {
			return ErrPurchaseOrderReceiptItem
		}

		var orderedItems []model.PurchaseOrderItem
		if err := tx.Where("purchase_order_id = ?", po.ID).Find(&orderedItems).Error; err != nil {
			return err
		}
		byID := make(map[uint]model.PurchaseOrderItem, len(orderedItems))
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
			if err := tx.Model(&model.Product{}).
				Where("id = ? AND store_id = ?", item.ProductID, storeID).
				Count(&productCount).Error; err != nil {
				return err
			}
			if productCount == 0 {
				return ErrPurchaseOrderReceiptTarget
			}

			if item.VariantID != nil {
				var variantCount int64
				if err := tx.Model(&model.ProductVariant{}).
					Where("id = ? AND store_id = ? AND product_id = ?", *item.VariantID, storeID, item.ProductID).
					Count(&variantCount).Error; err != nil {
					return err
				}
				if variantCount == 0 {
					return ErrPurchaseOrderReceiptTarget
				}
				if _, _, _, err := changeOutletStock(tx, storeID, outletID, item.ProductID, item.VariantID, received.Qty, false); err != nil {
					if errors.Is(err, ErrNotFound) {
						return ErrPurchaseOrderReceiptTarget
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
					return ErrPurchaseOrderReceiptTarget
				}
				if _, _, _, err := changeOutletStock(tx, storeID, outletID, item.ProductID, nil, received.Qty, false); err != nil {
					if errors.Is(err, ErrNotFound) {
						return ErrPurchaseOrderReceiptTarget
					}
					return err
				}
			}

			updated := tx.Model(&model.PurchaseOrderItem{}).
				Where("id = ? AND purchase_order_id = ? AND qty_received + ? <= qty_ordered", received.ItemID, po.ID, received.Qty).
				Update("qty_received", gorm.Expr("qty_received + ?", received.Qty))
			if updated.Error != nil {
				return updated.Error
			}
			if updated.RowsAffected == 0 {
				return ErrPurchaseOrderReceiptOver
			}

			movement := model.Movement{
				StoreID: storeID, OutletID: uintPointer(outletID), ProductID: item.ProductID, VariantID: item.VariantID,
				Type: model.MovementPurchase, Qty: received.Qty,
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
		if err := tx.Model(&model.PurchaseOrderItem{}).
			Where("purchase_order_id = ? AND qty_received < qty_ordered", po.ID).Count(&remaining).Error; err != nil {
			return err
		}
		status := model.PurchaseOrderPartiallyReceived
		if remaining == 0 {
			status = model.PurchaseOrderFullyReceived
		}
		if err := tx.Model(po).Update("status", status).Error; err != nil {
			return err
		}
		payableStatus := model.DebtUnpaid
		if payableAmount == 0 {
			payableStatus = model.DebtPaid
		}
		return tx.Create(&model.Payable{
			StoreID: storeID, SupplierID: po.SupplierID, PurchaseOrderID: po.ID,
			Amount: payableAmount, Status: payableStatus,
		}).Error
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, storeID, outletID, id)
}

func (r *PurchaseOrderRepo) Cancel(ctx context.Context, storeID, outletID, id uint) (*model.PurchaseOrder, error) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		po, err := lockPurchaseOrder(tx, storeID, outletID, id)
		if err != nil {
			return err
		}
		if po.Status != model.PurchaseOrderDraft && po.Status != model.PurchaseOrderOrdered {
			return ErrPurchaseOrderState
		}
		return tx.Model(po).Update("status", model.PurchaseOrderCancelled).Error
	})
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, storeID, outletID, id)
}

func lockPurchaseOrder(tx *gorm.DB, storeID, outletID, id uint) (*model.PurchaseOrder, error) {
	var po model.PurchaseOrder
	query := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND store_id = ?", id, storeID)
	if outletID != 0 {
		query = query.Where("outlet_id = ?", outletID)
	}
	if err := query.First(&po).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &po, nil
}

func createPurchaseOrderItems(tx *gorm.DB, poID uint, items []model.PurchaseOrderItem) error {
	for i := range items {
		items[i].PurchaseOrderID = poID
		if err := tx.Create(&items[i]).Error; err != nil {
			return mapDBErr(err)
		}
	}
	return nil
}
