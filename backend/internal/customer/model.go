package customer

import (
	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Customer adalah model pelanggan (dipindah verbatim dari
// model/customer.go, BE-F4).
type Customer struct {
	database.BaseModel
	StoreID uint   `gorm:"not null;uniqueIndex:uq_customers_store_phone,priority:1,where:phone <> '' AND deleted_at IS NULL" json:"store_id"`
	Name    string `gorm:"not null" json:"name"`
	Phone   string `gorm:"not null;default:'';uniqueIndex:uq_customers_store_phone,priority:2,where:phone <> '' AND deleted_at IS NULL" json:"phone"`
	Email   string `gorm:"not null;default:''" json:"email"`
	Address string `gorm:"not null;default:''" json:"address"`
	Points  int64  `gorm:"not null;default:0;check:points >= 0" json:"points"`
}

func (Customer) TableName() string { return "customers" }
