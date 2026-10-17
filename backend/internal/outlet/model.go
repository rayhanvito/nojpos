package outlet

import (
	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Store adalah entitas toko (dipindahkan verbatim dari model/models.go).
type Store struct {
	database.BaseModel
	Name          string  `gorm:"not null" json:"name"`
	Address       string  `gorm:"not null;default:''" json:"address"`
	Phone         string  `gorm:"not null;default:''" json:"phone"`
	TaxEnabled    bool    `gorm:"not null;default:false" json:"taxEnabled"`
	TaxPct        float64 `gorm:"not null;default:0" json:"taxPct"`
	ReceiptHeader string  `gorm:"not null;default:'Terima kasih sudah berbelanja'" json:"receiptHeader"`
	ReceiptFooter string  `gorm:"not null;default:'Barang yang sudah dibeli tidak dapat ditukar'" json:"receiptFooter"`
	Paper         string  `gorm:"not null;default:'58mm'" json:"paper"`
	Timezone      string  `gorm:"not null;default:'Asia/Makassar'" json:"timezone"`
	// Extended settings (kontrak Extended Store Settings).
	BusinessType        string `gorm:"not null;default:''" json:"businessType"`
	Email               string `gorm:"not null;default:''" json:"email"`
	City                string `gorm:"not null;default:''" json:"city"`
	Province            string `gorm:"not null;default:''" json:"province"`
	Currency            string `gorm:"not null;default:'IDR'" json:"currency"`
	Hours               string `gorm:"not null;default:'[]'" json:"hours"`
	ReceiptShowLogo     bool   `gorm:"not null;default:true" json:"receiptShowLogo"`
	ReceiptShowCashier  bool   `gorm:"not null;default:true" json:"receiptShowCashier"`
	ReceiptShowMethod   bool   `gorm:"not null;default:true" json:"receiptShowMethod"`
	ReceiptShowTax      bool   `gorm:"not null;default:true" json:"receiptShowTax"`
	ReceiptShowDiscount bool   `gorm:"not null;default:true" json:"receiptShowDiscount"`
	ReceiptShowNote     bool   `gorm:"not null;default:true" json:"receiptShowNote"`
	TaxName             string `gorm:"not null;default:''" json:"taxName"`
	TaxInclusive        bool   `gorm:"not null;default:false" json:"taxInclusive"`
	TaxRounding         string `gorm:"not null;default:'none'" json:"taxRounding"`
	TaxApplyTo          string `gorm:"not null;default:'all'" json:"taxApplyTo"`
	LoyaltyEnabled      bool   `gorm:"not null;default:false" json:"loyaltyEnabled"`
	LoyaltyEarnPer      int64  `gorm:"not null;default:10000" json:"loyaltyEarnPer"`
	LoyaltyPointValue   int64  `gorm:"not null;default:100" json:"loyaltyPointValue"`
	CreditDueDays       int    `gorm:"not null;default:14;check:credit_due_days >= 0" json:"creditDueDays"`
	LowStockDefault     int    `gorm:"not null;default:0;check:low_stock_default >= 0" json:"lowStockDefault"`
}

func (Store) TableName() string {
	return "stores"
}

// Outlet adalah entitas outlet/cabang toko (dipindahkan verbatim dari
// model/outlet.go).
type Outlet struct {
	database.BaseModel
	StoreID   uint   `gorm:"not null;index" json:"store_id"`
	Name      string `gorm:"not null" json:"name"`
	Address   string `gorm:"not null;default:''" json:"address"`
	IsDefault bool   `gorm:"not null;default:false" json:"is_default"`
}

func (Outlet) TableName() string { return "outlets" }

// StoreSettings adalah DTO pengaturan toko (dipindahkan verbatim dari
// model/reporting.go; bukan entitas, tidak ada di registri migrasi).
// Tipe jam operasional adalah StoreHours di paket ini.
type StoreSettings struct {
	Name                string       `json:"storeName"`
	Address             string       `json:"address"`
	Phone               string       `json:"phone"`
	TaxEnabled          bool         `json:"taxEnabled"`
	TaxPct              float64      `json:"taxPct"`
	ReceiptHeader       string       `json:"receiptHeader"`
	ReceiptFooter       string       `json:"receiptFooter"`
	Paper               string       `json:"paper"`
	Timezone            string       `json:"timezone"`
	BusinessType        string       `json:"businessType"`
	Email               string       `json:"email"`
	City                string       `json:"city"`
	Province            string       `json:"province"`
	Currency            string       `json:"currency"`
	Hours               []StoreHours `json:"hours"`
	ReceiptShowLogo     bool         `json:"receiptShowLogo"`
	ReceiptShowCashier  bool         `json:"receiptShowCashier"`
	ReceiptShowMethod   bool         `json:"receiptShowMethod"`
	ReceiptShowTax      bool         `json:"receiptShowTax"`
	ReceiptShowDiscount bool         `json:"receiptShowDiscount"`
	ReceiptShowNote     bool         `json:"receiptShowNote"`
	TaxName             string       `json:"taxName"`
	TaxInclusive        bool         `json:"taxInclusive"`
	TaxRounding         string       `json:"taxRounding"`
	TaxApplyTo          string       `json:"taxApplyTo"`
	LoyaltyEnabled      bool         `json:"loyalty_enabled"`
	LoyaltyEarnPer      int64        `json:"loyalty_earn_per"`
	LoyaltyPointValue   int64        `json:"loyalty_point_value"`
	CreditDueDays       int          `json:"creditDueDays"`
	LowStockDefault     int          `json:"lowStockDefault"`
}

// StoreHours adalah tipe jam operasional toko (dipindah verbatim dari
// model/reporting.go; definisi kanonisnya di slice ini sejak BE-F10
// karena StoreSettings di paket ini memakainya dan slice report di
// puncak DAG tidak bisa diimpor dari sini).
type StoreHours struct {
	Days  string  `json:"days"`
	Open  *string `json:"open"`
	Close *string `json:"close"`
}
