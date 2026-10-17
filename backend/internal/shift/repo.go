package shift

import (
	"context"
	"errors"
	"math/big"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
	"github.com/rayhanvito/nojpos/backend/internal/staff"
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

func isNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// shiftExpense adalah baris privat untuk agregat pengeluaran laci di
// getShiftStats. Tipe kanonisnya tinggal di slice finance (BE-F6), dan
// finance mengimpor shift (kontrak Shifts ExpenseService) — shift tidak
// boleh mengimpor balik finance (siklus). Baris ini menunjuk tabel yang
// sama dan menyematkan database.BaseModel, jadi cakupan soft-delete kueri
// SUM-nya identik dengan Model(&model.Expense{}) sebelumnya.
type shiftExpense struct {
	database.BaseModel
	Amount int64 `gorm:"not null"`
}

func (shiftExpense) TableName() string { return "expenses" }

// loadLoc, parseLocalDay, roundRatioHalfUp, dan uintPointer adalah
// salinan privat slice dari helper repo lama (repo/report.go dan
// repo/transaction.go) — badan verbatim.
func loadLoc(tz string) *time.Location {
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc, err = time.LoadLocation("Asia/Makassar")
		if err != nil {
			loc = time.UTC
		}
	}
	return loc
}

// parseLocalDay memvalidasi "YYYY-MM-DD" secara ketat dan mengembalikan
// tengah malam hari itu di loc. ok=false berarti param kosong/tak valid
// dan pemanggil harus memakai perilaku lama (hari ini).
func parseLocalDay(s string, loc *time.Location) (time.Time, bool) {
	if len(s) != 10 {
		return time.Time{}, false
	}
	parsed, err := time.Parse("2006-01-02", s)
	if err != nil || parsed.Format("2006-01-02") != s {
		return time.Time{}, false
	}
	return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, loc), true
}

func roundRatioHalfUp(value, numerator, denominator int64) int64 {
	if denominator <= 0 || value <= 0 || numerator <= 0 {
		return 0
	}
	product := new(big.Int).Mul(big.NewInt(value), big.NewInt(numerator))
	product.Add(product, big.NewInt(denominator/2))
	return product.Div(product, big.NewInt(denominator)).Int64()
}

func uintPointer(value uint) *uint { return &value }

// ShiftStats adalah agregat statistik shift dari repo (dipindah
// dari repo/shift.go bersama Repo).
type ShiftStats struct {
	Sales           int64
	TrxCount        int64
	SalesByMethod   map[string]int64
	RefundTotal     int64
	CashRefundTotal int64
	CashExpenses    int64
	Items           []trxItem
	Transactions    []trx
}

// Repo adalah repositori shift (ShiftRepo lama di repo/shift.go);
// metode-metodenya verbatim, akses DB lewat DB(ctx).
type Repo struct {
	database.BaseRepo
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *Repo) GetActiveShift(ctx context.Context, storeID, outletID, cashierID uint) (*CashierShift, error) {
	var s CashierShift
	err := r.DB(ctx).Where("store_id = ? AND outlet_id = ? AND cashier_id = ? AND closed_at IS NULL", storeID, outletID, cashierID).Order("started_at DESC").First(&s).Error
	if err != nil {
		return nil, mapDBErr(err)
	}
	return &s, nil
}

