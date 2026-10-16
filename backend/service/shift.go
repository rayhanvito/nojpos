package service

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/adrr-dev/openPOS/backend/model"
	"github.com/adrr-dev/openPOS/backend/repo"
)

var (
	ErrNoActiveShift        = errors.New("tidak ada shift yang berjalan")
	ErrCashierOutletChanged = errors.New("outlet kasir sudah berubah")
	ErrShiftNotFound        = errors.New("shift tidak ditemukan")
	ErrShiftAlreadyClosed   = errors.New("shift sudah ditutup")
	ErrShiftAlreadyOpen     = errors.New("shift masih aktif")
)

type ShiftService struct {
	shifts   ShiftRepository
	cashiers CashierRepository
	stores   StoreRepository
}

type ShiftSummary struct {
	ID                uint             `json:"id"`
	OutletID          *uint            `json:"outlet_id,omitempty"`
	CashierName       string           `json:"cashier_name"`
	ClosedByName      *string          `json:"closed_by_name"`
	StartedAt         time.Time        `json:"started_at"`
	ClosedAt          *time.Time       `json:"closed_at"`
	OpeningCash       int64            `json:"opening_cash"`
	SalesByMethod     map[string]int64 `json:"sales_by_method"`
	RefundTotal       int64            `json:"refund_total"`
	Sales             int64            `json:"sales"`
	TrxCount          int64            `json:"trx_count"`
	ExpectedCash      int64            `json:"expected_cash"`
	CashExpenses      int64            `json:"cash_expenses"`
	ClosingCashActual *int64           `json:"closing_cash_actual"`
	CashVariance      *int64           `json:"cash_variance"`
	CloseNote         *string          `json:"close_note"`
}

type ShiftTransactionPage struct {
	Items []*model.Trx `json:"items"`
	Page  int          `json:"page"`
	Limit int          `json:"limit"`
	Total int64        `json:"total"`
}

type ShiftDetail struct {
	Shift        ShiftSummary         `json:"shift"`
	Transactions ShiftTransactionPage `json:"transactions"`
	Hourly       []ShiftHourPoint     `json:"hourly"`
	TopProducts  []ShiftTopProduct    `json:"top_products"`
}

type ShiftHourPoint struct {
	Hour  int   `json:"hour"`
	Omzet int64 `json:"omzet"`
}

type ShiftTopProduct struct {
	ProductID uint   `json:"product_id"`
	Name      string `json:"name"`
	Qty       int    `json:"qty"`
}

func NewShiftService(shifts ShiftRepository, cashiers CashierRepository, stores StoreRepository) *ShiftService {
	return &ShiftService{shifts: shifts, cashiers: cashiers, stores: stores}
}

func ShiftPaymentMethods() []string {
	return []string{
		string(model.PayCash), string(model.PayBankTransfer), string(model.PayQRIS),
		string(model.PayEWallet), string(model.PayCard), string(model.PayCredit),
	}
}

func (s *ShiftService) resolveCashier(ctx context.Context, storeID uint, actingAsCashierID *uint, fallbackName string) (uint, string, error) {
	// Admin tanpa switch memakai cashier_id=0 (shift admin sendiri),
	// TANPA membuat baris cashiers kloningan. Kasir memakai id-nya.
	cashierName := fallbackName
	if actingAsCashierID != nil {
		if c, err := s.cashiers.GetByID(ctx, *actingAsCashierID); err == nil {
			cashierName = c.Name
		}
		return *actingAsCashierID, cashierName, nil
	}
	return 0, cashierName, nil
}

func (s *ShiftService) GetCashierShift(ctx context.Context, storeID, outletID uint, actingAsCashierID *uint, fallbackName string) (any, error) {
	cashierID, _, err := s.resolveCashier(ctx, storeID, actingAsCashierID, fallbackName)
	if err != nil {
		return nil, err
	}

	shift, err := s.shifts.GetActiveShift(ctx, storeID, outletID, cashierID)
	if err != nil {
		return nil, ErrNoActiveShift
	}
	stats, err := s.shifts.GetShiftStats(ctx, *shift, time.Now().UTC())
	if err != nil {
		return nil, err
	}
	return currentShiftResponse(*shift, stats), nil
}

