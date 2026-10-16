package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/middleware"
	"github.com/rayhanvito/nojpos/backend/model"
	"github.com/rayhanvito/nojpos/backend/repo"
	"github.com/rayhanvito/nojpos/backend/service"
)

type StockHandler struct {
	svc *service.CatalogService
}

func NewStockHandler(svc *service.CatalogService) *StockHandler { return &StockHandler{svc: svc} }

func (h *StockHandler) ListMovements(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	q := c.Query
	outletID := claims.OutletID
	if claims.Role == model.RoleAdmin {
		if value, provided := c.GetQuery("outletId"); provided {
			id, err := strconv.ParseUint(value, 10, 64)
			if err != nil || id == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "outletId tidak valid."})
				return
			}
			outletID = uint(id)
		}
	}

	f := repo.MovementFilter{Type: q("type"), ProductID: queryID(q, "productId"), Q: q("q")}
	if v := q("page"); v != "" {
		f.Page, _ = strconv.Atoi(v)
	}
	if v := q("limit"); v != "" {
		f.Limit, _ = strconv.Atoi(v)
	}

	pageData, err := h.svc.ListMovements(c.Request.Context(), claims.StoreID, outletID, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat riwayat pergerakan."})
		return
	}
	c.JSON(http.StatusOK, pageData)
}

type adjustReq struct {
	ProductID uint   `json:"productId"`
	VariantID *uint  `json:"variantId"`
	Direction string `json:"direction"`
	Qty       int64  `json:"qty"`
	Reason    string `json:"reason"`
}

func (h *StockHandler) AdjustStock(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req adjustReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	p, err := h.svc.AdjustStock(c.Request.Context(), claims.StoreID, claims.OutletID, req.ProductID, req.VariantID, claims.Name, req.Direction, req.Qty, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrBadDirection):
			c.JSON(http.StatusBadRequest, gin.H{"error": "arah penyesuaian harus 'plus' atau 'minus'"})
		case errors.Is(err, service.ErrStoreMismatch):
			c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan."})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"product": p})
}
