package customer

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// Handler adalah CustomerHandler lama di handler/customer.go, dipindah
// verbatim.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// pathUint parses a numeric path param, writing 400 on garbage
// (semantik disalin persis dari paket handler lama).
func pathUint(c *gin.Context, name string) (uint, bool) {
	n, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid."})
		return 0, false
	}
	return uint(n), true
}

func (h *Handler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))
	data, err := h.svc.List(c.Request.Context(), claims.StoreID, CustomerFilter{
		Q: c.Query("q"), Page: page, Limit: limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat pelanggan."})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) Get(c *gin.Context) {
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

func (h *Handler) Transactions(c *gin.Context) {
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

func (h *Handler) Create(c *gin.Context) {
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

func (h *Handler) Update(c *gin.Context) {
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

func (h *Handler) Delete(c *gin.Context) {
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
	case errors.Is(err, ErrCustomerNotFound), errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Pelanggan tidak ditemukan di toko Anda."})
	case errors.Is(err, ErrCustomerPhoneTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "Nomor telepon sudah digunakan di toko ini."})
	case errors.Is(err, ErrCustomerNameRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama pelanggan wajib diisi."})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses pelanggan."})
	}
}
