package router

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/rayhanvito/nojpos/backend/model"
	"github.com/rayhanvito/nojpos/backend/service"
)

func TestOutletAPIAndSessionSwitchRespectOutletBindings(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Outlet")
	cashierToken := api.createCashier(adminToken)
	otherAdminToken := api.registerAdmin("Toko Outlet Lain")

	w, outlets := api.doReq(http.MethodGet, "/outlets", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("list default outlet: %d %s", w.Code, w.Body.String())
	}
	items, ok := outlets["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("default outlet list must use items: %v", outlets)
	}
	defaultOutlet := items[0].(map[string]any)
	if defaultOutlet["name"] != "Toko Outlet" || defaultOutlet["is_default"] != true {
		t.Fatalf("default outlet: %v", defaultOutlet)
	}
	defaultOutletID := uint(defaultOutlet["id"].(float64))
	w, partialUpdate := api.doReq(http.MethodPut, fmt.Sprintf("/outlets/%d", defaultOutletID), map[string]any{
		"name": "Toko Outlet Diedit", "address": "Alamat Baru",
	}, adminToken)
	if w.Code != http.StatusOK || partialUpdate["name"] != "Toko Outlet Diedit" || partialUpdate["is_default"] != true {
		t.Fatalf("partial update must preserve default status: %d %s", w.Code, w.Body.String())
	}
	w, rejectedDefault := api.doReq(http.MethodPut, fmt.Sprintf("/outlets/%d", defaultOutletID), map[string]any{
		"name": "Toko Outlet Diedit", "address": "Alamat Baru", "isDefault": false,
	}, adminToken)
	if w.Code < http.StatusBadRequest || w.Code >= http.StatusInternalServerError || rejectedDefault["error"] == nil {
		t.Fatalf("explicitly disabling the default outlet must be rejected: %d %s", w.Code, w.Body.String())
	}

	w, outletB := api.doReq(http.MethodPost, "/outlets", map[string]any{
		"name": "Cabang Timur", "address": "Jl Timur",
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create outlet: %d %s", w.Code, w.Body.String())
	}
	outletBID := uint(outletB["id"].(float64))
	w, _ = api.doReq(http.MethodPost, "/outlets", map[string]any{"name": "Ditolak"}, cashierToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("cashier outlet create: want 403, got %d %s", w.Code, w.Body.String())
	}
	w, updatedOutlet := api.doReq(http.MethodPut, fmt.Sprintf("/outlets/%d", outletBID), map[string]any{
		"name": "Cabang Timur Baru", "address": "Jl Timur No. 2", "isDefault": true,
	}, adminToken)
	if w.Code != http.StatusOK || updatedOutlet["name"] != "Cabang Timur Baru" || updatedOutlet["is_default"] != true {
		t.Fatalf("update/default outlet: %d %s", w.Code, w.Body.String())
	}
	w, outlets = api.doReq(http.MethodGet, "/outlets", nil, adminToken)
	items = outlets["items"].([]any)
	defaults := 0
	for _, item := range items {
		if item.(map[string]any)["is_default"] == true {
			defaults++
		}
	}
	if defaults != 1 {
		t.Fatalf("store must retain exactly one default outlet: %v", items)
	}

	w, switched := api.doReq(http.MethodPost, "/outlets/switch", map[string]any{"outletId": outletBID}, adminToken)
	if w.Code != http.StatusOK || outletIDFromUser(t, switched) != outletBID {
		t.Fatalf("admin switch outlet: %d %s", w.Code, w.Body.String())
	}
	adminOutletToken := switched["access_token"].(string)
	w, me := api.doReq(http.MethodGet, "/auth/me", nil, adminOutletToken)
	if w.Code != http.StatusOK || outletIDFromUser(t, me) != outletBID {
		t.Fatalf("active outlet in /auth/me: %d %s", w.Code, w.Body.String())
	}
	w, refreshed := api.doReq(http.MethodPost, "/auth/refresh", map[string]any{"refresh_token": switched["refresh_token"]}, "")
	if w.Code != http.StatusOK || outletIDFromUser(t, refreshed) != outletBID {
		t.Fatalf("refresh preserves active outlet: %d %s", w.Code, w.Body.String())
	}

	w, denied := api.doReq(http.MethodPost, "/outlets/switch", map[string]any{"outletId": outletBID}, cashierToken)
	if w.Code != http.StatusForbidden || denied["error"] == nil {
		t.Fatalf("cashier switch to another outlet: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodGet, "/outlets", nil, cashierToken)
	if w.Code != http.StatusOK {
		t.Fatalf("cashier may list store outlets: %d %s", w.Code, w.Body.String())
	}

	w, cashierB := api.doReq(http.MethodPost, "/users", map[string]any{"name": "Kasir Cabang", "outletId": outletBID}, adminOutletToken)
	if w.Code != http.StatusCreated || outletIDFromUser(t, cashierB) != outletBID {
		t.Fatalf("create cashier at outlet B: %d %s", w.Code, w.Body.String())
	}
	cashierBID := uint(cashierB["user"].(map[string]any)["id"].(float64))
	w, updatedUser := api.doReq(http.MethodPatch, fmt.Sprintf("/users/%d", cashierBID), map[string]any{
		"name": "Kasir Cabang Revisi", "outletId": defaultOutletID,
	}, adminOutletToken)
	if w.Code != http.StatusOK {
		t.Fatalf("update cashier outlet: %d %s", w.Code, w.Body.String())
	}
	if updatedUser["message"] == nil {
		t.Fatalf("update cashier response: %v", updatedUser)
	}
	w, userList := api.doReq(http.MethodGet, "/users", nil, adminOutletToken)
	if w.Code != http.StatusOK {
		t.Fatalf("list users: %d %s", w.Code, w.Body.String())
	}
	foundOutlet := false
	for _, user := range userList["users"].([]any) {
		row := user.(map[string]any)
		if uint(row["id"].(float64)) == cashierBID {
			foundOutlet = uint(row["outlet_id"].(float64)) == defaultOutletID
		}
	}
	if !foundOutlet {
		t.Fatalf("updated cashier outlet missing from user response: %v", userList)
	}

	w, otherOutlets := api.doReq(http.MethodGet, "/outlets", nil, otherAdminToken)
	if w.Code != http.StatusOK || len(otherOutlets["items"].([]any)) != 1 {
		t.Fatalf("other store default outlet: %d %s", w.Code, w.Body.String())
	}
	foreignOutletID := uint(otherOutlets["items"].([]any)[0].(map[string]any)["id"].(float64))
	w, response := api.doReq(http.MethodPost, "/outlets/switch", map[string]any{"outletId": foreignOutletID}, adminOutletToken)
	if w.Code != http.StatusNotFound || response["error"] == nil {
		t.Fatalf("switch to foreign outlet must not disclose it: %d %s", w.Code, w.Body.String())
	}
}

func TestCashierCannotChangeOutletDuringOpenShift(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Shift Outlet")
	w, outlets := api.doReq(http.MethodGet, "/outlets", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("list outlets: %d %s", w.Code, w.Body.String())
	}
	outletAID := uint(outlets["items"].([]any)[0].(map[string]any)["id"].(float64))
	w, outletB := api.doReq(http.MethodPost, "/outlets", map[string]any{"name": "Cabang Shift"}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create outlet B: %d %s", w.Code, w.Body.String())
	}
	outletBID := uint(outletB["id"].(float64))
	w, createdCashier := api.doReq(http.MethodPost, "/users", map[string]any{
		"name": "Kasir Shift", "outletId": outletAID,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create cashier: %d %s", w.Code, w.Body.String())
	}
	cashierID := uint(createdCashier["user"].(map[string]any)["id"].(float64))
	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/users/%d/passcode", cashierID), map[string]any{"passcode": "12345"}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("set cashier passcode: %d %s", w.Code, w.Body.String())
	}
	w, cashierSession := api.doReq(http.MethodPost, "/auth/switch", map[string]any{
		"target_user_id": cashierID, "passcode": "12345",
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("switch to cashier: %d %s", w.Code, w.Body.String())
	}
	cashierToken := cashierSession["access_token"].(string)
	w, _ = api.doReq(http.MethodPost, "/cashier/shift/start", map[string]any{"opening_cash": 5000}, cashierToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("start cashier shift: %d %s", w.Code, w.Body.String())
	}
	w, rejected := api.doReq(http.MethodPatch, fmt.Sprintf("/users/%d", cashierID), map[string]any{"outletId": outletBID}, adminToken)
	wantError := "Kasir tidak dapat dipindahkan ke outlet lain selama shift masih terbuka. Tutup shift terlebih dahulu."
	if w.Code != http.StatusConflict || rejected["error"] != wantError {
		t.Fatalf("open shift must block outlet reassignment with a clear conflict: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, "/cashier/shift/close", map[string]any{"closingCash": 0}, cashierToken)
	if w.Code != http.StatusOK {
		t.Fatalf("close cashier shift: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPatch, fmt.Sprintf("/users/%d", cashierID), map[string]any{"outletId": outletBID}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("outlet reassignment after shift close: %d %s", w.Code, w.Body.String())
	}
}

func TestTransactionAndMovementListsFilterByOutlet(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Filter Outlet")
	w, outlets := api.doReq(http.MethodGet, "/outlets", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("list outlets: %d %s", w.Code, w.Body.String())
	}
	outletAID := uint(outlets["items"].([]any)[0].(map[string]any)["id"].(float64))
	w, product := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Filter Outlet", "sku": "FILTER-OUTLET", "sellPrice": 1000, "stock": 5,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create product: %d %s", w.Code, w.Body.String())
	}
	productID := uint(product["id"].(float64))
	w, outletB := api.doReq(http.MethodPost, "/outlets", map[string]any{"name": "Cabang Filter"}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create outlet B: %d %s", w.Code, w.Body.String())
	}
	outletBID := uint(outletB["id"].(float64))
	w, switchedB := api.doReq(http.MethodPost, "/outlets/switch", map[string]any{"outletId": outletBID}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("switch to outlet B: %d %s", w.Code, w.Body.String())
	}
	adminTokenB := switchedB["access_token"].(string)
	w, _ = api.doReq(http.MethodPost, "/stock/adjustments", map[string]any{
		"productId": productID, "direction": "plus", "qty": 5, "reason": "stok outlet B",
	}, adminTokenB)
	if w.Code != http.StatusOK {
		t.Fatalf("add stock at outlet B: %d %s", w.Code, w.Body.String())
	}
	w, saleB := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "Cash", "paid": 1000,
	}, adminTokenB)
	if w.Code != http.StatusCreated {
		t.Fatalf("sale at outlet B: %d %s", w.Code, w.Body.String())
	}
	w, switchedA := api.doReq(http.MethodPost, "/outlets/switch", map[string]any{"outletId": outletAID}, adminTokenB)
	if w.Code != http.StatusOK {
		t.Fatalf("switch to outlet A: %d %s", w.Code, w.Body.String())
	}
	adminTokenA := switchedA["access_token"].(string)
	w, saleA := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "Cash", "paid": 1000,
	}, adminTokenA)
	if w.Code != http.StatusCreated {
		t.Fatalf("sale at outlet A: %d %s", w.Code, w.Body.String())
	}
	trxAID := uint(saleA["id"].(float64))
	trxBID := uint(saleB["id"].(float64))

	for _, tc := range []struct {
		outletID uint
		trxID    uint
	}{{outletAID, trxAID}, {outletBID, trxBID}} {
		w, list := api.doReq(http.MethodGet, fmt.Sprintf("/transactions?outletId=%d", tc.outletID), nil, adminTokenA)
		items, ok := list["items"].([]any)
		if w.Code != http.StatusOK || !ok || len(items) != 1 || uint(items[0].(map[string]any)["id"].(float64)) != tc.trxID {
			t.Fatalf("transaction list for outlet %d: %d %s", tc.outletID, w.Code, w.Body.String())
		}
	}
	w, allTransactions := api.doReq(http.MethodGet, "/transactions", nil, adminTokenA)
	if w.Code != http.StatusOK || len(allTransactions["items"].([]any)) != 2 {
		t.Fatalf("admin transaction list without outletId must retain aggregate behavior: %d %s", w.Code, w.Body.String())
	}

	w, movementList := api.doReq(http.MethodGet, fmt.Sprintf("/movements?outletId=%d", outletBID), nil, adminTokenA)
	movements, ok := movementList["items"].([]any)
	if w.Code != http.StatusOK || !ok || len(movements) == 0 {
		t.Fatalf("movement list for outlet B: %d %s", w.Code, w.Body.String())
	}
	for _, movement := range movements {
		if uint(movement.(map[string]any)["outlet_id"].(float64)) != outletBID {
			t.Fatalf("movement filter leaked another outlet: %v", movement)
		}
	}
	w, movementList = api.doReq(http.MethodGet, "/movements", nil, adminTokenA)
	movements, ok = movementList["items"].([]any)
	if w.Code != http.StatusOK || !ok || len(movements) == 0 {
		t.Fatalf("movement list without outletId: %d %s", w.Code, w.Body.String())
	}
	for _, movement := range movements {
		if uint(movement.(map[string]any)["outlet_id"].(float64)) != outletAID {
			t.Fatalf("movement list without outletId must retain active outlet behavior: %v", movement)
		}
	}

	w, cashier := api.doReq(http.MethodPost, "/users", map[string]any{"name": "Kasir Filter", "outletId": outletAID}, adminTokenA)
	if w.Code != http.StatusCreated {
		t.Fatalf("create cashier at outlet B: %d %s", w.Code, w.Body.String())
	}
	cashierID := uint(cashier["user"].(map[string]any)["id"].(float64))
	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/users/%d/passcode", cashierID), map[string]any{"passcode": "12345"}, adminTokenA)
	if w.Code != http.StatusOK {
		t.Fatalf("set cashier passcode: %d %s", w.Code, w.Body.String())
	}
	w, cashierSession := api.doReq(http.MethodPost, "/auth/switch", map[string]any{"target_user_id": cashierID, "passcode": "12345"}, adminTokenA)
	if w.Code != http.StatusOK {
		t.Fatalf("switch to cashier: %d %s", w.Code, w.Body.String())
	}
	cashierToken := cashierSession["access_token"].(string)
	w, historicalCashierSale := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "Cash", "paid": 1000,
	}, cashierToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("cashier sale at outlet A: %d %s", w.Code, w.Body.String())
	}
	historicalCashierTrxID := uint(historicalCashierSale["id"].(float64))
	// Gerbang shift: kasir wajib menutup shift sebelum bisa pindah outlet.
	api.doReq(http.MethodPost, "/cashier/shift/close", map[string]any{"closingCash": 1000}, cashierToken)
	w, movedCashier := api.doReq(http.MethodPatch, fmt.Sprintf("/users/%d", cashierID), map[string]any{"outletId": outletBID}, adminTokenA)
	if w.Code != http.StatusOK {
		t.Fatalf("move cashier to outlet B: %d %s", w.Code, w.Body.String())
	}
	w, cashierSession = api.doReq(http.MethodPost, "/auth/switch", map[string]any{"target_user_id": cashierID, "passcode": "12345"}, adminTokenA)
	if w.Code != http.StatusOK || outletIDFromUser(t, cashierSession) != outletBID {
		t.Fatalf("switch to cashier after outlet move: %d %s %v", w.Code, w.Body.String(), movedCashier)
	}
	cashierToken = cashierSession["access_token"].(string)
	w, cashierSale := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "Cash", "paid": 1000,
	}, cashierToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("cashier sale at outlet B: %d %s", w.Code, w.Body.String())
	}
	cashierTrxID := uint(cashierSale["id"].(float64))
	for _, path := range []string{"/transactions", fmt.Sprintf("/transactions?outletId=%d", outletAID)} {
		w, list := api.doReq(http.MethodGet, path, nil, cashierToken)
		items, ok := list["items"].([]any)
		if w.Code != http.StatusOK || !ok || len(items) != 1 || uint(items[0].(map[string]any)["id"].(float64)) != cashierTrxID || cashierTrxID == historicalCashierTrxID {
			t.Fatalf("cashier transaction list must stay in active outlet for %s: %d %s", path, w.Code, w.Body.String())
		}
	}
}

