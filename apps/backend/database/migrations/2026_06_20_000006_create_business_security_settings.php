<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('business_security_settings', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->unique();
            $table->boolean('pin_required_void')->default(true);
            $table->boolean('pin_required_refund')->default(true);
            $table->boolean('pin_required_discount_override')->default(true);
            $table->boolean('pin_required_cash_out_over_limit')->default(true);
            $table->boolean('pin_required_close_shift')->default(true);
            $table->boolean('pin_required_store_open_close')->default(true);
            $table->boolean('pin_required_settings_change')->default(false);
            $table->unsignedSmallInteger('pin_lockout_max_attempts')->default(5);
            $table->unsignedSmallInteger('pin_lockout_decay_minutes')->default(15);
            $table->unsignedInteger('idle_lock_timeout_seconds')->default(180);
            $table->unsignedInteger('terminal_session_timeout_seconds')->default(900);
            $table->uuid('updated_by')->nullable()->index();
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('business_security_settings');
    }
};
