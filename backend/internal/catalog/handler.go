package catalog

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/httpx"
	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
	"github.com/rayhanvito/nojpos/backend/internal/user"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler { return &Handler{svc: svc} }

type categoryReq struct {
	Name string `json:"name"`
}

func (h *Handler) ListCategories(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	cats, err := h.svc.ListCategories(c.Request.Context(), claims.StoreID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat kategori."})
		return
	}
	c.JSON(http.StatusOK, httpx.PaginateList(c, cats))
}

func (h *Handler) CreateCategory(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req categoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	cat, err := h.svc.CreateCategory(c.Request.Context(), claims.StoreID, req.Name)
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"category": cat})
}

func (h *Handler) RenameCategory(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req categoryReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	cat, err := h.svc.RenameCategory(c.Request.Context(), claims.StoreID, id, req.Name)
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"category": cat})
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	soft, err := h.svc.DeleteCategory(c.Request.Context(), claims.StoreID, id)
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"soft_deleted": soft})
}

type productReq struct {
	Name       string `json:"name"`
	SKU        string `json:"sku"`
	Barcode    string `json:"barcode"`
	CategoryID *uint  `json:"categoryId"`
	BuyPrice   *int64 `json:"buyPrice"`
	SellPrice  int64  `json:"sellPrice"`
	Stock      *int   `json:"stock"`
	MinStock   *int   `json:"minStock"`
	Unit       string `json:"unit"`
}

type variantReq struct {
	Name      string `json:"name"`
	SKU       string `json:"sku"`
	BuyPrice  int64  `json:"buyPrice"`
	SellPrice int64  `json:"sellPrice"`
	Stock     int    `json:"stock"`
	MinStock  int    `json:"minStock"`
}

type priceTierReq struct {
	MinQty    int   `json:"minQty"`
	SellPrice int64 `json:"sellPrice"`
}

func (h *Handler) ListProducts(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	q := c.Query

	f := ProductFilter{Q: q("q"), CategoryID: queryID(q, "categoryId")}
	if v := q("page"); v != "" {
		f.Page, _ = strconv.Atoi(v)
	}
	if v := q("limit"); v != "" {
		f.Limit, _ = strconv.Atoi(v)
	}
	if v := q("active"); v == "true" || v == "false" {
		b := v == "true"
		f.Active = &b
	}

	pageData, err := h.svc.ListProducts(c.Request.Context(), claims.StoreID, claims.OutletID, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat produk."})
		return
	}
	c.JSON(http.StatusOK, pageData)
}

func (h *Handler) GetProduct(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	p, err := h.svc.GetProduct(c.Request.Context(), claims.StoreID, claims.OutletID, id)
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) CreateProduct(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req productReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	p, err := h.svc.CreateProduct(c.Request.Context(), claims.StoreID, claims.OutletID, claims.Name, productInputFrom(&req))
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, p)
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req productReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	p, err := h.svc.UpdateProduct(c.Request.Context(), claims.StoreID, claims.OutletID, id, productInputFrom(&req))
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusOK, p)
}

