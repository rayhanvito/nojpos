package router

import (
	"net/http"
	"testing"
)

// Regresi audit menyeluruh 2026-10-16 (perintah Rayhan: tidak ada lagi
// kelolosan): gerbang shift di checkout, larangan buka shift ganda,
// pengeluaran tanpa metode Credit, dan penjualan offline-shortage yang
// terhitung sebagai penjualan + bisa direfund.
func TestAuditFixesShiftGateAndAggregates(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Audit Gate")
	cashierToken := api.createCashier(adminToken)
	productID := createInventoryProduct(t, api, adminToken, "Produk Gate", "GATE-1", 10000, 50, 0)

	checkout := map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "Cash", "paid": 10000,
	}

	// 1. Tanpa shift aktif, checkout owner DITOLAK 409 SHIFT_REQUIRED.
	w, body := api.doRaw(http.MethodPost, "/transactions", checkout, adminToken)
	if w.Code != http.StatusConflict || body["code"] != "SHIFT_REQUIRED" {
		t.Fatalf("owner checkout without shift: %d %v", w.Code, body)
	}
	// 2. Tanpa shift aktif, checkout kasir DITOLAK juga.
	w, _ = api.doRaw(http.MethodPost, "/transactions", checkout, cashierToken)
	if w.Code != http.StatusConflict {
		t.Fatalf("cashier checkout without shift: %d %s", w.Code, w.Body.String())
	}
	// 3. Buka shift -> 201; buka lagi -> 409 (tidak menutup diam-diam).
	w, _ = api.doRaw(http.MethodPost, "/cashier/shift/start", map[string]any{"opening_cash": 5000}, cashierToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("start shift: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doRaw(http.MethodPost, "/cashier/shift/start", map[string]any{"opening_cash": 5000}, cashierToken)
	if w.Code != http.StatusConflict {
		t.Fatalf("double start shift must be 409: %d %s", w.Code, w.Body.String())
	}
	// 4. Dengan shift aktif, checkout kasir lolos.
	w, _ = api.doRaw(http.MethodPost, "/transactions", checkout, cashierToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("cashier checkout with shift: %d %s", w.Code, w.Body.String())
	}
	// 5. Pengeluaran metode Credit ditolak (hanya metode standar).
	w, _ = api.doRaw(http.MethodPost, "/expenses", map[string]any{
		"date": "2026-10-16", "category": "Lainnya", "note": "uji", "amount": 1000, "method": "Credit", "source": "outside",
	}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expense with Credit method must be 400: %d %s", w.Code, w.Body.String())
	}

	// 6. Penjualan offline-shortage terhitung di dashboard + bisa direfund.
	w, _ = api.doRaw(http.MethodPost, "/cashier/shift/start", map[string]any{"opening_cash": 5000}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("start admin shift: %d %s", w.Code, w.Body.String())
	}
	scarceID := createInventoryProduct(t, api, adminToken, "Produk Langka Gate", "GATE-2", 2000, 2, 0)
	w, sale := api.doRaw(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": scarceID, "qty": 5}}, "method": "Cash", "paid": 10000, "clientTxId": "audit-gate-offline-1",
	}, adminToken)
	if w.Code != http.StatusCreated || sale["status"] != "completed_offline_shortage" {
		t.Fatalf("offline shortage sale: %d %v", w.Code, sale)
	}
	saleID := uint(sale["id"].(float64))
	w, dash := api.doRaw(http.MethodGet, "/dashboard", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("dashboard: %d %s", w.Code, w.Body.String())
	}
	today := dash["today"].(map[string]any)
	// Omzet hari ini mencakup penjualan kasir (10000) + shortage (10000).
	if today["omzet"].(float64) != 20000 {
		t.Fatalf("dashboard omzet must include shortage sale: %v", today)
	}
	w, _ = api.doRaw(http.MethodPost, "/transactions/"+itoa(sale["id"].(float64))+"/refund", map[string]any{
		"items": []any{map[string]any{"productId": scarceID, "qty": 1}}, "reason": "uji refund shortage",
	}, adminToken)
	_ = saleID
	if w.Code != http.StatusOK {
		t.Fatalf("refund of shortage sale must be 200: %d %s", w.Code, w.Body.String())
	}
}
