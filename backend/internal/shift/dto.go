package shift

import (
	"time"
)

// DTO ringkasan/detail shift (dipindah verbatim dari service/shift.go,
// BE-F5) beserta tipe input handler shift (dari handler/shift.go).

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
	Items []*trx `json:"items"`
	Page  int    `json:"page"`
	Limit int    `json:"limit"`
	Total int64  `json:"total"`
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

// startShiftReq adalah bentuk JSON permintaan buka shift (startShiftReq
// lama di paket handler); tag JSON tidak berubah.
type startShiftReq struct {
	OpeningCash int64 `json:"opening_cash"`
}

// closeShiftReq adalah bentuk JSON permintaan tutup shift
// (closeShiftReq lama di paket handler); tag JSON tidak berubah.
type closeShiftReq struct {
	ClosingCash *int64  `json:"closingCash"`
	Note        *string `json:"note"`
}
