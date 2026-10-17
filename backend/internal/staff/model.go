package staff

import (
	"time"

	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Cashier adalah entitas kasir (dipindahkan verbatim dari
// model/models.go).
type Cashier struct {
	database.BaseModel
	StoreID      uint       `gorm:"not null;index" json:"store_id"`
	OutletID     *uint      `gorm:"index" json:"outlet_id,omitempty"`
	Name         string     `gorm:"not null" json:"name"`
	PasscodeHash *string    `json:"-"`
	Active       bool       `gorm:"not null;default:true" json:"active"`
	LastSeenAt   *time.Time `gorm:"index" json:"last_seen_at,omitempty"`
}

func (Cashier) TableName() string {
	return "cashiers"
}
