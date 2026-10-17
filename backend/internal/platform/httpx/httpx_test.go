package httpx

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	apierr "github.com/rayhanvito/nojpos/backend/internal/platform/errors"
)

func testContext() (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	return c, w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("body bukan JSON valid: %v (%q)", err, w.Body.String())
	}
	return body
}

func TestFailRendersApierr(t *testing.T) {
	c, w := testContext()
	Fail(c, apierr.New(http.StatusConflict, "SKU_TAKEN", "SKU sudah digunakan di toko ini."))
	if w.Code != http.StatusConflict {
		t.Fatalf("status = %d, mau 409", w.Code)
	}
	body := decodeBody(t, w)
	if body["error"] != "SKU sudah digunakan di toko ini." || body["code"] != "SKU_TAKEN" {
		t.Fatalf("body salah: %v", body)
	}
}

func TestFailRendersUnknownAsGeneric500(t *testing.T) {
	c, w := testContext()
	c.Set(RequestIDKey, "req-123")
	Fail(c, errors.New("detail internal yang sensitif"))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, mau 500", w.Code)
	}
	body := decodeBody(t, w)
	if body["error"] != "Terjadi kesalahan pada server." || body["code"] != "INTERNAL" {
		t.Fatalf("body harus generik: %v", body)
	}
}

func TestRequestIDHelper(t *testing.T) {
	c, _ := testContext()
	if got := RequestID(c); got != "" {
		t.Fatalf("tanpa middleware harus kosong, dapat %q", got)
	}
	c.Set(RequestIDKey, "abc-def")
	if got := RequestID(c); got != "abc-def" {
		t.Fatalf("RequestID = %q", got)
	}
}
