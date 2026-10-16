package repo

import (
	"context"

	"gorm.io/gorm"

	"github.com/adrr-dev/openPOS/backend/model"
)

type SupplierRepo struct {
	db *gorm.DB
}

func NewSupplierRepo(db *gorm.DB) *SupplierRepo { return &SupplierRepo{db: db} }

func (r *SupplierRepo) List(ctx context.Context, storeID uint) ([]*model.Supplier, error) {
	items := make([]*model.Supplier, 0)
	err := r.db.WithContext(ctx).Where("store_id = ?", storeID).Order("name ASC, id ASC").Find(&items).Error
	return items, err
}

func (r *SupplierRepo) GetByID(ctx context.Context, storeID, id uint) (*model.Supplier, error) {
	var supplier model.Supplier
	if err := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&supplier).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &supplier, nil
}

func (r *SupplierRepo) Create(ctx context.Context, storeID uint, supplier *model.Supplier) (*model.Supplier, error) {
	supplier.StoreID = storeID
	if err := r.db.WithContext(ctx).Create(supplier).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return supplier, nil
}

func (r *SupplierRepo) Update(ctx context.Context, storeID, id uint, supplier *model.Supplier) (*model.Supplier, error) {
	res := r.db.WithContext(ctx).Model(&model.Supplier{}).Where("id = ? AND store_id = ?", id, storeID).Updates(map[string]any{
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
	res := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", id, storeID).Delete(&model.Supplier{})
	if res.Error != nil {
		return mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
