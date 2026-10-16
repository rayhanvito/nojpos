package service

import (
	"context"
	"time"

	"github.com/adrr-dev/openPOS/backend/model"
	"github.com/adrr-dev/openPOS/backend/repo"
)

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByID(ctx context.Context, id uint) (*model.User, error)
	ListByStore(ctx context.Context, storeID uint) ([]*model.User, error)
	SetActive(ctx context.Context, id uint, active bool) error
	SetPasscode(ctx context.Context, id uint, hash *string) error
	RegisterTx(ctx context.Context, storeName, email, name, passwordHash string) (*model.User, error)
	UpdatePassword(ctx context.Context, email string, passwordHash string) error
	UpdatePasswordByID(ctx context.Context, id uint, passwordHash string) error
	UpdateLastSeenAt(ctx context.Context, id uint, lastSeenAt *time.Time) error
}

type CashierRepository interface {
	GetByID(ctx context.Context, id uint) (*model.Cashier, error)
	ListByStore(ctx context.Context, storeID uint) ([]*model.Cashier, error)
	Create(ctx context.Context, storeID, outletID uint, name string) (*model.Cashier, error)
	UpdateProfile(ctx context.Context, storeID, id uint, name *string, outletID *uint) error
	SetActive(ctx context.Context, id uint, active bool) error
	SetPasscode(ctx context.Context, id uint, hash *string) error
	GetOrCreateByName(ctx context.Context, storeID uint, name string) (uint, error)
	Delete(ctx context.Context, id uint) error
	UpdateLastSeenAt(ctx context.Context, id uint, lastSeenAt *time.Time) error
	UpdateName(ctx context.Context, id uint, name string) error
}

type OutletRepository interface {
	List(ctx context.Context, storeID uint) ([]model.Outlet, error)
	GetByID(ctx context.Context, storeID, id uint) (*model.Outlet, error)
	GetDefault(ctx context.Context, storeID uint) (*model.Outlet, error)
	Create(ctx context.Context, storeID uint, outlet *model.Outlet) (*model.Outlet, error)
	Update(ctx context.Context, storeID, id uint, outlet model.Outlet, isDefault *bool) (*model.Outlet, error)
}

type RefreshRepository interface {
	Create(ctx context.Context, userID uint, actingAsCashierID, outletID *uint, tokenHash string, expiresAt time.Time) error
	GetActiveByHash(ctx context.Context, tokenHash string) (*model.RefreshToken, error)
	Revoke(ctx context.Context, tokenHash string) error
	RevokeAllForUser(ctx context.Context, userID uint) error
}

type OtpRepository interface {
	UpsertOTP(ctx context.Context, email, codeHash string, expiresAt time.Time) error
	GetOTP(ctx context.Context, email string) (*model.EmailOtp, error)
	IncrementAttempts(ctx context.Context, email string) (int, error)
	MarkVerified(ctx context.Context, email string) error
	IsEmailVerified(ctx context.Context, email string) (bool, error)
	RevokeOTP(ctx context.Context, email string) error
}

type CategoryRepository interface {
	ListByStore(ctx context.Context, storeID uint) ([]*model.Category, error)
	GetByID(ctx context.Context, storeID, id uint) (*model.Category, error)
	Create(ctx context.Context, storeID uint, name string) (*model.Category, error)
	Delete(ctx context.Context, storeID, id uint) (softDeleted bool, err error)
}

