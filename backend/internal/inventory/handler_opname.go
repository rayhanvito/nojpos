package inventory

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

type OpnameHandler struct {
	svc *OpnameService
}

func NewOpnameHandler(svc *OpnameService) *OpnameHandler {
	return &OpnameHandler{svc: svc}
}

type stockOpnameCreateReq struct {
	Note string `json:"note"`
}

type stockOpnameCountItemReq struct {
	ItemID       *uint `json:"itemId"`
	CountedStock *int  `json:"countedStock"`
}

type stockOpnameCountReq struct {
	Items []stockOpnameCountItemReq `json:"items"`
}

func (h *OpnameHandler) List(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	items, err := h.svc.List(c.Request.Context(), claims.StoreID, claims.OutletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat sesi stock opname."})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *OpnameHandler) Get(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	session, err := h.svc.Get(c.Request.Context(), claims.StoreID, claims.OutletID, id)
	if err != nil {
		respondStockOpnameErr(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *OpnameHandler) Create(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req stockOpnameCreateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body sesi stock opname tidak valid."})
		return
	}
	session, err := h.svc.Create(c.Request.Context(), claims.StoreID, claims.OutletID, req.Note, claims.Name)
	if err != nil {
		respondStockOpnameErr(c, err)
		return
	}
	c.JSON(http.StatusCreated, session)
}

func (h *OpnameHandler) UpdateItems(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	var req stockOpnameCountReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Body item stock opname tidak valid."})
		return
	}
	items := make([]StockOpnameItemCount, 0, len(req.Items))
	for _, item := range req.Items {
		if item.ItemID == nil || item.CountedStock == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "itemId dan countedStock wajib diisi."})
			return
		}
		items = append(items, StockOpnameItemCount{ItemID: *item.ItemID, CountedStock: *item.CountedStock})
	}
	session, err := h.svc.UpdateItems(c.Request.Context(), claims.StoreID, claims.OutletID, id, items)
	if err != nil {
		respondStockOpnameErr(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *OpnameHandler) Post(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	session, err := h.svc.Post(c.Request.Context(), claims.StoreID, claims.OutletID, id, claims.Name)
	if err != nil {
		respondStockOpnameErr(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func (h *OpnameHandler) Cancel(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	id, ok := pathUint(c, "id")
	if !ok {
		return
	}
	session, err := h.svc.Cancel(c.Request.Context(), claims.StoreID, claims.OutletID, id)
	if err != nil {
		respondStockOpnameErr(c, err)
		return
	}
	c.JSON(http.StatusOK, session)
}

func respondStockOpnameErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrStockOpnameInvalid):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, ErrStockOpnameActive):
		c.JSON(http.StatusConflict, gin.H{"error": "Masih ada sesi stock opname yang berstatus counting di toko ini."})
	case errors.Is(err, ErrStockOpnameState):
		c.JSON(http.StatusConflict, gin.H{"error": "Sesi stock opname sudah dikunci dan tidak dapat diubah."})
	case errors.Is(err, ErrStockOpnameNotFound), errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "Sesi stock opname tidak ditemukan di toko Anda."})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memproses sesi stock opname."})
	}
}
