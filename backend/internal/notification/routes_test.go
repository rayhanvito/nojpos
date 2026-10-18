package notification

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

type fakeNotificationRepo struct{}

func (fakeNotificationRepo) Create(context.Context, *Notification) error { return nil }

func (fakeNotificationRepo) List(context.Context, uint, string, string, bool, int, int) ([]*Notification, int, error) {
	return []*Notification{}, 0, nil
}

func (fakeNotificationRepo) UnreadCount(context.Context, uint) (int64, error) { return 0, nil }

func (fakeNotificationRepo) GetByID(context.Context, uint, uint) (*Notification, error) {
	return nil, nil
}

func (fakeNotificationRepo) MarkAsRead(context.Context, uint, uint) error { return nil }

func (fakeNotificationRepo) MarkAllAsRead(context.Context, uint) error { return nil }

func (fakeNotificationRepo) Delete(context.Context, uint, uint) error { return nil }

func (fakeNotificationRepo) HasUnreadForReference(context.Context, uint, string, string) (bool, error) {
	return false, nil
}

func notificationTestEngine(role middleware.Role) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	group := r.Group("/api/v1")
	group.Use(func(c *gin.Context) {
		c.Set("claims", &middleware.Claims{UserID: 1, StoreID: 1, Role: role})
		c.Next()
	})
	admin := group.Group("")
	admin.Use(middleware.RequireRole(string(middleware.RoleAdmin)))
	RegisterRoutes(admin, NewHandler(NewService(fakeNotificationRepo{})))
	return r
}

func TestNotificationRoutesRequireAdmin(t *testing.T) {
	tests := []struct {
		name string
		role middleware.Role
		want int
	}{
		{name: "kasir ditolak", role: middleware.RoleCashier, want: http.StatusForbidden},
		{name: "admin diizinkan", role: middleware.RoleAdmin, want: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/notifications", nil)
			notificationTestEngine(tt.role).ServeHTTP(w, req)
			if w.Code != tt.want {
				t.Fatalf("status = %d, mau %d: %s", w.Code, tt.want, w.Body.String())
			}
		})
	}
}

func TestNotificationListHandlerDoesNotEnforceRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(NewService(fakeNotificationRepo{}))
	r.GET("/notifications", func(c *gin.Context) {
		c.Set("claims", &middleware.Claims{UserID: 1, StoreID: 1, Role: middleware.RoleCashier})
		c.Next()
	}, h.List)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/notifications", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("handler list harus menyerahkan pemeriksaan peran ke middleware, status = %d: %s", w.Code, w.Body.String())
	}
}
