package inventory

import (
	"context"

	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Engine adalah fasad sadar-konteks di atas mesin stok kanonis paket ini
// (stock.go: Get/Set/ChangeOutletStock). Sejak Unit B, paket inventory
// adalah SATU-SATUNYA pemilik mutasi stok: slice lain (catalog) tidak
// lagi menyimpan salinan mesin stok, melainkan mengorkestrasi Engine ini
// lewat Transactor — DB(ctx) menggabungkan panggilan ke transaksi aktif
// yang dibawa ctx, jadi atomisitas pemanggil utuh.
type Engine struct {
	database.BaseRepo
}

func NewEngine(db *gorm.DB) *Engine {
	return &Engine{BaseRepo: database.NewBaseRepo(db)}
}

// GetStock membaca stok satu produk/varian di satu outlet (0 bila baris
// stok belum ada).
func (e *Engine) GetStock(ctx context.Context, outletID, productID uint, variantID *uint) (int, error) {
	return GetOutletStock(e.DB(ctx), outletID, productID, variantID)
}

// SetStock menetapkan stok absolut satu produk/varian di satu outlet.
func (e *Engine) SetStock(ctx context.Context, storeID, outletID, productID uint, variantID *uint, qty int) error {
	return SetOutletStock(e.DB(ctx), storeID, outletID, productID, variantID, qty)
}

// ChangeStock mengubah stok sebesar delta; allowShortage=false menolak
// hasil negatif dengan ErrStockInsufficient kanonis (catalog).
func (e *Engine) ChangeStock(ctx context.Context, storeID, outletID, productID uint, variantID *uint, delta int, allowShortage bool) (before, after, actualDelta int, err error) {
	return ChangeOutletStock(e.DB(ctx), storeID, outletID, productID, variantID, delta, allowShortage)
}

// InsertMovement mencatat satu baris stock_movements.
func (e *Engine) InsertMovement(ctx context.Context, m *Movement) error {
	return e.DB(ctx).Create(m).Error
}

// RecordMovement adalah bentuk primitif InsertMovement untuk konsumen
// luar paket (catalog) yang tidak boleh membangun tipe Movement.
func (e *Engine) RecordMovement(ctx context.Context, storeID, outletID, productID uint, variantID *uint, movementType string, qty int, reason, actor string) error {
	return e.InsertMovement(ctx, &Movement{
		StoreID:   storeID,
		OutletID:  &outletID,
		ProductID: productID,
		VariantID: variantID,
		Type:      MovementType(movementType),
		Qty:       qty,
		Reason:    reason,
		Actor:     actor,
	})
}

// IsDefaultOutlet melaporkan apakah outlet adalah outlet default toko.
func (e *Engine) IsDefaultOutlet(ctx context.Context, storeID, outletID uint) (bool, error) {
	return isDefaultOutlet(e.DB(ctx), storeID, outletID)
}

// DefaultOutletID mengembalikan ID outlet default toko (ErrNotFound bila
// toko tidak punya outlet default).
func (e *Engine) DefaultOutletID(ctx context.Context, storeID uint) (uint, error) {
	var o outlet.Outlet
	if err := e.DB(ctx).Where("store_id = ? AND is_default = ?", storeID, true).First(&o).Error; err != nil {
		return 0, mapDBErr(err)
	}
	return o.ID, nil
}
