<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('payment_method_configs', function (Blueprint $table): void {
            $table->boolean('async_confirmation_enabled')->default(false)->after('is_cash');
            $table->string('provider')->nullable();
            $table->unsignedSmallInteger('async_expiry_minutes')->default(10);
        });

        Schema::table('payments', function (Blueprint $table): void {
            $table->string('provider')->nullable();
            $table->string('provider_reference')->nullable();
            $table->timestamp('confirm_expires_at')->nullable();
            $table->timestamp('failed_at')->nullable();
            $table->string('failure_reason')->nullable();
            $table->index(['business_id', 'provider_reference']);
            $table->index(['business_id', 'confirm_expires_at']);
        });

        Schema::create('payment_webhook_events', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->nullable()->index();
            $table->string('provider');
            $table->string('event_id');
            $table->uuid('payment_id')->nullable()->index();
            $table->string('outcome');
            $table->timestamp('received_at');
            $table->timestamps();
            $table->unique(['provider', 'event_id']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('payment_webhook_events');
        Schema::table('payments', function (Blueprint $table): void {
            $table->dropIndex(['business_id', 'provider_reference']);
            $table->dropIndex(['business_id', 'confirm_expires_at']);
            $table->dropColumn(['provider', 'provider_reference', 'confirm_expires_at', 'failed_at', 'failure_reason']);
        });
        Schema::table('payment_method_configs', function (Blueprint $table): void {
            $table->dropColumn(['async_confirmation_enabled', 'provider', 'async_expiry_minutes']);
        });
    }
};
