package model

type StockOpnameStatus string

const (
	StockOpnameCounting  StockOpnameStatus = "counting"
	StockOpnamePosted    StockOpnameStatus = "posted"
	StockOpnameCancelled StockOpnameStatus = "cancelled"
)

type StockOpnameSession struct {
	Model
	StoreID   uint              `gorm:"not null;index;uniqueIndex:uq_stock_opname_counting_store,where:status = 'counting' AND deleted_at IS NULL" json:"store_id"`
	OutletID  *uint             `gorm:"index" json:"outlet_id,omitempty"`
	Status    StockOpnameStatus `gorm:"not null;default:'counting';index" json:"status"`
	Note      string            `gorm:"not null;default:''" json:"note"`
	CreatedBy string            `gorm:"not null;default:''" json:"created_by"`
	Items     []StockOpnameItem `gorm:"-" json:"items,omitempty"`
}

func (StockOpnameSession) TableName() string { return "stock_opname_sessions" }

type StockOpnameItem struct {
	Model
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
