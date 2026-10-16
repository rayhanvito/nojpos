package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/adrr-dev/openPOS/backend/middleware"
	"github.com/adrr-dev/openPOS/backend/repo"
	"github.com/adrr-dev/openPOS/backend/service"
)

type SupplierHandler struct {
	svc *service.SupplierService
}

func NewSupplierHandler(svc *service.SupplierService) *SupplierHandler {
	return &SupplierHandler{svc: svc}
}

type supplierReq struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Address string `json:"address"`
}

func (r supplierReq) input() service.SupplierInput {
	return service.SupplierInput{Name: r.Name, Contact: r.Contact, Address: r.Address}
}

func (h *SupplierHandler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	items, err := h.svc.List(c.Request.Context(), claims.StoreID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat supplier."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *SupplierHandler) Get(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	supplier, err := h.svc.Get(c.Request.Context(), claims.StoreID, id)
	if err != nil {
		respondSupplierErr(c, err)
		return
	}
	c.JSON(http.StatusOK, supplier)
}

func (h *SupplierHandler) Create(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req supplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body supplier tidak valid."})
		return
	}
	supplier, err := h.svc.Create(c.Request.Context(), claims.StoreID, req.input())
	if err != nil {
		respondSupplierErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, supplier)
}

func (h *SupplierHandler) Update(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req supplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body supplier tidak valid."})
		return
	}
	supplier, err := h.svc.Update(c.Request.Context(), claims.StoreID, id, req.input())
	if err != nil {
		respondSupplierErr(c, err)
		return
	}
	c.JSON(http.StatusOK, supplier)
}

func (h *SupplierHandler) Delete(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), claims.StoreID, id); err != nil {
		respondSupplierErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Supplier berhasil dihapus."})
}

func respondSupplierErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrSupplierNameRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama supplier wajib diisi."})
	case errors.Is(err, service.ErrSupplierNotFound), errors.Is(err, repo.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Supplier tidak ditemukan di toko Anda."})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses supplier."})
	}
}