type ProductRepository interface {
	GetByID(ctx context.Context, storeID, id uint) (*model.Product, error)
	GetByIDAtOutlet(ctx context.Context, storeID, outletID, id uint) (*model.Product, error)
	List(ctx context.Context, storeID uint, f repo.ProductFilter) (*repo.ProductPage, error)
	ListAtOutlet(ctx context.Context, storeID, outletID uint, f repo.ProductFilter) (*repo.ProductPage, error)
	ListForExport(ctx context.Context, storeID uint) ([]*model.Product, error)
	ListForExportAtOutlet(ctx context.Context, storeID, outletID uint) ([]*model.Product, error)
	Create(ctx context.Context, storeID uint, p *model.Product) (*model.Product, error)
	Update(ctx context.Context, storeID, id uint, p *model.Product) (*model.Product, error)
	SetActive(ctx context.Context, storeID, id uint, active bool) error
	IsSkuTaken(ctx context.Context, storeID uint, sku string, exceptProductID uint) (bool, error)
	CreateWithInitialMovement(ctx context.Context, storeID uint, p *model.Product, actor string) (*model.Product, error)
	CreateWithInitialMovementAtOutlet(ctx context.Context, storeID, outletID uint, p *model.Product, actor string) (*model.Product, error)
	AdjustStock(ctx context.Context, storeID, productID uint, delta int, reason, actor string) (*model.Product, error)
	AdjustStockAtOutlet(ctx context.Context, storeID, outletID, productID uint, delta int, reason, actor string) (*model.Product, error)
	AdjustVariantStock(ctx context.Context, storeID, productID, variantID uint, delta int, reason, actor string) (*model.ProductVariant, error)
	AdjustVariantStockAtOutlet(ctx context.Context, storeID, outletID, productID, variantID uint, delta int, reason, actor string) (*model.ProductVariant, error)
	ListVariants(ctx context.Context, storeID, productID uint) ([]model.ProductVariant, error)
	ListVariantsAtOutlet(ctx context.Context, storeID, outletID, productID uint) ([]model.ProductVariant, error)
	GetVariant(ctx context.Context, storeID, productID, variantID uint) (*model.ProductVariant, error)
	GetVariantAtOutlet(ctx context.Context, storeID, outletID, productID, variantID uint) (*model.ProductVariant, error)
	CreateVariant(ctx context.Context, storeID, productID uint, variant *model.ProductVariant, actor string) (*model.ProductVariant, error)
	CreateVariantAtOutlet(ctx context.Context, storeID, outletID, productID uint, variant *model.ProductVariant, actor string) (*model.ProductVariant, error)
	UpdateVariant(ctx context.Context, storeID, productID, variantID uint, variant *model.ProductVariant, actor string) (*model.ProductVariant, error)
	UpdateVariantAtOutlet(ctx context.Context, storeID, outletID, productID, variantID uint, variant *model.ProductVariant, actor string) (*model.ProductVariant, error)
	DeleteVariant(ctx context.Context, storeID, productID, variantID uint) error
	IsVariantSkuTaken(ctx context.Context, storeID uint, sku string, exceptID uint) (bool, error)
	ListPriceTiers(ctx context.Context, storeID, productID uint, variantID *uint) ([]model.PriceTier, error)
	CreatePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tier *model.PriceTier) (*model.PriceTier, error)
	UpdatePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tierID uint, tier *model.PriceTier) (*model.PriceTier, error)
	DeletePriceTier(ctx context.Context, storeID, productID uint, variantID *uint, tierID uint) error
	Delete(ctx context.Context, storeID, id uint) error
}

type MovementRepository interface {
	List(ctx context.Context, storeID uint, f repo.MovementFilter) (*repo.MovementPage, error)
	ListAtOutlet(ctx context.Context, storeID, outletID uint, f repo.MovementFilter) (*repo.MovementPage, error)
}

type PromoRepository interface {
	List(ctx context.Context, storeID uint) ([]*model.Promo, error)
	GetByID(ctx context.Context, storeID, id uint) (*model.Promo, error)
	Create(ctx context.Context, storeID uint, promo *model.Promo) (*model.Promo, error)
	Update(ctx context.Context, storeID, id uint, promo *model.Promo) (*model.Promo, error)
	SetActive(ctx context.Context, storeID, id uint, active bool) error
	Delete(ctx context.Context, storeID, id uint) error
}

type CustomerRepository interface {
	List(ctx context.Context, storeID uint, f repo.CustomerFilter) (*repo.CustomerPage, error)
	GetByID(ctx context.Context, storeID, id uint) (*model.Customer, error)
	Create(ctx context.Context, storeID uint, customer *model.Customer) (*model.Customer, error)
	Update(ctx context.Context, storeID, id uint, customer *model.Customer) (*model.Customer, error)
	Delete(ctx context.Context, storeID, id uint) error
	ListTransactions(ctx context.Context, storeID, customerID uint) ([]*model.Trx, error)
}

type ExpenseRepository interface {
	List(ctx context.Context, storeID uint, f repo.ExpenseFilter) (*repo.ExpensePage, error)
	GetByID(ctx context.Context, storeID, id uint) (*model.Expense, error)
	Create(ctx context.Context, expense *model.Expense) (*model.Expense, error)
	Update(ctx context.Context, storeID, id uint, expense *model.Expense) (*model.Expense, error)
	Delete(ctx context.Context, storeID, id uint) error
}

type SupplierRepository interface {
	List(ctx context.Context, storeID uint) ([]*model.Supplier, error)
	GetByID(ctx context.Context, storeID, id uint) (*model.Supplier, error)
	Create(ctx context.Context, storeID uint, supplier *model.Supplier) (*model.Supplier, error)
	Update(ctx context.Context, storeID, id uint, supplier *model.Supplier) (*model.Supplier, error)
	Delete(ctx context.Context, storeID, id uint) error
}

