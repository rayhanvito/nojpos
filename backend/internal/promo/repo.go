package promo

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

type Repo struct {
	database.BaseRepo
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{BaseRepo: database.NewBaseRepo(db)}
}

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

func (r *Repo) List(ctx context.Context, storeID uint) ([]*Promo, error) {
	items := make([]*Promo, 0)
	err := r.DB(ctx).Where("store_id = ?", storeID).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *Repo) GetByID(ctx context.Context, storeID, id uint) (*Promo, error) {
	var promo Promo
	if err := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&promo).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &promo, nil
}

func (r *Repo) Create(ctx context.Context, storeID uint, promo *Promo) (*Promo, error) {
	promo.StoreID = storeID
	if err := r.DB(ctx).Create(promo).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return promo, nil
}

func (r *Repo) Update(ctx context.Context, storeID, id uint, in *Promo) (*Promo, error) {
	res := r.DB(ctx).Model(&Promo{}).Where("id = ? AND store_id = ?", id, storeID).Updates(map[string]any{
		"name": in.Name, "type": in.Type, "value": in.Value, "amount_value": in.AmountValue, "scope": in.Scope,
		"category_id": in.CategoryID, "product_id": in.ProductID, "voucher_code": in.VoucherCode,
		"starts_at": in.StartsAt, "ends_at": in.EndsAt, "happy_start": in.HappyStart,
		"happy_end": in.HappyEnd, "min_purchase": in.MinPurchase, "max_discount": in.MaxDiscount,
	})
	if res.Error != nil {
		return nil, mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, storeID, id)
}

func (r *Repo) SetActive(ctx context.Context, storeID, id uint, active bool) error {
	res := r.DB(ctx).Model(&Promo{}).Where("id = ? AND store_id = ?", id, storeID).Update("active", active)
	if res.Error != nil {
		return mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) Delete(ctx context.Context, storeID, id uint) error {
	res := r.DB(ctx).Unscoped().Where("id = ? AND store_id = ?", id, storeID).Delete(&Promo{})
	if res.Error != nil {
		return mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
