package handler

import (
	"encoding/csv"
	"io"
	"net/http"
	"strings"

	"github.com/adrr-dev/openPOS/backend/middleware"
	"github.com/adrr-dev/openPOS/backend/service"
	"github.com/gin-gonic/gin"
)

// F2.8 — POST /products/import (admin): unggah CSV produk, create-only.
// Header wajib persis format ekspor F0.2, dengan atau tanpa kolom id di depan.
var importHeaderWithID = []string{"id", "nama", "sku", "barcode", "kategori", "harga beli", "harga jual", "stok", "satuan", "aktif"}
var importHeaderNoID = []string{"nama", "sku", "barcode", "kategori", "harga beli", "harga jual", "stok", "satuan", "aktif"}

func sameHeader(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if strings.ToLower(strings.TrimSpace(got[i])) != want[i] {
			return false
		}
	}
	return true
}

func (h *CatalogHandler) ImportProducts(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File CSV wajib diunggah pada field 'file'."})
		return
	}
	if fileHeader.Size > 5<<20 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Ukuran file melebihi 5 MB."})
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File tidak dapat dibaca."})
		return
	}
	defer f.Close()
	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	records, err := reader.ReadAll()
	if err != nil && err != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format CSV tidak valid: " + err.Error()})
		return
	}
	if len(records) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File CSV kosong."})
		return
	}
	header := records[0]
	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], "\ufeff")
	}
	hasID := sameHeader(header, importHeaderWithID)
	if !hasID && !sameHeader(header, importHeaderNoID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Header CSV harus persis format ekspor produk: " + strings.Join(importHeaderWithID, ", ") + " (kolom id boleh dikosongkan/dihilangkan)."})
		return
	}
	rows := records[1:]
	nonBlank := 0
	for _, rec := range rows {
		for _, cell := range rec {
			if strings.TrimSpace(cell) != "" {
				nonBlank++
				break
			}
		}
	}
	if nonBlank > service.MaxImportRows {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Maksimal 1.000 baris data per file."})
		return
	}
	result, err := h.svc.ImportProducts(c.Request.Context(), claims.StoreID, claims.OutletID, claims.Name, rows, hasID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal mengimpor produk."})
		return
	}
	c.JSON(http.StatusOK, result)
}
