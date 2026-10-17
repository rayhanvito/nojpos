package purchasing

import (
	"errors"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
)

var (
	// ErrNotFound dan ErrDuplicate adalah alias ke instans kanonis
	// slice outlet (BE-F1a) — instans yang sama dengan
	// repo.ErrNotFound / repo.ErrDuplicate lama, jadi pemeriksaan
	// errors.Is konsumen menetap tetap cocok.
	ErrNotFound  = outlet.ErrNotFound
	ErrDuplicate = outlet.ErrDuplicate

	// Sentinel repo purchase order, dipindahkan verbatim dari
	// repo/purchase_order.go (BE-F7) — definisi kanonisnya kini di
	// slice ini.
	ErrPurchaseOrderItemsLocked   = errors.New("barang pada PO sudah terkunci")
	ErrPurchaseOrderState         = errors.New("status PO tidak mengizinkan aksi ini")
	ErrPurchaseOrderNumberLimit   = errors.New("nomor PO harian sudah mencapai batas")
	ErrPurchaseOrderReceiptItem   = errors.New("barang penerimaan tidak sesuai dengan PO")
	ErrPurchaseOrderReceiptOver   = errors.New("jumlah penerimaan melebihi sisa pesanan")
	ErrPurchaseOrderReceiptTarget = errors.New("target stok PO sudah tidak tersedia")

	// Sentinel service purchase order, dipindahkan verbatim dari
	// service/purchase_order.go (BE-F7).
	ErrPurchaseOrderNotFound = errors.New("purchase order tidak ditemukan")
	ErrPurchaseOrderInvalid  = errors.New("data purchase order tidak valid")
	ErrPurchaseOrderLocked   = errors.New("barang pada PO yang sudah dipesan tidak dapat diubah")

	// Sentinel service supplier, dipindahkan verbatim dari
	// service/supplier.go (BE-F7).
	ErrSupplierNotFound     = errors.New("supplier tidak ditemukan")
	ErrSupplierNameRequired = errors.New("nama supplier wajib diisi")
)
