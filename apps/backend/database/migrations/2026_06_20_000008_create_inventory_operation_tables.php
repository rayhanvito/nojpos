<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('inventory_counts', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->index();
            $table->uuid('actor_id')->nullable()->index();
            $table->string('number')->index();
            $table->string('status')->default('completed')->index();
            $table->text('notes')->nullable();
            $table->timestamp('counted_at');
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('inventory_count_lines', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('inventory_count_id')->index();
            $table->uuid('product_id')->index();
            $table->integer('system_quantity');
            $table->integer('counted_quantity');
            $table->integer('delta_quantity');
            $table->string('reason')->nullable();
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('inventory_waste_records', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('outlet_id')->index();
            $table->uuid('product_id')->index();
            $table->uuid('actor_id')->nullable()->index();
            $table->string('number')->index();
            $table->integer('quantity');
            $table->string('reason');
            $table->string('status')->default('completed')->index();
            $table->timestamp('occurred_at');
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('inventory_transfers', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('source_outlet_id')->index();
            $table->uuid('destination_outlet_id')->index();
            $table->uuid('actor_id')->nullable()->index();
            $table->string('number')->index();
            $table->string('status')->default('requested')->index();
            $table->text('notes')->nullable();
            $table->timestamp('sent_at')->nullable();
            $table->timestamp('received_at')->nullable();
            $table->timestamp('cancelled_at')->nullable();
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::create('inventory_transfer_lines', function (Blueprint $table): void {
            $table->uuid('id')->primary();
            $table->uuid('business_id')->index();
            $table->uuid('inventory_transfer_id')->index();
            $table->uuid('product_id')->index();
            $table->integer('requested_quantity');
            $table->integer('sent_quantity')->default(0);
            $table->integer('received_quantity')->default(0);
            $table->timestamps();
            $table->softDeletes();
        });

        Schema::table('stock_movements', function (Blueprint $table): void {
            $table->uuid('inventory_count_id')->nullable()->index();
            $table->uuid('waste_id')->nullable()->index();
            $table->uuid('inventory_transfer_id')->nullable()->index();
            $table->uuid('actor_id')->nullable()->index();
            $table->integer('before_quantity')->nullable();
            $table->integer('after_quantity')->nullable();
            $table->string('reason')->nullable();
        });
    }

    public function down(): void
    {
        Schema::table('stock_movements', function (Blueprint $table): void {
            $table->dropColumn([
                'inventory_count_id',
                'waste_id',
                'inventory_transfer_id',
                'actor_id',
                'before_quantity',
                'after_quantity',
                'reason',
            ]);
        });

        Schema::dropIfExists('inventory_transfer_lines');
        Schema::dropIfExists('inventory_transfers');
        Schema::dropIfExists('inventory_waste_records');
        Schema::dropIfExists('inventory_count_lines');
        Schema::dropIfExists('inventory_counts');
    }
};
