<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('terminal_lock_states', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->index();
            $table->uuid('device_id')->index();
            $table->uuid('current_user_id')->nullable()->index();
            $table->uuid('current_cashier_id')->nullable()->index();
            $table->uuid('shift_id')->nullable()->index();
            $table->boolean('locked')->default(false)->index();
            $table->timestamp('locked_at')->nullable();
            $table->string('lock_reason')->nullable();
            $table->unsignedSmallInteger('failed_attempts')->default(0);
            $table->timestamp('lockout_until')->nullable();
            $table->timestamp('unlocked_at')->nullable();
            $table->uuid('unlocked_by_user_id')->nullable()->index();
            $table->timestamps();

            $table->unique(['business_id', 'outlet_id', 'device_id'], 'terminal_lock_state_context_unique');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('terminal_lock_states');
    }
};