func TestFreshAdminLoginUsesCurrentDefaultOutlet(t *testing.T) {
	api := newPromoTestAPI(t)
	email := fmt.Sprintf("outlet-login-%d@example.com", time.Now().UnixNano())
	var otp string
	previous := service.TestOnOTPSent
	service.TestOnOTPSent = func(to, code string) {
		if to == email {
			otp = code
		}
	}
	t.Cleanup(func() { service.TestOnOTPSent = previous })
	w, _ := api.doReq(http.MethodPost, "/auth/otp/send", map[string]any{"email": email}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("send registration OTP: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, "/auth/otp/verify", map[string]any{"email": email, "code": otp}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("verify registration OTP: %d %s", w.Code, w.Body.String())
	}
	w, registered := api.doReq(http.MethodPost, "/auth/register", map[string]any{
		"name": "Admin Login Outlet", "email": email, "password": "password123", "storeName": "Toko Login Outlet",
	}, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("register admin: %d %s", w.Code, w.Body.String())
	}
	adminToken := registered["access_token"].(string)
	w, outletB := api.doReq(http.MethodPost, "/outlets", map[string]any{"name": "Outlet Default Baru"}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create new default outlet: %d %s", w.Code, w.Body.String())
	}
	outletBID := uint(outletB["id"].(float64))
	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/outlets/%d", outletBID), map[string]any{
		"name": "Outlet Default Baru", "address": "", "isDefault": true,
	}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("set new default outlet: %d %s", w.Code, w.Body.String())
	}
	w, loggedIn := api.doReq(http.MethodPost, "/auth/login", map[string]any{"email": email, "password": "password123"}, "")
	if w.Code != http.StatusOK || outletIDFromUser(t, loggedIn) != outletBID {
		t.Fatalf("fresh login must select current default outlet: %d %s", w.Code, w.Body.String())
	}
}

