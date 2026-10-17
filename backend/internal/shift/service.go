package shift

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/rayhanvito/nojpos/backend/internal/staff"
)

// Shifts adalah kontrak penyimpanan shift yang dipakai Service
// (ShiftRepository lama di service/interfaces.go, bertipe slice).
type Shifts interface {
	GetActiveShift(ctx context.Context, storeID, outletID, cashierID uint) (*CashierShift, error)
	StartShift(ctx context.Context, storeID, outletID, cashierID uint, cashierName string, openingCash int64) (*CashierShift, error)
	CloseShift(ctx context.Context, storeID, outletID, cashierID uint, closingCash int64, note *string, closedByName string) (*CashierShift, *ShiftStats, error)
	ForceCloseShift(ctx context.Context, storeID, shiftID uint, closingCash *int64, note *string, closedByName string) (*CashierShift, *ShiftStats, error)
	ListStoreShifts(ctx context.Context, storeID uint, outletID *uint, from string, to string) ([]*CashierShift, error)
	GetShiftByID(ctx context.Context, storeID, shiftID uint, outletID *uint) (*CashierShift, error)
	GetShiftStats(ctx context.Context, shift CashierShift, endTime time.Time) (*ShiftStats, error)
	ListShiftTransactions(ctx context.Context, storeID uint, shift CashierShift, page, limit int) ([]*trx, int64, error)
}

// Cashiers adalah kontrak kasir yang dipakai Service — hanya
// GetByIDUnscoped dari CashierRepository lama yang benar-benar dipakai
// (tipe staff.Cashier).
type Cashiers interface {
	GetByIDUnscoped(ctx context.Context, id uint) (*staff.Cashier, error)
}

// Stores adalah kontrak toko yang dipakai Service — hanya GetTimezone
// dari StoreRepository lama yang benar-benar dipakai (diimplementasikan
// outlet.StoreRepo).
type Stores interface {
	GetTimezone(ctx context.Context, storeID uint) (string, error)
}

// Service adalah ShiftService lama di service/shift.go, dipindah
// verbatim.
type Service struct {
	shifts   Shifts
	cashiers Cashiers
	stores   Stores
}

func NewService(shifts Shifts, cashiers Cashiers, stores Stores) *Service {
	return &Service{shifts: shifts, cashiers: cashiers, stores: stores}
}

func ShiftPaymentMethods() []string {
	// Nilai literal sama persis dengan konstanta Pay* kanonis di
	// slice sales (BE-F8); slice ini di bawah sales di DAG sehingga
	// tidak mengimpornya.
	return []string{
		"Cash", "Bank Transfer", "QRIS",
		"E-Wallet", "Card", "Credit",
	}
}

func (s *Service) resolveCashier(ctx context.Context, storeID uint, actingAsCashierID *uint, fallbackName string) (uint, string, error) {
	// Admin tanpa switch memakai cashier_id=0 (shift admin sendiri),
	// TANPA membuat baris cashiers kloningan. Kasir memakai id-nya.
	cashierName := fallbackName
	if actingAsCashierID != nil {
		if c, err := s.cashiers.GetByIDUnscoped(ctx, *actingAsCashierID); err == nil {
			cashierName = c.Name
		}
		return *actingAsCashierID, cashierName, nil
	}
	return 0, cashierName, nil
}

func (s *Service) GetCashierShift(ctx context.Context, storeID, outletID uint, actingAsCashierID *uint, fallbackName string) (any, error) {
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

func currentShiftResponse(shift CashierShift, stats *ShiftStats) map[string]any {
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

func (s *Service) StartShift(ctx context.Context, storeID, outletID uint, actingAsCashierID *uint, fallbackName string, openingCash int64) (any, error) {
	cashierID, cashierName, err := s.resolveCashier(ctx, storeID, actingAsCashierID, fallbackName)
	if err != nil {
		return nil, err
	}

	shift, err := s.shifts.StartShift(ctx, storeID, outletID, cashierID, cashierName, openingCash)
	if err != nil {
		// BE-F5: sentinel repo dan service kini satu instans kanonis
		// di paket ini (errors.go), jadi pemetaan lama menjadi
		// pass-through — errors.Is di handler identik seperti dulu.
		if errors.Is(err, ErrCashierOutletChanged) {
			return nil, ErrCashierOutletChanged
		}
		if errors.Is(err, ErrShiftAlreadyOpen) {
			return nil, ErrShiftAlreadyOpen
		}
		return nil, err
	}
	return map[string]any{"shift": summarizeShift(*shift, emptyShiftStats())}, nil
}

func (s *Service) CloseShift(ctx context.Context, storeID, outletID uint, actingAsCashierID *uint, fallbackName string, closingCash int64, note *string) (any, error) {
	cashierID, cashierName, err := s.resolveCashier(ctx, storeID, actingAsCashierID, fallbackName)
	if err != nil {
		return nil, err
	}
	closed, stats, err := s.shifts.CloseShift(ctx, storeID, outletID, cashierID, closingCash, note, cashierName)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrNoActiveShift
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{"message": "Shift ditutup.", "summary": summarizeShift(*closed, stats)}, nil
}

func (s *Service) ForceCloseShift(ctx context.Context, storeID, shiftID uint, closingCash *int64, note *string, closedByName string) (any, error) {
	if note == nil || strings.TrimSpace(*note) == "" {
		defaultNote := "Ditutup paksa oleh admin"
		note = &defaultNote
	}
	closed, stats, err := s.shifts.ForceCloseShift(ctx, storeID, shiftID, closingCash, note, closedByName)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrShiftNotFound
	}
	if errors.Is(err, ErrShiftClosed) {
		return nil, ErrShiftAlreadyClosed
	}
	if err != nil {
		return nil, err
	}
	return map[string]any{"message": "Shift ditutup paksa.", "summary": summarizeShift(*closed, stats)}, nil
}

func (s *Service) ListShifts(ctx context.Context, storeID uint, outletID *uint, from string, to string) ([]ShiftSummary, error) {
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

func (s *Service) GetShift(ctx context.Context, storeID, shiftID uint, outletID *uint, page, limit int) (*ShiftDetail, error) {
	shift, err := s.shifts.GetShiftByID(ctx, storeID, shiftID, outletID)
	if errors.Is(err, ErrNotFound) {
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

func (s *Service) ExportShifts(ctx context.Context, storeID uint, outletID *uint, from string, to string) ([]ShiftSummary, string, error) {
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

func summarizeShift(shift CashierShift, stats *ShiftStats) ShiftSummary {
	methodSales := make(map[string]int64, len(ShiftPaymentMethods()))
	for _, method := range ShiftPaymentMethods() {
		methodSales[method] = stats.SalesByMethod[method]
	}
	expectedCash := shift.OpeningCash + methodSales["Cash"] - stats.CashRefundTotal - stats.CashExpenses
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

func emptyShiftStats() *ShiftStats {
	return &ShiftStats{SalesByMethod: map[string]int64{}, Items: []trxItem{}, Transactions: []trx{}}
}
