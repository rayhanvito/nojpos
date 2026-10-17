package notification

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan rute notifikasi (path & method sama
// seperti router lama): semua di grup auth, handler-nya sendiri
// membatasi ke admin.
func RegisterRoutes(auth *gin.RouterGroup, h *Handler) {
	auth.GET("/notifications", h.List)
	auth.GET("/notifications/unread-count", h.UnreadCount)
	auth.PATCH("/notifications/read-all", h.MarkAllAsRead)
	auth.PATCH("/notifications/:id/read", h.MarkAsRead)
	auth.DELETE("/notifications/:id", h.Delete)
}
