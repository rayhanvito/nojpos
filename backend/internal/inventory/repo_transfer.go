package inventory

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	database "github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// TransferRepo adalah penyimpanan header + item transfer stok (Unit D).
// Semua metodenya memakai DB(ctx), jadi panggilan dari dalam
// Transactor.WithinTx otomatis bergabung ke transaksi pemanggil.
type TransferRepo struct {
	database.BaseRepo
}

func NewTransferRepo(db *gorm.DB) *TransferRepo {
	return &TransferRepo{BaseRepo: database.NewBaseRepo(db)}
}

// OutletBelongsToStore melaporkan apakah outlet adalah milik toko.
func (r *TransferRepo) OutletBelongsToStore(ctx context.Context, storeID, outletID uint) (bool, error) {
	var count int64
	err := r.DB(ctx).Model(&outlet.Outlet{}).
		Where("id = ? AND store_id = ?", outletID, storeID).
		Count(&count).Error
	return count > 0, err
}

// CreateHeader menyisipkan header transfer (Items belum termasuk) dan
// mengisi ID-nya.
func (r *TransferRepo) CreateHeader(ctx context.Context, t *StockTransfer) error {
	return mapDBErr(r.DB(ctx).Omit("Items").Create(t).Error)
}

// CreateItem menyisipkan satu baris item transfer.
func (r *TransferRepo) CreateItem(ctx context.Context, item *StockTransferItem) error {
	return mapDBErr(r.DB(ctx).Create(item).Error)
}

// List mengembalikan semua transfer toko yang cocok dengan filter,
// terbaru dulu, masing-masing beserta items-nya. Paginasi diterapkan
// pemanggil lewat httpx.PaginateList (amplop kanonis Unit B).
func (r *TransferRepo) List(ctx context.Context, storeID uint, f TransferListFilter) ([]StockTransfer, error) {
	query := database.ScopeStore(r.DB(ctx).Model(&StockTransfer{}), storeID)
	if f.FromOutletID != 0 {
		query = query.Where("from_outlet_id = ?", f.FromOutletID)
	}
	if f.ToOutletID != 0 {
		query = query.Where("to_outlet_id = ?", f.ToOutletID)
	}
	if from, err := time.Parse("2006-01-02", f.From); f.From != "" && err == nil {
		query = query.Where("created_at >= ?", from)
	}
	if to, err := time.Parse("2006-01-02", f.To); f.To != "" && err == nil {
		query = query.Where("created_at < ?", to.AddDate(0, 0, 1))
	}

	items := make([]StockTransfer, 0)
	err := query.Preload("Items", func(db *gorm.DB) *gorm.DB {
		return db.Order("id ASC")
	}).Order("created_at DESC, id DESC").Find(&items).Error
	if err != nil {
		return nil, err
	}
	return items, nil
}
