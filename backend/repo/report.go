package repo

import (
	"context"
	"math/big"
	"sort"
	"time"

	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/model"
)

type ReportRepo struct {
	db *gorm.DB
}

func NewReportRepo(db *gorm.DB) *ReportRepo { return &ReportRepo{db: db} }

func reportOutlet(query *gorm.DB, outletID *uint, column string) *gorm.DB {
	if outletID != nil {
		query = query.Where(column+" = ?", *outletID)
	}
	return query
}

func (r *ReportRepo) validateReportOutlet(ctx context.Context, storeID uint, outletID *uint) error {
	if outletID == nil {
		return nil
	}
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Outlet{}).
		Where("id = ? AND store_id = ?", *outletID, storeID).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return ErrOutletNotFound
	}
	return nil
}

func periodStarts(nowUTC time.Time, loc *time.Location, period string) (time.Time, time.Time, error) {
	local := nowUTC.In(loc)
	today := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, loc)
	switch period {
	case "today":
		return today, today.AddDate(0, 0, 1), nil
	case "yesterday":
		return today.AddDate(0, 0, -1), today, nil
	case "week":
		offset := (int(local.Weekday()) + 6) % 7
		start := today.AddDate(0, 0, -offset)
		return start, today.AddDate(0, 0, 1), nil
	case "month":
		start := time.Date(local.Year(), local.Month(), 1, 0, 0, 0, 0, loc)
		return start, start.AddDate(0, 1, 0), nil
	case "", "all":
		return time.Time{}, time.Time{}, nil
	default:
		return time.Time{}, time.Time{}, ErrNotFound
	}
}

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

type refundedProduct struct {
	name   string
	qty    int64
	value  int64
	profit int64
}

type refundImpact struct {
	total     int64
	count     int64
	items     int64
	hpp       int64
	byDate    map[string]int64
	byMethod  map[string]int64
	byProduct map[uint]refundedProduct
}

func roundRatioHalfUp(value, numerator, denominator int64) int64 {
	if denominator <= 0 || value <= 0 || numerator <= 0 {
		return 0
	}
	product := new(big.Int).Mul(big.NewInt(value), big.NewInt(numerator))
	product.Add(product, big.NewInt(denominator/2))
	return product.Div(product, big.NewInt(denominator)).Int64()
}

func (r *ReportRepo) refunds(ctx context.Context, storeID uint, outletID *uint, start, end time.Time, loc *time.Location, cashierID uint) (refundImpact, error) {
	if cashierID == 0 {
		return r.refundsFor(ctx, storeID, outletID, start, end, loc, nil)
	}
	return r.refundsFor(ctx, storeID, outletID, start, end, loc, &cashierID)
}

