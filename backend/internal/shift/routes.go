package shift

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan rute shift (path & method sama seperti
// router lama): shift kasir di grup auth, daftar/detail/ekspor/
// force-close shift di grup admin.
func RegisterRoutes(auth *gin.RouterGroup, admin *gin.RouterGroup, h *Handler) {
	auth.GET("/cashier/shift", h.GetCurrentShift)
	auth.POST("/cashier/shift/start", h.StartShift)
	auth.POST("/cashier/shift/close", h.CloseShift)
	admin.GET("/shifts/export", h.ExportShifts)
	admin.GET("/shifts", h.ListShifts)
	admin.GET("/shifts/:id", h.GetShift)
	admin.POST("/shifts/:id/force-close", h.ForceCloseShift)
}
