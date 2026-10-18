package shift

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

func TestListShiftsUsesCanonicalPagination(t *testing.T) {
	shifts := make([]*CashierShift, 5)
	for i := range shifts {
		shifts[i] = &CashierShift{CashierName: fmt.Sprintf("Kasir %d", i+1)}
	}
	h := NewHandler(NewService(&fakeShifts{listed: shifts}, fakeCashiers{}, fakeStores{}))
	for _, tc := range []struct {
		name  string
		query string
		items []string
		page  int
		limit int
	}{
		{name: "tanpa parameter", items: []string{"Kasir 1", "Kasir 2", "Kasir 3", "Kasir 4", "Kasir 5"}, page: 1, limit: 5},
		{name: "halaman kedua", query: "?page=2&limit=2", items: []string{"Kasir 3", "Kasir 4"}, page: 2, limit: 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/shifts"+tc.query, nil)
			c.Set("claims", &middleware.Claims{StoreID: 1, Role: middleware.RoleAdmin})
			h.ListShifts(c)
			assertShiftListResponse(t, w, tc.items, 5, tc.page, tc.limit)
		})
	}
}

func assertShiftListResponse(t *testing.T, w *httptest.ResponseRecorder, wantNames []string, total, page, limit int) {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
	}
	var body map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body) != 4 {
		t.Fatalf("respons harus hanya berisi items, total, page, limit, dapat %s", w.Body.String())
	}
	if _, ok := body["shifts"]; ok {
		t.Fatalf("respons tidak boleh memiliki kunci shifts: %s", w.Body.String())
	}
	var items []map[string]json.RawMessage
	var gotTotal, gotPage, gotLimit int
	for key, target := range map[string]any{"items": &items, "total": &gotTotal, "page": &gotPage, "limit": &gotLimit} {
		if err := json.Unmarshal(body[key], target); err != nil {
			t.Fatalf("field %s: %v", key, err)
		}
	}
	if len(items) != len(wantNames) || gotTotal != total || gotPage != page || gotLimit != limit {
		t.Fatalf("respons daftar salah: items=%d total=%d page=%d limit=%d, mau %d/%d/%d/%d", len(items), gotTotal, gotPage, gotLimit, len(wantNames), total, page, limit)
	}
	for i, wantName := range wantNames {
		var gotName string
		if err := json.Unmarshal(items[i]["cashier_name"], &gotName); err != nil {
			t.Fatalf("nama kasir %d: %v", i, err)
		}
		if gotName != wantName {
			t.Fatalf("kasir %d = %q, mau %q", i, gotName, wantName)
		}
	}
}
