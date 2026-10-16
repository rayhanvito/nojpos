package model

type Category struct {
	Model
	StoreID uint   `gorm:"not null;index;uniqueIndex:uq_category_store_name" json:"store_id"`
	Name    string `gorm:"not null;uniqueIndex:uq_category_store_name" json:"name"`
	Active  bool   `gorm:"not null;default:true" json:"active"`
}

func (Category) TableName() string {
	return "categories"
}

type Product struct {
	Model
	StoreID      uint             `gorm:"not null;index;uniqueIndex:uq_product_store_sku" json:"-"`
	OutletID     *uint            `gorm:"index" json:"outlet_id,omitempty"`
	CategoryID   *uint            `gorm:"index" json:"category_id"`
	CategoryName *string          `gorm:"-" json:"category_name"`
	Name         string           `gorm:"not null" json:"name"`
	SKU          string           `gorm:"not null;uniqueIndex:uq_product_store_sku" json:"sku"`
	Barcode      string           `gorm:"not null;default:''" json:"barcode"`
	BuyPrice     int64            `gorm:"not null;default:0;check:buy_price >= 0" json:"buy_price"`
	SellPrice    int64            `gorm:"not null;default:0;check:sell_price >= 0" json:"sell_price"`
	Stock        int              `gorm:"not null;default:0;check:stock >= 0" json:"stock"`
	MinStock     int              `gorm:"not null;default:0;check:min_stock >= 0" json:"min_stock"`
	Unit         string           `gorm:"not null;default:'pcs'" json:"unit"`
	Active       bool             `gorm:"not null;default:true" json:"active"`
	Variants     []ProductVariant `gorm:"-" json:"variants"`
	PriceTiers   []PriceTier      `gorm:"-" json:"price_tiers"`
}

func (Product) TableName() string {
	return "products"
}
