package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/webhook"
)

// URL uji memakai IP literal publik agar validasi tidak bergantung DNS.
const webhookPublicURL = "https://93.184.216.34/hook"

func TestWebhooksCRUDValidationAndAPIKeyRestriction(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Webhook")
	otherAdminToken := api.registerAdmin("Toko Webhook Lain")

	// Validasi: URL privat/loopback ditolak 400 (pertahanan SSRF).
	for _, badURL := range []string{"http://127.0.0.1:9999/hook", "http://192.168.10.9/hook", "http://10.1.2.3/hook"} {
		w, _ := api.doReq(http.MethodPost, "/webhooks", map[string]any{
			"url": badURL, "events": []string{"transaction.created"},
		}, adminToken)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("url privat %s: mau 400, dapat %d, body=%s", badURL, w.Code, w.Body.String())
		}
	}
	// Validasi: events kosong dan event tak dikenal ditolak 400.
	w, _ := api.doReq(http.MethodPost, "/webhooks", map[string]any{
		"url": webhookPublicURL, "events": []string{},
	}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("events kosong: mau 400, dapat %d, body=%s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, "/webhooks", map[string]any{
		"url": webhookPublicURL, "events": []string{"stok.berubah"},
	}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("event asing: mau 400, dapat %d, body=%s", w.Code, w.Body.String())
	}

	// Buat: 201 dan secret hanya muncul di respons ini.
	w, created := api.doReq(http.MethodPost, "/webhooks", map[string]any{
		"url": webhookPublicURL, "events": []string{"transaction.created", "stock.updated"},
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("buat webhook: mau 201, dapat %d, body=%s", w.Code, w.Body.String())
	}
	secret, ok := created["secret"].(string)
	if !ok || !strings.HasPrefix(secret, "whsec_") || len(secret) != len("whsec_")+40 {
		t.Fatalf("secret pembuatan tidak sah: %v", created["secret"])
	}
	hookID := uint(created["id"].(float64))
	if created["url"] != webhookPublicURL || len(created["events"].([]any)) != 2 {
		t.Fatalf("metadata pembuatan tidak sesuai: %s", w.Body.String())
	}

	// Daftar: metadata saja — secret tidak boleh muncul dalam bentuk apa pun.
	w, listed := api.doReq(http.MethodGet, "/webhooks", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("daftar webhook: mau 200, dapat %d, body=%s", w.Code, w.Body.String())
	}
	items := listed["items"].([]any)
	if len(items) != 1 || uint(items[0].(map[string]any)["id"].(float64)) != hookID {
		t.Fatalf("daftar webhook tidak sesuai: %s", w.Body.String())
	}
	if strings.Contains(w.Body.String(), secret) || strings.Contains(w.Body.String(), `"secret"`) {
		t.Fatalf("daftar webhook membocorkan secret: %s", w.Body.String())
	}

	// Hapus milik toko lain → 404.
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/webhooks/%d", hookID), nil, otherAdminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("hapus webhook toko lain: mau 404, dapat %d, body=%s", w.Code, w.Body.String())
	}

	// Kunci API tidak boleh mengelola webhook sama sekali → 403.
	w, keyResp := api.doReq(http.MethodPost, "/api-keys", map[string]any{"name": "Uji Webhook"}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("buat API key: mau 201, dapat %d, body=%s", w.Code, w.Body.String())
	}
	apiKey := keyResp["apiKey"].(string)
	doKeyReq := func(method, path string, body any) *httptest.ResponseRecorder {
		t.Helper()
		var bodyBytes []byte
		if body != nil {
			bodyBytes, _ = json.Marshal(body)
		}
		req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(bodyBytes))
		req.RemoteAddr = api.remote
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", apiKey)
		rec := httptest.NewRecorder()
		api.handler.ServeHTTP(rec, req)
		return rec
	}
	for _, tc := range []struct{ method, path string }{
		{http.MethodGet, "/webhooks"},
		{http.MethodPost, "/webhooks"},
		{http.MethodDelete, fmt.Sprintf("/webhooks/%d", hookID)},
	} {
		rec := doKeyReq(tc.method, tc.path, map[string]any{"url": webhookPublicURL, "events": []string{"transaction.created"}})
		if rec.Code != http.StatusForbidden {
			t.Fatalf("API key %s %s: mau 403, dapat %d, body=%s", tc.method, tc.path, rec.Code, rec.Body.String())
		}
	}

	// Hapus milik sendiri → 200, daftar kosong sesudahnya.
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/webhooks/%d", hookID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("hapus webhook: mau 200, dapat %d, body=%s", w.Code, w.Body.String())
	}
	w, listed = api.doReq(http.MethodGet, "/webhooks", nil, adminToken)
	if w.Code != http.StatusOK || len(listed["items"].([]any)) != 0 {
		t.Fatalf("daftar sesudah hapus harus kosong: %d %s", w.Code, w.Body.String())
	}
}

func TestWebhookDeadReceiverDoesNotFailCheckout(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Webhook Penerima Mati")

	db, err := gorm.Open(sqlite.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	var storeRow struct{ ID uint }
	if err := db.Table("stores").Order("id").First(&storeRow).Error; err != nil {
		t.Fatal(err)
	}

	// Webhook ditanam langsung ke DB (menembus validasi API memang
	// disengaja di sini): penerimanya port tertutup di loopback —
	// setiap pengiriman pasti gagal jaringan.
	repo := webhook.NewRepo(db)
	hook := &webhook.Webhook{
		StoreID: storeRow.ID, URL: "http://127.0.0.1:1/dead",
		Events: `["transaction.created"]`, Secret: "whsec_deadtest",
	}
	if err := repo.Create(context.Background(), hook); err != nil {
		t.Fatal(err)
	}

	productID := createInventoryProduct(t, api, adminToken, "Kopi Webhook", "WH-KOPI", 5000, 10, 0)
	w, trx := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items":  []any{map[string]any{"productId": productID, "qty": 1}},
		"method": "QRIS",
	}, adminToken)
	if w.Code != http.StatusCreated || trx["status"] != "completed" {
		t.Fatalf("checkout dengan penerima webhook mati harus tetap sukses: %d %s", w.Code, w.Body.String())
	}

	// Percobaan pengiriman tercatat gagal di webhook_deliveries.
	deadline := time.Now().Add(10 * time.Second)
	var rows []webhook.Delivery
	for time.Now().Before(deadline) {
		rows = nil
		if err := db.Where("webhook_id = ?", hook.ID).Find(&rows).Error; err == nil && len(rows) > 0 && rows[0].Error != "" {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if len(rows) == 0 {
		t.Fatal("tidak ada baris delivery tercatat untuk webhook penerima mati")
	}
	row := rows[0]
	if row.Success {
		t.Fatal("delivery ke penerima mati tercatat sukses")
	}
	if row.ResponseStatus != nil {
		t.Fatalf("response_status harus NULL untuk gagal jaringan, dapat %d", *row.ResponseStatus)
	}
	if !strings.Contains(row.Payload, `"event":"transaction.created"`) {
		t.Fatalf("payload delivery tidak memuat event: %s", row.Payload)
	}

	// Bisnis tetap sehat sesudahnya.
	w, list := api.doReq(http.MethodGet, "/transactions", nil, adminToken)
	if w.Code != http.StatusOK || len(list["items"].([]any)) != 1 {
		t.Fatalf("daftar transaksi sesudah checkout: %d %s", w.Code, w.Body.String())
	}
}