func (h *Handler) SetProductActive(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req struct {
		Active bool `json:"active"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	if err := h.svc.SetProductActive(c.Request.Context(), claims.StoreID, id, req.Active); err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Status produk diperbarui."})
}

func (h *Handler) DeleteProduct(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	if err := h.svc.DeleteProduct(c.Request.Context(), claims.StoreID, id); err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Produk berhasil dihapus."})
}

func productInputFrom(req *productReq) ProductInput {
	in := ProductInput{
		Name:       req.Name,
		SKU:        req.SKU,
		Barcode:    req.Barcode,
		CategoryID: req.CategoryID,
		SellPrice:  req.SellPrice,
		Unit:       req.Unit,
	}
	if req.BuyPrice != nil {
		in.BuyPrice = *req.BuyPrice
	}
	if req.Stock != nil {
		in.Stock = *req.Stock
	}
	in.MinStock = req.MinStock
	return in
}

func (h *Handler) ListVariants(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	productID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	variants, err := h.svc.ListVariants(c.Request.Context(), claims.StoreID, claims.OutletID, productID)
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": variants})
}

func (h *Handler) CreateVariant(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	productID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req variantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	variant, err := h.svc.CreateVariant(c.Request.Context(), claims.StoreID, claims.OutletID, productID, claims.Name, VariantInput{
		Name: req.Name, SKU: req.SKU, BuyPrice: req.BuyPrice, SellPrice: req.SellPrice, Stock: req.Stock, MinStock: req.MinStock,
	})
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, variant)
}

func (h *Handler) UpdateVariant(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	productID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	variantID, ok := pathUint(c, "variantId")
	if !ok {
		return
	}
	var req variantReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	variant, err := h.svc.UpdateVariant(c.Request.Context(), claims.StoreID, claims.OutletID, productID, variantID, claims.Name, VariantInput{
		Name: req.Name, SKU: req.SKU, BuyPrice: req.BuyPrice, SellPrice: req.SellPrice, Stock: req.Stock, MinStock: req.MinStock,
	})
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusOK, variant)
}

func (h *Handler) DeleteVariant(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	productID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	variantID, ok := pathUint(c, "variantId")
	if !ok {
		return
	}
	if err := h.svc.DeleteVariant(c.Request.Context(), claims.StoreID, productID, variantID); err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Varian berhasil dihapus."})
}

func (h *Handler) ListPriceTiers(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	productID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var variantID *uint
	if c.Param("variantId") != "" {
		id, valid := pathUint(c, "variantId")
		if !valid {
			return
		}
		variantID = &id
	}
	tiers, err := h.svc.ListPriceTiers(c.Request.Context(), claims.StoreID, productID, variantID)
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": tiers})
}

func (h *Handler) CreatePriceTier(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	productID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var variantID *uint
	if c.Param("variantId") != "" {
		id, valid := pathUint(c, "variantId")
		if !valid {
			return
		}
		variantID = &id
	}
	var req priceTierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	tier, err := h.svc.CreatePriceTier(c.Request.Context(), claims.StoreID, productID, variantID, PriceTierInput{MinQty: req.MinQty, SellPrice: req.SellPrice})
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, tier)
}

func (h *Handler) UpdatePriceTier(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	productID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	tierID, ok := pathUint(c, "tierId")
	if !ok {
		return
	}
	var variantID *uint
	if c.Param("variantId") != "" {
		id, valid := pathUint(c, "variantId")
		if !valid {
			return
		}
		variantID = &id
	}
	var req priceTierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}
	tier, err := h.svc.UpdatePriceTier(c.Request.Context(), claims.StoreID, productID, variantID, tierID, PriceTierInput{MinQty: req.MinQty, SellPrice: req.SellPrice})
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusOK, tier)
}

func (h *Handler) DeletePriceTier(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	productID, ok := pathUint(c, "id")
	if !ok {
		return
	}
	tierID, ok := pathUint(c, "tierId")
	if !ok {
		return
	}
	var variantID *uint
	if c.Param("variantId") != "" {
		id, valid := pathUint(c, "variantId")
		if !valid {
			return
		}
		variantID = &id
	}
	if err := h.svc.DeletePriceTier(c.Request.Context(), claims.StoreID, productID, variantID, tierID); err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Harga bertingkat berhasil dihapus."})
}

func respondCatalogErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrSkuTaken), errors.Is(err, ErrCategoryTaken), errors.Is(err, ErrCategoryInvalid):
		// Sentinel yang sudah dikonversi ke apierr: status + pesan + kode
		// berasal dari sentinel itu sendiri lewat httpx.Fail.
		httpx.Fail(c, err)
	case errors.Is(err, ErrVariantRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pilih varian produk terlebih dahulu."})
	case errors.Is(err, user.ErrStoreMismatch):
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan."})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}

// pathUint parses a numeric path param, writing 400 on garbage
// (semantik disalin persis dari paket handler lama).
func pathUint(c *gin.Context, name string) (uint, bool) {
	n, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid."})
		return 0, false
	}
	return uint(n), true
}

// queryID parses an optional numeric query filter.
// Returns 0 when absent. When present but invalid, returns max uint so the
// filter matches nothing (same visible result as before with bad input).
func queryID(q func(string) string, name string) uint {
	v := q(name)
	if v == "" {
		return 0
	}
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return ^uint(0)
	}
	return uint(n)
}

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

func (h *Handler) ImportProducts(c *gin.Context) {
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
	if nonBlank > MaxImportRows {
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
