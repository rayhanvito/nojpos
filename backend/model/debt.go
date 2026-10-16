package model

import "time"

type DebtStatus string

const (
	DebtUnpaid        DebtStatus = "unpaid"
	DebtPartiallyPaid DebtStatus = "partially_paid"
	DebtPaid          DebtStatus = "paid"
)

type Receivable struct {
	Model
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
	Model
	StoreID      uint   `gorm:"not null;index" json:"-"`
	ReceivableID uint   `gorm:"not null;index" json:"receivable_id"`
	Amount       int64  `gorm:"not null;check:amount > 0" json:"amount"`
	Method       string `gorm:"not null" json:"method"`
	CreatedBy    string `gorm:"not null;default:''" json:"created_by"`
}

func (ReceivablePayment) TableName() string { return "receivable_payments" }

type Payable struct {
	Model
	StoreID           uint            `gorm:"not null;index" json:"-"`
	SupplierID        uint            `gorm:"not null;index" json:"supplier_id"`
	PurchaseOrderID   uint            `gorm:"not null;index" json:"purchase_order_id"`
	Amount            int64           `gorm:"not null;check:amount >= 0" json:"amount"`
	Status            DebtStatus      `gorm:"not null;default:'unpaid';index" json:"status"`
	Remaining         int64           `gorm:"-" json:"remaining"`
	SupplierName      string          `gorm:"-" json:"supplier_name"`
	PurchaseOrderNumber string        `gorm:"-" json:"purchase_order_number"`
	Payments          []PayablePayment `gorm:"-" json:"payments,omitempty"`
}

func (Payable) TableName() string { return "payables" }

type PayablePayment struct {
	Model
	StoreID    uint   `gorm:"not null;index" json:"-"`
	PayableID  uint   `gorm:"not null;index" json:"payable_id"`
	Amount     int64  `gorm:"not null;check:amount > 0" json:"amount"`
	Method     string `gorm:"not null" json:"method"`
	CreatedBy  string `gorm:"not null;default:''" json:"created_by"`
}

func (PayablePayment) TableName() string { return "payable_payments" }
