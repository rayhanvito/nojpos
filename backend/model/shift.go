package model

import (
	"time"
)

type CashierShift struct {
	Model
	StoreID           uint       `gorm:"not null;index" json:"store_id"`
	OutletID          *uint      `gorm:"index" json:"outlet_id,omitempty"`
	CashierID         uint       `gorm:"not null;index" json:"cashier_id"`
	CashierName       string     `gorm:"not null" json:"cashier_name"`
	StartedAt         time.Time  `gorm:"not null" json:"started_at"`
	ClosedAt          *time.Time `json:"closed_at,omitempty"`
	OpeningCash       int64      `gorm:"not null;default:0" json:"opening_cash"`
	ClosingCashActual *int64     `json:"closing_cash_actual"`
	ExpectedCash      *int64     `json:"expected_cash"`
	CashVariance      *int64     `json:"cash_variance"`
	CloseNote         *string    `json:"close_note"`
	ClosedByName      *string    `json:"closed_by_name"`
}

func (CashierShift) TableName() string {
	return "cashier_shifts"
}
