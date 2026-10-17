// Package httpx merender error ke respons HTTP dengan satu pola:
// *apierr.Error dirender dengan status + pesan + kode miliknya; error lain
// menjadi 500 generik (pesan internal tidak pernah bocor ke pengguna) dan
// dicatat satu baris JSON lewat slog.
package httpx

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"

	apierr "github.com/rayhanvito/nojpos/backend/internal/platform/errors"
)

// logger menulis satu baris JSON per error internal ke stderr.
var logger = slog.New(slog.NewJSONHandler(os.Stderr, nil))

// RequestIDKey adalah key konteks gin tempat middleware request-id menyimpan ID.
const RequestIDKey = "requestId"

// RequestID membaca request ID dari konteks gin ("" bila tidak ada).
func RequestID(c *gin.Context) string {
	if v, ok := c.Get(RequestIDKey); ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// Fail merender err sebagai respons JSON {"error": ..., "code": ...}.
func Fail(c *gin.Context, err error) {
	if e, ok := apierr.Of(err); ok {
		c.JSON(e.Status, gin.H{"error": e.Message, "code": e.Code})
		return
	}
	logger.Error("internal error",
		"request_id", RequestID(c),
		"method", c.Request.Method,
		"path", c.Request.URL.Path,
		"error", err.Error(),
	)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "Terjadi kesalahan pada server.", "code": "INTERNAL"})
}
