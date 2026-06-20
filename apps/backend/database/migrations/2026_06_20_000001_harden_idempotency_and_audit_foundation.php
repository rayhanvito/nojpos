<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('idempotency_keys', function (Blueprint $table): void {
            $table->string('state')->default('completed')->index()->after('request_hash');
            $table->timestamp('reserved_at')->nullable()->after('state');
            $table->timestamp('completed_at')->nullable()->after('reserved_at');
        });

        Schema::table('audit_logs', function (Blueprint $table): void {
            $table->uuid('approver_id')->nullable()->index()->after('actor_id');
            $table->uuid('outlet_id')->nullable()->index()->after('approver_id');
            $table->uuid('device_id')->nullable()->index()->after('outlet_id');
            $table->string('request_id')->nullable()->index()->after('device_id');
            $table->string('idempotency_key')->nullable()->index()->after('request_id');
            $table->unsignedSmallInteger('event_version')->default(1)->after('idempotency_key');
        });
    }

    public function down(): void
    {
        Schema::table('audit_logs', function (Blueprint $table): void {
            $table->dropColumn([
                'approver_id',
                'outlet_id',
                'device_id',
                'request_id',
                'idempotency_key',
                'event_version',
            ]);
        });

        Schema::table('idempotency_keys', function (Blueprint $table): void {
            $table->dropColumn([
                'state',
                'reserved_at',
                'completed_at',
            ]);
        });
    }
};
