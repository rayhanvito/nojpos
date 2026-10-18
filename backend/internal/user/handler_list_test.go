package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/auth"
	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
	"github.com/rayhanvito/nojpos/backend/internal/staff"
)

type listUsersRepo struct{}

func (listUsersRepo) GetByID(context.Context, uint) (*auth.User, error) { return nil, nil }

func (listUsersRepo) ListByStore(context.Context, uint) ([]*auth.User, error) {
	return []*auth.User{}, nil
}

func (listUsersRepo) SetPasscode(context.Context, uint, *string) error { return nil }

type listCashiersRepo struct{ items []*staff.Cashier }

func (listCashiersRepo) GetByIDUnscoped(context.Context, uint) (*staff.Cashier, error) {
	return nil, nil
}

func (f listCashiersRepo) ListByStore(context.Context, uint) ([]*staff.Cashier, error) {
	return f.items, nil
}

func (listCashiersRepo) Create(context.Context, uint, uint, string) (*staff.Cashier, error) {
	return nil, nil
}

func (listCashiersRepo) UpdateProfile(context.Context, uint, uint, *string, *uint) error {
	return nil
}

func (listCashiersRepo) SetActive(context.Context, uint, bool) error { return nil }
func (listCashiersRepo) SetPasscode(context.Context, uint, *string) error {
	return nil
}
func (listCashiersRepo) Delete(context.Context, uint) error { return nil }
func (listCashiersRepo) UpdateName(context.Context, uint, string) error {
	return nil
}

func TestUserListUsesCanonicalPagination(t *testing.T) {
	items := make([]*staff.Cashier, 5)
	for i := range items {
		items[i] = &staff.Cashier{Name: fmt.Sprintf("Kasir %d", i+1), Active: true}
	}
	h := NewHandler(NewService(listUsersRepo{}, listCashiersRepo{items: items}))
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
			c.Request = httptest.NewRequest(http.MethodGet, "/users"+tc.query, nil)
			c.Set("claims", &middleware.Claims{StoreID: 1, Role: middleware.RoleAdmin})
			h.List(c)
			assertUserListResponse(t, w, tc.items, 5, tc.page, tc.limit)
		})
	}
}

func assertUserListResponse(t *testing.T, w *httptest.ResponseRecorder, wantNames []string, total, page, limit int) {
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
		if err := json.Unmarshal(items[i]["name"], &gotName); err != nil {
			t.Fatalf("nama item %d: %v", i, err)
		}
		if gotName != wantName {
			t.Fatalf("item %d = %q, mau %q", i, gotName, wantName)
		}
	}
}
