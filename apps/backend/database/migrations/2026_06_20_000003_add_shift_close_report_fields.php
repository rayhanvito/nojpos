<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('shift_sessions', function (Blueprint $table): void {
            $table->text('variance_reason')->nullable()->after('cash_difference');
            $table->uuid('approved_by')->nullable()->after('variance_reason');
            $table->json('close_report_snapshot')->nullable()->after('approved_by');
        });
    }

    public function down(): void
    {
        Schema::table('shift_sessions', function (Blueprint $table): void {
            $table->dropColumn(['variance_reason', 'approved_by', 'close_report_snapshot']);
        });
    }
};
