package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/adrr-dev/openPOS/backend/middleware"
	"github.com/adrr-dev/openPOS/backend/model"
	"github.com/adrr-dev/openPOS/backend/repo"
	"github.com/adrr-dev/openPOS/backend/service"
)

type OutletHandler struct {
	outlets *service.OutletService
	auth    *service.AuthService
}

func NewOutletHandler(outlets *service.OutletService, auth *service.AuthService) *OutletHandler {
	return &OutletHandler{outlets: outlets, auth: auth}
}

func (h *OutletHandler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	items, err := h.outlets.List(c.Request.Context(), claims.StoreID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat daftar outlet."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

type outletReq struct {
	Name      string `json:"name"`
	Address   string `json:"address"`
	IsDefault *bool  `json:"isDefault"`
}

func (h *OutletHandler) Create(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req outletReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	outlet, err := h.outlets.Create(c.Request.Context(), claims.StoreID, model.Outlet{
		Name: req.Name, Address: req.Address, IsDefault: req.IsDefault != nil && *req.IsDefault,
	})
	if err != nil {
		respondOutletErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, outlet)
}

func (h *OutletHandler) Update(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req outletReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	outlet, err := h.outlets.Update(c.Request.Context(), claims.StoreID, id, model.Outlet{
		Name: req.Name, Address: req.Address,
	}, req.IsDefault)
	if err != nil {
		respondOutletErr(c, err)
		return
	}
	c.JSON(http.StatusOK, outlet)
}

func (h *OutletHandler) Switch(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req struct {
		OutletID uint `json:"outletId"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || req.OutletID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Outlet tujuan tidak valid."})
		return
	}
	user, pair, err := h.auth.SwitchOutlet(c.Request.Context(), claims, req.OutletID)
	if err != nil {
		respondOutletErr(c, err)
		return
	}
	c.JSON(http.StatusOK, authResponse{User: *user, TokenPair: *pair})
}

func respondOutletErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrOutletForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "Kasir hanya dapat menggunakan outlet yang ditetapkan."})
	case errors.Is(err, service.ErrOutletNotFound), errors.Is(err, repo.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Outlet tidak ditemukan di toko Anda."})
	case errors.Is(err, repo.ErrOutletDefault):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Outlet default tidak dapat dinonaktifkan."})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
