package integration

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestSupplierCRUDIsStoreScopedAndAdminOnly(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Supplier Satu")
	cashierToken := api.createCashier(adminToken)
	otherAdminToken := api.registerAdmin("Toko Supplier Dua")

	w, supplier := api.doReq(http.MethodPost, "/suppliers", map[string]any{
		"name": "  Distributor Beras  ", "contact": "081234567890", "address": "Jl. Padi No. 1",
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create supplier: %d %s", w.Code, w.Body.String())
	}
	supplierID := uint(supplier["id"].(float64))
	if supplier["name"] != "Distributor Beras" || supplier["contact"] != "081234567890" || supplier["address"] != "Jl. Padi No. 1" {
		t.Fatalf("supplier fields: %v", supplier)
	}

	for _, token := range []string{adminToken, cashierToken} {
		w, list := api.doReq(http.MethodGet, "/suppliers", nil, token)
		items, ok := list["items"].([]any)
		if w.Code != http.StatusOK || !ok || len(items) != 1 {
			t.Fatalf("list suppliers: %d %s", w.Code, w.Body.String())
		}
		w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/suppliers/%d", supplierID), nil, token)
		if w.Code != http.StatusOK || detail["id"] != float64(supplierID) {
			t.Fatalf("get supplier as logged-in user: %d %s", w.Code, w.Body.String())
		}
	}

	w, updated := api.doReq(http.MethodPut, fmt.Sprintf("/suppliers/%d", supplierID), map[string]any{
		"name": "Distributor Padi", "contact": "082233445566", "address": "Jl. Sawah No. 2",
	}, adminToken)
	if w.Code != http.StatusOK || updated["name"] != "Distributor Padi" || updated["contact"] != "082233445566" {
		t.Fatalf("update supplier: %d %s", w.Code, w.Body.String())
	}

	w, _ = api.doReq(http.MethodGet, fmt.Sprintf("/suppliers/%d", supplierID), nil, otherAdminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("foreign store supplier detail: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/suppliers/%d", supplierID), map[string]any{
		"name": "Diubah lintas toko", "contact": "", "address": "",
	}, otherAdminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("foreign store supplier update: %d %s", w.Code, w.Body.String())
	}

	for _, request := range []struct {
		method string
		path   string
		body   any
	}{
		{http.MethodPost, "/suppliers", map[string]any{"name": "Supplier Kasir"}},
		{http.MethodPut, fmt.Sprintf("/suppliers/%d", supplierID), map[string]any{"name": "Supplier Kasir"}},
		{http.MethodDelete, fmt.Sprintf("/suppliers/%d", supplierID), nil},
	} {
		w, _ = api.doReq(request.method, request.path, request.body, cashierToken)
		if w.Code != http.StatusForbidden {
			t.Fatalf("cashier %s %s: %d %s", request.method, request.path, w.Code, w.Body.String())
		}
	}

	w, _ = api.doReq(http.MethodPost, "/suppliers", map[string]any{"name": "  "}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("empty supplier name: %d %s", w.Code, w.Body.String())
	}

	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/suppliers/%d", supplierID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("delete supplier: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodGet, fmt.Sprintf("/suppliers/%d", supplierID), nil, adminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("soft-deleted supplier should be hidden: %d %s", w.Code, w.Body.String())
	}
	w, list := api.doReq(http.MethodGet, "/suppliers", nil, adminToken)
	if w.Code != http.StatusOK || len(list["items"].([]any)) != 0 {
		t.Fatalf("soft-deleted supplier should leave list: %d %s", w.Code, w.Body.String())
	}
}

func createPurchaseTestSupplier(t *testing.T, api *promoTestAPI, token, name string) uint {
	t.Helper()
	w, supplier := api.doReq(http.MethodPost, "/suppliers", map[string]any{
		"name": name, "contact": "081200000000", "address": "Gudang",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create supplier %q: %d %s", name, w.Code, w.Body.String())
	}
	return uint(supplier["id"].(float64))
}

func createPurchaseTestProduct(t *testing.T, api *promoTestAPI, token, name, sku string, stock int) uint {
	t.Helper()
	w, product := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": name, "sku": sku, "buyPrice": 200, "sellPrice": 500, "stock": stock,
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create product %q: %d %s", name, w.Code, w.Body.String())
	}
	return uint(product["id"].(float64))
}

func createPurchaseTestVariant(t *testing.T, api *promoTestAPI, token string, productID uint, name, sku string, stock int) uint {
	t.Helper()
	w, variant := api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/variants", productID), map[string]any{
		"name": name, "sku": sku, "buyPrice": 300, "sellPrice": 600, "stock": stock,
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create variant %q: %d %s", name, w.Code, w.Body.String())
	}
	return uint(variant["id"].(float64))
}

