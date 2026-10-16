package router

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/rayhanvito/nojpos/backend/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestOfflineCheckoutIdempotencyAndStoreScope(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Checkout Offline")
	otherAdminToken := api.registerAdmin("Toko Checkout Offline Lain")
	productID := createInventoryProduct(t, api, adminToken, "Kopi Offline", "OFFLINE-KOPI", 1000, 5, 0)
	key := "550e8400-e29b-41d4-a716-446655440000"
	body := map[string]any{
		"items":      []any{map[string]any{"productId": productID, "qty": 2}},
		"method":     "QRIS",
		"clientTxId": key,
	}

	w, first := api.doReq(http.MethodPost, "/transactions", body, adminToken)
	if w.Code != http.StatusCreated || first["status"] != "completed" {
		t.Fatalf("first offline checkout: %d %s", w.Code, w.Body.String())
	}
	firstID := uint(first["id"].(float64))
	assertOfflineFinalStock(t, first, productID, nil, 3)

	w, replay := api.doReq(http.MethodPost, "/transactions", body, adminToken)
	if w.Code != http.StatusOK || uint(replay["id"].(float64)) != firstID {
		t.Fatalf("replayed checkout must return the first transaction: %d %s", w.Code, w.Body.String())
	}
	assertOfflineFinalStock(t, replay, productID, nil, 3)
	w, product := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d", productID), nil, adminToken)
	if w.Code != http.StatusOK || product["stock"].(float64) != 3 {
		t.Fatalf("replay must decrement stock once: %d %s", w.Code, w.Body.String())
	}
	w, list := api.doReq(http.MethodGet, "/transactions", nil, adminToken)
	if w.Code != http.StatusOK || len(list["items"].([]any)) != 1 {
		t.Fatalf("same-store replay must create one transaction: %d %s", w.Code, w.Body.String())
	}
	w, detail := api.doReq(http.MethodGet, fmt.Sprintf("/transactions/%d", firstID), nil, adminToken)
	if w.Code != http.StatusOK || uint(detail["id"].(float64)) != firstID {
		t.Fatalf("checkout detail: %d %s", w.Code, w.Body.String())
	}

	otherProductID := createInventoryProduct(t, api, otherAdminToken, "Kopi Offline", "OFFLINE-KOPI", 1000, 5, 0)
	otherBody := map[string]any{
		"items":      []any{map[string]any{"productId": otherProductID, "qty": 1}},
		"method":     "QRIS",
		"clientTxId": key,
	}
	w, other := api.doReq(http.MethodPost, "/transactions", otherBody, otherAdminToken)
	if w.Code != http.StatusCreated || uint(other["id"].(float64)) == firstID {
		t.Fatalf("same key in another store must create a distinct transaction: %d %s", w.Code, w.Body.String())
	}
	assertOfflineFinalStock(t, other, otherProductID, nil, 4)
}

