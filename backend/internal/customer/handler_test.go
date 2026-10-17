package customer

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// Dipindah verbatim dari handler/customer_test.go (BE-F4): fungsi
// respondCustomerErr kini tinggal di paket ini.
func TestRespondCustomerErrHidesUnexpectedError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)

	respondCustomerErr(ctx, errors.New("database secret"))

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("unexpected error status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
	if got := recorder.Body.String(); got != `{"error":"Gagal memproses pelanggan."}` {
		t.Fatalf("unexpected error response = %s", got)
	}
}
