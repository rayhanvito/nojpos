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
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/rayhanvito/nojpos/backend/internal/auth"
	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
	"github.com/rayhanvito/nojpos/backend/internal/sales"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCSVExportsAreScopedAndAdminOnly(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "csv-export.db")
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DATABASE_URL", dbPath)
	t.Setenv("JWT_SECRET", "test-secret-csv-export")

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

	otpCodes := map[string]string{}
	auth.TestOnOTPSent = func(email, code string) { otpCodes[email] = code }
	defer func() { auth.TestOnOTPSent = nil }()
	register := func(email, name, storeName string) string {
		t.Helper()
		w := doReq(http.MethodPost, "/auth/otp/send", map[string]any{"email": email}, "")
		assertStatus(w, http.StatusOK)
		w = doReq(http.MethodPost, "/auth/otp/verify", map[string]any{"email": email, "code": otpCodes[email]}, "")
		assertStatus(w, http.StatusOK)
		w = doReq(http.MethodPost, "/auth/register", map[string]any{
			"name": name, "email": email, "password": "password123", "storeName": storeName,
		}, "")
		assertStatus(w, http.StatusCreated)
		return responseMap(w)["access_token"].(string)
	}
	createProduct := func(token, name, sku, barcode string, categoryID any, buy, sell int64, stock int, unit string) map[string]any {
		t.Helper()
		body := map[string]any{
			"name": name, "sku": sku, "barcode": barcode, "buyPrice": buy, "sellPrice": sell, "stock": stock, "unit": unit,
		}
		if categoryID != nil {
			body["categoryId"] = categoryID
		}
		w := doReq(http.MethodPost, "/products", body, token)
		assertStatus(w, http.StatusCreated)
		return responseMap(w)
	}

	adminToken := register("csv-export-admin@example.com", "Admin CSV", "Toko CSV")
	w := doReq(http.MethodPut, "/settings", map[string]any{"storeName": "Toko CSV", "timezone": "Asia/Jakarta"}, adminToken)
	assertStatus(w, http.StatusOK)
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		t.Fatal(err)
	}

	w = doReq(http.MethodPost, "/categories", map[string]any{"name": "Minuman, \"Kopi\""}, adminToken)
	assertStatus(w, http.StatusCreated)
	categoryID := responseMap(w)["category"].(map[string]any)["id"].(float64)
	product := createProduct(adminToken, "Kopi, \"Premium\"\nBubuk", "CSV-1", "899-CSV-1", categoryID, 500, 1234, 10, "bungkus")
	productID := uint(product["id"].(float64))
	createProduct(adminToken, "Teh", "CSV-2", "899-CSV-2", nil, 700, 999, 5, "pcs")

	doReq(http.MethodPost, "/cashier/shift/start", map[string]any{"opening_cash": 100000}, adminToken)
	w = doReq(http.MethodPost, "/transactions", map[string]any{
		"items":    []any{map[string]any{"productId": productID, "qty": 2}},
		"discount": 34, "method": "Cash", "paid": 2500, "customer": "=1+1, \"VIP\"\nBuyer",
	}, adminToken)
	assertStatus(w, http.StatusCreated)
	trx := responseMap(w)
	trxID := uint(trx["id"].(float64))
	trxCreatedAt, err := time.Parse(time.RFC3339Nano, trx["created_at"].(string))
	if err != nil {
		t.Fatalf("parse created_at: %v", err)
	}

	w = doReq(http.MethodPatch, fmt.Sprintf("/products/%d/active", productID), map[string]any{"active": false}, adminToken)
	assertStatus(w, http.StatusOK)

	otherStoreToken := register("csv-export-other@example.com", "Admin Lain", "Toko Lain")
	doReq(http.MethodPost, "/cashier/shift/start", map[string]any{"opening_cash": 100000}, otherStoreToken)
	otherProduct := createProduct(otherStoreToken, "Produk Rahasia", "OTHER-1", "899-OTHER-1", nil, 100, 777, 3, "pcs")
	w = doReq(http.MethodPost, "/transactions", map[string]any{
		"items":  []any{map[string]any{"productId": uint(otherProduct["id"].(float64)), "qty": 1}},
		"method": "Cash", "paid": 777,
	}, otherStoreToken)
	assertStatus(w, http.StatusCreated)

	localDate := trxCreatedAt.In(loc).Format("2006-01-02")
	w = doReq(http.MethodGet, "/transactions/export?period=today&date="+localDate, nil, adminToken)
	assertStatus(w, http.StatusOK)
	if !strings.HasPrefix(w.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("transactions content type: %q", w.Header().Get("Content-Type"))
	}
	if want := fmt.Sprintf(`attachment; filename="transactions-%s.csv"`, time.Now().In(loc).Format("2006-01-02")); w.Header().Get("Content-Disposition") != want {
		t.Fatalf("transactions content disposition: want %q, got %q", want, w.Header().Get("Content-Disposition"))
	}
	trxRows, err := csv.NewReader(strings.NewReader(w.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse transactions CSV: %v", err)
	}
	wantTrxHeader := []string{"id", "tanggal", "kasir", "pelanggan", "metode", "status", "subtotal", "diskon", "pajak", "total", "dibayar", "kembalian", "jumlah item"}
	if len(trxRows) != 2 || strings.Join(trxRows[0], "|") != strings.Join(wantTrxHeader, "|") {
		t.Fatalf("transaction CSV header/rows: %#v", trxRows)
	}
	wantTimestamp := trxCreatedAt.In(loc).Format(time.RFC3339)
	wantTrxRow := []string{strconv.FormatUint(uint64(trxID), 10), wantTimestamp, "Admin CSV", "'=1+1, \"VIP\"\nBuyer", "Cash", "completed", "2468", "34", "0", "2434", "2500", "66", "2"}
	if strings.Join(trxRows[1], "|") != strings.Join(wantTrxRow, "|") {
		t.Fatalf("transaction CSV row: want %#v, got %#v", wantTrxRow, trxRows[1])
	}
	w = doReq(http.MethodGet, "/transactions/export?period=all&date=1900-01-01", nil, adminToken)
	assertStatus(w, http.StatusOK)
	allRows, err := csv.NewReader(strings.NewReader(w.Body.String())).ReadAll()
	if err != nil || len(allRows) != 2 {
		t.Fatalf("period=all should ignore date and include the transaction: rows=%#v err=%v", allRows, err)
	}
	testDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	defer func() {
		sqlDB, _ := testDB.DB()
		_ = sqlDB.Close()
	}()
	var exportStore outlet.Store
	if err := testDB.Where("name = ?", "Toko CSV").First(&exportStore).Error; err != nil {
		t.Fatalf("find export store: %v", err)
	}
	var admin auth.User
	if err := testDB.Where("email = ?", "csv-export-admin@example.com").First(&admin).Error; err != nil {
		t.Fatalf("find export admin: %v", err)
	}
	bulkTransactions := make([]sales.Trx, 201)
	bulkCreatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := range bulkTransactions {
		bulkTransactions[i] = sales.Trx{
			BaseModel:   database.BaseModel{CreatedAt: bulkCreatedAt, UpdatedAt: bulkCreatedAt},
			StoreID:     exportStore.ID,
			CashierID:   admin.ID,
			CashierName: "Admin CSV",
			Items:       sales.TrxItemList{},
			Method:      "Cash",
			Status:      sales.TrxCompleted,
		}
	}
	if err := testDB.CreateInBatches(&bulkTransactions, 50).Error; err != nil {
		t.Fatalf("seed unpaginated export rows: %v", err)
	}
	w = doReq(http.MethodGet, "/transactions/export?period=all", nil, adminToken)
	assertStatus(w, http.StatusOK)
	allRows, err = csv.NewReader(strings.NewReader(w.Body.String())).ReadAll()
	if err != nil || len(allRows) != len(bulkTransactions)+2 {
		t.Fatalf("period=all should include more than one page of store rows: count=%d err=%v", len(allRows), err)
	}
	w = doReq(http.MethodGet, "/transactions/export?period=today&date=1900-01-01", nil, adminToken)
	assertStatus(w, http.StatusOK)
	emptyDayRows, err := csv.NewReader(strings.NewReader(w.Body.String())).ReadAll()
	if err != nil || len(emptyDayRows) != 1 || strings.Join(emptyDayRows[0], "|") != strings.Join(wantTrxHeader, "|") {
		t.Fatalf("empty local date should return only the header: rows=%#v err=%v", emptyDayRows, err)
	}

	w = doReq(http.MethodGet, "/products/export", nil, adminToken)
	assertStatus(w, http.StatusOK)
	if !strings.HasPrefix(w.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("products content type: %q", w.Header().Get("Content-Type"))
	}
	if want := fmt.Sprintf(`attachment; filename="products-%s.csv"`, time.Now().In(loc).Format("2006-01-02")); w.Header().Get("Content-Disposition") != want {
		t.Fatalf("products content disposition: want %q, got %q", want, w.Header().Get("Content-Disposition"))
	}
	productRows, err := csv.NewReader(strings.NewReader(w.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse products CSV: %v", err)
	}
	wantProductHeader := []string{"id", "nama", "sku", "barcode", "kategori", "harga beli", "harga jual", "stok", "satuan", "aktif"}
	if len(productRows) != 3 || strings.Join(productRows[0], "|") != strings.Join(wantProductHeader, "|") {
		t.Fatalf("product CSV header/rows: %#v", productRows)
	}
	productsBySKU := make(map[string][]string, len(productRows)-1)
	for _, row := range productRows[1:] {
		productsBySKU[row[2]] = row
	}
	wantProductRow := []string{strconv.FormatUint(uint64(productID), 10), "Kopi, \"Premium\"\nBubuk", "CSV-1", "899-CSV-1", "Minuman, \"Kopi\"", "500", "1234", "8", "bungkus", "false"}
	if got := productsBySKU["CSV-1"]; strings.Join(got, "|") != strings.Join(wantProductRow, "|") {
		t.Fatalf("primary product CSV row: want %#v, got %#v", wantProductRow, got)
	}
	if got := productsBySKU["CSV-2"]; len(got) != len(wantProductHeader) || got[4] != "" || got[9] != "true" {
		t.Fatalf("uncategorized product CSV row: %#v", got)
	}
	if _, leaked := productsBySKU["OTHER-1"]; leaked {
		t.Fatalf("product from another store leaked into export: %#v", productsBySKU["OTHER-1"])
	}
	bulkProducts := make([]catalog.Product, 201)
	for i := range bulkProducts {
		bulkProducts[i] = catalog.Product{
			BaseModel: database.BaseModel{CreatedAt: bulkCreatedAt, UpdatedAt: bulkCreatedAt},
			StoreID:   exportStore.ID,
			Name:      fmt.Sprintf("Produk %03d", i),
			SKU:       fmt.Sprintf("BULK-%03d", i),
			BuyPrice:  100,
			SellPrice: 200,
			Unit:      "pcs",
			Active:    true,
		}
	}
	if err := testDB.CreateInBatches(&bulkProducts, 50).Error; err != nil {
		t.Fatalf("seed unpaginated product rows: %v", err)
	}
	w = doReq(http.MethodGet, "/products/export", nil, adminToken)
	assertStatus(w, http.StatusOK)
	productRows, err = csv.NewReader(strings.NewReader(w.Body.String())).ReadAll()
	if err != nil || len(productRows) != len(bulkProducts)+3 {
		t.Fatalf("product export should include more than one page of store rows: count=%d err=%v", len(productRows), err)
	}
	localMidnight := time.Date(2026, 10, 14, 0, 0, 0, 0, loc)
	moveTransaction := func(createdAt time.Time) {
		t.Helper()
		if err := testDB.Model(&sales.Trx{}).Where("id = ?", trxID).Update("created_at", createdAt.UTC()).Error; err != nil {
			t.Fatalf("set transaction local boundary: %v", err)
		}
	}
	rowsForDate := func(date string) [][]string {
		t.Helper()
		w := doReq(http.MethodGet, "/transactions/export?period=today&date="+date, nil, adminToken)
		assertStatus(w, http.StatusOK)
		rows, err := csv.NewReader(strings.NewReader(w.Body.String())).ReadAll()
		if err != nil {
			t.Fatalf("parse local-date CSV: %v", err)
		}
		return rows
	}
	moveTransaction(localMidnight.Add(-time.Nanosecond))
	if rows := rowsForDate(localMidnight.Format("2006-01-02")); len(rows) != 1 {
		t.Fatalf("transaction immediately before local midnight must be excluded: rows=%#v", rows)
	}
	moveTransaction(localMidnight)
	if rows := rowsForDate(localMidnight.Format("2006-01-02")); len(rows) != 2 || rows[1][0] != strconv.FormatUint(uint64(trxID), 10) {
		t.Fatalf("transaction at local midnight must be included: rows=%#v", rows)
	}
	nextLocalMidnight := localMidnight.AddDate(0, 0, 1)
	moveTransaction(nextLocalMidnight)
	if rows := rowsForDate(localMidnight.Format("2006-01-02")); len(rows) != 1 {
		t.Fatalf("transaction at next local midnight must be excluded: rows=%#v", rows)
	}
	if rows := rowsForDate(nextLocalMidnight.Format("2006-01-02")); len(rows) != 2 || rows[1][0] != strconv.FormatUint(uint64(trxID), 10) {
		t.Fatalf("transaction at next local midnight must be included on the next date: rows=%#v", rows)
	}

	w = doReq(http.MethodGet, "/transactions/export?period=invalid", nil, adminToken)
	assertStatus(w, http.StatusBadRequest)
	if responseMap(w)["error"] == nil {
		t.Fatalf("invalid-period response has no error field: %s", w.Body.String())
	}

	w = doReq(http.MethodPost, "/users", map[string]any{"name": "Kasir CSV"}, adminToken)
	assertStatus(w, http.StatusCreated)
	cashierID := uint(responseMap(w)["user"].(map[string]any)["id"].(float64))
	w = doReq(http.MethodPut, fmt.Sprintf("/users/%d/passcode", cashierID), map[string]any{"passcode": "12345"}, adminToken)
	assertStatus(w, http.StatusOK)
	w = doReq(http.MethodPost, "/auth/switch", map[string]any{"target_user_id": cashierID, "passcode": "12345", "role": "cashier"}, adminToken)
	assertStatus(w, http.StatusOK)
	cashierToken := responseMap(w)["access_token"].(string)
	for _, path := range []string{"/transactions/export?period=all", "/products/export"} {
		w = doReq(http.MethodGet, path, nil, cashierToken)
		if w.Code != http.StatusForbidden {
			t.Fatalf("cashier GET %s: want 403, got %d: %s", path, w.Code, w.Body.String())
		}
	}
}
