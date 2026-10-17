package finance

import (
	"time"

	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Tipe utang-piutang (dipindah verbatim dari model/debt.go, BE-F6).

type DebtStatus string

const (
	DebtUnpaid        DebtStatus = "unpaid"
	DebtPartiallyPaid DebtStatus = "partially_paid"
	DebtPaid          DebtStatus = "paid"
)

type Receivable struct {
	database.BaseModel
	StoreID       uint                `gorm:"not null;index" json:"-"`
	CustomerID    uint                `gorm:"not null;index" json:"customer_id"`
	TransactionID uint                `gorm:"not null;uniqueIndex" json:"transaction_id"`
	Amount        int64               `gorm:"not null;check:amount >= 0" json:"amount"`
	DueDate       time.Time           `gorm:"not null;index" json:"due_date"`
	Status        DebtStatus          `gorm:"not null;default:'unpaid';index" json:"status"`
	Remaining     int64               `gorm:"-" json:"remaining"`
	CustomerName  string              `gorm:"-" json:"customer_name"`
	InvoiceNumber string              `gorm:"-" json:"invoice_number"`
	Payments      []ReceivablePayment `gorm:"-" json:"payments,omitempty"`
}

func (Receivable) TableName() string { return "receivables" }

type ReceivablePayment struct {
	database.BaseModel
	StoreID      uint   `gorm:"not null;index" json:"-"`
	ReceivableID uint   `gorm:"not null;index" json:"receivable_id"`
	Amount       int64  `gorm:"not null;check:amount > 0" json:"amount"`
	Method       string `gorm:"not null" json:"method"`
	CreatedBy    string `gorm:"not null;default:''" json:"created_by"`
}

func (ReceivablePayment) TableName() string { return "receivable_payments" }

type Payable struct {
	database.BaseModel
	StoreID             uint             `gorm:"not null;index" json:"-"`
	SupplierID          uint             `gorm:"not null;index" json:"supplier_id"`
	PurchaseOrderID     uint             `gorm:"not null;index" json:"purchase_order_id"`
	Amount              int64            `gorm:"not null;check:amount >= 0" json:"amount"`
	Status              DebtStatus       `gorm:"not null;default:'unpaid';index" json:"status"`
	Remaining           int64            `gorm:"-" json:"remaining"`
	SupplierName        string           `gorm:"-" json:"supplier_name"`
	PurchaseOrderNumber string           `gorm:"-" json:"purchase_order_number"`
	Payments            []PayablePayment `gorm:"-" json:"payments,omitempty"`
}

func (Payable) TableName() string { return "payables" }

type PayablePayment struct {
	database.BaseModel
	StoreID   uint   `gorm:"not null;index" json:"-"`
	PayableID uint   `gorm:"not null;index" json:"payable_id"`
	Amount    int64  `gorm:"not null;check:amount > 0" json:"amount"`
	Method    string `gorm:"not null" json:"method"`
	CreatedBy string `gorm:"not null;default:''" json:"created_by"`
}

func (PayablePayment) TableName() string { return "payable_payments" }

// Expense = pengeluaran operasional toko (token listrik, wifi, sewa, ...).
// Bukan pembelian stok: tidak menyentuh inventaris maupun hutang supplier.
// Source "drawer" + Method Cash berarti uang diambil dari laci kasir; bila
// pencatat punya shift terbuka, pengeluaran tertaut ke shift itu dan
// mengurangi kas yang diharapkan saat tutup shift.
// (Dipindah verbatim dari model/expense.go, BE-F6.)
type Expense struct {
	database.BaseModel
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

// validStandardPayMethod adalah salinan privat slice dari
// ValidStandardPayMethod kanonis di slice sales (BE-F8): sales
// mengimpor finance (piutang Credit di checkout), jadi finance —
// di bawah sales di DAG — tidak boleh mengimpor balik sales.
// Himpunan metodenya verbatim: Cash, Bank Transfer, QRIS. Metode
// tunai dibandingkan sebagai literal "Cash" setempat dengan nilai
// yang sama (payCash di pemakai di bawah).
const payCash = "Cash"

func validStandardPayMethod(m string) bool {
	switch m {
	case payCash, "Bank Transfer", "QRIS":
		return true
	}
	return false
}
