package inventory

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan semua rute stok/movement/opname persis
// seperti router lama: stock-opname di grup auth, movements,
// stock/adjustments, dan stock/transfers (Unit D) di grup admin.
func RegisterRoutes(auth, admin *gin.RouterGroup, h *Handler, oh *OpnameHandler) {
	auth.GET("/stock-opnames", oh.List)
	auth.POST("/stock-opnames", oh.Create)
	auth.GET("/stock-opnames/:id", oh.Get)
	auth.PUT("/stock-opnames/:id/items", oh.UpdateItems)
	auth.POST("/stock-opnames/:id/post", oh.Post)
	auth.POST("/stock-opnames/:id/cancel", oh.Cancel)

	admin.GET("/movements", h.ListMovements)
	admin.POST("/stock/adjustments", h.AdjustStock)
	admin.GET("/stock/transfers", h.ListTransfers)
	admin.POST("/stock/transfers", h.CreateTransfer)
}
