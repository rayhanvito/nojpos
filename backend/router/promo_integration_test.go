package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/adrr-dev/openPOS/backend/service"
)

type promoTestAPI struct {
	t       *testing.T
	handler http.Handler
	remote  string
}

var promoTestIPSequence atomic.Uint32

func newPromoTestAPI(t *testing.T) *promoTestAPI {
	t.Helper()
	t.Setenv("DB_DRIVER", "sqlite")
	t.Setenv("DATABASE_URL", filepath.Join(t.TempDir(), "promos.db"))
	t.Setenv("JWT_SECRET", "test-secret-promos")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	t.Cleanup(cancel)
	srv, err := New(ctx)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if srv.Cleanup != nil {
		t.Cleanup(srv.Cleanup)
	}
	sequence := promoTestIPSequence.Add(1)
	return &promoTestAPI{
		t: t, handler: srv.Handler,
		// Rate limiter keys include client IP while each test has its own database.
		remote: fmt.Sprintf("198.18.%d.%d:1234", sequence/254+1, sequence%254+1),
	}
}

// doRaw = doReq tanpa bantuan fixture apa pun (dipakai tes gerbang shift).
func (api *promoTestAPI) doRaw(method, path string, body any, token string) (*httptest.ResponseRecorder, map[string]any) {
	api.t.Helper()
	return api.doReqInner(method, path, body, token)
}

func (api *promoTestAPI) doReq(method, path string, body any, token string) (*httptest.ResponseRecorder, map[string]any) {
	api.t.Helper()
	// Sejak gerbang shift (2026-10-16) checkout mewajibkan shift aktif.
	// Fixture: pastikan pelaku punya shift sebelum POST /transactions —
	// start yang ditolak 409 (sudah ada shift) diabaikan, jadi tes yang
	// mengatur shift sendiri tidak terganggu.
	if method == http.MethodPost && path == "/transactions" && token != "" {
		_, _ = api.doReqInner(http.MethodPost, "/cashier/shift/start", map[string]any{"opening_cash": 100000}, token)
	}
	return api.doReqInner(method, path, body, token)
}

func (api *promoTestAPI) doReqInner(method, path string, body any, token string) (*httptest.ResponseRecorder, map[string]any) {
	api.t.Helper()
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(bodyBytes))
	req.RemoteAddr = api.remote
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	api.handler.ServeHTTP(w, req)
	var response map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	return w, response
}