type PurchaseOrderRepository interface {
	List(ctx context.Context, storeID, outletID uint, f repo.PurchaseOrderFilter) ([]*model.PurchaseOrder, error)
	GetByID(ctx context.Context, storeID, outletID, id uint) (*model.PurchaseOrder, error)
	CreateDraft(ctx context.Context, storeID, outletID, supplierID uint, items []model.PurchaseOrderItem, datePrefix string) (*model.PurchaseOrder, error)
	UpdateDraft(ctx context.Context, storeID, outletID, id, supplierID uint, items []model.PurchaseOrderItem) (*model.PurchaseOrder, error)
	DeleteDraft(ctx context.Context, storeID, outletID, id uint) error
	Order(ctx context.Context, storeID, outletID, id uint) (*model.PurchaseOrder, error)
	Receive(ctx context.Context, storeID, outletID, id uint, items []repo.PurchaseOrderReceiptItem, actor string) (*model.PurchaseOrder, error)
	Cancel(ctx context.Context, storeID, outletID, id uint) (*model.PurchaseOrder, error)
}

type StockOpnameRepository interface {
	List(ctx context.Context, storeID, outletID uint) ([]model.StockOpnameSession, error)
	GetByID(ctx context.Context, storeID, outletID, id uint) (*model.StockOpnameSession, error)
	Create(ctx context.Context, storeID, outletID uint, note, createdBy string) (*model.StockOpnameSession, error)
	UpdateItems(ctx context.Context, storeID, outletID, id uint, items []repo.StockOpnameItemCount) (*model.StockOpnameSession, error)
	Post(ctx context.Context, storeID, outletID, id uint, actor string) (*model.StockOpnameSession, error)
	Cancel(ctx context.Context, storeID, outletID, id uint) (*model.StockOpnameSession, error)
}

type TrxRepository interface {
	Checkout(ctx context.Context, in repo.CheckoutInput) (*model.Trx, bool, error)
	PreviewPromo(ctx context.Context, in repo.PromoPreviewInput) (*model.PromoPreview, error)
	Refund(ctx context.Context, storeID, trxID uint, items []repo.RefundItem, reason, byName string) (*model.Trx, error)
	RefundAtOutlet(ctx context.Context, storeID, outletID, trxID uint, items []repo.RefundItem, reason, byName string) (*model.Trx, error)
	List(ctx context.Context, storeID uint, outletID *uint, cashierID uint, q, method, date, from, to string, status string, cashierFilter *uint, page, limit int) ([]*model.Trx, int, error)
	ListForExport(ctx context.Context, storeID uint, period, timezone, date, from, to string) ([]*model.Trx, error)
	GetByID(ctx context.Context, storeID, id uint) (*model.Trx, error)
}

type StoreRepository interface {
	GetSettings(ctx context.Context, storeID uint) (*model.StoreSettings, error)
	GetTimezone(ctx context.Context, storeID uint) (string, error)
	UpdateSettings(ctx context.Context, storeID uint, s *model.StoreSettings) (*model.StoreSettings, error)
	SetPasscode(ctx context.Context, storeID, userID uint, hash *string) error
}

type ReportRepository interface {
	Dashboard(ctx context.Context, storeID, cashierID uint, cashierView bool, outletID *uint, tz, dateStr, fromStr, toStr string) (any, error)
	Report(ctx context.Context, storeID uint, outletID *uint, period, tz, dateStr, fromStr, toStr string) (*model.ReportBundle, error)
	TaxReport(ctx context.Context, storeID uint, outletID *uint, start, end time.Time, timezone string) (*model.TaxReport, error)
}

type ShiftRepository interface {
	GetActiveShift(ctx context.Context, storeID, outletID, cashierID uint) (*model.CashierShift, error)
	StartShift(ctx context.Context, storeID, outletID, cashierID uint, cashierName string, openingCash int64) (*model.CashierShift, error)
	CloseShift(ctx context.Context, storeID, outletID, cashierID uint, closingCash int64, note *string, closedByName string) (*model.CashierShift, *repo.ShiftStats, error)
	ForceCloseShift(ctx context.Context, storeID, shiftID uint, closingCash *int64, note *string, closedByName string) (*model.CashierShift, *repo.ShiftStats, error)
	ListStoreShifts(ctx context.Context, storeID uint, outletID *uint, from string, to string) ([]*model.CashierShift, error)
	GetShiftByID(ctx context.Context, storeID, shiftID uint, outletID *uint) (*model.CashierShift, error)
	GetShiftStats(ctx context.Context, shift model.CashierShift, endTime time.Time) (*repo.ShiftStats, error)
	ListShiftTransactions(ctx context.Context, storeID uint, shift model.CashierShift, page, limit int) ([]*model.Trx, int64, error)
}
