package inventory

import (
	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// StockTransfer adalah header pemindahan stok antar outlet dalam satu
// toko (Unit D). Items diisi dari stock_transfer_items; relasinya logis
// (tanpa FK database, mengikuti konvensi skema proyek).
type StockTransfer struct {
	database.BaseModel
	StoreID       uint                `gorm:"not null;index" json:"-"`
	FromOutletID  uint                `gorm:"not null;index" json:"from_outlet_id"`
	ToOutletID    uint                `gorm:"not null;index" json:"to_outlet_id"`
	Note          string              `gorm:"not null;default:''" json:"note"`
	CreatedBy     uint                `gorm:"not null;default:0" json:"created_by"`
	CreatedByName string              `gorm:"not null;default:''" json:"created_by_name"`
	Items         []StockTransferItem `gorm:"foreignKey:TransferID" json:"items"`
}

func (StockTransfer) TableName() string { return "stock_transfers" }

// StockTransferItem adalah satu baris barang dalam sebuah transfer stok.
type StockTransferItem struct {
	database.BaseModel
	TransferID uint  `gorm:"not null;index" json:"transfer_id"`
	ProductID  uint  `gorm:"not null;index" json:"product_id"`
	VariantID  *uint `gorm:"index" json:"variant_id,omitempty"`
	Qty        int   `gorm:"not null" json:"qty"`
}

func (StockTransferItem) TableName() string { return "stock_transfer_items" }
