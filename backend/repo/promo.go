package repo

import (
	"context"

	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/model"
)

type PromoRepo struct {
	db *gorm.DB
}

func NewPromoRepo(db *gorm.DB) *PromoRepo { return &PromoRepo{db: db} }

func (r *PromoRepo) List(ctx context.Context, storeID uint) ([]*model.Promo, error) {
	items := make([]*model.Promo, 0)
	err := r.db.WithContext(ctx).Where("store_id = ?", storeID).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *PromoRepo) GetByID(ctx context.Context, storeID, id uint) (*model.Promo, error) {
	var promo model.Promo
	if err := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&promo).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &promo, nil
}

func (r *PromoRepo) Create(ctx context.Context, storeID uint, promo *model.Promo) (*model.Promo, error) {
	promo.StoreID = storeID
	if err := r.db.WithContext(ctx).Create(promo).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return promo, nil
}

func (r *PromoRepo) Update(ctx context.Context, storeID, id uint, promo *model.Promo) (*model.Promo, error) {
	res := r.db.WithContext(ctx).Model(&model.Promo{}).Where("id = ? AND store_id = ?", id, storeID).Updates(map[string]any{
		"name": promo.Name, "type": promo.Type, "value": promo.Value, "amount_value": promo.AmountValue, "scope": promo.Scope,
		"category_id": promo.CategoryID, "product_id": promo.ProductID, "voucher_code": promo.VoucherCode,
		"starts_at": promo.StartsAt, "ends_at": promo.EndsAt, "happy_start": promo.HappyStart,
		"happy_end": promo.HappyEnd, "min_purchase": promo.MinPurchase, "max_discount": promo.MaxDiscount,
	})
	if res.Error != nil {
		return nil, mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, storeID, id)
}

func (r *PromoRepo) SetActive(ctx context.Context, storeID, id uint, active bool) error {
	res := r.db.WithContext(ctx).Model(&model.Promo{}).Where("id = ? AND store_id = ?", id, storeID).Update("active", active)
	if res.Error != nil {
		return mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PromoRepo) Delete(ctx context.Context, storeID, id uint) error {
	res := r.db.WithContext(ctx).Unscoped().Where("id = ? AND store_id = ?", id, storeID).Delete(&model.Promo{})
	if res.Error != nil {
		return mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
