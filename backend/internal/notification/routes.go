package notification

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan rute notifikasi di grup admin.
func RegisterRoutes(admin *gin.RouterGroup, h *Handler) {
	admin.GET("/notifications", h.List)
	admin.GET("/notifications/unread-count", h.UnreadCount)
	admin.PATCH("/notifications/read-all", h.MarkAllAsRead)
	admin.PATCH("/notifications/:id/read", h.MarkAsRead)
	admin.DELETE("/notifications/:id", h.Delete)
}
