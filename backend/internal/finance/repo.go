package finance

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/rayhanvito/nojpos/backend/internal/customer"
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

// DebtRepo adalah repositori utang-piutang (DebtRepo lama di
// repo/debt.go); metode-metodenya verbatim, akses DB lewat DB(ctx).
type DebtRepo struct {
	database.BaseRepo
}

func NewDebtRepo(db *gorm.DB) *DebtRepo {
	return &DebtRepo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *DebtRepo) ListReceivables(ctx context.Context, storeID uint, status string, overdueBefore *time.Time) ([]*Receivable, error) {
	query := r.DB(ctx).Where("store_id = ?", storeID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if overdueBefore != nil {
		query = query.Where("due_date < ? AND status <> ?", *overdueBefore, DebtPaid)
	}
	items := make([]*Receivable, 0)
	if err := query.Order("created_at DESC, id DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	if err := r.fillReceivables(ctx, storeID, items, false); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *DebtRepo) GetReceivable(ctx context.Context, storeID, id uint) (*Receivable, error) {
	var item Receivable
	if err := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&item).Error; err != nil {
		return nil, mapDBErr(err)
	}
	items := []*Receivable{&item}
	if err := r.fillReceivables(ctx, storeID, items, true); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *DebtRepo) PayReceivable(ctx context.Context, storeID, id uint, amount int64, method, actor string) (*Receivable, error) {
	if amount <= 0 || !validStandardPayMethod(method) {
		return nil, ErrDebtPaymentInvalid
	}
	err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var item Receivable
		query := tx.Where("id = ? AND store_id = ?", id, storeID)
		if tx.Dialector.Name() != "sqlite" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := query.First(&item).Error; err != nil {
			return mapDBErr(err)
		}
		var paid int64
		if err := tx.Model(&ReceivablePayment{}).
			Where("receivable_id = ? AND store_id = ?", id, storeID).
			Select("COALESCE(SUM(amount), 0)").Scan(&paid).Error; err != nil {
			return err
		}
		remaining := item.Amount - paid
		if amount > remaining {
			return ErrDebtPaymentExceedsRemain
		}
		payment := ReceivablePayment{StoreID: storeID, ReceivableID: id, Amount: amount, Method: method, CreatedBy: actor}
		if err := tx.Create(&payment).Error; err != nil {
			return err
		}
		return tx.Model(&item).Update("status", debtStatus(item.Amount, paid+amount)).Error
	})
	if err != nil {
		return nil, err
	}
	return r.GetReceivable(ctx, storeID, id)
}

func (r *DebtRepo) fillReceivables(ctx context.Context, storeID uint, items []*Receivable, includePayments bool) error {
	ids := make([]uint, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	var customers []customer.Customer
	if err := r.DB(ctx).Unscoped().Where("store_id = ? AND id IN ?", storeID, receivableCustomerIDs(items)).Find(&customers).Error; err != nil {
		return err
	}
	customerNames := make(map[uint]string, len(customers))
	for _, customer := range customers {
		customerNames[customer.ID] = customer.Name
	}
	var payments []ReceivablePayment
	query := r.DB(ctx).Where("store_id = ? AND receivable_id IN ?", storeID, ids)
	if !includePayments {
		query = query.Select("id, receivable_id, amount")
	}
	if err := query.Order("created_at ASC, id ASC").Find(&payments).Error; err != nil {
		return err
	}
	paid := make(map[uint]int64, len(items))
	paymentRows := make(map[uint][]ReceivablePayment, len(items))
	for _, payment := range payments {
		paid[payment.ReceivableID] += payment.Amount
		if includePayments {
			paymentRows[payment.ReceivableID] = append(paymentRows[payment.ReceivableID], payment)
		}
	}
	for _, item := range items {
		item.Remaining = item.Amount - paid[item.ID]
		item.CustomerName = customerNames[item.CustomerID]
		item.InvoiceNumber = fmt.Sprintf("TRX-%05d", item.TransactionID)
		if includePayments {
			item.Payments = paymentRows[item.ID]
			if item.Payments == nil {
				item.Payments = []ReceivablePayment{}
			}
		}
	}
	return nil
}

func receivableCustomerIDs(items []*Receivable) []uint {
	ids := make([]uint, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.CustomerID)
	}
	return ids
}

func (r *DebtRepo) ListPayables(ctx context.Context, storeID uint, status string) ([]*Payable, error) {
	query := r.DB(ctx).Where("store_id = ?", storeID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	items := make([]*Payable, 0)
	if err := query.Order("created_at DESC, id DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	if err := r.fillPayables(ctx, storeID, items, false); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *DebtRepo) GetPayable(ctx context.Context, storeID, id uint) (*Payable, error) {
	var item Payable
	if err := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&item).Error; err != nil {
		return nil, mapDBErr(err)
	}
	items := []*Payable{&item}
	if err := r.fillPayables(ctx, storeID, items, true); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *DebtRepo) PayPayable(ctx context.Context, storeID, id uint, amount int64, method, actor string) (*Payable, error) {
	if amount <= 0 || !validStandardPayMethod(method) {
		return nil, ErrDebtPaymentInvalid
	}
	err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		var item Payable
		query := tx.Where("id = ? AND store_id = ?", id, storeID)
		if tx.Dialector.Name() != "sqlite" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := query.First(&item).Error; err != nil {
			return mapDBErr(err)
		}
		var paid int64
		if err := tx.Model(&PayablePayment{}).
			Where("payable_id = ? AND store_id = ?", id, storeID).
			Select("COALESCE(SUM(amount), 0)").Scan(&paid).Error; err != nil {
			return err
		}
		remaining := item.Amount - paid
		if amount > remaining {
			return ErrDebtPaymentExceedsRemain
		}
		payment := PayablePayment{StoreID: storeID, PayableID: id, Amount: amount, Method: method, CreatedBy: actor}
		if err := tx.Create(&payment).Error; err != nil {
			return err
		}
		return tx.Model(&item).Update("status", debtStatus(item.Amount, paid+amount)).Error
	})
	if err != nil {
		return nil, err
	}
	return r.GetPayable(ctx, storeID, id)
}

