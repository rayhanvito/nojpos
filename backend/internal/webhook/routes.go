package webhook

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan rute manajemen webhook (grup admin saja;
// kunci API juga dilarang mengelola webhook lewat RestrictAPIKeyRoutes).
func RegisterRoutes(admin *gin.RouterGroup, h *Handler) {
	admin.GET("/webhooks", h.List)
	admin.POST("/webhooks", h.Create)
	admin.DELETE("/webhooks/:id", h.Delete)
}
