package report

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan rute laporan (path & method sama persis
// seperti router lama): GET /dashboard di grup auth; laporan, laporan
// pajak, dan ekspor (termasuk GET /transactions/export) di grup admin.
func RegisterRoutes(auth *gin.RouterGroup, admin *gin.RouterGroup, h *Handler, th *TaxHandler, eh *ExportHandler) {
	auth.GET("/dashboard", h.Dashboard)
	admin.GET("/reports", h.Report)
	admin.GET("/reports/tax", th.Get)
	admin.GET("/reports/tax/export", th.Export)
	admin.GET("/transactions/export", eh.Transactions)
	admin.GET("/products/export", eh.Products)
}
