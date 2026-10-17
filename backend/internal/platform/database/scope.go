package database

import "gorm.io/gorm"

// ScopeStore membatasi query ke satu toko.
func ScopeStore(db *gorm.DB, storeID uint) *gorm.DB {
	return db.Where("store_id = ?", storeID)
}

// ScopeOutlet membatasi query ke satu outlet.
func ScopeOutlet(db *gorm.DB, outletID uint) *gorm.DB {
	return db.Where("outlet_id = ?", outletID)
}