func outletIDFromUser(t *testing.T, response map[string]any) uint {
	t.Helper()
	user, ok := response["user"].(map[string]any)
	if !ok {
		user = response
	}
	id, ok := user["outlet_id"].(float64)
	if !ok || id <= 0 {
		t.Fatalf("user missing active outlet_id: %v", user)
	}
	return uint(id)
}

func TestOutletStockSeparatesSalesAndCatalogViews(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Stok Outlet")
	w, outletsA := api.doReq(http.MethodGet, "/outlets", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("list default outlet: %d %s", w.Code, w.Body.String())
	}
	defaultOutletID := uint(outletsA["items"].([]any)[0].(map[string]any)["id"].(float64))
	w, product := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Kopi Outlet", "sku": "KOPI-OUTLET", "sellPrice": 1000, "stock": 5,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create product: %d %s", w.Code, w.Body.String())
	}
	productID := uint(product["id"].(float64))
	w, zeroProduct := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Hanya Outlet A", "sku": "A-ONLY-OUTLET", "sellPrice": 1000, "stock": 4,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create product with A-only stock: %d %s", w.Code, w.Body.String())
	}
	zeroProductID := uint(zeroProduct["id"].(float64))
	w, listA := api.doReq(http.MethodGet, "/products", nil, adminToken)
	if w.Code != http.StatusOK || productStockFromList(t, listA, productID) != 5 {
		t.Fatalf("default outlet initial stock: %d %s", w.Code, w.Body.String())
	}

	w, outletB := api.doReq(http.MethodPost, "/outlets", map[string]any{"name": "Cabang B"}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create outlet B: %d %s", w.Code, w.Body.String())
	}
	outletBID := uint(outletB["id"].(float64))
	w, switchedB := api.doReq(http.MethodPost, "/outlets/switch", map[string]any{"outletId": outletBID}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("switch to outlet B: %d %s", w.Code, w.Body.String())
	}
	outletBToken := switchedB["access_token"].(string)
	w, listB := api.doReq(http.MethodGet, "/products", nil, outletBToken)
	if w.Code != http.StatusOK || productStockFromList(t, listB, productID) != 0 || productStockFromList(t, listB, zeroProductID) != 0 {
		t.Fatalf("new outlet must show zero for products without outlet stock: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, "/stock/adjustments", map[string]any{
		"productId": productID, "direction": "plus", "qty": 7, "reason": "stok cabang B",
	}, outletBToken)
	if w.Code != http.StatusOK {
		t.Fatalf("add stock in outlet B: %d %s", w.Code, w.Body.String())
	}
	w, listB = api.doReq(http.MethodGet, "/products", nil, outletBToken)
	if w.Code != http.StatusOK || productStockFromList(t, listB, productID) != 7 {
		t.Fatalf("outlet B product stock: %d %s", w.Code, w.Body.String())
	}

	w, switchedA := api.doReq(http.MethodPost, "/outlets/switch", map[string]any{"outletId": defaultOutletID}, outletBToken)
	if w.Code != http.StatusOK {
		t.Fatalf("switch back to outlet A: %d %s", w.Code, w.Body.String())
	}
	outletAToken := switchedA["access_token"].(string)
	w, sale := api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 2}}, "method": "Cash", "paid": 2000,
	}, outletAToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("sell from outlet A: %d %s", w.Code, w.Body.String())
	}
	finalStocks := sale["final_stocks"].([]any)
	if len(finalStocks) != 1 || finalStocks[0].(map[string]any)["stock"] != float64(3) {
		t.Fatalf("final stock must be read from outlet A: %v", sale["final_stocks"])
	}
	w, listA = api.doReq(http.MethodGet, "/products", nil, outletAToken)
	if w.Code != http.StatusOK || productStockFromList(t, listA, productID) != 3 {
		t.Fatalf("outlet A stock after sale: %d %s", w.Code, w.Body.String())
	}
	w, listB = api.doReq(http.MethodGet, "/products", nil, outletBToken)
	if w.Code != http.StatusOK || productStockFromList(t, listB, productID) != 7 {
		t.Fatalf("sale in A changed outlet B stock: %d %s", w.Code, w.Body.String())
	}

	database, err := gorm.Open(sqlite.Open(os.Getenv("DATABASE_URL")), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	var stocks []model.ProductStock
	if err := database.Where("product_id = ?", productID).Find(&stocks).Error; err != nil {
		t.Fatal(err)
	}
	stockByOutlet := map[uint]int{}
	for _, stock := range stocks {
		stockByOutlet[stock.OutletID] = stock.Qty
	}
	if stockByOutlet[defaultOutletID] != 3 || stockByOutlet[outletBID] != 7 {
		t.Fatalf("product_stocks must keep outlet quantities separate: %v", stockByOutlet)
	}
	var legacyStock struct{ Stock int }
	if err := database.Table("products").Select("stock").Where("id = ?", productID).Scan(&legacyStock).Error; err != nil {
		t.Fatal(err)
	}
	if legacyStock.Stock != 3 {
		t.Fatalf("legacy product stock must mirror the default outlet: got %d want 3", legacyStock.Stock)
	}
	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/outlets/%d", outletBID), map[string]any{
		"name": "Cabang B", "address": "", "isDefault": true,
	}, outletAToken)
	if w.Code != http.StatusOK {
		t.Fatalf("change default outlet: %d %s", w.Code, w.Body.String())
	}
	if err := database.Table("products").Select("stock").Where("id = ?", productID).Scan(&legacyStock).Error; err != nil {
		t.Fatal(err)
	}
	if legacyStock.Stock != 7 {
		t.Fatalf("legacy product stock must follow the newly selected default outlet: got %d want 7", legacyStock.Stock)
	}
	if err := database.Table("products").Select("stock").Where("id = ?", zeroProductID).Scan(&legacyStock).Error; err != nil {
		t.Fatal(err)
	}
	if legacyStock.Stock != 0 {
		t.Fatalf("new default without a stock row must mirror zero: got %d want 0", legacyStock.Stock)
	}
}