func TestPurchaseOrderDraftNumberingAndOrderLock(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko PO Satu")
	cashierToken := api.createCashier(adminToken)
	otherAdminToken := api.registerAdmin("Toko PO Dua")

	supplierOneID := createPurchaseTestSupplier(t, api, adminToken, "Supplier Satu")
	supplierTwoID := createPurchaseTestSupplier(t, api, adminToken, "Supplier Dua")
	productID := createPurchaseTestProduct(t, api, adminToken, "Beras", "PO-BERAS", 5)

	w, created := api.doReq(http.MethodPost, "/purchase-orders", map[string]any{
		"supplierId": supplierOneID,
		"items":      []any{map[string]any{"productId": productID, "qty": 6, "buyPrice": 450}},
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create draft PO: %d %s", w.Code, w.Body.String())
	}
	poID := uint(created["id"].(float64))
	poNumber := created["number"].(string)
	if matched, _ := regexp.MatchString(`^PO-[0-9]{8}-0001$`, poNumber); !matched || created["status"] != "draft" {
		t.Fatalf("initial PO number/status: %v", created)
	}

	w, updated := api.doReq(http.MethodPut, fmt.Sprintf("/purchase-orders/%d", poID), map[string]any{
		"supplierId": supplierTwoID,
		"items":      []any{map[string]any{"productId": productID, "qty": 4, "buyPrice": 525}},
	}, adminToken)
	if w.Code != http.StatusOK || updated["supplier_id"] != float64(supplierTwoID) {
		t.Fatalf("replace draft PO: %d %s", w.Code, w.Body.String())
	}
	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/purchase-orders/%d", poID), nil, adminToken)
	items := detail["items"].([]any)
	supplier := detail["supplier"].(map[string]any)
	item := items[0].(map[string]any)
	if w.Code != http.StatusOK || supplier["id"] != float64(supplierTwoID) || len(items) != 1 || item["qty_ordered"] != float64(4) || item["buy_price"] != float64(525) || item["qty_received"] != float64(0) {
		t.Fatalf("draft detail with supplier and replaced items: %d %s", w.Code, w.Body.String())
	}

	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/suppliers/%d", supplierTwoID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("soft-delete referenced supplier: %d %s", w.Code, w.Body.String())
	}
	w, detail = api.doReq(http.MethodGet, fmt.Sprintf("/purchase-orders/%d", poID), nil, adminToken)
	if w.Code != http.StatusOK || detail["supplier"].(map[string]any)["id"] != float64(supplierTwoID) {
		t.Fatalf("PO detail should retain its soft-deleted supplier: %d %s", w.Code, w.Body.String())
	}

	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/order", poID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("order draft PO: %d %s", w.Code, w.Body.String())
	}
	w, ordered := api.doReq(http.MethodGet, fmt.Sprintf("/purchase-orders/%d", poID), nil, adminToken)
	if w.Code != http.StatusOK || ordered["status"] != "ordered" {
		t.Fatalf("PO status after order: %d %s", w.Code, w.Body.String())
	}
	w, locked := api.doReq(http.MethodPut, fmt.Sprintf("/purchase-orders/%d", poID), map[string]any{
		"supplierId": supplierOneID,
		"items":      []any{map[string]any{"productId": productID, "qty": 1, "buyPrice": 1}},
	}, adminToken)
	if (w.Code != http.StatusBadRequest && w.Code != http.StatusConflict) || !strings.Contains(strings.ToLower(locked["error"].(string)), "ubah") {
		t.Fatalf("ordered PO PUT must explain item lock: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/purchase-orders/%d", poID), nil, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("ordered PO delete: %d %s", w.Code, w.Body.String())
	}

	for _, roleToken := range []string{adminToken, cashierToken} {
		w, list := api.doReq(http.MethodGet, "/purchase-orders?status=ordered&supplierId="+fmt.Sprint(supplierTwoID), nil, roleToken)
		rows, ok := list["items"].([]any)
		if w.Code != http.StatusOK || !ok || len(rows) != 1 || rows[0].(map[string]any)["id"] != float64(poID) {
			t.Fatalf("filtered PO list: %d %s", w.Code, w.Body.String())
		}
	}

	w, secondPO := api.doReq(http.MethodPost, "/purchase-orders", map[string]any{
		"supplierId": supplierOneID,
		"items":      []any{map[string]any{"productId": productID, "qty": 2, "buyPrice": 400}},
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create second PO: %d %s", w.Code, w.Body.String())
	}
	deletedPOID := uint(secondPO["id"].(float64))
	secondNumber := secondPO["number"].(string)
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/purchase-orders/%d", deletedPOID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("delete draft PO: %d %s", w.Code, w.Body.String())
	}

	w, thirdPO := api.doReq(http.MethodPost, "/purchase-orders", map[string]any{
		"supplierId": supplierOneID,
		"items":      []any{map[string]any{"productId": productID, "qty": 2, "buyPrice": 400}},
	}, adminToken)
	if w.Code != http.StatusCreated || !strings.HasSuffix(secondNumber, "-0002") || !strings.HasSuffix(thirdPO["number"].(string), "-0003") {
		t.Fatalf("deleted draft number must remain consumed: second=%v third=%v status=%d body=%s", secondNumber, thirdPO["number"], w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodGet, fmt.Sprintf("/purchase-orders/%d", deletedPOID), nil, adminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("deleted draft PO should be hidden: %d %s", w.Code, w.Body.String())
	}

	otherSupplierID := createPurchaseTestSupplier(t, api, otherAdminToken, "Supplier Toko Dua")
	otherProductID := createPurchaseTestProduct(t, api, otherAdminToken, "Gula", "PO-GULA", 0)
	w, otherPO := api.doReq(http.MethodPost, "/purchase-orders", map[string]any{
		"supplierId": otherSupplierID,
		"items":      []any{map[string]any{"productId": otherProductID, "qty": 1, "buyPrice": 100}},
	}, otherAdminToken)
	if w.Code != http.StatusCreated || !strings.HasSuffix(otherPO["number"].(string), "-0001") {
		t.Fatalf("daily PO sequence should be store scoped: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodGet, fmt.Sprintf("/purchase-orders/%d", poID), nil, otherAdminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("foreign store PO detail: %d %s", w.Code, w.Body.String())
	}
	w, draftList := api.doReq(http.MethodGet, "/purchase-orders?status=draft&supplierId="+fmt.Sprint(supplierOneID), nil, adminToken)
	if w.Code != http.StatusOK || len(draftList["items"].([]any)) != 1 {
		t.Fatalf("draft/supplier filtered list: %d %s", w.Code, w.Body.String())
	}

	w, _ = api.doReq(http.MethodPost, "/purchase-orders", map[string]any{"supplierId": supplierOneID, "items": []any{}}, cashierToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("cashier create PO: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/order", poID), map[string]any{}, cashierToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("cashier order PO: %d %s", w.Code, w.Body.String())
	}
}

func TestPurchaseOrderRejectsForeignCatalogAndInvalidItems(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko PO Validasi")
	otherAdminToken := api.registerAdmin("Toko PO Asing")
	supplierID := createPurchaseTestSupplier(t, api, adminToken, "Supplier Lokal")
	foreignSupplierID := createPurchaseTestSupplier(t, api, otherAdminToken, "Supplier Asing")
	productID := createPurchaseTestProduct(t, api, adminToken, "Barang Lokal", "PO-LOCAL", 0)
	foreignProductID := createPurchaseTestProduct(t, api, otherAdminToken, "Barang Asing", "PO-FOREIGN", 0)
	variantProductID := createPurchaseTestProduct(t, api, adminToken, "Barang Varian", "PO-VARIANT", 0)
	variantID := createPurchaseTestVariant(t, api, adminToken, variantProductID, "Merah", "PO-VAR-RED", 0)
	foreignVariantProductID := createPurchaseTestProduct(t, api, otherAdminToken, "Barang Varian Asing", "PO-VARIANT-FOREIGN", 0)
	foreignVariantID := createPurchaseTestVariant(t, api, otherAdminToken, foreignVariantProductID, "Biru", "PO-VAR-BLUE", 0)

	validItem := map[string]any{"productId": productID, "qty": 1, "buyPrice": 0}
	for _, tc := range []struct {
		name string
		body map[string]any
	}{
		{"foreign supplier", map[string]any{"supplierId": foreignSupplierID, "items": []any{validItem}}},
		{"foreign product", map[string]any{"supplierId": supplierID, "items": []any{map[string]any{"productId": foreignProductID, "qty": 1, "buyPrice": 1}}}},
		{"foreign variant", map[string]any{"supplierId": supplierID, "items": []any{map[string]any{"productId": productID, "variantId": foreignVariantID, "qty": 1, "buyPrice": 1}}}},
		{"variant belongs to different product", map[string]any{"supplierId": supplierID, "items": []any{map[string]any{"productId": productID, "variantId": variantID, "qty": 1, "buyPrice": 1}}}},
		{"variant required", map[string]any{"supplierId": supplierID, "items": []any{map[string]any{"productId": variantProductID, "qty": 1, "buyPrice": 1}}}},
		{"supplier required", map[string]any{"items": []any{validItem}}},
		{"items required", map[string]any{"supplierId": supplierID}},
		{"items cannot be empty", map[string]any{"supplierId": supplierID, "items": []any{}}},
		{"product required", map[string]any{"supplierId": supplierID, "items": []any{map[string]any{"qty": 1, "buyPrice": 1}}}},
		{"quantity must be positive", map[string]any{"supplierId": supplierID, "items": []any{map[string]any{"productId": productID, "qty": 0, "buyPrice": 1}}}},
		{"quantity cannot be negative", map[string]any{"supplierId": supplierID, "items": []any{map[string]any{"productId": productID, "qty": -1, "buyPrice": 1}}}},
		{"buy price required", map[string]any{"supplierId": supplierID, "items": []any{map[string]any{"productId": productID, "qty": 1}}}},
		{"buy price cannot be negative", map[string]any{"supplierId": supplierID, "items": []any{map[string]any{"productId": productID, "qty": 1, "buyPrice": -1}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w, _ := api.doReq(http.MethodPost, "/purchase-orders", tc.body, adminToken)
			if w.Code != http.StatusBadRequest {
				t.Fatalf("invalid PO should return 400: %d %s", w.Code, w.Body.String())
			}
		})
	}
	w, remaining := api.doReq(http.MethodGet, "/purchase-orders", nil, adminToken)
	if w.Code != http.StatusOK || len(remaining["items"].([]any)) != 0 {
		t.Fatalf("list after invalid PO requests: %d %s", w.Code, w.Body.String())
	}
}

func purchaseOrderItemsByProduct(items any) map[uint]map[string]any {
	result := make(map[uint]map[string]any)
	for _, raw := range items.([]any) {
		item := raw.(map[string]any)
		result[uint(item["product_id"].(float64))] = item
	}
	return result
}

func purchaseTestStock(t *testing.T, api *promoTestAPI, token string, productID, variantID uint) (int, int) {
	t.Helper()
	w, product := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d", productID), nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("get stock for product %d: %d %s", productID, w.Code, w.Body.String())
	}
	productStock := int(product["stock"].(float64))
	if variantID == 0 {
		return productStock, 0
	}
	for _, raw := range product["variants"].([]any) {
		variant := raw.(map[string]any)
		if uint(variant["id"].(float64)) == variantID {
			return productStock, int(variant["stock"].(float64))
		}
	}
	t.Fatalf("variant %d missing from product %d: %s", variantID, productID, w.Body.String())
	return 0, 0
}

func TestPurchaseOrderPartialAndFullReceipt(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko PO Penerimaan")
	cashierToken := api.createCashier(adminToken)
	supplierID := createPurchaseTestSupplier(t, api, adminToken, "Supplier Penerimaan")
	productID := createPurchaseTestProduct(t, api, adminToken, "Beras Penerimaan", "PO-RECEIVE", 3)
	variantProductID := createPurchaseTestProduct(t, api, adminToken, "Susu Varian", "PO-MILK", 0)
	variantID := createPurchaseTestVariant(t, api, adminToken, variantProductID, "Cokelat", "PO-MILK-CHOC", 2)

	w, created := api.doReq(http.MethodPost, "/purchase-orders", map[string]any{
		"supplierId": supplierID,
		"items": []any{
			map[string]any{"productId": productID, "qty": 5, "buyPrice": 450},
			map[string]any{"productId": variantProductID, "variantId": variantID, "qty": 4, "buyPrice": 550},
		},
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create PO: %d %s", w.Code, w.Body.String())
	}
	poID := uint(created["id"].(float64))
	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/purchase-orders/%d", poID), nil, cashierToken)
	if w.Code != http.StatusOK || detail["supplier"] == nil {
		t.Fatalf("cashier should read PO detail: %d %s", w.Code, w.Body.String())
	}
	itemsByProduct := purchaseOrderItemsByProduct(detail["items"])
	productItemID := uint(itemsByProduct[productID]["id"].(float64))
	variantItemID := uint(itemsByProduct[variantProductID]["id"].(float64))
	if itemsByProduct[productID]["buy_price"] != float64(450) || itemsByProduct[variantProductID]["buy_price"] != float64(550) {
		t.Fatalf("buy price snapshots: %s", w.Body.String())
	}

	for _, request := range []struct {
		method string
		path   string
		body   any
	}{
		{http.MethodPut, fmt.Sprintf("/purchase-orders/%d", poID), map[string]any{"supplierId": supplierID, "items": []any{map[string]any{"productId": productID, "qty": 1, "buyPrice": 1}}}},
		{http.MethodDelete, fmt.Sprintf("/purchase-orders/%d", poID), nil},
		{http.MethodPost, fmt.Sprintf("/purchase-orders/%d/order", poID), map[string]any{}},
		{http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", poID), map[string]any{"items": []any{map[string]any{"itemId": productItemID, "qty": 1}}}},
		{http.MethodPost, fmt.Sprintf("/purchase-orders/%d/cancel", poID), map[string]any{}},
	} {
		w, _ = api.doReq(request.method, request.path, request.body, cashierToken)
		if w.Code != http.StatusForbidden {
			t.Fatalf("cashier %s %s: %d %s", request.method, request.path, w.Code, w.Body.String())
		}
	}

	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/order", poID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("order PO: %d %s", w.Code, w.Body.String())
	}
	w, partial := api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", poID), map[string]any{
		"items": []any{
			map[string]any{"itemId": productItemID, "qty": 2},
			map[string]any{"itemId": variantItemID, "qty": 1},
		},
	}, adminToken)
	if w.Code != http.StatusOK || partial["status"] != "partially_received" {
		t.Fatalf("partial receipt: %d %s", w.Code, w.Body.String())
	}
	productStock, _ := purchaseTestStock(t, api, adminToken, productID, 0)
	_, variantStock := purchaseTestStock(t, api, adminToken, variantProductID, variantID)
	if productStock != 5 || variantStock != 3 {
		t.Fatalf("stock after partial receipt: product=%d variant=%d", productStock, variantStock)
	}

	w, full := api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", poID), map[string]any{
		"items": []any{
			map[string]any{"itemId": productItemID, "qty": 3},
			map[string]any{"itemId": variantItemID, "qty": 3},
		},
	}, adminToken)
	if w.Code != http.StatusOK || full["status"] != "fully_received" {
		t.Fatalf("final receipt: %d %s", w.Code, w.Body.String())
	}
	productStock, _ = purchaseTestStock(t, api, adminToken, productID, 0)
	_, variantStock = purchaseTestStock(t, api, adminToken, variantProductID, variantID)
	if productStock != 8 || variantStock != 6 {
		t.Fatalf("stock after full receipt: product=%d variant=%d", productStock, variantStock)
	}

	w, movements := api.doReq(http.MethodGet, "/movements?type=purchase", nil, adminToken)
	rows, ok := movements["items"].([]any)
	if w.Code != http.StatusOK || !ok || len(rows) != 4 {
		t.Fatalf("purchase movements: %d %s", w.Code, w.Body.String())
	}
	for _, raw := range rows {
		movement := raw.(map[string]any)
		if movement["type"] != "purchase" || movement["qty"].(float64) <= 0 || movement["actor"] == "" || !strings.Contains(movement["reason"].(string), created["number"].(string)) {
			t.Fatalf("purchase movement fields: %v", movement)
		}
	}
	w, detail = api.doReq(http.MethodGet, fmt.Sprintf("/purchase-orders/%d", poID), nil, adminToken)
	itemsByProduct = purchaseOrderItemsByProduct(detail["items"])
	if itemsByProduct[productID]["qty_received"] != float64(5) || itemsByProduct[variantProductID]["qty_received"] != float64(4) {
		t.Fatalf("received quantities: %s", w.Body.String())
	}
}

