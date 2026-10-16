package router

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func stockOpnameItem(t *testing.T, items any, productID, variantID uint) map[string]any {
	t.Helper()
	for _, raw := range items.([]any) {
		item := raw.(map[string]any)
		itemProductID := uint(item["product_id"].(float64))
		var itemVariantID uint
		if value, ok := item["variant_id"].(float64); ok {
			itemVariantID = uint(value)
		}
		if itemProductID == productID && itemVariantID == variantID {
			return item
		}
	}
	t.Fatalf("stock opname item product=%d variant=%d not found: %v", productID, variantID, items)
	return nil
}

func createStockOpname(t *testing.T, api *promoTestAPI, token, note string) (uint, map[string]any) {
	t.Helper()
	w, session := api.doReq(http.MethodPost, "/stock-opnames", map[string]any{"note": note}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("create stock opname: %d %s", w.Code, w.Body.String())
	}
	return uint(session["id"].(float64)), session
}

func assertStockOpnameMovement(t *testing.T, item map[string]any, productID, variantID uint, qty int, sessionID uint) {
	t.Helper()
	if uint(item["product_id"].(float64)) != productID || int(item["qty"].(float64)) != qty {
		t.Fatalf("stock opname movement target/qty: %v", item)
	}
	var gotVariantID uint
	if value, ok := item["variant_id"].(float64); ok {
		gotVariantID = uint(value)
	}
	if gotVariantID != variantID || !strings.Contains(item["reason"].(string), fmt.Sprintf("%d", sessionID)) {
		t.Fatalf("stock opname movement variant/reason: %v", item)
	}
}

