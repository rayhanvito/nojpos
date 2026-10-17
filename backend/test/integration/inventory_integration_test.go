package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func createInventoryProduct(t *testing.T, api *promoTestAPI, token, name, sku string, sellPrice int64, stock, minStock int) uint {
	t.Helper()
	w, product := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": name, "sku": sku, "buyPrice": 300, "sellPrice": sellPrice, "stock": stock, "minStock": minStock,
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create product %q: %d %s", name, w.Code, w.Body.String())
	}
	return uint(product["id"].(float64))
}

func TestInventoryCatalogNestedCRUDAndStoreScope(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Inventori Satu")
	cashierToken := api.createCashier(adminToken)
	otherAdminToken := api.registerAdmin("Toko Inventori Dua")

	productID := createInventoryProduct(t, api, adminToken, "Kemeja", "INV-KEMEJA", 1000, 20, 4)
	otherProductID := createInventoryProduct(t, api, otherAdminToken, "Celana", "INV-CELANA", 1000, 20, 0)
	createInventoryProduct(t, api, adminToken, "Cendol", "INV-SHARED", 1000, 10, 0)

	w, createdVariant := api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/variants", productID), map[string]any{
		"name": "Merah M", "sku": "INV-KEMEJA-M", "buyPrice": 400, "sellPrice": 1200, "stock": 8, "minStock": 2,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create variant: %d %s", w.Code, w.Body.String())
	}
	variantID := uint(createdVariant["id"].(float64))
	if createdVariant["name"] != "Merah M" || createdVariant["sku"] != "INV-KEMEJA-M" || createdVariant["min_stock"].(float64) != 2 {
		t.Fatalf("variant response fields: %v", createdVariant)
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/variants", productID), map[string]any{
		"name": "SKU Produk", "sku": "INV-SHARED", "buyPrice": 1, "sellPrice": 1, "stock": 1,
	}, adminToken)
	if w.Code != http.StatusConflict {
		t.Fatalf("variant SKU must be unique across the store catalog: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk duplikat varian", "sku": "INV-KEMEJA-M", "sellPrice": 100, "stock": 1,
	}, adminToken)
	if w.Code != http.StatusConflict {
		t.Fatalf("product SKU cannot reuse a variant SKU: %d %s", w.Code, w.Body.String())
	}

	w, variantList := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d/variants", productID), nil, cashierToken)
	if w.Code != http.StatusOK || len(variantList["items"].([]any)) != 1 {
		t.Fatalf("cashier may read variant list: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/variants", productID), map[string]any{
		"name": "Kasir Tidak Boleh", "sku": "INV-NO", "sellPrice": 1, "stock": 1,
	}, cashierToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("cashier cannot create variants: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodGet, fmt.Sprintf("/products/%d/variants", otherProductID), nil, adminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("foreign product variants are hidden: %d %s", w.Code, w.Body.String())
	}

	w, productTier := api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/price-tiers", productID), map[string]any{
		"minQty": 10, "sellPrice": 900,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create product tier: %d %s", w.Code, w.Body.String())
	}
	productTierID := uint(productTier["id"].(float64))
	w, variantTier := api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/variants/%d/price-tiers", productID, variantID), map[string]any{
		"minQty": 5, "sellPrice": 1000,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create variant tier: %d %s", w.Code, w.Body.String())
	}
	variantTierID := uint(variantTier["id"].(float64))

	w, productTiers := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d/price-tiers", productID), nil, cashierToken)
	if w.Code != http.StatusOK || len(productTiers["items"].([]any)) != 1 {
		t.Fatalf("read product tiers: %d %s", w.Code, w.Body.String())
	}
	w, variantTiers := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d/variants/%d/price-tiers", productID, variantID), nil, cashierToken)
	if w.Code != http.StatusOK || len(variantTiers["items"].([]any)) != 1 {
		t.Fatalf("read variant tiers: %d %s", w.Code, w.Body.String())
	}

	w, updatedVariant := api.doReq(http.MethodPut, fmt.Sprintf("/products/%d/variants/%d", productID, variantID), map[string]any{
		"name": "Merah L", "sku": "INV-KEMEJA-L", "buyPrice": 450, "sellPrice": 1300, "stock": 9, "minStock": 3,
	}, adminToken)
	if w.Code != http.StatusOK || updatedVariant["name"] != "Merah L" || updatedVariant["stock"].(float64) != 9 {
		t.Fatalf("update variant: %d %s", w.Code, w.Body.String())
	}
	w, updatedTier := api.doReq(http.MethodPut, fmt.Sprintf("/products/%d/price-tiers/%d", productID, productTierID), map[string]any{
		"minQty": 12, "sellPrice": 850,
	}, adminToken)
	if w.Code != http.StatusOK || updatedTier["min_qty"].(float64) != 12 || updatedTier["sell_price"].(float64) != 850 {
		t.Fatalf("update product tier: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/products/%d/variants/%d/price-tiers/%d", productID, variantID, variantTierID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("delete variant tier: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/products/%d/variants/%d", productID, variantID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("delete variant: %d %s", w.Code, w.Body.String())
	}
	w, deletedVariants := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d/variants", productID), nil, cashierToken)
	if w.Code != http.StatusOK || len(deletedVariants["items"].([]any)) != 0 {
		t.Fatalf("deleted variant should be hidden: %d %s", w.Code, w.Body.String())
	}

	w, productDetail := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d", productID), nil, cashierToken)
	if w.Code != http.StatusOK || len(productDetail["variants"].([]any)) != 0 || len(productDetail["price_tiers"].([]any)) != 1 || productDetail["min_stock"].(float64) != 4 {
		t.Fatalf("product detail embeds variants, tiers, and threshold: %d %s", w.Code, w.Body.String())
	}
	w, productPage := api.doReq(http.MethodGet, "/products", nil, cashierToken)
	items := productPage["items"].([]any)
	var listed map[string]any
	for _, item := range items {
		if uint(item.(map[string]any)["id"].(float64)) == productID {
			listed = item.(map[string]any)
		}
	}
	if w.Code != http.StatusOK || listed == nil || len(listed["variants"].([]any)) != 0 || len(listed["price_tiers"].([]any)) != 1 {
		t.Fatalf("product list embeds inventory data: %d %s", w.Code, w.Body.String())
	}
}

func TestCheckoutUsesVariantStockAndQuantityTiers(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Checkout Inventori")
	otherAdminToken := api.registerAdmin("Toko Checkout Inventori Lain")

	variantProductID := createInventoryProduct(t, api, adminToken, "Jaket", "INV-JAKET", 1000, 30, 0)
	w, _ := api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/price-tiers", variantProductID), map[string]any{"minQty": 1, "sellPrice": 100}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create product tier: %d %s", w.Code, w.Body.String())
	}
	w, variant := api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/variants", variantProductID), map[string]any{
		"name": "Hitam L", "sku": "INV-JAKET-L", "buyPrice": 350, "sellPrice": 1200, "stock": 10,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create checkout variant: %d %s", w.Code, w.Body.String())
	}
	variantID := uint(variant["id"].(float64))
	otherProductID := createInventoryProduct(t, api, adminToken, "Celana", "INV-CELANA", 1000, 10, 0)
	w, otherVariant := api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/variants", otherProductID), map[string]any{
		"name": "Biru", "sku": "INV-CELANA-B", "sellPrice": 1000, "stock": 10,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create other product variant: %d %s", w.Code, w.Body.String())
	}
	foreignProductID := createInventoryProduct(t, api, otherAdminToken, "Sepatu", "INV-SEPATU", 1000, 10, 0)
	w, foreignVariant := api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/variants", foreignProductID), map[string]any{
		"name": "Putih", "sku": "INV-SEPATU-P", "sellPrice": 1000, "stock": 10,
	}, otherAdminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create foreign store variant: %d %s", w.Code, w.Body.String())
	}
	for _, invalidVariantID := range []uint{uint(otherVariant["id"].(float64)), uint(foreignVariant["id"].(float64))} {
		w, invalidVariantCheckout := api.doReq(http.MethodPost, "/transactions", map[string]any{
			"items": []any{map[string]any{"productId": variantProductID, "variantId": invalidVariantID, "qty": 1}}, "method": "QRIS",
		}, adminToken)
		if w.Code != http.StatusBadRequest || invalidVariantCheckout["error"] == nil {
			t.Fatalf("foreign or wrong-parent variant rejected: %d %s", w.Code, w.Body.String())
		}
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/variants/%d/price-tiers", variantProductID, variantID), map[string]any{
		"minQty": 2, "sellPrice": 900,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create variant tier: %d %s", w.Code, w.Body.String())
	}

	checkout := func(items []any, paid int64) (*httptest.ResponseRecorder, map[string]any) {
		t.Helper()
		return api.doReq(http.MethodPost, "/transactions", map[string]any{"items": items, "method": "Cash", "paid": paid}, adminToken)
	}
	w, missingVariant := checkout([]any{map[string]any{"productId": variantProductID, "qty": 1}}, 2000)
	if w.Code != http.StatusBadRequest || missingVariant["error"] == nil {
		t.Fatalf("variant selection is required: %d %s", w.Code, w.Body.String())
	}
	w, twoVariantUnits := checkout([]any{map[string]any{"productId": variantProductID, "variantId": variantID, "qty": 2}}, 1800)
	if w.Code != http.StatusCreated {
		t.Fatalf("checkout variant: %d %s", w.Code, w.Body.String())
	}
	items := twoVariantUnits["items"].([]any)
	item := items[0].(map[string]any)
	if twoVariantUnits["subtotal"].(float64) != 1800 || item["price"].(float64) != 900 || uint(item["variant_id"].(float64)) != variantID || item["buy_price"].(float64) != 350 {
		t.Fatalf("variant tier price snapshot: %v", twoVariantUnits)
	}
	w, productAfter := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d", variantProductID), nil, adminToken)
	if w.Code != http.StatusOK || productAfter["stock"].(float64) != 30 || productAfter["variants"].([]any)[0].(map[string]any)["stock"].(float64) != 8 {
		t.Fatalf("variant checkout stock isolation: %d %s", w.Code, w.Body.String())
	}
	w, movements := api.doReq(http.MethodGet, fmt.Sprintf("/movements?productId=%d", variantProductID), nil, adminToken)
	movementItems := movements["items"].([]any)
	firstMovement := movementItems[0].(map[string]any)
	if w.Code != http.StatusOK || firstMovement["type"] != "sale" || firstMovement["qty"].(float64) != -2 || uint(firstMovement["variant_id"].(float64)) != variantID {
		t.Fatalf("variant sale movement: %d %s", w.Code, w.Body.String())
	}
	w, variantBase := checkout([]any{map[string]any{"productId": variantProductID, "variantId": variantID, "qty": 1}}, 1200)
	if w.Code != http.StatusCreated || variantBase["items"].([]any)[0].(map[string]any)["price"].(float64) != 1200 {
		t.Fatalf("variant base price must ignore product tier: %d %s", w.Code, w.Body.String())
	}

	productID := createInventoryProduct(t, api, adminToken, "Kopi Grosir", "INV-KOPI", 1000, 40, 0)
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/price-tiers", productID), map[string]any{"minQty": 10, "sellPrice": 800}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create bulk tier: %d %s", w.Code, w.Body.String())
	}
	promo := validPromoRequest("Promo sesudah tier", "TIER-PROMO")
	promo["type"] = "percent"
	promo["value"] = 10
	w, _ = api.doReq(http.MethodPost, "/promos", promo, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create checkout promo: %d %s", w.Code, w.Body.String())
	}
	w, belowTier := checkout([]any{map[string]any{"productId": productID, "qty": 9}}, 9000)
	if w.Code != http.StatusCreated || belowTier["items"].([]any)[0].(map[string]any)["price"].(float64) != 1000 {
		t.Fatalf("qty below tier uses base product price: %d %s", w.Code, w.Body.String())
	}
	w, invalidZeroVariant := checkout([]any{
		map[string]any{"productId": productID, "qty": 1},
		map[string]any{"productId": productID, "variantId": 0, "qty": 1},
	}, 2000)
	if w.Code != http.StatusBadRequest || invalidZeroVariant["error"] == nil {
		t.Fatalf("explicit zero variant ID must not merge into a product-only line: %d %s", w.Code, w.Body.String())
	}
	w, tierPromo := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 5}, map[string]any{"productId": productID, "qty": 5}}, "method": "Cash", "paid": 7200, "voucherCode": "TIER-PROMO",
	}, adminToken)
	if w.Code != http.StatusCreated || tierPromo["subtotal"].(float64) != 8000 || tierPromo["voucher_discount"].(float64) != 800 || tierPromo["total"].(float64) != 7200 || tierPromo["items"].([]any)[0].(map[string]any)["price"].(float64) != 800 {
		t.Fatalf("tier price is applied before promo: %d %s", w.Code, w.Body.String())
	}
	w, promoPreview := api.doReq(http.MethodPost, "/promos/preview", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 5}, map[string]any{"productId": productID, "qty": 5}}, "voucherCode": "TIER-PROMO",
	}, adminToken)
	if w.Code != http.StatusOK || promoPreview["voucher_discount"].(float64) != 800 {
		t.Fatalf("promo preview uses the same aggregated tier price: %d %s", w.Code, w.Body.String())
	}
	productWithoutInventory := createInventoryProduct(t, api, adminToken, "Produk Lama", "INV-LAMA", 700, 5, 0)
	w, legacyCheckout := checkout([]any{map[string]any{"productId": productWithoutInventory, "qty": 1}}, 700)
	if w.Code != http.StatusCreated || legacyCheckout["subtotal"].(float64) != 700 || legacyCheckout["items"].([]any)[0].(map[string]any)["price"].(float64) != 700 {
		t.Fatalf("legacy checkout without variants or tiers: %d %s", w.Code, w.Body.String())
	}
}

