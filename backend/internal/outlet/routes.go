package outlet

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan rute outlet (path & method sama seperti
// router lama): List di grup auth, Create/Update di grup admin, plus
// rute pengaturan toko sejak BE-F1d (GET /settings di grup auth,
// PUT /settings di grup admin, dari SettingsHandler) dan pengaturan
// per outlet sejak Unit C (GET /outlets/:id/settings di grup auth,
// PUT /outlets/:id/settings di grup admin). Rute
// POST /outlets/switch TIDAK terdaftar di sini — Switch milik
// sesi/auth dan tinggal di internal/auth (sejak BE-F1c).
func RegisterRoutes(auth *gin.RouterGroup, admin *gin.RouterGroup, h *Handler, settingsH *SettingsHandler) {
	auth.GET("/outlets", h.List)
	admin.POST("/outlets", h.Create)
	admin.PUT("/outlets/:id", h.Update)
	auth.GET("/settings", settingsH.Get)
	admin.PUT("/settings", settingsH.Update)
	auth.GET("/outlets/:id/settings", settingsH.GetOutletSettings)
	admin.PUT("/outlets/:id/settings", settingsH.UpdateOutletSettings)
}
