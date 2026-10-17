package activity

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// Handler adalah ActivityHandler lama di handler/activity.go, dipindah
// verbatim.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	if claims == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	page, _ := strconv.Atoi(c.Query("page"))
	limit, _ := strconv.Atoi(c.Query("limit"))

	f := ActivityFilter{
		Actor:  c.Query("actor"),
		Action: c.Query("action"),
		Date:   c.Query("date"),
		Page:   page,
		Limit:  limit,
	}

	logs, total, err := h.svc.List(c.Request.Context(), claims.StoreID, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat aktivitas."})
		return
	}

	if logs == nil {
		logs = []ActivityLog{}
	}

	p := page
	if p < 1 {
		p = 1
	}
	l := limit
	if l < 1 {
		l = 20
	}

	c.JSON(http.StatusOK, gin.H{
		"items": logs,
		"total": total,
		"page":  p,
		"limit": l,
	})
}
