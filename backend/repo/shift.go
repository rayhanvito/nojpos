package repo

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/rayhanvito/nojpos/backend/model"
)

var ErrShiftClosed = errors.New("shift sudah ditutup")
var ErrShiftAlreadyOpen = errors.New("shift masih aktif")

type ShiftStats struct {
	Sales           int64
	TrxCount        int64
	SalesByMethod   map[string]int64
	RefundTotal     int64
	CashRefundTotal int64
	CashExpenses    int64
	Items           []model.TrxItem
	Transactions    []model.Trx
}

type ShiftRepo struct {
	db *gorm.DB
}

func NewShiftRepo(db *gorm.DB) *ShiftRepo {
	return &ShiftRepo{db: db}
}

func (r *ShiftRepo) GetActiveShift(ctx context.Context, storeID, outletID, cashierID uint) (*model.CashierShift, error) {
	var s model.CashierShift
	err := r.db.WithContext(ctx).Where("store_id = ? AND outlet_id = ? AND cashier_id = ? AND closed_at IS NULL", storeID, outletID, cashierID).Order("started_at DESC").First(&s).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	return &s, nil
}

func (r *ShiftRepo) StartShift(ctx context.Context, storeID, outletID, cashierID uint, cashierName string, openingCash int64) (*model.CashierShift, error) {
	var shift model.CashierShift
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.lockStoreTransactions(tx, storeID); err != nil {
			return err
		}
		if cashierID != 0 {
			var cashier model.Cashier
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND store_id = ?", cashierID, storeID).First(&cashier).Error; err != nil {
				return mapDBErr(err)
			}
			if cashier.OutletID == nil || *cashier.OutletID != outletID {
				return ErrCashierOutletChanged
			}
		}

		var active model.CashierShift
		query := tx.Where("store_id = ? AND outlet_id = ? AND cashier_id = ? AND closed_at IS NULL", storeID, outletID, cashierID)
		if r.isPostgres() {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		err := query.First(&active).Error
		if err != nil && !isNotFound(err) {
			return err
		}
		now := time.Now().UTC()
		if err == nil {
			// Shift aktif TIDAK ditimpa diam-diam: menutup shift wajib lewat
			// hitungan kas. Tolak agar tidak ada shift yang tertutup tanpa
			// closing_cash dan selisih (audit 2026-10-16).
			return ErrShiftAlreadyOpen
		}

		shift = model.CashierShift{
			StoreID:     storeID,
			OutletID:    uintPointer(outletID),
			CashierID:   cashierID,
			CashierName: cashierName,
			StartedAt:   now,
			OpeningCash: openingCash,
			ClosedAt:    nil,
		}
		return tx.Create(&shift).Error
	})
	if err != nil {
		return nil, mapDBErr(err)
	}
	return &shift, nil
}

func (r *ShiftRepo) CloseShift(ctx context.Context, storeID, outletID, cashierID uint, closingCash int64, note *string, closedByName string) (*model.CashierShift, *ShiftStats, error) {
	var closed model.CashierShift
	var stats *ShiftStats
	cash := closingCash
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.lockStoreTransactions(tx, storeID); err != nil {
			return err
		}
		query := tx.Where("store_id = ? AND outlet_id = ? AND cashier_id = ? AND closed_at IS NULL", storeID, outletID, cashierID)
		if r.isPostgres() {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		var active model.CashierShift
		if err := query.First(&active).Error; err != nil {
			return mapDBErr(err)
		}
		var err error
		stats, err = r.finishShift(ctx, tx, &active, time.Now().UTC(), &cash, note, closedByName)
		if err != nil {
			return err
		}
		closed = active
		return nil
	})
	if err != nil {
		return nil, nil, mapDBErr(err)
	}
	return &closed, stats, nil
}

func (r *ShiftRepo) ForceCloseShift(ctx context.Context, storeID, shiftID uint, closingCash *int64, note *string, closedByName string) (*model.CashierShift, *ShiftStats, error) {
	var closed model.CashierShift
	var stats *ShiftStats
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.lockStoreTransactions(tx, storeID); err != nil {
			return err
		}
		query := tx.Where("id = ? AND store_id = ?", shiftID, storeID)
		if r.isPostgres() {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		var active model.CashierShift
		if err := query.First(&active).Error; err != nil {
			return mapDBErr(err)
		}
		if active.ClosedAt != nil {
			return ErrShiftClosed
		}
		var err error
		stats, err = r.finishShift(ctx, tx, &active, time.Now().UTC(), closingCash, note, closedByName)
		if err != nil {
			return err
		}
		closed = active
		return nil
	})
	if err != nil {
		return nil, nil, mapDBErr(err)
	}
	return &closed, stats, nil
}

