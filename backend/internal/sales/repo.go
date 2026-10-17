package sales

import (
	"context"
	"errors"
	"fmt"
	cust "github.com/rayhanvito/nojpos/backend/internal/customer"
	"github.com/rayhanvito/nojpos/backend/internal/inventory"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/finance"
	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
	"github.com/rayhanvito/nojpos/backend/internal/promo"
)

type CheckoutItem struct {
	ProductID uint
	VariantID *uint
	Qty       int
}

type RefundItem struct {
	ProductID uint
	VariantID *uint
	Qty       int
}

type CheckoutInput struct {
	StoreID      uint
	OutletID     uint
	CashierID    uint
	CashierName  string
	ClientTxID   string
	Items        []CheckoutItem
	Discount     int64
	VoucherCode  string
	Method       string
	Paid         int64
	Customer     string
	CustomerID   *uint
	PointsRedeem int64
}

type PromoPreviewInput struct {
	StoreID     uint
	OutletID    uint
	Items       []CheckoutItem
	VoucherCode string
}

type promoLine struct {
	Product catalog.Product
	Value   int64
}

// Repo adalah repositori transaksi (TrxRepo lama di
// repo/transaction.go): menyematkan BaseRepo dan menyimpan
// Transactor, meniru pola purchasing.NewPurchaseOrderRepo —
// metode checkout/refund memakai tx.WithinTx JOIN, jadi mereka ikut
// transaksi luar yang dibuka Service (atau membuka sendiri bila
// belum ada).
type Repo struct {
	database.BaseRepo
	tx *database.Transactor
}

type checkoutLine struct {
	item    TrxItem
	product catalog.Product
	variant *catalog.ProductVariant
	qty     int
	stock   int
}

type checkoutItemKey struct {
	productID  uint
	variantID  uint
	hasVariant bool
}

func uintPointer(value uint) *uint { return &value }

func NewRepo(db *gorm.DB, tx *database.Transactor) *Repo {
	return &Repo{BaseRepo: database.NewBaseRepo(db), tx: tx}
}

// isPostgres reports whether the dialector supports pg locks (FOR UPDATE,
// pg_advisory_xact_lock). Sqlite path skips them — same results, no crash.
func (r *Repo) isPostgres(ctx context.Context) bool {
	return r.DB(ctx).Dialector.Name() != "sqlite"
}

func mergeCheckoutItems(items []CheckoutItem) []CheckoutItem {
	merged := make([]CheckoutItem, 0, len(items))
	idx := make(map[checkoutItemKey]int, len(items))
	for _, item := range items {
		key := checkoutItemKey{productID: item.ProductID}
		if item.VariantID != nil {
			key.variantID = *item.VariantID
			key.hasVariant = true
		}
		if i, ok := idx[key]; ok {
			merged[i].Qty += item.Qty
			continue
		}
		idx[key] = len(merged)
		merged = append(merged, item)
	}
	return merged
}

