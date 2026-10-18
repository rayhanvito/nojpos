package integration

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

func createShiftProduct(t *testing.T, api *promoTestAPI, token string) uint {
	t.Helper()
	w, product := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Shift", "sku": "SHIFT-PRODUCT", "buyPrice": 500, "sellPrice": 1000, "stock": 100, "unit": "pcs",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create shift product: %d %s", w.Code, w.Body.String())
	}
	return uint(product["id"].(float64))
}

func createNamedShiftCashier(t *testing.T, api *promoTestAPI, adminToken, name string) string {
	t.Helper()
	w, cashier := api.doReq(http.MethodPost, "/users", map[string]any{"name": name}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create shift cashier: %d %s", w.Code, w.Body.String())
	}
	cashierID := uint(cashier["user"].(map[string]any)["id"].(float64))
	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/users/%d/passcode", cashierID), map[string]any{"passcode": "12345"}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("set shift cashier passcode: %d %s", w.Code, w.Body.String())
	}
	w, switched := api.doReq(http.MethodPost, "/auth/switch", map[string]any{"target_user_id": cashierID, "passcode": "12345"}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("switch to shift cashier: %d %s", w.Code, w.Body.String())
	}
	return switched["access_token"].(string)
}

func startShiftForTest(t *testing.T, api *promoTestAPI, token string, openingCash int64) string {
	t.Helper()
	w, response := api.doReq(http.MethodPost, "/cashier/shift/start", map[string]any{"opening_cash": openingCash}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("start shift: %d %s", w.Code, w.Body.String())
	}
	return response["shift"].(map[string]any)["started_at"].(string)
}

func shiftByStartTime(t *testing.T, api *promoTestAPI, adminToken, startedAt string) map[string]any {
	t.Helper()
	w, response := api.doReq(http.MethodGet, "/shifts", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("list shifts: %d %s", w.Code, w.Body.String())
	}
	items, ok := response["items"].([]any)
	if !ok {
		t.Fatalf("shift list has no items: %v", response)
	}
	for _, raw := range items {
		shift := raw.(map[string]any)
		if shift["started_at"] == startedAt {
			return shift
		}
	}
	t.Fatalf("shift started at %s not found: %v", startedAt, response)
	return nil
}

func assertShiftNumber(t *testing.T, values map[string]any, key string, want float64) {
	t.Helper()
	got, ok := values[key].(float64)
	if !ok || got != want {
		t.Fatalf("%s: want %v, got %v", key, want, values[key])
	}
}

func TestShiftCloseStoresCashSummaryWithCashRefund(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Shift Tutup")
	cashierToken := api.createCashier(adminToken)
	productID := createShiftProduct(t, api, adminToken)

	w, customer := api.doReq(http.MethodPost, "/customers", map[string]any{"name": "Pelanggan Shift", "phone": "081299900001"}, cashierToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create shift credit customer: %d %s", w.Code, w.Body.String())
	}
	customerID := uint(customer["id"].(float64))
	startedAt := startShiftForTest(t, api, cashierToken, 10000)

	methods := []struct {
		name string
		qty  int
	}{
		{name: "Cash", qty: 1},
		{name: "Bank Transfer", qty: 2},
		{name: "QRIS", qty: 3},
		{name: "Credit", qty: 6},
	}
	var cashTransactionID uint
	for _, method := range methods {
		body := map[string]any{
			"items":  []any{map[string]any{"productId": productID, "qty": method.qty}},
			"method": method.name,
			"paid":   int64(method.qty * 1000),
		}
		if method.name == "Credit" {
			body["paid"] = 0
			body["customerId"] = customerID
		}
		w, transaction := api.doReq(http.MethodPost, "/transactions", body, cashierToken)
		if w.Code != http.StatusCreated {
			t.Fatalf("checkout %s: %d %s", method.name, w.Code, w.Body.String())
		}
		if method.name == "Cash" {
			cashTransactionID = uint(transaction["id"].(float64))
		}
	}

	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/transactions/%d/refund", cashTransactionID), map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "reason": "Salah bayar",
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("refund cash transaction: %d %s", w.Code, w.Body.String())
	}

	w, current := api.doReq(http.MethodGet, "/cashier/shift", nil, cashierToken)
	if w.Code != http.StatusOK {
		t.Fatalf("review current shift: %d %s", w.Code, w.Body.String())
	}
	review := current["shift"].(map[string]any)
	assertShiftNumber(t, review, "sales", 12000)
	assertShiftNumber(t, review, "trx_count", 4)
	assertShiftNumber(t, review, "refund_total", 1000)
	assertShiftNumber(t, review, "expected_cash", 10000)
	methodSales := review["sales_by_method"].(map[string]any)
	for method, want := range map[string]float64{
		"Cash": 1000, "Bank Transfer": 2000, "QRIS": 3000, "Credit": 6000,
	} {
		assertShiftNumber(t, methodSales, method, want)
	}

	w, closed := api.doReq(http.MethodPost, "/cashier/shift/close", map[string]any{"closingCash": 12000, "note": "Hitung akhir"}, cashierToken)
	if w.Code != http.StatusOK {
		t.Fatalf("close shift: %d %s", w.Code, w.Body.String())
	}
	summary := closed["summary"].(map[string]any)
	assertShiftNumber(t, summary, "sales", 12000)
	assertShiftNumber(t, summary, "trx_count", 4)
	assertShiftNumber(t, summary, "refund_total", 1000)
	assertShiftNumber(t, summary, "expected_cash", 10000)
	assertShiftNumber(t, summary, "closing_cash_actual", 12000)
	assertShiftNumber(t, summary, "cash_variance", 2000)
	if summary["close_note"] != "Hitung akhir" || summary["closed_by_name"] != "Kasir Promo" {
		t.Fatalf("close actor/note: %v", summary)
	}
	stored := shiftByStartTime(t, api, adminToken, startedAt)
	assertShiftNumber(t, stored, "expected_cash", 10000)
	assertShiftNumber(t, stored, "cash_variance", 2000)
}

