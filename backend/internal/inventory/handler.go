package inventory

import (
	"errors"
	"github.com/rayhanvito/nojpos/backend/internal/user"
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

func (h *Handler) ListMovements(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	q := c.Query
	outletID := claims.OutletID
	if claims.Role == middleware.RoleAdmin {
		if value, provided := c.GetQuery("outletId"); provided {
			id, err := strconv.ParseUint(value, 10, 64)
			if err != nil || id == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "outletId tidak valid."})
				return
			}
			outletID = uint(id)
		}
	}

	f := MovementListFilter{Type: q("type"), ProductID: queryID(q, "productId"), Q: q("q")}
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

func (h *Handler) AdjustStock(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req adjustReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	p, err := h.svc.AdjustStock(c.Request.Context(), claims.StoreID, claims.OutletID, req.ProductID, req.VariantID, claims.Name, req.Direction, req.Qty, req.Reason)
	if err != nil {
		switch {
		case errors.Is(err, ErrBadDirection):
			httpx.Fail(c, err)
		case errors.Is(err, user.ErrStoreMismatch):
			c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan."})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"product": p})
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

// queryID parses an optional numeric query filter.
// Returns 0 when absent. When present but invalid, returns max uint so the
// filter matches nothing (same visible result as before with bad input).
func queryID(q func(string) string, name string) uint {
	v := q(name)
	if v == "" {
		return 0
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return ^uint(0)
	}
	return uint(n)
}