// refundsFor: cashierID nil = semua kasir; pointer (termasuk 0 = owner) =
// hanya kasir itu. Pointer diperlukan karena owner tercatat cashier_id 0.
func (r *ReportRepo) refundsFor(ctx context.Context, storeID uint, outletID *uint, start, end time.Time, loc *time.Location, cashierID *uint) (refundImpact, error) {
	impact := refundImpact{
		byDate:    map[string]int64{},
		byMethod:  map[string]int64{},
		byProduct: map[uint]refundedProduct{},
	}
	query := r.db.WithContext(ctx).Model(&model.Refund{}).
		Select("refunds.*").
		Joins("JOIN transactions ON transactions.id = refunds.trx_id").
		Where("refunds.store_id = ? AND transactions.store_id = ?", storeID, storeID)
	query = reportOutlet(query, outletID, "transactions.outlet_id")
	if !start.IsZero() {
		query = query.Where("refunds.created_at >= ? AND refunds.created_at < ?", start.UTC(), end.UTC())
	}
	if cashierID != nil {
		query = query.Where("transactions.cashier_id = ?", *cashierID)
	}
	var records []model.Refund
	if err := query.Order("refunds.created_at ASC, refunds.id ASC").Find(&records).Error; err != nil {
		return impact, err
	}
	if len(records) == 0 {
		return impact, nil
	}

	trxIDs := make([]uint, 0, len(records))
	seen := map[uint]bool{}
	for _, ref := range records {
		if !seen[ref.TrxID] {
			seen[ref.TrxID] = true
			trxIDs = append(trxIDs, ref.TrxID)
		}
	}
	var trxs []model.Trx
	if err := r.db.WithContext(ctx).Where("store_id = ? AND id IN ?", storeID, trxIDs).Find(&trxs).Error; err != nil {
		return impact, err
	}
	var items []model.TransactionItem
	if err := r.db.WithContext(ctx).Where("trx_id IN ?", trxIDs).Order("id ASC").Find(&items).Error; err != nil {
		return impact, err
	}
	trxByID := make(map[uint]model.Trx, len(trxs))
	for _, trx := range trxs {
		trx.Items = []model.TrxItem{}
		trxByID[trx.ID] = trx
	}
	for _, item := range items {
		trx, ok := trxByID[item.TrxID]
		if !ok {
			continue
		}
		trx.Items = append(trx.Items, model.TrxItem{
			ProductID: item.ProductID, VariantID: item.VariantID, VariantName: item.VariantName,
			Name: item.Name, BuyPrice: item.BuyPrice, Price: item.Price, Qty: item.Qty,
		})
		trxByID[item.TrxID] = trx
	}

	for _, ref := range records {
		trx, ok := trxByID[ref.TrxID]
		if !ok {
			continue
		}
		impact.count++
		date := ref.CreatedAt.In(loc).Format("2006-01-02")
		var refundTotal int64
		for _, item := range trx.Items {
			qty := ref.Items[item.ProductID]
			if item.VariantID != nil {
				qty = ref.VariantItems[*item.VariantID]
			}
			if qty <= 0 {
				continue
			}
			itemValue := item.Price * int64(qty)
			refundTotal += itemValue + roundRatioHalfUp(trx.Tax, itemValue, trx.Subtotal-trx.Discount-trx.PromoDiscount-trx.VoucherDiscount)
			impact.items += int64(qty)
			impact.hpp += item.BuyPrice * int64(qty)
			product := impact.byProduct[item.ProductID]
			product.name = item.Name
			product.qty += int64(qty)
			product.value += itemValue
			product.profit += (item.Price - item.BuyPrice) * int64(qty)
			impact.byProduct[item.ProductID] = product
		}
		impact.total += refundTotal
		impact.byDate[date] += refundTotal
		impact.byMethod[trx.Method] += refundTotal
	}
	return impact, nil
}