func TestPurchaseOrderOverflowReceiptRollsBack(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko PO Overflow")
	supplierID := createPurchaseTestSupplier(t, api, adminToken, "Supplier Overflow")
	firstProductID := createPurchaseTestProduct(t, api, adminToken, "Barang Aman", "PO-SAFE", 3)
	secondProductID := createPurchaseTestProduct(t, api, adminToken, "Barang Lebih", "PO-OVER", 8)
	w, created := api.doReq(http.MethodPost, "/purchase-orders", map[string]any{
		"supplierId": supplierID,
		"items": []any{
			map[string]any{"productId": firstProductID, "qty": 2, "buyPrice": 100},
			map[string]any{"productId": secondProductID, "qty": 2, "buyPrice": 200},
		},
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create PO: %d %s", w.Code, w.Body.String())
	}
	poID := uint(created["id"].(float64))
	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/purchase-orders/%d", poID), nil, adminToken)
	itemsByProduct := purchaseOrderItemsByProduct(detail["items"])
	firstItemID := uint(itemsByProduct[firstProductID]["id"].(float64))
	secondItemID := uint(itemsByProduct[secondProductID]["id"].(float64))
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/order", poID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("order PO: %d %s", w.Code, w.Body.String())
	}
	w, partial := api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", poID), map[string]any{
		"items": []any{
			map[string]any{"itemId": firstItemID, "qty": 1},
			map[string]any{"itemId": secondItemID, "qty": 1},
		},
	}, adminToken)
	if w.Code != http.StatusOK || partial["status"] != "partially_received" {
		t.Fatalf("initial partial receipt: %d %s", w.Code, w.Body.String())
	}

	w, response := api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", poID), map[string]any{
		"items": []any{
			map[string]any{"itemId": firstItemID, "qty": 1},
			map[string]any{"itemId": secondItemID, "qty": 2},
		},
	}, adminToken)
	if w.Code != http.StatusBadRequest || response["error"] == "" {
		t.Fatalf("overflow receipt must be rejected: %d %s", w.Code, w.Body.String())
	}
	firstStock, _ := purchaseTestStock(t, api, adminToken, firstProductID, 0)
	secondStock, _ := purchaseTestStock(t, api, adminToken, secondProductID, 0)
	if firstStock != 4 || secondStock != 9 {
		t.Fatalf("overflow receipt changed stock: first=%d second=%d", firstStock, secondStock)
	}
	w, detail = api.doReq(http.MethodGet, fmt.Sprintf("/purchase-orders/%d", poID), nil, adminToken)
	itemsByProduct = purchaseOrderItemsByProduct(detail["items"])
	if detail["status"] != "partially_received" || itemsByProduct[firstProductID]["qty_received"] != float64(1) || itemsByProduct[secondProductID]["qty_received"] != float64(1) {
		t.Fatalf("overflow receipt was not rolled back: %d %s", w.Code, w.Body.String())
	}
	w, movements := api.doReq(http.MethodGet, "/movements?type=purchase", nil, adminToken)
	if w.Code != http.StatusOK || len(movements["items"].([]any)) != 2 {
		t.Fatalf("overflow receipt created movements: %d %s", w.Code, w.Body.String())
	}
}

