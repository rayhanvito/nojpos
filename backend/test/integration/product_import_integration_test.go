package integration

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func (api *promoTestAPI) doImport(t *testing.T, csvContent, token string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", "produk.csv")
	if err != nil {
		t.Fatalf("multipart: %v", err)
	}
	if _, err := part.Write([]byte(csvContent)); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products/import", &buf)
	req.RemoteAddr = api.remote
	req.Header.Set("Content-Type", mw.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	w := httptest.NewRecorder()
	api.handler.ServeHTTP(w, req)
	var response map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &response)
	return w, response
}

const importCSVFixture = "id,nama,sku,barcode,kategori,harga beli,harga jual,stok,satuan,aktif\n" +
	",Teh Melati,TEH-001,,Minuman,3000,6000,12,pcs,true\n" +
	",Gula 1kg,GULA-001,,Bahan,12000,15000,0,pack,false\n" +
	",SKU Lama,LEGACY-DEL,,Minuman,1000,2000,5,pcs,true\n" +
	",Duplikat File,TEH-001,,Minuman,3000,6000,1,pcs,true\n" +
	",Harga Rusak,BAD-001,,Minuman,abc,2000,1,pcs,true\n" +
	",Stok Awal,STOK-001,,Minuman,500,1000,7,pcs,true\n"

func TestProductImportPartialSuccess(t *testing.T) {
	api := newPromoTestAPI(t)
	token := api.registerAdmin("Import Store")

	// Produk yang sudah ada + satu soft-deleted dengan SKU yang akan ditolak.
	w, _ := api.doReq(http.MethodPost, "/products", map[string]any{
		"name": "Sudah Ada", "sku": "LEGACY-DEL", "buyPrice": 1, "sellPrice": 2, "stock": 0, "unit": "pcs",
	}, token)
	if w.Code != http.StatusCreated {
		t.Fatalf("seed product: %d %s", w.Code, w.Body.String())
	}
	w, list := api.doReq(http.MethodGet, "/products", nil, token)
	var legacyID float64
	for _, item := range list["items"].([]any) {
		if item.(map[string]any)["sku"] == "LEGACY-DEL" {
			legacyID = item.(map[string]any)["id"].(float64)
		}
	}
	w, _ = api.doReq(http.MethodDelete, "/products/"+itoa(legacyID), nil, token)
	if w.Code != http.StatusOK && w.Code != http.StatusNoContent {
		t.Fatalf("soft delete seed: %d %s", w.Code, w.Body.String())
	}

	w, res := api.doImport(t, importCSVFixture, token)
	if w.Code != http.StatusOK {
		t.Fatalf("import: %d %s", w.Code, w.Body.String())
	}
	if res["total"] != float64(6) || res["berhasil"] != float64(3) || res["gagal"] != float64(3) {
		t.Fatalf("ringkasan salah: %v", res)
	}
	kegagalan, _ := res["kesalahan"].([]any)
	if len(kegagalan) != 3 {
		t.Fatalf("kesalahan harus 3: %v", res["kesalahan"])
	}
	alasan := map[float64]string{}
	for _, k := range kegagalan {
		m := k.(map[string]any)
		alasan[m["baris"].(float64)] = m["alasan"].(string)
	}
	if !strings.Contains(alasan[4], "terdaftar") {
		t.Fatalf("baris 4 harus ditolak SKU terdaftar (soft-deleted): %v", alasan)
	}
	if !strings.Contains(alasan[5], "duplikat") {
		t.Fatalf("baris 5 harus ditolak duplikat file: %v", alasan)
	}
	if !strings.Contains(alasan[6], "harga beli") {
		t.Fatalf("baris 6 harus ditolak harga beli: %v", alasan)
	}

	// Kategori otomatis dibuat, produk nonaktif sesuai kolom aktif, stok awal tercatat sebagai adjust/import.
	w, list = api.doReq(http.MethodGet, "/products", nil, token)
	bySKU := map[string]map[string]any{}
	for _, item := range list["items"].([]any) {
		m := item.(map[string]any)
		bySKU[m["sku"].(string)] = m
	}
	if _, ok := bySKU["TEH-001"]; !ok {
		t.Fatalf("TEH-001 harus terimpor: %v", list)
	}
	if got := bySKU["STOK-001"]["stock"]; got != float64(7) {
		t.Fatalf("stok STOK-001 harus 7, dapat %v", got)
	}
	w, movs := api.doReq(http.MethodGet, "/movements", nil, token)
	foundImport := false
	for _, item := range movs["items"].([]any) {
		m := item.(map[string]any)
		if m["type"] == "adjust" && m["reason"] == "import" {
			foundImport = true
		}
	}
	if !foundImport {
		t.Fatalf("movement adjust dengan reason import tidak ditemukan: %v", movs)
	}
	w, cats := api.doReq(http.MethodGet, "/categories", nil, token)
	names := ""
	for _, item := range cats["items"].([]any) {
		names += item.(map[string]any)["name"].(string) + ","
	}
	if !strings.Contains(names, "Minuman") || !strings.Contains(names, "Bahan") {
		t.Fatalf("kategori harus otomatis dibuat: %s", names)
	}
}

func itoa(f float64) string {
	return strconv.FormatInt(int64(f), 10)
}
