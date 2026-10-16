package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

type PayMethod string

const (
	PayCash         PayMethod = "Cash"
	PayBankTransfer PayMethod = "Bank Transfer"
	PayQRIS         PayMethod = "QRIS"
	PayEWallet      PayMethod = "E-Wallet"
	PayCard         PayMethod = "Card"
	PayCredit       PayMethod = "Credit"
)

func ValidPayMethod(m string) bool {
	return PayMethod(m) == PayCredit || ValidStandardPayMethod(m)
}

func ValidStandardPayMethod(m string) bool {
	switch PayMethod(m) {
	case PayCash, PayBankTransfer, PayQRIS:
		return true
	}
	return false
}

type TrxStatus string

const (
	TrxPending                  TrxStatus = "pending"
	TrxCompleted                TrxStatus = "completed"
	TrxCompletedOfflineShortage TrxStatus = "completed_offline_shortage"
	TrxCancelled                TrxStatus = "cancelled"
	TrxRefunded                 TrxStatus = "refunded"
)

type FinalStock struct {
	ProductID uint  `json:"product_id"`
	VariantID *uint `json:"variant_id"`
	Stock     int   `json:"stock"`
}

type TrxItem struct {
	ProductID   uint   `json:"product_id"`
	VariantID   *uint  `json:"variant_id,omitempty"`
	VariantName string `json:"variant_name,omitempty"`
	Name        string `json:"name"`
	BuyPrice    int64  `json:"buy_price"`
	Price       int64  `json:"price"`
	Qty         int    `json:"qty"`
}

type TrxItemList []TrxItem

func (l *TrxItemList) Scan(value interface{}) error {
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

func (l TrxItemList) Value() (driver.Value, error) {
	return json.Marshal(l)
}

type RefundItemList map[uint]int

func (m *RefundItemList) Scan(value interface{}) error {
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

func (m RefundItemList) Value() (driver.Value, error) {
	return json.Marshal(m)
}

type Trx struct {
	Model
	StoreID         uint           `gorm:"not null;index;uniqueIndex:uq_transactions_store_client_tx_id,priority:1,where:client_tx_id <> ''" json:"-"`
	OutletID        *uint          `gorm:"index" json:"outlet_id,omitempty"`
	ClientTxID      string         `gorm:"not null;default:'';uniqueIndex:uq_transactions_store_client_tx_id,priority:2,where:client_tx_id <> ''" json:"client_tx_id,omitempty"`
	CashierID       uint           `gorm:"not null;index" json:"-"`
	CashierName     string         `gorm:"not null" json:"cashier_name"`
	Items           TrxItemList    `gorm:"serializer:json" json:"items"`
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
	Status          TrxStatus      `gorm:"not null;default:'completed'" json:"status"`
	Customer        string         `gorm:"not null;default:''" json:"customer"`
	CustomerID      *uint          `gorm:"index" json:"customer_id"`
	PointsEarned    int64          `gorm:"not null;default:0;check:points_earned >= 0" json:"points_earned"`
	PointsRedeemed  int64          `gorm:"not null;default:0;check:points_redeemed >= 0" json:"points_redeemed"`
	PointsDiscount  int64          `gorm:"not null;default:0;check:points_discount >= 0" json:"points_discount"`
	FinalStocks     []FinalStock   `gorm:"-" json:"final_stocks,omitempty"`
	Refunds         []RefundDetail `gorm:"-" json:"-"`
}

func (Trx) TableName() string {
	return "transactions"
}

type TransactionItem struct {
	Model
	TrxID       uint   `gorm:"not null;index" json:"trx_id"`
	ProductID   uint   `gorm:"not null" json:"product_id"`
	VariantID   *uint  `gorm:"index" json:"variant_id,omitempty"`
	VariantName string `gorm:"not null;default:''" json:"variant_name,omitempty"`
	Name        string `gorm:"not null" json:"name"`
	BuyPrice    int64  `gorm:"not null" json:"buy_price"`
	Price       int64  `gorm:"not null" json:"price"`
	Qty         int    `gorm:"not null;check:qty > 0" json:"qty"`
}

func (TransactionItem) TableName() string {
	return "transaction_items"
}

type Refund struct {
	Model
	StoreID      uint           `gorm:"not null" json:"-"`
	TrxID        uint           `gorm:"not null;index" json:"trx_id"`
	Items        RefundItemList `gorm:"serializer:json" json:"-"`
	VariantItems map[uint]int   `gorm:"serializer:json" json:"-"`
	RawItems     []TrxItem      `gorm:"-" json:"items"`
	Reason       string         `gorm:"not null" json:"reason"`
	ByName       string         `gorm:"not null" json:"by"`
}

type RefundDetail struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Items     []TrxItem `json:"items"`
	Reason    string    `json:"reason"`
	Actor     string    `json:"actor"`
}

func (Refund) TableName() string {
	return "refunds"
}