func (api *promoTestAPI) registerAdmin(storeName string) string {
	api.t.Helper()
	email := fmt.Sprintf("promo-%d@example.com", time.Now().UnixNano())
	var otp string
	previous := service.TestOnOTPSent
	service.TestOnOTPSent = func(to, code string) {
		if to == email {
			otp = code
		}
	}
	api.t.Cleanup(func() { service.TestOnOTPSent = previous })

	w, _ := api.doReq(http.MethodPost, "/auth/otp/send", map[string]any{"email": email}, "")
	if w.Code != http.StatusOK {
		api.t.Fatalf("send OTP: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, "/auth/otp/verify", map[string]any{"email": email, "code": otp}, "")
	if w.Code != http.StatusOK {
		api.t.Fatalf("verify OTP: %d %s", w.Code, w.Body.String())
	}
	w, response := api.doReq(http.MethodPost, "/auth/register", map[string]any{
		"name": "Admin Promo", "email": email, "password": "password123", "storeName": storeName,
	}, "")
	if w.Code != http.StatusCreated {
		api.t.Fatalf("register admin: %d %s", w.Code, w.Body.String())
	}
	return response["access_token"].(string)
}

func (api *promoTestAPI) createCashier(adminToken string) string {
	api.t.Helper()
	w, response := api.doReq(http.MethodPost, "/users", map[string]any{"name": "Kasir Promo"}, adminToken)
	if w.Code != http.StatusCreated {
		api.t.Fatalf("create cashier: %d %s", w.Code, w.Body.String())
	}
	id := uint(response["user"].(map[string]any)["id"].(float64))
	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/users/%d/passcode", id), map[string]any{"passcode": "12345"}, adminToken)
	if w.Code != http.StatusOK {
		api.t.Fatalf("set cashier passcode: %d %s", w.Code, w.Body.String())
	}
	w, response = api.doReq(http.MethodPost, "/auth/switch", map[string]any{"target_user_id": id, "passcode": "12345"}, adminToken)
	if w.Code != http.StatusOK {
		api.t.Fatalf("switch cashier: %d %s", w.Code, w.Body.String())
	}
	return response["access_token"].(string)
}

func validPromoRequest(name, code string) map[string]any {
	return map[string]any{
		"name": name, "type": "amount", "value": 1000, "scope": "all", "voucherCode": code,
		"startsAt":    time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
		"endsAt":      time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		"minPurchase": 5000, "maxDiscount": 0,
	}
}

func TestPromoCRUDIsAdminOnlyAndStoreScoped(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Promo Satu")
	cashierToken := api.createCashier(adminToken)
	otherAdminToken := api.registerAdmin("Toko Promo Dua")

	w, otherCategory := api.doReq(http.MethodPost, "/categories", map[string]any{"name": "Kategori Toko Dua"}, otherAdminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create other-store category: %d %s", w.Code, w.Body.String())
	}
	otherCategoryID := uint(otherCategory["category"].(map[string]any)["id"].(float64))
	w, otherProduct := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Toko Dua", "sku": "PROMO-TOKO-DUA", "sellPrice": 1000, "stock": 5,
	}, otherAdminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create other-store product: %d %s", w.Code, w.Body.String())
	}
	otherProductID := uint(otherProduct["id"].(float64))

	create := validPromoRequest("Promo Hemat", "HEMAT")
	w, created := api.doReq(http.MethodPost, "/promos", create, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create promo: %d %s", w.Code, w.Body.String())
	}
	promoID := uint(created["id"].(float64))
	if created["active"] != true || created["voucher_code"] != "HEMAT" {
		t.Fatalf("created promo fields: %v", created)
	}

	w, _ = api.doReq(http.MethodPost, "/promos", validPromoRequest("Duplikat", "HEMAT"), adminToken)
	if w.Code != http.StatusConflict {
		t.Fatalf("duplicate code in same store: want 409, got %d %s", w.Code, w.Body.String())
	}
	w, secondStorePromo := api.doReq(http.MethodPost, "/promos", validPromoRequest("Promo Toko Dua", "HEMAT"), otherAdminToken)
	if w.Code != http.StatusCreated || secondStorePromo["voucher_code"] != "HEMAT" {
		t.Fatalf("same code in another store: %d %s", w.Code, w.Body.String())
	}

	w, listed := api.doReq(http.MethodGet, "/promos", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("list promos: %d %s", w.Code, w.Body.String())
	}
	items, ok := listed["items"].([]any)
	if !ok || len(items) != 1 || uint(items[0].(map[string]any)["id"].(float64)) != promoID {
		t.Fatalf("store-scoped promo list: %v", listed)
	}
	w, fetched := api.doReq(http.MethodGet, fmt.Sprintf("/promos/%d", promoID), nil, adminToken)
	if w.Code != http.StatusOK || uint(fetched["id"].(float64)) != promoID {
		t.Fatalf("get promo: %d %s", w.Code, w.Body.String())
	}

	update := validPromoRequest("Promo Hemat Revisi", "HEMAT")
	update["value"] = 1500
	w, updated := api.doReq(http.MethodPut, fmt.Sprintf("/promos/%d", promoID), update, adminToken)
	if w.Code != http.StatusOK || updated["name"] != "Promo Hemat Revisi" || updated["value"].(float64) != 1500 {
		t.Fatalf("update promo: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPatch, fmt.Sprintf("/promos/%d/active", promoID), map[string]any{"active": false}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("toggle promo: %d %s", w.Code, w.Body.String())
	}
	w, fetched = api.doReq(http.MethodGet, fmt.Sprintf("/promos/%d", promoID), nil, adminToken)
	if w.Code != http.StatusOK || fetched["active"] != false {
		t.Fatalf("promo active after toggle: %d %s", w.Code, w.Body.String())
	}

	foreignScope := validPromoRequest("Promo Kategori Asing", "ASING")
	foreignScope["scope"] = "category"
	foreignScope["categoryId"] = otherCategoryID
	w, _ = api.doReq(http.MethodPost, "/promos", foreignScope, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("foreign category scope: want 400, got %d %s", w.Code, w.Body.String())
	}
	foreignScope = validPromoRequest("Promo Produk Asing", "ASING-PRODUK")
	foreignScope["scope"] = "product"
	foreignScope["productId"] = otherProductID
	w, _ = api.doReq(http.MethodPost, "/promos", foreignScope, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("foreign product scope: want 400, got %d %s", w.Code, w.Body.String())
	}

	for _, route := range []struct {
		method string
		path   string
		body   any
	}{
		{http.MethodGet, "/promos", nil},
		{http.MethodPost, "/promos", validPromoRequest("Ditolak", "TOLAK")},
		{http.MethodGet, fmt.Sprintf("/promos/%d", promoID), nil},
		{http.MethodPut, fmt.Sprintf("/promos/%d", promoID), update},
		{http.MethodPatch, fmt.Sprintf("/promos/%d/active", promoID), map[string]any{"active": true}},
		{http.MethodDelete, fmt.Sprintf("/promos/%d", promoID), nil},
	} {
		w, _ = api.doReq(route.method, route.path, route.body, cashierToken)
		if w.Code != http.StatusForbidden {
			t.Fatalf("cashier %s %s: want 403, got %d %s", route.method, route.path, w.Code, w.Body.String())
		}
	}

	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/promos/%d", promoID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("delete promo: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, "/promos", validPromoRequest("Kode Dipakai Lagi", "HEMAT"), adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("reuse deleted voucher: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodGet, fmt.Sprintf("/promos/%d", promoID), nil, adminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get deleted promo: want 404, got %d %s", w.Code, w.Body.String())
	}
}

