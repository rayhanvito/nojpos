package apikey

import (
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

func (h *Handler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	items, err := h.svc.List(c.Request.Context(), claims.StoreID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, httpx.PaginateList(c, items))
}

func (h *Handler) Create(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body kunci API tidak valid."})
		return
	}
	claims := middleware.ClaimsFrom(c)
	metadata, fullKey, err := h.svc.Create(c.Request.Context(), claims.StoreID, claims.UserID, req.Name)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, struct {
		Metadata
		APIKey string `json:"apiKey"`
	}{Metadata: metadata, APIKey: fullKey})
}

func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID kunci API tidak valid."})
		return
	}
	claims := middleware.ClaimsFrom(c)
	if err := h.svc.Delete(c.Request.Context(), claims.StoreID, uint(id)); err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Kunci API berhasil dicabut."})
}
