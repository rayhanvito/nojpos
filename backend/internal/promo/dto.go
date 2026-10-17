package promo

import (
	"encoding/json"
	"strconv"
	"time"
)

// PromoInput adalah masukan tervalidasi untuk membuat/mengubah promo
// (dipindah verbatim dari service.PromoInput lama).
type PromoInput struct {
	Name        string
	Type        string
	Value       float64
	AmountValue int64
	Scope       string
	CategoryID  *uint
	ProductID   *uint
	VoucherCode *string
	StartsAt    time.Time
	EndsAt      time.Time
	HappyStart  *string
	HappyEnd    *string
	MinPurchase int64
	MaxDiscount int64
	Combinable  bool
}

// promoRequest adalah bentuk JSON permintaan promo (promoReq lama di
// paket handler); tag JSON tidak berubah.
type promoRequest struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Value       json.Number `json:"value"`
	Scope       string      `json:"scope"`
	CategoryID  *uint       `json:"categoryId"`
	ProductID   *uint       `json:"productId"`
	VoucherCode *string     `json:"voucherCode"`
	StartsAt    time.Time   `json:"startsAt"`
	EndsAt      time.Time   `json:"endsAt"`
	HappyStart  *string     `json:"happyStart"`
	HappyEnd    *string     `json:"happyEnd"`
	MinPurchase int64       `json:"minPurchase"`
	MaxDiscount int64       `json:"maxDiscount"`
	Combinable  bool        `json:"combinable"`
}

func promoInput(req promoRequest) (PromoInput, error) {
	in := PromoInput{
		Name: req.Name, Type: req.Type, Scope: req.Scope,
		CategoryID: req.CategoryID, ProductID: req.ProductID, VoucherCode: req.VoucherCode,
		StartsAt: req.StartsAt, EndsAt: req.EndsAt, HappyStart: req.HappyStart, HappyEnd: req.HappyEnd,
		MinPurchase: req.MinPurchase, MaxDiscount: req.MaxDiscount, Combinable: req.Combinable,
	}
	switch req.Type {
	case "percent":
		value, err := strconv.ParseFloat(req.Value.String(), 64)
		if err != nil {
			return in, err
		}
		in.Value = value
	case "amount":
		value, err := strconv.ParseInt(req.Value.String(), 10, 64)
		if err != nil {
			return in, err
		}
		in.AmountValue = value
	}
	return in, nil
}
