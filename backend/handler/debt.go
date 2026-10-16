package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/adrr-dev/openPOS/backend/middleware"
	"github.com/adrr-dev/openPOS/backend/repo"
	"github.com/adrr-dev/openPOS/backend/service"
)

type DebtHandler struct {
	svc *service.DebtService
}

func NewDebtHandler(svc *service.DebtService) *DebtHandler { return &DebtHandler{svc: svc} }

type debtPaymentReq struct {
	Amount int64  `json:"amount"`
	Method string `json:"method"`
}

func (h *DebtHandler) ListReceivables(c *gin.Context) {
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

func (h *DebtHandler) GetReceivable(c *gin.Context) {
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

func (h *DebtHandler) PayReceivable(c *gin.Context) {
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

func (h *DebtHandler) ListPayables(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	items, err := h.svc.ListPayables(c.Request.Context(), claims.StoreID, c.Query("status"))
	if err != nil {
		respondDebtErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *DebtHandler) GetPayable(c *gin.Context) {
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

func (h *DebtHandler) PayPayable(c *gin.Context) {
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
	case errors.Is(err, service.ErrDebtStatusInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Filter status harus unpaid, partially_paid, atau paid."})
	case errors.Is(err, service.ErrDebtPaymentInput), errors.Is(err, repo.ErrDebtPaymentInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jumlah pembayaran harus lebih dari 0 dan metode harus salah satu dari Cash, Bank Transfer, QRIS, E-Wallet, atau Card."})
	case errors.Is(err, repo.ErrDebtPaymentExceedsRemain):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Jumlah pembayaran melebihi sisa hutang-piutang."})
	case errors.Is(err, repo.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Data hutang-piutang tidak ditemukan."})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses hutang-piutang."})
	}
}

func parseDebtID(c *gin.Context) (uint, bool) {
	value, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || value == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID hutang-piutang tidak valid."})
		return 0, false
	}
	return uint(value), true
}
