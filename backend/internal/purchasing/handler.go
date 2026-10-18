package purchasing

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/httpx"
	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

// pathUint dan queryID adalah salinan privat slice dari helper paket
// handler lama (handler/params.go) — badan verbatim.
func pathUint(c *gin.Context, name string) (uint, bool) {
	n, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid."})
		return 0, false
	}
	return uint(n), true
}

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

// Handler adalah PurchaseOrderHandler lama di
// handler/purchase_order.go, dipindah verbatim.
type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

type purchaseOrderItemReq struct {
	ProductID *uint  `json:"productId"`
	VariantID *uint  `json:"variantId"`
	Qty       *int   `json:"qty"`
	BuyPrice  *int64 `json:"buyPrice"`
}

type purchaseOrderReq struct {
	SupplierID *uint                  `json:"supplierId"`
	Items      []purchaseOrderItemReq `json:"items"`
}

type purchaseOrderReceiptItemReq struct {
	ItemID *uint `json:"itemId"`
	Qty    *int  `json:"qty"`
}

type purchaseOrderReceiptReq struct {
	Items []purchaseOrderReceiptItemReq `json:"items"`
}

func (r purchaseOrderReq) input() (PurchaseOrderInput, error) {
	if r.SupplierID == nil || *r.SupplierID == 0 {
		return PurchaseOrderInput{}, fmt.Errorf("%w: supplierId wajib diisi.", ErrPurchaseOrderInvalid)
	}
	if len(r.Items) == 0 {
		return PurchaseOrderInput{}, fmt.Errorf("%w: minimal satu barang wajib diisi.", ErrPurchaseOrderInvalid)
	}
	in := PurchaseOrderInput{SupplierID: *r.SupplierID, Items: make([]PurchaseOrderItemInput, 0, len(r.Items))}
	for _, item := range r.Items {
		if item.ProductID == nil || *item.ProductID == 0 {
			return PurchaseOrderInput{}, fmt.Errorf("%w: productId wajib diisi.", ErrPurchaseOrderInvalid)
		}
		if item.Qty == nil {
			return PurchaseOrderInput{}, fmt.Errorf("%w: qty wajib diisi.", ErrPurchaseOrderInvalid)
		}
		if item.BuyPrice == nil {
			return PurchaseOrderInput{}, fmt.Errorf("%w: buyPrice wajib diisi.", ErrPurchaseOrderInvalid)
		}
		in.Items = append(in.Items, PurchaseOrderItemInput{
			ProductID: *item.ProductID, VariantID: item.VariantID, Qty: *item.Qty, BuyPrice: *item.BuyPrice,
		})
	}
	return in, nil
}

func (r purchaseOrderReceiptReq) input() ([]PurchaseOrderReceiptInput, error) {
	if len(r.Items) == 0 {
		return nil, fmt.Errorf("%w: minimal satu barang penerimaan wajib diisi.", ErrPurchaseOrderInvalid)
	}
	items := make([]PurchaseOrderReceiptInput, 0, len(r.Items))
	for _, item := range r.Items {
		if item.ItemID == nil || *item.ItemID == 0 || item.Qty == nil {
			return nil, fmt.Errorf("%w: itemId dan qty wajib diisi.", ErrPurchaseOrderInvalid)
		}
		items = append(items, PurchaseOrderReceiptInput{ItemID: *item.ItemID, Qty: *item.Qty})
	}
	return items, nil
}

func (h *Handler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	items, err := h.svc.List(c.Request.Context(), claims.StoreID, claims.OutletID, PurchaseOrderFilter{
		Status: c.Query("status"), SupplierID: queryID(c.Query, "supplierId"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat purchase order."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *Handler) Get(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	po, err := h.svc.Get(c.Request.Context(), claims.StoreID, claims.OutletID, id)
	if err != nil {
		respondPurchaseOrderErr(c, err)
		return
	}
	c.JSON(http.StatusOK, po)
}

func (h *Handler) Create(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req purchaseOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body purchase order tidak valid."})
		return
	}
	in, err := req.input()
	if err != nil {
		respondPurchaseOrderErr(c, err)
		return
	}
	po, err := h.svc.Create(c.Request.Context(), claims.StoreID, claims.OutletID, in)
	if err != nil {
		respondPurchaseOrderErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, po)
}

func (h *Handler) Update(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req purchaseOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body purchase order tidak valid."})
		return
	}
	in, err := req.input()
	if err != nil {
		respondPurchaseOrderErr(c, err)
		return
	}
	po, err := h.svc.Update(c.Request.Context(), claims.StoreID, claims.OutletID, id, in)
	if err != nil {
		respondPurchaseOrderErr(c, err)
		return
	}
	c.JSON(http.StatusOK, po)
}

func (h *Handler) Delete(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), claims.StoreID, claims.OutletID, id); err != nil {
		respondPurchaseOrderErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Purchase order draft berhasil dihapus."})
}