func TestPromoValidationAllowsDecimalPercentAndCrossMidnightHours(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Promo Validasi")

	percent := validPromoRequest("Persen Desimal", "PERSEN-DESIMAL")
	percent["type"] = "percent"
	percent["value"] = 12.5
	w, created := api.doReq(http.MethodPost, "/promos", percent, adminToken)
	if w.Code != http.StatusCreated || created["value"] != 12.5 {
		t.Fatalf("decimal percent should be accepted and preserved: %d %s", w.Code, w.Body.String())
	}

	for _, value := range []float64{0, 100.1} {
		invalid := validPromoRequest("Persen Tidak Sah", "")
		invalid["type"] = "percent"
		invalid["value"] = value
		w, _ = api.doReq(http.MethodPost, "/promos", invalid, adminToken)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("percent %v: want 400, got %d %s", value, w.Code, w.Body.String())
		}
	}

	amountFraction := validPromoRequest("Amount Desimal", "")
	amountFraction["value"] = 12.5
	w, _ = api.doReq(http.MethodPost, "/promos", amountFraction, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("fractional Rupiah amount: want 400, got %d %s", w.Code, w.Body.String())
	}

	incompleteHours := validPromoRequest("Jam Tidak Lengkap", "JAM-TIDAK-LENGKAP")
	incompleteHours["happyStart"] = "23:00"
	w, _ = api.doReq(http.MethodPost, "/promos", incompleteHours, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("incomplete happy hours: want 400, got %d %s", w.Code, w.Body.String())
	}

	crossMidnight := validPromoRequest("Happy Hour Lintas Hari", "HAPPY-LINTAS-HARI")
	crossMidnight["happyStart"] = "22:00"
	crossMidnight["happyEnd"] = "02:00"
	w, accepted := api.doReq(http.MethodPost, "/promos", crossMidnight, adminToken)
	if w.Code != http.StatusCreated || accepted["happy_start"] != "22:00" || accepted["happy_end"] != "02:00" {
		t.Fatalf("cross-midnight happy hours should be accepted: %d %s", w.Code, w.Body.String())
	}
}

