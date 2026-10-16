package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/adrr-dev/openPOS/backend/middleware"
	"github.com/adrr-dev/openPOS/backend/repo"
	"github.com/adrr-dev/openPOS/backend/service"
)

type CatalogHandler struct {
	svc *service.CatalogService
}

func NewCatalogHandler(svc *service.CatalogService) *CatalogHandler { return &CatalogHandler{svc: svc} }

type categoryReq struct {
	Name string `json:"name"`
}

func (h *CatalogHandler) ListCategories(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	cats, err := h.svc.ListCategories(c.Request.Context(), claims.StoreID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat kategori."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"categories": cats})
}

func (h *CatalogHandler) CreateCategory(c *gin.Context) {
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

func (h *CatalogHandler) DeleteCategory(c *gin.Context) {
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

func (h *CatalogHandler) ListProducts(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	q := c.Query

	f := repo.ProductFilter{Q: q("q"), CategoryID: queryID(q, "categoryId")}
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

func (h *CatalogHandler) GetProduct(c *gin.Context) {
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

func (h *CatalogHandler) CreateProduct(c *gin.Context) {
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

func (h *CatalogHandler) UpdateProduct(c *gin.Context) {
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

func (h *CatalogHandler) SetProductActive(c *gin.Context) {
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

func (h *CatalogHandler) DeleteProduct(c *gin.Context) {
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

func productInputFrom(req *productReq) service.ProductInput {
	in := service.ProductInput{
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

func (h *CatalogHandler) ListVariants(c *gin.Context) {
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

func (h *CatalogHandler) CreateVariant(c *gin.Context) {
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
	variant, err := h.svc.CreateVariant(c.Request.Context(), claims.StoreID, claims.OutletID, productID, claims.Name, service.VariantInput{
		Name: req.Name, SKU: req.SKU, BuyPrice: req.BuyPrice, SellPrice: req.SellPrice, Stock: req.Stock, MinStock: req.MinStock,
	})
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, variant)
}

func (h *CatalogHandler) UpdateVariant(c *gin.Context) {
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
	variant, err := h.svc.UpdateVariant(c.Request.Context(), claims.StoreID, claims.OutletID, productID, variantID, claims.Name, service.VariantInput{
		Name: req.Name, SKU: req.SKU, BuyPrice: req.BuyPrice, SellPrice: req.SellPrice, Stock: req.Stock, MinStock: req.MinStock,
	})
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusOK, variant)
}

func (h *CatalogHandler) DeleteVariant(c *gin.Context) {
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

func (h *CatalogHandler) ListPriceTiers(c *gin.Context) {
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

func (h *CatalogHandler) CreatePriceTier(c *gin.Context) {
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
	tier, err := h.svc.CreatePriceTier(c.Request.Context(), claims.StoreID, productID, variantID, service.PriceTierInput{MinQty: req.MinQty, SellPrice: req.SellPrice})
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, tier)
}

func (h *CatalogHandler) UpdatePriceTier(c *gin.Context) {
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
	tier, err := h.svc.UpdatePriceTier(c.Request.Context(), claims.StoreID, productID, variantID, tierID, service.PriceTierInput{MinQty: req.MinQty, SellPrice: req.SellPrice})
	if err != nil {
		respondCatalogErr(c, err)
		return
	}
	c.JSON(http.StatusOK, tier)
}

func (h *CatalogHandler) DeletePriceTier(c *gin.Context) {
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
	case errors.Is(err, service.ErrSkuTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "SKU sudah digunakan di toko ini."})
	case errors.Is(err, service.ErrCategoryTaken):
		c.JSON(http.StatusConflict, gin.H{"error": "Kategori dengan nama itu sudah ada."})
	case errors.Is(err, service.ErrCategoryInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kategori tidak ditemukan di toko Anda."})
	case errors.Is(err, repo.ErrVariantRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Pilih varian produk terlebih dahulu."})
	case errors.Is(err, service.ErrStoreMismatch):
		c.JSON(http.StatusNotFound, gin.H{"error": "Produk tidak ditemukan."})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}
}