func TestCheckoutUsesParentInventoryAfterAllVariantsAreDeleted(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Varian Terhapus")
	productID := createInventoryProduct(t, api, adminToken, "Produk Induk", "INV-INDUK-HAPUS", 4500, 7, 8)
	w, variant := api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/variants", productID), map[string]any{
		"name": "Varian Tunggal", "sku": "INV-INDUK-HAPUS-V1", "sellPrice": 6000, "stock": 2,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create variant: %d %s", w.Code, w.Body.String())
	}
	variantID := uint(variant["id"].(float64))

	checkoutBody := map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 2}}, "method": "Cash", "paid": 9000,
	}
	w, missingVariant := api.doReq(http.MethodPost, "/transactions", checkoutBody, adminToken)
	if w.Code != http.StatusBadRequest || missingVariant["error"] == nil {
		t.Fatalf("an active variant still requires selection: %d %s", w.Code, w.Body.String())
	}

	w, _ = api.doReq(http.MethodDelete, fmt.Sprintf("/products/%d/variants/%d", productID, variantID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("delete variant: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, "/promos/preview", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 2}},
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("promo preview should use parent inventory after variant deletion: %d %s", w.Code, w.Body.String())
	}
	w, dashboard := api.doReq(http.MethodGet, "/dashboard", nil, adminToken)
	if w.Code != http.StatusOK || dashboard["today"].(map[string]any)["low_stock"].(float64) != 1 {
		t.Fatalf("dashboard should count parent stock after variant deletion: %d %s", w.Code, w.Body.String())
	}
	w, checkout := api.doReq(http.MethodPost, "/transactions", checkoutBody, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("checkout should use parent inventory after variant deletion: %d %s", w.Code, w.Body.String())
	}
	items := checkout["items"].([]any)
	item := items[0].(map[string]any)
	if checkout["subtotal"].(float64) != 9000 || item["price"].(float64) != 4500 || item["variant_id"] != nil {
		t.Fatalf("checkout should use parent price without a variant snapshot: %v", checkout)
	}
	w, productAfter := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d", productID), nil, adminToken)
	if w.Code != http.StatusOK || productAfter["stock"].(float64) != 5 || len(productAfter["variants"].([]any)) != 0 {
		t.Fatalf("checkout should reduce parent stock after deletion: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, "/stock/adjustments", map[string]any{
		"productId": productID, "direction": "minus", "qty": 1, "reason": "Penyesuaian setelah varian dihapus",
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("parent stock adjustment should work after variant deletion: %d %s", w.Code, w.Body.String())
	}
}

func TestVariantRefundRestoresSameVariantAndMovement(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Refund Varian")
	productID := createInventoryProduct(t, api, adminToken, "Kaos", "INV-KAOS", 1000, 15, 0)
	w, variant := api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/variants", productID), map[string]any{
		"name": "Putih XL", "sku": "INV-KAOS-XL", "buyPrice": 400, "sellPrice": 1500, "stock": 6,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create refundable variant: %d %s", w.Code, w.Body.String())
	}
	variantID := uint(variant["id"].(float64))
	w, trx := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "variantId": variantID, "qty": 2}}, "method": "QRIS",
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("checkout refundable variant: %d %s", w.Code, w.Body.String())
	}
	trxID := uint(trx["id"].(float64))
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/transactions/%d/refund", trxID), map[string]any{
		"items": []any{map[string]any{"productId": productID, "variantId": variantID, "qty": 1}}, "reason": "Ukuran tidak sesuai",
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("refund variant: %d %s", w.Code, w.Body.String())
	}
	w, product := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d", productID), nil, adminToken)
	variants := product["variants"].([]any)
	if w.Code != http.StatusOK || product["stock"].(float64) != 15 || variants[0].(map[string]any)["stock"].(float64) != 5 {
		t.Fatalf("refund must restore only the selected variant: %d %s", w.Code, w.Body.String())
	}
	w, movements := api.doReq(http.MethodGet, fmt.Sprintf("/movements?productId=%d", productID), nil, adminToken)
	firstMovement := movements["items"].([]any)[0].(map[string]any)
	if w.Code != http.StatusOK || firstMovement["type"] != "refund" || firstMovement["qty"].(float64) != 1 || uint(firstMovement["variant_id"].(float64)) != variantID {
		t.Fatalf("refund movement keeps variant identity: %d %s", w.Code, w.Body.String())
	}
	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/transactions/%d", trxID), nil, adminToken)
	refundItems := detail["refunds"].([]any)[0].(map[string]any)["items"].([]any)
	refunded := refundItems[0].(map[string]any)
	if w.Code != http.StatusOK || refunded["qty"].(float64) != 1 || uint(refunded["variant_id"].(float64)) != variantID || refunded["variant_name"] != "Putih XL" {
		t.Fatalf("refund detail identifies the sold variant: %d %s", w.Code, w.Body.String())
	}
	w, dashboard := api.doReq(http.MethodGet, "/dashboard", nil, adminToken)
	today := dashboard["today"].(map[string]any)
	if w.Code != http.StatusOK || today["refund_total"].(float64) != 1500 || today["omzet"].(float64) != 1500 || today["items_sold"].(float64) != 1 {
		t.Fatalf("dashboard net totals include variant refunds: %d %s", w.Code, w.Body.String())
	}
}

