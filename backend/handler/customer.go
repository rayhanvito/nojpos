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

type CustomerHandler struct {
	svc *service.CustomerService
}

func NewCustomerHandler(svc *service.CustomerService) *CustomerHandler {
	return &CustomerHandler{svc: svc}
}

type customerReq struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Email   string `json:"email"`
	Address string `json:"address"`
}

func (r customerReq) input() service.CustomerInput {
	return service.CustomerInput{Name: r.Name, Phone: r.Phone, Email: r.Email, Address: r.Address}
}

func (h *CustomerHandler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	data, err := h.svc.List(c.Request.Context(), claims.StoreID, repo.CustomerFilter{
		Q: c.Query("q"), Page: page, Limit: limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat pelanggan."})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *CustomerHandler) Get(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	customer, err := h.svc.Get(c.Request.Context(), claims.StoreID, id)
	if err != nil {
		respondCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, customer)
}

func (h *CustomerHandler) Transactions(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	items, err := h.svc.Transactions(c.Request.Context(), claims.StoreID, id)
	if err != nil {
		respondCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *CustomerHandler) Create(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req customerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body JSON tidak valid."})
		return
	}
	customer, err := h.svc.Create(c.Request.Context(), claims.StoreID, req.input())
	if err != nil {
		respondCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, customer)
}

func (h *CustomerHandler) Update(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req customerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body JSON tidak valid."})
		return
	}
	customer, err := h.svc.Update(c.Request.Context(), claims.StoreID, id, req.input())
	if err != nil {
		respondCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, customer)
}

func (h *CustomerHandler) Delete(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), claims.StoreID, id); err != nil {
		respondCustomerErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Pelanggan berhasil dihapus."})
}

func respondCustomerErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCustomerNotFound), errors.Is(err, repo.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Pelanggan tidak ditemukan di toko Anda."})
	case errors.Is(err, service.ErrCustomerPhoneTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "Nomor telepon sudah digunakan di toko ini."})
	case errors.Is(err, service.ErrCustomerNameRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama pelanggan wajib diisi."})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses pelanggan."})
	}
}
