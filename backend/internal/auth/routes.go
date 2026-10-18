package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// RegisterPublicRoutes mendaftarkan 9 rute publik auth. Rate-limit:
// login/forgot/switch dari router lama, plus OTP dan register/google
// (Unit A — lindungi kuota email dan cegah pembuatan akun massal).
func RegisterPublicRoutes(v1 *gin.RouterGroup, h *Handler) {
	v1.POST("/auth/register", middleware.RateLimitRegister(), h.Register)
	v1.POST("/auth/login", middleware.RateLimitLogin(), h.Login)
	v1.POST("/auth/google", middleware.RateLimitRegister(), h.Google)
	v1.POST("/auth/refresh", h.Refresh)
	v1.POST("/auth/logout", h.Logout)
	v1.POST("/auth/otp/send", middleware.RateLimitOTP(), h.SendOTP)
	v1.POST("/auth/otp/verify", h.VerifyOTP)
	v1.POST("/auth/forgot-password/send", middleware.RateLimitForgot(), h.SendPasswordResetOTP)
	v1.POST("/auth/forgot-password/reset", h.ResetPassword)
}

// RegisterAuthRoutes mendaftarkan rute auth di grup terautentikasi
// (presence/heartbeat, auth/me, outlets/switch, auth/password,
// auth/switch +RateLimitSwitch).
func RegisterAuthRoutes(g *gin.RouterGroup, h *Handler) {
	g.POST("/presence/heartbeat", h.Heartbeat)
	g.GET("/auth/me", h.Me)
	g.POST("/outlets/switch", h.SwitchOutlet)
	g.PUT("/auth/password", h.ChangePassword)
	g.POST("/auth/switch", middleware.RateLimitSwitch(), h.Switch)
}