func (r *ShiftRepo) finishShift(ctx context.Context, tx *gorm.DB, shift *model.CashierShift, closedAt time.Time, closingCash *int64, note *string, closedByName string) (*ShiftStats, error) {
	stats, err := r.getShiftStats(ctx, tx, *shift, closedAt)
	if err != nil {
		return nil, err
	}
	expectedCash := shift.OpeningCash + stats.SalesByMethod[string(model.PayCash)] - stats.CashRefundTotal - stats.CashExpenses
	shift.ClosedAt = &closedAt
	shift.ExpectedCash = &expectedCash
	shift.ClosingCashActual = closingCash
	shift.CashVariance = nil
	if closingCash != nil {
		variance := *closingCash - expectedCash
		shift.CashVariance = &variance
	}
	shift.CloseNote = note
	shift.ClosedByName = nil
	if closedByName != "" {
		shift.ClosedByName = &closedByName
	}
	if err := tx.Save(shift).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *ShiftRepo) lockStoreTransactions(tx *gorm.DB, storeID uint) error {
	if !r.isPostgres() {
		return nil
	}
	return tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", strconv.FormatUint(uint64(storeID), 10)).Error
}

func (r *ShiftRepo) ListStoreShifts(ctx context.Context, storeID uint, outletID *uint, from string, to string) ([]*model.CashierShift, error) {
	if err := r.validateOutletFilter(ctx, storeID, outletID); err != nil {
		return nil, err
	}
	var shifts []*model.CashierShift
	query := r.db.WithContext(ctx).Where("store_id = ?", storeID)
	if outletID != nil {
		query = query.Where("outlet_id = ?", *outletID)
	}
	if fs, ts := strings.TrimSpace(from), strings.TrimSpace(to); fs != "" && ts != "" {
		// Rentang dari pemilih periode UI: inklusif per hari kalender lokal toko,
		// diterapkan pada waktu mulai shift (pola sama dengan daftar transaksi).
		tz := "Asia/Makassar"
		var st model.Store
		if err := r.db.WithContext(ctx).Where("id = ?", storeID).First(&st).Error; err == nil && st.Timezone != "" {
			tz = st.Timezone
		}
		loc := loadLoc(tz)
		if f, ok := parseLocalDay(fs, loc); ok {
			if t, ok2 := parseLocalDay(ts, loc); ok2 && !t.Before(f) {
				query = query.Where("started_at >= ? AND started_at < ?", f.UTC(), t.AddDate(0, 0, 1).UTC())
			}
		}
	}
	err := query.Order("started_at DESC").Find(&shifts).Error
	if err != nil {
		return nil, err
	}
	return shifts, nil
}

func (r *ShiftRepo) GetShiftByID(ctx context.Context, storeID, shiftID uint, outletID *uint) (*model.CashierShift, error) {
	if err := r.validateOutletFilter(ctx, storeID, outletID); err != nil {
		return nil, err
	}
	var shift model.CashierShift
	query := r.db.WithContext(ctx).Where("id = ? AND store_id = ?", shiftID, storeID)
	if outletID != nil {
		query = query.Where("outlet_id = ?", *outletID)
	}
	if err := query.First(&shift).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &shift, nil
}

func (r *ShiftRepo) GetShiftStats(ctx context.Context, shift model.CashierShift, endTime time.Time) (*ShiftStats, error) {
	return r.getShiftStats(ctx, r.db, shift, endTime)
}