func TestLowStockAlertsUseThresholdAndDeduplicateDaily(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Alert Stok")
	w, settings := api.doReq(http.MethodPut, "/settings", map[string]any{"lowStockDefault": 6}, adminToken)
	if w.Code != http.StatusOK || settings["lowStockDefault"].(float64) != 6 {
		t.Fatalf("save low stock suggestion: %d %s", w.Code, w.Body.String())
	}
	w, settings = api.doReq(http.MethodGet, "/settings", nil, adminToken)
	if w.Code != http.StatusOK || settings["lowStockDefault"].(float64) != 6 {
		t.Fatalf("read low stock suggestion: %d %s", w.Code, w.Body.String())
	}
	w, invalidSettings := api.doReq(http.MethodPut, "/settings", map[string]any{"lowStockDefault": -1}, adminToken)
	if w.Code != http.StatusBadRequest || !strings.Contains(invalidSettings["error"].(string), "tidak boleh negatif") {
		t.Fatalf("negative low stock suggestion gets a clear validation message: %d %s", w.Code, w.Body.String())
	}

	productID := createInventoryProduct(t, api, adminToken, "Minyak", "INV-MINYAK", 1000, 4, 2)
	zeroThresholdID := createInventoryProduct(t, api, adminToken, "Sabun", "INV-SABUN", 1000, 1, 0)
	adjustmentID := createInventoryProduct(t, api, adminToken, "Gula", "INV-GULA", 1000, 5, 3)
	variantProductID := createInventoryProduct(t, api, adminToken, "Sepatu", "INV-SEPATU-ALERT", 1000, 5, 4)
	w, variant := api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/variants", variantProductID), map[string]any{
		"name": "Hitam 42", "sku": "INV-SEPATU-42", "sellPrice": 1500, "stock": 3, "minStock": 2,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create low stock variant: %d %s", w.Code, w.Body.String())
	}
	variantID := uint(variant["id"].(float64))
	w, parentAdjustment := api.doReq(http.MethodPost, "/stock/adjustments", map[string]any{
		"productId": variantProductID, "direction": "minus", "qty": 1, "reason": "Uji tanpa varian",
	}, adminToken)
	if w.Code != http.StatusBadRequest || parentAdjustment["error"] == nil {
		t.Fatalf("variant-backed product stock requires a variant: %d %s", w.Code, w.Body.String())
	}
	w, unchangedParent := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d", variantProductID), nil, adminToken)
	if w.Code != http.StatusOK || unchangedParent["stock"].(float64) != 5 {
		t.Fatalf("variant stock adjustment cannot change parent stock: %d %s", w.Code, w.Body.String())
	}

	checkout := func(items []any) (*httptest.ResponseRecorder, map[string]any) {
		t.Helper()
		return api.doReq(http.MethodPost, "/transactions", map[string]any{"items": items, "method": "QRIS"}, adminToken)
	}
	w, _ = checkout([]any{map[string]any{"productId": productID, "qty": 1}})
	if w.Code != http.StatusCreated {
		t.Fatalf("first sale above threshold: %d %s", w.Code, w.Body.String())
	}
	w, _ = checkout([]any{map[string]any{"productId": productID, "qty": 1}})
	if w.Code != http.StatusCreated {
		t.Fatalf("sale reaching threshold: %d %s", w.Code, w.Body.String())
	}
	w, notifications := api.doReq(http.MethodGet, "/notifications?category=stok", nil, adminToken)
	items := notifications["items"].([]any)
	if w.Code != http.StatusOK || len(items) != 1 || items[0].(map[string]any)["reference_id"] != fmt.Sprint(productID) {
		t.Fatalf("threshold creates one product alert: %d %s", w.Code, w.Body.String())
	}
	notificationID := uint(items[0].(map[string]any)["id"].(float64))
	w, _ = api.doReq(http.MethodPatch, fmt.Sprintf("/notifications/%d/read", notificationID), nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("mark stock alert read: %d %s", w.Code, w.Body.String())
	}
	w, thirdSale := checkout([]any{map[string]any{"productId": productID, "qty": 1}})
	if w.Code != http.StatusCreated {
		t.Fatalf("second sale below threshold: %d %s", w.Code, w.Body.String())
	}
	thirdTrxID := uint(thirdSale["id"].(float64))
	w, _ = api.doReq(http.MethodPost, "/transactions/"+fmt.Sprint(thirdTrxID)+"/refund", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "reason": "Batal satu barang",
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("refund low stock product: %d %s", w.Code, w.Body.String())
	}

	w, _ = checkout([]any{map[string]any{"productId": zeroThresholdID, "qty": 1}})
	if w.Code != http.StatusCreated {
		t.Fatalf("checkout disabled threshold product: %d %s", w.Code, w.Body.String())
	}
	w, _ = checkout([]any{map[string]any{"productId": variantProductID, "variantId": variantID, "qty": 1}})
	if w.Code != http.StatusCreated {
		t.Fatalf("checkout low stock variant: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, "/stock/adjustments", map[string]any{
		"productId": variantProductID, "variantId": variantID, "direction": "minus", "qty": 1, "reason": "Cacat",
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("adjust low stock variant: %d %s", w.Code, w.Body.String())
	}
	w, variantMovements := api.doReq(http.MethodGet, fmt.Sprintf("/movements?productId=%d", variantProductID), nil, adminToken)
	firstVariantMovement := variantMovements["items"].([]any)[0].(map[string]any)
	if w.Code != http.StatusOK || firstVariantMovement["type"] != "adjust" || uint(firstVariantMovement["variant_id"].(float64)) != variantID {
		t.Fatalf("variant adjustment movement keeps variant identity: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, "/stock/adjustments", map[string]any{
		"productId": adjustmentID, "direction": "minus", "qty": 2, "reason": "Barang rusak",
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("adjust stock to threshold: %d %s", w.Code, w.Body.String())
	}

	w, notifications = api.doReq(http.MethodGet, "/notifications?category=stok", nil, adminToken)
	items = notifications["items"].([]any)
	if w.Code != http.StatusOK || len(items) != 3 {
		t.Fatalf("one alert for each low stock entity, disabled threshold excluded: %d %s", w.Code, w.Body.String())
	}
	alertsByReference := map[string]map[string]any{}
	for _, raw := range items {
		item := raw.(map[string]any)
		alertsByReference[item["reference_type"].(string)+":"+item["reference_id"].(string)] = item
	}
	variantAlert := alertsByReference["variant:"+fmt.Sprint(variantID)]
	if variantAlert == nil || !strings.Contains(variantAlert["message"].(string), "Sepatu") || !strings.Contains(variantAlert["message"].(string), "Hitam 42") {
		t.Fatalf("variant alert names product and variant: %v", variantAlert)
	}
	if alertsByReference["product:"+fmt.Sprint(zeroThresholdID)] != nil {
		t.Fatalf("zero threshold remains disabled despite global suggestion: %v", alertsByReference)
	}
	w, dashboard := api.doReq(http.MethodGet, "/dashboard", nil, adminToken)
	today := dashboard["today"].(map[string]any)
	if w.Code != http.StatusOK || today["low_stock"].(float64) != 3 {
		t.Fatalf("dashboard counts only enabled product/variant thresholds: %d %s", w.Code, w.Body.String())
	}
}
