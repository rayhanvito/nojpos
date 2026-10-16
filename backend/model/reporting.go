package model

import "time"

type StoreHours struct {
	Days  string  `json:"days"`
	Open  *string `json:"open"`
	Close *string `json:"close"`
}

type StoreSettings struct {
	Name                string       `json:"storeName"`
	Address             string       `json:"address"`
	Phone               string       `json:"phone"`
	TaxEnabled          bool         `json:"taxEnabled"`
	TaxPct              float64      `json:"taxPct"`
	ReceiptHeader       string       `json:"receiptHeader"`
	ReceiptFooter       string       `json:"receiptFooter"`
	Paper               string       `json:"paper"`
	Timezone            string       `json:"timezone"`
	BusinessType        string       `json:"businessType"`
	Email               string       `json:"email"`
	City                string       `json:"city"`
	Province            string       `json:"province"`
	Currency            string       `json:"currency"`
	Hours               []StoreHours `json:"hours"`
	ReceiptShowLogo     bool         `json:"receiptShowLogo"`
	ReceiptShowCashier  bool         `json:"receiptShowCashier"`
	ReceiptShowMethod   bool         `json:"receiptShowMethod"`
	ReceiptShowTax      bool         `json:"receiptShowTax"`
	ReceiptShowDiscount bool         `json:"receiptShowDiscount"`
	ReceiptShowNote     bool         `json:"receiptShowNote"`
	TaxName             string       `json:"taxName"`
	TaxInclusive        bool         `json:"taxInclusive"`
	TaxRounding         string       `json:"taxRounding"`
	TaxApplyTo          string       `json:"taxApplyTo"`
	LoyaltyEnabled      bool         `json:"loyalty_enabled"`
	LoyaltyEarnPer      int64        `json:"loyalty_earn_per"`
	LoyaltyPointValue   int64        `json:"loyalty_point_value"`
	CreditDueDays       int          `json:"creditDueDays"`
	LowStockDefault     int          `json:"lowStockDefault"`
}

type DayPoint struct {
	Date  string `json:"date"`
	Omzet int64  `json:"omzet"`
}

type MethodPoint struct {
	Method string `json:"method"`
	Total  int64  `json:"total"`
}

type TopProduct struct {
	ProductID uint   `json:"product_id"`
	Name      string `json:"name"`
	Qty       int    `json:"qty"`
	Revenue   int64  `json:"revenue"`
}

type TrxBrief struct {
	ID          uint      `json:"id"`
	CashierName string    `json:"cashier_name"`
	Total       int64     `json:"total"`
	Status      TrxStatus `json:"status"`
	Time        time.Time `json:"time"`
}

type DashboardToday struct {
	Omzet       int64 `json:"omzet"`
	TrxCount    int64 `json:"trx_count"`
	ItemsSold   int64 `json:"items_sold"`
	RefundTotal int64 `json:"refund_total"`
	LowStock    int64 `json:"low_stock"`
}

type DashboardAdmin struct {
	Role        string         `json:"role"`
	Today       DashboardToday `json:"today"`
	Sales7      []DayPoint     `json:"sales7"`
	Methods     []MethodPoint  `json:"methods"`
	TopProducts []TopProduct   `json:"top_products"`
	Recent      []TrxBrief     `json:"recent"`
}

type DashboardCashier struct {
	Role   string         `json:"role"`
	Today  DashboardToday `json:"today"`
	Recent []TrxBrief     `json:"recent"`
}

type ReportSummary struct {
	Omzet       int64 `json:"omzet"`
	TrxCount    int64 `json:"trx_count"`
	ItemsSold   int64 `json:"items_sold"`
	GrossProfit int64 `json:"gross_profit"`
	RefundTotal int64 `json:"refund_total"`
	RefundCount int64 `json:"refund_count"`
	// Total diskon menurut sumbernya: promo otomatis vs voucher berkode.
	PromoDiscountTotal   int64 `json:"promo_discount_total"`
	VoucherDiscountTotal int64 `json:"voucher_discount_total"`
	// Total pengeluaran operasional pada periode laporan.
	ExpenseTotal int64 `json:"expense_total"`
}

type TaxReport struct {
	Month            string `json:"month"`
	Gross            int64  `json:"omzet_bruto"`
	RefundTotal      int64  `json:"refund_total"`
	Net              int64  `json:"omzet_neto"`
	DPP              int64  `json:"dpp"`
	PPNCollected     int64  `json:"ppn_terkumpul"`
	TrxCount         int64  `json:"trx_count"`
	RefundCount      int64  `json:"refund_count"`
	PPhFinalEstimate int64  `json:"pph_final_estimate"`
}

type ExpenseCategoryRow struct {
	Category string `json:"category"`
	Total    int64  `json:"total"`
	Count    int64  `json:"count"`
}

type ProductReportRow struct {
	ProductID uint   `json:"product_id"`
	Name      string `json:"name"`
	SKU       string `json:"sku"`
	Qty       int    `json:"qty"`
	Revenue   int64  `json:"revenue"`
	Profit    int64  `json:"profit"`
}

type TrxProfitRow struct {
	Date    string    `json:"date"`
	ID      uint      `json:"id"`
	Cashier string    `json:"cashier"`
	Method  string    `json:"method"`
	Total   int64     `json:"total"`
	HPP     int64     `json:"hpp"`
	Profit  int64     `json:"profit"`
	Status  TrxStatus `json:"status"`
}

type StockRow struct {
	Name       string `json:"name"`
	SKU        string `json:"sku"`
	Stock      int    `json:"stock"`
	BuyPrice   int64  `json:"buy_price"`
	SellPrice  int64  `json:"sell_price"`
	StockValue int64  `json:"stock_value"`
}

type StatusCount struct {
	Status TrxStatus `json:"status"`
	Count  int64     `json:"count"`
}

type CashierReportRow struct {
	CashierID     uint   `json:"cashier_id"`
	Name          string `json:"name"`
	Trx           int    `json:"trx"`
	Items         int64  `json:"items"`
	Gross         int64  `json:"gross"`
	Refund        int64  `json:"refund"`
	Net           int64  `json:"net"`
	TopProduct    string `json:"top_product"`
	TopProductQty int    `json:"top_product_qty"`
}

type ReportBundle struct {
	Period       string             `json:"period"`
	Summary      ReportSummary      `json:"summary"`
	ByMethod     []MethodPoint      `json:"by_method"`
	ByStatus     []StatusCount      `json:"by_status"`
	Products     []ProductReportRow `json:"products"`
	Transactions []TrxProfitRow     `json:"transactions"`
	Cashiers     []CashierReportRow `json:"cashiers"`
	Stock        []StockRow         `json:"stock"`
	Expenses     []ExpenseCategoryRow `json:"expenses"`
}