func (r *ReportRepo) Dashboard(ctx context.Context, storeID, cashierID uint, cashierView bool, outletID *uint, tz, dateStr, fromStr, toStr string) (any, error) {
	if err := r.validateReportOutlet(ctx, storeID, outletID); err != nil {
		return nil, err
	}
	loc := loadLoc(tz)
	now := time.Now()
	todayLocal := now.In(loc)
	anchor := time.Date(todayLocal.Year(), todayLocal.Month(), todayLocal.Day(), 0, 0, 0, 0, loc)
	if d, ok := parseLocalDay(dateStr, loc); ok {
		anchor = d
	}
	dayStart := anchor
	dayEnd := dayStart.AddDate(0, 0, 1)
	// Rentang eksplisit (dari PeriodPicker standar) menimpa jendela satu hari:
	// ringkasan/methods/top/recent dihitung atas [from, to]; sales7 tetap 7 hari
	// berlabuh di hari terakhir rentang.
	if f, okF := parseLocalDay(fromStr, loc); okF {
		if t, okT := parseLocalDay(toStr, loc); okT && !t.Before(f) {
			anchor = t
			dayStart = f
			dayEnd = t.AddDate(0, 0, 1)
		}
	}

	query := r.db.WithContext(ctx).Model(&model.Trx{}).Where("store_id = ? AND status IN ? AND created_at >= ? AND created_at < ?", storeID, []model.TrxStatus{model.TrxCompleted, model.TrxRefunded, model.TrxCompletedOfflineShortage}, dayStart.UTC(), dayEnd.UTC())
	query = reportOutlet(query, outletID, "outlet_id")
	if cashierView && cashierID != 0 {
		query = query.Where("cashier_id = ?", cashierID)
	}
	refundCashierID := uint(0)
	if cashierView {
		refundCashierID = cashierID
	}
	refunds, err := r.refunds(ctx, storeID, outletID, dayStart, dayEnd, loc, refundCashierID)
	if err != nil {
		return nil, err
	}

	var omzet, cnt int64
	var totalOmzet struct {
		Sum int64
		Cnt int64
	}
	query.Select("COALESCE(SUM(total), 0) as sum, COUNT(*) as cnt").Scan(&totalOmzet)
	omzet = totalOmzet.Sum - refunds.total
	cnt = totalOmzet.Cnt

	// Items sold today
	itemsQuery := r.db.WithContext(ctx).Model(&model.TransactionItem{}).
		Joins("JOIN transactions ON transactions.id = transaction_items.trx_id").
		Where("transactions.store_id = ? AND transactions.status IN ? AND transactions.created_at >= ? AND transactions.created_at < ?", storeID, []model.TrxStatus{model.TrxCompleted, model.TrxRefunded, model.TrxCompletedOfflineShortage}, dayStart.UTC(), dayEnd.UTC())
	itemsQuery = reportOutlet(itemsQuery, outletID, "transactions.outlet_id")
	if cashierView && cashierID != 0 {
		itemsQuery = itemsQuery.Where("transactions.cashier_id = ?", cashierID)
	}
	var itemsSold int64
	itemsQuery.Select("COALESCE(SUM(transaction_items.qty), 0)").Scan(&itemsSold)
	itemsSold -= refunds.items

	lowStock := int64(0)
	if !cashierView {
		productLow := r.db.WithContext(ctx).Model(&model.Product{}).
			Joins("CROSS JOIN outlets").
			Joins("LEFT JOIN product_stocks ON product_stocks.product_id = products.id AND product_stocks.outlet_id = outlets.id AND product_stocks.variant_id IS NULL").
			Where("products.store_id = ? AND outlets.store_id = ? AND products.active = ? AND products.min_stock > 0 AND COALESCE(product_stocks.qty, 0) <= products.min_stock", storeID, storeID, true)
		if outletID != nil {
			productLow = productLow.Where("outlets.id = ?", *outletID)
		}
		productLow = productLow.Where("NOT EXISTS (SELECT 1 FROM product_variants WHERE product_variants.store_id = products.store_id AND product_variants.product_id = products.id AND product_variants.deleted_at IS NULL)")
		productLow.Count(&lowStock)
		var variantLowStock int64
		variantLow := r.db.WithContext(ctx).Model(&model.ProductVariant{}).
			Joins("JOIN products ON products.id = product_variants.product_id").
			Joins("CROSS JOIN outlets").
			Joins("LEFT JOIN product_stocks ON product_stocks.variant_id = product_variants.id AND product_stocks.outlet_id = outlets.id").
			Where("product_variants.store_id = ? AND products.store_id = ? AND outlets.store_id = ? AND products.active = ? AND products.deleted_at IS NULL", storeID, storeID, storeID, true).
			Where("product_variants.min_stock > 0 AND COALESCE(product_stocks.qty, 0) <= product_variants.min_stock")
		if outletID != nil {
			variantLow = variantLow.Where("outlets.id = ?", *outletID)
		}
		variantLow.Count(&variantLowStock)
		lowStock += variantLowStock
	}

	recent, err := r.recentTrx(ctx, storeID, cashierID, cashierView, outletID, dayStart, dayEnd)
	if err != nil {
		return nil, err
	}
	todayDTO := model.DashboardToday{Omzet: omzet, TrxCount: cnt, ItemsSold: itemsSold, RefundTotal: refunds.total, LowStock: lowStock}

	if cashierView {
		return &model.DashboardCashier{Role: "cashier", Today: todayDTO, Recent: recent}, nil
	}

	weekStart := anchor.AddDate(0, 0, -6)
	weekRefunds, err := r.refunds(ctx, storeID, outletID, weekStart, dayEnd, loc, 0)
	if err != nil {
		return nil, err
	}
	var salesRows []struct {
		D time.Time
		V int64
	}
	if r.db.Dialector.Name() == "sqlite" {
		// sqlite has no AT TIME ZONE / ::date: aggregate in Go for local testing.
		var trxs []model.Trx
		salesQuery := r.db.WithContext(ctx).
			Where("store_id = ? AND status IN ? AND created_at >= ? AND created_at < ?", storeID, []model.TrxStatus{model.TrxCompleted, model.TrxRefunded, model.TrxCompletedOfflineShortage}, weekStart.UTC(), dayEnd.UTC())
		salesQuery = reportOutlet(salesQuery, outletID, "outlet_id")
		if err := salesQuery.Find(&trxs).Error; err == nil {
			daySum := map[string]int64{}
			for _, t := range trxs {
				k := t.CreatedAt.In(loc).Format("2006-01-02")
				if k >= weekStart.Format("2006-01-02") && k <= anchor.Format("2006-01-02") {
					daySum[k] += t.Total
				}
			}
			for k, v := range daySum {
				parsed, _ := time.ParseInLocation("2006-01-02", k, loc)
				salesRows = append(salesRows, struct {
					D time.Time
					V int64
				}{D: parsed, V: v})
			}
		}
		if salesRows == nil {
			salesRows = []struct {
				D time.Time
				V int64
			}{}
		}
	} else {
		query := `
			SELECT (created_at AT TIME ZONE ?)::date as d, COALESCE(SUM(total), 0) as v
			FROM transactions
			WHERE store_id = ? AND status IN ('completed', 'refunded', 'completed_offline_shortage')
			  AND (created_at AT TIME ZONE ?)::date >= ?::date
			  AND (created_at AT TIME ZONE ?)::date <= ?::date
		`
		args := []any{tz, storeID, tz, weekStart.Format("2006-01-02"), tz, anchor.Format("2006-01-02")}
		if outletID != nil {
			query += " AND outlet_id = ?"
			args = append(args, *outletID)
		}
		query += " GROUP BY (created_at AT TIME ZONE ?)::date ORDER BY d ASC"
		args = append(args, tz)
		r.db.WithContext(ctx).Raw(query, args...).Scan(&salesRows)
	}

	byDate := map[string]int64{}
	for _, sr := range salesRows {
		byDate[sr.D.Format("2006-01-02")] = sr.V
	}
	for date, amount := range weekRefunds.byDate {
		byDate[date] -= amount
	}

	sales7 := make([]model.DayPoint, 0, 7)
	for i := 0; i < 7; i++ {
		d := weekStart.AddDate(0, 0, i)
		key := d.Format("2006-01-02")
		sales7 = append(sales7, model.DayPoint{Date: key, Omzet: byDate[key]})
	}

	methods := []model.MethodPoint{}
	methodQuery := r.db.WithContext(ctx).Model(&model.Trx{}).
		Select("method, COALESCE(SUM(total), 0) as total").
		Where("store_id = ? AND status IN ? AND created_at >= ? AND created_at < ?", storeID, []model.TrxStatus{model.TrxCompleted, model.TrxRefunded, model.TrxCompletedOfflineShortage}, dayStart.UTC(), dayEnd.UTC())
	methodQuery = reportOutlet(methodQuery, outletID, "outlet_id")
	methodQuery.Group("method").Order("total DESC").Scan(&methods)
	for method, amount := range refunds.byMethod {
		found := false
		for i := range methods {
			if methods[i].Method == method {
				methods[i].Total -= amount
				found = true
				break
			}
		}
		if !found {
			methods = append(methods, model.MethodPoint{Method: method, Total: -amount})
		}
	}
	sort.Slice(methods, func(i, j int) bool { return methods[i].Total > methods[j].Total })
	if methods == nil {
		methods = []model.MethodPoint{}
	}

	var grossTop []model.TopProduct
	topQuery := r.db.WithContext(ctx).Model(&model.TransactionItem{}).
		Select("transaction_items.product_id, transaction_items.name, SUM(transaction_items.qty) as qty, SUM(transaction_items.qty * transaction_items.price) as revenue").
		Joins("JOIN transactions ON transactions.id = transaction_items.trx_id").
		Where("transactions.store_id = ? AND transactions.status IN ? AND transactions.created_at >= ? AND transactions.created_at < ?", storeID, []model.TrxStatus{model.TrxCompleted, model.TrxRefunded, model.TrxCompletedOfflineShortage}, dayStart.UTC(), dayEnd.UTC())
	topQuery = reportOutlet(topQuery, outletID, "transactions.outlet_id")
	topQuery.Group("transaction_items.product_id, transaction_items.name").Scan(&grossTop)
	topByProduct := make(map[uint]model.TopProduct, len(grossTop)+len(refunds.byProduct))
	for _, product := range grossTop {
		topByProduct[product.ProductID] = product
	}
	for productID, refunded := range refunds.byProduct {
		product := topByProduct[productID]
		product.ProductID = productID
		if product.Name == "" {
			product.Name = refunded.name
		}
		product.Qty -= int(refunded.qty)
		product.Revenue -= refunded.value
		topByProduct[productID] = product
	}
	top := make([]model.TopProduct, 0, len(topByProduct))
	for _, product := range topByProduct {
		top = append(top, product)
	}
	sort.Slice(top, func(i, j int) bool { return top[i].Qty > top[j].Qty })
	if len(top) > 5 {
		top = top[:5]
	}

	return &model.DashboardAdmin{
		Role:        "admin",
		Today:       todayDTO,
		Sales7:      sales7,
		Methods:     methods,
		TopProducts: top,
		Recent:      recent,
	}, nil
}

