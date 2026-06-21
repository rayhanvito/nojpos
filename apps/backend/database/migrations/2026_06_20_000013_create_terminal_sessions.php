<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('terminal_sessions', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->index();
            $table->uuid('device_id')->index();
            $table->uuid('user_id')->index();
            $table->uuid('cashier_id')->index();
            $table->uuid('token_id')->nullable()->index();
            $table->string('status')->default('active')->index();
            $table->timestamp('opened_at')->index();
            $table->timestamp('last_seen_at')->nullable();
            $table->timestamp('expires_at')->nullable()->index();
            $table->timestamp('locked_at')->nullable();
            $table->timestamp('revoked_at')->nullable();
            $table->json('metadata')->nullable();
            $table->timestamps();

            $table->index(['business_id', 'outlet_id', 'device_id', 'status'], 'terminal_sessions_context_status_idx');
            $table->index(['business_id', 'token_id', 'status'], 'terminal_sessions_token_status_idx');
            $table->index(['business_id', 'user_id', 'status'], 'terminal_sessions_user_status_idx');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('terminal_sessions');
    }
};