func currentShiftResponse(shift model.CashierShift, stats *repo.ShiftStats) map[string]any {
	hourMap := make(map[int]int64)
	for _, transaction := range stats.Transactions {
		hourMap[transaction.CreatedAt.Hour()] += transaction.Total
	}
	var hourly []map[string]any
	now := time.Now().UTC()
	for hour := 0; hour < 24; hour++ {
		if hour >= shift.StartedAt.Hour() && hour <= now.Hour() {
			hourly = append(hourly, map[string]any{"hour": hour, "omzet": hourMap[hour]})
		}
	}
	if hourly == nil {
		hourly = []map[string]any{}
	}

	prodQty := make(map[uint]struct {
		name string
		qty  int
	})
	for _, item := range stats.Items {
		product := prodQty[item.ProductID]
		product.name = item.Name
		product.qty += item.Qty
		prodQty[item.ProductID] = product
	}
	type topProd struct {
		ProductID uint   `json:"product_id"`
		Name      string `json:"name"`
		Qty       int    `json:"qty"`
	}
	var tops []topProd
	for productID, product := range prodQty {
		tops = append(tops, topProd{ProductID: productID, Name: product.name, Qty: product.qty})
	}
	sort.Slice(tops, func(i, j int) bool { return tops[i].Qty > tops[j].Qty })
	if len(tops) > 5 {
		tops = tops[:5]
	}
	if tops == nil {
		tops = []topProd{}
	}
	return map[string]any{"shift": summarizeShift(shift, stats), "hourly": hourly, "top_products": tops}
}

func (s *ShiftService) StartShift(ctx context.Context, storeID, outletID uint, actingAsCashierID *uint, fallbackName string, openingCash int64) (any, error) {
	cashierID, cashierName, err := s.resolveCashier(ctx, storeID, actingAsCashierID, fallbackName)
	if err != nil {
		return nil, err
	}

	shift, err := s.shifts.StartShift(ctx, storeID, outletID, cashierID, cashierName, openingCash)
	if err != nil {
		if errors.Is(err, repo.ErrCashierOutletChanged) {
			return nil, ErrCashierOutletChanged
		}
		if errors.Is(err, repo.ErrShiftAlreadyOpen) {
			return nil, ErrShiftAlreadyOpen
		}
		return nil, err
	}
	return map[string]any{"shift": summarizeShift(*shift, emptyShiftStats())}, nil
}

func (s *ShiftService) CloseShift(ctx context.Context, storeID, outletID uint, actingAsCashierID *uint, fallbackName string, closingCash int64, note *string) (any, error) {
	cashierID, cashierName, err := s.resolveCashier(ctx, storeID, actingAsCashierID, fallbackName)
	if err != nil {
		return nil, err
	}
	closed, stats, err := s.shifts.CloseShift(ctx, storeID, outletID, cashierID, closingCash, note, cashierName)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrNoActiveShift
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{"message": "Shift ditutup.", "summary": summarizeShift(*closed, stats)}, nil
}

func (s *ShiftService) ForceCloseShift(ctx context.Context, storeID, shiftID uint, closingCash *int64, note *string, closedByName string) (any, error) {
	if note == nil || strings.TrimSpace(*note) == "" {
		defaultNote := "Ditutup paksa oleh admin"
		note = &defaultNote
	}
	closed, stats, err := s.shifts.ForceCloseShift(ctx, storeID, shiftID, closingCash, note, closedByName)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrShiftNotFound
	}
	if errors.Is(err, repo.ErrShiftClosed) {
		return nil, ErrShiftAlreadyClosed
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{"message": "Shift ditutup paksa.", "summary": summarizeShift(*closed, stats)}, nil
}

func (s *ShiftService) ListShifts(ctx context.Context, storeID uint, outletID *uint, from string, to string) ([]ShiftSummary, error) {
	shifts, err := s.shifts.ListStoreShifts(ctx, storeID, outletID, from, to)
	if err != nil {
		return nil, err
	}
	results := make([]ShiftSummary, 0, len(shifts))
	for _, shift := range shifts {
		stats, err := s.shifts.GetShiftStats(ctx, *shift, shiftEndTime(*shift))
		if err != nil {
			return nil, err
		}
		results = append(results, summarizeShift(*shift, stats))
	}
	return results, nil
}

