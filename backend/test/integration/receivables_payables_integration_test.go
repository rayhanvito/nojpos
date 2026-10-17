package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCreditCheckoutReceivablePaymentsAndCashRegression(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Piutang")
	cashierToken := api.createCashier(adminToken)
	otherAdminToken := api.registerAdmin("Toko Piutang Lain")

	w, settings := api.doReq(http.MethodGet, "/settings", nil, adminToken)
	if w.Code != http.StatusOK || settings["creditDueDays"] != float64(14) {
		t.Fatalf("default credit due days: %d %s", w.Code, w.Body.String())
	}
	setCRMSettings(t, api, adminToken, map[string]any{
		"creditDueDays": 3, "taxEnabled": true, "taxPct": 10,
	})
	customer := createCRMCustomer(t, api, adminToken, "Pelanggan Kredit", "081211110001")
	customerID := uint(customer["id"].(float64))
	productID := createCRMProduct(t, api, adminToken, "Barang Kredit", 10000, 10)

	w, rejected := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "Credit", "paid": 0,
	}, cashierToken)
	if w.Code != http.StatusBadRequest || !strings.Contains(strings.ToLower(fmt.Sprint(rejected["error"])), "pelanggan") {
		t.Fatalf("Credit must require customerId: %d %s", w.Code, w.Body.String())
	}
	w, rejected = api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "Credit", "paid": 1, "customerId": customerID,
	}, cashierToken)
	if w.Code != http.StatusBadRequest || !strings.Contains(strings.ToLower(fmt.Sprint(rejected["error"])), "paid") {
		t.Fatalf("Credit must reject paid greater than zero: %d %s", w.Code, w.Body.String())
	}

	w, credit := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "discount": 1000,
		"method": "Credit", "paid": 0, "customerId": customerID,
	}, cashierToken)
	if w.Code != http.StatusCreated || credit["total"] != float64(9900) || credit["paid"] != float64(0) || credit["change"] != float64(0) {
		t.Fatalf("Credit checkout totals: %d %s", w.Code, w.Body.String())
	}
	creditID := uint(credit["id"].(float64))
	createdAt, err := time.Parse(time.RFC3339Nano, credit["created_at"].(string))
	if err != nil {
		t.Fatalf("parse transaction created_at: %v", err)
	}
	location, err := time.LoadLocation("Asia/Makassar")
	if err != nil {
		t.Fatal(err)
	}
	createdDate := createdAt.In(location).AddDate(0, 0, 3)

	w, list := api.doReq(http.MethodGet, "/receivables?status=unpaid", nil, adminToken)
	items := list["items"].([]any)
	if w.Code != http.StatusOK || len(items) != 1 {
		t.Fatalf("list receivables: %d %s", w.Code, w.Body.String())
	}
	receivable := items[0].(map[string]any)
	receivableID := uint(receivable["id"].(float64))
	if receivable["customer_id"] != float64(customerID) || receivable["transaction_id"] != float64(creditID) || receivable["amount"] != float64(9900) || receivable["remaining"] != float64(9900) || receivable["status"] != "unpaid" || receivable["customer_name"] != "Pelanggan Kredit" || receivable["invoice_number"] != fmt.Sprintf("TRX-%05d", creditID) {
		t.Fatalf("receivable fields: %v", receivable)
	}
	dueDate, err := time.Parse(time.RFC3339Nano, receivable["due_date"].(string))
	if err != nil || dueDate.In(location).Format("2006-01-02") != createdDate.Format("2006-01-02") {
		t.Fatalf("receivable due date: %v, err=%v; transaction=%v", receivable["due_date"], err, createdAt)
	}
	w, byPaidStatus := api.doReq(http.MethodGet, "/receivables?status=paid", nil, adminToken)
	if w.Code != http.StatusOK || len(byPaidStatus["items"].([]any)) != 0 {
		t.Fatalf("receivable status filter: %d %s", w.Code, w.Body.String())
	}
	w, notOverdue := api.doReq(http.MethodGet, "/receivables?overdue=true", nil, adminToken)
	if w.Code != http.StatusOK || len(notOverdue["items"].([]any)) != 0 {
		t.Fatalf("future receivable must not be overdue: %d %s", w.Code, w.Body.String())
	}

	for _, token := range []string{cashierToken, otherAdminToken} {
		w, _ = api.doReq(http.MethodGet, "/receivables", nil, token)
		wantStatus := http.StatusForbidden
		if token == otherAdminToken {
			wantStatus = http.StatusOK
		}
		if w.Code != wantStatus {
			t.Fatalf("receivable list role/scope: got %d want %d: %s", w.Code, wantStatus, w.Body.String())
		}
	}
	w, _ = api.doReq(http.MethodGet, fmt.Sprintf("/receivables/%d", receivableID), nil, otherAdminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("foreign receivable detail: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/receivables/%d/payments", receivableID), map[string]any{"amount": 1, "method": "Cash"}, cashierToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("cashier receivable payment: %d %s", w.Code, w.Body.String())
	}
	w, rejected = api.doReq(http.MethodPost, fmt.Sprintf("/receivables/%d/payments", receivableID), map[string]any{"amount": 1, "method": "Credit"}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("Credit is not a receivable payment method: %d %s", w.Code, w.Body.String())
	}

	w, partial := api.doReq(http.MethodPost, fmt.Sprintf("/receivables/%d/payments", receivableID), map[string]any{"amount": 3000, "method": "Cash"}, adminToken)
	if w.Code != http.StatusOK || partial["status"] != "partially_paid" || partial["remaining"] != float64(6900) {
		t.Fatalf("partial receivable payment: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/receivables/%d/payments", receivableID), map[string]any{"amount": 6901, "method": "QRIS"}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("overpayment must be rejected: %d %s", w.Code, w.Body.String())
	}
	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/receivables/%d", receivableID), nil, adminToken)
	if w.Code != http.StatusOK || detail["remaining"] != float64(6900) || len(detail["payments"].([]any)) != 1 {
		t.Fatalf("rejected payment must not change balance/history: %d %s", w.Code, w.Body.String())
	}
	w, paid := api.doReq(http.MethodPost, fmt.Sprintf("/receivables/%d/payments", receivableID), map[string]any{"amount": 6900, "method": "Bank Transfer"}, adminToken)
	if w.Code != http.StatusOK || paid["status"] != "paid" || paid["remaining"] != float64(0) || len(paid["payments"].([]any)) != 2 {
		t.Fatalf("settle receivable: %d %s", w.Code, w.Body.String())
	}

	w, cash := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "Cash", "paid": 11000,
	}, cashierToken)
	if w.Code != http.StatusCreated || cash["method"] != "Cash" {
		t.Fatalf("regular cash checkout regression: %d %s", w.Code, w.Body.String())
	}
	w, list = api.doReq(http.MethodGet, "/receivables", nil, adminToken)
	if w.Code != http.StatusOK || len(list["items"].([]any)) != 1 {
		t.Fatalf("non-Credit checkout must not create receivable: %d %s", w.Code, w.Body.String())
	}
}

