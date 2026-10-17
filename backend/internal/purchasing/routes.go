package purchasing

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan rute pembelian (path & method sama
// seperti router lama): supplier List/Get dan purchase-order List/Get
// di grup auth, seluruh mutasi supplier + purchase-order di grup
// admin.
func RegisterRoutes(auth *gin.RouterGroup, admin *gin.RouterGroup, h *Handler, sh *SupplierHandler) {
	auth.GET("/suppliers", sh.List)
	auth.GET("/suppliers/:id", sh.Get)
	auth.GET("/purchase-orders", h.List)
	auth.GET("/purchase-orders/:id", h.Get)
	admin.POST("/suppliers", sh.Create)
	admin.PUT("/suppliers/:id", sh.Update)
	admin.DELETE("/suppliers/:id", sh.Delete)
	admin.POST("/purchase-orders", h.Create)
	admin.PUT("/purchase-orders/:id", h.Update)
	admin.DELETE("/purchase-orders/:id", h.Delete)
	admin.POST("/purchase-orders/:id/order", h.Order)
	admin.POST("/purchase-orders/:id/receive", h.Receive)
	admin.POST("/purchase-orders/:id/cancel", h.Cancel)
}
