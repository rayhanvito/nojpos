package repo

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/adrr-dev/openPOS/backend/model"
)

type CustomerFilter struct {
	Q     string
	Page  int
	Limit int
}

type CustomerPage struct {
	Items []*model.Customer `json:"items"`
	Total int               `json:"total"`
	Page  int               `json:"page"`
	Limit int               `json:"limit"`
}

type CustomerRepo struct {
	db *gorm.DB
}

func NewCustomerRepo(db *gorm.DB) *CustomerRepo { return &CustomerRepo{db: db} }

func (r *CustomerRepo) List(ctx context.Context, storeID uint, f CustomerFilter) (*CustomerPage, error) {
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

	query := r.db.WithContext(ctx).Model(&model.Customer{}).Where("store_id = ?", storeID)
	if q := strings.TrimSpace(f.Q); q != "" {
		searchTerm := "%" + strings.ToLower(q) + "%"
		query = query.Where("lower(name) LIKE ? OR lower(phone) LIKE ?", searchTerm, searchTerm)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	items := make([]*model.Customer, 0)
	err := query.Order("created_at DESC, id DESC").Limit(limit).Offset((page - 1) * limit).Find(&items).Error
	if err != nil {
		return nil, err
	}
	return &CustomerPage{Items: items, Total: int(total), Page: page, Limit: limit}, nil
}

func (r *CustomerRepo) GetByID(ctx context.Context, storeID, id uint) (*model.Customer, error) {
	var customer model.Customer
	if err := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&customer).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &customer, nil
}

func (r *CustomerRepo) Create(ctx context.Context, storeID uint, customer *model.Customer) (*model.Customer, error) {
	customer.StoreID = storeID
	err := r.db.WithContext(ctx).Create(customer).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	return r.GetByID(ctx, storeID, customer.ID)
}

func (r *CustomerRepo) Update(ctx context.Context, storeID, id uint, customer *model.Customer) (*model.Customer, error) {
	res := r.db.WithContext(ctx).Model(&model.Customer{}).Where("id = ? AND store_id = ?", id, storeID).Updates(map[string]any{
		"name": customer.Name, "phone": customer.Phone, "email": customer.Email, "address": customer.Address,
	})
	if res.Error != nil {
		return nil, mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, storeID, id)
}

func (r *CustomerRepo) Delete(ctx context.Context, storeID, id uint) error {
	res := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", id, storeID).Delete(&model.Customer{})
	if res.Error != nil {
		return mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *CustomerRepo) ListTransactions(ctx context.Context, storeID, customerID uint) ([]*model.Trx, error) {
	items := make([]*model.Trx, 0)
	err := r.db.WithContext(ctx).Where("store_id = ? AND customer_id = ?", storeID, customerID).
		Order("created_at DESC, id DESC").Find(&items).Error
	return items, err
}
