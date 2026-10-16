package model

type MovementType string

const (
	MovementSale     MovementType = "sale"
	MovementRefund   MovementType = "refund"
	MovementAdjust   MovementType = "adjust"
	MovementInitial  MovementType = "initial"
	MovementPurchase MovementType = "purchase"
	MovementOpname   MovementType = "opname"
)

type Movement struct {
	Model
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
