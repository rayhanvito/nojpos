package model

type NotificationCategory string

const (
	CategoryStok      NotificationCategory = "stok"
	CategoryTransaksi NotificationCategory = "transaksi"
	CategorySistem    NotificationCategory = "sistem"
)

type Notification struct {
	Model
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