func (r *Repo) resolveCheckoutItem(ctx context.Context, tx *gorm.DB, storeID, outletID uint, item CheckoutItem, checkStock bool) (checkoutLine, error) {
	if item.Qty < 1 {
		return checkoutLine{}, fmt.Errorf("qty harus minimal 1")
	}
	var product catalog.Product
	productQuery := tx.WithContext(ctx)
	if r.isPostgres(ctx) {
		productQuery = productQuery.Clauses(clause.Locking{Strength: "UPDATE"})
	}
	if err := productQuery.Where("id = ? AND store_id = ?", item.ProductID, storeID).First(&product).Error; err != nil {
		if isNotFound(err) {
			return checkoutLine{}, ErrNotFound
		}
		return checkoutLine{}, err
	}
	if !product.Active {
		return checkoutLine{}, ErrProductInactive
	}

	var variantCount int64
	if err := tx.WithContext(ctx).Model(&catalog.ProductVariant{}).Where("store_id = ? AND product_id = ?", storeID, product.ID).Count(&variantCount).Error; err != nil {
		return checkoutLine{}, err
	}
	var variant *catalog.ProductVariant
	if item.VariantID != nil {
		if variantCount == 0 {
			return checkoutLine{}, ErrVariantInvalid
		}
		var record catalog.ProductVariant
		variantQuery := tx.WithContext(ctx)
		if r.isPostgres(ctx) {
			variantQuery = variantQuery.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := variantQuery.Where("id = ? AND store_id = ? AND product_id = ?", *item.VariantID, storeID, product.ID).First(&record).Error; err != nil {
			if isNotFound(err) {
				return checkoutLine{}, ErrVariantInvalid
			}
			return checkoutLine{}, err
		}
		variant = &record
	} else if variantCount > 0 {
		return checkoutLine{}, ErrVariantRequired
	}

	buyPrice, sellPrice := product.BuyPrice, product.SellPrice
	variantID := (*uint)(nil)
	if variant != nil {
		buyPrice, sellPrice = variant.BuyPrice, variant.SellPrice
		id := variant.ID
		variantID = &id
	}
	stock, err := inventory.GetOutletStock(tx, outletID, product.ID, variantID)
	if err != nil {
		return checkoutLine{}, err
	}
	if checkStock && stock < item.Qty {
		return checkoutLine{}, ErrStockInsufficient
	}
	priceQuery := tx.WithContext(ctx).Model(&catalog.PriceTier{}).
		Where("store_id = ? AND product_id = ? AND min_qty <= ?", storeID, product.ID, item.Qty)
	if variant == nil {
		priceQuery = priceQuery.Where("variant_id IS NULL")
	} else {
		priceQuery = priceQuery.Where("variant_id = ?", variant.ID)
	}
	var tier catalog.PriceTier
	err = priceQuery.Order("min_qty DESC, id DESC").First(&tier).Error
	if err != nil && !isNotFound(err) {
		return checkoutLine{}, err
	}
	if err == nil {
		sellPrice = tier.SellPrice
	}

	line := checkoutLine{
		item:    TrxItem{ProductID: product.ID, Name: product.Name, BuyPrice: buyPrice, Price: sellPrice, Qty: item.Qty},
		product: product,
		qty:     item.Qty,
		stock:   stock,
	}
	if variant != nil {
		variantID := variant.ID
		line.item.VariantID = &variantID
		line.item.VariantName = variant.Name
		line.variant = variant
	}
	return line, nil
}

func (r *Repo) Checkout(ctx context.Context, in CheckoutInput) (*Trx, bool, error) {
	if in.ClientTxID != "" {
		if existing, err := r.findByClientTxID(ctx, in.StoreID, in.ClientTxID); err == nil {
			return existing, false, nil
		} else if !errors.Is(err, ErrNotFound) {
			return nil, false, err
		}
	}
	if len(in.Items) == 0 {
		return nil, false, ErrEmptyItems
	}
	if !ValidPayMethod(in.Method) {
		return nil, false, fmt.Errorf("metode pembayaran tidak valid")
	}
	if in.OutletID == 0 {
		var outlet outlet.Outlet
		if err := r.DB(ctx).Where("store_id = ? AND is_default = ?", in.StoreID, true).First(&outlet).Error; err != nil {
			return nil, false, mapDBErr(err)
		}
		in.OutletID = outlet.ID
	}
	if in.Method == string(PayCredit) {
		if in.CustomerID == nil {
			return nil, false, ErrCreditCustomerRequired
		}
		if in.Paid != 0 {
			return nil, false, ErrCreditPaid
		}
	}

	in.Items = mergeCheckoutItems(in.Items)

	var resultTrx *Trx
	var existingTrxID uint

	err := r.tx.WithinTx(ctx, func(ctx context.Context) error {
		tx := r.DB(ctx)
		// Serialize checkouts per store on postgres. Skipped on sqlite.
		// IDs are numeric now, so pass the store as text for hashtext.
		if r.isPostgres(ctx) {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", strconv.FormatUint(uint64(in.StoreID), 10)).Error; err != nil {
				return err
			}
		}
		if in.ClientTxID != "" {
			var existing Trx
			err := tx.Where("store_id = ? AND client_tx_id = ?", in.StoreID, in.ClientTxID).First(&existing).Error
			if err == nil {
				existingTrxID = existing.ID
				return nil
			}
			if !isNotFound(err) {
				return err
			}
		}

		var store outlet.Store
		if err := tx.Where("id = ?", in.StoreID).First(&store).Error; err != nil {
			return mapDBErr(err)
		}
		var customer *cust.Customer
		if in.CustomerID != nil {
			var record cust.Customer
			q := tx
			if r.isPostgres(ctx) {
				q = q.Clauses(clause.Locking{Strength: "UPDATE"})
			}
			if err := q.Where("id = ? AND store_id = ?", *in.CustomerID, in.StoreID).First(&record).Error; err != nil {
				if isNotFound(err) {
					return ErrCustomerNotFound
				}
				return err
			}
			customer = &record
		}
		if store.LoyaltyEnabled {
			if in.PointsRedeem < 0 {
				return ErrPointsRedeemInvalid
			}
			if in.PointsRedeem > 0 && customer == nil {
				return ErrPointsCustomerRequired
			}
			if in.PointsRedeem > 0 && in.PointsRedeem > customer.Points {
				return ErrPointsBalance
			}
			if store.LoyaltyPointValue < 0 || (customer != nil && store.LoyaltyEarnPer <= 0) {
				return ErrLoyaltySettings
			}
		}

		lines := make([]checkoutLine, 0, len(in.Items))
		var subtotal int64
		shortage := false
		offline := in.ClientTxID != ""

		for _, it := range in.Items {
			line, err := r.resolveCheckoutItem(ctx, tx, in.StoreID, in.OutletID, it, !offline)
			if err != nil {
				return err
			}
			if offline && line.stock < line.qty {
				shortage = true
			}
			subtotal += line.item.Price * int64(line.qty)
			lines = append(lines, line)
		}
		promoLines := make([]promoLine, len(lines))
		for i, l := range lines {
			promoLines[i] = promoLine{Product: l.product, Value: l.item.Price * int64(l.qty)}
		}
		promo, err := r.resolvePromo(ctx, tx, store, subtotal, promoLines, in.VoucherCode)
		if err != nil {
			return err
		}

		discount := in.Discount
		var promoID *uint
		var promoName string
		var promoDiscount int64
		var voucherID *uint
		var voucherCode string
		var voucherDiscount int64
		if promo != nil {
			discount = 0
			if promo.PromoID != 0 {
				id := promo.PromoID
				promoID = &id
				promoName = promo.PromoName
				promoDiscount = promo.PromoDiscount
			}
			if promo.VoucherID != 0 {
				id := promo.VoucherID
				voucherID = &id
				voucherCode = promo.VoucherCode
				voucherDiscount = promo.VoucherDiscount
			}
		}
		if discount < 0 || discount > subtotal {
			return ErrBadDiscount
		}

		var tax int64
		base := subtotal - discount - promoDiscount - voucherDiscount
		total := base
		if store.TaxEnabled && store.TaxPct > 0 {
			if store.TaxInclusive {
				// Harga sudah termasuk pajak: total tetap = base,
				// tax = base - base/(1+pct/100).
				tax = applyTaxRounding(float64(base)-float64(base)/(1+store.TaxPct/100), store.TaxRounding)
			} else {
				tax = applyTaxRounding(float64(base)*store.TaxPct/100, store.TaxRounding)
				total = base + tax
			}
		}

		pointsRedeemed, pointsDiscount := int64(0), int64(0)
		if store.LoyaltyEnabled && customer != nil {
			pointsRedeemed = in.PointsRedeem
			if store.LoyaltyPointValue > 0 && pointsRedeemed > int64(^uint64(0)>>1)/store.LoyaltyPointValue {
				return ErrPointsRedeemInvalid
			}
			pointsDiscount = pointsRedeemed * store.LoyaltyPointValue
			if pointsDiscount > total {
				return ErrPointsExceedTotal
			}
		}
		payable := total - pointsDiscount
		paid, change := in.Paid, int64(0)
		if in.Method == string(PayCash) {
			if paid < payable {
				return ErrPaidInsufficient
			}
			change = paid - payable
		} else if in.Method == string(PayCredit) {
			paid = 0
		} else {
			paid = payable
		}
		pointsEarned := int64(0)
		if store.LoyaltyEnabled && customer != nil {
			pointsEarned = total / store.LoyaltyEarnPer
			balanceAfterRedeem := customer.Points - pointsRedeemed
			if pointsEarned > int64(^uint64(0)>>1)-balanceAfterRedeem {
				return ErrLoyaltySettings
			}
		}

		trxItemsList := make(TrxItemList, len(lines))
		for i, l := range lines {
			trxItemsList[i] = l.item
		}

		status := TrxCompleted
		if offline && shortage {
			status = TrxCompletedOfflineShortage
		}
		trx := Trx{
			StoreID:         in.StoreID,
			OutletID:        uintPointer(in.OutletID),
			ClientTxID:      in.ClientTxID,
			CashierID:       in.CashierID,
			CashierName:     in.CashierName,
			Items:           trxItemsList,
			Subtotal:        subtotal,
			Discount:        discount,
			PromoID:         promoID,
			PromoName:       promoName,
			PromoDiscount:   promoDiscount,
			VoucherID:       voucherID,
			VoucherCode:     voucherCode,
			VoucherDiscount: voucherDiscount,
			Tax:             tax,
			Total:           total,
			Method:          in.Method,
			Paid:            paid,
			Change:          change,
			Status:          status,
			Customer:        in.Customer,
			CustomerID:      in.CustomerID,
			PointsEarned:    pointsEarned,
			PointsRedeemed:  pointsRedeemed,
			PointsDiscount:  pointsDiscount,
		}

		if err := tx.Create(&trx).Error; err != nil {
			return mapDBErr(err)
		}
		if in.Method == string(PayCredit) {
			location, err := time.LoadLocation(store.Timezone)
			if err != nil {
				location, _ = time.LoadLocation("Asia/Makassar")
			}
			transactionDate := trx.CreatedAt.In(location)
			dueDate := time.Date(transactionDate.Year(), transactionDate.Month(), transactionDate.Day(), 0, 0, 0, 0, location).AddDate(0, 0, store.CreditDueDays)
			status := finance.DebtUnpaid
			if payable == 0 {
				status = finance.DebtPaid
			}
			receivable := finance.Receivable{
				StoreID: in.StoreID, CustomerID: *in.CustomerID, TransactionID: trx.ID,
				Amount: payable, DueDate: dueDate, Status: status,
			}
			if err := tx.Create(&receivable).Error; err != nil {
				return err
			}
		}
		if store.LoyaltyEnabled && customer != nil && (pointsRedeemed > 0 || pointsEarned > 0) {
			res := tx.Model(&cust.Customer{}).Where("id = ? AND store_id = ?", customer.ID, in.StoreID).
				Update("points", gorm.Expr("points - ? + ?", pointsRedeemed, pointsEarned))
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return ErrCustomerNotFound
			}
		}

		for _, l := range lines {
			tItem := TransactionItem{
				TrxID:       trx.ID,
				ProductID:   l.item.ProductID,
				VariantID:   l.item.VariantID,
				VariantName: l.item.VariantName,
				Name:        l.item.Name,
				BuyPrice:    l.item.BuyPrice,
				Price:       l.item.Price,
				Qty:         l.qty,
			}
			if err := tx.Create(&tItem).Error; err != nil {
				return err
			}

			_, _, actualDelta, err := inventory.ChangeOutletStock(tx, in.StoreID, in.OutletID, l.item.ProductID, l.item.VariantID, -l.qty, offline)
			if err != nil {
				return err
			}
			deductQty := -actualDelta

			if deductQty > 0 {
				mv := inventory.Movement{
					StoreID:   in.StoreID,
					OutletID:  uintPointer(in.OutletID),
					ProductID: l.item.ProductID,
					VariantID: l.item.VariantID,
					Type:      inventory.MovementSale,
					Qty:       -deductQty,
					Reason:    fmt.Sprintf("Penjualan %d", trx.ID),
					Actor:     in.CashierName,
				}
				if err := tx.Create(&mv).Error; err != nil {
					return err
				}
			}
		}

		trx.FinalStocks, err = r.loadFinalStocks(ctx, tx, in.StoreID, in.OutletID, trxItemsList)
		if err != nil {
			return err
		}
		resultTrx = &trx
		return nil
	})

	if err != nil {
		if in.ClientTxID != "" {
			if existing, lookupErr := r.findByClientTxID(ctx, in.StoreID, in.ClientTxID); lookupErr == nil {
				return existing, false, nil
			}
		}
		return nil, false, err
	}
	if existingTrxID != 0 {
		existing, err := r.findByClientTxID(ctx, in.StoreID, in.ClientTxID)
		if err != nil {
			return nil, false, err
		}
		return existing, false, nil
	}
	return resultTrx, true, nil
}

