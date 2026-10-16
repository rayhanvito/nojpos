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

// Kebijakan Rayhan 2026-10-16: metode pembayaran hanya Cash, QRIS,
// Bank Transfer, dan Credit. E-Wallet + Card ditolak di semua pintu
// (checkout, pengeluaran, pembayaran hutang) tetapi data historisnya
// tetap terbaca.
func TestPaymentMethodsRestrictedToFour(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Metode Bayar")
	cashierToken := api.createCashier(adminToken)
	productID := createInventoryProduct(t, api, adminToken, "Produk Metode", "MET-1", 10000, 50, 0)

	w, _ := api.doRaw(http.MethodPost, "/cashier/shift/start", map[string]any{"opening_cash": 5000}, cashierToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("start shift: %d %s", w.Code, w.Body.String())
	}
	for _, method := range []string{"E-Wallet", "Card"} {
		w, _ := api.doRaw(http.MethodPost, "/transactions", map[string]any{
			"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": method, "paid": 10000,
		}, cashierToken)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("checkout %s must be 400: %d %s", method, w.Code, w.Body.String())
		}
	}
	w, _ = api.doRaw(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "QRIS", "paid": 10000,
	}, cashierToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("checkout QRIS must pass: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doRaw(http.MethodPost, "/expenses", map[string]any{
		"date": "2026-10-16", "category": "Lainnya", "note": "uji", "amount": 1000, "method": "E-Wallet", "source": "outside",
	}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expense with E-Wallet must be 400: %d %s", w.Code, w.Body.String())
	}
}

// Pembuktian kelompok 1 + perbaikan kelompok 2 (2026-10-16): rentang
// tanggal terbalik kini ditolak 400 (bukan diam-diam diganti periode
// bawaan), dan notifikasi stok rendah juga dinilai saat produk dibuat.
func TestReversedRangeRejectedAndLowStockOnCreate(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Rentang dan Stok")

	w, _ := api.doRaw(http.MethodGet, "/dashboard?from=2026-10-20&to=2026-10-10", nil, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("dashboard reversed range must be 400: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doRaw(http.MethodGet, "/reports?from=2026-10-20&to=2026-10-10", nil, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("reports reversed range must be 400: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doRaw(http.MethodGet, "/reports?from=2026-10-10&to=2026-10-20", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("reports normal range must pass: %d %s", w.Code, w.Body.String())
	}

	// Produk yang sejak dibuat sudah di bawah stok minimum -> notifikasi.
	_ = createInventoryProduct(t, api, adminToken, "Produk Langsung Menipis", "LOW-1", 5000, 3, 5)
	w, list := api.doRaw(http.MethodGet, "/notifications?limit=50", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("list notifications: %d %s", w.Code, w.Body.String())
	}
	items, _ := list["items"].([]any)
	found := false
	for _, raw := range items {
		n, _ := raw.(map[string]any)
		if n["type"] == "low_stock" {
			found = true
		}
	}
	if !found {
		t.Fatalf("low_stock notification must exist after creating a below-minimum product: %v", list)
	}
}
