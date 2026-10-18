package webhook

import (
	"time"

	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Nama event yang sah untuk langganan webhook (daftar final Unit F).
const (
	EventTransactionCreated  = "transaction.created"
	EventTransactionRefunded = "transaction.refunded"
	EventStockUpdated        = "stock.updated"
)

// ValidEvent melaporkan apakah name adalah nama event webhook yang sah.
func ValidEvent(name string) bool {
	switch name {
	case EventTransactionCreated, EventTransactionRefunded, EventStockUpdated:
		return true
	}
	return false
}

// Webhook adalah satu langganan sistem luar milik sebuah toko.
//
// Secret disimpan apa adanya (BUKAN hash) karena server memerlukannya
// untuk menandatangani setiap pengiriman dengan HMAC-SHA256; ia bukan
// kredensial akses ke NOJPOS — kebocorannya hanya memungkinkan pemalsuan
// kiriman KE penerima. Karena itu kolom ini json:"-" dan tidak pernah
// ditulis ke log; satu-satunya tempat secret tampil adalah respons
// pembuatan webhook (201).
type Webhook struct {
	database.BaseModel
	StoreID uint   `gorm:"not null;index" json:"-"`
	URL     string `gorm:"not null" json:"-"`
	Events  string `gorm:"not null" json:"-"` // array JSON nama event
	Secret  string `gorm:"not null" json:"-"`
}

func (Webhook) TableName() string { return "webhooks" }

// Delivery adalah catatan satu percobaan pengiriman webhook. Payload
// adalah body persis seperti terkirim (yang ditandatangani), disimpan
// apa adanya untuk audit dan verifikasi tanda tangan.
type Delivery struct {
	ID             uint      `gorm:"primarykey" json:"id"`
	CreatedAt      time.Time `json:"created_at"`
	WebhookID      uint      `gorm:"not null;index" json:"webhook_id"`
	StoreID        uint      `gorm:"not null;index" json:"store_id"`
	Event          string    `gorm:"not null" json:"event"`
	Payload        string    `gorm:"not null" json:"payload"`
	ResponseStatus *int      `json:"response_status,omitempty"`
	Success        bool      `gorm:"not null;default:false" json:"success"`
	Error          string    `gorm:"not null;default:''" json:"error"`
}

func (Delivery) TableName() string { return "webhook_deliveries" }

// Metadata adalah bentuk baca webhook: tanpa secret, selamanya.
type Metadata struct {
	ID        uint      `json:"id"`
	URL       string    `json:"url"`
	Events    []string  `json:"events"`
	CreatedAt time.Time `json:"created_at"`
}
