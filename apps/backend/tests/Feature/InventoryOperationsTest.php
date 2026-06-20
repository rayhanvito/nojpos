<?php

namespace Tests\Feature;

use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class InventoryOperationsTest extends TestCase
{
    use RefreshDatabase;

    public function test_owner_can_list_inventory_and_movements_for_own_tenant(): void
    {
        $ctx = $this->inventoryContext();
        $this->seedStock($ctx, 8);

        $inventory = $this->withToken($ctx['token'])->getJson('/api/v1/inventory?outlet_id='.$ctx['source_outlet'])->assertOk();
        $inventory->assertJsonStructure(['data' => ['items' => [[
            'product_id',
            'product_name',
            'outlet_id',
            'outlet_name',
            'on_hand_quantity',
            'in_transit_out_quantity',
            'available_quantity',
        ]], 'page', 'per_page'], 'meta']);
        $inventory->assertJsonPath('data.items.0.product_id', $ctx['product']);
        $inventory->assertJsonPath('data.items.0.on_hand_quantity', 8);

        $movements = $this->withToken($ctx['token'])->getJson('/api/v1/inventory/movements?outlet_id='.$ctx['source_outlet'])->assertOk();
        $movements->assertJsonPath('data.items.0.type', 'purchase');
        $movements->assertJsonPath('data.items.0.quantity_delta', 8);
    }

    public function test_cashier_cannot_perform_privileged_stock_writes(): void
    {
        $ctx = $this->inventoryContext(role: 'cashier');

        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/waste', [
            'outlet_id' => $ctx['source_outlet'],
            'product_id' => $ctx['product'],
            'quantity' => 1,
            'reason' => 'Rusak',
        ], $this->idempotencyHeader())->assertForbidden()->assertJsonPath('error.code', 'FORBIDDEN');
    }

    public function test_cross_business_outlet_and_product_are_blocked(): void
    {
        $ctx = $this->inventoryContext();
        $other = $this->inventoryContext(email: 'other@example.test');

        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/waste', [
            'outlet_id' => $other['source_outlet'],
            'product_id' => $ctx['product'],
            'quantity' => 1,
            'reason' => 'Rusak',
        ], $this->idempotencyHeader())->assertForbidden()->assertJsonPath('error.code', 'FORBIDDEN');

        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/waste', [
            'outlet_id' => $ctx['source_outlet'],
            'product_id' => $other['product'],
            'quantity' => 1,
            'reason' => 'Rusak',
        ], $this->idempotencyHeader())->assertForbidden()->assertJsonPath('error.code', 'FORBIDDEN');
    }

    public function test_inventory_count_creates_adjustment_movement_and_is_idempotent(): void
    {
        $ctx = $this->inventoryContext();
        $this->seedStock($ctx, 10);
        $key = $this->idempotencyHeader();
        $payload = [
            'outlet_id' => $ctx['source_outlet'],
            'notes' => 'Cycle count',
            'lines' => [[
                'product_id' => $ctx['product'],
                'counted_quantity' => 7,
                'reason' => 'Selisih opname',
            ]],
        ];

        $first = $this->withToken($ctx['token'])->postJson('/api/v1/inventory/counts', $payload, $key)->assertCreated();
        $second = $this->withToken($ctx['token'])->postJson('/api/v1/inventory/counts', $payload, $key)->assertCreated();

        $this->assertSame($first->json('data.id'), $second->json('data.id'));
        $first->assertJsonPath('data.lines.0.system_quantity', 10)
            ->assertJsonPath('data.lines.0.counted_quantity', 7)
            ->assertJsonPath('data.lines.0.delta_quantity', -3);
        $this->assertSame(7, $this->stock($ctx, $ctx['source_outlet']));
        $this->assertSame(1, DB::table('stock_movements')->where('business_id', $ctx['business'])->where('type', 'adjustment')->count());
        $this->assertSame(1, DB::table('audit_logs')->where('business_id', $ctx['business'])->where('action', 'inventory.count.finalize')->count());
    }

    public function test_waste_reduces_stock_requires_reason_and_rejects_shortage(): void
    {
        $ctx = $this->inventoryContext();
        $this->seedStock($ctx, 5);

        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/waste', [
            'outlet_id' => $ctx['source_outlet'],
            'product_id' => $ctx['product'],
            'quantity' => 1,
        ], $this->idempotencyHeader())->assertUnprocessable()->assertJsonPath('error.code', 'VALIDATION_ERROR');

        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/waste', [
            'outlet_id' => $ctx['source_outlet'],
            'product_id' => $ctx['product'],
            'quantity' => 6,
            'reason' => 'Rusak',
        ], $this->idempotencyHeader())->assertConflict()->assertJsonPath('error.code', 'STOCK_INSUFFICIENT');

        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/waste', [
            'outlet_id' => $ctx['source_outlet'],
            'product_id' => $ctx['product'],
            'quantity' => 2,
            'reason' => 'Kadaluarsa',
        ], $this->idempotencyHeader())->assertCreated()->assertJsonPath('data.quantity', 2)->assertJsonPath('data.reason', 'Kadaluarsa');

        $this->assertSame(3, $this->stock($ctx, $ctx['source_outlet']));
        $this->assertDatabaseHas('stock_movements', ['business_id' => $ctx['business'], 'product_id' => $ctx['product'], 'type' => 'waste', 'quantity_delta' => -2, 'reason' => 'Kadaluarsa']);
        $this->assertSame(1, DB::table('audit_logs')->where('business_id', $ctx['business'])->where('action', 'inventory.waste.finalize')->count());
    }

    public function test_transfer_create_send_receive_lifecycle_reconciles_in_transit_stock(): void
    {
        $ctx = $this->inventoryContext();
        $this->seedStock($ctx, 10);
        $create = $this->createTransfer($ctx, quantity: 4);

        $create->assertCreated()
            ->assertJsonPath('data.status', 'requested')
            ->assertJsonPath('data.lines.0.requested_quantity', 4);
        $transferId = $create->json('data.id');
        $this->assertSame(10, $this->stock($ctx, $ctx['source_outlet']));
        $this->assertSame(0, $this->stock($ctx, $ctx['destination_outlet']));

        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/transfers/'.$transferId.'/send', [], $this->idempotencyHeader())
            ->assertOk()
            ->assertJsonPath('data.status', 'in_transit')
            ->assertJsonPath('data.lines.0.sent_quantity', 4)
            ->assertJsonPath('data.lines.0.in_transit_quantity', 4);

        $this->assertSame(6, $this->stock($ctx, $ctx['source_outlet']));
        $this->assertSame(0, $this->stock($ctx, $ctx['destination_outlet']));
        $inTransit = $this->withToken($ctx['token'])->getJson('/api/v1/inventory/transfers/in-transit?outlet_id='.$ctx['source_outlet'])->assertOk();
        $inTransit->assertJsonPath('data.items.0.id', $transferId);
        $overview = $this->withToken($ctx['token'])->getJson('/api/v1/inventory?outlet_id='.$ctx['source_outlet'])->assertOk();
        $overview->assertJsonPath('data.items.0.on_hand_quantity', 6)
            ->assertJsonPath('data.items.0.available_quantity', 6)
            ->assertJsonPath('data.items.0.in_transit_out_quantity', 4);

        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/transfers/'.$transferId.'/receive', [], $this->idempotencyHeader())
            ->assertOk()
            ->assertJsonPath('data.status', 'received')
            ->assertJsonPath('data.lines.0.received_quantity', 4)
            ->assertJsonPath('data.lines.0.in_transit_quantity', 0);

        $this->assertSame(6, $this->stock($ctx, $ctx['source_outlet']));
        $this->assertSame(4, $this->stock($ctx, $ctx['destination_outlet']));
        $this->assertDatabaseHas('stock_movements', ['business_id' => $ctx['business'], 'type' => 'transfer_out', 'quantity_delta' => -4]);
        $this->assertDatabaseHas('stock_movements', ['business_id' => $ctx['business'], 'type' => 'transfer_in', 'quantity_delta' => 4]);
        $this->assertSame(1, DB::table('audit_logs')->where('business_id', $ctx['business'])->where('action', 'inventory.transfer.dispatch')->count());
        $this->assertSame(1, DB::table('audit_logs')->where('business_id', $ctx['business'])->where('action', 'inventory.transfer.receive')->count());
    }

    public function test_transfer_cannot_send_or_receive_twice_and_cannot_cancel_after_send(): void
    {
        $ctx = $this->inventoryContext();
        $this->seedStock($ctx, 10);
        $transferId = $this->createTransfer($ctx, quantity: 2)->json('data.id');

        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/transfers/'.$transferId.'/send', [], $this->idempotencyHeader())->assertOk();
        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/transfers/'.$transferId.'/send', [], $this->idempotencyHeader())->assertConflict()->assertJsonPath('error.code', 'TRANSFER_STATE_INVALID');
        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/transfers/'.$transferId.'/cancel', [], $this->idempotencyHeader())->assertConflict()->assertJsonPath('error.code', 'TRANSFER_STATE_INVALID');
        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/transfers/'.$transferId.'/receive', [], $this->idempotencyHeader())->assertOk();
        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/transfers/'.$transferId.'/receive', [], $this->idempotencyHeader())->assertConflict()->assertJsonPath('error.code', 'TRANSFER_STATE_INVALID');
    }

    public function test_cancel_before_send_works_without_stock_effect(): void
    {
        $ctx = $this->inventoryContext();
        $this->seedStock($ctx, 5);
        $transferId = $this->createTransfer($ctx, quantity: 3)->json('data.id');

        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/transfers/'.$transferId.'/cancel', [], $this->idempotencyHeader())
            ->assertOk()
            ->assertJsonPath('data.status', 'cancelled');

        $this->assertSame(5, $this->stock($ctx, $ctx['source_outlet']));
        $this->assertSame(0, DB::table('stock_movements')->where('business_id', $ctx['business'])->where('type', 'transfer_out')->count());
        $this->assertSame(1, DB::table('audit_logs')->where('business_id', $ctx['business'])->where('action', 'inventory.transfer.cancel')->count());
    }

    public function test_idempotency_mismatch_rejects_same_key_different_body(): void
    {
        $ctx = $this->inventoryContext();
        $this->seedStock($ctx, 5);
        $headers = $this->idempotencyHeader();

        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/waste', [
            'outlet_id' => $ctx['source_outlet'],
            'product_id' => $ctx['product'],
            'quantity' => 1,
            'reason' => 'Rusak',
        ], $headers)->assertCreated();

        $this->withToken($ctx['token'])->postJson('/api/v1/inventory/waste', [
            'outlet_id' => $ctx['source_outlet'],
            'product_id' => $ctx['product'],
            'quantity' => 2,
            'reason' => 'Rusak',
        ], $headers)->assertConflict()->assertJsonPath('error.code', 'IDEMPOTENCY_MISMATCH');
    }

    public function test_existing_cash_checkout_stock_deduction_still_works(): void
    {
        $ctx = $this->inventoryContext(role: 'cashier');
        $this->seedStock($ctx, 3);
        DB::table('devices')->insert(['id' => $ctx['device'], 'business_id' => $ctx['business'], 'outlet_id' => $ctx['source_outlet'], 'device_uuid' => 'dev-'.$ctx['device'], 'name' => 'Device', 'created_at' => now(), 'updated_at' => now()]);
        DB::table('shift_sessions')->insert(['id' => $ctx['shift'], 'business_id' => $ctx['business'], 'outlet_id' => $ctx['source_outlet'], 'device_id' => $ctx['device'], 'cashier_id' => $ctx['user'], 'status' => 'open', 'opening_cash' => 0, 'opened_at' => now(), 'created_at' => now(), 'updated_at' => now()]);

        $this->withToken($ctx['token'])->postJson('/api/v1/transactions', [
            'outlet_id' => $ctx['source_outlet'],
            'device_id' => $ctx['device'],
            'cashier_id' => $ctx['user'],
            'shift_id' => $ctx['shift'],
            'items' => [[
                'product_id' => $ctx['product'],
                'quantity' => 2,
                'unit_price' => 10000,
                'discount' => 0,
            ]],
            'payments' => [[
                'method' => 'cash',
                'amount' => 20000,
            ]],
        ], $this->idempotencyHeader())->assertCreated()->assertJsonPath('data.status', 'paid');

        $this->assertSame(1, $this->stock($ctx, $ctx['source_outlet']));
        $this->assertSame(1, DB::table('stock_movements')->where('business_id', $ctx['business'])->where('type', 'sale')->count());
    }

    private function inventoryContext(string $role = 'owner', string $email = 'owner@example.test'): array
    {
        $business = (string) Str::uuid();
        $user = (string) Str::uuid();
        $sourceOutlet = (string) Str::uuid();
        $destinationOutlet = (string) Str::uuid();
        $product = (string) Str::uuid();
        $device = (string) Str::uuid();
        $shift = (string) Str::uuid();

        DB::table('businesses')->insert(['id' => $business, 'name' => 'Business '.$business, 'created_at' => now(), 'updated_at' => now()]);
        DB::table('users')->insert(['id' => $user, 'business_id' => $business, 'name' => 'User '.$role, 'email' => $email, 'password' => Hash::make('password'), 'role' => $role, 'pin_hash' => Hash::make('1234'), 'created_at' => now(), 'updated_at' => now()]);
        DB::table('outlets')->insert([
            ['id' => $sourceOutlet, 'business_id' => $business, 'name' => 'Source', 'service_charge_rate' => 0, 'tax_rate' => 0, 'created_at' => now(), 'updated_at' => now()],
            ['id' => $destinationOutlet, 'business_id' => $business, 'name' => 'Destination', 'service_charge_rate' => 0, 'tax_rate' => 0, 'created_at' => now(), 'updated_at' => now()],
        ]);
        DB::table('products')->insert(['id' => $product, 'business_id' => $business, 'outlet_id' => $sourceOutlet, 'name' => 'Tracked product', 'barcode' => 'SKU-'.$product, 'price' => 10000, 'track_stock' => true, 'created_at' => now(), 'updated_at' => now()]);
        DB::table('payment_method_configs')->insert(['id' => (string) Str::uuid(), 'business_id' => $business, 'outlet_id' => $sourceOutlet, 'method' => 'cash', 'is_cash' => true, 'created_at' => now(), 'updated_at' => now()]);
        DB::table('personal_access_tokens')->insert(['id' => (string) Str::uuid(), 'tokenable_type' => 'App\\Models\\User', 'tokenable_id' => $user, 'name' => 'test', 'token' => hash('sha256', 'plain-'.$user), 'abilities' => json_encode(['*']), 'created_at' => now(), 'updated_at' => now()]);

        return [
            'business' => $business,
            'user' => $user,
            'token' => 'plain-'.$user,
            'source_outlet' => $sourceOutlet,
            'destination_outlet' => $destinationOutlet,
            'product' => $product,
            'device' => $device,
            'shift' => $shift,
        ];
    }

    private function seedStock(array $ctx, int $quantity): void
    {
        DB::table('stock_movements')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $ctx['business'],
            'outlet_id' => $ctx['source_outlet'],
            'product_id' => $ctx['product'],
            'type' => 'purchase',
            'quantity_delta' => $quantity,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
    }

    private function createTransfer(array $ctx, int $quantity)
    {
        return $this->withToken($ctx['token'])->postJson('/api/v1/inventory/transfers', [
            'source_outlet_id' => $ctx['source_outlet'],
            'destination_outlet_id' => $ctx['destination_outlet'],
            'lines' => [[
                'product_id' => $ctx['product'],
                'quantity' => $quantity,
            ]],
        ], $this->idempotencyHeader());
    }

    private function stock(array $ctx, string $outlet): int
    {
        return (int) DB::table('stock_movements')
            ->where('business_id', $ctx['business'])
            ->where('outlet_id', $outlet)
            ->where('product_id', $ctx['product'])
            ->sum('quantity_delta');
    }

    private function idempotencyHeader(): array
    {
        return ['Idempotency-Key' => (string) Str::uuid()];
    }
}