func TestPurchaseOrderReceiptRejectsChangedCatalogTargets(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko PO Katalog Berubah")
	supplierID := createPurchaseTestSupplier(t, api, adminToken, "Supplier Katalog Berubah")

	baseProductID := createPurchaseTestProduct(t, api, adminToken, "Produk Awal", "PO-BASE-CHANGED", 3)
	w, basePO := api.doReq(http.MethodPost, "/purchase-orders", map[string]any{
		"supplierId": supplierID,
		"items":      []any{map[string]any{"productId": baseProductID, "qty": 2, "buyPrice": 250}},
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create product PO: %d %s", w.Code, w.Body.String())
	}
	basePOID := uint(basePO["id"].(float64))
	baseItemID := uint(basePO["items"].([]any)[0].(map[string]any)["id"].(float64))
	variantID := createPurchaseTestVariant(t, api, adminToken, baseProductID, "Baru", "PO-BASE-NEW-VARIANT", 0)
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/order", basePOID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("order product PO: %d %s", w.Code, w.Body.String())
	}
	w, response := api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", basePOID), map[string]any{"items": []any{map[string]any{"itemId": baseItemID, "qty": 1}}}, adminToken)
	if w.Code != http.StatusBadRequest || response["error"] == "" {
		t.Fatalf("receipt must reject a product that now requires a variant: %d %s", w.Code, w.Body.String())
	}
	baseStock, variantStock := purchaseTestStock(t, api, adminToken, baseProductID, variantID)
	if baseStock != 3 || variantStock != 0 {
		t.Fatalf("changed product target stock: base=%d variant=%d", baseStock, variantStock)
	}

	variantProductID := createPurchaseTestProduct(t, api, adminToken, "Varian Dihapus", "PO-VARIANT-REMOVED", 0)
	deletedVariantID := createPurchaseTestVariant(t, api, adminToken, variantProductID, "Biru", "PO-VARIANT-REMOVED-BLUE", 2)
	w, variantPO := api.doReq(http.MethodPost, "/purchase-orders", map[string]any{
		"supplierId": supplierID,
		"items":      []any{map[string]any{"productId": variantProductID, "variantId": deletedVariantID, "qty": 2, "buyPrice": 300}},
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create variant PO: %d %s", w.Code, w.Body.String())
	}
	variantPOID := uint(variantPO["id"].(float64))
	variantItemID := uint(variantPO["items"].([]any)[0].(map[string]any)["id"].(float64))
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/order", variantPOID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("order variant PO: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/products/%d/variants/%d", variantProductID, deletedVariantID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("delete ordered variant: %d %s", w.Code, w.Body.String())
	}
	w, response = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", variantPOID), map[string]any{"items": []any{map[string]any{"itemId": variantItemID, "qty": 1}}}, adminToken)
	if w.Code != http.StatusBadRequest || response["error"] == "" {
		t.Fatalf("receipt must reject a deleted variant: %d %s", w.Code, w.Body.String())
	}

	deletedProductID := createPurchaseTestProduct(t, api, adminToken, "Produk Dihapus", "PO-PRODUCT-REMOVED", 1)
	w, deletedProductPO := api.doReq(http.MethodPost, "/purchase-orders", map[string]any{
		"supplierId": supplierID,
		"items":      []any{map[string]any{"productId": deletedProductID, "qty": 2, "buyPrice": 350}},
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create PO for product to delete: %d %s", w.Code, w.Body.String())
	}
	deletedProductPOID := uint(deletedProductPO["id"].(float64))
	deletedProductItemID := uint(deletedProductPO["items"].([]any)[0].(map[string]any)["id"].(float64))
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/order", deletedProductPOID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("order PO for product to delete: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/products/%d", deletedProductID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("delete ordered product: %d %s", w.Code, w.Body.String())
	}
	w, response = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", deletedProductPOID), map[string]any{"items": []any{map[string]any{"itemId": deletedProductItemID, "qty": 1}}}, adminToken)
	if w.Code != http.StatusBadRequest || response["error"] == "" {
		t.Fatalf("receipt must reject a deleted product: %d %s", w.Code, w.Body.String())
	}

	deletedParentID := createPurchaseTestProduct(t, api, adminToken, "Induk Varian Dihapus", "PO-PARENT-REMOVED", 0)
	deletedParentVariantID := createPurchaseTestVariant(t, api, adminToken, deletedParentID, "Merah", "PO-PARENT-REMOVED-RED", 1)
	w, deletedParentPO := api.doReq(http.MethodPost, "/purchase-orders", map[string]any{
		"supplierId": supplierID,
		"items":      []any{map[string]any{"productId": deletedParentID, "variantId": deletedParentVariantID, "qty": 2, "buyPrice": 400}},
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create PO for variant whose product will be deleted: %d %s", w.Code, w.Body.String())
	}
	deletedParentPOID := uint(deletedParentPO["id"].(float64))
	deletedParentItemID := uint(deletedParentPO["items"].([]any)[0].(map[string]any)["id"].(float64))
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/order", deletedParentPOID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("order PO for deleted parent variant: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/products/%d", deletedParentID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("delete parent product: %d %s", w.Code, w.Body.String())
	}
	w, response = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", deletedParentPOID), map[string]any{"items": []any{map[string]any{"itemId": deletedParentItemID, "qty": 1}}}, adminToken)
	if w.Code != http.StatusBadRequest || response["error"] == "" {
		t.Fatalf("receipt must reject a deleted product's variant: %d %s", w.Code, w.Body.String())
	}

	for _, poID := range []uint{basePOID, variantPOID, deletedProductPOID, deletedParentPOID} {
		w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/purchase-orders/%d", poID), nil, adminToken)
		if w.Code != http.StatusOK || detail["status"] != "ordered" || detail["items"].([]any)[0].(map[string]any)["qty_received"] != float64(0) {
			t.Fatalf("changed catalog target must remain unreceived for PO %d: %d %s", poID, w.Code, w.Body.String())
		}
	}
	w, movements := api.doReq(http.MethodGet, "/movements?type=purchase", nil, adminToken)
	if w.Code != http.StatusOK || len(movements["items"].([]any)) != 0 {
		t.Fatalf("rejected catalog receipts created purchase movements: %d %s", w.Code, w.Body.String())
	}
}

