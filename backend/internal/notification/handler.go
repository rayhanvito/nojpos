package notification

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// Handler adalah NotificationHandler lama di handler/notification.go,
// dipindah verbatim.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// pathUint adalah salinan privat slice dari helper paket handler lama.
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
	q := c.Query

	category := q("category")
	status := q("status")
	unreadOnly := q("unread") == "true"
	page, _ := strconv.Atoi(q("page"))
	limit, _ := strconv.Atoi(q("limit"))

	pageData, err := h.svc.List(c.Request.Context(), claims.StoreID, category, status, unreadOnly, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat notifikasi."})
		return
	}
	c.JSON(http.StatusOK, pageData)
}

func (h *Handler) UnreadCount(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	total, err := h.svc.UnreadCount(c.Request.Context(), claims.StoreID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menghitung notifikasi."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total})
}

func (h *Handler) MarkAsRead(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}

	if err := h.svc.MarkAsRead(c.Request.Context(), claims.StoreID, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notifikasi tidak ditemukan."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) MarkAllAsRead(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	if err := h.svc.MarkAllAsRead(c.Request.Context(), claims.StoreID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memperbarui notifikasi."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *Handler) Delete(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}

	if err := h.svc.Delete(c.Request.Context(), claims.StoreID, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notifikasi tidak ditemukan."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}
