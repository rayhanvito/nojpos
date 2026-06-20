<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('transaction_quotes', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->index();
            $table->uuid('device_id')->index();
            $table->uuid('cashier_id')->index();
            $table->uuid('shift_id')->index();
            $table->uuid('customer_id')->nullable()->index();
            $table->uuid('served_by')->nullable()->index();
            $table->string('config_version');
            $table->string('quote_hash');
            $table->uuid('checkout_idempotency_key')->unique();
            $table->string('checkout_token', 64)->unique();
            $table->json('input_snapshot');
            $table->json('quote_snapshot');
            $table->timestamp('expires_at')->index();
            $table->uuid('used_transaction_id')->nullable()->index();
            $table->timestamps();

            $table->index(['business_id', 'id']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('transaction_quotes');
    }
};
