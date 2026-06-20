<?php

use App\Http\Controllers\Api\V1\AttendanceController;
use App\Http\Controllers\Api\V1\AuthController;
use App\Http\Controllers\Api\V1\CategoryController;
use App\Http\Controllers\Api\V1\CustomerController;
use App\Http\Controllers\Api\V1\InventoryController;
use App\Http\Controllers\Api\V1\ParkedOrderController;
use App\Http\Controllers\Api\V1\PaymentController;
use App\Http\Controllers\Api\V1\PaymentWebhookController;
use App\Http\Controllers\Api\V1\ProductController;
use App\Http\Controllers\Api\V1\ReportController;
use App\Http\Controllers\Api\V1\SettingsController;
use App\Http\Controllers\Api\V1\ShiftController;
use App\Http\Controllers\Api\V1\StaffController;
use App\Http\Controllers\Api\V1\SubscriptionController;
use App\Http\Controllers\Api\V1\Superadmin\BusinessController as SuperadminBusinessController;
use App\Http\Controllers\Api\V1\Superadmin\PlanController as SuperadminPlanController;
use App\Http\Controllers\Api\V1\Superadmin\SubscriptionController as SuperadminSubscriptionController;
use App\Http\Controllers\Api\V1\Superadmin\SummaryController as SuperadminSummaryController;
use App\Http\Controllers\Api\V1\TransactionController;
use App\Http\Controllers\Api\V1\VoidController;
use Illuminate\Support\Facades\Route;

