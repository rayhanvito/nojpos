package finance

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan rute keuangan (path & method sama
// seperti router lama): pengeluaran List/Create di grup auth,
// Update/Delete pengeluaran dan seluruh utang-piutang di grup admin.
func RegisterRoutes(auth *gin.RouterGroup, admin *gin.RouterGroup, h *Handler, eh *ExpenseHandler) {
	auth.GET("/expenses", eh.List)
	auth.POST("/expenses", eh.Create)
	admin.PUT("/expenses/:id", eh.Update)
	admin.DELETE("/expenses/:id", eh.Delete)
	admin.GET("/receivables", h.ListReceivables)
	admin.GET("/receivables/:id", h.GetReceivable)
	admin.POST("/receivables/:id/payments", h.PayReceivable)
	admin.GET("/payables", h.ListPayables)
	admin.GET("/payables/:id", h.GetPayable)
	admin.POST("/payables/:id/payments", h.PayPayable)
}