func (r *Repo) findByClientTxID(ctx context.Context, storeID uint, clientTxID string) (*Trx, error) {
	var trx Trx
	if err := r.DB(ctx).Where("store_id = ? AND client_tx_id = ?", storeID, clientTxID).First(&trx).Error; err != nil {
		return nil, mapDBErr(err)
	}
	fullTrx, err := r.GetByID(ctx, storeID, trx.ID)
	if err != nil {
		return nil, err
	}
	finalOutletID := uint(0)
	if fullTrx.OutletID != nil {
		finalOutletID = *fullTrx.OutletID
	} else {
		var outlet outlet.Outlet
		if err := r.DB(ctx).Where("store_id = ? AND is_default = ?", storeID, true).First(&outlet).Error; err != nil {
			return nil, err
		}
		finalOutletID = outlet.ID
	}
	fullTrx.FinalStocks, err = r.loadFinalStocks(ctx, r.DB(ctx), storeID, finalOutletID, fullTrx.Items)
	if err != nil {
		return nil, err
	}
	return fullTrx, nil
}

func (r *Repo) loadFinalStocks(ctx context.Context, tx *gorm.DB, storeID, outletID uint, items []TrxItem) ([]FinalStock, error) {
	stocks := make([]FinalStock, 0, len(items))
	for _, item := range items {
		var product catalog.Product
		if err := tx.WithContext(ctx).Select("id").Where("id = ? AND store_id = ?", item.ProductID, storeID).Take(&product).Error; err != nil {
			return nil, err
		}
		stock, err := inventory.GetOutletStock(tx.WithContext(ctx), outletID, item.ProductID, item.VariantID)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, FinalStock{ProductID: item.ProductID, VariantID: item.VariantID, Stock: stock})
	}
	return stocks, nil
}

