package repo

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/rayhanvito/nojpos/backend/model"
)

type CashierRepo struct {
	db *gorm.DB
}

func NewCashierRepo(db *gorm.DB) *CashierRepo { return &CashierRepo{db: db} }

func (r *CashierRepo) GetByID(ctx context.Context, id uint) (*model.Cashier, error) {
	var c model.Cashier
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&c).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	return &c, nil
}

func (r *CashierRepo) ListByStore(ctx context.Context, storeID uint) ([]*model.Cashier, error) {
	out := make([]*model.Cashier, 0)
	err := r.db.WithContext(ctx).Where("store_id = ?", storeID).Order("created_at ASC").Find(&out).Error
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (r *CashierRepo) Create(ctx context.Context, storeID, outletID uint, name string) (*model.Cashier, error) {
	c := model.Cashier{
		StoreID:  storeID,
		OutletID: &outletID,
		Name:     name,
		Active:   true,
	}
	err := r.db.WithContext(ctx).Create(&c).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	return &c, nil
}

func (r *CashierRepo) SetActive(ctx context.Context, id uint, active bool) error {
	res := r.db.WithContext(ctx).Model(&model.Cashier{}).Where("id = ?", id).Update("active", active)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CashierRepo) SetPasscode(ctx context.Context, id uint, hash *string) error {
	res := r.db.WithContext(ctx).Model(&model.Cashier{}).Where("id = ?", id).Update("passcode_hash", hash)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CashierRepo) GetOrCreateByName(ctx context.Context, storeID uint, name string) (uint, error) {
	var c model.Cashier
	err := r.db.WithContext(ctx).Where("store_id = ? AND name = ?", storeID, name).First(&c).Error
	if err == nil {
		return c.ID, nil
	}
	var outlet model.Outlet
	if err := r.db.WithContext(ctx).Where("store_id = ? AND is_default = ?", storeID, true).First(&outlet).Error; err != nil {
		return 0, mapDBErr(err)
	}
	created, err := r.Create(ctx, storeID, outlet.ID, name)
	if err != nil {
		return 0, err
	}
	return created.ID, nil
}

func (r *CashierRepo) Delete(ctx context.Context, id uint) error {
	res := r.db.WithContext(ctx).Delete(&model.Cashier{}, id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CashierRepo) UpdateLastSeenAt(ctx context.Context, id uint, lastSeenAt *time.Time) error {
	res := r.db.WithContext(ctx).Model(&model.Cashier{}).Where("id = ?", id).Update("last_seen_at", lastSeenAt)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CashierRepo) UpdateName(ctx context.Context, id uint, name string) error {
	res := r.db.WithContext(ctx).Model(&model.Cashier{}).Where("id = ?", id).Update("name", name)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CashierRepo) UpdateProfile(ctx context.Context, storeID, id uint, name *string, outletID *uint) error {
	updates := make(map[string]any, 2)
	if name != nil {
		updates["name"] = *name
	}
	if outletID != nil {
		updates["outlet_id"] = *outletID
	}
	if len(updates) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var cashier model.Cashier
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND store_id = ?", id, storeID).First(&cashier).Error; err != nil {
			return mapDBErr(err)
		}

		if outletID != nil && (cashier.OutletID == nil || *cashier.OutletID != *outletID) {
			var openShifts int64
			if err := tx.Model(&model.CashierShift{}).Where("store_id = ? AND cashier_id = ? AND closed_at IS NULL", storeID, id).Count(&openShifts).Error; err != nil {
				return err
			}
			if openShifts > 0 {
				return ErrCashierOpenShift
			}
		}

		result := tx.Model(&cashier).Updates(updates)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return ErrNotFound
		}
		return nil
	})
}
