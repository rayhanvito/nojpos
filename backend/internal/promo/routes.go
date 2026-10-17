package promo

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan 6 rute admin promo (path & method sama
// seperti router lama).
func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	rg.GET("/promos", h.List)
	rg.POST("/promos", h.Create)
	rg.GET("/promos/:id", h.Get)
	rg.PUT("/promos/:id", h.Update)
	rg.PATCH("/promos/:id/active", h.SetActive)
	rg.DELETE("/promos/:id", h.Delete)
}
