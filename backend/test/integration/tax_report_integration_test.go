package integration

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/rayhanvito/nojpos/backend/internal/app"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rayhanvito/nojpos/backend/internal/auth"
	"github.com/rayhanvito/nojpos/backend/internal/sales"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTaxReportAndCSV(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "tax-report.db")
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DATABASE_URL", dbPath)
	t.Setenv("JWT_SECRET", "test-secret-tax-report")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	srv, err := app.New(ctx)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if srv.Cleanup != nil {
		defer srv.Cleanup()
	}

	doReq := func(method, path string, body any, token string) *httptest.ResponseRecorder {
		t.Helper()
		var bodyBytes []byte
		if body != nil {
			bodyBytes, _ = json.Marshal(body)
		}
		req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(bodyBytes))
		req.RemoteAddr = "198.51.100.42:12345"
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		srv.Handler.ServeHTTP(w, req)
		return w
	}
	responseMap := func(w *httptest.ResponseRecorder) map[string]any {
		t.Helper()
		var response map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode JSON response: %v; body: %s", err, w.Body.String())
		}
		return response
	}
	assertStatus := func(w *httptest.ResponseRecorder, want int) {
		t.Helper()
		if w.Code != want {
			t.Fatalf("status: want %d, got %d: %s", want, w.Code, w.Body.String())
		}
	}
	assertNum := func(m map[string]any, key string, want int64) {
		t.Helper()
		got, ok := m[key].(float64)
		if !ok || int64(got) != want {
			t.Fatalf("%s: want %d, got %v", key, want, m[key])
		}
	}

	email := fmt.Sprintf("tax-report-%d@example.com", time.Now().UnixNano())
	var otp string
	auth.TestOnOTPSent = func(to, code string) {
		if to == email {
			otp = code
		}
	}
	defer func() { auth.TestOnOTPSent = nil }()
	w := doReq(http.MethodPost, "/auth/otp/send", map[string]any{"email": email}, "")
	assertStatus(w, http.StatusOK)
	w = doReq(http.MethodPost, "/auth/otp/verify", map[string]any{"email": email, "code": otp}, "")
	assertStatus(w, http.StatusOK)
	w = doReq(http.MethodPost, "/auth/register", map[string]any{
		"name": "Admin Pajak", "email": email, "password": "password123", "storeName": "Toko Pajak",
	}, "")
	assertStatus(w, http.StatusCreated)
	adminToken := responseMap(w)["access_token"].(string)

	w = doReq(http.MethodPut, "/settings", map[string]any{
		"storeName": "Toko Pajak", "taxEnabled": true, "taxPct": 10,
		"taxInclusive": false, "taxRounding": "none", "timezone": "Asia/Tokyo",
	}, adminToken)
	assertStatus(w, http.StatusOK)
	w = doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Pajak", "sku": "TAX-1", "buyPrice": 500, "sellPrice": 1000, "stock": 10, "unit": "pcs",
	}, adminToken)
	assertStatus(w, http.StatusCreated)
	productID := uint(responseMap(w)["id"].(float64))
	doReq(http.MethodPost, "/cashier/shift/start", map[string]any{"opening_cash": 100000}, adminToken)
	w = doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 2}}, "method": "Cash", "paid": 2200,
	}, adminToken)
	assertStatus(w, http.StatusCreated)
	trxID := uint(responseMap(w)["id"].(float64))
	w = doReq(http.MethodPost, fmt.Sprintf("/transactions/%d/refund", trxID), map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "reason": "Salah beli",
	}, adminToken)
	assertStatus(w, http.StatusOK)

	testDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer func() {
		sqlDB, _ := testDB.DB()
		_ = sqlDB.Close()
	}()
	localMonthStart := time.Date(2026, time.November, 1, 0, 0, 0, 0, time.FixedZone("JST", 9*60*60))
	saleAt := localMonthStart.Add(30 * time.Minute).UTC()
	refundAt := localMonthStart.Add(time.Hour).UTC()
	if err := testDB.Model(&sales.Trx{}).Where("id = ?", trxID).Update("created_at", saleAt).Error; err != nil {
		t.Fatalf("set transaction month boundary: %v", err)
	}
	if err := testDB.Model(&sales.Refund{}).Where("trx_id = ?", trxID).Update("created_at", refundAt).Error; err != nil {
		t.Fatalf("set refund month boundary: %v", err)
	}

	w = doReq(http.MethodGet, "/reports/tax?month=2026-11", nil, adminToken)
	assertStatus(w, http.StatusOK)
	report := responseMap(w)
	if report["month"] != "2026-11" {
		t.Fatalf("report month: want 2026-11, got %v", report["month"])
	}
	assertNum(report, "omzet_bruto", 2200)
	assertNum(report, "refund_total", 1100)
	assertNum(report, "omzet_neto", 1100)
	assertNum(report, "dpp", 2000)
	assertNum(report, "ppn_terkumpul", 200)
	assertNum(report, "trx_count", 1)
	assertNum(report, "refund_count", 1)
	assertNum(report, "pph_final_estimate", 6)

	w = doReq(http.MethodGet, "/reports/tax/export?month=2026-11", nil, adminToken)
	assertStatus(w, http.StatusOK)
	if !strings.HasPrefix(w.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("tax CSV content type: %q", w.Header().Get("Content-Type"))
	}
	if want := `attachment; filename="laporan-pajak-2026-11.csv"`; w.Header().Get("Content-Disposition") != want {
		t.Fatalf("tax CSV content disposition: want %q, got %q", want, w.Header().Get("Content-Disposition"))
	}
	rows, err := csv.NewReader(strings.NewReader(w.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse tax CSV: %v", err)
	}
	wantHeader := []string{"bulan", "omzet bruto", "total refund", "omzet neto", "DPP", "PPN terkumpul", "jumlah transaksi", "jumlah refund", "estimasi PPh Final"}
	wantRow := []string{"2026-11", "2200", "1100", "1100", "2000", "200", "1", "1", "6"}
	if len(rows) != 2 || strings.Join(rows[0], "|") != strings.Join(wantHeader, "|") || strings.Join(rows[1], "|") != strings.Join(wantRow, "|") {
		t.Fatalf("tax CSV rows: want header %v and row %v, got %#v", wantHeader, wantRow, rows)
	}
	jsonRow := []string{
		report["month"].(string),
		fmt.Sprint(report["omzet_bruto"]),
		fmt.Sprint(report["refund_total"]),
		fmt.Sprint(report["omzet_neto"]),
		fmt.Sprint(report["dpp"]),
		fmt.Sprint(report["ppn_terkumpul"]),
		fmt.Sprint(report["trx_count"]),
		fmt.Sprint(report["refund_count"]),
		fmt.Sprint(report["pph_final_estimate"]),
	}
	if strings.Join(rows[1], "|") != strings.Join(jsonRow, "|") {
		t.Fatalf("tax CSV row does not match JSON report: csv=%v json=%v", rows[1], jsonRow)
	}

	w = doReq(http.MethodGet, "/reports/tax?month=2026-10", nil, adminToken)
	assertStatus(w, http.StatusOK)
	empty := responseMap(w)
	if empty["month"] != "2026-10" {
		t.Fatalf("empty report month: want 2026-10, got %v", empty["month"])
	}
	for _, key := range []string{"omzet_bruto", "refund_total", "omzet_neto", "dpp", "ppn_terkumpul", "trx_count", "refund_count", "pph_final_estimate"} {
		assertNum(empty, key, 0)
	}

	for _, path := range []string{"/reports/tax?month=2026-13", "/reports/tax?month="} {
		w = doReq(http.MethodGet, path, nil, adminToken)
		assertStatus(w, http.StatusBadRequest)
		if responseMap(w)["error"] == nil {
			t.Fatalf("invalid-month response has no error field: %s", w.Body.String())
		}
	}

	w = doReq(http.MethodGet, "/reports/tax", nil, adminToken)
	assertStatus(w, http.StatusOK)
	defaultReport := responseMap(w)
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatal(err)
	}
	if want := time.Now().In(loc).Format("2006-01"); defaultReport["month"] != want {
		t.Fatalf("default report month: want %s, got %v", want, defaultReport["month"])
	}

	w = doReq(http.MethodPost, "/users", map[string]any{"name": "Kasir Pajak"}, adminToken)
	assertStatus(w, http.StatusCreated)
	cashierID := uint(responseMap(w)["user"].(map[string]any)["id"].(float64))
	w = doReq(http.MethodPut, fmt.Sprintf("/users/%d/passcode", cashierID), map[string]any{"passcode": "12345"}, adminToken)
	assertStatus(w, http.StatusOK)
	w = doReq(http.MethodPost, "/auth/switch", map[string]any{"target_user_id": cashierID, "passcode": "12345", "role": "cashier"}, adminToken)
	assertStatus(w, http.StatusOK)
	cashierToken := responseMap(w)["access_token"].(string)
	for _, path := range []string{"/reports/tax?month=2026-11", "/reports/tax/export?month=2026-11"} {
		w = doReq(http.MethodGet, path, nil, cashierToken)
		if w.Code != http.StatusForbidden {
			t.Fatalf("cashier GET %s: want 403, got %d: %s", path, w.Code, w.Body.String())
		}
	}
}
