package router

import (
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestExpenseCRUDPermissionsAndValidation(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Pengeluaran")
	cashierToken := api.createCashier(adminToken)
	today := time.Now().Format("2006-01-02")

	// Kasir boleh mencatat; pencatat terekam dari sesi.
	w, created := api.doReq(http.MethodPost, "/expenses", map[string]any{
		"date": today, "category": "Listrik & Air", "note": "Token listrik", "amount": 200000, "method": "Cash", "source": "outside",
	}, cashierToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("cashier create expense: %d %s", w.Code, w.Body.String())
	}
	if created["created_by_name"] != "Kasir Promo" {
		t.Fatalf("created_by_name = %v, want Kasir Promo", created["created_by_name"])
	}
	expenseID := uint(created["id"].(float64))

	// Validasi: nominal 0, metode tak sah, kategori kosong -> 400.
	for name, body := range map[string]map[string]any{
		"amount nol":     {"date": today, "category": "Internet", "amount": 0, "method": "Cash"},
		"metode invalid": {"date": today, "category": "Internet", "amount": 1000, "method": "Cek"},
		"kategori kosong": {"date": today, "category": "", "amount": 1000, "method": "Cash"},
		"tanggal rusak":  {"date": "15-10-2026", "category": "Internet", "amount": 1000, "method": "Cash"},
	} {
		w, _ = api.doReq(http.MethodPost, "/expenses", body, adminToken)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("%s: status = %d, want 400 (%s)", name, w.Code, w.Body.String())
		}
	}

	// Daftar terbaca kasir dan berisi tepat satu baris.
	w, list := api.doReq(http.MethodGet, "/expenses", nil, cashierToken)
	if w.Code != http.StatusOK || list["total"].(float64) != 1 {
		t.Fatalf("list expenses: %d total=%v", w.Code, list["total"])
	}

	// Ubah/hapus khusus admin.
	update := map[string]any{"date": today, "category": "Listrik & Air", "note": "Token listrik (koreksi)", "amount": 210000, "method": "Cash", "source": "outside"}
	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/expenses/%d", expenseID), update, cashierToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("cashier update: %d, want 403", w.Code)
	}
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/expenses/%d", expenseID), nil, cashierToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("cashier delete: %d, want 403", w.Code)
	}
	w, updated := api.doReq(http.MethodPut, fmt.Sprintf("/expenses/%d", expenseID), update, adminToken)
	if w.Code != http.StatusOK || updated["amount"].(float64) != 210000 {
		t.Fatalf("admin update: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/expenses/%d", expenseID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("admin delete: %d %s", w.Code, w.Body.String())
	}
	w, list = api.doReq(http.MethodGet, "/expenses", nil, adminToken)
	if list["total"].(float64) != 0 {
		t.Fatalf("after delete total = %v, want 0", list["total"])
	}
}

func TestExpenseDrawerReducesShiftExpectedCashAndReport(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Laci")
	cashierToken := api.createCashier(adminToken)
	today := time.Now().Format("2006-01-02")

	startShiftForTest(t, api, cashierToken, 100000)

	// Pengeluaran tunai dari laci oleh kasir yang sedang shift.
	w, drawer := api.doReq(http.MethodPost, "/expenses", map[string]any{
		"date": today, "category": "Internet", "note": "WiFi bulanan", "amount": 30000, "method": "Cash", "source": "drawer",
	}, cashierToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create drawer expense: %d %s", w.Code, w.Body.String())
	}
	if drawer["shift_id"] == nil {
		t.Fatalf("drawer expense harus tertaut shift: %v", drawer)
	}
	// Pengeluaran tunai di luar laci tidak mengurangi kas shift.
	w, _ = api.doReq(http.MethodPost, "/expenses", map[string]any{
		"date": today, "category": "Listrik & Air", "note": "Token dari rekening", "amount": 50000, "method": "Cash", "source": "outside",
	}, cashierToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create outside expense: %d %s", w.Code, w.Body.String())
	}

	// Shift berjalan: kas diharapkan = 100000 - 30000 = 70000.
	w, current := api.doReq(http.MethodGet, "/cashier/shift", nil, cashierToken)
	if w.Code != http.StatusOK {
		t.Fatalf("current shift: %d %s", w.Code, w.Body.String())
	}
	shift := current["shift"].(map[string]any)
	if shift["cash_expenses"].(float64) != 30000 {
		t.Fatalf("cash_expenses = %v, want 30000", shift["cash_expenses"])
	}
	if shift["expected_cash"].(float64) != 70000 {
		t.Fatalf("expected_cash = %v, want 70000", shift["expected_cash"])
	}

	// Laporan periode berjalan memuat total + rincian per kategori.
	w, report := api.doReq(http.MethodGet, "/reports?period=month", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("report: %d %s", w.Code, w.Body.String())
	}
	summary := report["summary"].(map[string]any)
	if summary["expense_total"].(float64) != 80000 {
		t.Fatalf("expense_total = %v, want 80000", summary["expense_total"])
	}
	rows, ok := report["expenses"].([]any)
	if !ok || len(rows) != 2 {
		t.Fatalf("expenses rows = %v, want 2 kategori", report["expenses"])
	}
}
