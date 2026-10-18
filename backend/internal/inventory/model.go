package inventory

import (
	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Tipe-tipe di file ini dipindahkan verbatim dari model/stock.go,
// model/stock_opname.go, dan model/outlet.go (ProductStock yang satu-
// satunya tipe tersisa di file itu) pada BE-F3-1.

type MovementType string

const (
	MovementSale        MovementType = "sale"
	MovementRefund      MovementType = "refund"
	MovementAdjust      MovementType = "adjust"
	MovementInitial     MovementType = "initial"
	MovementPurchase    MovementType = "purchase"
	MovementOpname      MovementType = "opname"
	MovementTransferOut MovementType = "transfer_out"
	MovementTransferIn  MovementType = "transfer_in"
)

type Movement struct {
	database.BaseModel
	StoreID     uint         `gorm:"not null;index" json:"-"`
	OutletID    *uint        `gorm:"index" json:"outlet_id,omitempty"`
	ProductID   uint         `gorm:"not null;index" json:"product_id"`
	VariantID   *uint        `gorm:"index" json:"variant_id,omitempty"`
	ProductName *string      `gorm:"-" json:"product_name"`
	Type        MovementType `gorm:"not null" json:"type"`
	Qty         int          `gorm:"not null" json:"qty"`
	Reason      string       `gorm:"not null;default:''" json:"reason"`
	Actor       string       `gorm:"not null;default:''" json:"actor"`
}

func (Movement) TableName() string {
	return "stock_movements"
}

type StockOpnameStatus string

const (
	StockOpnameCounting  StockOpnameStatus = "counting"
	StockOpnamePosted    StockOpnameStatus = "posted"
	StockOpnameCancelled StockOpnameStatus = "cancelled"
)

type StockOpnameSession struct {
	database.BaseModel
	StoreID   uint              `gorm:"not null;index;uniqueIndex:uq_stock_opname_counting_store,where:status = 'counting' AND deleted_at IS NULL" json:"store_id"`
	OutletID  *uint             `gorm:"index" json:"outlet_id,omitempty"`
	Status    StockOpnameStatus `gorm:"not null;default:'counting';index" json:"status"`
	Note      string            `gorm:"not null;default:''" json:"note"`
	CreatedBy string            `gorm:"not null;default:''" json:"created_by"`
	Items     []StockOpnameItem `gorm:"-" json:"items,omitempty"`
}

func (StockOpnameSession) TableName() string { return "stock_opname_sessions" }

type StockOpnameItem struct {
	database.BaseModel
	SessionID    uint   `gorm:"not null;index" json:"session_id"`
	ProductID    uint   `gorm:"not null;index" json:"product_id"`
	VariantID    *uint  `gorm:"index" json:"variant_id,omitempty"`
	SystemStock  int    `gorm:"not null;check:system_stock >= 0" json:"system_stock"`
	CountedStock *int   `gorm:"check:counted_stock IS NULL OR counted_stock >= 0" json:"counted_stock"`
	Difference   *int   `gorm:"-" json:"difference"`
	ProductName  string `gorm:"-" json:"product_name,omitempty"`
	VariantName  string `gorm:"-" json:"variant_name,omitempty"`
}

func (StockOpnameItem) TableName() string { return "stock_opname_items" }

type ProductStock struct {
	database.BaseModel
	ProductID uint  `gorm:"not null;uniqueIndex:uq_product_stock_product_outlet,priority:1,where:variant_id IS NULL;uniqueIndex:uq_product_stock_variant_outlet,priority:1,where:variant_id IS NOT NULL" json:"product_id"`
	VariantID *uint `gorm:"index;uniqueIndex:uq_product_stock_variant_outlet,priority:2,where:variant_id IS NOT NULL" json:"variant_id,omitempty"`
	OutletID  uint  `gorm:"not null;index;uniqueIndex:uq_product_stock_product_outlet,priority:2,where:variant_id IS NULL;uniqueIndex:uq_product_stock_variant_outlet,priority:3,where:variant_id IS NOT NULL" json:"outlet_id"`
	Qty       int   `gorm:"not null;default:0;check:qty >= 0" json:"qty"`
}

func (ProductStock) TableName() string { return "product_stocks" }
