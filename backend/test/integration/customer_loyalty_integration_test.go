package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func createCRMCustomer(t *testing.T, api *promoTestAPI, token, name, phone string) map[string]any {
	t.Helper()
	w, response := api.doReq(http.MethodPost, "/customers", map[string]any{
		"name": name, "phone": phone, "email": strings.ToLower(strings.ReplaceAll(name, " ", "-")) + "@example.com", "address": "Jalan Contoh",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create customer %q: %d %s", name, w.Code, w.Body.String())
	}
	return response
}

func createCRMProduct(t *testing.T, api *promoTestAPI, token, name string, price int64, stock int) uint {
	t.Helper()
	w, response := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": name, "sku": strings.ToUpper(strings.ReplaceAll(name, " ", "-")), "sellPrice": price, "stock": stock,
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create product %q: %d %s", name, w.Code, w.Body.String())
	}
	return uint(response["id"].(float64))
}

func setCRMSettings(t *testing.T, api *promoTestAPI, token string, settings map[string]any) map[string]any {
	t.Helper()
	w, response := api.doReq(http.MethodPut, "/settings", settings, token)
	if w.Code != http.StatusOK {
		t.Fatalf("update settings: %d %s", w.Code, w.Body.String())
	}
	return response
}

func TestCustomerCRUDRolesAndStoreScope(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko CRM Satu")
	cashierToken := api.createCashier(adminToken)
	otherAdminToken := api.registerAdmin("Toko CRM Dua")

	created := createCRMCustomer(t, api, cashierToken, "Alice Pembeli", "081234567890")
	customerID := uint(created["id"].(float64))
	if created["points"].(float64) != 0 || created["name"] != "Alice Pembeli" {
		t.Fatalf("created customer fields: %v", created)
	}
	w, optionalFields := api.doReq(http.MethodPost, "/customers", map[string]any{"name": "Kontak Opsional"}, cashierToken)
	if w.Code != http.StatusCreated || optionalFields["email"] != "" || optionalFields["address"] != "" {
		t.Fatalf("email and address must be optional: %d %s", w.Code, w.Body.String())
	}
	w, invalidName := api.doReq(http.MethodPost, "/customers", map[string]any{"name": "   "}, cashierToken)
	if w.Code != http.StatusBadRequest || invalidName["error"] != "Nama pelanggan wajib diisi." {
		t.Fatalf("customer name is required: %d %s", w.Code, w.Body.String())
	}

	w, duplicate := api.doReq(http.MethodPost, "/customers", map[string]any{"name": "Alice Duplikat", "phone": "081234567890"}, adminToken)
	if w.Code != http.StatusConflict || !strings.Contains(strings.ToLower(duplicate["error"].(string)), "telepon") {
		t.Fatalf("duplicate phone: %d %s", w.Code, w.Body.String())
	}

	if other := createCRMCustomer(t, api, otherAdminToken, "Alice Toko Lain", "081234567890"); other["id"] == created["id"] {
		t.Fatalf("same phone must be allowed in another store: %v", other)
	}
	createCRMCustomer(t, api, adminToken, "Tanpa Telepon Satu", "")
	createCRMCustomer(t, api, adminToken, "Tanpa Telepon Dua", "")

	w, listed := api.doReq(http.MethodGet, "/customers?q=ALICE&page=1&limit=1", nil, cashierToken)
	items, ok := listed["items"].([]any)
	if w.Code != http.StatusOK || !ok || len(items) != 1 || uint(items[0].(map[string]any)["id"].(float64)) != customerID {
		t.Fatalf("case-insensitive, paginated customer list: %d %s", w.Code, w.Body.String())
	}
	w, byPhone := api.doReq(http.MethodGet, "/customers?q=081234567890", nil, cashierToken)
	if w.Code != http.StatusOK || len(byPhone["items"].([]any)) != 1 {
		t.Fatalf("phone search: %d %s", w.Code, w.Body.String())
	}

	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/customers/%d", customerID), nil, cashierToken)
	if w.Code != http.StatusOK || detail["points"].(float64) != 0 {
		t.Fatalf("customer detail with points: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodGet, fmt.Sprintf("/customers/%d", uint(created["id"].(float64))), nil, otherAdminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("cross-store customer detail must be hidden: %d %s", w.Code, w.Body.String())
	}

	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/customers/%d", customerID), map[string]any{"name": "Alice Baru", "phone": "081234567890"}, cashierToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("cashier PUT must be forbidden: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/customers/%d", customerID), nil, cashierToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("cashier DELETE must be forbidden: %d %s", w.Code, w.Body.String())
	}
	w, updated := api.doReq(http.MethodPut, fmt.Sprintf("/customers/%d", customerID), map[string]any{
		"name": "Alice Baru", "phone": "081234567890", "email": "alice-baru@example.com", "address": "Jalan Baru",
	}, adminToken)
	if w.Code != http.StatusOK || updated["name"] != "Alice Baru" {
		t.Fatalf("admin update: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/customers/%d", customerID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("admin delete: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodGet, fmt.Sprintf("/customers/%d", customerID), nil, adminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("soft-deleted customer should not be visible: %d %s", w.Code, w.Body.String())
	}
}

func TestLoyaltyCheckoutAndRefundRules(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Loyalty")
	cashierToken := api.createCashier(adminToken)
	otherAdminToken := api.registerAdmin("Toko Loyalty Lain")
	w, defaults := api.doReq(http.MethodGet, "/settings", nil, adminToken)
	if w.Code != http.StatusOK || defaults["loyalty_enabled"] != false || defaults["loyalty_earn_per"].(float64) != 10000 || defaults["loyalty_point_value"].(float64) != 100 {
		t.Fatalf("default loyalty settings: %d %s", w.Code, w.Body.String())
	}
	setCRMSettings(t, api, adminToken, map[string]any{
		"loyaltyEnabled": true, "loyaltyEarnPer": 27500, "loyaltyPointValue": 100,
		"taxEnabled": true, "taxPct": 10, "taxInclusive": false,
	})
	w, settings := api.doReq(http.MethodGet, "/settings", nil, adminToken)
	if w.Code != http.StatusOK || settings["loyalty_enabled"] != true || settings["loyalty_earn_per"].(float64) != 27500 || settings["loyalty_point_value"].(float64) != 100 {
		t.Fatalf("loyalty settings response/default keys: %d %s", w.Code, w.Body.String())
	}

	customer := createCRMCustomer(t, api, cashierToken, "Budi Loyal", "081200000001")
	customerID := uint(customer["id"].(float64))
	otherCustomer := createCRMCustomer(t, api, otherAdminToken, "Pelanggan Asing", "081200000002")
	productID := createCRMProduct(t, api, adminToken, "Barang Loyalty", 25000, 20)

	checkout := func(body map[string]any) (*httptest.ResponseRecorder, map[string]any) {
		t.Helper()
		return api.doReq(http.MethodPost, "/transactions", body, cashierToken)
	}
	baseCheckout := func() map[string]any {
		return map[string]any{"items": []any{map[string]any{"productId": productID, "qty": 2}}, "method": "Cash", "paid": 60000}
	}
	assertNoCheckoutMutation := func(caseName string) {
		t.Helper()
		w, customerState := api.doReq(http.MethodGet, fmt.Sprintf("/customers/%d", customerID), nil, cashierToken)
		if w.Code != http.StatusOK || customerState["points"].(float64) != 2 {
			t.Fatalf("%s changed customer points: %d %s", caseName, w.Code, w.Body.String())
		}
		w, productState := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d", productID), nil, cashierToken)
		if w.Code != http.StatusOK || productState["stock"].(float64) != 18 {
			t.Fatalf("%s changed product stock: %d %s", caseName, w.Code, w.Body.String())
		}
		w, transactionState := api.doReq(http.MethodGet, "/transactions", nil, cashierToken)
		items, ok := transactionState["items"].([]any)
		if w.Code != http.StatusOK || !ok || len(items) != 1 {
			t.Fatalf("%s created a transaction: %d %s", caseName, w.Code, w.Body.String())
		}
	}

	first := baseCheckout()
	first["customerId"] = customerID
	w, firstTrx := checkout(first)
	if w.Code != http.StatusCreated || firstTrx["tax"].(float64) != 5000 || firstTrx["total"].(float64) != 55000 || firstTrx["points_earned"].(float64) != 2 {
		t.Fatalf("earn based on tax-inclusive checkout total: %d %s", w.Code, w.Body.String())
	}
	if uint(firstTrx["customer_id"].(float64)) != customerID || firstTrx["points_redeemed"].(float64) != 0 || firstTrx["points_discount"].(float64) != 0 {
		t.Fatalf("checkout loyalty response: %v", firstTrx)
	}

	tooManyPoints := baseCheckout()
	tooManyPoints["customerId"], tooManyPoints["pointsRedeem"] = customerID, 6
	w, _ = checkout(tooManyPoints)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("redeem above balance must be rejected: %d %s", w.Code, w.Body.String())
	}
	assertNoCheckoutMutation("redeem above balance")
	setCRMSettings(t, api, adminToken, map[string]any{"loyaltyPointValue": 55001})
	tooMuchDiscount := baseCheckout()
	tooMuchDiscount["customerId"], tooMuchDiscount["pointsRedeem"] = customerID, 1
	w, _ = checkout(tooMuchDiscount)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("points discount above total must be rejected: %d %s", w.Code, w.Body.String())
	}
	assertNoCheckoutMutation("points discount above total")
	noCustomer := baseCheckout()
	noCustomer["pointsRedeem"] = 1
	w, _ = checkout(noCustomer)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("redeem without customer must be rejected while enabled: %d %s", w.Code, w.Body.String())
	}
	assertNoCheckoutMutation("redeem without customer")
	foreignCustomer := baseCheckout()
	foreignCustomer["customerId"] = otherCustomer["id"]
	w, _ = checkout(foreignCustomer)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("cross-store customer checkout: %d %s", w.Code, w.Body.String())
	}
	assertNoCheckoutMutation("cross-store customer")
	setCRMSettings(t, api, adminToken, map[string]any{"loyaltyPointValue": 100})

	validRedeem := baseCheckout()
	validRedeem["customerId"], validRedeem["pointsRedeem"] = customerID, 1
	w, redeemed := checkout(validRedeem)
	if w.Code != http.StatusCreated || redeemed["tax"].(float64) != 5000 || redeemed["total"].(float64) != 55000 || redeemed["points_redeemed"].(float64) != 1 || redeemed["points_discount"].(float64) != 100 || redeemed["paid"].(float64) != 60000 || redeemed["change"].(float64) != 5100 {
		t.Fatalf("redeem is a post-tax payment and leaves tax unchanged: %d %s", w.Code, w.Body.String())
	}

	w, customerAfterRedeem := api.doReq(http.MethodGet, fmt.Sprintf("/customers/%d", customerID), nil, cashierToken)
	if w.Code != http.StatusOK || customerAfterRedeem["points"].(float64) != 3 {
		t.Fatalf("balance after earn and redeem: %d %s", w.Code, w.Body.String())
	}

	walkIn := baseCheckout()
	walkIn["customer"] = "Pelanggan Walk-in"
	w, walkInTrx := checkout(walkIn)
	if w.Code != http.StatusCreated || walkInTrx["customer"] != "Pelanggan Walk-in" || walkInTrx["points_earned"].(float64) != 0 {
		t.Fatalf("legacy walk-in checkout changed: %d %s", w.Code, w.Body.String())
	}

	setCRMSettings(t, api, adminToken, map[string]any{"loyaltyEnabled": false})
	disabled := baseCheckout()
	disabled["customerId"], disabled["pointsRedeem"] = customerID, 999
	w, disabledTrx := checkout(disabled)
	if w.Code != http.StatusCreated || uint(disabledTrx["customer_id"].(float64)) != customerID || disabledTrx["points_earned"].(float64) != 0 || disabledTrx["points_redeemed"].(float64) != 0 || disabledTrx["points_discount"].(float64) != 0 {
		t.Fatalf("disabled loyalty should record CRM customer and ignore points: %d %s", w.Code, w.Body.String())
	}
	setCRMSettings(t, api, adminToken, map[string]any{"loyaltyEnabled": true})

	// Separate low-priced item earns no points; refunding the redemption must not restore spent points.
	cheapProductID := createCRMProduct(t, api, adminToken, "Barang Murah Loyalty", 1000, 5)
	w, redeemedTrx := checkout(map[string]any{
		"items": []any{map[string]any{"productId": cheapProductID, "qty": 1}}, "method": "Cash", "paid": 1000,
		"customerId": customerID, "pointsRedeem": 1,
	})
	if w.Code != http.StatusCreated || redeemedTrx["points_discount"].(float64) != 100 {
		t.Fatalf("spend balance before refund test: %d %s", w.Code, w.Body.String())
	}
	newestTrxID := uint(redeemedTrx["id"].(float64))
	w, history := api.doReq(http.MethodGet, fmt.Sprintf("/customers/%d/transactions", customerID), nil, cashierToken)
	historyItems, ok := history["items"].([]any)
	if w.Code != http.StatusOK || !ok || len(historyItems) < 3 || uint(historyItems[0].(map[string]any)["id"].(float64)) != newestTrxID {
		t.Fatalf("customer transaction history must be newest first: %d %s", w.Code, w.Body.String())
	}

	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/transactions/%d/refund", newestTrxID), map[string]any{
		"items": []any{map[string]any{"productId": cheapProductID, "qty": 1}}, "reason": "Barang dikembalikan",
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("refund redemption transaction: %d %s", w.Code, w.Body.String())
	}
	w, customerAfterRefund := api.doReq(http.MethodGet, fmt.Sprintf("/customers/%d", customerID), nil, cashierToken)
	if w.Code != http.StatusOK || customerAfterRefund["points"].(float64) != 2 {
		t.Fatalf("refund must not restore redeemed points: %d %s", w.Code, w.Body.String())
	}
}