func (r *Repo) Refund(ctx context.Context, storeID, trxID uint, items []RefundItem, reason, byName string) (*Trx, error) {
	var outlet outlet.Outlet
	if err := r.DB(ctx).Where("store_id = ? AND is_default = ?", storeID, true).First(&outlet).Error; err != nil {
		return nil, mapDBErr(err)
	}
	return r.RefundAtOutlet(ctx, storeID, outlet.ID, trxID, items, reason, byName)
}

func (r *Repo) RefundAtOutlet(ctx context.Context, storeID, outletID, trxID uint, items []RefundItem, reason, byName string) (*Trx, error) {
	if len(items) == 0 {
		return nil, ErrEmptyItems
	}
	var outletCount int64
	if err := r.DB(ctx).Model(&outlet.Outlet{}).Where("id = ? AND store_id = ?", outletID, storeID).Count(&outletCount).Error; err != nil {
		return nil, err
	}
	if outletCount == 0 {
		return nil, ErrNotFound
	}
	if strings.TrimSpace(reason) == "" {
		return nil, fmt.Errorf("alasan refund wajib diisi")
	}

	err := r.tx.WithinTx(ctx, func(ctx context.Context) error {
		tx := r.DB(ctx)
		if r.isPostgres(ctx) {
			if err := tx.Exec("SELECT pg_advisory_xact_lock(hashtext(?))", strconv.FormatUint(uint64(storeID), 10)).Error; err != nil {
				return err
			}
		}
		var t Trx
		q := tx
		if r.isPostgres(ctx) {
			q = q.Clauses(clause.Locking{Strength: "UPDATE"})
		}
		if err := q.Where("id = ? AND store_id = ?", trxID, storeID).First(&t).Error; err != nil {
			if isNotFound(err) {
				return ErrNotFound
			}
			return err
		}
		if t.Status != TrxCompleted && t.Status != TrxCompletedOfflineShortage {
			return ErrNotRefundable
		}

		var tItems []TransactionItem
		if err := tx.Where("trx_id = ?", trxID).Find(&tItems).Error; err != nil {
			return err
		}

		soldProducts := map[uint]TrxItem{}
		soldVariants := map[uint]TrxItem{}
		for _, ti := range tItems {
			item := TrxItem{
				ProductID: ti.ProductID, VariantID: ti.VariantID, VariantName: ti.VariantName,
				Name: ti.Name, BuyPrice: ti.BuyPrice, Price: ti.Price, Qty: ti.Qty,
			}
			if ti.VariantID != nil {
				if prev, ok := soldVariants[*ti.VariantID]; ok {
					prev.Qty += ti.Qty
					soldVariants[*ti.VariantID] = prev
				} else {
					soldVariants[*ti.VariantID] = item
				}
				continue
			}
			if prev, ok := soldProducts[ti.ProductID]; ok {
				prev.Qty += ti.Qty
				soldProducts[ti.ProductID] = prev
			} else {
				soldProducts[ti.ProductID] = item
			}
		}

		var refunds []Refund
		if err := tx.Where("trx_id = ?", trxID).Find(&refunds).Error; err != nil {
			return err
		}

		refundedProducts := map[uint]int{}
		refundedVariants := map[uint]int{}
		for _, ref := range refunds {
			for pid, q := range ref.Items {
				refundedProducts[pid] += q
			}
			for variantID, q := range ref.VariantItems {
				refundedVariants[variantID] += q
			}
		}

		wantedProducts := map[uint]int{}
		wantedVariants := map[uint]int{}
		variantProducts := map[uint]uint{}
		for _, item := range items {
			if item.VariantID == nil {
				wantedProducts[item.ProductID] += item.Qty
			} else {
				if productID, exists := variantProducts[*item.VariantID]; exists && productID != item.ProductID {
					return ErrRefundTooMuch
				}
				wantedVariants[*item.VariantID] += item.Qty
				variantProducts[*item.VariantID] = item.ProductID
			}
		}
		processedProducts := RefundItemList{}
		processedVariants := map[uint]int{}
		var refundValue int64
		for pid, want := range wantedProducts {
			s, ok := soldProducts[pid]
			if !ok || want < 0 {
				return ErrRefundTooMuch
			}
			if want == 0 {
				continue
			}
			if refundedProducts[pid]+want > s.Qty {
				return ErrRefundTooMuch
			}
			itemValue := s.Price * int64(want)
			refundValue += itemValue + roundRatioHalfUp(t.Tax, itemValue, t.Subtotal-t.Discount-t.PromoDiscount-t.VoucherDiscount)

			if _, _, _, err := inventory.ChangeOutletStock(tx, storeID, outletID, pid, nil, want, false); err != nil {
				return err
			}

			mv := inventory.Movement{
				StoreID:   storeID,
				OutletID:  uintPointer(outletID),
				ProductID: pid,
				Type:      inventory.MovementRefund,
				Qty:       want,
				Reason:    fmt.Sprintf("Refund %d", trxID),
				Actor:     byName,
			}
			if err := tx.Create(&mv).Error; err != nil {
				return err
			}

			processedProducts[pid] = want
		}
		for variantID, want := range wantedVariants {
			sold, ok := soldVariants[variantID]
			if !ok || sold.ProductID != variantProducts[variantID] || want < 0 {
				return ErrRefundTooMuch
			}
			if want == 0 {
				continue
			}
			if refundedVariants[variantID]+want > sold.Qty {
				return ErrRefundTooMuch
			}
			itemValue := sold.Price * int64(want)
			refundValue += itemValue + roundRatioHalfUp(t.Tax, itemValue, t.Subtotal-t.Discount-t.PromoDiscount-t.VoucherDiscount)
			if _, _, _, err := inventory.ChangeOutletStock(tx, storeID, outletID, sold.ProductID, &variantID, want, false); errors.Is(err, ErrNotFound) {
				return ErrVariantInvalid
			} else if err != nil {
				return err
			}
			movement := inventory.Movement{
				StoreID: storeID, OutletID: uintPointer(outletID), ProductID: sold.ProductID, VariantID: &variantID,
				Type: inventory.MovementRefund, Qty: want, Reason: fmt.Sprintf("Refund %d", trxID), Actor: byName,
			}
			if err := tx.Create(&movement).Error; err != nil {
				return err
			}
			processedVariants[variantID] = want
		}

		if len(processedProducts)+len(processedVariants) == 0 {
			return ErrEmptyItems
		}

		full := true
		for pid, s := range soldProducts {
			if refundedProducts[pid]+processedProducts[pid] < s.Qty {
				full = false
				break
			}
		}
		for variantID, s := range soldVariants {
			if refundedVariants[variantID]+processedVariants[variantID] < s.Qty {
				full = false
				break
			}
		}

		refRecord := Refund{
			StoreID: storeID,
			TrxID:   trxID,
			Items:   processedProducts, VariantItems: processedVariants,
			Reason: strings.TrimSpace(reason), ByName: byName,
		}
		if err := tx.Create(&refRecord).Error; err != nil {
			return err
		}
		if t.CustomerID != nil && t.PointsEarned > 0 {
			var store outlet.Store
			if err := tx.Where("id = ?", storeID).First(&store).Error; err != nil {
				return err
			}
			if store.LoyaltyEarnPer <= 0 {
				return ErrLoyaltySettings
			}
			pointsToRemove := refundValue / store.LoyaltyEarnPer
			if pointsToRemove > 0 {
				if err := tx.Model(&cust.Customer{}).Where("id = ? AND store_id = ?", *t.CustomerID, storeID).
					Update("points", gorm.Expr("CASE WHEN points > ? THEN points - ? ELSE 0 END", pointsToRemove, pointsToRemove)).Error; err != nil {
					return err
				}
			}
		}

		if full {
			if err := tx.Model(&Trx{}).Where("id = ? AND store_id = ?", trxID, storeID).Update("status", TrxRefunded).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, storeID, trxID)
}

func (r *Repo) List(ctx context.Context, storeID uint, outletID *uint, cashierID uint, q, method, date, from, to string, status string, cashierFilter *uint, page, limit int) ([]*Trx, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 200 {
		limit = 200
	}

	query := r.DB(ctx).Model(&Trx{}).Where("store_id = ?", storeID)
	if outletID != nil {
		query = query.Where("outlet_id = ?", *outletID)
	}
	if cashierID != 0 {
		query = query.Where("cashier_id = ?", cashierID)
	}
	if qs := strings.TrimSpace(q); qs != "" {
		searchTerm := "%" + strings.ToLower(qs) + "%"
		// IDs are numeric now: exact match when q is a number, name search otherwise.
		if n, err := strconv.ParseUint(qs, 10, 64); err == nil {
			query = query.Where("id = ? OR lower(cashier_name) LIKE ?", n, searchTerm)
		} else {
			query = query.Where("lower(cashier_name) LIKE ?", searchTerm)
		}
	}
	if method != "" {
		query = query.Where("method = ?", method)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if cashierFilter != nil {
		query = query.Where("cashier_id = ?", *cashierFilter)
	}
	if ds := strings.TrimSpace(date); ds != "" {
		// date = hari kalender lokal toko (00:00–23:59). Invalid/kosong = abaikan.
		tz := "Asia/Makassar"
		var st outlet.Store
		if err := r.DB(ctx).Where("id = ?", storeID).First(&st).Error; err == nil && st.Timezone != "" {
			tz = st.Timezone
		}
		loc, err := time.LoadLocation(tz)
		if err != nil {
			loc, err = time.LoadLocation("Asia/Makassar")
			if err != nil {
				loc = time.UTC
			}
		}
		if parsed, perr := time.Parse("2006-01-02", ds); perr == nil && parsed.Format("2006-01-02") == ds {
			dayStart := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, loc)
			dayEnd := dayStart.AddDate(0, 0, 1)
			query = query.Where("created_at >= ? AND created_at < ?", dayStart.UTC(), dayEnd.UTC())
		}
	}
	if fs, ts := strings.TrimSpace(from), strings.TrimSpace(to); fs != "" && ts != "" {
		// Rentang khusus dari pemilih tanggal UI: inklusif per hari kalender lokal toko.
		tz := "Asia/Makassar"
		var st outlet.Store
		if err := r.DB(ctx).Where("id = ?", storeID).First(&st).Error; err == nil && st.Timezone != "" {
			tz = st.Timezone
		}
		loc := loadLoc(tz)
		if f, ok := parseLocalDay(fs, loc); ok {
			if t, ok2 := parseLocalDay(ts, loc); ok2 && !t.Before(f) {
				query = query.Where("created_at >= ? AND created_at < ?", f.UTC(), t.AddDate(0, 0, 1).UTC())
			}
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	list := make([]*Trx, 0)
	offset := (page - 1) * limit
	err := query.Order("created_at DESC, id DESC").Limit(limit).Offset(offset).Find(&list).Error
	if err != nil {
		return nil, 0, err
	}

	if len(list) > 0 {
		ids := make([]uint, len(list))
		for i, t := range list {
			ids[i] = t.ID
		}
		var items []TransactionItem
		if err := r.DB(ctx).Where("trx_id IN ?", ids).Order("id ASC").Find(&items).Error; err == nil {
			byID := map[uint]*Trx{}
			for _, t := range list {
				t.Items = []TrxItem{}
				byID[t.ID] = t
			}
			for _, it := range items {
				if t, ok := byID[it.TrxID]; ok {
					t.Items = append(t.Items, TrxItem{
						ProductID:   it.ProductID,
						VariantID:   it.VariantID,
						VariantName: it.VariantName,
						Name:        it.Name,
						BuyPrice:    it.BuyPrice,
						Price:       it.Price,
						Qty:         it.Qty,
					})
				}
			}
		}
	}

	return list, int(total), nil
}

func (r *Repo) ListForExport(ctx context.Context, storeID uint, period, timezone, dateStr, fromStr, toStr string) ([]*Trx, error) {
	loc := loadLoc(timezone)
	start, end, err := periodStarts(time.Now().UTC(), loc, period)
	if err != nil {
		return nil, err
	}
	if anchor, ok := parseLocalDay(dateStr, loc); ok {
		switch period {
		case "today":
			start, end = anchor, anchor.AddDate(0, 0, 1)
		case "yesterday":
			start, end = anchor.AddDate(0, 0, -1), anchor
		}
	}
	if f, ok := parseLocalDay(fromStr, loc); ok {
		if t, ok2 := parseLocalDay(toStr, loc); ok2 && !t.Before(f) {
			start, end = f, t.AddDate(0, 0, 1)
		}
	}

	query := r.DB(ctx).Where("store_id = ?", storeID)
	if !start.IsZero() {
		query = query.Where("created_at >= ? AND created_at < ?", start.UTC(), end.UTC())
	}
	list := make([]*Trx, 0)
	if err := query.Order("created_at ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return list, nil
	}

	byID := make(map[uint]*Trx, len(list))
	for _, trx := range list {
		trx.Items = []TrxItem{}
		byID[trx.ID] = trx
	}
	var items []TransactionItem
	itemQuery := r.DB(ctx).
		Joins("JOIN transactions ON transactions.id = transaction_items.trx_id").
		Where("transactions.store_id = ?", storeID)
	if !start.IsZero() {
		itemQuery = itemQuery.Where("transactions.created_at >= ? AND transactions.created_at < ?", start.UTC(), end.UTC())
	}
	if err := itemQuery.Order("transaction_items.id ASC").Find(&items).Error; err != nil {
		return nil, err
	}
	for _, item := range items {
		if trx, ok := byID[item.TrxID]; ok {
			trx.Items = append(trx.Items, TrxItem{
				ProductID:   item.ProductID,
				VariantID:   item.VariantID,
				VariantName: item.VariantName,
				Name:        item.Name,
				BuyPrice:    item.BuyPrice,
				Price:       item.Price,
				Qty:         item.Qty,
			})
		}
	}
	return list, nil
}

func (r *Repo) GetByID(ctx context.Context, storeID, id uint) (*Trx, error) {
	var t Trx
	if err := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).First(&t).Error; err != nil {
		return nil, mapDBErr(err)
	}

	t.Items = []TrxItem{}
	var items []TransactionItem
	if err := r.DB(ctx).Where("trx_id = ?", id).Order("id ASC").Find(&items).Error; err == nil {
		for _, it := range items {
			t.Items = append(t.Items, TrxItem{
				ProductID:   it.ProductID,
				VariantID:   it.VariantID,
				VariantName: it.VariantName,
				Name:        it.Name,
				BuyPrice:    it.BuyPrice,
				Price:       it.Price,
				Qty:         it.Qty,
			})
		}
	}

	var refunds []Refund
	if err := r.DB(ctx).
		Where("store_id = ? AND trx_id = ?", storeID, id).
		Order("created_at ASC, id ASC").
		Find(&refunds).Error; err != nil {
		return nil, err
	}
	t.Refunds = make([]RefundDetail, 0, len(refunds))
	if len(refunds) > 0 {
		for _, ref := range refunds {
			detail := RefundDetail{
				ID:        ref.ID,
				CreatedAt: ref.CreatedAt,
				Items:     []TrxItem{},
				Reason:    ref.Reason,
				Actor:     ref.ByName,
			}
			for _, item := range t.Items {
				qty := ref.Items[item.ProductID]
				if item.VariantID != nil {
					qty = ref.VariantItems[*item.VariantID]
				}
				if qty > 0 {
					item.Qty = qty
					detail.Items = append(detail.Items, item)
				}
			}
			t.Refunds = append(t.Refunds, detail)
		}
	}
	return &t, nil
}

func (r *Repo) PreviewPromo(ctx context.Context, in PromoPreviewInput) (*promo.Preview, error) {
	if len(in.Items) == 0 {
		return nil, ErrEmptyItems
	}
	var store outlet.Store
	if err := r.DB(ctx).Where("id = ?", in.StoreID).First(&store).Error; err != nil {
		return nil, mapDBErr(err)
	}
	if in.OutletID == 0 {
		var outlet outlet.Outlet
		if err := r.DB(ctx).Where("store_id = ? AND is_default = ?", in.StoreID, true).First(&outlet).Error; err != nil {
			return nil, mapDBErr(err)
		}
		in.OutletID = outlet.ID
	}
	lines := make([]promoLine, 0, len(in.Items))
	var subtotal int64
	for _, item := range mergeCheckoutItems(in.Items) {
		line, err := r.resolveCheckoutItem(ctx, r.DB(ctx), in.StoreID, in.OutletID, item, false)
		if err != nil {
			return nil, err
		}
		value := line.item.Price * int64(item.Qty)
		subtotal += value
		lines = append(lines, promoLine{Product: line.product, Value: value})
	}
	return r.resolvePromo(ctx, r.DB(ctx), store, subtotal, lines, in.VoucherCode)
}

// evalPromoDiscount menilai satu promo terhadap keranjang: mengembalikan nilai
// yang tercakup (covered) beserta diskonnya. ok=false bila promo tidak berlaku
// (di luar periode, di bawah min. belanja, di luar happy hour, atau tidak ada
// item yang tercakup).
func evalPromoDiscount(promo *promo.Promo, store outlet.Store, subtotal int64, lines []promoLine, now time.Time) (int64, int64, bool, error) {
	if now.Before(promo.StartsAt) || !now.Before(promo.EndsAt) || subtotal < promo.MinPurchase {
		return 0, 0, false, nil
	}
	if promo.HappyStart != nil && promo.HappyEnd != nil {
		location, err := time.LoadLocation(store.Timezone)
		if err != nil {
			return 0, 0, false, fmt.Errorf("zona waktu toko tidak valid: %w", err)
		}
		if !promoHappyHourContains(*promo.HappyStart, *promo.HappyEnd, now.In(location)) {
			return 0, 0, false, nil
		}
	}
	var covered int64
	for _, line := range lines {
		switch promo.Scope {
		case "all":
			covered += line.Value
		case "category":
			if line.Product.CategoryID != nil && promo.CategoryID != nil && *line.Product.CategoryID == *promo.CategoryID {
				covered += line.Value
			}
		case "product":
			if promo.ProductID != nil && line.Product.ID == *promo.ProductID {
				covered += line.Value
			}
		}
	}
	if covered == 0 {
		return 0, 0, false, nil
	}
	return covered, promoDiscountFor(promo, covered), true, nil
}

// promoDiscountFor menghitung diskon satu promo dari suatu basis nilai,
// dibatasi basis itu sendiri dan MaxDiscount promo.
func promoDiscountFor(promo *promo.Promo, base int64) int64 {
	if base <= 0 {
		return 0
	}
	var discount int64
	if promo.Type == "percent" {
		discount = roundPercentHalfUp(base, promo.Value)
		if discount > base {
			discount = base
		}
		if promo.MaxDiscount > 0 && discount > promo.MaxDiscount {
			discount = promo.MaxDiscount
		}
	} else if promo.Type == "amount" {
		discount = promo.AmountValue
		if discount > base {
			discount = base
		}
	}
	return discount
}

// resolvePromo menyelesaikan dua lapisan diskon (keputusan produk 2026-10-15):
// lapisan 1 promo otomatis terbaik (tanpa kode); lapisan 2 voucher berkode.
// Voucher yang ditandai Combinable menumpuk DI ATAS promo otomatis — nilainya
// dihitung dari sisa nilai tercakup setelah promo — sedangkan voucher biasa
// menggantikan promo otomatis. Min. belanja selalu dinilai dari subtotal awal.
func (r *Repo) resolvePromo(ctx context.Context, tx *gorm.DB, store outlet.Store, subtotal int64, lines []promoLine, voucherCode string) (*promo.Preview, error) {
	voucherCode = strings.TrimSpace(voucherCode)
	now := time.Now().UTC()

	var autos []promo.Promo
	if err := tx.WithContext(ctx).Where("store_id = ? AND active = ? AND voucher_code IS NULL", store.ID, true).Order("id ASC").Find(&autos).Error; err != nil {
		return nil, err
	}
	var best *promo.Preview
	for i := range autos {
		_, discount, ok, err := evalPromoDiscount(&autos[i], store, subtotal, lines, now)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		if best == nil || discount > best.PromoDiscount {
			best = &promo.Preview{PromoID: autos[i].ID, PromoName: autos[i].Name, PromoDiscount: discount}
		}
	}

	if voucherCode == "" {
		return best, nil
	}

	var coded []promo.Promo
	if err := tx.WithContext(ctx).Where("store_id = ? AND active = ? AND voucher_code = ?", store.ID, true, voucherCode).Find(&coded).Error; err != nil {
		return nil, err
	}
	if len(coded) == 0 {
		return nil, ErrPromoVoucherInvalid
	}
	voucher := &coded[0]
	covered, discount, ok, err := evalPromoDiscount(voucher, store, subtotal, lines, now)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrPromoVoucherInvalid
	}
	out := &promo.Preview{VoucherID: voucher.ID, VoucherName: voucher.Name, VoucherCode: *voucher.VoucherCode, VoucherDiscount: discount}
	if best != nil && voucher.Combinable {
		base := covered - best.PromoDiscount
		if base < 0 {
			base = 0
		}
		out.PromoID = best.PromoID
		out.PromoName = best.PromoName
		out.PromoDiscount = best.PromoDiscount
		out.VoucherDiscount = promoDiscountFor(voucher, base)
	}
	return out, nil
}

func promoHappyHourContains(start, end string, local time.Time) bool {
	startTime, startErr := time.Parse("15:04", start)
	endTime, endErr := time.Parse("15:04", end)
	if startErr != nil || endErr != nil || len(start) != 5 || len(end) != 5 {
		return false
	}
	startMinute := startTime.Hour()*60 + startTime.Minute()
	endMinute := endTime.Hour()*60 + endTime.Minute()
	currentMinute := local.Hour()*60 + local.Minute()
	if startMinute <= endMinute {
		return currentMinute >= startMinute && currentMinute < endMinute
	}
	return currentMinute >= startMinute || currentMinute < endMinute
}

func roundHalfUp(f float64) int64 {
	return int64(f + 0.5)
}

func roundPercentHalfUp(amount int64, percent float64) int64 {
	if amount <= 0 || percent <= 0 || percent > 100 || math.IsNaN(percent) || math.IsInf(percent, 0) {
		return 0
	}
	rate, ok := new(big.Rat).SetString(strconv.FormatFloat(percent, 'f', -1, 64))
	if !ok {
		return 0
	}
	numerator := new(big.Int).Mul(big.NewInt(amount), rate.Num())
	denominator := new(big.Int).Mul(big.NewInt(100), rate.Denom())
	quotient, remainder := new(big.Int).QuoRem(numerator, denominator, new(big.Int))
	if new(big.Int).Lsh(remainder, 1).Cmp(denominator) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}
	return quotient.Int64()
}

// applyTaxRounding menerapkan mode pembulatan pajak: none = round_half_up
// (perilaku lama), down = floor, up = ceil.
func applyTaxRounding(f float64, mode string) int64 {
	switch mode {
	case "down":
		return int64(math.Floor(f))
	case "up":
		return int64(math.Ceil(f))
	default:
		return roundHalfUp(f)
	}
}

// mapDBErr, isNotFound, loadLoc, parseLocalDay, periodStarts, dan
// roundRatioHalfUp adalah salinan privat slice dari helper repo lama
// (repo/base.go dan repo/report.go) — badan verbatim; periodStarts
// dipakai ListForExport, sisanya dipakai List dan Refund.
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

func roundRatioHalfUp(value, numerator, denominator int64) int64 {
	if denominator <= 0 || value <= 0 || numerator <= 0 {
		return 0
	}
	product := new(big.Int).Mul(big.NewInt(value), big.NewInt(numerator))
	product.Add(product, big.NewInt(denominator/2))
	return product.Div(product, big.NewInt(denominator)).Int64()
}
