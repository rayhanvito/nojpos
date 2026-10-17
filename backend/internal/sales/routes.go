package sales

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan rute transaksi (path & method sama
// persis seperti router lama): preview promo, checkout, daftar, dan
// detail transaksi di grup auth; refund di grup admin.
func RegisterRoutes(auth *gin.RouterGroup, admin *gin.RouterGroup, h *Handler) {
	auth.POST("/promos/preview", h.PreviewPromo)
	auth.POST("/transactions", h.Checkout)
	auth.GET("/transactions", h.List)
	auth.GET("/transactions/:id", h.Get)
	admin.POST("/transactions/:id/refund", h.Refund)
}