// payableSupplierRow dan payableOrderRow adalah baris baca privat slice
// untuk nama supplier dan nomor PO pada fillPayables. Tipe kanonisnya
// tinggal di slice purchasing, dan purchasing mengimpor finance
// (pembuatan Payable) — finance tidak boleh mengimpor balik purchasing
// (siklus), jadi kueri nama/nomor ini memakai baris lokal dengan nama
// tabel dan semantik (Unscoped) yang sama persis seperti sebelumnya.
type payableSupplierRow struct {
	database.BaseModel
	StoreID uint   `gorm:"not null;index"`
	Name    string `gorm:"not null"`
}

func (payableSupplierRow) TableName() string { return "suppliers" }

type payableOrderRow struct {
	database.BaseModel
	StoreID uint   `gorm:"not null;index"`
	Number  string `gorm:"not null"`
}

func (payableOrderRow) TableName() string { return "purchase_orders" }

func (r *DebtRepo) fillPayables(ctx context.Context, storeID uint, items []*Payable, includePayments bool) error {
	ids := make([]uint, 0, len(items))
	supplierIDs := make([]uint, 0, len(items))
	poIDs := make([]uint, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
		supplierIDs = append(supplierIDs, item.SupplierID)
		poIDs = append(poIDs, item.PurchaseOrderID)
	}
	if len(ids) == 0 {
		return nil
	}
	var suppliers []payableSupplierRow
	if err := r.DB(ctx).Unscoped().Where("store_id = ? AND id IN ?", storeID, supplierIDs).Find(&suppliers).Error; err != nil {
		return err
	}
	supplierNames := make(map[uint]string, len(suppliers))
	for _, supplier := range suppliers {
		supplierNames[supplier.ID] = supplier.Name
	}
	var orders []payableOrderRow
	if err := r.DB(ctx).Unscoped().Where("store_id = ? AND id IN ?", storeID, poIDs).Find(&orders).Error; err != nil {
		return err
	}
	poNumbers := make(map[uint]string, len(orders))
	for _, order := range orders {
		poNumbers[order.ID] = order.Number
	}
	var payments []PayablePayment
	query := r.DB(ctx).Where("store_id = ? AND payable_id IN ?", storeID, ids)
	if !includePayments {
		query = query.Select("id, payable_id, amount")
	}
	if err := query.Order("created_at ASC, id ASC").Find(&payments).Error; err != nil {
		return err
	}
	paid := make(map[uint]int64, len(items))
	paymentRows := make(map[uint][]PayablePayment, len(items))
	for _, payment := range payments {
		paid[payment.PayableID] += payment.Amount
		if includePayments {
			paymentRows[payment.PayableID] = append(paymentRows[payment.PayableID], payment)
		}
	}
	for _, item := range items {
		item.Remaining = item.Amount - paid[item.ID]
		item.SupplierName = supplierNames[item.SupplierID]
		item.PurchaseOrderNumber = poNumbers[item.PurchaseOrderID]
		if includePayments {
			item.Payments = paymentRows[item.ID]
			if item.Payments == nil {
				item.Payments = []PayablePayment{}
			}
		}
	}
	return nil
}

func debtStatus(amount, paid int64) DebtStatus {
	if paid >= amount {
		return DebtPaid
	}
	if paid > 0 {
		return DebtPartiallyPaid
	}
	return DebtUnpaid
}

// ExpenseFilter dan ExpensePage dipindah verbatim dari repo/expense.go
// (BE-F6).
type ExpenseFilter struct {
	From     string
	To       string
	Category string
	Q        string
	Page     int
	Limit    int
}

type ExpensePage struct {
	Items []*Expense `json:"items"`
	Total int        `json:"total"`
	Page  int        `json:"page"`
	Limit int        `json:"limit"`
}

// ExpenseRepo adalah repositori pengeluaran operasional (ExpenseRepo
// lama di repo/expense.go); metode-metodenya verbatim, akses DB lewat
// DB(ctx).
type ExpenseRepo struct {
	database.BaseRepo
}

func NewExpenseRepo(db *gorm.DB) *ExpenseRepo {
	return &ExpenseRepo{BaseRepo: database.NewBaseRepo(db)}
}

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

	query := r.DB(ctx).Model(&Expense{}).Where("store_id = ?", storeID)
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

	items := make([]*Expense, 0)
	offset := (page - 1) * limit
	if err := query.Order("date DESC, id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}
	return &ExpensePage{Items: items, Total: int(total), Page: page, Limit: limit}, nil
}

func (r *ExpenseRepo) GetByID(ctx context.Context, storeID, id uint) (*Expense, error) {
	var expense Expense
	if err := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&expense).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &expense, nil
}

func (r *ExpenseRepo) Create(ctx context.Context, expense *Expense) (*Expense, error) {
	if err := r.DB(ctx).Create(expense).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return expense, nil
}

func (r *ExpenseRepo) Update(ctx context.Context, storeID, id uint, expense *Expense) (*Expense, error) {
	res := r.DB(ctx).Model(&Expense{}).Where("id = ? AND store_id = ?", id, storeID).Updates(map[string]any{
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
	res := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).Delete(&Expense{})
	if res.Error != nil {
		return mapDBErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