func TestShiftCloseRequiresNonnegativeClosingCash(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Validasi Tutup Shift")
	cashierToken := api.createCashier(adminToken)
	startShiftForTest(t, api, cashierToken, 0)

	for _, body := range []any{nil, map[string]any{"closingCash": -1}} {
		w, _ := api.doReq(http.MethodPost, "/cashier/shift/close", body, cashierToken)
		if w.Code < http.StatusBadRequest || w.Code >= http.StatusInternalServerError {
			t.Fatalf("invalid closing cash must return 4xx, got %d %s", w.Code, w.Body.String())
		}
		w, _ = api.doReq(http.MethodGet, "/cashier/shift", nil, cashierToken)
		if w.Code != http.StatusOK {
			t.Fatalf("invalid close must leave shift open: %d %s", w.Code, w.Body.String())
		}
	}
}

func TestShiftOwnershipStaysWithOpeningCashierAfterPINSwitch(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Shift Operator")
	openingCashier := api.createCashier(adminToken)
	otherCashier := createNamedShiftCashier(t, api, adminToken, "Kasir Kedua")
	productID := createShiftProduct(t, api, adminToken)
	startedAt := startShiftForTest(t, api, openingCashier, 5000)

	// Penjualan pembuka tercatat di shift miliknya sendiri.
	wc, _ := api.doRaw(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "Cash", "paid": 1000,
	}, openingCashier)
	if wc.Code != http.StatusCreated {
		t.Fatalf("checkout as opening cashier: %d %s", wc.Code, wc.Body.String())
	}
	// Kasir lain TANPA shift sendiri ditolak (gerbang shift 2026-10-16):
	// ia tidak mewarisi shift kasir pembuka.
	wc, _ = api.doRaw(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 2}}, "method": "Cash", "paid": 2000,
	}, otherCashier)
	if wc.Code != http.StatusConflict {
		t.Fatalf("checkout without own shift must be rejected: %d %s", wc.Code, wc.Body.String())
	}

	shift := shiftByStartTime(t, api, adminToken, startedAt)
	assertShiftNumber(t, shift, "sales", 1000)
	assertShiftNumber(t, shift, "trx_count", 1)
	w, _ := api.doReq(http.MethodGet, "/cashier/shift", nil, otherCashier)
	if w.Code != http.StatusNotFound {
		t.Fatalf("switched cashier must not inherit the other cashier shift: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, "/cashier/shift/close", map[string]any{"closingCash": 1000}, otherCashier)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("cashier without an owned shift cannot close another: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodGet, "/cashier/shift", nil, openingCashier)
	if w.Code != http.StatusOK {
		t.Fatalf("opening cashier shift must remain open: %d %s", w.Code, w.Body.String())
	}
}

