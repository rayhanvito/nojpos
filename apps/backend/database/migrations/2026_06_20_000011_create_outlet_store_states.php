<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('outlets', function (Blueprint $table): void {
            $table->boolean('store_open_close_enabled')->default(false)->after('timezone');
        });

        Schema::create('outlet_store_states', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->unique();
            $table->string('status')->default('open')->index();
            $table->timestamp('opened_at')->nullable();
            $table->uuid('opened_by_user_id')->nullable()->index();
            $table->timestamp('closed_at')->nullable();
            $table->uuid('closed_by_user_id')->nullable()->index();
            $table->string('reason')->nullable();
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('outlet_store_states');

        Schema::table('outlets', function (Blueprint $table): void {
            $table->dropColumn('store_open_close_enabled');
        });
    }
};
