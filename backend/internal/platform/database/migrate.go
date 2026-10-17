package database

import "gorm.io/gorm"

// Migrate menjalankan AutoMigrate atas daftar model yang diberikan
// pemanggil. Registri model diserahkan ke akar komposisi (router) agar
// paket ini tidak mengimpor model domain mana pun. Jalur Postgres tidak
// memakai fungsi ini — skemanya dikelola migrasi goose.
func Migrate(db *gorm.DB, models ...any) error {
	return db.AutoMigrate(models...)
}
