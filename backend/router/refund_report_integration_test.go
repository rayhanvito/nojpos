package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/adrr-dev/openPOS/backend/model"
	"github.com/adrr-dev/openPOS/backend/service"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRefundsAreNetAndAttributedToRefundDate(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "refund-report.db")
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DATABASE_URL", dbPath)
	t.Setenv("JWT_SECRET", "test-secret-refund-report")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	srv, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if srv.Cleanup != nil {
		defer srv.Cleanup()
	}

	doReq := func(method, path string, body any, token string) (*httptest.ResponseRecorder, map[string]any) {
		t.Helper()
		var bodyBytes []byte
		if body != nil {
			bodyBytes, _ = json.Marshal(body)
		}
		req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
		w := httptest.NewRecorder()
		srv.Handler.ServeHTTP(w, req)
		var response map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		return w, response
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

	email := fmt.Sprintf("refund-report-%d@example.com", time.Now().UnixNano())
	var otp string
	service.TestOnOTPSent = func(to, code string) {
		if to == email {
			otp = code
		}
	}
	defer func() { service.TestOnOTPSent = nil }()
	w, _ := doReq(http.MethodPost, "/auth/otp/send", map[string]any{"email": email}, "")
	assertStatus(w, http.StatusOK)
	w, _ = doReq(http.MethodPost, "/auth/otp/verify", map[string]any{"email": email, "code": otp}, "")
	assertStatus(w, http.StatusOK)
	w, response := doReq(http.MethodPost, "/auth/register", map[string]any{
		"name": "Admin Refund", "email": email, "password": "password123", "storeName": "Toko Refund",
	}, "")
	assertStatus(w, http.StatusCreated)
	token := response["access_token"].(string)

	w, product := doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Refund", "sku": "RFD-1", "buyPrice": 400, "sellPrice": 1001, "stock": 20, "unit": "pcs",
	}, token)
	assertStatus(w, http.StatusCreated)
	productID := uint(product["id"].(float64))
	w, _ = doReq(http.MethodPut, "/settings", map[string]any{
		"storeName": "Toko Refund", "taxEnabled": true, "taxPct": 10, "taxInclusive": false, "taxRounding": "none", "timezone": "Asia/Makassar",
	}, token)
	assertStatus(w, http.StatusOK)

	loc, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		t.Fatal(err)
	}
	today := time.Now().In(loc).Format("2006-01-02")
	_, _ = doReq(http.MethodPost, "/cashier/shift/start", map[string]any{"opening_cash": 100000}, token)
	checkout := func(qty int) uint {
		t.Helper()
		w, body := doReq(http.MethodPost, "/transactions", map[string]any{
			"items": []any{map[string]any{"productId": productID, "qty": qty}}, "method": "Cash", "paid": 100000,
		}, token)
		assertStatus(w, http.StatusCreated)
		return uint(body["id"].(float64))
	}
	refund := func(trxID uint, qty int) map[string]any {
		t.Helper()
		w, body := doReq(http.MethodPost, fmt.Sprintf("/transactions/%d/refund", trxID), map[string]any{
			"items": []any{map[string]any{"productId": productID, "qty": qty}}, "reason": "Salah beli",
		}, token)
		assertStatus(w, http.StatusOK)
		return body
	}
	getDashboard := func(date string) map[string]any {
		t.Helper()
		w, body := doReq(http.MethodGet, "/dashboard?date="+date, nil, token)
		assertStatus(w, http.StatusOK)
		return body["today"].(map[string]any)
	}
	getReport := func(date string) map[string]any {
		t.Helper()
		w, body := doReq(http.MethodGet, "/reports?period=today&date="+date, nil, token)
		assertStatus(w, http.StatusOK)
		return body["summary"].(map[string]any)
	}

	partialTrxID := checkout(6) // subtotal 6,006; tax 601; total 6,607.
	refund(partialTrxID, 3)     // item 3,003 + half-up proportional tax 301.
	dashboard := getDashboard(today)
	assertNum(dashboard, "omzet", 3303)
	assertNum(dashboard, "trx_count", 1)
	assertNum(dashboard, "items_sold", 3)
	assertNum(dashboard, "refund_total", 3304)
	report := getReport(today)
	assertNum(report, "omzet", 3303)
	assertNum(report, "trx_count", 1)
	assertNum(report, "items_sold", 3)
	assertNum(report, "gross_profit", 2103) // 3,303 net revenue less 1,200 net HPP.
	assertNum(report, "refund_total", 3304)
	assertNum(report, "refund_count", 1)

	w, detail := doReq(http.MethodGet, fmt.Sprintf("/transactions/%d", partialTrxID), nil, token)
	assertStatus(w, http.StatusOK)
	refunds, ok := detail["refunds"].([]any)
	if !ok || len(refunds) != 1 {
		t.Fatalf("transaction refund history: want one refund, got %v", detail["refunds"])
	}
	refundDetail := refunds[0].(map[string]any)
	if refundDetail["reason"] != "Salah beli" || refundDetail["actor"] != "Admin Refund" {
		t.Fatalf("refund detail fields: %v", refundDetail)
	}
	if refundDetail["id"].(float64) == 0 || refundDetail["created_at"] == "" {
		t.Fatalf("refund identity fields: %v", refundDetail)
	}
	refundItem := refundDetail["items"].([]any)[0].(map[string]any)
	if got := refundItem["qty"].(float64); got != 3 {
		t.Fatalf("refund item qty: want 3, got %v", got)
	}
	if got := refundItem["price"].(float64); got != 1001 {
		t.Fatalf("refund item snapshot price: want 1001, got %v", got)
	}

	fullTrxID := checkout(1)
	fullRefund := refund(fullTrxID, 1)
	if fullRefund["status"] != "refunded" {
		t.Fatalf("existing full refund status changed: %v", fullRefund["status"])
	}
	dashboard = getDashboard(today)
	assertNum(dashboard, "omzet", 3303)
	assertNum(dashboard, "trx_count", 2)
	assertNum(dashboard, "items_sold", 3)
	assertNum(dashboard, "refund_total", 4405)
	report = getReport(today)
	assertNum(report, "omzet", 3303)
	assertNum(report, "trx_count", 2)
	assertNum(report, "items_sold", 3)
	assertNum(report, "gross_profit", 2103)
	assertNum(report, "refund_total", 4405)
	assertNum(report, "refund_count", 2)

	crossDayTrxID := checkout(1)
	refund(crossDayTrxID, 1)
	dbConn, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer func() {
		sqlDB, _ := dbConn.DB()
		_ = sqlDB.Close()
	}()
	yesterdayNoon := time.Now().In(loc)
	yesterdayNoon = time.Date(yesterdayNoon.Year(), yesterdayNoon.Month(), yesterdayNoon.Day(), 12, 0, 0, 0, loc).AddDate(0, 0, -1)
	if err := dbConn.Model(&model.Trx{}).Where("id = ?", crossDayTrxID).Update("created_at", yesterdayNoon.UTC()).Error; err != nil {
		t.Fatalf("move original sale to previous local day: %v", err)
	}
	yesterday := yesterdayNoon.Format("2006-01-02")
	dashboard = getDashboard(yesterday)
	assertNum(dashboard, "omzet", 1101)
	assertNum(dashboard, "trx_count", 1)
	assertNum(dashboard, "items_sold", 1)
	assertNum(dashboard, "refund_total", 0)
	report = getReport(yesterday)
	assertNum(report, "omzet", 1101)
	assertNum(report, "trx_count", 1)
	assertNum(report, "items_sold", 1)
	assertNum(report, "gross_profit", 701)
	assertNum(report, "refund_total", 0)
	assertNum(report, "refund_count", 0)

	w, promoProduct := doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Refund Promo", "sku": "RFD-PROMO-1", "buyPrice": 2000, "sellPrice": 5000, "stock": 10, "unit": "pcs",
	}, token)
	assertStatus(w, http.StatusCreated)
	promoProductID := uint(promoProduct["id"].(float64))
	w, _ = doReq(http.MethodPost, "/promos", map[string]any{
		"name": "Voucher Refund", "type": "amount", "value": 1000, "scope": "all", "voucherCode": "RFD-PROMO",
		"startsAt": time.Now().UTC().Add(-time.Hour).Format(time.RFC3339), "endsAt": time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
		"minPurchase": 0,
	}, token)
	assertStatus(w, http.StatusCreated)
	w, promoSale := doReq(http.MethodPost, "/transactions", map[string]any{
		"items":       []any{map[string]any{"productId": promoProductID, "qty": 2}},
		"voucherCode": "RFD-PROMO", "method": "Cash", "paid": 100000,
	}, token)
	assertStatus(w, http.StatusCreated)
	promoTrxID := uint(promoSale["id"].(float64))
	if promoSale["subtotal"] != float64(10000) || promoSale["voucher_discount"] != float64(1000) || promoSale["tax"] != float64(900) {
		t.Fatalf("promo transaction before refund: %v", promoSale)
	}
	w, _ = doReq(http.MethodPost, fmt.Sprintf("/transactions/%d/refund", promoTrxID), map[string]any{
		"items": []any{map[string]any{"productId": promoProductID, "qty": 1}}, "reason": "Refund transaksi promo",
	}, token)
	assertStatus(w, http.StatusOK)
	assertNum(getDashboard(today), "refund_total", 11006) // Refund yang sudah ada + Rp5.000 item dan Rp500 pajak.
	assertNum(getReport(today), "refund_total", 11006)
}
