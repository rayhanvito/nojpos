package app

import (
	"context"
	"github.com/rayhanvito/nojpos/backend/internal/inventory"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/rayhanvito/nojpos/backend/internal/activity"
	"github.com/rayhanvito/nojpos/backend/internal/apikey"
	"github.com/rayhanvito/nojpos/backend/internal/auth"
	"github.com/rayhanvito/nojpos/backend/internal/catalog"
	"github.com/rayhanvito/nojpos/backend/internal/customer"
	"github.com/rayhanvito/nojpos/backend/internal/finance"
	"github.com/rayhanvito/nojpos/backend/internal/notification"
	"github.com/rayhanvito/nojpos/backend/internal/outlet"
	"github.com/rayhanvito/nojpos/backend/internal/platform/config"
	"github.com/rayhanvito/nojpos/backend/internal/platform/database"
	"github.com/rayhanvito/nojpos/backend/internal/platform/middleware"
	"github.com/rayhanvito/nojpos/backend/internal/promo"
	"github.com/rayhanvito/nojpos/backend/internal/purchasing"
	"github.com/rayhanvito/nojpos/backend/internal/report"
	"github.com/rayhanvito/nojpos/backend/internal/sales"
	"github.com/rayhanvito/nojpos/backend/internal/shift"
	"github.com/rayhanvito/nojpos/backend/internal/staff"
	"github.com/rayhanvito/nojpos/backend/internal/user"
	"github.com/rayhanvito/nojpos/backend/internal/webhook"
)

type App struct {
	Handler http.Handler
	Port    string
	Cleanup func()
}

