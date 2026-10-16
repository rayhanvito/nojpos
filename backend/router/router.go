package router

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/adrr-dev/openPOS/backend/config"
	"github.com/adrr-dev/openPOS/backend/db"
	"github.com/adrr-dev/openPOS/backend/handler"
	"github.com/adrr-dev/openPOS/backend/middleware"
	"github.com/adrr-dev/openPOS/backend/model"
	"github.com/adrr-dev/openPOS/backend/repo"
	"github.com/adrr-dev/openPOS/backend/service"
)

type Server struct {
	Handler http.Handler
	Port    string
	Cleanup func()
}

func New(ctx context.Context) (*Server, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	database, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	sqlDB, _ := database.DB()
	cleanup := func() { _ = sqlDB.Close() }

	if err := db.Migrate(database); err != nil {
		cleanup()
		return nil, err
	}
	log.Println("migrasi GORM: ok")

	userRepo := repo.NewUserRepo(database)
	cashierRepo := repo.NewCashierRepo(database)
	outletRepo := repo.NewOutletRepo(database)
	refreshRepo := repo.NewRefreshRepo(database)
	otpRepo := repo.NewOtpRepo(database)
	categoryRepo := repo.NewCategoryRepo(database)
	customerRepo := repo.NewCustomerRepo(database)
	supplierRepo := repo.NewSupplierRepo(database)
	expenseRepo := repo.NewExpenseRepo(database)
	purchaseOrderRepo := repo.NewPurchaseOrderRepo(database)
	stockOpnameRepo := repo.NewStockOpnameRepo(database)
	debtRepo := repo.NewDebtRepo(database)
	productRepo := repo.NewProductRepo(database)
	movementRepo := repo.NewMovementRepo(database)
	trxRepo := repo.NewTrxRepo(database)
	storeRepo := repo.NewStoreRepo(database)
	reportRepo := repo.NewReportRepo(database)
	shiftRepo := repo.NewShiftRepo(database)
	notificationRepo := repo.NewNotificationRepo(database)
	activityRepo := repo.NewActivityRepo(database)
	promoRepo := repo.NewPromoRepo(database)

	authSvc := service.NewAuthService(userRepo, cashierRepo, refreshRepo, otpRepo, cfg.JWTSecret, cfg.AccessTTL, time.Duration(cfg.RefreshTTLDays)*24*time.Hour, cfg.GoogleClientID)
	authSvc.SetOutlets(outletRepo)
	userSvc := service.NewUserService(userRepo, cashierRepo, outletRepo)
	userSvc.SetShiftRepository(shiftRepo)
	outletSvc := service.NewOutletService(outletRepo)
	notificationSvc := service.NewNotificationService(notificationRepo, storeRepo)
	activitySvc := service.NewActivityService(activityRepo)
	promoSvc := service.NewPromoService(promoRepo, categoryRepo, productRepo)
	authSvc.SetActivity(activitySvc)
	catalogSvc := service.NewCatalogService(categoryRepo, productRepo, movementRepo, notificationSvc)
	customerSvc := service.NewCustomerService(customerRepo)
	supplierSvc := service.NewSupplierService(supplierRepo)
	expenseSvc := service.NewExpenseService(expenseRepo, cashierRepo, shiftRepo)
	purchaseOrderSvc := service.NewPurchaseOrderService(purchaseOrderRepo, supplierRepo, productRepo, storeRepo)
	stockOpnameSvc := service.NewStockOpnameService(stockOpnameRepo)
	debtSvc := service.NewDebtService(debtRepo, storeRepo)
	trxSvc := service.NewTrxService(trxRepo, cashierRepo, productRepo, shiftRepo, notificationSvc)
	settingsSvc := service.NewSettingsService(storeRepo, userRepo, cashierRepo, reportRepo)
	exportSvc := service.NewExportService(storeRepo, trxRepo, productRepo)
	taxReportSvc := service.NewTaxReportService(storeRepo, reportRepo)
	shiftSvc := service.NewShiftService(shiftRepo, cashierRepo, storeRepo)

	authH := handler.NewAuthHandler(authSvc)
	userH := handler.NewUserHandler(userSvc, activitySvc)
	outletH := handler.NewOutletHandler(outletSvc, authSvc)
	catalogH := handler.NewCatalogHandler(catalogSvc)
	customerH := handler.NewCustomerHandler(customerSvc)
	supplierH := handler.NewSupplierHandler(supplierSvc)
	expenseH := handler.NewExpenseHandler(expenseSvc)
	purchaseOrderH := handler.NewPurchaseOrderHandler(purchaseOrderSvc)
	stockOpnameH := handler.NewStockOpnameHandler(stockOpnameSvc)
	debtH := handler.NewDebtHandler(debtSvc)
	stockH := handler.NewStockHandler(catalogSvc)
	trxH := handler.NewTrxHandler(trxSvc)
	settingsH := handler.NewSettingsHandler(settingsSvc, activitySvc)
	exportH := handler.NewExportHandler(exportSvc)
	taxReportH := handler.NewTaxReportHandler(taxReportSvc)
	shiftH := handler.NewShiftHandler(shiftSvc)
	notificationH := handler.NewNotificationHandler(notificationSvc)
	activityH := handler.NewActivityHandler(activitySvc)
	promoH := handler.NewPromoHandler(promoSvc)
	presenceH := handler.NewPresenceHandler(authSvc)
	healthH := handler.NewHealthHandler(database)

	r := gin.New()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
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
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		// Fixed: was 300 (300ns) — cors.Config expects time.Duration.
		MaxAge: 300 * time.Second,
	}))

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", healthH.Health)

		// Public
		v1.POST("/auth/register", authH.Register)
		v1.POST("/auth/login", middleware.RateLimitLogin(), authH.Login)
		v1.POST("/auth/google", authH.Google)
		v1.POST("/auth/refresh", authH.Refresh)
		v1.POST("/auth/logout", authH.Logout)
		v1.POST("/auth/otp/send", authH.SendOTP)
		v1.POST("/auth/otp/verify", authH.VerifyOTP)
		v1.POST("/auth/forgot-password/send", middleware.RateLimitForgot(), authH.SendPasswordResetOTP)
		v1.POST("/auth/forgot-password/reset", authH.ResetPassword)

		// Authenticated
		authGroup := v1.Group("")
		authGroup.Use(middleware.Auth(authSvc), middleware.ActiveCheck(authSvc))
		{
			authGroup.POST("/presence/heartbeat", presenceH.Heartbeat)
			authGroup.GET("/auth/me", authH.Me)
			authGroup.GET("/outlets", outletH.List)
			authGroup.POST("/outlets/switch", outletH.Switch)
			authGroup.PUT("/auth/password", authH.ChangePassword)
			authGroup.POST("/auth/switch", middleware.RateLimitSwitch(), authH.Switch)

			authGroup.GET("/categories", catalogH.ListCategories)
			authGroup.GET("/products", catalogH.ListProducts)
			authGroup.GET("/products/:id", catalogH.GetProduct)
			authGroup.GET("/products/:id/variants", catalogH.ListVariants)
			authGroup.GET("/products/:id/price-tiers", catalogH.ListPriceTiers)
			authGroup.GET("/products/:id/variants/:variantId/price-tiers", catalogH.ListPriceTiers)
			authGroup.GET("/customers", customerH.List)
			authGroup.GET("/suppliers", supplierH.List)
			authGroup.GET("/suppliers/:id", supplierH.Get)
			authGroup.GET("/expenses", expenseH.List)
			authGroup.POST("/expenses", expenseH.Create)
			authGroup.GET("/purchase-orders", purchaseOrderH.List)
			authGroup.GET("/purchase-orders/:id", purchaseOrderH.Get)
			authGroup.GET("/stock-opnames", stockOpnameH.List)
			authGroup.POST("/stock-opnames", stockOpnameH.Create)
			authGroup.GET("/stock-opnames/:id", stockOpnameH.Get)
			authGroup.PUT("/stock-opnames/:id/items", stockOpnameH.UpdateItems)
			authGroup.POST("/stock-opnames/:id/post", stockOpnameH.Post)
			authGroup.POST("/stock-opnames/:id/cancel", stockOpnameH.Cancel)
			authGroup.GET("/customers/:id/transactions", customerH.Transactions)
			authGroup.GET("/customers/:id", customerH.Get)
			authGroup.POST("/customers", customerH.Create)
			authGroup.POST("/promos/preview", trxH.PreviewPromo)
			authGroup.POST("/transactions", trxH.Checkout)
			authGroup.GET("/transactions", trxH.List)
			authGroup.GET("/transactions/:id", trxH.Get)
			authGroup.GET("/settings", settingsH.Get)
			authGroup.GET("/dashboard", settingsH.Dashboard)
			authGroup.GET("/cashier/shift", shiftH.GetCurrentShift)
			authGroup.POST("/cashier/shift/start", shiftH.StartShift)
			authGroup.POST("/cashier/shift/close", shiftH.CloseShift)
			authGroup.GET("/notifications", notificationH.List)
			authGroup.GET("/notifications/unread-count", notificationH.UnreadCount)
			authGroup.PATCH("/notifications/read-all", notificationH.MarkAllAsRead)
			authGroup.PATCH("/notifications/:id/read", notificationH.MarkAsRead)
			authGroup.DELETE("/notifications/:id", notificationH.Delete)

			// Admin only
			adminGroup := authGroup.Group("")
			adminGroup.Use(middleware.RequireRole(string(model.RoleAdmin)))
			{
				adminGroup.GET("/users", userH.List)
				adminGroup.POST("/users", userH.Create)
				adminGroup.PATCH("/users/:id", userH.Rename)
				adminGroup.PATCH("/users/:id/active", userH.SetActive)
				adminGroup.DELETE("/users/:id", userH.Delete)
				adminGroup.POST("/outlets", outletH.Create)
				adminGroup.PUT("/outlets/:id", outletH.Update)
				adminGroup.POST("/categories", catalogH.CreateCategory)
				adminGroup.DELETE("/categories/:id", catalogH.DeleteCategory)
				adminGroup.POST("/suppliers", supplierH.Create)
				adminGroup.PUT("/suppliers/:id", supplierH.Update)
				adminGroup.DELETE("/suppliers/:id", supplierH.Delete)
				adminGroup.PUT("/expenses/:id", expenseH.Update)
				adminGroup.DELETE("/expenses/:id", expenseH.Delete)
				adminGroup.POST("/purchase-orders", purchaseOrderH.Create)
				adminGroup.PUT("/purchase-orders/:id", purchaseOrderH.Update)
				adminGroup.DELETE("/purchase-orders/:id", purchaseOrderH.Delete)
				adminGroup.POST("/purchase-orders/:id/order", purchaseOrderH.Order)
				adminGroup.POST("/purchase-orders/:id/receive", purchaseOrderH.Receive)
				adminGroup.POST("/purchase-orders/:id/cancel", purchaseOrderH.Cancel)
				adminGroup.GET("/receivables", debtH.ListReceivables)
				adminGroup.GET("/receivables/:id", debtH.GetReceivable)
				adminGroup.POST("/receivables/:id/payments", debtH.PayReceivable)
				adminGroup.GET("/payables", debtH.ListPayables)
				adminGroup.GET("/payables/:id", debtH.GetPayable)
				adminGroup.POST("/payables/:id/payments", debtH.PayPayable)
				adminGroup.PUT("/customers/:id", customerH.Update)
				adminGroup.DELETE("/customers/:id", customerH.Delete)
				adminGroup.POST("/products", catalogH.CreateProduct)
				adminGroup.POST("/products/import", catalogH.ImportProducts)
				adminGroup.PUT("/products/:id", catalogH.UpdateProduct)
				adminGroup.POST("/products/:id/variants", catalogH.CreateVariant)
				adminGroup.PUT("/products/:id/variants/:variantId", catalogH.UpdateVariant)
				adminGroup.DELETE("/products/:id/variants/:variantId", catalogH.DeleteVariant)
				adminGroup.POST("/products/:id/price-tiers", catalogH.CreatePriceTier)
				adminGroup.PUT("/products/:id/price-tiers/:tierId", catalogH.UpdatePriceTier)
				adminGroup.DELETE("/products/:id/price-tiers/:tierId", catalogH.DeletePriceTier)
				adminGroup.POST("/products/:id/variants/:variantId/price-tiers", catalogH.CreatePriceTier)
				adminGroup.PUT("/products/:id/variants/:variantId/price-tiers/:tierId", catalogH.UpdatePriceTier)
				adminGroup.DELETE("/products/:id/variants/:variantId/price-tiers/:tierId", catalogH.DeletePriceTier)
				adminGroup.PATCH("/products/:id/active", catalogH.SetProductActive)
				adminGroup.DELETE("/products/:id", catalogH.DeleteProduct)
				adminGroup.GET("/promos", promoH.List)
				adminGroup.POST("/promos", promoH.Create)
				adminGroup.GET("/promos/:id", promoH.Get)
				adminGroup.PUT("/promos/:id", promoH.Update)
				adminGroup.PATCH("/promos/:id/active", promoH.SetActive)
				adminGroup.DELETE("/promos/:id", promoH.Delete)
				adminGroup.GET("/movements", stockH.ListMovements)
				adminGroup.POST("/stock/adjustments", stockH.AdjustStock)
				adminGroup.POST("/transactions/:id/refund", trxH.Refund)
				adminGroup.PUT("/settings", settingsH.Update)
				adminGroup.PUT("/users/:id/passcode", settingsH.SetPasscode)
				adminGroup.GET("/reports", settingsH.Report)
				adminGroup.GET("/reports/tax", taxReportH.Get)
				adminGroup.GET("/reports/tax/export", taxReportH.Export)
				adminGroup.GET("/transactions/export", exportH.Transactions)
				adminGroup.GET("/products/export", exportH.Products)
				adminGroup.GET("/shifts/export", shiftH.ExportShifts)
				adminGroup.GET("/shifts", shiftH.ListShifts)
				adminGroup.GET("/shifts/:id", shiftH.GetShift)
				adminGroup.POST("/shifts/:id/force-close", shiftH.ForceCloseShift)
				adminGroup.GET("/activity", activityH.List)
			}
		}
	}

	return &Server{Handler: r, Port: cfg.Port, Cleanup: cleanup}, nil
}