func (r *ReportRepo) recentTrx(ctx context.Context, storeID, cashierID uint, cashierView bool, outletID *uint, dayStart, dayEnd time.Time) ([]model.TrxBrief, error) {
	query := r.db.WithContext(ctx).Model(&model.Trx{}).
		Select("id, cashier_name, total, status, created_at as time").
		Where("store_id = ? AND status = 'completed' AND created_at >= ? AND created_at < ?", storeID, dayStart.UTC(), dayEnd.UTC())
	query = reportOutlet(query, outletID, "outlet_id")
	if cashierView && cashierID != 0 {
		query = query.Where("cashier_id = ?", cashierID)
	}
	recent := []model.TrxBrief{}
	err := query.Order("created_at DESC").Limit(5).Scan(&recent).Error
	if recent == nil {
		recent = []model.TrxBrief{}
	}
	return recent, err
}

func (r *ReportRepo) Report(ctx context.Context, storeID uint, outletID *uint, period, tz, dateStr, fromStr, toStr string) (*model.ReportBundle, error) {
	if err := r.validateReportOutlet(ctx, storeID, outletID); err != nil {
		return nil, err
	}
	loc := loadLoc(tz)
	start, end, err := periodStarts(time.Now().UTC(), loc, period)
	if err != nil {
		return nil, err
	}
	// date opsional hanya untuk today/yesterday: today+D = rentang D,
	// yesterday+D = rentang D-1. Period lain mengabaikan date.
	// date tak valid/kosong = abaikan (perilaku lama).
	if anchor, ok := parseLocalDay(dateStr, loc); ok {
		switch period {
		case "today":
			start, end = anchor, anchor.AddDate(0, 0, 1)
		case "yesterday":
			start, end = anchor.AddDate(0, 0, -1), anchor
		}
	}
	// Rentang khusus dari pemilih periode UI: from/to inklusif per hari lokal,
	// mengalahkan period/date bila keduanya valid dan berurutan.
	if f, ok := parseLocalDay(fromStr, loc); ok {
		if t, ok2 := parseLocalDay(toStr, loc); ok2 && !t.Before(f) {
			start, end = f, t.AddDate(0, 0, 1)
		}
	}

	query := r.db.WithContext(ctx).Model(&model.Trx{}).Where("store_id = ? AND status IN ?", storeID, []model.TrxStatus{model.TrxCompleted, model.TrxRefunded, model.TrxCompletedOfflineShortage})
	query = reportOutlet(query, outletID, "outlet_id")
	if !start.IsZero() {
		query = query.Where("created_at >= ? AND created_at < ?", start.UTC(), end.UTC())
	}
	refunds, err := r.refunds(ctx, storeID, outletID, start, end, loc, 0)
	if err != nil {
		return nil, err
	}

	var list []*model.Trx
	if err := query.Order("created_at DESC").Find(&list).Error; err != nil {
		return nil, err
	}

	if len(list) > 0 {
		ids := make([]uint, len(list))
		for i, t := range list {
			ids[i] = t.ID
		}
		var items []model.TransactionItem
		if err := r.db.WithContext(ctx).Where("trx_id IN ?", ids).Order("id ASC").Find(&items).Error; err == nil {
			byID := map[uint]*model.Trx{}
			for _, t := range list {
				t.Items = []model.TrxItem{}
				byID[t.ID] = t
			}
			for _, it := range items {
				if t, ok := byID[it.TrxID]; ok {
					t.Items = append(t.Items, model.TrxItem{
						ProductID: it.ProductID,
						Name:      it.Name,
						BuyPrice:  it.BuyPrice,
						Price:     it.Price,
						Qty:       it.Qty,
					})
				}
			}
		}
	}

	// Transaksi lama mencatat kemenangan voucher di kolom promo; klasifikasikan
	// ulang total diskon menurut ada/tidaknya kode pada promo sumbernya.
	promoHasCode := map[uint]bool{}
	var storePromos []model.Promo
	if err := r.db.WithContext(ctx).Where("store_id = ?", storeID).Find(&storePromos).Error; err == nil {
		for _, pr := range storePromos {
			promoHasCode[pr.ID] = pr.VoucherCode != nil && *pr.VoucherCode != ""
		}
	}

	bundle := &model.ReportBundle{
		Period:       period,
		ByMethod:     []model.MethodPoint{},
		ByStatus:     []model.StatusCount{},
		Products:     []model.ProductReportRow{},
		Transactions: []model.TrxProfitRow{},
		Cashiers:     []model.CashierReportRow{},
		Stock:        []model.StockRow{},
	}

	prodAgg := map[uint]*model.ProductReportRow{}
	cashierAggs := map[uint]*cashierAgg{}
	for _, t := range list {
		ca := cashierAggFor(cashierAggs, t.CashierID, t.CashierName)
		var hpp int64
		for _, it := range t.Items {
			hpp += it.BuyPrice * int64(it.Qty)
			bundle.Summary.ItemsSold += int64(it.Qty)
			ca.row.Items += int64(it.Qty)
			ca.prodQty[it.ProductID] += it.Qty
			ca.prodName[it.ProductID] = it.Name

			pr, ok := prodAgg[it.ProductID]
			if !ok {
				pr = &model.ProductReportRow{ProductID: it.ProductID, Name: it.Name}
				prodAgg[it.ProductID] = pr
			}
			pr.Qty += it.Qty
			pr.Revenue += it.Price * int64(it.Qty)
			pr.Profit += (it.Price - it.BuyPrice) * int64(it.Qty)
		}
		ca.row.Trx++
		ca.row.Gross += t.Total
		bundle.Summary.Omzet += t.Total
		bundle.Summary.TrxCount++
		bundle.Summary.GrossProfit += t.Total - hpp
		if t.VoucherDiscount > 0 {
			bundle.Summary.VoucherDiscountTotal += t.VoucherDiscount
			bundle.Summary.PromoDiscountTotal += t.PromoDiscount
		} else if t.PromoID != nil && promoHasCode[*t.PromoID] {
			bundle.Summary.VoucherDiscountTotal += t.PromoDiscount
		} else {
			bundle.Summary.PromoDiscountTotal += t.PromoDiscount
		}
		bundle.ByMethod = addMethod(bundle.ByMethod, t.Method, t.Total)

		dateStr := ""
		if !t.CreatedAt.IsZero() {
			dateStr = t.CreatedAt.In(loc).Format("2006-01-02")
		}
		bundle.Transactions = append(bundle.Transactions, model.TrxProfitRow{
			Date:    dateStr,
			ID:      t.ID,
			Cashier: t.CashierName,
			Method:  t.Method,
			Total:   t.Total,
			HPP:     hpp,
			Profit:  t.Total - hpp,
			Status:  t.Status,
		})
	}
	bundle.Summary.Omzet -= refunds.total
	bundle.Summary.ItemsSold -= refunds.items
	bundle.Summary.GrossProfit += refunds.hpp - refunds.total
	bundle.Summary.RefundTotal = refunds.total
	bundle.Summary.RefundCount = refunds.count

	// Pengeluaran operasional pada rentang yang sama, per tanggal kalender
	// lokal (kolom date memang tanggal kalender, jadi bandingkan sebagai
	// string tanggal — awal inklusif, akhir eksklusif seperti transaksi).
	expQuery := r.db.WithContext(ctx).Model(&model.Expense{}).Where("store_id = ?", storeID)
	expQuery = reportOutlet(expQuery, outletID, "outlet_id")
	if !start.IsZero() {
		expQuery = expQuery.Where("date >= ? AND date < ?", start.In(loc).Format("2006-01-02"), end.In(loc).Format("2006-01-02"))
	}
	var expRows []model.ExpenseCategoryRow
	if err := expQuery.Select("category, COALESCE(SUM(amount), 0) AS total, COUNT(*) AS count").
		Group("category").Order("total DESC").Scan(&expRows).Error; err == nil {
		bundle.Expenses = expRows
		for _, row := range expRows {
			bundle.Summary.ExpenseTotal += row.Total
		}
	}
	for method, amount := range refunds.byMethod {
		bundle.ByMethod = addMethod(bundle.ByMethod, method, -amount)
	}
	for productID, refunded := range refunds.byProduct {
		product, ok := prodAgg[productID]
		if !ok {
			product = &model.ProductReportRow{ProductID: productID, Name: refunded.name}
			prodAgg[productID] = product
		}
		product.Qty -= int(refunded.qty)
		product.Revenue -= refunded.value
		product.Profit -= refunded.profit
	}

	// Finalisasi agregat per kasir: refund diatribusikan per kasir (termasuk
	// kasir yang hanya punya refund di periode ini, tanpa penjualan).
	{
		var refundOnly []struct {
			CashierID   uint
			CashierName string
		}
		rq := r.db.WithContext(ctx).Model(&model.Refund{}).
			Select("DISTINCT transactions.cashier_id AS cashier_id, transactions.cashier_name AS cashier_name").
			Joins("JOIN transactions ON transactions.id = refunds.trx_id").
			Where("refunds.store_id = ? AND transactions.store_id = ?", storeID, storeID)
		rq = reportOutlet(rq, outletID, "transactions.outlet_id")
		if !start.IsZero() {
			rq = rq.Where("refunds.created_at >= ? AND refunds.created_at < ?", start.UTC(), end.UTC())
		}
		_ = rq.Scan(&refundOnly).Error
		for _, ro := range refundOnly {
			cashierAggFor(cashierAggs, ro.CashierID, ro.CashierName)
		}
		for id, ca := range cashierAggs {
			cid := id
			if ri, rerr := r.refundsFor(ctx, storeID, outletID, start, end, loc, &cid); rerr == nil {
				ca.row.Refund = ri.total
				ca.row.Items -= ri.items
				for pid, rp := range ri.byProduct {
					ca.prodQty[pid] -= int(rp.qty)
				}
			}
			ca.row.Net = ca.row.Gross - ca.row.Refund
			bestQty := 0
			for pid, q := range ca.prodQty {
				if q > bestQty {
					bestQty = q
					ca.row.TopProduct = ca.prodName[pid]
					ca.row.TopProductQty = q
				}
			}
			bundle.Cashiers = append(bundle.Cashiers, *ca.row)
		}
		sort.Slice(bundle.Cashiers, func(i, j int) bool { return bundle.Cashiers[i].Net > bundle.Cashiers[j].Net })
	}

	// Fill SKU via product lookup (TransactionItem has no SKU column).
	if len(prodAgg) > 0 {
		var prodsForSKU []model.Product
		_ = r.db.WithContext(ctx).Where("store_id = ?", storeID).Find(&prodsForSKU).Error
		skuByID := make(map[uint]string, len(prodsForSKU))
		for _, p := range prodsForSKU {
			skuByID[p.ID] = p.SKU
		}
		for _, pr := range prodAgg {
			if sku, ok := skuByID[pr.ProductID]; ok {
				pr.SKU = sku
			}
			bundle.Products = append(bundle.Products, *pr)
		}
	}

	sort.Slice(bundle.Products, func(i, j int) bool { return bundle.Products[i].Qty > bundle.Products[j].Qty })
	sort.Slice(bundle.ByMethod, func(i, j int) bool { return bundle.ByMethod[i].Total > bundle.ByMethod[j].Total })

	statusQuery := r.db.WithContext(ctx).Model(&model.Trx{}).
		Select("status, count(*) as count").
		Where("store_id = ?", storeID)
	statusQuery = reportOutlet(statusQuery, outletID, "outlet_id")
	if !start.IsZero() {
		statusQuery = statusQuery.Where("created_at >= ? AND created_at < ?", start.UTC(), end.UTC())
	}
	statusQuery.Group("status").
		Scan(&bundle.ByStatus)
	if bundle.ByStatus == nil {
		bundle.ByStatus = []model.StatusCount{}
	}
	if bundle.Transactions == nil {
		bundle.Transactions = []model.TrxProfitRow{}
	}
	if bundle.Stock == nil {
		bundle.Stock = []model.StockRow{}
	}

	var prods []model.Product
	r.db.WithContext(ctx).Where("store_id = ?", storeID).Order("name ASC").Find(&prods)
	stockQuery := r.db.WithContext(ctx).Table("product_stocks").
		Select("product_stocks.product_id, COALESCE(SUM(product_stocks.qty), 0) AS qty").
		Joins("JOIN outlets ON outlets.id = product_stocks.outlet_id").
		Joins("JOIN products ON products.id = product_stocks.product_id").
		Where("outlets.store_id = ? AND products.store_id = ? AND product_stocks.variant_id IS NULL", storeID, storeID)
	if outletID != nil {
		stockQuery = stockQuery.Where("product_stocks.outlet_id = ?", *outletID)
	}
	var productStocks []struct {
		ProductID uint
		Qty       int
	}
	if err := stockQuery.Group("product_stocks.product_id").Scan(&productStocks).Error; err != nil {
		return nil, err
	}
	stockByProduct := make(map[uint]int, len(productStocks))
	for _, stock := range productStocks {
		stockByProduct[stock.ProductID] = stock.Qty
	}
	for _, p := range prods {
		stock := stockByProduct[p.ID]
		sr := model.StockRow{
			Name:       p.Name,
			SKU:        p.SKU,
			Stock:      stock,
			BuyPrice:   p.BuyPrice,
			SellPrice:  p.SellPrice,
			StockValue: p.BuyPrice * int64(stock),
		}
		bundle.Stock = append(bundle.Stock, sr)
	}

	return bundle, nil
}

