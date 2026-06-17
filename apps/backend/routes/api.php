<?php

use App\Http\Controllers\Api\V1\AuthController;
use App\Http\Controllers\Api\V1\PaymentController;
use App\Http\Controllers\Api\V1\ProductController;
use App\Http\Controllers\Api\V1\ShiftController;
use App\Http\Controllers\Api\V1\TransactionController;
use Illuminate\Support\Facades\Route;

Route::prefix('v1')->group(function (): void {
    Route::post('/auth/login', [AuthController::class, 'login']);

    Route::middleware('auth:sanctum')->group(function (): void {
        Route::post('/auth/pin-switch', [AuthController::class, 'pinSwitch']);
        Route::post('/auth/logout', [AuthController::class, 'logout']);
        Route::get('/me', [AuthController::class, 'me']);
        Route::get('/outlets', [AuthController::class, 'outlets']);

        Route::get('/products', [ProductController::class, 'index']);

        Route::post('/shifts/open', [ShiftController::class, 'open']);
        Route::get('/shifts/current', [ShiftController::class, 'current']);
        Route::post('/shifts/{shift}/cash-movements', [ShiftController::class, 'cashMovement']);
        Route::post('/shifts/{shift}/close', [ShiftController::class, 'close'])->middleware('idempotency');

        Route::post('/transactions', [TransactionController::class, 'store'])->middleware('idempotency');
        Route::get('/transactions', [TransactionController::class, 'index']);
        Route::post('/payments', [PaymentController::class, 'store'])->middleware('idempotency');
    });
});
