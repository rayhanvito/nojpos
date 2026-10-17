package customer

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Tipe baris privat slice untuk tabel transactions,
// transaction_items, dan refunds. Tipe kanonisnya kini milik slice
// internal/sales (BE-F8), tetapi sales mengimpor customer
// (pelanggan + poin di repo checkout sales) dan finance mengimpor
// customer lalu diimpor sales — customer tidak boleh mengimpor
// balik sales (siklus), jadi riwayat transaksi pelanggan memakai
// baris lokal ini:
// skema (nama tabel, kolom, tag) identik verbatim dengan tipe
// kanonis, hanya dipakai di dalam slice customer. Konstanta status dan
// metode bayar yang dipakai slice ini ("Cash" dkk.) ditulis sebagai
// literal setempat dengan nilai yang sama.

type trxStatus string

const (
	trxPending                  trxStatus = "pending"
	trxCompleted                trxStatus = "completed"
	trxCompletedOfflineShortage trxStatus = "completed_offline_shortage"
	trxCancelled                trxStatus = "cancelled"
	trxRefunded                 trxStatus = "refunded"
)

type finalStock struct {
	ProductID uint  `json:"product_id"`
	VariantID *uint `json:"variant_id"`
	Stock     int   `json:"stock"`
}

type trxItem struct {
	ProductID   uint   `json:"product_id"`
	VariantID   *uint  `json:"variant_id,omitempty"`
	VariantName string `json:"variant_name,omitempty"`
	Name        string `json:"name"`
	BuyPrice    int64  `json:"buy_price"`
	Price       int64  `json:"price"`
	Qty         int    `json:"qty"`
}

type trxItemList []trxItem

func (l *trxItemList) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, okStr := value.(string)
		if !okStr {
			return errors.New("type assertion to []byte or string failed")
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, l)
}

func (l trxItemList) Value() (driver.Value, error) {
	return json.Marshal(l)
}

type refundItemList map[uint]int

func (m *refundItemList) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		str, okStr := value.(string)
		if !okStr {
			return errors.New("type assertion to []byte or string failed")
		}
		bytes = []byte(str)
	}
	return json.Unmarshal(bytes, m)
}

func (m refundItemList) Value() (driver.Value, error) {
	return json.Marshal(m)
}

type trx struct {
	database.BaseModel
	StoreID         uint           `gorm:"not null;index;uniqueIndex:uq_transactions_store_client_tx_id,priority:1,where:client_tx_id <> ''" json:"-"`
	OutletID        *uint          `gorm:"index" json:"outlet_id,omitempty"`
	ClientTxID      string         `gorm:"not null;default:'';uniqueIndex:uq_transactions_store_client_tx_id,priority:2,where:client_tx_id <> ''" json:"client_tx_id,omitempty"`
	CashierID       uint           `gorm:"not null;index" json:"-"`
	CashierName     string         `gorm:"not null" json:"cashier_name"`
	Items           trxItemList    `gorm:"serializer:json" json:"items"`
	Subtotal        int64          `gorm:"not null;check:subtotal >= 0" json:"subtotal"`
	Discount        int64          `gorm:"not null;default:0;check:discount >= 0" json:"discount"`
	PromoID         *uint          `gorm:"index" json:"promo_id,omitempty"`
	PromoName       string         `gorm:"not null;default:''" json:"promo_name,omitempty"`
	PromoDiscount   int64          `gorm:"not null;default:0;check:promo_discount >= 0" json:"promo_discount"`
	VoucherID       *uint          `gorm:"index" json:"voucher_id,omitempty"`
	VoucherCode     string         `gorm:"not null;default:''" json:"voucher_code,omitempty"`
	VoucherDiscount int64          `gorm:"not null;default:0;check:voucher_discount >= 0" json:"voucher_discount"`
	Tax             int64          `gorm:"not null;default:0;check:tax >= 0" json:"tax"`
	Total           int64          `gorm:"not null;check:total >= 0" json:"total"`
	Method          string         `gorm:"not null" json:"method"`
	Paid            int64          `gorm:"not null;check:paid >= 0" json:"paid"`
	Change          int64          `gorm:"not null;default:0;check:change >= 0" json:"change"`
	Status          trxStatus      `gorm:"not null;default:'completed'" json:"status"`
	Customer        string         `gorm:"not null;default:''" json:"customer"`
	CustomerID      *uint          `gorm:"index" json:"customer_id"`
	PointsEarned    int64          `gorm:"not null;default:0;check:points_earned >= 0" json:"points_earned"`
	PointsRedeemed  int64          `gorm:"not null;default:0;check:points_redeemed >= 0" json:"points_redeemed"`
	PointsDiscount  int64          `gorm:"not null;default:0;check:points_discount >= 0" json:"points_discount"`
	FinalStocks     []finalStock   `gorm:"-" json:"final_stocks,omitempty"`
	Refunds         []refundDetail `gorm:"-" json:"-"`
}

func (trx) TableName() string {
	return "transactions"
}

type transactionItem struct {
	database.BaseModel
	TrxID       uint   `gorm:"not null;index" json:"trx_id"`
	ProductID   uint   `gorm:"not null" json:"product_id"`
	VariantID   *uint  `gorm:"index" json:"variant_id,omitempty"`
	VariantName string `gorm:"not null;default:''" json:"variant_name,omitempty"`
	Name        string `gorm:"not null" json:"name"`
	BuyPrice    int64  `gorm:"not null" json:"buy_price"`
	Price       int64  `gorm:"not null" json:"price"`
	Qty         int    `gorm:"not null;check:qty > 0" json:"qty"`
}

func (transactionItem) TableName() string {
	return "transaction_items"
}

type refund struct {
	database.BaseModel
	StoreID      uint           `gorm:"not null" json:"-"`
	TrxID        uint           `gorm:"not null;index" json:"trx_id"`
	Items        refundItemList `gorm:"serializer:json" json:"-"`
	VariantItems map[uint]int   `gorm:"serializer:json" json:"-"`
	RawItems     []trxItem      `gorm:"-" json:"items"`
	Reason       string         `gorm:"not null" json:"reason"`
	ByName       string         `gorm:"not null" json:"by"`
}

type refundDetail struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Items     []trxItem `json:"items"`
	Reason    string    `json:"reason"`
	Actor     string    `json:"actor"`
}

func (refund) TableName() string {
	return "refunds"
}
