package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SessionValidator adalah kontrak pemeriksa sesi aktif yang dipakai
// ActiveCheck: diimplementasikan oleh service auth (ValidateSession,
// yang membungkus Me), didefinisikan di sini supaya middleware tidak
// lagi mengimpor paket service.
type SessionValidator interface {
	ValidateSession(ctx context.Context, claims *Claims) error
}

func ActiveCheck(authSvc SessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := ClaimsFrom(c)
		if claims == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "sesi tidak valid"})
			return
		}
		// Reuse Me logic to validate active status (checks user/cashier Active)
		if err := authSvc.ValidateSession(c.Request.Context(), claims); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "akun dinonaktifkan"})
			return
		}
		c.Next()
	}
}