func (r *ShiftRepo) getShiftStats(ctx context.Context, db *gorm.DB, shift model.CashierShift, endTime time.Time) (*ShiftStats, error) {
	stats := &ShiftStats{SalesByMethod: map[string]int64{}, Items: []model.TrxItem{}, Transactions: []model.Trx{}}
	query := db.WithContext(ctx).Where("store_id = ? AND cashier_id = ? AND status IN ? AND created_at >= ? AND created_at <= ?",
		shift.StoreID, shift.CashierID, completedShiftStatuses, shift.StartedAt, endTime)
	if shift.OutletID == nil {
		query = query.Where("outlet_id IS NULL")
	} else {
		query = query.Where("outlet_id = ?", *shift.OutletID)
	}
	if err := query.Order("created_at ASC, id ASC").Find(&stats.Transactions).Error; err != nil {
		return nil, err
	}
	stats.TrxCount = int64(len(stats.Transactions))
	transactionsByID := make(map[uint]model.Trx, len(stats.Transactions))
	for _, transaction := range stats.Transactions {
		stats.Sales += transaction.Total
		stats.SalesByMethod[transaction.Method] += transaction.Total
		stats.Items = append(stats.Items, transaction.Items...)
		transactionsByID[transaction.ID] = transaction
	}
	if err := r.shiftRefunds(ctx, db, shift, endTime, transactionsByID, stats); err != nil {
		return nil, err
	}
	// Pengeluaran laci yang tertaut ke shift ini mengurangi kas fisik.
	if err := db.WithContext(ctx).Model(&model.Expense{}).Where("shift_id = ?", shift.ID).
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.CashExpenses).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *ShiftRepo) shiftRefunds(ctx context.Context, db *gorm.DB, shift model.CashierShift, endTime time.Time, transactions map[uint]model.Trx, stats *ShiftStats) error {
	query := db.WithContext(ctx).Model(&model.Refund{}).Select("refunds.*").
		Joins("JOIN transactions ON transactions.id = refunds.trx_id").
		Where("refunds.store_id = ? AND transactions.store_id = ? AND transactions.cashier_id = ? AND transactions.created_at >= ? AND transactions.created_at <= ? AND refunds.created_at >= ? AND refunds.created_at <= ?",
			shift.StoreID, shift.StoreID, shift.CashierID, shift.StartedAt, endTime, shift.StartedAt, endTime)
	if shift.OutletID == nil {
		query = query.Where("transactions.outlet_id IS NULL")
	} else {
		query = query.Where("transactions.outlet_id = ?", *shift.OutletID)
	}
	var refunds []model.Refund
	if err := query.Order("refunds.created_at ASC, refunds.id ASC").Find(&refunds).Error; err != nil {
		return err
	}
	if len(refunds) == 0 {
		return nil
	}

	trxIDs := make([]uint, 0, len(refunds))
	for _, refund := range refunds {
		trxIDs = append(trxIDs, refund.TrxID)
	}
	var items []model.TransactionItem
	if err := db.WithContext(ctx).Where("trx_id IN ?", trxIDs).Order("id ASC").Find(&items).Error; err != nil {
		return err
	}
	itemsByTrx := make(map[uint][]model.TrxItem, len(trxIDs))
	for _, item := range items {
		itemsByTrx[item.TrxID] = append(itemsByTrx[item.TrxID], model.TrxItem{
			ProductID: item.ProductID, VariantID: item.VariantID, Price: item.Price,
		})
	}

	for _, refund := range refunds {
		transaction, ok := transactions[refund.TrxID]
		if !ok {
			continue
		}
		trxItems := itemsByTrx[refund.TrxID]
		if len(trxItems) == 0 {
			trxItems = transaction.Items
		}
		var amount int64
		for _, item := range trxItems {
			qty := refund.Items[item.ProductID]
			if item.VariantID != nil {
				qty = refund.VariantItems[*item.VariantID]
			}
			if qty <= 0 {
				continue
			}
			itemValue := item.Price * int64(qty)
			amount += itemValue + roundRatioHalfUp(transaction.Tax, itemValue, transaction.Subtotal-transaction.Discount-transaction.PromoDiscount-transaction.VoucherDiscount)
		}
		stats.RefundTotal += amount
		if transaction.Method == string(model.PayCash) {
			stats.CashRefundTotal += amount
		}
	}
	return nil
}

func (r *ShiftRepo) ListShiftTransactions(ctx context.Context, storeID uint, shift model.CashierShift, page, limit int) ([]*model.Trx, int64, error) {
	query := r.db.WithContext(ctx).Where("store_id = ? AND cashier_id = ? AND status IN ? AND created_at >= ? AND created_at <= ?",
		storeID, shift.CashierID, completedShiftStatuses, shift.StartedAt, shiftEndTime(shift))
	if shift.OutletID == nil {
		query = query.Where("outlet_id IS NULL")
	} else {
		query = query.Where("outlet_id = ?", *shift.OutletID)
	}
	var total int64
	if err := query.Model(&model.Trx{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*model.Trx, 0, limit)
	if err := query.Order("created_at ASC, id ASC").Offset((page - 1) * limit).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *ShiftRepo) validateOutletFilter(ctx context.Context, storeID uint, outletID *uint) error {
	if outletID == nil {
		return nil
	}
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Outlet{}).Where("id = ? AND store_id = ?", *outletID, storeID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrOutletNotFound
	}
	return nil
}

func shiftEndTime(shift model.CashierShift) time.Time {
	if shift.ClosedAt != nil {
		return *shift.ClosedAt
	}
	return time.Now().UTC()
}

var completedShiftStatuses = []model.TrxStatus{model.TrxCompleted, model.TrxRefunded, model.TrxCompletedOfflineShortage}

func (r *ShiftRepo) isPostgres() bool {
	return r.db.Dialector.Name() == "postgres"
}
