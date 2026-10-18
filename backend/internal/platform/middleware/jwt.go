package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
)

const claimsKey = "claims"

// TokenParser adalah kontrak penerbit token yang dipakai Auth:
// diimplementasikan oleh service auth (ParseAccess), didefinisikan di
// sini supaya middleware tidak lagi mengimpor paket service.
type TokenParser interface {
	ParseAccess(token string) (*Claims, error)
}

type APIKeyAuthenticator interface {
	Authenticate(ctx context.Context, rawKey string) (*Claims, error)
}

var ErrInvalidAPIKey = errors.New("kunci API tidak valid")

func Auth(authSvc TokenParser, apiKeys ...APIKeyAuthenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rawKey := c.GetHeader("X-Api-Key"); rawKey != "" {
			if len(apiKeys) == 0 || apiKeys[0] == nil {
				c.AbortWithStatusJSON(401, gin.H{"error": "kunci API tidak valid"})
				return
			}
			claims, err := apiKeys[0].Authenticate(c.Request.Context(), rawKey)
			if errors.Is(err, ErrInvalidAPIKey) || claims == nil && err == nil {
				c.AbortWithStatusJSON(401, gin.H{"error": "kunci API tidak valid"})
				return
			}
			if err != nil {
				c.AbortWithStatusJSON(500, gin.H{"error": "gagal memvalidasi kunci API"})
				return
			}
			c.Set(claimsKey, claims)
			c.Next()
			return
		}
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "token tidak disertakan"})
			return
		}
		claims, err := authSvc.ParseAccess(strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "sesi tidak valid, silakan masuk kembali"})
			return
		}
		c.Set(claimsKey, claims)
		c.Next()
	}
}

func RequireRole(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := ClaimsFrom(c)
		if claims == nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "sesi tidak valid"})
			return
		}
		for _, role := range roles {
			if string(claims.Role) == role {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(403, gin.H{"error": "akses ditolak untuk peran Anda"})
	}
}

func ClaimsFrom(c *gin.Context) *Claims {
	v, exists := c.Get(claimsKey)
	if !exists {
		return nil
	}
	claims, _ := v.(*Claims)
	return claims
}