func TestForceCloseShiftByAdminAndOptionalCash(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Force Close")
	firstCashier := createNamedShiftCashier(t, api, adminToken, "Kasir Force Satu")
	secondCashier := createNamedShiftCashier(t, api, adminToken, "Kasir Force Dua")
	productID := createShiftProduct(t, api, adminToken)
	firstStart := startShiftForTest(t, api, firstCashier, 500)
	secondStart := startShiftForTest(t, api, secondCashier, 700)
	w, _ := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "Cash", "paid": 1000,
	}, firstCashier)
	if w.Code != http.StatusCreated {
		t.Fatalf("cashier checkout before force-close: %d %s", w.Code, w.Body.String())
	}
	first := shiftByStartTime(t, api, adminToken, firstStart)
	second := shiftByStartTime(t, api, adminToken, secondStart)
	firstID := uint(first["id"].(float64))
	secondID := uint(second["id"].(float64))

	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/shifts/%d/force-close", secondID), map[string]any{}, firstCashier)
	if w.Code != http.StatusForbidden {
		t.Fatalf("cashier force-close must be forbidden: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/shifts/%d/force-close", firstID), map[string]any{"closingCash": -1}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("negative force-close cash must be rejected: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodGet, "/cashier/shift", nil, firstCashier)
	if w.Code != http.StatusOK {
		t.Fatalf("invalid force-close must leave shift open: %d %s", w.Code, w.Body.String())
	}
	w, forced := api.doReq(http.MethodPost, fmt.Sprintf("/shifts/%d/force-close", firstID), map[string]any{"closingCash": 1400, "note": "Dihitung admin"}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("force-close with cash: %d %s", w.Code, w.Body.String())
	}
	summary := forced["summary"].(map[string]any)
	assertShiftNumber(t, summary, "expected_cash", 1500)
	assertShiftNumber(t, summary, "closing_cash_actual", 1400)
	assertShiftNumber(t, summary, "cash_variance", -100)
	if summary["close_note"] != "Dihitung admin" || summary["closed_by_name"] != "Admin Promo" {
		t.Fatalf("force-close actor/note: %v", summary)
	}

	w, forced = api.doReq(http.MethodPost, fmt.Sprintf("/shifts/%d/force-close", secondID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("force-close without cash: %d %s", w.Code, w.Body.String())
	}
	summary = forced["summary"].(map[string]any)
	assertShiftNumber(t, summary, "expected_cash", 700)
	if summary["closing_cash_actual"] != nil || summary["cash_variance"] != nil || summary["close_note"] != "Ditutup paksa oleh admin" {
		t.Fatalf("force-close without actual cash: %v", summary)
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/shifts/%d/force-close", secondID), nil, adminToken)
	if w.Code != http.StatusConflict {
		t.Fatalf("closed shift cannot be force-closed again: %d %s", w.Code, w.Body.String())
	}
	if firstStart == secondStart {
		t.Fatal("test setup expected distinct shift start timestamps")
	}
}

func TestShiftAdminReportsFilterOutletAndPaginateDetail(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Laporan Shift")
	cashierToken := api.createCashier(adminToken)
	productID := createShiftProduct(t, api, adminToken)
	startedAt := startShiftForTest(t, api, cashierToken, 1000)
	for _, method := range []string{"Cash", "Bank Transfer"} {
		w, _ := api.doReq(http.MethodPost, "/transactions", map[string]any{
			"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": method, "paid": 1000,
		}, cashierToken)
		if w.Code != http.StatusCreated {
			t.Fatalf("checkout for detail pagination: %d %s", w.Code, w.Body.String())
		}
	}
	shiftA := shiftByStartTime(t, api, adminToken, startedAt)
	shiftAID := uint(shiftA["id"].(float64))
	if shiftA["closing_cash_actual"] != nil || shiftA["cash_variance"] != nil || shiftA["close_note"] != nil || shiftA["closed_by_name"] != nil {
		t.Fatalf("open shift close fields must be null: %v", shiftA)
	}
	assertShiftNumber(t, shiftA, "expected_cash", 2000)

	w, outletB := api.doReq(http.MethodPost, "/outlets", map[string]any{"name": "Outlet Dua"}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create second outlet: %d %s", w.Code, w.Body.String())
	}
	outletBID := uint(outletB["id"].(float64))
	w, adminOutletB := api.doReq(http.MethodPost, "/outlets/switch", map[string]any{"outletId": outletBID}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("switch admin outlet: %d %s", w.Code, w.Body.String())
	}
	adminOutletBToken := adminOutletB["access_token"].(string)
	secondStart := startShiftForTest(t, api, adminOutletBToken, 2500)
	w, _ = api.doReq(http.MethodPost, "/cashier/shift/close", map[string]any{"closingCash": 2500}, adminOutletBToken)
	if w.Code != http.StatusOK {
		t.Fatalf("close admin shift in second outlet: %d %s", w.Code, w.Body.String())
	}
	shiftB := shiftByStartTime(t, api, adminToken, secondStart)
	shiftBID := uint(shiftB["id"].(float64))
	outletAID := uint(shiftA["outlet_id"].(float64))

	w, combined := api.doReq(http.MethodGet, "/shifts", nil, adminToken)
	if w.Code != http.StatusOK || len(combined["items"].([]any)) != 2 {
		t.Fatalf("admin shifts without outlet filter must combine outlets: %d %s", w.Code, w.Body.String())
	}
	if _, ok := combined["shifts"]; ok {
		t.Fatalf("admin shifts response must not contain legacy shifts key: %v", combined)
	}
	w, filtered := api.doReq(http.MethodGet, fmt.Sprintf("/shifts?outletId=%d", outletBID), nil, adminToken)
	if w.Code != http.StatusOK || len(filtered["items"].([]any)) != 1 || uint(filtered["items"].([]any)[0].(map[string]any)["id"].(float64)) != shiftBID {
		t.Fatalf("admin outlet filter: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodGet, fmt.Sprintf("/shifts/export?outletId=%d", outletBID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("export one outlet's shifts: %d %s", w.Code, w.Body.String())
	}
	exportRows, err := csv.NewReader(strings.NewReader(w.Body.String())).ReadAll()
	if err != nil || len(exportRows) != 2 || exportRows[1][0] != strconv.FormatUint(uint64(shiftBID), 10) {
		t.Fatalf("shift export outlet filter: rows=%#v err=%v", exportRows, err)
	}
	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/shifts/%d?page=1&limit=1&outletId=%d", shiftAID, outletAID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("get shift detail: %d %s", w.Code, w.Body.String())
	}
	transactions := detail["transactions"].(map[string]any)
	if len(transactions["items"].([]any)) != 1 || transactions["page"] != float64(1) || transactions["limit"] != float64(1) || transactions["total"] != float64(2) {
		t.Fatalf("shift detail pagination: %v", transactions)
	}
	w, _ = api.doReq(http.MethodGet, fmt.Sprintf("/shifts/%d?outletId=%d", shiftAID, outletBID), nil, adminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("shift detail outside requested outlet must be hidden: %d %s", w.Code, w.Body.String())
	}
	otherAdminToken := api.registerAdmin("Toko Laporan Shift Asing")
	w, _ = api.doReq(http.MethodGet, fmt.Sprintf("/shifts?outletId=%d", outletBID), nil, otherAdminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("foreign outlet filter must be rejected: %d %s", w.Code, w.Body.String())
	}
}

func TestShiftCSVExportHeaderAndValues(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Ekspor Shift")
	cashierToken := api.createCashier(adminToken)
	productID := createShiftProduct(t, api, adminToken)
	startedAt := startShiftForTest(t, api, cashierToken, 5000)
	for _, method := range []struct {
		name string
		qty  int
	}{{"Cash", 1}, {"Bank Transfer", 2}} {
		w, _ := api.doReq(http.MethodPost, "/transactions", map[string]any{
			"items": []any{map[string]any{"productId": productID, "qty": method.qty}}, "method": method.name, "paid": method.qty * 1000,
		}, cashierToken)
		if w.Code != http.StatusCreated {
			t.Fatalf("checkout before shift CSV: %d %s", w.Code, w.Body.String())
		}
	}
	shift := shiftByStartTime(t, api, adminToken, startedAt)
	shiftID := uint(shift["id"].(float64))

	w, _ := api.doReq(http.MethodGet, "/shifts/export", nil, adminToken)
	if w.Code != http.StatusOK || !strings.HasPrefix(w.Header().Get("Content-Type"), "text/csv") {
		t.Fatalf("export shifts: %d %s", w.Code, w.Body.String())
	}
	if !strings.HasPrefix(w.Header().Get("Content-Disposition"), "attachment; filename=\"laporan-shift-") {
		t.Fatalf("shift export must use the download filename pattern: %q", w.Header().Get("Content-Disposition"))
	}
	rows, err := csv.NewReader(strings.NewReader(w.Body.String())).ReadAll()
	if err != nil {
		t.Fatalf("parse shift CSV: %v", err)
	}
	wantHeader := []string{
		"id", "kasir", "mulai", "selesai", "modal awal", "penjualan Cash", "penjualan Bank Transfer", "penjualan QRIS",
		"penjualan E-Wallet", "penjualan Card", "penjualan Credit", "refund", "total penjualan", "jumlah transaksi",
		"expected", "aktual", "selisih",
	}
	if len(rows) != 2 || strings.Join(rows[0], "|") != strings.Join(wantHeader, "|") {
		t.Fatalf("shift CSV header/rows: %#v", rows)
	}
	var shiftRow []string
	for _, row := range rows[1:] {
		if row[0] == strconv.FormatUint(uint64(shiftID), 10) {
			shiftRow = row
		}
	}
	if len(shiftRow) == 0 {
		t.Fatalf("shift row %d missing from CSV: %#v", shiftID, rows)
	}
	wantValues := map[int]string{5: "1000", 6: "2000", 11: "0", 12: "3000", 13: "2", 14: "6000", 15: "", 16: ""}
	for index, want := range wantValues {
		if shiftRow[index] != want {
			t.Fatalf("shift CSV column %q: want %q, got %q (row=%#v)", wantHeader[index], want, shiftRow[index], shiftRow)
		}
	}
}
