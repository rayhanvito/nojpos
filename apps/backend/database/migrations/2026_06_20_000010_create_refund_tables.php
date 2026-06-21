<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('refunds', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->index();
            $table->uuid('transaction_id')->index();
            $table->uuid('shift_id')->nullable()->index();
            $table->uuid('actor_id')->index();
            $table->uuid('authorizer_id')->nullable()->index();
            $table->string('refund_method')->index();
            $table->unsignedBigInteger('total_amount');
            $table->string('reason');
            $table->text('notes')->nullable();
            $table->string('status')->default('finalized')->index();
            $table->string('idempotency_key')->nullable()->index();
            $table->timestamp('finalized_at');
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('refund_lines', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('refund_id')->index();
            $table->uuid('transaction_item_id')->index();
            $table->uuid('product_id')->index();
            $table->unsignedInteger('quantity');
            $table->unsignedBigInteger('amount');
            $table->boolean('restock')->default(true);
            $table->string('non_restock_reason')->nullable();
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::table('stock_movements', function (Blueprint $table): void {
            $table->uuid('refund_id')->nullable()->index();
        });

        Schema::table('cash_movements', function (Blueprint $table): void {
            $table->uuid('refund_id')->nullable()->index();
        });
    }

    public function down(): void
    {
        Schema::table('cash_movements', function (Blueprint $table): void {
            $table->dropColumn('refund_id');
        });

        Schema::table('stock_movements', function (Blueprint $table): void {
            $table->dropColumn('refund_id');
        });

        Schema::dropIfExists('refund_lines');
        Schema::dropIfExists('refunds');
    }
};