func TestOutletReportsAggregateAndFilterWithoutCrossStoreLeak(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Laporan Outlet")
	w, outlets := api.doReq(http.MethodGet, "/outlets", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("list default outlet: %d %s", w.Code, w.Body.String())
	}
	outletAID := uint(outlets["items"].([]any)[0].(map[string]any)["id"].(float64))
	w, product := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Laporan Outlet", "sku": "LAPORAN-OUTLET", "sellPrice": 1000, "stock": 5,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create report product: %d %s", w.Code, w.Body.String())
	}
	productID := uint(product["id"].(float64))
	w, outletB := api.doReq(http.MethodPost, "/outlets", map[string]any{"name": "Cabang Laporan"}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create report outlet B: %d %s", w.Code, w.Body.String())
	}
	outletBID := uint(outletB["id"].(float64))
	w, switchedB := api.doReq(http.MethodPost, "/outlets/switch", map[string]any{"outletId": outletBID}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("switch to report outlet B: %d %s", w.Code, w.Body.String())
	}
	outletBToken := switchedB["access_token"].(string)
	w, _ = api.doReq(http.MethodPost, "/stock/adjustments", map[string]any{
		"productId": productID, "direction": "plus", "qty": 7, "reason": "stok laporan outlet B",
	}, outletBToken)
	if w.Code != http.StatusOK {
		t.Fatalf("stock outlet B: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 2}}, "method": "Cash", "paid": 2000,
	}, outletBToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("sale outlet B: %d %s", w.Code, w.Body.String())
	}
	w, switchedA := api.doReq(http.MethodPost, "/outlets/switch", map[string]any{"outletId": outletAID}, outletBToken)
	if w.Code != http.StatusOK {
		t.Fatalf("switch to report outlet A: %d %s", w.Code, w.Body.String())
	}
	outletAToken := switchedA["access_token"].(string)
	w, _ = api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "Cash", "paid": 1000,
	}, outletAToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("sale outlet A: %d %s", w.Code, w.Body.String())
	}

	w, dashboard := api.doReq(http.MethodGet, "/dashboard", nil, outletAToken)
	today, ok := dashboard["today"].(map[string]any)
	if w.Code != http.StatusOK || !ok || today["omzet"] != float64(3000) {
		t.Fatalf("admin dashboard without outletId must combine both outlets: %d %s", w.Code, w.Body.String())
	}
	w, dashboard = api.doReq(http.MethodGet, fmt.Sprintf("/dashboard?outletId=%d", outletBID), nil, outletAToken)
	today = dashboard["today"].(map[string]any)
	if w.Code != http.StatusOK || today["omzet"] != float64(2000) {
		t.Fatalf("admin dashboard outlet B filter: %d %s", w.Code, w.Body.String())
	}
	w, dashboard = api.doReq(http.MethodGet, fmt.Sprintf("/dashboard?outletId=%d", outletAID), nil, outletAToken)
	today = dashboard["today"].(map[string]any)
	if w.Code != http.StatusOK || today["omzet"] != float64(1000) {
		t.Fatalf("admin dashboard outlet A filter: %d %s", w.Code, w.Body.String())
	}

	w, report := api.doReq(http.MethodGet, "/reports?period=today", nil, outletAToken)
	summary, ok := report["summary"].(map[string]any)
	if w.Code != http.StatusOK || !ok || summary["omzet"] != float64(3000) {
		t.Fatalf("admin report without outletId must combine both outlets: %d %s", w.Code, w.Body.String())
	}
	w, report = api.doReq(http.MethodGet, fmt.Sprintf("/reports?period=today&outletId=%d", outletBID), nil, outletAToken)
	summary = report["summary"].(map[string]any)
	if w.Code != http.StatusOK || summary["omzet"] != float64(2000) {
		t.Fatalf("admin report outlet B filter: %d %s", w.Code, w.Body.String())
	}
	w, tax := api.doReq(http.MethodGet, fmt.Sprintf("/reports/tax?outletId=%d", outletBID), nil, outletAToken)
	if w.Code != http.StatusOK || tax["omzet_bruto"] != float64(2000) {
		t.Fatalf("tax report outlet B filter: %d %s", w.Code, w.Body.String())
	}

	otherAdminToken := api.registerAdmin("Toko Laporan Asing")
	w, foreignOutlets := api.doReq(http.MethodGet, "/outlets", nil, otherAdminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("list foreign default outlet: %d %s", w.Code, w.Body.String())
	}
	foreignOutletID := uint(foreignOutlets["items"].([]any)[0].(map[string]any)["id"].(float64))
	w, dashboard = api.doReq(http.MethodGet, fmt.Sprintf("/dashboard?outletId=%d", foreignOutletID), nil, outletAToken)
	if w.Code != http.StatusNotFound || dashboard["error"] == nil {
		t.Fatalf("foreign outlet filter must be rejected without disclosing sales: %d %s", w.Code, w.Body.String())
	}
	w, tax = api.doReq(http.MethodGet, fmt.Sprintf("/reports/tax?outletId=%d", foreignOutletID), nil, outletAToken)
	if w.Code != http.StatusNotFound || tax["error"] == nil {
		t.Fatalf("foreign outlet tax filter must be rejected without disclosing sales: %d %s", w.Code, w.Body.String())
	}

	cashierToken := api.createCashier(outletAToken)
	w, _ = api.doReq(http.MethodPost, "/transactions", map[string]any{
		"items": []any{map[string]any{"productId": productID, "qty": 1}}, "method": "Cash", "paid": 1000,
	}, cashierToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("cashier sale in bound outlet: %d %s", w.Code, w.Body.String())
	}
	w, dashboard = api.doReq(http.MethodGet, "/dashboard", nil, cashierToken)
	today = dashboard["today"].(map[string]any)
	if w.Code != http.StatusOK || dashboard["role"] != "cashier" || today["omzet"] != float64(1000) {
		t.Fatalf("cashier dashboard without outletId must show only bound outlet sales: %d %s", w.Code, w.Body.String())
	}
}