func (s *ShiftService) GetShift(ctx context.Context, storeID, shiftID uint, outletID *uint, page, limit int) (*ShiftDetail, error) {
	shift, err := s.shifts.GetShiftByID(ctx, storeID, shiftID, outletID)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrShiftNotFound
	}
	if err != nil {
		return nil, err
	}
	stats, err := s.shifts.GetShiftStats(ctx, *shift, shiftEndTime(*shift))
	if err != nil {
		return nil, err
	}
	transactions, total, err := s.shifts.ListShiftTransactions(ctx, storeID, *shift, page, limit)
	if err != nil {
		return nil, err
	}
	hourMap := make(map[int]int64)
	for _, tr := range stats.Transactions {
		hourMap[tr.CreatedAt.Hour()] += tr.Total
	}
	hourly := []ShiftHourPoint{}
	if len(hourMap) > 0 {
		minH, maxH := 24, -1
		for h := range hourMap {
			if h < minH {
				minH = h
			}
			if h > maxH {
				maxH = h
			}
		}
		for h := minH; h <= maxH; h++ {
			hourly = append(hourly, ShiftHourPoint{Hour: h, Omzet: hourMap[h]})
		}
	}
	prodQty := make(map[uint]struct {
		name string
		qty  int
	})
	for _, item := range stats.Items {
		product := prodQty[item.ProductID]
		product.name = item.Name
		product.qty += item.Qty
		prodQty[item.ProductID] = product
	}
	tops := []ShiftTopProduct{}
	for productID, product := range prodQty {
		tops = append(tops, ShiftTopProduct{ProductID: productID, Name: product.name, Qty: product.qty})
	}
	sort.Slice(tops, func(i, j int) bool { return tops[i].Qty > tops[j].Qty })
	if len(tops) > 5 {
		tops = tops[:5]
	}
	return &ShiftDetail{
		Shift:        summarizeShift(*shift, stats),
		Transactions: ShiftTransactionPage{Items: transactions, Page: page, Limit: limit, Total: total},
		Hourly:       hourly,
		TopProducts:  tops,
	}, nil
}

func (s *ShiftService) ExportShifts(ctx context.Context, storeID uint, outletID *uint, from string, to string) ([]ShiftSummary, string, error) {
	shifts, err := s.ListShifts(ctx, storeID, outletID, from, to)
	if err != nil {
		return nil, "", err
	}
	timezone, err := s.stores.GetTimezone(ctx, storeID)
	if err != nil || timezone == "" {
		timezone = "Asia/Makassar"
	}
	return shifts, timezone, nil
}

func summarizeShift(shift model.CashierShift, stats *repo.ShiftStats) ShiftSummary {
	methodSales := make(map[string]int64, len(ShiftPaymentMethods()))
	for _, method := range ShiftPaymentMethods() {
		methodSales[method] = stats.SalesByMethod[method]
	}
	expectedCash := shift.OpeningCash + methodSales[string(model.PayCash)] - stats.CashRefundTotal - stats.CashExpenses
	if shift.ExpectedCash != nil {
		expectedCash = *shift.ExpectedCash
	}
	return ShiftSummary{
		ID: shift.ID, OutletID: shift.OutletID, CashierName: shift.CashierName, ClosedByName: shift.ClosedByName,
		StartedAt: shift.StartedAt, ClosedAt: shift.ClosedAt, OpeningCash: shift.OpeningCash,
		SalesByMethod: methodSales, RefundTotal: stats.RefundTotal, Sales: stats.Sales, TrxCount: stats.TrxCount,
		ExpectedCash: expectedCash, CashExpenses: stats.CashExpenses, ClosingCashActual: shift.ClosingCashActual, CashVariance: shift.CashVariance, CloseNote: shift.CloseNote,
	}
}

func emptyShiftStats() *repo.ShiftStats {
	return &repo.ShiftStats{SalesByMethod: map[string]int64{}, Items: []model.TrxItem{}, Transactions: []model.Trx{}}
}

func shiftEndTime(shift model.CashierShift) time.Time {
	if shift.ClosedAt != nil {
		return *shift.ClosedAt
	}
	return time.Now().UTC()
}
