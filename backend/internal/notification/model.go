package notification

import (
	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// NotificationCategory adalah kategori notifikasi (dipindah verbatim
// dari model/notification.go, BE-F9a).
type NotificationCategory string

const (
	CategoryStok      NotificationCategory = "stok"
	CategoryTransaksi NotificationCategory = "transaksi"
	CategorySistem    NotificationCategory = "sistem"
)

// Notification adalah model notifikasi (dipindah verbatim dari
// model/notification.go, BE-F9a).
type Notification struct {
	database.BaseModel
	StoreID       uint                 `gorm:"not null;index;uniqueIndex:uq_notification_daily_key,priority:1" json:"store_id"`
	DedupKey      *string              `gorm:"uniqueIndex:uq_notification_daily_key,priority:2" json:"-"`
	Title         string               `gorm:"not null" json:"title"`
	Message       string               `gorm:"not null" json:"message"`
	Category      NotificationCategory `gorm:"not null;default:'sistem'" json:"category"`
	Type          string               `gorm:"not null;default:'info'" json:"type"`
	ActorID       string               `gorm:"not null;default:''" json:"actor_id,omitempty"`
	ActorName     string               `gorm:"not null;default:''" json:"actor_name,omitempty"`
	ReferenceType string               `gorm:"not null;default:''" json:"reference_type,omitempty"`
	ReferenceID   string               `gorm:"not null;default:''" json:"reference_id,omitempty"`
	Read          bool                 `gorm:"not null;default:false" json:"read"`
}

func (Notification) TableName() string {
	return "notifications"
}
