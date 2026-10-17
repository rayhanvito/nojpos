package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers"
	legacyrouter "github.com/getkin/kin-openapi/routers/legacy"
)

// loadOpenAPISpec memuat api/openapi.yaml, memvalidasi dokumennya, dan
// membangun router pencocok rute dari dokumen yang sama.
func loadOpenAPISpec(t *testing.T) routers.Router {
	t.Helper()
	loader := &openapi3.Loader{IsExternalRefsAllowed: false}
	doc, err := loader.LoadFromFile("../../api/openapi.yaml")
	if err != nil {
		t.Fatalf("muat openapi.yaml: %v", err)
	}
	if err := doc.Validate(context.Background()); err != nil {
		t.Fatalf("dokumen openapi.yaml tidak valid: %v", err)
	}
	rtr, err := legacyrouter.NewRouter(doc)
	if err != nil {
		t.Fatalf("bangun router openapi: %v", err)
	}
	return rtr
}

// doContract menjalankan satu request HTTP nyata lalu memvalidasi respons
// terhadap spec: status harus terdokumentasi untuk path+method dan body
// harus lolos schema respons operasi tersebut.
func (api *promoTestAPI) doContract(doc routers.Router, method, path string, body any, token string, wantStatus int) *httptest.ResponseRecorder {
	api.t.Helper()
	endpoint := method + " " + path
	var bodyBytes []byte
	if body != nil {
		bodyBytes, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(bodyBytes))
	req.RemoteAddr = api.remote
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	route, pathParams, err := doc.FindRoute(req)
	if err != nil {
		api.t.Fatalf("%s: rute tidak ditemukan di spec: %v", endpoint, err)
	}
	w := httptest.NewRecorder()
	api.handler.ServeHTTP(w, req)
	if w.Code != wantStatus {
		api.t.Fatalf("%s: status = %d, mau %d (body: %s)", endpoint, w.Code, wantStatus, w.Body.String())
	}
	respInput := &openapi3filter.ResponseValidationInput{
		RequestValidationInput: &openapi3filter.RequestValidationInput{
			Request:    req,
			PathParams: pathParams,
			Route:      route,
		},
		Status: w.Code,
		Header: w.Header(),
		Body:   io.NopCloser(bytes.NewReader(w.Body.Bytes())),
	}
	if err := openapi3filter.ValidateResponse(context.Background(), respInput); err != nil {
		api.t.Fatalf("%s: respons tidak lolos validasi kontrak OpenAPI: %v (body: %s)", endpoint, err, w.Body.String())
	}
	return w
}

func contractPromoBody(name string, value float64) map[string]any {
	return map[string]any{
		"name": name, "type": "percent", "value": value, "scope": "all",
		"startsAt":    time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
		"endsAt":      time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
		"minPurchase": 0, "maxDiscount": 0,
	}
}

func decodeContractBody(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
		t.Fatalf("body bukan JSON objek: %v (%s)", err, w.Body.String())
	}
	return out
}

// TestOpenAPIContractPromo memvalidasi 9 skenario domain promo terhadap
// api/openapi.yaml: CRUD admin /promos, PATCH active, POST
// /promos/preview, 404 sesudah hapus, dan 400 untuk masukan tidak valid.
func TestOpenAPIContractPromo(t *testing.T) {
	doc := loadOpenAPISpec(t)
	api := newPromoTestAPI(t)
	adminToken := api.registerAdmin("Toko Kontrak OpenAPI")

	// (1) POST /promos persen valid -> 201. Persen desimal dipakai supaya
	// field "value" tak ambigu terhadap oneOf[number, integer] di spec.
	w := api.doContract(doc, http.MethodPost, "/promos", contractPromoBody("Promo Kontrak", 12.5), adminToken, http.StatusCreated)
	created := decodeContractBody(t, w)
	promoID := uint(created["id"].(float64))
	promoPath := fmt.Sprintf("/promos/%d", promoID)

	// (2) GET /promos -> 200.
	api.doContract(doc, http.MethodGet, "/promos", nil, adminToken, http.StatusOK)

	// (3) GET /promos/{id} -> 200.
	api.doContract(doc, http.MethodGet, promoPath, nil, adminToken, http.StatusOK)

	// (4) PUT /promos/{id} -> 200.
	api.doContract(doc, http.MethodPut, promoPath, contractPromoBody("Promo Kontrak Revisi", 15.5), adminToken, http.StatusOK)

	// (5) PATCH /promos/{id}/active -> 200 (nonaktif lalu aktif lagi
	// supaya pratinjau di langkah 6 tetap berlaku).
	api.doContract(doc, http.MethodPatch, promoPath+"/active", map[string]any{"active": false}, adminToken, http.StatusOK)
	api.doContract(doc, http.MethodPatch, promoPath+"/active", map[string]any{"active": true}, adminToken, http.StatusOK)

	// (6) POST /promos/preview dengan item produk nyata -> 200.
	w, product := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Produk Kontrak", "sku": "KONTRAK-OPENAPI", "sellPrice": 10000, "stock": 10,
	}, adminToken)
	if w.Code != http.StatusCreated {
		t.Fatalf("create product: %d %s", w.Code, w.Body.String())
	}
	productID := uint(product["id"].(float64))
	previewBody := map[string]any{
		"items":       []any{map[string]any{"productId": productID, "qty": 1}},
		"voucherCode": "",
	}
	w = api.doContract(doc, http.MethodPost, "/promos/preview", previewBody, adminToken, http.StatusOK)
	if w.Body.String() == "null" {
		t.Fatalf("preview harus menghasilkan objek PromoPreview (promo otomatis aktif), dapat null")
	}

	// (7) DELETE /promos/{id} -> 200.
	api.doContract(doc, http.MethodDelete, promoPath, nil, adminToken, http.StatusOK)

	// (8) GET /promos/{id} sesudah hapus -> 404.
	api.doContract(doc, http.MethodGet, promoPath, nil, adminToken, http.StatusNotFound)

	// (9) POST /promos tidak valid (persen 0) -> 400.
	invalid := contractPromoBody("Promo Tidak Valid", 0)
	api.doContract(doc, http.MethodPost, "/promos", invalid, adminToken, http.StatusBadRequest)
}
