package apikey

import "github.com/gin-gonic/gin"

func RegisterRoutes(admin *gin.RouterGroup, h *Handler) {
	admin.GET("/api-keys", h.List)
	admin.POST("/api-keys", h.Create)
	admin.DELETE("/api-keys/:id", h.Delete)
}