func New(ctx context.Context) (*App, error) {
	// Penghitung rate-limit bersifat global per proses. Server produksi
	// hanya dibangun sekali sehingga reset ini tidak berpengaruh di sana;
	// pada tes integrasi setiap boot App mulai dengan penghitung bersih
	// agar limiter antar-tes tidak saling mencemari (satu IP httptest).
	middleware.ResetRateLimiters()

	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	gormDB, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	sqlDB, _ := gormDB.DB()
	cleanup := func() { _ = sqlDB.Close() }

	if database.IsSQLite(gormDB) {
		if err := database.Migrate(gormDB, autoMigrateModels...); err != nil {
			cleanup()
			return nil, err
		}
		if err := backfillDefaultOutlets(gormDB); err != nil {
			cleanup()
			return nil, err
		}
		log.Println("migrasi GORM: ok")
	} else {
		log.Println("migrasi GORM: dilewati untuk Postgres — skema dikelola migrasi goose (cmd/migrate)")
	}

	userRepo := auth.NewUserRepo(gormDB)
	cashierRepo := staff.NewRepo(gormDB)
	outletRepo := outlet.NewOutletRepo(gormDB)
	apiKeyRepo := apikey.NewRepo(gormDB)
	refreshRepo := auth.NewRefreshRepo(gormDB)
	otpRepo := auth.NewOtpRepo(gormDB)
	categoryRepo := catalog.NewCategoryRepo(gormDB)
	customerRepo := customer.NewRepo(gormDB)
	supplierRepo := purchasing.NewSupplierRepo(gormDB)
	expenseRepo := finance.NewExpenseRepo(gormDB)
	purchaseOrderRepo := purchasing.NewPurchaseOrderRepo(gormDB)
	stockOpnameRepo := inventory.NewStockOpnameRepo(gormDB)
	debtRepo := finance.NewDebtRepo(gormDB)
	productRepo := catalog.NewProductRepo(gormDB)
	movementRepo := inventory.NewMovementRepo(gormDB)
	transferRepo := inventory.NewTransferRepo(gormDB)
	trxRepo := sales.NewRepo(gormDB, database.NewTransactor(gormDB))
	storeRepo := outlet.NewStoreRepo(gormDB)
	reportRepo := report.NewRepo(gormDB)
	shiftRepo := shift.NewRepo(gormDB)
	notificationRepo := notification.NewRepo(gormDB)
	activityRepo := activity.NewRepo(gormDB)
	promoRepo := promo.NewRepo(gormDB)
	webhookRepo := webhook.NewRepo(gormDB)

	authSvc := auth.NewService(userRepo, cashierRepo, refreshRepo, otpRepo, cfg.JWTSecret, cfg.AccessTTL, time.Duration(cfg.RefreshTTLDays)*24*time.Hour, cfg.GoogleClientID)
	authSvc.SetOutlets(outletRepo)
	apiKeySvc := apikey.NewService(apiKeyRepo, outletRepo)
	userSvc := user.NewService(userRepo, cashierRepo, outletRepo)
	userSvc.SetShiftRepository(shiftRepo)
	outletSvc := outlet.NewService(outletRepo)
	notificationSvc := notification.NewService(notificationRepo, storeRepo)
	activitySvc := activity.NewService(activityRepo)
	promoSvc := promo.NewService(promoRepo, categoryRepo, productRepo)
	authSvc.SetActivity(activitySvc)
	stockEngine := inventory.NewEngine(gormDB)
	inventorySvc := inventory.NewService(productRepo, movementRepo, notificationSvc)
	inventorySvc.SetStockEngine(stockEngine, database.NewTransactor(gormDB))
	inventorySvc.SetTransferRepo(transferRepo)
	catalogSvc := catalog.NewService(categoryRepo, productRepo, notificationSvc)
	catalogSvc.SetStockEngine(stockEngine, database.NewTransactor(gormDB))
	catalogSvc.SetInventory(inventorySvc)
	customerSvc := customer.NewService(customerRepo)
	supplierSvc := purchasing.NewSupplierService(supplierRepo)
	expenseSvc := finance.NewExpenseService(expenseRepo, cashierRepo, shiftRepo)
	purchaseOrderSvc := purchasing.NewService(purchaseOrderRepo, supplierRepo, productRepo, storeRepo, database.NewTransactor(gormDB))
	stockOpnameSvc := inventory.NewOpnameService(stockOpnameRepo)
	debtSvc := finance.NewService(debtRepo, storeRepo)
	trxSvc := sales.NewService(trxRepo, cashierRepo, productRepo, shiftRepo, notificationSvc, database.NewTransactor(gormDB))
	settingsSvc := outlet.NewSettingsService(storeRepo)
	// Unit C: checkout memakai tarif pajak efektif outlet transaksi
	// (override per outlet menimpa tarif toko bila ditetapkan).
	trxSvc.SetSettingsResolver(settingsSvc)
	// Unit F: webhook adalah sink seperti notifikasi — dispatcher-nya
	// dipasang ke domain pemancar event lewat setter; pengiriman asinkron
	// dan tidak pernah menggagalkan operasi bisnis.
	webhookSvc := webhook.NewService(webhookRepo)
	webhookDispatcher := webhook.NewDispatcher(webhookRepo)
	trxSvc.SetWebhookEmitter(webhookDispatcher)
	inventorySvc.SetWebhookEmitter(webhookDispatcher)
	stockOpnameSvc.SetWebhookEmitter(webhookDispatcher)
	reportSvc := report.NewService(storeRepo, reportRepo)
	exportSvc := report.NewExportService(storeRepo, trxRepo, productRepo)
	taxReportSvc := report.NewTaxService(storeRepo, reportRepo)
	shiftSvc := shift.NewService(shiftRepo, cashierRepo, storeRepo)

	authH := auth.NewHandler(authSvc)
	userH := user.NewHandler(userSvc, activitySvc)
	outletH := outlet.NewHandler(outletSvc)
	catalogH := catalog.NewHandler(catalogSvc)
	customerH := customer.NewHandler(customerSvc)
	supplierH := purchasing.NewSupplierHandler(supplierSvc)
	expenseH := finance.NewExpenseHandler(expenseSvc)
	purchaseOrderH := purchasing.NewHandler(purchaseOrderSvc)
	stockOpnameH := inventory.NewOpnameHandler(stockOpnameSvc)
	debtH := finance.NewHandler(debtSvc)
	stockH := inventory.NewHandler(inventorySvc)
	trxH := sales.NewHandler(trxSvc)
	settingsH := outlet.NewSettingsHandler(settingsSvc, activitySvc)
	reportH := report.NewHandler(reportSvc)
	exportH := report.NewExportHandler(exportSvc)
	taxReportH := report.NewTaxHandler(taxReportSvc)
	shiftH := shift.NewHandler(shiftSvc)
	notificationH := notification.NewHandler(notificationSvc)
	activityH := activity.NewHandler(activitySvc)
	apiKeyH := apikey.NewHandler(apiKeySvc)
	webhookH := webhook.NewHandler(webhookSvc)
	promoH := promo.NewHandler(promoSvc)
	healthH := NewHealthHandler(gormDB)

	r := gin.New()
	// Hanya percaya X-Forwarded-For dari reverse proxy setempat (nginx di host
	// yang sama). Tanpa ini Gin percaya semua proxy dan ClientIP() — kunci rate
	// limiter — bisa dipalsukan dari internet (arsitektur BE §11 langkah 3).
	if err := r.SetTrustedProxies([]string{"127.0.0.1", "::1"}); err != nil {
		panic(err)
	}
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.RequestID())
	r.Use(middleware.SecurityHeaders())
	// 30s request timeout like old chi backend (chimw.Timeout).
	// Literal bug: slow queries could hang forever.
	r.Use(func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	})
	r.Use(cors.New(cors.Config{
		AllowOrigins:     cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-Api-Key"},
		AllowCredentials: false,
		// Fixed: was 300 (300ns) — cors.Config expects time.Duration.
		MaxAge: 300 * time.Second,
	}))

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthH.Health)

		// Public
		auth.RegisterPublicRoutes(v1, authH)

		// Authenticated
		authGroup := v1.Group("")
		authGroup.Use(middleware.Auth(authSvc, apiKeySvc), middleware.ActiveCheck(authSvc), middleware.RestrictAPIKeyRoutes())
		{
			auth.RegisterAuthRoutes(authGroup, authH)

			// Admin only
			adminGroup := authGroup.Group("")
			adminGroup.Use(middleware.RequireRole(string(middleware.RoleAdmin)))
			{
				apikey.RegisterRoutes(adminGroup, apiKeyH)
				webhook.RegisterRoutes(adminGroup, webhookH)
				notification.RegisterRoutes(adminGroup, notificationH)
				user.RegisterRoutes(adminGroup, userH)
				outlet.RegisterRoutes(authGroup, adminGroup, outletH, settingsH)
				catalog.RegisterRoutes(authGroup, adminGroup, catalogH)
				inventory.RegisterRoutes(authGroup, adminGroup, stockH, stockOpnameH)
				finance.RegisterRoutes(authGroup, adminGroup, debtH, expenseH)
				purchasing.RegisterRoutes(authGroup, adminGroup, purchaseOrderH, supplierH)
				customer.RegisterRoutes(authGroup, adminGroup, customerH)
				sales.RegisterRoutes(authGroup, adminGroup, trxH)
				promo.RegisterRoutes(adminGroup, promoH)
				report.RegisterRoutes(authGroup, adminGroup, reportH, taxReportH, exportH)
				shift.RegisterRoutes(authGroup, adminGroup, shiftH)
				activity.RegisterRoutes(adminGroup, activityH)
			}
		}
	}

	return &App{Handler: r, Port: cfg.Port, Cleanup: cleanup}, nil
}
