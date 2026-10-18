package inventory

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/platform/httpx"
	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
)

type transferItemReq struct {
	ProductID uint  `json:"productId"`
	VariantID *uint `json:"variantId"`
	Qty       int   `json:"qty"`
}

type createTransferReq struct {
	FromOutletID uint              `json:"fromOutletId"`
	ToOutletID   uint              `json:"toOutletId"`
	Note         string            `json:"note"`
	Items        []transferItemReq `json:"items"`
}

// CreateTransfer melayani POST /stock/transfers (admin): memindahkan
// stok antar outlet secara atomik dan menjawab 201 + transfer-nya.
func (h *Handler) CreateTransfer(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	var req createTransferReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "body JSON tidak valid"})
		return
	}

	cmd := CreateTransferCmd{
		FromOutletID: req.FromOutletID,
		ToOutletID:   req.ToOutletID,
		Note:         req.Note,
	}
	for _, item := range req.Items {
		cmd.Items = append(cmd.Items, TransferItemCmd{
			ProductID: item.ProductID,
			VariantID: item.VariantID,
			Qty:       item.Qty,
		})
	}

	transfer, err := h.svc.CreateTransfer(c.Request.Context(), claims.StoreID, claims.UserID, claims.Name, cmd)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Outlet atau produk tidak ditemukan."})
		case errors.Is(err, catalog.ErrVariantRequired):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			httpx.Fail(c, err)
		}
		return
	}
	c.JSON(http.StatusCreated, transfer)
}

// ListTransfers melayani GET /stock/transfers (admin): daftar transfer
// toko dalam amplop paginasi kanonis {items,total,page,limit}.
func (h *Handler) ListTransfers(c *gin.Context) {
	claims := middleware.ClaimsFrom(c)
	q := c.Query
	f := TransferListFilter{
		FromOutletID: queryID(q, "fromOutletId"),
		ToOutletID:   queryID(q, "toOutletId"),
		From:         q("from"),
		To:           q("to"),
	}

	items, err := h.svc.ListTransfers(c.Request.Context(), claims.StoreID, f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal memuat daftar transfer stok."})
		return
	}
	c.JSON(http.StatusOK, httpx.PaginateList(c, items))
}