func (r *ReportRepo) TaxReport(ctx context.Context, storeID uint, outletID *uint, start, end time.Time, timezone string) (*model.TaxReport, error) {
	if err := r.validateReportOutlet(ctx, storeID, outletID); err != nil {
		return nil, err
	}
	var totals struct {
		Gross    int64 `gorm:"column:gross"`
		DPP      int64 `gorm:"column:dpp"`
		PPN      int64 `gorm:"column:ppn"`
		TrxCount int64 `gorm:"column:trx_count"`
	}
	trxQuery := r.db.WithContext(ctx).Model(&model.Trx{}).
		Select("COALESCE(SUM(total), 0) AS gross, COALESCE(SUM(total - tax), 0) AS dpp, COALESCE(SUM(tax), 0) AS ppn, COUNT(*) AS trx_count").
		Where("store_id = ? AND status IN ? AND created_at >= ? AND created_at < ?", storeID, []model.TrxStatus{model.TrxCompleted, model.TrxRefunded, model.TrxCompletedOfflineShortage}, start.UTC(), end.UTC())
	trxQuery = reportOutlet(trxQuery, outletID, "outlet_id")
	if err := trxQuery.Scan(&totals).Error; err != nil {
		return nil, err
	}

	refunds, err := r.refunds(ctx, storeID, outletID, start, end, loadLoc(timezone), 0)
	if err != nil {
		return nil, err
	}
	return &model.TaxReport{
		Gross:        totals.Gross,
		RefundTotal:  refunds.total,
		DPP:          totals.DPP,
		PPNCollected: totals.PPN,
		TrxCount:     totals.TrxCount,
		RefundCount:  refunds.count,
	}, nil
}

func addMethod(list []model.MethodPoint, method string, total int64) []model.MethodPoint {
	for i := range list {
		if list[i].Method == method {
			list[i].Total += total
			return list
		}
	}
	return append(list, model.MethodPoint{Method: method, Total: total})
}

type cashierAgg struct {
	row      *model.CashierReportRow
	prodQty  map[uint]int
	prodName map[uint]string
}

func cashierAggFor(m map[uint]*cashierAgg, id uint, name string) *cashierAgg {
	if ca, ok := m[id]; ok {
		return ca
	}
	ca := &cashierAgg{
		row:      &model.CashierReportRow{CashierID: id, Name: name},
		prodQty:  map[uint]int{},
		prodName: map[uint]string{},
	}
	m[id] = ca
	return ca
}
