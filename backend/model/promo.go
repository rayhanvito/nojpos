package model

import (
	"encoding/json"
	"time"
)

type Promo struct {
	Model
	StoreID     uint      `gorm:"not null;index;uniqueIndex:uq_promo_store_voucher" json:"store_id"`
	Name        string    `gorm:"not null" json:"name"`
	Type        string    `gorm:"not null" json:"type"`
	Value       float64   `gorm:"not null;default:0" json:"-"`
	AmountValue int64     `gorm:"not null;default:0;check:amount_value >= 0" json:"-"`
	Scope       string    `gorm:"not null" json:"scope"`
	CategoryID  *uint     `gorm:"index" json:"category_id,omitempty"`
	ProductID   *uint     `gorm:"index" json:"product_id,omitempty"`
	VoucherCode *string   `gorm:"uniqueIndex:uq_promo_store_voucher" json:"voucher_code"`
	StartsAt    time.Time `gorm:"not null;index" json:"starts_at"`
	EndsAt      time.Time `gorm:"not null;index" json:"ends_at"`
	HappyStart  *string   `json:"happy_start,omitempty"`
	HappyEnd    *string   `json:"happy_end,omitempty"`
	MinPurchase int64     `gorm:"not null;default:0;check:min_purchase >= 0" json:"min_purchase"`
	MaxDiscount int64     `gorm:"not null;default:0;check:max_discount >= 0" json:"max_discount"`
	Combinable  bool      `gorm:"not null;default:false" json:"combinable"`
	Active      bool      `gorm:"not null;default:true" json:"active"`
}

func (Promo) TableName() string { return "promos" }

func (p Promo) MarshalJSON() ([]byte, error) {
	type promoJSON Promo
	value := any(p.Value)
	if p.Type == "amount" {
		value = p.AmountValue
	}
	return json.Marshal(struct {
		promoJSON
		Value any `json:"value"`
	}{promoJSON(p), value})
}

type PromoPreview struct {
	PromoID       uint   `json:"promo_id"`
	PromoName     string `json:"promo_name"`
	PromoDiscount int64  `json:"promo_discount"`
	// Lapisan voucher (berkode) — terpisah dari promo otomatis supaya
	// pencatatan dan laporan bisa membedakan sumber diskon.
	VoucherID       uint   `json:"voucher_id"`
	VoucherName     string `json:"voucher_name"`
	VoucherCode     string `json:"voucher_code"`
	VoucherDiscount int64  `json:"voucher_discount"`
}