func TestPurchaseOrderReceiptRollsBackAfterMovementFailure(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko PO Error Movement")
	supplierID := createPurchaseTestSupplier(t, api, adminToken, "Supplier Error Movement")
	firstProductID := createPurchaseTestProduct(t, api, adminToken, "Barang Pertama", "PO-MOVE-FIRST", 3)
	secondProductID := createPurchaseTestProduct(t, api, adminToken, "Barang Kedua", "PO-MOVE-SECOND", 8)
	w, created := api.doReq(http.MethodPost, "/purchase-orders", map[string]any{
		"supplierId": supplierID,
		"items": []any{
			map[string]any{"productId": firstProductID, "qty": 2, "buyPrice": 100},
			map[string]any{"productId": secondProductID, "qty": 2, "buyPrice": 200},
		},
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create PO: %d %s", w.Code, w.Body.String())
	}
	poID := uint(created["id"].(float64))
	itemsByProduct := purchaseOrderItemsByProduct(created["items"])
	firstItemID := uint(itemsByProduct[firstProductID]["id"].(float64))
	secondItemID := uint(itemsByProduct[secondProductID]["id"].(float64))
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/order", poID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("order PO: %d %s", w.Code, w.Body.String())
	}

	database, err := gorm.Open(sqlite.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("get test database connection: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	trigger := fmt.Sprintf(`CREATE TRIGGER fail_second_purchase_movement BEFORE INSERT ON stock_movements
		WHEN NEW.type = 'purchase' AND NEW.product_id = %d
		BEGIN SELECT RAISE(ABORT, 'forced purchase movement failure'); END`, secondProductID)
	if err := database.Exec(trigger).Error; err != nil {
		t.Fatalf("create movement failure trigger: %v", err)
	}

	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", poID), map[string]any{
		"items": []any{
			map[string]any{"itemId": firstItemID, "qty": 1},
			map[string]any{"itemId": secondItemID, "qty": 1},
		},
	}, adminToken)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("movement failure should fail receipt: %d %s", w.Code, w.Body.String())
	}
	firstStock, _ := purchaseTestStock(t, api, adminToken, firstProductID, 0)
	secondStock, _ := purchaseTestStock(t, api, adminToken, secondProductID, 0)
	if firstStock != 3 || secondStock != 8 {
		t.Fatalf("movement failure changed stock: first=%d second=%d", firstStock, secondStock)
	}
	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/purchase-orders/%d", poID), nil, adminToken)
	itemsByProduct = purchaseOrderItemsByProduct(detail["items"])
	if w.Code != http.StatusOK || detail["status"] != "ordered" || itemsByProduct[firstProductID]["qty_received"] != float64(0) || itemsByProduct[secondProductID]["qty_received"] != float64(0) {
		t.Fatalf("movement failure did not roll back PO quantities/status: %d %s", w.Code, w.Body.String())
	}
	w, movements := api.doReq(http.MethodGet, "/movements?type=purchase", nil, adminToken)
	if w.Code != http.StatusOK || len(movements["items"].([]any)) != 0 {
		t.Fatalf("movement failure left purchase movements: %d %s", w.Code, w.Body.String())
	}
}

func TestPurchaseOrderCancelAndInvalidReceiptStates(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko PO Batal")
	supplierID := createPurchaseTestSupplier(t, api, adminToken, "Supplier Batal")
	productID := createPurchaseTestProduct(t, api, adminToken, "Barang Batal", "PO-CANCEL", 7)
	createPO := func() uint {
		t.Helper()
		w, po := api.doReq(http.MethodPost, "/purchase-orders", map[string]any{
			"supplierId": supplierID,
			"items":      []any{map[string]any{"productId": productID, "qty": 2, "buyPrice": 250}},
		}, adminToken)
		if w.Code != http.StatusCreated {
			t.Fatalf("create PO: %d %s", w.Code, w.Body.String())
		}
		return uint(po["id"].(float64))
	}
	getItemID := func(poID uint) uint {
		t.Helper()
		w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/purchase-orders/%d", poID), nil, adminToken)
		if w.Code != http.StatusOK {
			t.Fatalf("get PO: %d %s", w.Code, w.Body.String())
		}
		return uint(detail["items"].([]any)[0].(map[string]any)["id"].(float64))
	}
	stockBefore, _ := purchaseTestStock(t, api, adminToken, productID, 0)

	draftID := createPO()
	draftItemID := getItemID(draftID)
	w, _ := api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", draftID), map[string]any{"items": []any{map[string]any{"itemId": draftItemID, "qty": 1}}}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("receive draft PO: %d %s", w.Code, w.Body.String())
	}
	w, cancelledDraft := api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/cancel", draftID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK || cancelledDraft["status"] != "cancelled" {
		t.Fatalf("cancel draft PO: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", draftID), map[string]any{"items": []any{map[string]any{"itemId": draftItemID, "qty": 1}}}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("receive cancelled draft PO: %d %s", w.Code, w.Body.String())
	}

	orderedID := createPO()
	orderedItemID := getItemID(orderedID)
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/order", orderedID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("order PO before cancel: %d %s", w.Code, w.Body.String())
	}
	w, cancelledOrdered := api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/cancel", orderedID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK || cancelledOrdered["status"] != "cancelled" {
		t.Fatalf("cancel ordered PO: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", orderedID), map[string]any{"items": []any{map[string]any{"itemId": orderedItemID, "qty": 1}}}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("receive cancelled ordered PO: %d %s", w.Code, w.Body.String())
	}

	partialID := createPO()
	partialItemID := getItemID(partialID)
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/order", partialID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("order PO before partial receipt: %d %s", w.Code, w.Body.String())
	}
	w, partial := api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", partialID), map[string]any{"items": []any{map[string]any{"itemId": partialItemID, "qty": 1}}}, adminToken)
	if w.Code != http.StatusOK || partial["status"] != "partially_received" {
		t.Fatalf("partial receipt: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/cancel", partialID), map[string]any{}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("cancel partially received PO: %d %s", w.Code, w.Body.String())
	}
	w, full := api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", partialID), map[string]any{"items": []any{map[string]any{"itemId": partialItemID, "qty": 1}}}, adminToken)
	if w.Code != http.StatusOK || full["status"] != "fully_received" {
		t.Fatalf("complete PO receipt: %d %s", w.Code, w.Body.String())
	}
	for _, request := range []struct {
		path string
		body any
	}{
		{fmt.Sprintf("/purchase-orders/%d/receive", partialID), map[string]any{"items": []any{map[string]any{"itemId": partialItemID, "qty": 1}}}},
		{fmt.Sprintf("/purchase-orders/%d/cancel", partialID), map[string]any{}},
	} {
		w, _ = api.doReq(http.MethodPost, request.path, request.body, adminToken)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("invalid action for fully received PO %s: %d %s", request.path, w.Code, w.Body.String())
		}
	}
	stockAfter, _ := purchaseTestStock(t, api, adminToken, productID, 0)
	if stockAfter != stockBefore+2 {
		t.Fatalf("cancelled POs changed stock: before=%d after=%d", stockBefore, stockAfter)
	}
}
