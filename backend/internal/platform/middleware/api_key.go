package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func RestrictAPIKeyRoutes() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := ClaimsFrom(c)
		if claims == nil || !claims.APIKey {
			c.Next()
			return
		}
		path := strings.TrimPrefix(c.Request.URL.Path, "/api/v1")
		method := c.Request.Method
		blocked := isPathOrChild(path, "/api-keys") ||
			isPathOrChild(path, "/webhooks") ||
			isPathOrChild(path, "/users") ||
			isPathOrChild(path, "/auth") ||
			(method == http.MethodPost || method == http.MethodPut) && isPathOrChild(path, "/outlets") ||
			method == http.MethodPut && path == "/settings" ||
			method == http.MethodPost && path == "/presence/heartbeat"
		if blocked {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Kunci API tidak diizinkan untuk endpoint ini."})
			return
		}
		c.Next()
	}
}

func isPathOrChild(path, root string) bool {
	return path == root || strings.HasPrefix(path, root+"/")
}
