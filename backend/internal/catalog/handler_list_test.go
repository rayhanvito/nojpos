package catalog

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

func TestListCategoriesUsesCanonicalPagination(t *testing.T) {
	categories := make([]*Category, 5)
	for i := range categories {
		categories[i] = &Category{Name: fmt.Sprintf("Kategori %d", i+1)}
	}
	h := NewHandler(NewService(&fakeCats{items: categories}, &fakeProds{}, nil))
	for _, tc := range []struct {
		query string
		items []string
		page  int
		limit int
	}{
		{query: "", items: []string{"Kategori 1", "Kategori 2", "Kategori 3", "Kategori 4", "Kategori 5"}, page: 1, limit: 5},
		{query: "?page=2&limit=2", items: []string{"Kategori 3", "Kategori 4"}, page: 2, limit: 2},
	} {
		t.Run(tc.query, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest(http.MethodGet, "/categories"+tc.query, nil)
			c.Set("claims", &middleware.Claims{StoreID: 1, Role: middleware.RoleAdmin})
			h.ListCategories(c)
			assertCanonicalListResponse(t, w, "name", tc.items, 5, tc.page, tc.limit)
		})
	}
}

func TestListCategoriesSerializesProductCount(t *testing.T) {
	s := newTenantCatalogSeed(t)
	h := NewHandler(NewService(NewCategoryRepo(s.db), NewProductRepo(s.db), nil))
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/categories", nil)
	c.Set("claims", &middleware.Claims{StoreID: s.storeA, Role: middleware.RoleAdmin})

	h.ListCategories(c)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, body=%s", w.Code, w.Body.String())
	}
	var body struct {
		Items []map[string]json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Items) != 1 {
		t.Fatalf("items = %d, mau 1; body=%s", len(body.Items), w.Body.String())
	}
	var got int64
	if err := json.Unmarshal(body.Items[0]["product_count"], &got); err != nil {
		t.Fatalf("product_count tidak terserialisasi: %v; body=%s", err, w.Body.String())
	}
	if got != 1 {
		t.Fatalf("product_count = %d, mau 1", got)
	}
}

func assertCanonicalListResponse(t *testing.T, w *httptest.ResponseRecorder, itemNameKey string, wantNames []string, total, page, limit int) {
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
	var items []map[string]json.RawMessage
	var gotTotal, gotPage, gotLimit int
	if err := json.Unmarshal(body["items"], &items); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body["total"], &gotTotal); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body["page"], &gotPage); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(body["limit"], &gotLimit); err != nil {
		t.Fatal(err)
	}
	if len(items) != len(wantNames) || gotTotal != total || gotPage != page || gotLimit != limit {
		t.Fatalf("respons daftar salah: items=%d total=%d page=%d limit=%d, mau %d/%d/%d/%d", len(items), gotTotal, gotPage, gotLimit, len(wantNames), total, page, limit)
	}
	for i, wantName := range wantNames {
		var gotName string
		if err := json.Unmarshal(items[i][itemNameKey], &gotName); err != nil {
			t.Fatalf("nama item %d: %v", i, err)
		}
		if gotName != wantName {
			t.Fatalf("item %d = %q, mau %q", i, gotName, wantName)
		}
	}
}
