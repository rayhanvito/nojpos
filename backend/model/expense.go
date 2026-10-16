package model

import "time"

// Expense = pengeluaran operasional toko (token listrik, wifi, sewa, ...).
// Bukan pembelian stok: tidak menyentuh inventaris maupun hutang supplier.
// Source "drawer" + Method Cash berarti uang diambil dari laci kasir; bila
// pencatat punya shift terbuka, pengeluaran tertaut ke shift itu dan
// mengurangi kas yang diharapkan saat tutup shift.
type Expense struct {
	Model
	StoreID       uint      `gorm:"not null;index" json:"store_id"`
	OutletID      *uint     `gorm:"index" json:"outlet_id"`
	Date          time.Time `gorm:"type:date;not null;index" json:"date"`
	Category      string    `gorm:"not null" json:"category"`
	Note          string    `gorm:"not null;default:''" json:"note"`
	Amount        int64     `gorm:"not null" json:"amount"`
	Method        string    `gorm:"not null" json:"method"`
	Source        string    `gorm:"not null;default:'outside'" json:"source"`
	ShiftID       *uint     `gorm:"index" json:"shift_id"`
	CreatedByID   uint      `gorm:"not null;default:0" json:"created_by_id"`
	CreatedByName string    `gorm:"not null;default:''" json:"created_by_name"`
}

func (Expense) TableName() string { return "expenses" }

const (
	ExpenseSourceDrawer  = "drawer"
	ExpenseSourceOutside = "outside"
)