func TestPromoPreviewUsesStoreScheduleScopeAndDoesNotWrite(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Preview Promo")
	cashierToken := api.createCashier(adminToken)

	timezone := "Asia/Kathmandu"
	var location *time.Location
	for _, candidate := range []string{"Asia/Kathmandu", "Asia/Kolkata", "UTC"} {
		location, _ = time.LoadLocation(candidate)
		localNow := time.Now().In(location)
		if localNow.Hour()*60+localNow.Minute() <= 1434 {
			timezone = candidate
			break
		}
	}
	w, _ := api.doReq(http.MethodPut, "/settings", map[string]any{"timezone": timezone}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("set store timezone: %d %s", w.Code, w.Body.String())
	}

	w, category := api.doReq(http.MethodPost, "/categories", map[string]any{"name": "Kategori Cakupan"}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create category: %d %s", w.Code, w.Body.String())
	}
	categoryID := uint(category["category"].(map[string]any)["id"].(float64))
	w, covered := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Cakupan", "sku": "PREVIEW-COVERED", "categoryId": categoryID,
		"sellPrice": 1000, "stock": 10,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create covered product: %d %s", w.Code, w.Body.String())
	}
	coveredID := uint(covered["id"].(float64))
	w, uncovered := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Lain", "sku": "PREVIEW-UNCOVERED", "sellPrice": 1000, "stock": 10,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create uncovered product: %d %s", w.Code, w.Body.String())
	}
	uncoveredID := uint(uncovered["id"].(float64))

	voucher := validPromoRequest("Voucher Kategori", "KATEGORI")
	voucher["scope"] = "category"
	voucher["categoryId"] = categoryID
	voucher["value"] = 500
	voucher["minPurchase"] = 1500
	w, _ = api.doReq(http.MethodPost, "/promos", voucher, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create category voucher: %d %s", w.Code, w.Body.String())
	}
	expired := validPromoRequest("Voucher Kedaluwarsa", "KADALUARSA")
	expired["startsAt"] = time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339)
	expired["endsAt"] = time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
	w, _ = api.doReq(http.MethodPost, "/promos", expired, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create expired voucher: %d %s", w.Code, w.Body.String())
	}
	otherAdminToken := api.registerAdmin("Toko Promo Pembanding")
	otherStorePromo := validPromoRequest("Promo Toko Lain", "TOKO-LAIN")
	otherStorePromo["value"], otherStorePromo["minPurchase"] = 750, 0
	w, _ = api.doReq(http.MethodPost, "/promos", otherStorePromo, otherAdminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create other-store promo: %d %s", w.Code, w.Body.String())
	}
	w, foreignProduct := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Toko Lain", "sku": "PREVIEW-FOREIGN", "sellPrice": 1000, "stock": 5,
	}, otherAdminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create other-store product: %d %s", w.Code, w.Body.String())
	}
	foreignProductID := uint(foreignProduct["id"].(float64))

	previewBody := map[string]any{
		"items": []any{
			map[string]any{"productId": coveredID, "qty": 1},
			map[string]any{"productId": uncoveredID, "qty": 1},
		},
		"voucherCode": "KATEGORI",
	}
	w, preview := api.doReq(http.MethodPost, "/promos/preview", previewBody, cashierToken)
	if w.Code != http.StatusOK || preview["voucher_name"] != "Voucher Kategori" || preview["voucher_discount"] != float64(500) || preview["promo_discount"] != float64(0) {
		t.Fatalf("category voucher preview: %d %s", w.Code, w.Body.String())
	}
	for _, code := range []string{"KADALUARSA", "SALAH"} {
		previewBody["voucherCode"] = code
		w, response := api.doReq(http.MethodPost, "/promos/preview", previewBody, cashierToken)
		if w.Code != http.StatusBadRequest || response == nil || response["code"] != "PROMO_INVALID" || response["error"] == nil {
			t.Fatalf("preview for invalid voucher %q: %d %s", code, w.Code, w.Body.String())
		}
	}
	previewBody["voucherCode"] = ""
	w, response := api.doReq(http.MethodPost, "/promos/preview", previewBody, cashierToken)
	if w.Code != http.StatusOK || response != nil || w.Body.String() != "null" {
		t.Fatalf("preview without a voucher and no automatic promo: %d %s", w.Code, w.Body.String())
	}
	previewBody["voucherCode"] = "TOKO-LAIN"
	w, response = api.doReq(http.MethodPost, "/promos/preview", previewBody, cashierToken)
	if w.Code != http.StatusBadRequest || response == nil || response["code"] != "PROMO_INVALID" || response["error"] == nil {
		t.Fatalf("preview must not apply another store's voucher: %d %s", w.Code, w.Body.String())
	}
	w, response = api.doReq(http.MethodPost, "/promos/preview", map[string]any{
		"items": []any{map[string]any{"productId": foreignProductID, "qty": 1}},
	}, cashierToken)
	if w.Code != http.StatusNotFound || response["error"] != "Produk tidak ditemukan di toko Anda." {
		t.Fatalf("preview must reject another store's product clearly: %d %s", w.Code, w.Body.String())
	}

	localNow := time.Now().In(location)
	minute := localNow.Hour()*60 + localNow.Minute()
	insideStart, insideEnd := 1439, minute+5
	if minute > 1433 {
		insideStart, insideEnd = minute-5, minute-10
	}
	happy := validPromoRequest("Promo Happy", "")
	happy["value"] = 400
	happy["minPurchase"] = 0
	happy["happyStart"] = fmt.Sprintf("%02d:%02d", insideStart/60, insideStart%60)
	happy["happyEnd"] = fmt.Sprintf("%02d:%02d", insideEnd/60, insideEnd%60)
	w, createdHappy := api.doReq(http.MethodPost, "/promos", happy, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create cross-midnight promo: %d %s", w.Code, w.Body.String())
	}
	previewBody["voucherCode"] = ""
	w, preview = api.doReq(http.MethodPost, "/promos/preview", previewBody, cashierToken)
	if w.Code != http.StatusOK || preview["promo_name"] != "Promo Happy" || preview["promo_discount"] != float64(400) {
		t.Fatalf("cross-midnight window should include local store time %s: %d %s", localNow.Format("15:04"), w.Code, w.Body.String())
	}

	outsideStart, outsideEnd := minute+5, minute-1
	if outsideEnd < 0 {
		outsideEnd = 0
	}
	happy["happyStart"] = fmt.Sprintf("%02d:%02d", outsideStart/60, outsideStart%60)
	happy["happyEnd"] = fmt.Sprintf("%02d:%02d", outsideEnd/60, outsideEnd%60)
	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/promos/%d", uint(createdHappy["id"].(float64))), happy, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("update cross-midnight window: %d %s", w.Code, w.Body.String())
	}
	w, response = api.doReq(http.MethodPost, "/promos/preview", previewBody, cashierToken)
	if w.Code != http.StatusOK || response != nil || w.Body.String() != "null" {
		t.Fatalf("cross-midnight window should exclude local time %s: %d %s", localNow.Format("15:04"), w.Code, w.Body.String())
	}

	w, transactions := api.doReq(http.MethodGet, "/transactions", nil, adminToken)
	if w.Code != http.StatusOK || len(transactions["items"].([]any)) != 0 {
		t.Fatalf("preview wrote transactions: %d %s", w.Code, w.Body.String())
	}
	w, products := api.doReq(http.MethodGet, "/products", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("read products after preview: %d %s", w.Code, w.Body.String())
	}
	for _, item := range products["items"].([]any) {
		if item.(map[string]any)["stock"] != float64(10) {
			t.Fatalf("preview changed stock: %v", item)
		}
	}
	w, movements := api.doReq(http.MethodGet, "/movements", nil, adminToken)
	if w.Code != http.StatusOK || movements["total"] != float64(2) {
		t.Fatalf("preview wrote stock movements: %d %s", w.Code, w.Body.String())
	}
}

