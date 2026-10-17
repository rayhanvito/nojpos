package sales

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// Handler adalah TrxHandler lama di handler/transaction.go, dipindah
// verbatim.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// pathUint, reportOutletFilter, clampPage, dan clampLimit adalah
// salinan privat slice dari helper paket handler lama
// (handler/params.go) — badan verbatim.
func pathUint(c *gin.Context, name string) (uint, bool) {
	n, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid."})
		return 0, false
	}
	return uint(n), true
}

func reportOutletFilter(c *gin.Context, admin bool, activeOutletID uint) (*uint, bool) {
	if !admin {
		return &activeOutletID, true
	}
	value, provided := c.GetQuery("outletId")
	if !provided {
		return nil, true
	}
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "outletId tidak valid."})
		return nil, false
	}
	outletID := uint(id)
	return &outletID, true
}

func clampLimit(limit int) int {
	if limit < 1 {
		return 20
	}
	if limit > 200 {
		return 200
	}
	return limit
}

func clampPage(page int) int {
	if page < 1 {
		return 1
	}
	return page
}

func (h *Handler) Checkout(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req checkoutReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	cmd := CheckoutCmd{
		Discount: req.Discount, VoucherCode: req.VoucherCode, Method: req.Method, Paid: req.Paid, Customer: req.Customer,
		CustomerID: req.CustomerID, PointsRedeem: req.PointsRedeem,
		ClientTxID: req.ClientTxID,
	}
	for _, it := range req.Items {
		cmd.Items = append(cmd.Items, CheckoutItemCmd{ProductID: it.ProductID, VariantID: it.VariantID, Qty: it.Qty})
	}

	trx, created, err := h.svc.Checkout(c.Request.Context(), claims.StoreID, claims.OutletID, claims.ActingAsCashierID, claims.Name, cmd)
	if err != nil {
		respondTrxErr(c, err)
		return
	}
	status := http.StatusCreated
	if !created {
		status = http.StatusOK
	}
	c.JSON(status, trx)
}

func (h *Handler) PreviewPromo(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req struct {
		Items       []checkoutItemReq `json:"items"`
		VoucherCode string            `json:"voucherCode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	cmd := PromoPreviewCmd{VoucherCode: req.VoucherCode}
	for _, item := range req.Items {
		cmd.Items = append(cmd.Items, CheckoutItemCmd{ProductID: item.ProductID, VariantID: item.VariantID, Qty: item.Qty})
	}
	preview, err := h.svc.PreviewPromo(c.Request.Context(), claims.StoreID, claims.OutletID, cmd)
	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan di toko Anda."})
		return
	}
	if err != nil {
		respondTrxErr(c, err)
		return
	}
	c.JSON(http.StatusOK, preview)
}

func (h *Handler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	qp := c.Query
	outletID, ok := reportOutletFilter(c, claims.Role == middleware.RoleAdmin, claims.OutletID)
	if !ok {
		return
	}

	var cashierID uint
	if claims.ActingAsCashierID != nil {
		cashierID = *claims.ActingAsCashierID
	}
	page, _ := strconv.Atoi(qp("page"))
	limit, _ := strconv.Atoi(qp("limit"))
	page = clampPage(page)
	limit = clampLimit(limit)

	// Filter kasir dari query (admin): "0" sah (penjualan owner), jadi
	// diurai sebagai pointer — bukan queryID yang menganggap 0 = kosong.
	var cashierFilter *uint
	if raw := qp("cashierId"); raw != "" {
		if n, err := strconv.ParseUint(raw, 10, 64); err == nil {
			v := uint(n)
			cashierFilter = &v
		}
	}
	list, total, err := h.svc.List(c.Request.Context(), claims.StoreID, outletID, cashierID,
		qp("q"), qp("method"), qp("date"), qp("from"), qp("to"), qp("status"), cashierFilter, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat transaksi."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": list, "total": total, "page": page, "limit": limit})
}

func (h *Handler) Get(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	trx, err := h.svc.Get(c.Request.Context(), claims.StoreID, id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat transaksi."})
		}
		return
	}
	if claims.ActingAsCashierID != nil && trx.CashierID != *claims.ActingAsCashierID {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan."})
		return
	}
	c.JSON(http.StatusOK, struct {
		*Trx
		Refunds []RefundDetail `json:"refunds"`
	}{Trx: trx, Refunds: trx.Refunds})
}

func (h *Handler) Refund(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req refundReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	items := make([]RefundItem, 0, len(req.Items))
	for _, it := range req.Items {
		items = append(items, RefundItem{ProductID: it.ProductID, VariantID: it.VariantID, Qty: it.Qty})
	}
	trx, err := h.svc.Refund(c.Request.Context(), claims.StoreID, claims.OutletID, id, items, req.Reason, claims.Name)
	if err != nil {
		respondTrxErr(c, err)
		return
	}
	c.JSON(http.StatusOK, trx)
}

func respondTrxErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrShiftRequired):
		c.JSON(http.StatusConflict, gin.H{"error": "Belum ada shift aktif. Buka shift dulu sebelum transaksi.", "code": "SHIFT_REQUIRED"})
	case errors.Is(err, ErrEmptyItems):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Keranjang kosong."})
	case errors.Is(err, ErrStockInsufficient):
		c.JSON(http.StatusConflict, gin.H{"error": "Stok tidak cukup untuk menyelesaikan transaksi."})
	case errors.Is(err, ErrPaidInsufficient):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jumlah bayar kurang dari total."})
	case errors.Is(err, ErrBadDiscount):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Diskon melebihi subtotal."})
	case errors.Is(err, ErrProductInactive):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ada produk yang tidak aktif."})
	case errors.Is(err, ErrVariantRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pilih varian produk terlebih dahulu."})
	case errors.Is(err, ErrVariantInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Varian produk tidak sesuai dengan produk atau toko Anda."})
	case errors.Is(err, ErrNotRefundable):
		c.JSON(http.StatusConflict, gin.H{"error": "Transaksi ini tidak dapat direfund."})
	case errors.Is(err, ErrRefundTooMuch):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Qty refund melebihi jumlah terjual."})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaksi tidak ditemukan."})
	case errors.Is(err, ErrPromoVoucherInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kode voucher tidak valid.", "code": "PROMO_INVALID"})
	case errors.Is(err, ErrCustomerNotFound):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pelanggan tidak ditemukan di toko Anda."})
	case errors.Is(err, ErrPointsCustomerRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pilih pelanggan untuk menukar poin."})
	case errors.Is(err, ErrPointsBalance):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Saldo poin pelanggan tidak mencukupi."})
	case errors.Is(err, ErrPointsExceedTotal):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nilai penukaran poin melebihi total transaksi."})
	case errors.Is(err, ErrPointsRedeemInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jumlah poin yang ditukar tidak valid."})
	case errors.Is(err, ErrLoyaltySettings):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pengaturan loyalty tidak valid."})
	case errors.Is(err, ErrCreditCustomerRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pilih pelanggan untuk transaksi Credit."})
	case errors.Is(err, ErrCreditPaid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Field paid harus 0 untuk transaksi Credit."})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
