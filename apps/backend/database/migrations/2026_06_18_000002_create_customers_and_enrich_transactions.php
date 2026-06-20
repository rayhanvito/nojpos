<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('customers', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->string('name')->index();
            $table->string('phone')->nullable()->index();
            $table->string('group')->default('Tanpa Grup')->index();
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::table('transactions', function (Blueprint $table) {
            $table->uuid('customer_id')->nullable()->index()->after('shift_id');
            $table->uuid('served_by')->nullable()->index()->after('customer_id');
            $table->unsignedBigInteger('item_discount_total')->default(0)->after('discount_total');
            $table->unsignedBigInteger('cart_discount_total')->default(0)->after('item_discount_total');
        });
    }

    public function down(): void
    {
        Schema::table('transactions', function (Blueprint $table) {
            $table->dropColumn([
                'customer_id',
                'served_by',
                'item_discount_total',
                'cart_discount_total',
            ]);
        });

        Schema::dropIfExists('customers');
    }
};
