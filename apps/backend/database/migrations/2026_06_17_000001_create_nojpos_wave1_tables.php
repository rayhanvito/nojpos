<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('businesses', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->string('name');
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('outlets', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->string('name');
            $table->unsignedInteger('service_charge_rate')->default(0);
            $table->unsignedInteger('tax_rate')->default(0);
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('devices', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->index();
            $table->string('device_uuid')->index();
            $table->string('name');
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('roles', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->nullable()->index();
            $table->string('name')->index();
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('product_categories', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->string('name');
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('products', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->nullable()->index();
            $table->uuid('product_category_id')->nullable()->index();
            $table->string('name');
            $table->string('barcode')->nullable()->index();
            $table->unsignedBigInteger('price');
            $table->boolean('track_stock')->default(false);
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('shift_sessions', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->index();
            $table->uuid('device_id')->index();
            $table->uuid('cashier_id')->index();
            $table->string('status')->default('open')->index();
            $table->unsignedBigInteger('opening_cash')->default(0);
            $table->unsignedBigInteger('expected_cash')->nullable();
            $table->unsignedBigInteger('actual_cash')->nullable();
            $table->bigInteger('cash_difference')->nullable();
            $table->timestamp('opened_at');
            $table->timestamp('closed_at')->nullable();
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('cash_movements', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->index();
            $table->uuid('shift_id')->index();
            $table->uuid('actor_id')->index();
            $table->string('type')->index();
            $table->unsignedBigInteger('amount');
            $table->string('reason')->nullable();
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('transactions', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->index();
            $table->uuid('device_id')->index();
            $table->uuid('cashier_id')->index();
            $table->uuid('shift_id')->index();
            $table->string('number')->index();
            $table->string('status')->index();
            $table->unsignedBigInteger('subtotal')->default(0);
            $table->unsignedBigInteger('discount_total')->default(0);
            $table->unsignedBigInteger('service_charge_total')->default(0);
            $table->unsignedBigInteger('tax_total')->default(0);
            $table->bigInteger('rounding_total')->default(0);
            $table->unsignedBigInteger('grand_total')->default(0);
            $table->text('notes')->nullable();
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('transaction_items', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('transaction_id')->index();
            $table->uuid('product_id')->index();
            $table->string('name');
            $table->unsignedInteger('quantity');
            $table->unsignedBigInteger('unit_price');
            $table->unsignedBigInteger('discount')->default(0);
            $table->unsignedBigInteger('subtotal');
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('payments', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('transaction_id')->index();
            $table->string('method')->index();
            $table->string('reference')->nullable();
            $table->unsignedBigInteger('amount');
            $table->string('status')->default('pending')->index();
            $table->boolean('is_cash')->default(false);
            $table->uuid('confirmed_by')->nullable();
            $table->timestamp('confirmed_at')->nullable();
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('stock_movements', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->index();
            $table->uuid('product_id')->index();
            $table->uuid('transaction_id')->nullable()->index();
            $table->string('type')->index();
            $table->integer('quantity_delta');
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('payment_method_configs', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->nullable()->index();
            $table->string('method')->index();
            $table->boolean('is_cash')->default(false);
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('idempotency_keys', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->string('endpoint')->index();
            $table->string('key');
            $table->string('request_hash');
            $table->longText('response_snapshot');
            $table->unsignedSmallInteger('status');
            $table->timestamps();
            $table->unique(['business_id', 'endpoint', 'key']);
        });

        Schema::create('audit_logs', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->nullable()->index();
            $table->uuid('actor_id')->nullable()->index();
            $table->string('action')->index();
            $table->string('entity_type')->nullable();
            $table->uuid('entity_id')->nullable();
            $table->json('before')->nullable();
            $table->json('after')->nullable();
            $table->timestamp('created_at')->useCurrent();
        });

        Schema::create('pin_attempts', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->nullable()->index();
            $table->uuid('device_id')->nullable()->index();
            $table->string('pin_key')->nullable()->index();
            $table->unsignedInteger('attempts')->default(0);
            $table->timestamp('locked_until')->nullable();
            $table->timestamps();
        });

        Schema::create('personal_access_tokens', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->string('tokenable_type');
            $table->uuid('tokenable_id');
            $table->string('name');
            $table->string('token', 64)->unique();
            $table->text('abilities')->nullable();
            $table->timestamp('last_used_at')->nullable();
            $table->timestamp('expires_at')->nullable();
            $table->timestamps();
            $table->index(['tokenable_type', 'tokenable_id']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('personal_access_tokens');
        Schema::dropIfExists('pin_attempts');
        Schema::dropIfExists('audit_logs');
        Schema::dropIfExists('idempotency_keys');
        Schema::dropIfExists('payment_method_configs');
        Schema::dropIfExists('stock_movements');
        Schema::dropIfExists('payments');
        Schema::dropIfExists('transaction_items');
        Schema::dropIfExists('transactions');
        Schema::dropIfExists('cash_movements');
        Schema::dropIfExists('shift_sessions');
        Schema::dropIfExists('products');
        Schema::dropIfExists('product_categories');
        Schema::dropIfExists('roles');
        Schema::dropIfExists('devices');
        Schema::dropIfExists('outlets');
        Schema::dropIfExists('businesses');
    }
};
