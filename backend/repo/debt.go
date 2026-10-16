package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/rayhanvito/nojpos/backend/model"
)

var (
	ErrDebtPaymentInvalid       = errors.New("pembayaran hutang-piutang tidak valid")
	ErrDebtPaymentExceedsRemain = errors.New("pembayaran melebihi sisa hutang-piutang")
)

type DebtRepo struct {
	db *gorm.DB
}

func NewDebtRepo(db *gorm.DB) *DebtRepo { return &DebtRepo{db: db} }

func (r *DebtRepo) ListReceivables(ctx context.Context, storeID uint, status string, overdueBefore *time.Time) ([]*model.Receivable, error) {
	query := r.db.WithContext(ctx).Where("store_id = ?", storeID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if overdueBefore != nil {
		query = query.Where("due_date < ? AND status <> ?", *overdueBefore, model.DebtPaid)
	}
	items := make([]*model.Receivable, 0)
	if err := query.Order("created_at DESC, id DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	if err := r.fillReceivables(ctx, storeID, items, false); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *DebtRepo) GetReceivable(ctx context.Context, storeID, id uint) (*model.Receivable, error) {
	var item model.Receivable
	if err := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&item).Error; err != nil {
		return nil, mapDBErr(err)
	}
	items := []*model.Receivable{&item}
	if err := r.fillReceivables(ctx, storeID, items, true); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *DebtRepo) PayReceivable(ctx context.Context, storeID, id uint, amount int64, method, actor string) (*model.Receivable, error) {
	if amount <= 0 || !model.ValidStandardPayMethod(method) {
		return nil, ErrDebtPaymentInvalid
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item model.Receivable
		query := tx.Where("id = ? AND store_id = ?", id, storeID)
		if tx.Dialector.Name() != "sqlite" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := query.First(&item).Error; err != nil {
			return mapDBErr(err)
		}
		var paid int64
		if err := tx.Model(&model.ReceivablePayment{}).
			Where("receivable_id = ? AND store_id = ?", id, storeID).
			Select("COALESCE(SUM(amount), 0)").Scan(&paid).Error; err != nil {
			return err
		}
		remaining := item.Amount - paid
		if amount > remaining {
			return ErrDebtPaymentExceedsRemain
		}
		payment := model.ReceivablePayment{StoreID: storeID, ReceivableID: id, Amount: amount, Method: method, CreatedBy: actor}
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

func (r *DebtRepo) fillReceivables(ctx context.Context, storeID uint, items []*model.Receivable, includePayments bool) error {
	ids := make([]uint, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	var customers []model.Customer
	if err := r.db.WithContext(ctx).Unscoped().Where("store_id = ? AND id IN ?", storeID, receivableCustomerIDs(items)).Find(&customers).Error; err != nil {
		return err
	}
	customerNames := make(map[uint]string, len(customers))
	for _, customer := range customers {
		customerNames[customer.ID] = customer.Name
	}
	var payments []model.ReceivablePayment
	query := r.db.WithContext(ctx).Where("store_id = ? AND receivable_id IN ?", storeID, ids)
	if !includePayments {
		query = query.Select("id, receivable_id, amount")
	}
	if err := query.Order("created_at ASC, id ASC").Find(&payments).Error; err != nil {
		return err
	}
	paid := make(map[uint]int64, len(items))
	paymentRows := make(map[uint][]model.ReceivablePayment, len(items))
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
				item.Payments = []model.ReceivablePayment{}
			}
		}
	}
	return nil
}

func receivableCustomerIDs(items []*model.Receivable) []uint {
	ids := make([]uint, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.CustomerID)
	}
	return ids
}

func (r *DebtRepo) ListPayables(ctx context.Context, storeID uint, status string) ([]*model.Payable, error) {
	query := r.db.WithContext(ctx).Where("store_id = ?", storeID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	items := make([]*model.Payable, 0)
	if err := query.Order("created_at DESC, id DESC").Find(&items).Error; err != nil {
		return nil, err
	}
	if err := r.fillPayables(ctx, storeID, items, false); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *DebtRepo) GetPayable(ctx context.Context, storeID, id uint) (*model.Payable, error) {
	var item model.Payable
	if err := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&item).Error; err != nil {
		return nil, mapDBErr(err)
	}
	items := []*model.Payable{&item}
	if err := r.fillPayables(ctx, storeID, items, true); err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *DebtRepo) PayPayable(ctx context.Context, storeID, id uint, amount int64, method, actor string) (*model.Payable, error) {
	if amount <= 0 || !model.ValidStandardPayMethod(method) {
		return nil, ErrDebtPaymentInvalid
	}
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var item model.Payable
		query := tx.Where("id = ? AND store_id = ?", id, storeID)
		if tx.Dialector.Name() != "sqlite" {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := query.First(&item).Error; err != nil {
			return mapDBErr(err)
		}
		var paid int64
		if err := tx.Model(&model.PayablePayment{}).
			Where("payable_id = ? AND store_id = ?", id, storeID).
			Select("COALESCE(SUM(amount), 0)").Scan(&paid).Error; err != nil {
			return err
		}
		remaining := item.Amount - paid
		if amount > remaining {
			return ErrDebtPaymentExceedsRemain
		}
		payment := model.PayablePayment{StoreID: storeID, PayableID: id, Amount: amount, Method: method, CreatedBy: actor}
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

func (r *DebtRepo) fillPayables(ctx context.Context, storeID uint, items []*model.Payable, includePayments bool) error {
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
	var suppliers []model.Supplier
	if err := r.db.WithContext(ctx).Unscoped().Where("store_id = ? AND id IN ?", storeID, supplierIDs).Find(&suppliers).Error; err != nil {
		return err
	}
	supplierNames := make(map[uint]string, len(suppliers))
	for _, supplier := range suppliers {
		supplierNames[supplier.ID] = supplier.Name
	}
	var orders []model.PurchaseOrder
	if err := r.db.WithContext(ctx).Unscoped().Where("store_id = ? AND id IN ?", storeID, poIDs).Find(&orders).Error; err != nil {
		return err
	}
	poNumbers := make(map[uint]string, len(orders))
	for _, order := range orders {
		poNumbers[order.ID] = order.Number
	}
	var payments []model.PayablePayment
	query := r.db.WithContext(ctx).Where("store_id = ? AND payable_id IN ?", storeID, ids)
	if !includePayments {
		query = query.Select("id, payable_id, amount")
	}
	if err := query.Order("created_at ASC, id ASC").Find(&payments).Error; err != nil {
		return err
	}
	paid := make(map[uint]int64, len(items))
	paymentRows := make(map[uint][]model.PayablePayment, len(items))
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
				item.Payments = []model.PayablePayment{}
			}
		}
	}
	return nil
}

func debtStatus(amount, paid int64) model.DebtStatus {
	if paid >= amount {
		return model.DebtPaid
	}
	if paid > 0 {
		return model.DebtPartiallyPaid
	}
	return model.DebtUnpaid
}
