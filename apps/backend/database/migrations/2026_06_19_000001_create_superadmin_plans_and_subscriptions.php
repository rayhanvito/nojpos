<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('plans', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->string('name');
            $table->string('code');
            $table->string('active_code')->nullable()->unique();
            $table->unsignedBigInteger('price')->default(0);
            $table->string('billing_period')->nullable();
            $table->unsignedInteger('max_outlets')->nullable();
            $table->unsignedInteger('max_devices')->nullable();
            $table->unsignedInteger('max_users')->nullable();
            $table->unsignedInteger('max_products')->nullable();
            $table->boolean('is_active')->default(true)->index();
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('subscriptions', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('active_business_id')->nullable()->unique();
            $table->uuid('plan_id')->nullable()->index();
            $table->string('status')->index();
            $table->timestamp('trial_ends_at')->nullable();
            $table->timestamp('current_period_ends_at')->nullable();
            $table->unsignedInteger('max_outlets_override')->nullable();
            $table->unsignedInteger('max_devices_override')->nullable();
            $table->unsignedInteger('max_users_override')->nullable();
            $table->unsignedInteger('max_products_override')->nullable();
            $table->text('notes')->nullable();
            $table->uuid('created_by')->nullable();
            $table->uuid('updated_by')->nullable();
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::table('businesses', function (Blueprint $table) {
            $table->string('status')->default('active')->index();
        });
    }

    public function down(): void
    {
        Schema::table('businesses', function (Blueprint $table) {
            $table->dropColumn('status');
        });

        Schema::dropIfExists('subscriptions');
        Schema::dropIfExists('plans');
    }
};