func TestCheckoutAppliesPromoSnapshotAndRejectsInvalidVoucher(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Checkout Promo")
	cashierToken := api.createCashier(adminToken)
	w, _ := api.doReq(http.MethodPut, "/settings", map[string]any{
		"taxEnabled": true, "taxPct": 10, "taxInclusive": false, "taxRounding": "none",
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("enable additive tax: %d %s", w.Code, w.Body.String())
	}
	w, category := api.doReq(http.MethodPost, "/categories", map[string]any{"name": "Kategori Promo Checkout"}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create category: %d %s", w.Code, w.Body.String())
	}
	categoryID := uint(category["category"].(map[string]any)["id"].(float64))
	w, covered := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Kategori", "sku": "CHECKOUT-COVERED", "categoryId": categoryID,
		"sellPrice": 300, "stock": 30,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create covered product: %d %s", w.Code, w.Body.String())
	}
	coveredID := uint(covered["id"].(float64))
	w, uncovered := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Umum", "sku": "CHECKOUT-UNCOVERED", "sellPrice": 1004, "stock": 30,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create uncovered product: %d %s", w.Code, w.Body.String())
	}
	uncoveredID := uint(uncovered["id"].(float64))

	checkout := func(items []any, code string, manualDiscount int64) (*httptest.ResponseRecorder, map[string]any) {
		body := map[string]any{
			"items": items, "discount": manualDiscount, "method": "Cash", "paid": 100000,
		}
		if code != "" {
			body["voucherCode"] = code
		}
		return api.doReq(http.MethodPost, "/transactions", body, cashierToken)
	}
	assertInvalidVoucher := func(code string) {
		t.Helper()
		w, response := checkout([]any{map[string]any{"productId": uncoveredID, "qty": 1}}, code, 0)
		if w.Code != http.StatusBadRequest || response["code"] != "PROMO_INVALID" || response["error"] == nil {
			t.Fatalf("invalid voucher %q: want 400 PROMO_INVALID, got %d %s", code, w.Code, w.Body.String())
		}
	}
	assertInvalidVoucher("TIDAK-ADA")

	expired := validPromoRequest("Voucher Kedaluwarsa Checkout", "KEDALUARSA-CHECKOUT")
	expired["minPurchase"] = 0
	expired["startsAt"] = time.Now().UTC().Add(-48 * time.Hour).Format(time.RFC3339)
	expired["endsAt"] = time.Now().UTC().Add(-24 * time.Hour).Format(time.RFC3339)
	w, _ = api.doReq(http.MethodPost, "/promos", expired, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create expired voucher: %d %s", w.Code, w.Body.String())
	}
	assertInvalidVoucher("KEDALUARSA-CHECKOUT")

	inactive := validPromoRequest("Voucher Nonaktif", "NONAKTIF")
	inactive["minPurchase"] = 0
	w, inactivePromo := api.doReq(http.MethodPost, "/promos", inactive, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create inactive voucher: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPatch, fmt.Sprintf("/promos/%d/active", uint(inactivePromo["id"].(float64))), map[string]any{"active": false}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("disable voucher: %d %s", w.Code, w.Body.String())
	}
	assertInvalidVoucher("NONAKTIF")

	outsideWindow := validPromoRequest("Voucher Di Luar Jam", "DI-LUAR-JAM")
	outsideWindow["minPurchase"] = 0
	storeLocation, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		t.Fatalf("load store timezone: %v", err)
	}
	localNow := time.Now().In(storeLocation)
	localMinute := localNow.Hour()*60 + localNow.Minute()
	outsideStart := 0
	if localMinute < 6 || localMinute >= 1434 {
		outsideStart = 12 * 60
	}
	outsideWindow["happyStart"] = fmt.Sprintf("%02d:%02d", outsideStart/60, outsideStart%60)
	outsideWindow["happyEnd"] = fmt.Sprintf("%02d:%02d", (outsideStart+1)/60, (outsideStart+1)%60)
	w, _ = api.doReq(http.MethodPost, "/promos", outsideWindow, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create out-of-window voucher: %d %s", w.Code, w.Body.String())
	}
	assertInvalidVoucher("DI-LUAR-JAM")

	minimum := validPromoRequest("Voucher Minimum", "MINIMUM")
	minimum["minPurchase"] = 2000
	w, _ = api.doReq(http.MethodPost, "/promos", minimum, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create minimum voucher: %d %s", w.Code, w.Body.String())
	}
	assertInvalidVoucher("MINIMUM")

	percent := validPromoRequest("Persen Dua Belas Setengah", "DESIMAL")
	percent["type"] = "percent"
	percent["value"] = 12.5
	percent["scope"] = "product"
	percent["productId"] = uncoveredID
	percent["minPurchase"] = 0
	w, decimalPromo := api.doReq(http.MethodPost, "/promos", percent, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create decimal percent voucher: %d %s", w.Code, w.Body.String())
	}
	w, sale := checkout([]any{map[string]any{"productId": uncoveredID, "qty": 1}}, "DESIMAL", 900)
	if w.Code != http.StatusCreated {
		t.Fatalf("checkout with decimal voucher: %d %s", w.Code, w.Body.String())
	}
	if sale["subtotal"] != float64(1004) || sale["discount"] != float64(0) || sale["promo_discount"] != float64(0) || sale["voucher_discount"] != float64(126) || sale["tax"] != float64(88) || sale["total"] != float64(966) || sale["voucher_code"] != "DESIMAL" || sale["voucher_id"] != decimalPromo["id"] {
		t.Fatalf("checkout pricing order, half-up rounding, manual discount, or snapshot: %v", sale)
	}
	trxID := uint(sale["id"].(float64))
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/promos/%d", uint(decimalPromo["id"].(float64))), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("delete promo after checkout: %d %s", w.Code, w.Body.String())
	}
	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/transactions/%d", trxID), nil, adminToken)
	if w.Code != http.StatusOK || detail["voucher_code"] != "DESIMAL" || detail["voucher_discount"] != float64(126) || detail["voucher_id"] != decimalPromo["id"] {
		t.Fatalf("transaction promo snapshot after promo deletion: %d %s", w.Code, w.Body.String())
	}

	w, halfRupiahProduct := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Batas Pembulatan", "sku": "PROMO-ROUND-HALF-UP", "sellPrice": 5000, "stock": 5,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create rounding test product: %d %s", w.Code, w.Body.String())
	}
	halfRupiahProductID := uint(halfRupiahProduct["id"].(float64))
	halfUp := validPromoRequest("Persen Batas Setengah", "SETENGAH")
	halfUp["type"], halfUp["value"], halfUp["minPurchase"] = "percent", 0.57, 0
	w, _ = api.doReq(http.MethodPost, "/promos", halfUp, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create half-up boundary promo: %d %s", w.Code, w.Body.String())
	}
	w, preview := api.doReq(http.MethodPost, "/promos/preview", map[string]any{
		"items": []any{map[string]any{"productId": halfRupiahProductID, "qty": 1}}, "voucherCode": "SETENGAH",
	}, cashierToken)
	if w.Code != http.StatusOK || preview["voucher_discount"] != float64(29) {
		t.Fatalf("Rp5.000 x 0,57%% half-up preview: %d %s", w.Code, w.Body.String())
	}

	autoPercent := validPromoRequest("Auto Persen", "")
	autoPercent["type"], autoPercent["value"], autoPercent["minPurchase"] = "percent", 20.0, 0
	w, _ = api.doReq(http.MethodPost, "/promos", autoPercent, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create automatic percent promo: %d %s", w.Code, w.Body.String())
	}
	autoAmount := validPromoRequest("Auto Nominal Terbesar", "")
	autoAmount["value"], autoAmount["minPurchase"] = 400, 0
	w, _ = api.doReq(http.MethodPost, "/promos", autoAmount, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create automatic amount promo: %d %s", w.Code, w.Body.String())
	}
	w, sale = checkout([]any{map[string]any{"productId": uncoveredID, "qty": 1}}, "", 50)
	if w.Code != http.StatusCreated || sale["promo_name"] != "Auto Nominal Terbesar" || sale["promo_discount"] != float64(400) || sale["discount"] != float64(0) {
		t.Fatalf("automatic promo winner and manual discount: %d %s", w.Code, w.Body.String())
	}

	capPromo := validPromoRequest("Voucher Dengan Batas", "BATAS")
	capPromo["type"], capPromo["value"], capPromo["maxDiscount"], capPromo["minPurchase"] = "percent", 90.0, 500, 0
	w, _ = api.doReq(http.MethodPost, "/promos", capPromo, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create capped percent voucher: %d %s", w.Code, w.Body.String())
	}
	w, sale = checkout([]any{map[string]any{"productId": uncoveredID, "qty": 1}}, "BATAS", 0)
	if w.Code != http.StatusCreated || sale["voucher_discount"] != float64(500) || sale["promo_discount"] != float64(0) {
		t.Fatalf("max discount cap: %d %s", w.Code, w.Body.String())
	}

	categoryPromo := validPromoRequest("Voucher Kategori", "KATEGORI-CHECKOUT")
	categoryPromo["scope"], categoryPromo["categoryId"] = "category", categoryID
	categoryPromo["value"], categoryPromo["minPurchase"] = 1000, 1304
	w, _ = api.doReq(http.MethodPost, "/promos", categoryPromo, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create category amount voucher: %d %s", w.Code, w.Body.String())
	}
	w, sale = checkout([]any{
		map[string]any{"productId": coveredID, "qty": 1},
		map[string]any{"productId": uncoveredID, "qty": 1},
	}, "KATEGORI-CHECKOUT", 0)
	if w.Code != http.StatusCreated || sale["subtotal"] != float64(1304) || sale["voucher_discount"] != float64(300) || sale["promo_discount"] != float64(0) {
		t.Fatalf("category coverage, amount cap, or pre-promo minimum: %d %s", w.Code, w.Body.String())
	}

}

