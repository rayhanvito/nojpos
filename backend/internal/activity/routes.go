package activity

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan rute aktivitas (path & method sama
// seperti router lama): daftar aktivitas di grup admin.
func RegisterRoutes(admin *gin.RouterGroup, h *Handler) {
	admin.GET("/activity", h.List)
}
