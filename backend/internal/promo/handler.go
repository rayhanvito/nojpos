package promo

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/httpx"
	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

// storeIDFromClaims membaca StoreID dari claims yang disimpan
// middleware Auth pada konteks gin (middleware.ClaimsFrom bertipe).
// Klaim nil (tak terautentikasi) menghasilkan 0, sama seperti jalur
// refleksi sebelumnya.
func storeIDFromClaims(c *gin.Context) uint {
	claims := middleware.ClaimsFrom(c)
	if claims == nil {
		return 0
	}
	return claims.StoreID
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
	items, err := h.svc.List(c.Request.Context(), storeIDFromClaims(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat promo."})
		return
	}
	c.JSON(http.StatusOK, httpx.PaginateList(c, items))
}

func (h *Handler) Get(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	promo, err := h.svc.Get(c.Request.Context(), storeIDFromClaims(c), id)
	if err != nil {
		respondPromoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, promo)
}

func (h *Handler) Create(c *gin.Context) {
	var req promoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body promo tidak valid."})
		return
	}
	in, err := promoInput(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nilai promo tidak valid; nominal harus berupa bilangan Rupiah int64."})
		return
	}
	promo, err := h.svc.Create(c.Request.Context(), storeIDFromClaims(c), in)
	if err != nil {
		respondPromoErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, promo)
}

func (h *Handler) Update(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req promoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body promo tidak valid."})
		return
	}
	in, err := promoInput(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nilai promo tidak valid; nominal harus berupa bilangan Rupiah int64."})
		return
	}
	promo, err := h.svc.Update(c.Request.Context(), storeIDFromClaims(c), id, in)
	if err != nil {
		respondPromoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, promo)
}

func (h *Handler) SetActive(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req struct {
		Active *bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.Active == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status aktif wajib diisi."})
		return
	}
	if err := h.svc.SetActive(c.Request.Context(), storeIDFromClaims(c), id, *req.Active); err != nil {
		respondPromoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status promo diperbarui."})
}

func (h *Handler) Delete(c *gin.Context) {
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), storeIDFromClaims(c), id); err != nil {
		respondPromoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Promo berhasil dihapus."})
}

func respondPromoErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrCodeTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "Kode voucher sudah digunakan di toko ini.", "code": "PROMO_CODE_TAKEN"})
	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Promo tidak ditemukan.", "code": "NOT_FOUND"})
	case errors.Is(err, ErrInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "PROMO_INVALID"})
	default:
		httpx.Fail(c, err)
	}
}
