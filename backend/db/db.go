package db

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	"github.com/adrr-dev/openPOS/backend/model"
)

// DB_DRIVER selects the database explicitly:
//
//	DB_DRIVER=sqlite   -> local file DB (dev/test, no Supabase needed)
//	DB_DRIVER=postgres  -> force Supabase/Postgres even if URL looks like a file
//	DB_DRIVER unset     -> auto-detect: file:/.db/sqlite in DATABASE_URL means
//	                       sqlite, otherwise postgres. On Vercel just leave it
//	                       unset and set DATABASE_URL to the pooler string.
func Connect(ctx context.Context, databaseURL string) (*gorm.DB, error) {
	var dialector gorm.Dialector

	driver := strings.ToLower(strings.TrimSpace(os.Getenv("DB_DRIVER")))
	lowerURL := strings.ToLower(databaseURL)
	isSQLite := driver == "sqlite" ||
		(driver == "" && (strings.HasPrefix(lowerURL, "file:") ||
			strings.Contains(lowerURL, ".db") ||
			strings.Contains(lowerURL, "sqlite")))

	if isSQLite {
		if databaseURL == "" || databaseURL == "postgres://..." {
			databaseURL = "openpos_test.db"
		}
		dialector = sqlite.Open(databaseURL)
	} else {
		// Force simple protocol so Supabase PgBouncer (port 6543,
		// transaction mode) works, which doesn't support prepared statements.
		dialector = postgres.New(postgres.Config{
			DSN:                  databaseURL,
			PreferSimpleProtocol: true,
		})
	}

	// Quiet SQL logs on hosted Postgres (Vercel log spam), verbose locally.
	// Set LOG_SQL=true to force query logging anywhere.
	logLevel := logger.Warn
	if isSQLite || strings.EqualFold(os.Getenv("LOG_SQL"), "true") {
		logLevel = logger.Info
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
		// All auto timestamps in UTC: sqlite compares datetimes
		// lexicographically, so a single offset keeps range filters
		// (dashboard/reports "today") chronological. Postgres stores
		// the same instants as timestamptz either way.
		NowFunc: func() time.Time { return time.Now().UTC() },
	})
	if err != nil {
		return nil, fmt.Errorf("gagal terhubung ke database via GORM: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan sql.DB: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("gagal ping database: %w", err)
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.Outlet{},
		&model.Store{},
		&model.User{},
		&model.Cashier{},
		&model.RefreshToken{},
		&model.EmailOtp{},
		&model.Category{},
		&model.Product{},
		&model.ProductVariant{},
		&model.ProductStock{},
		&model.PriceTier{},
		&model.Promo{},
		&model.Customer{},
		&model.Supplier{},
		&model.PurchaseOrder{},
		&model.PurchaseOrderItem{},
		&model.StockOpnameSession{},
		&model.StockOpnameItem{},
		&model.Movement{},
		&model.Trx{},
		&model.TransactionItem{},
		&model.Refund{},
		&model.Receivable{},
		&model.ReceivablePayment{},
		&model.Payable{},
		&model.PayablePayment{},
		&model.Expense{},
		&model.CashierShift{},
		&model.Notification{},
		&model.ActivityLog{},
	); err != nil {
		return err
	}
	return backfillDefaultOutlets(db)
}

func backfillDefaultOutlets(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		var stores []model.Store
		if err := tx.Find(&stores).Error; err != nil {
			return err
		}
		for _, store := range stores {
			outlet, err := defaultOutlet(tx, store)
			if err != nil {
				return err
			}
			for _, entity := range []any{
				&model.Product{}, &model.ProductVariant{}, &model.User{}, &model.Cashier{},
				&model.Trx{}, &model.CashierShift{}, &model.PurchaseOrder{},
				&model.StockOpnameSession{}, &model.Movement{},
			} {
				if err := tx.Unscoped().Model(entity).
					Where("store_id = ? AND outlet_id IS NULL", store.ID).
					Update("outlet_id", outlet.ID).Error; err != nil {
					return err
				}
			}
			if err := tx.Model(&model.RefreshToken{}).
				Where("user_id IN (SELECT id FROM users WHERE store_id = ?) AND outlet_id IS NULL", store.ID).
				Update("outlet_id", outlet.ID).Error; err != nil {
				return err
			}
			if err := seedDefaultOutletStocks(tx, store.ID, outlet.ID); err != nil {
				return err
			}
		}
		return nil
	})
}

func defaultOutlet(tx *gorm.DB, store model.Store) (*model.Outlet, error) {
	var outlets []model.Outlet
	if err := tx.Unscoped().Where("store_id = ?", store.ID).Order("id ASC").Find(&outlets).Error; err != nil {
		return nil, err
	}
	var selected *model.Outlet
	for i := range outlets {
		if outlets[i].IsDefault {
			selected = &outlets[i]
			break
		}
	}
	if selected == nil && len(outlets) > 0 {
		selected = &outlets[0]
	}
	if selected == nil {
		name := strings.TrimSpace(store.Name)
		if name == "" {
			name = "Pusat"
		}
		created := model.Outlet{StoreID: store.ID, Name: name, Address: store.Address, IsDefault: true}
		if err := tx.Create(&created).Error; err != nil {
			return nil, err
		}
		return &created, nil
	}
	if err := tx.Unscoped().Model(&model.Outlet{}).Where("store_id = ?", store.ID).Update("is_default", false).Error; err != nil {
		return nil, err
	}
	if err := tx.Unscoped().Model(&model.Outlet{}).Where("id = ?", selected.ID).Update("is_default", true).Error; err != nil {
		return nil, err
	}
	selected.IsDefault = true
	return selected, nil
}

func seedDefaultOutletStocks(tx *gorm.DB, storeID, outletID uint) error {
	var products []model.Product
	if err := tx.Unscoped().Where("store_id = ?", storeID).Order("id ASC").Find(&products).Error; err != nil {
		return err
	}
	for _, product := range products {
		stock := model.ProductStock{ProductID: product.ID, OutletID: outletID, Qty: product.Stock}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&stock).Error; err != nil {
			return err
		}
	}
	var variants []model.ProductVariant
	if err := tx.Unscoped().Where("store_id = ?", storeID).Order("id ASC").Find(&variants).Error; err != nil {
		return err
	}
	for _, variant := range variants {
		variantID := variant.ID
		stock := model.ProductStock{ProductID: variant.ProductID, VariantID: &variantID, OutletID: outletID, Qty: variant.Stock}
		if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&stock).Error; err != nil {
			return err
		}
	}
	return nil
}