func TestCreditCheckoutReceivableUsesAmountAfterPointsDiscount(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Kredit Loyalty")
	customer := createCRMCustomer(t, api, adminToken, "Pelanggan Poin", "081211110002")
	customerID := uint(customer["id"].(float64))
	productID := createCRMProduct(t, api, adminToken, "Barang Poin Kredit", 20000, 5)
	setCRMSettings(t, api, adminToken, map[string]any{"loyaltyEnabled": true})

	w, earned := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "Cash", "paid": 20000, "customerId": customerID,
	}, adminToken)
	if w.Code != http.StatusCreated || earned["points_earned"] != float64(2) {
		t.Fatalf("earn points before Credit checkout: %d %s", w.Code, w.Body.String())
	}
	w, credit := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "Credit", "paid": 0,
		"customerId": customerID, "pointsRedeem": 1,
	}, adminToken)
	if w.Code != http.StatusCreated || credit["total"] != float64(20000) || credit["points_discount"] != float64(100) || credit["paid"] != float64(0) || credit["change"] != float64(0) {
		t.Fatalf("Credit checkout with points: %d %s", w.Code, w.Body.String())
	}
	w, list := api.doReq(http.MethodGet, "/receivables", nil, adminToken)
	if w.Code != http.StatusOK || len(list["items"].([]any)) != 1 || list["items"].([]any)[0].(map[string]any)["amount"] != float64(19900) {
		t.Fatalf("receivable after points discount: %d %s", w.Code, w.Body.String())
	}
}

func TestPurchaseReceiptsCreateSeparatePayablesAndSettle(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Hutang")
	cashierToken := api.createCashier(adminToken)
	supplierID := createPurchaseTestSupplier(t, api, adminToken, "Supplier Beras")
	productID := createPurchaseTestProduct(t, api, adminToken, "Beras Hutang", "PAYABLE-RICE", 0)
	w, created := api.doReq(http.MethodPost, "/purchase-orders", map[string]any{
		"supplierId": supplierID,
		"items":      []any{map[string]any{"productId": productID, "qty": 10, "buyPrice": 450}},
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create PO: %d %s", w.Code, w.Body.String())
	}
	poID := uint(created["id"].(float64))
	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/purchase-orders/%d", poID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("load PO: %d %s", w.Code, w.Body.String())
	}
	itemID := uint(detail["items"].([]any)[0].(map[string]any)["id"].(float64))
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/order", poID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("order PO: %d %s", w.Code, w.Body.String())
	}
	for _, qty := range []int{4, 2} {
		w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", poID), map[string]any{
			"items": []any{map[string]any{"itemId": itemID, "qty": qty}},
		}, adminToken)
		if w.Code != http.StatusOK {
			t.Fatalf("receive %d units: %d %s", qty, w.Code, w.Body.String())
		}
	}
	w, list := api.doReq(http.MethodGet, "/payables?status=unpaid", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("list payables: %d %s", w.Code, w.Body.String())
	}
	items, ok := list["items"].([]any)
	if !ok || len(items) != 2 {
		t.Fatalf("one payable per receipt: %d %s", w.Code, w.Body.String())
	}
	byAmount := map[float64]map[string]any{}
	for _, raw := range items {
		payable := raw.(map[string]any)
		byAmount[payable["amount"].(float64)] = payable
	}
	first := byAmount[1800]
	second := byAmount[900]
	if first == nil || second == nil || first["remaining"] != float64(1800) || second["remaining"] != float64(900) || first["supplier_name"] != "Supplier Beras" || first["purchase_order_number"] != created["number"] {
		t.Fatalf("payable receipt snapshots and relations: %v", items)
	}
	firstID := uint(first["id"].(float64))
	w, _ = api.doReq(http.MethodGet, "/payables", nil, cashierToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("cashier payable list: %d %s", w.Code, w.Body.String())
	}
	w, paid := api.doReq(http.MethodPost, fmt.Sprintf("/payables/%d/payments", firstID), map[string]any{"amount": 1800, "method": "QRIS"}, adminToken)
	if w.Code != http.StatusOK || paid["status"] != "paid" || paid["remaining"] != float64(0) || len(paid["payments"].([]any)) != 1 {
		t.Fatalf("settle payable: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/payables/%d/payments", firstID), map[string]any{"amount": 1, "method": "Bank Transfer"}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("paid payable cannot be overpaid: %d %s", w.Code, w.Body.String())
	}
}