func TestLoyaltyRefundUsesFloorAndClampsBalance(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Refund Poin")
	cashierToken := api.createCashier(adminToken)
	setCRMSettings(t, api, adminToken, map[string]any{"loyaltyEnabled": true, "loyaltyEarnPer": 10000, "loyaltyPointValue": 100})
	customer := createCRMCustomer(t, api, adminToken, "Dewi Refund", "081200000003")
	customerID := uint(customer["id"].(float64))
	productID := createCRMProduct(t, api, adminToken, "Produk Refund Poin", 15000, 5)

	w, sale := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 2}}, "method": "Cash", "paid": 33000, "customerId": customerID,
	}, cashierToken)
	if w.Code != http.StatusCreated || sale["points_earned"].(float64) != 3 {
		t.Fatalf("loyalty sale before partial refund: %d %s", w.Code, w.Body.String())
	}
	trxID := uint(sale["id"].(float64))
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/transactions/%d/refund", trxID), map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "reason": "Refund sebagian",
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("partial refund: %d %s", w.Code, w.Body.String())
	}
	w, afterPartial := api.doReq(http.MethodGet, fmt.Sprintf("/customers/%d", customerID), nil, adminToken)
	if w.Code != http.StatusOK || afterPartial["points"].(float64) != 2 {
		t.Fatalf("refund value 15000 must claw back floor(15000/10000)=1: %d %s", w.Code, w.Body.String())
	}

	cheapProductID := createCRMProduct(t, api, adminToken, "Produk Penukar Refund", 1000, 3)
	w, _ = api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": cheapProductID, "qty": 1}}, "method": "Cash", "paid": 800,
		"customerId": customerID, "pointsRedeem": 2,
	}, cashierToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("spend remaining points: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/transactions/%d/refund", trxID), map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "reason": "Refund sisa",
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("final partial refund: %d %s", w.Code, w.Body.String())
	}
	w, afterClamp := api.doReq(http.MethodGet, fmt.Sprintf("/customers/%d", customerID), nil, cashierToken)
	if w.Code != http.StatusOK || afterClamp["points"].(float64) != 0 {
		t.Fatalf("refund clawback must clamp balance at zero: %d %s", w.Code, w.Body.String())
	}
}