func TestStockOpnameSnapshotAndPostingUseSystemSnapshot(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Opname Snapshot")
	cashierToken := api.createCashier(adminToken)

	productID := createPurchaseTestProduct(t, api, adminToken, "Produk Stok", "OP-SNAPSHOT", 10)
	variantProductID := createPurchaseTestProduct(t, api, adminToken, "Produk Varian", "OP-VARIANTS", 99)
	redID := createPurchaseTestVariant(t, api, adminToken, variantProductID, "Merah", "OP-VARIANTS-RED", 8)
	blueID := createPurchaseTestVariant(t, api, adminToken, variantProductID, "Biru", "OP-VARIANTS-BLUE", 5)
	inactiveID := createPurchaseTestProduct(t, api, adminToken, "Produk Nonaktif", "OP-INACTIVE", 11)
	w, _ := api.doReq(http.MethodPatch, fmt.Sprintf("/products/%d/active", inactiveID), map[string]any{"active": false}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("deactivate product before snapshot: %d %s", w.Code, w.Body.String())
	}

	sessionID, created := createStockOpname(t, api, cashierToken, "Hitung rak utama")
	if created["status"] != "counting" || created["created_by"] == "" {
		t.Fatalf("created session status/actor: %v", created)
	}
	items, ok := created["items"].([]any)
	if !ok || len(items) != 3 {
		t.Fatalf("snapshot should contain the active base product and both variants, got: %v", created["items"])
	}
	baseItem := stockOpnameItem(t, items, productID, 0)
	redItem := stockOpnameItem(t, items, variantProductID, redID)
	blueItem := stockOpnameItem(t, items, variantProductID, blueID)
	if baseItem["system_stock"] != float64(10) || redItem["system_stock"] != float64(8) || blueItem["system_stock"] != float64(5) {
		t.Fatalf("system stock snapshots: %v", items)
	}
	if baseItem["counted_stock"] != nil || baseItem["difference"] != nil {
		t.Fatalf("uncounted item should have null count and difference: %v", baseItem)
	}

	for _, adjustment := range []map[string]any{
		{"productId": productID, "direction": "minus", "qty": 2, "reason": "Penjualan sesudah snapshot"},
		{"productId": variantProductID, "variantId": redID, "direction": "plus", "qty": 5, "reason": "Stok masuk sesudah snapshot"},
		{"productId": variantProductID, "variantId": blueID, "direction": "minus", "qty": 2, "reason": "Stok rusak sesudah snapshot"},
	} {
		w, _ = api.doReq(http.MethodPost, "/stock/adjustments", adjustment, adminToken)
		if w.Code != http.StatusOK {
			t.Fatalf("adjust stock after snapshot: %d %s", w.Code, w.Body.String())
		}
	}

	baseItemID := uint(baseItem["id"].(float64))
	redItemID := uint(redItem["id"].(float64))
	blueItemID := uint(blueItem["id"].(float64))
	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/stock-opnames/%d/items", sessionID), map[string]any{
		"items": []any{
			map[string]any{"itemId": baseItemID, "countedStock": 7},
			map[string]any{"itemId": redItemID, "countedStock": 12},
		},
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("first partial count: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/stock-opnames/%d/items", sessionID), map[string]any{
		"items": []any{map[string]any{"itemId": blueItemID, "countedStock": 0}},
	}, cashierToken)
	if w.Code != http.StatusOK {
		t.Fatalf("second partial count: %d %s", w.Code, w.Body.String())
	}

	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/stock-opnames/%d", sessionID), nil, cashierToken)
	if w.Code != http.StatusOK {
		t.Fatalf("get stock opname detail: %d %s", w.Code, w.Body.String())
	}
	items = detail["items"].([]any)
	baseItem = stockOpnameItem(t, items, productID, 0)
	redItem = stockOpnameItem(t, items, variantProductID, redID)
	blueItem = stockOpnameItem(t, items, variantProductID, blueID)
	if baseItem["difference"] != float64(-3) || redItem["difference"] != float64(4) || blueItem["difference"] != float64(-5) {
		t.Fatalf("differences must use session snapshots: %v", items)
	}
	if baseItem["product_name"] != "Produk Stok" || redItem["product_name"] != "Produk Varian" || redItem["variant_name"] != "Merah" {
		t.Fatalf("detail item relation names: %v", items)
	}

	w, posted := api.doReq(http.MethodPost, fmt.Sprintf("/stock-opnames/%d/post", sessionID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK || posted["status"] != "posted" {
		t.Fatalf("post stock opname: %d %s", w.Code, w.Body.String())
	}
	if stock, _ := purchaseTestStock(t, api, adminToken, productID, 0); stock != 7 {
		t.Fatalf("base product stock after post: got %d want 7", stock)
	}
	if parentStock, redStock := purchaseTestStock(t, api, adminToken, variantProductID, redID); parentStock != 99 || redStock != 12 {
		t.Fatalf("variant stock after post: parent=%d red=%d", parentStock, redStock)
	}
	if _, blueStock := purchaseTestStock(t, api, adminToken, variantProductID, blueID); blueStock != 0 {
		t.Fatalf("blue variant stock after post: got %d want 0", blueStock)
	}

	w, movements := api.doReq(http.MethodGet, "/movements?type=opname", nil, adminToken)
	if w.Code != http.StatusOK || len(movements["items"].([]any)) != 3 {
		t.Fatalf("opname movements: %d %s", w.Code, w.Body.String())
	}
	movementByTarget := map[string]map[string]any{}
	for _, raw := range movements["items"].([]any) {
		movement := raw.(map[string]any)
		variantID := uint(0)
		if value, ok := movement["variant_id"].(float64); ok {
			variantID = uint(value)
		}
		movementByTarget[fmt.Sprintf("%d:%d", uint(movement["product_id"].(float64)), variantID)] = movement
	}
	assertStockOpnameMovement(t, movementByTarget[fmt.Sprintf("%d:0", productID)], productID, 0, -3, sessionID)
	assertStockOpnameMovement(t, movementByTarget[fmt.Sprintf("%d:%d", variantProductID, redID)], variantProductID, redID, 4, sessionID)
	assertStockOpnameMovement(t, movementByTarget[fmt.Sprintf("%d:%d", variantProductID, blueID)], variantProductID, blueID, -5, sessionID)
}

func TestStockOpnameOneCountingSessionPerStoreAndStoreScope(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Opname Satu")
	cashierToken := api.createCashier(adminToken)
	otherAdminToken := api.registerAdmin("Toko Opname Dua")
	createPurchaseTestProduct(t, api, adminToken, "Produk Satu", "OP-STORE-ONE", 4)
	createPurchaseTestProduct(t, api, otherAdminToken, "Produk Dua", "OP-STORE-TWO", 8)

	firstID, _ := createStockOpname(t, api, adminToken, "Sesi toko satu")
	w, conflict := api.doReq(http.MethodPost, "/stock-opnames", map[string]any{"note": "Sesi duplikat"}, cashierToken)
	if w.Code != http.StatusConflict || !strings.Contains(strings.ToLower(conflict["error"].(string)), "opname") {
		t.Fatalf("second counting session should conflict with a clear message: %d %s", w.Code, w.Body.String())
	}
	otherID, _ := createStockOpname(t, api, otherAdminToken, "Sesi toko dua")

	for token, wantID := range map[string]uint{adminToken: firstID, otherAdminToken: otherID} {
		w, list := api.doReq(http.MethodGet, "/stock-opnames", nil, token)
		rows, ok := list["items"].([]any)
		if w.Code != http.StatusOK || !ok || len(rows) != 1 || uint(rows[0].(map[string]any)["id"].(float64)) != wantID {
			t.Fatalf("store-scoped stock opname list: %d %s", w.Code, w.Body.String())
		}
	}
	w, _ = api.doReq(http.MethodGet, fmt.Sprintf("/stock-opnames/%d", firstID), nil, otherAdminToken)
	if w.Code != http.StatusNotFound {
		t.Fatalf("foreign store session detail should be hidden: %d %s", w.Code, w.Body.String())
	}
}

func TestStockOpnameSkipsUncountedItemsAndLocksPostedSession(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Opname Terkunci")
	cashierToken := api.createCashier(adminToken)
	countedProductID := createPurchaseTestProduct(t, api, adminToken, "Dihitung", "OP-COUNTED", 5)
	uncountedProductID := createPurchaseTestProduct(t, api, adminToken, "Belum Dihitung", "OP-UNCOUNTED", 9)
	sessionID, created := createStockOpname(t, api, cashierToken, "Sesi sebagian")
	countedItem := stockOpnameItem(t, created["items"], countedProductID, 0)
	uncountedItem := stockOpnameItem(t, created["items"], uncountedProductID, 0)
	w, _ := api.doReq(http.MethodPut, fmt.Sprintf("/stock-opnames/%d/items", sessionID), map[string]any{
		"items": []any{map[string]any{"itemId": uint(countedItem["id"].(float64)), "countedStock": 4}},
	}, cashierToken)
	if w.Code != http.StatusOK {
		t.Fatalf("count one of two items: %d %s", w.Code, w.Body.String())
	}
	w, posted := api.doReq(http.MethodPost, fmt.Sprintf("/stock-opnames/%d/post", sessionID), map[string]any{}, cashierToken)
	if w.Code != http.StatusOK || posted["status"] != "posted" {
		t.Fatalf("post partially counted session: %d %s", w.Code, w.Body.String())
	}
	if stock, _ := purchaseTestStock(t, api, adminToken, countedProductID, 0); stock != 4 {
		t.Fatalf("counted stock: got %d want 4", stock)
	}
	if stock, _ := purchaseTestStock(t, api, adminToken, uncountedProductID, 0); stock != 9 {
		t.Fatalf("uncounted stock changed: got %d want 9", stock)
	}
	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/stock-opnames/%d", sessionID), nil, adminToken)
	if w.Code != http.StatusOK || stockOpnameItem(t, detail["items"], uncountedProductID, 0)["counted_stock"] != nil {
		t.Fatalf("uncounted item detail after post: %d %s", w.Code, w.Body.String())
	}
	for _, action := range []string{"items", "post", "cancel"} {
		method, path, body := http.MethodPost, fmt.Sprintf("/stock-opnames/%d/%s", sessionID, action), any(map[string]any{})
		if action == "items" {
			method = http.MethodPut
			body = map[string]any{"items": []any{map[string]any{"itemId": uint(uncountedItem["id"].(float64)), "countedStock": 8}}}
		}
		w, _ = api.doReq(method, path, body, adminToken)
		if w.Code != http.StatusConflict {
			t.Fatalf("posted session action %s should be locked: %d %s", action, w.Code, w.Body.String())
		}
	}
	w, movements := api.doReq(http.MethodGet, "/movements?type=opname", nil, adminToken)
	if w.Code != http.StatusOK || len(movements["items"].([]any)) != 1 || uint(movements["items"].([]any)[0].(map[string]any)["product_id"].(float64)) != countedProductID {
		t.Fatalf("only counted item should create movement: %d %s", w.Code, w.Body.String())
	}
}

func TestStockOpnameCancelDoesNotChangeStock(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Opname Batal")
	productID := createPurchaseTestProduct(t, api, adminToken, "Produk Batal", "OP-CANCEL", 7)
	sessionID, created := createStockOpname(t, api, adminToken, "Dibatalkan")
	item := stockOpnameItem(t, created["items"], productID, 0)
	w, _ := api.doReq(http.MethodPut, fmt.Sprintf("/stock-opnames/%d/items", sessionID), map[string]any{
		"items": []any{map[string]any{"itemId": uint(item["id"].(float64)), "countedStock": 3}},
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("count before cancel: %d %s", w.Code, w.Body.String())
	}
	w, cancelled := api.doReq(http.MethodPost, fmt.Sprintf("/stock-opnames/%d/cancel", sessionID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK || cancelled["status"] != "cancelled" {
		t.Fatalf("cancel stock opname: %d %s", w.Code, w.Body.String())
	}
	if stock, _ := purchaseTestStock(t, api, adminToken, productID, 0); stock != 7 {
		t.Fatalf("cancel changed stock: got %d want 7", stock)
	}
	w, movements := api.doReq(http.MethodGet, "/movements?type=opname", nil, adminToken)
	if w.Code != http.StatusOK || len(movements["items"].([]any)) != 0 {
		t.Fatalf("cancel should not create movements: %d %s", w.Code, w.Body.String())
	}
}

func TestStockOpnameRejectsInvalidAndForeignItems(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Opname Validasi")
	productID := createPurchaseTestProduct(t, api, adminToken, "Produk Validasi", "OP-VALIDATION", 6)
	firstID, first := createStockOpname(t, api, adminToken, "Sesi pertama")
	firstItemID := uint(stockOpnameItem(t, first["items"], productID, 0)["id"].(float64))
	w, _ := api.doReq(http.MethodPut, fmt.Sprintf("/stock-opnames/%d/items", firstID), map[string]any{
		"items": []any{map[string]any{"itemId": firstItemID, "countedStock": -1}},
	}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("negative counted stock should be rejected: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/stock-opnames/%d/items", firstID), map[string]any{
		"items": []any{map[string]any{"itemId": firstItemID}},
	}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("missing countedStock should be rejected: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/stock-opnames/%d/cancel", firstID), map[string]any{}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("cancel first session before creating another: %d %s", w.Code, w.Body.String())
	}
	secondID, second := createStockOpname(t, api, adminToken, "Sesi kedua")
	secondItemID := uint(stockOpnameItem(t, second["items"], productID, 0)["id"].(float64))
	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/stock-opnames/%d/items", secondID), map[string]any{
		"items": []any{
			map[string]any{"itemId": secondItemID, "countedStock": 5},
			map[string]any{"itemId": firstItemID, "countedStock": 2},
		},
	}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("item from another session should be rejected: %d %s", w.Code, w.Body.String())
	}
	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/stock-opnames/%d", secondID), nil, adminToken)
	if w.Code != http.StatusOK || stockOpnameItem(t, detail["items"], productID, 0)["counted_stock"] != nil {
		t.Fatalf("invalid batch must leave session item unchanged: %d %s", w.Code, w.Body.String())
	}
}

func TestStockOpnamePostRollsBackIfMovementInsertFails(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Opname Rollback")
	firstProductID := createPurchaseTestProduct(t, api, adminToken, "Produk Pertama", "OP-ROLLBACK-ONE", 5)
	secondProductID := createPurchaseTestProduct(t, api, adminToken, "Produk Kedua", "OP-ROLLBACK-TWO", 8)
	sessionID, created := createStockOpname(t, api, adminToken, "Rollback movement")
	firstItemID := uint(stockOpnameItem(t, created["items"], firstProductID, 0)["id"].(float64))
	secondItemID := uint(stockOpnameItem(t, created["items"], secondProductID, 0)["id"].(float64))
	w, _ := api.doReq(http.MethodPut, fmt.Sprintf("/stock-opnames/%d/items", sessionID), map[string]any{
		"items": []any{
			map[string]any{"itemId": firstItemID, "countedStock": 2},
			map[string]any{"itemId": secondItemID, "countedStock": 6},
		},
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("count rollback items: %d %s", w.Code, w.Body.String())
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
	trigger := fmt.Sprintf(`CREATE TRIGGER fail_second_opname_movement BEFORE INSERT ON stock_movements
		WHEN NEW.type = 'opname' AND NEW.product_id = %d
		BEGIN SELECT RAISE(ABORT, 'forced opname movement failure'); END`, secondProductID)
	if err := database.Exec(trigger).Error; err != nil {
		t.Fatalf("create movement failure trigger: %v", err)
	}

	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/stock-opnames/%d/post", sessionID), map[string]any{}, adminToken)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("movement failure should fail posting: %d %s", w.Code, w.Body.String())
	}
	if stock, _ := purchaseTestStock(t, api, adminToken, firstProductID, 0); stock != 5 {
		t.Fatalf("first stock update was not rolled back: %d", stock)
	}
	if stock, _ := purchaseTestStock(t, api, adminToken, secondProductID, 0); stock != 8 {
		t.Fatalf("second stock update changed after rollback: %d", stock)
	}
	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/stock-opnames/%d", sessionID), nil, adminToken)
	if w.Code != http.StatusOK || detail["status"] != "counting" {
		t.Fatalf("failed post must leave session counting: %d %s", w.Code, w.Body.String())
	}
	w, movements := api.doReq(http.MethodGet, "/movements?type=opname", nil, adminToken)
	if w.Code != http.StatusOK || len(movements["items"].([]any)) != 0 {
		t.Fatalf("failed post left movement rows: %d %s", w.Code, w.Body.String())
	}
}

func TestStockOpnamePostRejectsBaseProductThatGainedVariant(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Opname Target Berubah")
	firstProductID := createPurchaseTestProduct(t, api, adminToken, "Produk Tetap", "OP-TARGET-ONE", 5)
	changedProductID := createPurchaseTestProduct(t, api, adminToken, "Produk Berubah", "OP-TARGET-TWO", 8)
	sessionID, created := createStockOpname(t, api, adminToken, "Target berubah")
	firstItemID := uint(stockOpnameItem(t, created["items"], firstProductID, 0)["id"].(float64))
	changedItemID := uint(stockOpnameItem(t, created["items"], changedProductID, 0)["id"].(float64))
	w, _ := api.doReq(http.MethodPut, fmt.Sprintf("/stock-opnames/%d/items", sessionID), map[string]any{
		"items": []any{
			map[string]any{"itemId": firstItemID, "countedStock": 2},
			map[string]any{"itemId": changedItemID, "countedStock": 6},
		},
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("count items before target changes: %d %s", w.Code, w.Body.String())
	}
	variantID := createPurchaseTestVariant(t, api, adminToken, changedProductID, "Varian Baru", "OP-TARGET-TWO-NEW", 3)

	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/stock-opnames/%d/post", sessionID), map[string]any{}, adminToken)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("posting a base-product snapshot after variants were added should fail: %d %s", w.Code, w.Body.String())
	}
	if stock, _ := purchaseTestStock(t, api, adminToken, firstProductID, 0); stock != 5 {
		t.Fatalf("earlier item stock update was not rolled back: got %d want 5", stock)
	}
	if stock, _ := purchaseTestStock(t, api, adminToken, changedProductID, 0); stock != 8 {
		t.Fatalf("changed base-product stock: got %d want 8", stock)
	}
	if parentStock, variantStock := purchaseTestStock(t, api, adminToken, changedProductID, variantID); parentStock != 8 || variantStock != 3 {
		t.Fatalf("new variant stock changed: parent=%d variant=%d", parentStock, variantStock)
	}
	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/stock-opnames/%d", sessionID), nil, adminToken)
	if w.Code != http.StatusOK || detail["status"] != "counting" {
		t.Fatalf("rejected post must leave session counting: %d %s", w.Code, w.Body.String())
	}
	w, movements := api.doReq(http.MethodGet, "/movements?type=opname", nil, adminToken)
	if w.Code != http.StatusOK || len(movements["items"].([]any)) != 0 {
		t.Fatalf("rejected post left movement rows: %d %s", w.Code, w.Body.String())
	}
}
