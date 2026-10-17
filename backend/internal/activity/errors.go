package activity

import (
	"errors"
	"strings"

	"gorm.io/gorm"

	"github.com/rayhanvito/nojpos/backend/internal/outlet"
)

var (
	// ErrNotFound dan ErrDuplicate adalah alias ke instans kanonis
	// slice outlet (BE-F1a) — instans yang sama dengan
	// repo.ErrNotFound / repo.ErrDuplicate lama, jadi pemeriksaan
	// errors.Is konsumen menetap tetap cocok.
	ErrNotFound  = outlet.ErrNotFound
	ErrDuplicate = outlet.ErrDuplicate
)

// mapDBErr memetakan error database ke sentinel slice, setara mapDBErr
// lama di repo/base.go: record kosong -> ErrNotFound; pelanggaran unik
// (Postgres 23505 / SQLite) -> ErrDuplicate.
func mapDBErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "23505") ||
			strings.Contains(msg, "duplicate key") ||
			strings.Contains(msg, "UNIQUE constraint failed") ||
			strings.Contains(msg, "unique constraint") {
			return ErrDuplicate
		}
	}
	return err
}
