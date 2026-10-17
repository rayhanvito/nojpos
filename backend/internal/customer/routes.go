package customer

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan rute pelanggan (path & method sama
// seperti router lama): List/Get/Transactions/Create di grup auth,
// Update/Delete di grup admin.
func RegisterRoutes(auth *gin.RouterGroup, admin *gin.RouterGroup, h *Handler) {
	auth.GET("/customers", h.List)
	auth.GET("/customers/:id/transactions", h.Transactions)
	auth.GET("/customers/:id", h.Get)
	auth.POST("/customers", h.Create)
	admin.PUT("/customers/:id", h.Update)
	admin.DELETE("/customers/:id", h.Delete)
}
