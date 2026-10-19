package finance

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// Handler adalah DebtHandler lama di handler/debt.go, dipindah
// verbatim.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// pathUint adalah salinan privat slice dari helper paket handler lama
// (handler/params.go) — badan verbatim.
func pathUint(c *gin.Context, name string) (uint, bool) {
	n, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid."})
		return 0, false
	}
	return uint(n), true
}

type debtPaymentReq struct {
	Amount int64  `json:"amount"`
	Method string `json:"method"`
}

func (h *Handler) ListReceivables(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	overdue, ok := parseOverdue(c.Query("overdue"))
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filter overdue harus bernilai true atau false."})
		return
	}
	items, err := h.svc.ListReceivables(c.Request.Context(), claims.StoreID, c.Query("status"), overdue)
	if err != nil {
		respondDebtErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) GetReceivable(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	item, err := h.svc.GetReceivable(c.Request.Context(), claims.StoreID, id)
	if err != nil {
		respondDebtErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) PayReceivable(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req debtPaymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body pembayaran tidak valid."})
		return
	}
	item, err := h.svc.PayReceivable(c.Request.Context(), claims.StoreID, id, req.Amount, req.Method, claims.Name)
	if err != nil {
		respondDebtErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) ListPayables(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	items, err := h.svc.ListPayables(c.Request.Context(), claims.StoreID, c.Query("status"))
	if err != nil {
		respondDebtErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) GetPayable(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	item, err := h.svc.GetPayable(c.Request.Context(), claims.StoreID, id)
	if err != nil {
		respondDebtErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func (h *Handler) PayPayable(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req debtPaymentReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body pembayaran tidak valid."})
		return
	}
	item, err := h.svc.PayPayable(c.Request.Context(), claims.StoreID, id, req.Amount, req.Method, claims.Name)
	if err != nil {
		respondDebtErr(c, err)
		return
	}
	c.JSON(http.StatusOK, item)
}

func parseOverdue(value string) (bool, bool) {
	if value == "" || value == "false" {
		return false, true
	}
	if value == "true" {
		return true, true
	}
	return false, false
}

func respondDebtErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrDebtStatusInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filter status harus unpaid, partially_paid, atau paid."})
	case errors.Is(err, ErrDebtPaymentInput), errors.Is(err, ErrDebtPaymentInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jumlah pembayaran harus lebih dari 0 dan metode harus salah satu dari Cash, Bank Transfer, QRIS, E-Wallet, atau Card."})
	case errors.Is(err, ErrDebtPaymentExceedsRemain):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jumlah pembayaran melebihi sisa hutang-piutang."})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Data hutang-piutang tidak ditemukan."})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses hutang-piutang."})
	}
}

// ExpenseHandler adalah ExpenseHandler lama di handler/expense.go,
// dipindah verbatim.
type ExpenseHandler struct {
	svc *ExpenseService
}

func NewExpenseHandler(svc *ExpenseService) *ExpenseHandler {
	return &ExpenseHandler{svc: svc}
}

type expenseReq struct {
	Date     string `json:"date"`
	Category string `json:"category"`
	Note     string `json:"note"`
	Amount   int64  `json:"amount"`
	Method   string `json:"method"`
	Source   string `json:"source"`
}

func (r expenseReq) input() ExpenseInput {
	return ExpenseInput{
		Date: r.Date, Category: r.Category, Note: r.Note,
		Amount: r.Amount, Method: r.Method, Source: r.Source,
	}
}

func (h *ExpenseHandler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	qp := c.Query
	page, _ := strconv.Atoi(qp("page"))
	limit, _ := strconv.Atoi(qp("limit"))
	result, err := h.svc.List(c.Request.Context(), claims.StoreID, ExpenseFilter{
		From: qp("from"), To: qp("to"), Category: qp("category"), Q: qp("q"),
		Page: page, Limit: limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat pengeluaran."})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *ExpenseHandler) Create(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req expenseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body pengeluaran tidak valid."})
		return
	}
	expense, err := h.svc.Create(c.Request.Context(), claims.StoreID, claims.OutletID, claims.ActingAsCashierID, claims.Name, req.input())
	if err != nil {
		respondExpenseErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, expense)
}

func (h *ExpenseHandler) Update(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req expenseReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body pengeluaran tidak valid."})
		return
	}
	expense, err := h.svc.Update(c.Request.Context(), claims.StoreID, id, req.input())
	if err != nil {
		respondExpenseErr(c, err)
		return
	}
	c.JSON(http.StatusOK, expense)
}

func (h *ExpenseHandler) Delete(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), claims.StoreID, id); err != nil {
		respondExpenseErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func respondExpenseErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrExpenseNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, ErrExpenseInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data pengeluaran tidak valid: kategori dan nominal wajib diisi, tanggal serta metode harus sah."})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan pada server."})
	}
}