func (h *Handler) Order(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	po, err := h.svc.Order(c.Request.Context(), claims.StoreID, claims.OutletID, id)
	if err != nil {
		respondPurchaseOrderErr(c, err)
		return
	}
	c.JSON(http.StatusOK, po)
}

func (h *Handler) Receive(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req purchaseOrderReceiptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body penerimaan tidak valid."})
		return
	}
	items, err := req.input()
	if err != nil {
		respondPurchaseOrderErr(c, err)
		return
	}
	po, err := h.svc.Receive(c.Request.Context(), claims.StoreID, claims.OutletID, id, items, claims.Name)
	if err != nil {
		respondPurchaseOrderErr(c, err)
		return
	}
	c.JSON(http.StatusOK, po)
}

func (h *Handler) Cancel(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	po, err := h.svc.Cancel(c.Request.Context(), claims.StoreID, claims.OutletID, id)
	if err != nil {
		respondPurchaseOrderErr(c, err)
		return
	}
	c.JSON(http.StatusOK, po)
}

func respondPurchaseOrderErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrPurchaseOrderInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, ErrPurchaseOrderLocked):
		c.JSON(http.StatusConflict, gin.H{"error": "Barang pada PO yang sudah dipesan tidak dapat diubah."})
	case errors.Is(err, ErrPurchaseOrderState):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status PO tidak mengizinkan aksi ini."})
	case errors.Is(err, ErrPurchaseOrderNumberLimit):
		c.JSON(http.StatusConflict, gin.H{"error": "Nomor PO untuk toko ini hari ini telah mencapai batas."})
	case errors.Is(err, ErrDuplicate):
		c.JSON(http.StatusConflict, gin.H{"error": "Nomor purchase order sudah digunakan; silakan coba lagi."})
	case errors.Is(err, ErrPurchaseOrderNotFound), errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase order tidak ditemukan di toko Anda."})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses purchase order."})
	}
}

// SupplierHandler adalah SupplierHandler lama di handler/supplier.go,
// dipindah verbatim.
type SupplierHandler struct {
	svc *SupplierService
}

func NewSupplierHandler(svc *SupplierService) *SupplierHandler {
	return &SupplierHandler{svc: svc}
}

type supplierReq struct {
	Name    string `json:"name"`
	Contact string `json:"contact"`
	Address string `json:"address"`
}

func (r supplierReq) input() SupplierInput {
	return SupplierInput{Name: r.Name, Contact: r.Contact, Address: r.Address}
}

func (h *SupplierHandler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	items, err := h.svc.List(c.Request.Context(), claims.StoreID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat supplier."})
		return
	}
	c.JSON(http.StatusOK, httpx.PaginateList(c, items))
}

func (h *SupplierHandler) Get(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	supplier, err := h.svc.Get(c.Request.Context(), claims.StoreID, id)
	if err != nil {
		respondSupplierErr(c, err)
		return
	}
	c.JSON(http.StatusOK, supplier)
}

func (h *SupplierHandler) Create(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req supplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body supplier tidak valid."})
		return
	}
	supplier, err := h.svc.Create(c.Request.Context(), claims.StoreID, req.input())
	if err != nil {
		respondSupplierErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, supplier)
}

func (h *SupplierHandler) Update(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req supplierReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body supplier tidak valid."})
		return
	}
	supplier, err := h.svc.Update(c.Request.Context(), claims.StoreID, id, req.input())
	if err != nil {
		respondSupplierErr(c, err)
		return
	}
	c.JSON(http.StatusOK, supplier)
}

func (h *SupplierHandler) Delete(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	if err := h.svc.Delete(c.Request.Context(), claims.StoreID, id); err != nil {
		respondSupplierErr(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Supplier berhasil dihapus."})
}

func respondSupplierErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrSupplierNameRequired):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Nama supplier wajib diisi."})
	case errors.Is(err, ErrSupplierNotFound), errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Supplier tidak ditemukan di toko Anda."})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses supplier."})
	}
}
