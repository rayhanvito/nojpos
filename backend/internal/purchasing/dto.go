package purchasing

// Tipe input purchase order (dipindah verbatim dari
// service/purchase_order.go, BE-F7).
type PurchaseOrderItemInput struct {
	ProductID uint
	VariantID *uint
	Qty       int
	BuyPrice  int64
}

type PurchaseOrderInput struct {
	SupplierID uint
	Items      []PurchaseOrderItemInput
}

type PurchaseOrderReceiptInput struct {
	ItemID uint
	Qty    int
}

// SupplierInput adalah tipe input supplier (dipindah verbatim dari
// service/supplier.go, BE-F7).
type SupplierInput struct {
	Name    string
	Contact string
	Address string
}
