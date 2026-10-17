package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// requestIDCtxKey adalah key privat untuk request ID di context.Context.
type requestIDCtxKey struct{}

// RequestIDFromContext membaca request ID dari context.Context ("" bila tidak ada).
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(requestIDCtxKey{}).(string); ok {
		return v
	}
	return ""
}

// RequestID memastikan setiap request punya ID: memakai header X-Request-Id
// dari pemanggil bila ada, selain itu membuat ID acak 16 byte hex. ID disimpan
// di konteks gin ("requestId"), di request context, dan di header respons.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader("X-Request-Id")
		if id == "" {
			var b [16]byte
			if _, err := rand.Read(b[:]); err == nil {
				id = hex.EncodeToString(b[:])
			}
		}
		c.Set("requestId", id)
		c.Request = c.Request.WithContext(context.WithValue(c.Request.Context(), requestIDCtxKey{}, id))
		c.Header("X-Request-Id", id)
		c.Next()
	}
}
