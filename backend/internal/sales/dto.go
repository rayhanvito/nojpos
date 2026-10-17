package sales

// CheckoutCmd dan PromoPreviewCmd adalah perintah service transaksi
// (dipindah verbatim dari service/transaction.go, BE-F8) beserta tipe
// itemnya.
type CheckoutCmd struct {
	Items        []CheckoutItemCmd
	Discount     int64
	VoucherCode  string
	Method       string
	Paid         int64
	Customer     string
	CustomerID   *uint
	PointsRedeem int64
	ClientTxID   string
}

type CheckoutItemCmd struct {
	ProductID uint
	VariantID *uint
	Qty       int
}

type PromoPreviewCmd struct {
	Items       []CheckoutItemCmd
	VoucherCode string
}

// checkoutItemReq, checkoutReq, dan refundReq adalah bentuk JSON
// permintaan handler transaksi (dari handler/transaction.go); tag JSON
// tidak berubah.
type checkoutItemReq struct {
	ProductID uint  `json:"productId"`
	VariantID *uint `json:"variantId"`
	Qty       int   `json:"qty"`
}

type checkoutReq struct {
	Items        []checkoutItemReq `json:"items"`
	Discount     int64             `json:"discount"`
	VoucherCode  string            `json:"voucherCode"`
	Method       string            `json:"method"`
	Paid         int64             `json:"paid"`
	Customer     string            `json:"customer"`
	CustomerID   *uint             `json:"customerId"`
	PointsRedeem int64             `json:"pointsRedeem"`
	ClientTxID   string            `json:"clientTxId"`
}

type refundReq struct {
	Items  []checkoutItemReq `json:"items"`
	Reason string            `json:"reason"`
}
