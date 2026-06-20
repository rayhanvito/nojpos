<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('transactions', function (Blueprint $table): void {
            $table->unsignedInteger('revision')->default(1)->after('status');
            $table->uuid('lease_device_id')->nullable()->after('revision')->index();
            $table->uuid('leased_by_user_id')->nullable()->after('lease_device_id')->index();
            $table->timestamp('lease_expires_at')->nullable()->after('leased_by_user_id')->index();
        });
    }

    public function down(): void
    {
        Schema::table('transactions', function (Blueprint $table): void {
            $table->dropColumn([
                'revision',
                'lease_device_id',
                'leased_by_user_id',
                'lease_expires_at',
            ]);
        });
    }
};
