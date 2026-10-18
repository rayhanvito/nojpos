package app

import (
	"github.com/rayhanvito/nojpos/backend/internal/activity"
	"github.com/rayhanvito/nojpos/backend/internal/apikey"
	"github.com/rayhanvito/nojpos/backend/internal/auth"
	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/customer"
	"github.com/rayhanvito/nojpos/backend/internal/finance"
	"github.com/rayhanvito/nojpos/backend/internal/inventory"
	"github.com/rayhanvito/nojpos/backend/internal/notification"
	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	"github.com/rayhanvito/nojpos/backend/internal/promo"
	"github.com/rayhanvito/nojpos/backend/internal/purchasing"
	"github.com/rayhanvito/nojpos/backend/internal/sales"
	"github.com/rayhanvito/nojpos/backend/internal/shift"
	"github.com/rayhanvito/nojpos/backend/internal/staff"
	"github.com/rayhanvito/nojpos/backend/internal/webhook"
)

// autoMigrateModels adalah registri model untuk AutoMigrate jalur SQLite.
// Daftar dan urutannya verbatim dari Migrate paket db lama; registri ini
// tinggal di akar komposisi (internal/app) karena internal/platform/database
// tidak boleh mengimpor model domain (siklus impor).
var autoMigrateModels = []any{
	&outlet.Outlet{},
	&outlet.Store{},
	&outlet.OutletSettings{},
	&auth.User{},
	&staff.Cashier{},
	&auth.RefreshToken{},
	&auth.EmailOtp{},
	&catalog.Category{},
	&catalog.Product{},
	&catalog.ProductVariant{},
	&inventory.ProductStock{},
	&catalog.PriceTier{},
	&promo.Promo{},
	&customer.Customer{},
	&purchasing.Supplier{},
	&purchasing.PurchaseOrder{},
	&purchasing.PurchaseOrderItem{},
	&inventory.StockOpnameSession{},
	&inventory.StockOpnameItem{},
	&inventory.Movement{},
	&sales.Trx{},
	&sales.TransactionItem{},
	&sales.Refund{},
	&finance.Receivable{},
	&finance.ReceivablePayment{},
	&finance.Payable{},
	&finance.PayablePayment{},
	&finance.Expense{},
	&shift.CashierShift{},
	&notification.Notification{},
	&activity.ActivityLog{},
	&apikey.APIKey{},
	&webhook.Webhook{},
	&webhook.Delivery{},
}

// AutoMigrateModels adalah alias ekspor dari autoMigrateModels untuk suite
// integrasi di test/integration (paket luar tidak dapat menjangkau nama
// tak-ekspor); isinya registri yang sama persis.
var AutoMigrateModels = autoMigrateModels