func TestCheckoutStacksCombinableVoucherOnAutoPromo(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Tumpuk Promo")
	cashierToken := api.createCashier(adminToken)
	w, product := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Tumpuk", "sku": "TUMPUK-1", "sellPrice": 100000, "stock": 10,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create product: %d %s", w.Code, w.Body.String())
	}
	productID := uint(product["id"].(float64))

	auto := validPromoRequest("Auto Sepuluh", "")
	auto["type"], auto["value"], auto["minPurchase"] = "percent", 10.0, 0
	w, _ = api.doReq(http.MethodPost, "/promos", auto, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create automatic promo: %d %s", w.Code, w.Body.String())
	}

	stacked := validPromoRequest("Voucher Tumpuk", "TUMPUK10")
	stacked["type"], stacked["value"], stacked["minPurchase"] = "percent", 10.0, 0
	stacked["combinable"] = true
	w, _ = api.doReq(http.MethodPost, "/promos", stacked, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create combinable voucher: %d %s", w.Code, w.Body.String())
	}

	single := validPromoRequest("Voucher Tunggal", "TUNGGAL10")
	single["type"], single["value"], single["minPurchase"] = "percent", 10.0, 0
	w, _ = api.doReq(http.MethodPost, "/promos", single, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create single voucher: %d %s", w.Code, w.Body.String())
	}

	checkout := func(code string) map[string]any {
		t.Helper()
		body := map[string]any{
			"items": []any{map[string]any{"productId": productID, "qty": 1}},
			"method": "Cash", "paid": 1000000,
		}
		if code != "" {
			body["voucherCode"] = code
		}
		w, sale := api.doReq(http.MethodPost, "/transactions", body, cashierToken)
		if w.Code != http.StatusCreated {
			t.Fatalf("checkout %q: %d %s", code, w.Code, w.Body.String())
		}
		return sale
	}

	// Menumpuk: promo 10% x 100.000 = 10.000; voucher 10% dari sisa 90.000 = 9.000.
	sale := checkout("TUMPUK10")
	if sale["promo_discount"] != float64(10000) || sale["promo_name"] != "Auto Sepuluh" || sale["voucher_discount"] != float64(9000) || sale["voucher_code"] != "TUMPUK10" || sale["total"] != float64(81000) {
		t.Fatalf("stacked checkout: %v", sale)
	}
	// Tidak combinable: voucher menggantikan promo otomatis.
	sale = checkout("TUNGGAL10")
	if sale["promo_discount"] != float64(0) || sale["voucher_discount"] != float64(10000) || sale["total"] != float64(90000) {
		t.Fatalf("non-combinable checkout: %v", sale)
	}
	// Tanpa kode: hanya promo otomatis.
	sale = checkout("")
	if sale["promo_discount"] != float64(10000) || sale["voucher_discount"] != float64(0) || sale["total"] != float64(90000) {
		t.Fatalf("automatic-only checkout: %v", sale)
	}
}

func TestPromoAmountRejectsValuesOutsideInt64(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Promo Nominal")
	amount := validPromoRequest("Nominal Terlalu Besar", "NOMINAL-LUAR-RENTANG")
	amount["value"] = json.Number("10000000000000000000")
	w, _ := api.doReq(http.MethodPost, "/promos", amount, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("amount outside int64 range: want 400, got %d %s", w.Code, w.Body.String())
	}

	exactAmount := validPromoRequest("Nominal Presisi Int64", "NOMINAL-PRESISI")
	exactAmount["value"] = json.Number("9007199254740993")
	w, created := api.doReq(http.MethodPost, "/promos", exactAmount, adminToken)
	if w.Code != http.StatusCreated || !bytes.Contains(w.Body.Bytes(), []byte(`"value":9007199254740993`)) {
		t.Fatalf("exact amount response: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodGet, fmt.Sprintf("/promos/%d", uint(created["id"].(float64))), nil, adminToken)
	if w.Code != http.StatusOK || !bytes.Contains(w.Body.Bytes(), []byte(`"value":9007199254740993`)) {
		t.Fatalf("exact amount after database read: %d %s", w.Code, w.Body.String())
	}
}
