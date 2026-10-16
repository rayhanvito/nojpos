package model

type Supplier struct {
	Model
	StoreID uint   `gorm:"not null;index" json:"store_id"`
	Name    string `gorm:"not null" json:"name"`
	Contact string `gorm:"not null;default:''" json:"contact"`
	Address string `gorm:"not null;default:''" json:"address"`
}

func (Supplier) TableName() string { return "suppliers" }
