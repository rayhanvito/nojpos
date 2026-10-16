package model

type Outlet struct {
	Model
	StoreID   uint   `gorm:"not null;index" json:"store_id"`
	Name      string `gorm:"not null" json:"name"`
	Address   string `gorm:"not null;default:''" json:"address"`
	IsDefault bool   `gorm:"not null;default:false" json:"is_default"`
}

func (Outlet) TableName() string { return "outlets" }

type ProductStock struct {
	Model
	ProductID uint  `gorm:"not null;uniqueIndex:uq_product_stock_product_outlet,priority:1,where:variant_id IS NULL;uniqueIndex:uq_product_stock_variant_outlet,priority:1,where:variant_id IS NOT NULL" json:"product_id"`
	VariantID *uint `gorm:"index;uniqueIndex:uq_product_stock_variant_outlet,priority:2,where:variant_id IS NOT NULL" json:"variant_id,omitempty"`
	OutletID  uint  `gorm:"not null;index;uniqueIndex:uq_product_stock_product_outlet,priority:2,where:variant_id IS NULL;uniqueIndex:uq_product_stock_variant_outlet,priority:3,where:variant_id IS NOT NULL" json:"outlet_id"`
	Qty       int   `gorm:"not null;default:0;check:qty >= 0" json:"qty"`
}

func (ProductStock) TableName() string { return "product_stocks" }
