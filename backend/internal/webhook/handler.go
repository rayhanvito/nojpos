package webhook

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

// List mengembalikan metadata langganan toko — secret tidak pernah
// ikut (Metadata memang tidak memuatnya).
func (h *Handler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	items, err := h.svc.List(c.Request.Context(), claims.StoreID)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, httpx.PaginateList(c, items))
}

// Create mendaftarkan webhook baru. Respons 201 adalah SATU-SATUNYA
// respons yang memuat secret langganan.
func (h *Handler) Create(c *gin.Context) {
	var req struct {
		URL    string   `json:"url"`
		Events []string `json:"events"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body webhook tidak valid."})
		return
	}
	claims := middleware.ClaimsFrom(c)
	metadata, secret, err := h.svc.Create(c.Request.Context(), claims.StoreID, req.URL, req.Events)
	if err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, struct {
		Metadata
		Secret string `json:"secret"`
	}{Metadata: metadata, Secret: secret})
}

// Delete mencabut langganan (soft delete). Webhook milik toko lain
// dijawab 404, bukan 403, agar keberadaannya tidak bocor.
func (h *Handler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID webhook tidak valid."})
		return
	}
	claims := middleware.ClaimsFrom(c)
	if err := h.svc.Delete(c.Request.Context(), claims.StoreID, uint(id)); err != nil {
		httpx.Fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Webhook berhasil dihapus."})
}
