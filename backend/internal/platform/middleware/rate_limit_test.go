package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// runLimiter menjalankan n permintaan POST ke handler yang hanya dibungkus
// limiter, lalu mengembalikan urutan status HTTP-nya. Semua permintaan
// berasal dari IP httptest yang sama, seperti pemakaian nyata per-IP.
func runLimiter(t *testing.T, mw gin.HandlerFunc, n int) []int {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/x", mw, func(c *gin.Context) { c.Status(http.StatusOK) })
	codes := make([]int, 0, n)
	for i := 0; i < n; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/x", nil)
		r.ServeHTTP(w, req)
		codes = append(codes, w.Code)
	}
	return codes
}

func TestRateLimitOTP429SetelahAmbang(t *testing.T) {
	ResetRateLimiters()
	defer ResetRateLimiters()

	codes := runLimiter(t, RateLimitOTP(), 6)
	for i := 0; i < 5; i++ {
		if codes[i] != http.StatusOK {
			t.Fatalf("permintaan OTP ke-%d: mau 200, dapat %d", i+1, codes[i])
		}
	}
	if codes[5] != http.StatusTooManyRequests {
		t.Fatalf("permintaan OTP ke-6: mau 429, dapat %d", codes[5])
	}
}

func TestRateLimitRegister429SetelahAmbang(t *testing.T) {
	ResetRateLimiters()
	defer ResetRateLimiters()

	codes := runLimiter(t, RateLimitRegister(), 11)
	for i := 0; i < 10; i++ {
		if codes[i] != http.StatusOK {
			t.Fatalf("permintaan register ke-%d: mau 200, dapat %d", i+1, codes[i])
		}
	}
	if codes[10] != http.StatusTooManyRequests {
		t.Fatalf("permintaan register ke-11: mau 429, dapat %d", codes[10])
	}
}
