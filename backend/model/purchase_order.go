package model

type PurchaseOrderStatus string

const (
	PurchaseOrderDraft             PurchaseOrderStatus = "draft"
	PurchaseOrderOrdered           PurchaseOrderStatus = "ordered"
	PurchaseOrderPartiallyReceived PurchaseOrderStatus = "partially_received"
	PurchaseOrderFullyReceived     PurchaseOrderStatus = "fully_received"
	PurchaseOrderCancelled         PurchaseOrderStatus = "cancelled"
)

type PurchaseOrder struct {
	Model
	StoreID    uint                `gorm:"not null;index;uniqueIndex:uq_purchase_order_store_number,priority:1" json:"store_id"`
	OutletID   *uint               `gorm:"index" json:"outlet_id,omitempty"`
	SupplierID uint                `gorm:"not null;index" json:"supplier_id"`
	Number     string              `gorm:"not null;uniqueIndex:uq_purchase_order_store_number,priority:2" json:"number"`
	Status     PurchaseOrderStatus `gorm:"not null;default:'draft';index" json:"status"`
	Supplier   *Supplier           `gorm:"-" json:"supplier,omitempty"`
	Items      []PurchaseOrderItem `gorm:"-" json:"items,omitempty"`
}

func (PurchaseOrder) TableName() string { return "purchase_orders" }

type PurchaseOrderItem struct {
	ID              uint  `gorm:"primarykey" json:"id"`
	PurchaseOrderID uint  `gorm:"not null;index" json:"purchase_order_id"`
	ProductID       uint  `gorm:"not null;index" json:"product_id"`
	VariantID       *uint `gorm:"index" json:"variant_id,omitempty"`
	QtyOrdered      int   `gorm:"not null;check:qty_ordered > 0" json:"qty_ordered"`
	QtyReceived     int   `gorm:"not null;default:0;check:qty_received >= 0" json:"qty_received"`
	BuyPrice        int64 `gorm:"not null;check:buy_price >= 0" json:"buy_price"`
}

func (PurchaseOrderItem) TableName() string { return "purchase_order_items" }
