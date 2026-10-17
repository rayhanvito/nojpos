package customer

import (
	"context"
	"errors"
	"strings"

	"gorm.io/gorm"

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

type CustomerFilter struct {
	Q     string
	Page  int
	Limit int
}

type CustomerPage struct {
	Items []*Customer `json:"items"`
	Total int         `json:"total"`
	Page  int         `json:"page"`
	Limit int         `json:"limit"`
}

// Repo adalah repositori pelanggan (CustomerRepo lama di
// repo/customer.go); metode-metodenya verbatim, akses DB lewat DB(ctx).
type Repo struct {
	database.BaseRepo
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *Repo) List(ctx context.Context, storeID uint, f CustomerFilter) (*CustomerPage, error) {
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

	query := r.DB(ctx).Model(&Customer{}).Where("store_id = ?", storeID)
	if q := strings.TrimSpace(f.Q); q != "" {
		searchTerm := "%" + strings.ToLower(q) + "%"
		query = query.Where("lower(name) LIKE ? OR lower(phone) LIKE ?", searchTerm, searchTerm)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	items := make([]*Customer, 0)
	err := query.Order("created_at DESC, id DESC").Limit(limit).Offset((page - 1) * limit).Find(&items).Error
	if err != nil {
		return nil, err
	}
	return &CustomerPage{Items: items, Total: int(total), Page: page, Limit: limit}, nil
}

func (r *Repo) GetByID(ctx context.Context, storeID, id uint) (*Customer, error) {
	var customer Customer
	if err := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&customer).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &customer, nil
}

func (r *Repo) Create(ctx context.Context, storeID uint, customer *Customer) (*Customer, error) {
	customer.StoreID = storeID
	err := r.DB(ctx).Create(customer).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	return r.GetByID(ctx, storeID, customer.ID)
}

func (r *Repo) Update(ctx context.Context, storeID, id uint, customer *Customer) (*Customer, error) {
	res := r.DB(ctx).Model(&Customer{}).Where("id = ? AND store_id = ?", id, storeID).Updates(map[string]any{
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

func (r *Repo) Delete(ctx context.Context, storeID, id uint) error {
	res := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).Delete(&Customer{})
	if res.Error != nil {
		return mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repo) ListTransactions(ctx context.Context, storeID, customerID uint) ([]*trx, error) {
	items := make([]*trx, 0)
	err := r.DB(ctx).Where("store_id = ? AND customer_id = ?", storeID, customerID).
		Order("created_at DESC, id DESC").Find(&items).Error
	return items, err
}
