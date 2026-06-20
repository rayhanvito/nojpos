<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('outlets', function (Blueprint $table) {
            $table->string('receipt_paper_width')->default('58mm')->after('tax_rate');
            $table->string('receipt_header_name')->nullable()->after('receipt_paper_width');
            $table->string('receipt_header_address')->nullable()->after('receipt_header_name');
            $table->string('receipt_footer_note')->nullable()->after('receipt_header_address');
            $table->boolean('receipt_show_logo')->default(false)->after('receipt_footer_note');
            $table->boolean('receipt_show_qris_info')->default(true)->after('receipt_show_logo');
        });

        Schema::create('attendance_records', function (Blueprint $table) {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->index();
            $table->uuid('staff_id')->index();
            $table->uuid('created_by')->index();
            $table->timestamp('clock_in_at');
            $table->timestamp('clock_out_at')->nullable();
            $table->timestamps();
            $table->softDeletes();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('attendance_records');

        Schema::table('outlets', function (Blueprint $table) {
            $table->dropColumn([
                'receipt_paper_width',
                'receipt_header_name',
                'receipt_header_address',
                'receipt_footer_note',
                'receipt_show_logo',
                'receipt_show_qris_info',
            ]);
        });
    }
};