func TestOfflineCheckoutShortageAndNotification(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Stok Offline")
	productID := createInventoryProduct(t, api, adminToken, "Gula Offline", "OFFLINE-GULA", 1000, 1, 0)
	productBody := map[string]any{
		"items":      []any{map[string]any{"productId": productID, "qty": 3}},
		"method":     "QRIS",
		"clientTxId": "550e8400-e29b-41d4-a716-446655440001",
	}
	w, productShortage := api.doReq(http.MethodPost, "/transactions", productBody, adminToken)
	if w.Code != http.StatusCreated || productShortage["status"] != "completed_offline_shortage" {
		t.Fatalf("product shortage checkout: %d %s", w.Code, w.Body.String())
	}
	productTrxID := uint(productShortage["id"].(float64))
	if productShortage["total"].(float64) != 3000 || productShortage["items"].([]any)[0].(map[string]any)["qty"].(float64) != 3 {
		t.Fatalf("shortage must preserve the requested quantity and checkout amount: %v", productShortage)
	}
	assertOfflineFinalStock(t, productShortage, productID, nil, 0)
	w, product := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d", productID), nil, adminToken)
	if w.Code != http.StatusOK || product["stock"].(float64) != 0 {
		t.Fatalf("shortage must not make product stock negative: %d %s", w.Code, w.Body.String())
	}

	variantProductID := createInventoryProduct(t, api, adminToken, "Teh Offline", "OFFLINE-TEH", 1000, 0, 0)
	w, variant := api.doReq(http.MethodPost, fmt.Sprintf("/products/%d/variants", variantProductID), map[string]any{
		"name": "Kotak", "sku": "OFFLINE-TEH-KOTAK", "buyPrice": 300, "sellPrice": 1000, "stock": 2,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create offline variant: %d %s", w.Code, w.Body.String())
	}
	variantID := uint(variant["id"].(float64))
	variantBody := map[string]any{
		"items":      []any{map[string]any{"productId": variantProductID, "variantId": variantID, "qty": 4}},
		"method":     "QRIS",
		"clientTxId": "550e8400-e29b-41d4-a716-446655440002",
	}
	w, variantShortage := api.doReq(http.MethodPost, "/transactions", variantBody, adminToken)
	if w.Code != http.StatusCreated || variantShortage["status"] != "completed_offline_shortage" {
		t.Fatalf("variant shortage checkout: %d %s", w.Code, w.Body.String())
	}
	assertOfflineFinalStock(t, variantShortage, variantProductID, &variantID, 0)
	w, variants := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d/variants", variantProductID), nil, adminToken)
	if w.Code != http.StatusOK || variants["items"].([]any)[0].(map[string]any)["stock"].(float64) != 0 {
		t.Fatalf("shortage must not make variant stock negative: %d %s", w.Code, w.Body.String())
	}

	w, replay := api.doReq(http.MethodPost, "/transactions", productBody, adminToken)
	if w.Code != http.StatusOK || uint(replay["id"].(float64)) != productTrxID {
		t.Fatalf("shortage replay: %d %s", w.Code, w.Body.String())
	}
	assertOfflineFinalStock(t, replay, productID, nil, 0)
	w, notifications := api.doReq(http.MethodGet, "/notifications?category=transaksi", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("list shortage notifications: %d %s", w.Code, w.Body.String())
	}
	items := notifications["items"].([]any)
	if len(items) != 2 {
		t.Fatalf("each shortage checkout should notify once, got %d notifications: %v", len(items), items)
	}
	for _, raw := range items {
		notification := raw.(map[string]any)
		if notification["type"] != "transaction_offline_shortage" || notification["title"] != "Transaksi offline kekurangan stok" {
			t.Fatalf("shortage notification content: %v", notification)
		}
		if notification["reference_id"] != fmt.Sprint(productTrxID) && notification["reference_id"] != fmt.Sprint(variantShortage["id"]) {
			t.Fatalf("shortage notification must reference a transaction: %v", notification)
		}
		if notification["message"] == "" {
			t.Fatalf("shortage notification must describe the stock shortage: %v", notification)
		}
	}
}

func TestOfflineCheckoutWithoutClientTxIdKeepsOnlineBehavior(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Checkout Online")
	productID := createInventoryProduct(t, api, adminToken, "Garam Online", "ONLINE-GARAM", 1000, 4, 0)
	body := map[string]any{
		"items":  []any{map[string]any{"productId": productID, "qty": 1}},
		"method": "QRIS",
	}
	firstW, first := api.doReq(http.MethodPost, "/transactions", body, adminToken)
	if firstW.Code != http.StatusCreated {
		t.Fatalf("online checkout: %d %s", firstW.Code, firstW.Body.String())
	}
	body["clientTxId"] = ""
	secondW, second := api.doReq(http.MethodPost, "/transactions", body, adminToken)
	if secondW.Code != http.StatusCreated || uint(second["id"].(float64)) == uint(first["id"].(float64)) {
		t.Fatalf("empty clientTxId must not deduplicate online checkouts: %d %s", secondW.Code, secondW.Body.String())
	}
	w, list := api.doReq(http.MethodGet, "/transactions", nil, adminToken)
	if w.Code != http.StatusOK || len(list["items"].([]any)) != 2 {
		t.Fatalf("two online checkouts must create two transactions: %d %s", w.Code, w.Body.String())
	}
	w, product := api.doReq(http.MethodGet, fmt.Sprintf("/products/%d", productID), nil, adminToken)
	if w.Code != http.StatusOK || product["stock"].(float64) != 2 {
		t.Fatalf("two online checkouts must decrement stock twice: %d %s", w.Code, w.Body.String())
	}
	productID = createInventoryProduct(t, api, adminToken, "Lada Online", "ONLINE-LADA", 1000, 1, 0)
	w, rejected := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 2}}, "method": "QRIS",
	}, adminToken)
	if w.Code != http.StatusConflict || rejected["error"] == nil {
		t.Fatalf("online checkout with insufficient stock must remain rejected: %d %s", w.Code, w.Body.String())
	}
	w, product = api.doReq(http.MethodGet, fmt.Sprintf("/products/%d", productID), nil, adminToken)
	if w.Code != http.StatusOK || product["stock"].(float64) != 1 {
		t.Fatalf("rejected online checkout must not change stock: %d %s", w.Code, w.Body.String())
	}
}

func TestOfflineCheckoutClientTxIndexScopesKeysAndIgnoresEmptyIDs(t *testing.T) {
	newPromoTestAPI(t)
	db, err := gorm.Open(sqlite.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open migrated SQLite database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get migrated SQLite connection: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	if !db.Migrator().HasIndex(&model.Trx{}, "uq_transactions_store_client_tx_id") {
		t.Fatal("AutoMigrate must create the per-store client transaction index")
	}
	create := func(storeID uint, clientTxID string) error {
		return db.Create(&model.Trx{
			StoreID: storeID, ClientTxID: clientTxID, CashierName: "Uji Indeks", Items: model.TrxItemList{}, Method: "QRIS",
		}).Error
	}
	if err := create(1, "same-key"); err != nil {
		t.Fatalf("create first keyed transaction: %v", err)
	}
	if err := create(1, "same-key"); err == nil {
		t.Fatal("same non-empty clientTxId must be unique within a store")
	}
	if err := create(2, "same-key"); err != nil {
		t.Fatalf("same clientTxId must be allowed in another store: %v", err)
	}
	if err := create(1, ""); err != nil {
		t.Fatalf("create first empty clientTxId: %v", err)
	}
	if err := create(1, ""); err != nil {
		t.Fatalf("empty clientTxId must not be constrained: %v", err)
	}
}

func assertOfflineFinalStock(t *testing.T, transaction map[string]any, productID uint, variantID *uint, stock int) {
	t.Helper()
	rows, ok := transaction["final_stocks"].([]any)
	if !ok || len(rows) != 1 {
		t.Fatalf("expected one final stock entry, got %v", transaction["final_stocks"])
	}
	row := rows[0].(map[string]any)
	if uint(row["product_id"].(float64)) != productID || row["stock"].(float64) != float64(stock) {
		t.Fatalf("unexpected final stock entry: %v", row)
	}
	if variantID == nil {
		if row["variant_id"] != nil {
			t.Fatalf("product stock must have null variant_id: %v", row)
		}
	} else if uint(row["variant_id"].(float64)) != *variantID {
		t.Fatalf("unexpected variant_id in final stock entry: %v", row)
	}
}
