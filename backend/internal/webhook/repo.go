package webhook

import (
	"context"
	"encoding/json"

	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
)

// Repo adalah penyimpanan webhook + catatan pengirimannya, mengikuti
// pola BaseRepo proyek (DB(ctx) ikut bergabung ke transaksi pemanggil).
type Repo struct {
	database.BaseRepo
}

func NewRepo(db *gorm.DB) *Repo {
	return &Repo{BaseRepo: database.NewBaseRepo(db)}
}

func (r *Repo) Create(ctx context.Context, w *Webhook) error {
	return r.DB(ctx).Create(w).Error
}

// List mengembalikan semua webhook aktif (belum soft-delete) milik toko,
// terbaru dulu.
func (r *Repo) List(ctx context.Context, storeID uint) ([]Webhook, error) {
	items := make([]Webhook, 0)
	err := r.DB(ctx).Where("store_id = ?", storeID).
		Order("created_at DESC, id DESC").
		Find(&items).Error
	return items, err
}

// Delete menghapus lunak webhook milik toko; id milik toko lain (atau
// yang sudah terhapus) menghasilkan ErrNotFound.
func (r *Repo) Delete(ctx context.Context, storeID, id uint) error {
	result := r.DB(ctx).Where("id = ? AND store_id = ?", id, storeID).Delete(&Webhook{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

// ListSubscribed mengembalikan webhook aktif milik toko yang
// berlangganan event tertentu. Penyaringan dilakukan di Go dari kolom
// events (array JSON) agar identik di Postgres dan SQLite; jumlah
// webhook per toko kecil sehingga ini bukan jalur panas.
func (r *Repo) ListSubscribed(ctx context.Context, storeID uint, event string) ([]Webhook, error) {
	rows, err := r.List(ctx, storeID)
	if err != nil {
		return nil, err
	}
	out := make([]Webhook, 0, len(rows))
	for _, w := range rows {
		if subscribes(w.Events, event) {
			out = append(out, w)
		}
	}
	return out, nil
}

func (r *Repo) CreateDelivery(ctx context.Context, d *Delivery) error {
	return r.DB(ctx).Create(d).Error
}

func (r *Repo) UpdateDeliveryResult(ctx context.Context, id uint, status *int, success bool, errMsg string) error {
	return r.DB(ctx).Model(&Delivery{}).Where("id = ?", id).Updates(map[string]any{
		"response_status": status,
		"success":         success,
		"error":           errMsg,
	}).Error
}

// subscribes melaporkan apakah kolom events (array JSON) memuat event.
func subscribes(eventsJSON, event string) bool {
	var events []string
	if err := json.Unmarshal([]byte(eventsJSON), &events); err != nil {
		return false
	}
	for _, e := range events {
		if e == event {
			return true
		}
	}
	return false
}

// parseEvents membongkar kolom events untuk bentuk baca Metadata;
// data rusak diperlakukan sebagai daftar kosong (tidak pernah terjadi
// lewat jalur pembuatan yang tervalidasi).
func parseEvents(eventsJSON string) []string {
	var events []string
	if err := json.Unmarshal([]byte(eventsJSON), &events); err != nil || events == nil {
		return []string{}
	}
	return events
}
