package repo

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/adrr-dev/openPOS/backend/model"
)

type ExpenseFilter struct {
	From     string
	To       string
	Category string
	Q        string
	Page     int
	Limit    int
}

type ExpensePage struct {
	Items []*model.Expense `json:"items"`
	Total int              `json:"total"`
	Page  int              `json:"page"`
	Limit int              `json:"limit"`
}

type ExpenseRepo struct {
	db *gorm.DB
}

func NewExpenseRepo(db *gorm.DB) *ExpenseRepo { return &ExpenseRepo{db: db} }

func (r *ExpenseRepo) List(ctx context.Context, storeID uint, f ExpenseFilter) (*ExpensePage, error) {
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

	query := r.db.WithContext(ctx).Model(&model.Expense{}).Where("store_id = ?", storeID)
	if f.From != "" {
		query = query.Where("date >= ?", f.From)
	}
	if f.To != "" {
		query = query.Where("date <= ?", f.To)
	}
	if f.Category != "" {
		query = query.Where("category = ?", f.Category)
	}
	if q := strings.TrimSpace(f.Q); q != "" {
		like := "%" + strings.ToLower(q) + "%"
		query = query.Where("lower(note) LIKE ? OR lower(category) LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	items := make([]*model.Expense, 0)
	offset := (page - 1) * limit
	if err := query.Order("date DESC, id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}
	return &ExpensePage{Items: items, Total: int(total), Page: page, Limit: limit}, nil
}

func (r *ExpenseRepo) GetByID(ctx context.Context, storeID, id uint) (*model.Expense, error) {
	var expense model.Expense
	if err := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&expense).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &expense, nil
}

func (r *ExpenseRepo) Create(ctx context.Context, expense *model.Expense) (*model.Expense, error) {
	if err := r.db.WithContext(ctx).Create(expense).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return expense, nil
}

func (r *ExpenseRepo) Update(ctx context.Context, storeID, id uint, expense *model.Expense) (*model.Expense, error) {
	res := r.db.WithContext(ctx).Model(&model.Expense{}).Where("id = ? AND store_id = ?", id, storeID).Updates(map[string]any{
		"date": expense.Date, "category": expense.Category, "note": expense.Note,
		"amount": expense.Amount, "method": expense.Method, "source": expense.Source,
		"shift_id": expense.ShiftID,
	})
	if res.Error != nil {
		return nil, mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return nil, ErrNotFound
	}
	return r.GetByID(ctx, storeID, id)
}

func (r *ExpenseRepo) Delete(ctx context.Context, storeID, id uint) error {
	res := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", id, storeID).Delete(&model.Expense{})
	if res.Error != nil {
		return mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
