package integration

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func (api *promoTestAPI) doCategoryImport(t *testing.T, csvContent, token string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	part, err := mw.CreateFormFile("file", "kategori.csv")
	if err != nil {
		t.Fatalf("multipart: %v", err)
	}
	if _, err := part.Write([]byte(csvContent)); err != nil {
		t.Fatalf("write csv: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close multipart: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/api/v1/categories/import", &buf)
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

func TestCategoryImportPartialSuccessAndBlankRows(t *testing.T) {
	api := newPromoTestAPI(t)
	token := api.registerAdmin("Import Kategori")
	if w, _ := api.doReq(http.MethodPost, "/categories", map[string]any{"name": "Sudah Ada"}, token); w.Code != http.StatusCreated {
		t.Fatalf("seed category: %d %s", w.Code, w.Body.String())
	}

	csv := "\ufeffname\n" +
		"Kategori Baru\n" +
		"sudah ada\n" +
		"Dalam File\n" +
		" dalam file \n" +
		",kolom kedua\n" +
		",,\n"
	w, result := api.doCategoryImport(t, csv, token)
	if w.Code != http.StatusOK {
		t.Fatalf("import: %d %s", w.Code, w.Body.String())
	}
	if result["total"] != float64(5) || result["berhasil"] != float64(2) || result["gagal"] != float64(3) {
		t.Fatalf("ringkasan salah: %v", result)
	}
	failures, ok := result["kesalahan"].([]any)
	if !ok || len(failures) != 3 {
		t.Fatalf("kesalahan harus berisi tiga baris: %v", result["kesalahan"])
	}
	byRow := map[float64]string{}
	for _, failure := range failures {
		row := failure.(map[string]any)
		byRow[row["baris"].(float64)] = row["alasan"].(string)
	}
	if !strings.Contains(byRow[3], "sudah ada") {
		t.Fatalf("baris 3 harus melaporkan kategori lama duplikat: %v", byRow)
	}
	if !strings.Contains(byRow[5], "sudah ada") {
		t.Fatalf("baris 5 harus melaporkan duplikat dalam file: %v", byRow)
	}
	if byRow[6] != "Nama kategori wajib diisi." {
		t.Fatalf("baris 6 harus melaporkan nama kosong: %v", byRow)
	}

	w, categories := api.doReq(http.MethodGet, "/categories", nil, token)
	if w.Code != http.StatusOK {
		t.Fatalf("list categories: %d %s", w.Code, w.Body.String())
	}
	names := map[string]bool{}
	for _, item := range categories["items"].([]any) {
		names[item.(map[string]any)["name"].(string)] = true
	}
	for _, name := range []string{"Sudah Ada", "Kategori Baru", "Dalam File"} {
		if !names[name] {
			t.Fatalf("kategori %q tidak terlihat setelah impor: %v", name, categories)
		}
	}
}

func TestCategoryImportIsIsolatedByStoreAndAdminOnly(t *testing.T) {
	api := newPromoTestAPI(t)
	tokenA := api.registerAdmin("Toko A")
	tokenB := api.registerAdmin("Toko B")
	if w, _ := api.doReq(http.MethodPost, "/categories", map[string]any{"name": "Hanya Toko A"}, tokenA); w.Code != http.StatusCreated {
		t.Fatalf("seed category toko A: %d %s", w.Code, w.Body.String())
	}

	cashierToken := api.createCashier(tokenB)
	w, _ := api.doCategoryImport(t, "nama\nDitolak\n", cashierToken)
	if w.Code != http.StatusForbidden {
		t.Fatalf("kasir harus ditolak: %d %s", w.Code, w.Body.String())
	}

	w, result := api.doCategoryImport(t, "nama\nHanya Toko A\nHanya Toko B\n", tokenB)
	if w.Code != http.StatusOK {
		t.Fatalf("import toko B: %d %s", w.Code, w.Body.String())
	}
	if result["total"] != float64(2) || result["berhasil"] != float64(2) || result["gagal"] != float64(0) {
		t.Fatalf("kategori toko lain tidak boleh menghalangi impor: %v", result)
	}
	if failures, ok := result["kesalahan"].([]any); !ok || len(failures) != 0 {
		t.Fatalf("kesalahan kosong harus berupa array kosong: %v", result["kesalahan"])
	}

	w, categoriesA := api.doReq(http.MethodGet, "/categories", nil, tokenA)
	if w.Code != http.StatusOK {
		t.Fatalf("list toko A: %d %s", w.Code, w.Body.String())
	}
	for _, item := range categoriesA["items"].([]any) {
		if item.(map[string]any)["name"] == "Hanya Toko B" {
			t.Fatalf("kategori toko B bocor ke toko A: %v", categoriesA)
		}
	}

	w, categoriesB := api.doReq(http.MethodGet, "/categories", nil, tokenB)
	if w.Code != http.StatusOK {
		t.Fatalf("list toko B: %d %s", w.Code, w.Body.String())
	}
	w, me := api.doReq(http.MethodGet, "/auth/me", nil, tokenB)
	if w.Code != http.StatusOK {
		t.Fatalf("me toko B: %d %s", w.Code, w.Body.String())
	}
	storeB := me["user"].(map[string]any)["store_id"].(float64)
	foundSameNameInB := false
	foundBCategory := false
	for _, item := range categoriesB["items"].([]any) {
		category := item.(map[string]any)
		if category["name"] == "Hanya Toko A" {
			foundSameNameInB = true
			if category["store_id"] != storeB {
				t.Fatalf("kategori dengan nama sama harus dibuat untuk toko B: %v", category)
			}
		}
		if category["name"] == "Hanya Toko B" {
			foundBCategory = true
			if category["store_id"] != storeB {
				t.Fatalf("kategori baru harus dimiliki toko B: %v", category)
			}
		}
	}
	if !foundSameNameInB || !foundBCategory {
		t.Fatalf("kategori impor toko B tidak lengkap: %v", categoriesB)
	}
}

func TestCategoryImportRejectsEmptyFileAndWrongHeader(t *testing.T) {
	api := newPromoTestAPI(t)
	token := api.registerAdmin("Validasi Kategori")

	w, response := api.doCategoryImport(t, "", token)
	if w.Code != http.StatusBadRequest || response["error"] != "File CSV kosong." {
		t.Fatalf("file kosong harus ditolak, dapat %d %v", w.Code, response)
	}

	w, response = api.doCategoryImport(t, "kategori\nMakanan\n", token)
	if w.Code != http.StatusBadRequest || !strings.Contains(response["error"].(string), "satu kolom") {
		t.Fatalf("header salah harus ditolak, dapat %d %v", w.Code, response)
	}
}