Route::prefix('v1')->group(function (): void {
    Route::post('/auth/login', [AuthController::class, 'login']);
    Route::post('/payment-webhooks/{provider}', [PaymentWebhookController::class, 'store']);

    Route::middleware('auth:sanctum')->group(function (): void {
        Route::prefix('superadmin')->middleware('role:superadmin')->group(function (): void {
            Route::get('/summary', SuperadminSummaryController::class);
            Route::get('/plans', [SuperadminPlanController::class, 'index']);
            Route::post('/plans', [SuperadminPlanController::class, 'store']);
            Route::put('/plans/{plan}', [SuperadminPlanController::class, 'update']);
            Route::patch('/plans/{plan}/status', [SuperadminPlanController::class, 'status']);
            Route::get('/businesses', [SuperadminBusinessController::class, 'index']);
            Route::post('/businesses', [SuperadminBusinessController::class, 'store'])->middleware('idempotency');
            Route::get('/businesses/{business}', [SuperadminBusinessController::class, 'show']);
            Route::patch('/businesses/{business}', [SuperadminBusinessController::class, 'update']);
            Route::get('/businesses/{business}/subscription', [SuperadminSubscriptionController::class, 'show']);
            Route::post('/businesses/{business}/subscription', [SuperadminSubscriptionController::class, 'store'])->middleware('idempotency');
        });

        Route::post('/auth/pin-switch', [AuthController::class, 'pinSwitch']);
        Route::post('/auth/logout', [AuthController::class, 'logout']);
        Route::get('/me', [AuthController::class, 'me']);
        Route::get('/outlets', [AuthController::class, 'outlets']);
        Route::get('/subscription', [SubscriptionController::class, 'show']);
        Route::get('/settings', [SettingsController::class, 'index']);
        Route::middleware('role:owner,admin')->group(function (): void {
            Route::patch('/settings/business', [SettingsController::class, 'updateBusiness'])->middleware('idempotency');
            Route::patch('/settings/outlets/{outlet}', [SettingsController::class, 'updateOutlet'])->whereUuid('outlet')->middleware('idempotency');
            Route::patch('/settings/security', [SettingsController::class, 'updateSecurity'])->middleware('idempotency');
            Route::patch('/settings/payment-methods/{config}', [SettingsController::class, 'updatePaymentMethod'])->whereUuid('config')->middleware('idempotency');
        });

        Route::get('/products', [ProductController::class, 'index']);
        Route::post('/products', [ProductController::class, 'store']);
        Route::put('/products/{product}', [ProductController::class, 'update']);
        Route::get('/categories', [CategoryController::class, 'index']);
        Route::post('/categories', [CategoryController::class, 'store']);
        Route::put('/categories/{category}', [CategoryController::class, 'update']);
        Route::get('/customers', [CustomerController::class, 'index']);
        Route::post('/customers', [CustomerController::class, 'store']);
        Route::get('/staff', [StaffController::class, 'index']);
        Route::middleware('role:owner,admin')->group(function (): void {
            Route::post('/staff', [StaffController::class, 'store']);
            Route::put('/staff/{staff}', [StaffController::class, 'update']);
            Route::delete('/staff/{staff}', [StaffController::class, 'destroy']);
            Route::post('/staff/{staff}/delete', [StaffController::class, 'destroy']);
        });
        Route::get('/attendance', [AttendanceController::class, 'index']);
        Route::post('/attendance', [AttendanceController::class, 'store']);
        Route::get('/inventory', [InventoryController::class, 'index']);
        Route::get('/inventory/movements', [InventoryController::class, 'movements']);
        Route::get('/inventory/transfers', [InventoryController::class, 'transfers']);
        Route::get('/inventory/transfers/in-transit', [InventoryController::class, 'inTransit']);
        Route::middleware('role:owner,admin')->group(function (): void {
            Route::post('/inventory/purchases', [InventoryController::class, 'purchase']);
            Route::post('/inventory/counts', [InventoryController::class, 'counts'])->middleware('idempotency');
            Route::post('/inventory/waste', [InventoryController::class, 'waste'])->middleware('idempotency');
            Route::post('/inventory/transfers', [InventoryController::class, 'createTransfer'])->middleware('idempotency');
            Route::post('/inventory/transfers/{transfer}/send', [InventoryController::class, 'sendTransfer'])->whereUuid('transfer')->middleware('idempotency');
            Route::post('/inventory/transfers/{transfer}/receive', [InventoryController::class, 'receiveTransfer'])->whereUuid('transfer')->middleware('idempotency');
            Route::post('/inventory/transfers/{transfer}/cancel', [InventoryController::class, 'cancelTransfer'])->whereUuid('transfer')->middleware('idempotency');
        });

        Route::post('/shifts/open', [ShiftController::class, 'open']);
        Route::get('/shifts/current', [ShiftController::class, 'current']);
        Route::post('/shifts/{shift}/cash-movements', [ShiftController::class, 'cashMovement'])->middleware('idempotency');
        Route::post('/shifts/{shift}/close', [ShiftController::class, 'close'])->middleware('idempotency');

        Route::get('/parked-orders', [ParkedOrderController::class, 'index']);
        Route::post('/parked-orders', [ParkedOrderController::class, 'store']);
        Route::get('/parked-orders/{transaction}', [ParkedOrderController::class, 'show'])->whereUuid('transaction');
        Route::put('/parked-orders/{transaction}', [ParkedOrderController::class, 'update'])->whereUuid('transaction');
        Route::delete('/parked-orders/{transaction}', [ParkedOrderController::class, 'destroy'])->whereUuid('transaction');
        Route::post('/parked-orders/{transaction}/lease/acquire', [ParkedOrderController::class, 'acquireLease'])->whereUuid('transaction');
        Route::post('/parked-orders/{transaction}/lease/refresh', [ParkedOrderController::class, 'refreshLease'])->whereUuid('transaction');
        Route::post('/parked-orders/{transaction}/lease/release', [ParkedOrderController::class, 'releaseLease'])->whereUuid('transaction');

        Route::post('/transactions/quote', [TransactionController::class, 'quote']);
        Route::get('/transactions/recovery/{key}', [TransactionController::class, 'recovery'])->whereUuid('key');
        Route::get('/transactions/{transaction}', [TransactionController::class, 'show'])->whereUuid('transaction');
        Route::post('/transactions', [TransactionController::class, 'store'])->middleware('idempotency');
        Route::get('/transactions', [TransactionController::class, 'index']);
        Route::post('/payments', [PaymentController::class, 'store'])->middleware('idempotency');
        Route::post('/voids', [VoidController::class, 'store'])->middleware('idempotency');
        Route::get('/reports/sales-summary', [ReportController::class, 'salesSummary']);
        Route::get('/reports/sold-products', [ReportController::class, 'soldProducts']);
        Route::get('/reports/payment-methods', [ReportController::class, 'paymentMethods']);
        Route::get('/reports/cashier-shifts', [ReportController::class, 'cashierShifts']);
        Route::get('/reports/void-refund-audit', [ReportController::class, 'voidRefundAudit']);
        Route::get('/reports/top-10', [ReportController::class, 'topTen']);
    });
});
