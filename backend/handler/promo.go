package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/middleware"
	"github.com/rayhanvito/nojpos/backend/repo"
	"github.com/rayhanvito/nojpos/backend/service"
)

type PromoHandler struct {
	svc *service.PromoService
}

func NewPromoHandler(svc *service.PromoService) *PromoHandler { return &PromoHandler{svc: svc} }

type promoReq struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Value       json.Number `json:"value"`
	Scope       string      `json:"scope"`
	CategoryID  *uint       `json:"categoryId"`
	ProductID   *uint       `json:"productId"`
	VoucherCode *string     `json:"voucherCode"`
	StartsAt    time.Time   `json:"startsAt"`
	EndsAt      time.Time   `json:"endsAt"`
	HappyStart  *string     `json:"happyStart"`
	HappyEnd    *string     `json:"happyEnd"`
	MinPurchase int64       `json:"minPurchase"`
	MaxDiscount int64       `json:"maxDiscount"`
	Combinable  bool        `json:"combinable"`
}

func promoInput(req promoReq) (service.PromoInput, error) {
	in := service.PromoInput{
		Name: req.Name, Type: req.Type, Scope: req.Scope,
		CategoryID: req.CategoryID, ProductID: req.ProductID, VoucherCode: req.VoucherCode,
		StartsAt: req.StartsAt, EndsAt: req.EndsAt, HappyStart: req.HappyStart, HappyEnd: req.HappyEnd,
		MinPurchase: req.MinPurchase, MaxDiscount: req.MaxDiscount, Combinable: req.Combinable,
	}
	switch req.Type {
	case "percent":
		value, err := strconv.ParseFloat(req.Value.String(), 64)
		if err != nil {
			return in, err
		}
		in.Value = value
	case "amount":
		value, err := strconv.ParseInt(req.Value.String(), 10, 64)
		if err != nil {
			return in, err
		}
		in.AmountValue = value
	}
	return in, nil
}

func (h *PromoHandler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	items, err := h.svc.List(c.Request.Context(), claims.StoreID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat promo."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *PromoHandler) Get(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	promo, err := h.svc.Get(c.Request.Context(), claims.StoreID, id)
	if err != nil {
		respondPromoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, promo)
}

func (h *PromoHandler) Create(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req promoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body promo tidak valid."})
		return
	}
	in, err := promoInput(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nilai promo tidak valid; nominal harus berupa bilangan Rupiah int64."})
		return
	}
	promo, err := h.svc.Create(c.Request.Context(), claims.StoreID, in)
	if err != nil {
		respondPromoErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, promo)
}

func (h *PromoHandler) Update(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req promoReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body promo tidak valid."})
		return
	}
	in, err := promoInput(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nilai promo tidak valid; nominal harus berupa bilangan Rupiah int64."})
		return
	}
	promo, err := h.svc.Update(c.Request.Context(), claims.StoreID, id, in)
	if err != nil {
		respondPromoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, promo)
}

func (h *PromoHandler) SetActive(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
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
	if err := h.svc.SetActive(c.Request.Context(), claims.StoreID, id, *req.Active); err != nil {
		respondPromoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status promo diperbarui."})
}

func (h *PromoHandler) Delete(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), claims.StoreID, id); err != nil {
		respondPromoErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Promo berhasil dihapus."})
}

func respondPromoErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrPromoCodeTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "Kode voucher sudah digunakan di toko ini."})
	case errors.Is(err, repo.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Promo tidak ditemukan."})
	case errors.Is(err, service.ErrPromoInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses promo."})
	}
}
