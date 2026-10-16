package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/middleware"
	"github.com/rayhanvito/nojpos/backend/repo"
	"github.com/rayhanvito/nojpos/backend/service"
)

type ExpenseHandler struct {
	svc *service.ExpenseService
}

func NewExpenseHandler(svc *service.ExpenseService) *ExpenseHandler {
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

func (r expenseReq) input() service.ExpenseInput {
	return service.ExpenseInput{
		Date: r.Date, Category: r.Category, Note: r.Note,
		Amount: r.Amount, Method: r.Method, Source: r.Source,
	}
}

func (h *ExpenseHandler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	qp := c.Query
	page, _ := strconv.Atoi(qp("page"))
	limit, _ := strconv.Atoi(qp("limit"))
	result, err := h.svc.List(c.Request.Context(), claims.StoreID, repo.ExpenseFilter{
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
	case errors.Is(err, service.ErrExpenseNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrExpenseInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data pengeluaran tidak valid: kategori dan nominal wajib diisi, tanggal serta metode harus sah."})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan pada server."})
	}
}
