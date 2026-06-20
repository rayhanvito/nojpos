<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('inventory_purchases', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->index();
            $table->string('number')->index();
            $table->string('supplier_name')->nullable();
            $table->text('notes')->nullable();
            $table->unsignedBigInteger('total')->default(0);
            $table->timestamp('purchased_at');
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('inventory_purchase_items', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('purchase_id')->index();
            $table->uuid('product_id')->index();
            $table->unsignedInteger('quantity');
            $table->unsignedBigInteger('unit_cost');
            $table->unsignedBigInteger('subtotal');
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::table('stock_movements', function (Blueprint $table) {
            $table->uuid('purchase_id')->nullable()->index()->after('transaction_id');
        });
    }

    public function down(): void
    {
        Schema::table('stock_movements', function (Blueprint $table) {
            $table->dropColumn('purchase_id');
        });

        Schema::dropIfExists('inventory_purchase_items');
        Schema::dropIfExists('inventory_purchases');
    }
};
