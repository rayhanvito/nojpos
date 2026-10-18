package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAPIKeysCRUDAuthAndRestrictions(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko API Key")
	otherAdminToken := api.registerAdmin("Toko API Key Lain")

	var doKeyReq = func(method, path string, body any, key string) (*httptest.ResponseRecorder, map[string]any) {
		api.t.Helper()
		var bodyBytes []byte
		if body != nil {
			bodyBytes, _ = json.Marshal(body)
		}
		req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(bodyBytes))
		req.RemoteAddr = api.remote
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", key)
		w := httptest.NewRecorder()
		api.handler.ServeHTTP(w, req)
		var response map[string]any
		_ = json.Unmarshal(w.Body.Bytes(), &response)
		return w, response
	}

	w, _ := api.doReq(http.MethodPost, "/api-keys", map[string]any{"name": "   "}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("nama kosong: mau 400, dapat %d, body=%s", w.Code, w.Body.String())
	}

	w, created := api.doReq(http.MethodPost, "/api-keys", map[string]any{"name": "Integrasi Akuntansi"}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("buat API key: mau 201, dapat %d, body=%s", w.Code, w.Body.String())
	}
	key, ok := created["apiKey"].(string)
	if !ok || len(key) != 44 || !strings.HasPrefix(key, "njp_") {
		t.Fatalf("format apiKey tidak valid: %v", created["apiKey"])
	}
	if created["prefix"] != key[:11] {
		t.Fatalf("prefix = %v, mau %q", created["prefix"], key[:11])
	}
	keyID := uint(created["id"].(float64))

	w, listed := api.doReq(http.MethodGet, "/api-keys", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("daftar API key: mau 200, dapat %d, body=%s", w.Code, w.Body.String())
	}
	items, ok := listed["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("items daftar API key tidak sesuai: %s", w.Body.String())
	}
	item := items[0].(map[string]any)
	for _, field := range []string{"id", "name", "prefix", "created_at", "last_used_at"} {
		if _, ok := item[field]; !ok {
			t.Errorf("metadata daftar tidak memuat %q: %s", field, w.Body.String())
		}
	}
	for _, field := range []string{"apiKey", "key_hash", "key_prefix"} {
		if _, ok := item[field]; ok {
			t.Errorf("daftar API key membocorkan %q", field)
		}
	}
	if strings.Contains(w.Body.String(), key) {
		t.Fatal("daftar membocorkan kunci penuh")
	}

	createProduct := func(token, name, sku string) {
		t.Helper()
		w, _ := api.doReq(http.MethodPost, "/products", map[string]any{
			"name": name, "sku": sku, "buyPrice": 1000, "sellPrice": 2000, "stock": 5,
		}, token)
		if w.Code != http.StatusCreated {
			t.Fatalf("buat produk %q: mau 201, dapat %d, body=%s", name, w.Code, w.Body.String())
		}
	}
	createProduct(adminToken, "Produk Toko Pemilik", "APIKEY-A")
	createProduct(otherAdminToken, "Produk Toko Lain", "APIKEY-B")

	w, products := doKeyReq(http.MethodGet, "/products", nil, key)
	if w.Code != http.StatusOK {
		t.Fatalf("auth API key ke produk: mau 200, dapat %d, body=%s", w.Code, w.Body.String())
	}
	productItems := products["items"].([]any)
	if len(productItems) != 1 || productItems[0].(map[string]any)["name"] != "Produk Toko Pemilik" {
		t.Fatalf("produk tidak ter-scope ke toko pemilik key: %s", w.Body.String())
	}

	w, listed = api.doReq(http.MethodGet, "/api-keys", nil, adminToken)
	if w.Code != http.StatusOK || listed["items"].([]any)[0].(map[string]any)["last_used_at"] == nil {
		t.Fatalf("last_used_at belum diperbarui setelah auth sukses: %d %s", w.Code, w.Body.String())
	}

	w, _ = doKeyReq(http.MethodGet, "/products", nil, "njp_0000000000000000000000000000000000000000")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("key salah: mau 401, dapat %d, body=%s", w.Code, w.Body.String())
	}

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api-keys"},
		{http.MethodPost, "/api-keys"},
		{http.MethodDelete, fmt.Sprintf("/api-keys/%d", keyID)},
		{http.MethodGet, "/users"},
		{http.MethodPost, "/users"},
		{http.MethodPost, "/outlets"},
		{http.MethodPut, "/outlets/1"},
		{http.MethodPut, "/outlets/1/settings"},
		{http.MethodPut, "/settings"},
		{http.MethodGet, "/auth/me"},
		{http.MethodPost, "/auth/switch"},
		{http.MethodPut, "/auth/password"},
		{http.MethodPost, "/presence/heartbeat"},
		{http.MethodPost, "/outlets/switch"},
	} {
		w, _ := doKeyReq(tc.method, tc.path, map[string]any{}, key)
		if w.Code != http.StatusForbidden {
			t.Errorf("API key ke %s %s: mau 403, dapat %d, body=%s", tc.method, tc.path, w.Code, w.Body.String())
		}
	}

	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/api-keys/%d", keyID), nil, otherAdminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("hapus key toko lain: mau 404, dapat %d, body=%s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/api-keys/%d", keyID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("cabut API key: mau 200, dapat %d, body=%s", w.Code, w.Body.String())
	}
	w, _ = doKeyReq(http.MethodGet, "/products", nil, key)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("key dicabut: mau 401, dapat %d, body=%s", w.Code, w.Body.String())
	}
}
