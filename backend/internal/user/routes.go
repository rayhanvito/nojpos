package user

import "github.com/gin-gonic/gin"

// RegisterRoutes mendaftarkan semua rute /users di grup admin (path &
// method persis seperti router lama, termasuk PUT passcode yang dulu
// terdaftar dari SettingsHandler).
func RegisterRoutes(admin *gin.RouterGroup, h *Handler) {
	admin.GET("/users", h.List)
	admin.POST("/users", h.Create)
	admin.PATCH("/users/:id", h.Rename)
	admin.PATCH("/users/:id/active", h.SetActive)
	admin.DELETE("/users/:id", h.Delete)
	admin.PUT("/users/:id/passcode", h.SetPasscode)
}
