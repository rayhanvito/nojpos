package handler

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/middleware"
	"github.com/rayhanvito/nojpos/backend/repo"
	"github.com/rayhanvito/nojpos/backend/service"
)

type PurchaseOrderHandler struct {
	svc *service.PurchaseOrderService
}

func NewPurchaseOrderHandler(svc *service.PurchaseOrderService) *PurchaseOrderHandler {
	return &PurchaseOrderHandler{svc: svc}
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

func (r purchaseOrderReq) input() (service.PurchaseOrderInput, error) {
	if r.SupplierID == nil || *r.SupplierID == 0 {
		return service.PurchaseOrderInput{}, fmt.Errorf("%w: supplierId wajib diisi.", service.ErrPurchaseOrderInvalid)
	}
	if len(r.Items) == 0 {
		return service.PurchaseOrderInput{}, fmt.Errorf("%w: minimal satu barang wajib diisi.", service.ErrPurchaseOrderInvalid)
	}
	in := service.PurchaseOrderInput{SupplierID: *r.SupplierID, Items: make([]service.PurchaseOrderItemInput, 0, len(r.Items))}
	for _, item := range r.Items {
		if item.ProductID == nil || *item.ProductID == 0 {
			return service.PurchaseOrderInput{}, fmt.Errorf("%w: productId wajib diisi.", service.ErrPurchaseOrderInvalid)
		}
		if item.Qty == nil {
			return service.PurchaseOrderInput{}, fmt.Errorf("%w: qty wajib diisi.", service.ErrPurchaseOrderInvalid)
		}
		if item.BuyPrice == nil {
			return service.PurchaseOrderInput{}, fmt.Errorf("%w: buyPrice wajib diisi.", service.ErrPurchaseOrderInvalid)
		}
		in.Items = append(in.Items, service.PurchaseOrderItemInput{
			ProductID: *item.ProductID, VariantID: item.VariantID, Qty: *item.Qty, BuyPrice: *item.BuyPrice,
		})
	}
	return in, nil
}

func (r purchaseOrderReceiptReq) input() ([]service.PurchaseOrderReceiptInput, error) {
	if len(r.Items) == 0 {
		return nil, fmt.Errorf("%w: minimal satu barang penerimaan wajib diisi.", service.ErrPurchaseOrderInvalid)
	}
	items := make([]service.PurchaseOrderReceiptInput, 0, len(r.Items))
	for _, item := range r.Items {
		if item.ItemID == nil || *item.ItemID == 0 || item.Qty == nil {
			return nil, fmt.Errorf("%w: itemId dan qty wajib diisi.", service.ErrPurchaseOrderInvalid)
		}
		items = append(items, service.PurchaseOrderReceiptInput{ItemID: *item.ItemID, Qty: *item.Qty})
	}
	return items, nil
}

func (h *PurchaseOrderHandler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	items, err := h.svc.List(c.Request.Context(), claims.StoreID, claims.OutletID, repo.PurchaseOrderFilter{
		Status: c.Query("status"), SupplierID: queryID(c.Query, "supplierId"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat purchase order."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *PurchaseOrderHandler) Get(c *gin.Context) {
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

func (h *PurchaseOrderHandler) Create(c *gin.Context) {
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

func (h *PurchaseOrderHandler) Update(c *gin.Context) {
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

func (h *PurchaseOrderHandler) Delete(c *gin.Context) {
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

func (h *PurchaseOrderHandler) Order(c *gin.Context) {
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

func (h *PurchaseOrderHandler) Receive(c *gin.Context) {
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

func (h *PurchaseOrderHandler) Cancel(c *gin.Context) {
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
	case errors.Is(err, service.ErrPurchaseOrderInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrPurchaseOrderLocked):
		c.JSON(http.StatusConflict, gin.H{"error": "Barang pada PO yang sudah dipesan tidak dapat diubah."})
	case errors.Is(err, repo.ErrPurchaseOrderState):
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status PO tidak mengizinkan aksi ini."})
	case errors.Is(err, repo.ErrPurchaseOrderNumberLimit):
		c.JSON(http.StatusConflict, gin.H{"error": "Nomor PO untuk toko ini hari ini telah mencapai batas."})
	case errors.Is(err, repo.ErrDuplicate):
		c.JSON(http.StatusConflict, gin.H{"error": "Nomor purchase order sudah digunakan; silakan coba lagi."})
	case errors.Is(err, service.ErrPurchaseOrderNotFound), errors.Is(err, repo.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Purchase order tidak ditemukan di toko Anda."})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses purchase order."})
	}
}
