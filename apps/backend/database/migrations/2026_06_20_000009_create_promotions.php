<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('promotions', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->string('name');
            $table->string('code')->nullable();
            $table->string('type')->index();
            $table->unsignedBigInteger('value');
            $table->unsignedBigInteger('minimum_subtotal')->default(0);
            $table->uuid('product_id')->nullable()->index();
            $table->uuid('product_category_id')->nullable()->index();
            $table->boolean('is_active')->default(true)->index();
            $table->boolean('stackable')->default(false);
            $table->timestamp('starts_at')->nullable();
            $table->timestamp('ends_at')->nullable();
            $table->uuid('updated_by')->nullable()->index();
            $table->timestamps();
            $table->softDeletes();

            $table->index(['business_id', 'code']);
        });

        Schema::create('promotion_outlets', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('promotion_id')->index();
            $table->uuid('outlet_id')->index();
            $table->timestamps();

            $table->unique(['promotion_id', 'outlet_id']);
        });

        Schema::table('transactions', function (Blueprint $table): void {
            $table->unsignedBigInteger('promotion_discount_total')->default(0)->after('cart_discount_total');
        });
    }

    public function down(): void
    {
        Schema::table('transactions', function (Blueprint $table): void {
            $table->dropColumn('promotion_discount_total');
        });

        Schema::dropIfExists('promotion_outlets');
        Schema::dropIfExists('promotions');
    }
};