func TestPurchaseReceiptAndStockOpnameStayInActiveOutlet(t *testing.T) {
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Operasi Outlet")
	w, outlets := api.doReq(http.MethodGet, "/outlets", nil, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("list default outlet: %d %s", w.Code, w.Body.String())
	}
	outletAID := uint(outlets["items"].([]any)[0].(map[string]any)["id"].(float64))
	productID := createPurchaseTestProduct(t, api, adminToken, "Produk Operasi Outlet", "OP-OUTLET", 5)
	supplierID := createPurchaseTestSupplier(t, api, adminToken, "Supplier Operasi Outlet")
	w, outletB := api.doReq(http.MethodPost, "/outlets", map[string]any{"name": "Cabang Operasi"}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create outlet B: %d %s", w.Code, w.Body.String())
	}
	outletBID := uint(outletB["id"].(float64))
	w, switchedB := api.doReq(http.MethodPost, "/outlets/switch", map[string]any{"outletId": outletBID}, adminToken)
	if w.Code != http.StatusOK {
		t.Fatalf("switch to outlet B: %d %s", w.Code, w.Body.String())
	}
	outletBToken := switchedB["access_token"].(string)
	w, po := api.doReq(http.MethodPost, "/purchase-orders", map[string]any{
		"supplierId": supplierID,
		"items":      []any{map[string]any{"productId": productID, "qty": 4, "buyPrice": 200}},
	}, outletBToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create PO in B: %d %s", w.Code, w.Body.String())
	}
	poID := uint(po["id"].(float64))
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/order", poID), map[string]any{}, outletBToken)
	if w.Code != http.StatusOK {
		t.Fatalf("order PO in B: %d %s", w.Code, w.Body.String())
	}
	w, po = api.doReq(http.MethodGet, fmt.Sprintf("/purchase-orders/%d", poID), nil, outletBToken)
	if w.Code != http.StatusOK {
		t.Fatalf("get PO in B: %d %s", w.Code, w.Body.String())
	}
	poItem := po["items"].([]any)[0].(map[string]any)
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/purchase-orders/%d/receive", poID), map[string]any{
		"items": []any{map[string]any{"itemId": uint(poItem["id"].(float64)), "qty": 4}},
	}, outletBToken)
	if w.Code != http.StatusOK {
		t.Fatalf("receive PO in B: %d %s", w.Code, w.Body.String())
	}
	if stock, _ := purchaseTestStock(t, api, outletBToken, productID, 0); stock != 4 {
		t.Fatalf("PO receipt should add stock in B: got %d want 4", stock)
	}
	_, switchedA := api.doReq(http.MethodPost, "/outlets/switch", map[string]any{"outletId": outletAID}, outletBToken)
	outletAToken := switchedA["access_token"].(string)
	if stock, _ := purchaseTestStock(t, api, outletAToken, productID, 0); stock != 5 {
		t.Fatalf("PO receipt in B changed stock in A: got %d want 5", stock)
	}
	w, poList := api.doReq(http.MethodGet, "/purchase-orders", nil, outletAToken)
	if w.Code != http.StatusOK || len(poList["items"].([]any)) != 0 {
		t.Fatalf("PO list in A must not include B's PO: %d %s", w.Code, w.Body.String())
	}

	w, session := api.doReq(http.MethodPost, "/stock-opnames", map[string]any{"note": "Hitung outlet B"}, outletBToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create opname in B: %d %s", w.Code, w.Body.String())
	}
	sessionID := uint(session["id"].(float64))
	item := stockOpnameItem(t, session["items"], productID, 0)
	if item["system_stock"] != float64(4) {
		t.Fatalf("opname snapshot in B should read stock B: %v", item)
	}
	w, _ = api.doReq(http.MethodPut, fmt.Sprintf("/stock-opnames/%d/items", sessionID), map[string]any{
		"items": []any{map[string]any{"itemId": uint(item["id"].(float64)), "countedStock": 6}},
	}, outletBToken)
	if w.Code != http.StatusOK {
		t.Fatalf("count outlet B: %d %s", w.Code, w.Body.String())
	}
	w, _ = api.doReq(http.MethodPost, fmt.Sprintf("/stock-opnames/%d/post", sessionID), map[string]any{}, outletBToken)
	if w.Code != http.StatusOK {
		t.Fatalf("post opname in B: %d %s", w.Code, w.Body.String())
	}
	if stock, _ := purchaseTestStock(t, api, outletBToken, productID, 0); stock != 6 {
		t.Fatalf("opname should set stock in B: got %d want 6", stock)
	}
	if stock, _ := purchaseTestStock(t, api, outletAToken, productID, 0); stock != 5 {
		t.Fatalf("opname in B changed stock in A: got %d want 5", stock)
	}
	w, sessionA := api.doReq(http.MethodPost, "/stock-opnames", map[string]any{"note": "Hitung outlet A"}, outletAToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create opname in A after B was posted: %d %s", w.Code, w.Body.String())
	}
	w, sessionsA := api.doReq(http.MethodGet, "/stock-opnames", nil, outletAToken)
	if w.Code != http.StatusOK || len(sessionsA["items"].([]any)) != 1 || uint(sessionsA["items"].([]any)[0].(map[string]any)["id"].(float64)) != uint(sessionA["id"].(float64)) {
		t.Fatalf("opname list in A must show only its session: %d %s", w.Code, w.Body.String())
	}
	w, sessionsB := api.doReq(http.MethodGet, "/stock-opnames", nil, outletBToken)
	if w.Code != http.StatusOK || len(sessionsB["items"].([]any)) != 1 || uint(sessionsB["items"].([]any)[0].(map[string]any)["id"].(float64)) != uint(session["id"].(float64)) {
		t.Fatalf("opname list in B must show only its session: %d %s", w.Code, w.Body.String())
	}
}

func productStockFromList(t *testing.T, response map[string]any, productID uint) int {
	t.Helper()
	items, ok := response["items"].([]any)
	if !ok {
		t.Fatalf("product response missing items: %v", response)
	}
	for _, item := range items {
		product := item.(map[string]any)
		if uint(product["id"].(float64)) == productID {
			return int(product["stock"].(float64))
		}
	}
	t.Fatalf("product %d missing from list: %v", productID, response)
	return -1
}