func (r *Repo) StartShift(ctx context.Context, storeID, outletID, cashierID uint, cashierName string, openingCash int64) (*CashierShift, error) {
	var shift CashierShift
	err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.lockStoreTransactions(tx, storeID); err != nil {
			return err
		}
		if cashierID != 0 {
			var cashier staff.Cashier
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ? AND store_id = ?", cashierID, storeID).First(&cashier).Error; err != nil {
				return mapDBErr(err)
			}
			if cashier.OutletID == nil || *cashier.OutletID != outletID {
				return ErrCashierOutletChanged
			}
		}

		var active CashierShift
		query := tx.Where("store_id = ? AND outlet_id = ? AND cashier_id = ? AND closed_at IS NULL", storeID, outletID, cashierID)
		if r.isPostgres(ctx) {
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

		shift = CashierShift{
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

func (r *Repo) CloseShift(ctx context.Context, storeID, outletID, cashierID uint, closingCash int64, note *string, closedByName string) (*CashierShift, *ShiftStats, error) {
	var closed CashierShift
	var stats *ShiftStats
	cash := closingCash
	err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.lockStoreTransactions(tx, storeID); err != nil {
			return err
		}
		query := tx.Where("store_id = ? AND outlet_id = ? AND cashier_id = ? AND closed_at IS NULL", storeID, outletID, cashierID)
		if r.isPostgres(ctx) {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		var active CashierShift
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

func (r *Repo) ForceCloseShift(ctx context.Context, storeID, shiftID uint, closingCash *int64, note *string, closedByName string) (*CashierShift, *ShiftStats, error) {
	var closed CashierShift
	var stats *ShiftStats
	err := r.DB(ctx).Transaction(func(tx *gorm.DB) error {
		if err := r.lockStoreTransactions(tx, storeID); err != nil {
			return err
		}
		query := tx.Where("id = ? AND store_id = ?", shiftID, storeID)
		if r.isPostgres(ctx) {
			query = query.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		var active CashierShift
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

func (r *Repo) finishShift(ctx context.Context, tx *gorm.DB, shift *CashierShift, closedAt time.Time, closingCash *int64, note *string, closedByName string) (*ShiftStats, error) {
	stats, err := r.getShiftStats(ctx, tx, *shift, closedAt)
	if err != nil {
		return nil, err
	}
	expectedCash := shift.OpeningCash + stats.SalesByMethod["Cash"] - stats.CashRefundTotal - stats.CashExpenses
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

func (r *Repo) lockStoreTransactions(tx *gorm.DB, storeID uint) error {
	if tx.Dialector.Name() != "postgres" {
		return nil
	}
	return tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", strconv.FormatUint(uint64(storeID), 10)).Error
}

func (r *Repo) ListStoreShifts(ctx context.Context, storeID uint, outletID *uint, from string, to string) ([]*CashierShift, error) {
	if err := r.validateOutletFilter(ctx, storeID, outletID); err != nil {
		return nil, err
	}
	var shifts []*CashierShift
	query := r.DB(ctx).Where("store_id = ?", storeID)
	if outletID != nil {
		query = query.Where("outlet_id = ?", *outletID)
	}
	if fs, ts := strings.TrimSpace(from), strings.TrimSpace(to); fs != "" && ts != "" {
		// Rentang dari pemilih periode UI: inklusif per hari kalender lokal toko,
		// diterapkan pada waktu mulai shift (pola sama dengan daftar transaksi).
		tz := "Asia/Makassar"
		var st outlet.Store
		if err := r.DB(ctx).Where("id = ?", storeID).First(&st).Error; err == nil && st.Timezone != "" {
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

func (r *Repo) GetShiftByID(ctx context.Context, storeID, shiftID uint, outletID *uint) (*CashierShift, error) {
	if err := r.validateOutletFilter(ctx, storeID, outletID); err != nil {
		return nil, err
	}
	var shift CashierShift
	query := r.DB(ctx).Where("id = ? AND store_id = ?", shiftID, storeID)
	if outletID != nil {
		query = query.Where("outlet_id = ?", *outletID)
	}
	if err := query.First(&shift).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return &shift, nil
}

func (r *Repo) GetShiftStats(ctx context.Context, shift CashierShift, endTime time.Time) (*ShiftStats, error) {
	return r.getShiftStats(ctx, r.DB(ctx), shift, endTime)
}

func (r *Repo) getShiftStats(ctx context.Context, db *gorm.DB, shift CashierShift, endTime time.Time) (*ShiftStats, error) {
	stats := &ShiftStats{SalesByMethod: map[string]int64{}, Items: []trxItem{}, Transactions: []trx{}}
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
	transactionsByID := make(map[uint]trx, len(stats.Transactions))
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
	if err := db.WithContext(ctx).Model(&shiftExpense{}).Where("shift_id = ?", shift.ID).
		Select("COALESCE(SUM(amount), 0)").Scan(&stats.CashExpenses).Error; err != nil {
		return nil, err
	}
	return stats, nil
}

func (r *Repo) shiftRefunds(ctx context.Context, db *gorm.DB, shift CashierShift, endTime time.Time, transactions map[uint]trx, stats *ShiftStats) error {
	query := db.WithContext(ctx).Model(&refund{}).Select("refunds.*").
		Joins("JOIN transactions ON transactions.id = refunds.trx_id").
		Where("refunds.store_id = ? AND transactions.store_id = ? AND transactions.cashier_id = ? AND transactions.created_at >= ? AND transactions.created_at <= ? AND refunds.created_at >= ? AND refunds.created_at <= ?",
			shift.StoreID, shift.StoreID, shift.CashierID, shift.StartedAt, endTime, shift.StartedAt, endTime)
	if shift.OutletID == nil {
		query = query.Where("transactions.outlet_id IS NULL")
	} else {
		query = query.Where("transactions.outlet_id = ?", *shift.OutletID)
	}
	var refunds []refund
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
	var items []transactionItem
	if err := db.WithContext(ctx).Where("trx_id IN ?", trxIDs).Order("id ASC").Find(&items).Error; err != nil {
		return err
	}
	itemsByTrx := make(map[uint][]trxItem, len(trxIDs))
	for _, item := range items {
		itemsByTrx[item.TrxID] = append(itemsByTrx[item.TrxID], trxItem{
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
		if transaction.Method == "Cash" {
			stats.CashRefundTotal += amount
		}
	}
	return nil
}

func (r *Repo) ListShiftTransactions(ctx context.Context, storeID uint, shift CashierShift, page, limit int) ([]*trx, int64, error) {
	query := r.DB(ctx).Where("store_id = ? AND cashier_id = ? AND status IN ? AND created_at >= ? AND created_at <= ?",
		storeID, shift.CashierID, completedShiftStatuses, shift.StartedAt, shiftEndTime(shift))
	if shift.OutletID == nil {
		query = query.Where("outlet_id IS NULL")
	} else {
		query = query.Where("outlet_id = ?", *shift.OutletID)
	}
	var total int64
	if err := query.Model(&trx{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	items := make([]*trx, 0, limit)
	if err := query.Order("created_at ASC, id ASC").Offset((page - 1) * limit).Limit(limit).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *Repo) validateOutletFilter(ctx context.Context, storeID uint, outletID *uint) error {
	if outletID == nil {
		return nil
	}
	var count int64
	if err := r.DB(ctx).Model(&outlet.Outlet{}).Where("id = ? AND store_id = ?", *outletID, storeID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrOutletNotFound
	}
	return nil
}

func shiftEndTime(shift CashierShift) time.Time {
	if shift.ClosedAt != nil {
		return *shift.ClosedAt
	}
	return time.Now().UTC()
}

var completedShiftStatuses = []trxStatus{trxCompleted, trxRefunded, trxCompletedOfflineShortage}

func (r *Repo) isPostgres(ctx context.Context) bool {
	return r.DB(ctx).Dialector.Name() == "postgres"
}
