<?php

namespace Tests\Feature;

use App\Models\User;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Illuminate\Testing\TestResponse;
use Laravel\Sanctum\Sanctum;
use Tests\TestCase;

class ParkedOrderLeaseTest extends TestCase
{
    use RefreshDatabase;

    public function test_lease_acquire_blocks_other_device_until_expiry_then_allows_acquire(): void
    {
        [$businessId, $outletId, $deviceA, $cashierA, $tokenA] = $this->terminalContext();
        $deviceB = $this->device($businessId, $outletId, 'tablet-b');
        $cashierB = $this->user($businessId, 'cashier-b@example.test', 'cashier');
        $shiftId = $this->shift($businessId, $outletId, $deviceA, $cashierA);
        $productId = $this->product($businessId, $outletId, 'Nasi Goreng', 20000);
        $orderId = $this->parkedOrder($tokenA, $outletId, $deviceA, $cashierA, $shiftId, $productId)->json('data.id');

        $this->withToken($tokenA)->postJson("/api/v1/parked-orders/{$orderId}/lease/acquire", [
            'device_id' => $deviceA,
        ])->assertOk()
            ->assertJsonPath('data.lease.device_id', $deviceA)
            ->assertJsonPath('data.lease.user_id', $cashierA);

        Sanctum::actingAs(User::query()->findOrFail($cashierB));
        $this->postJson("/api/v1/parked-orders/{$orderId}/lease/acquire", [
            'device_id' => $deviceB,
        ])->assertStatus(409)
            ->assertJsonPath('error.code', 'ORDER_LOCKED')
            ->assertJsonPath('error.details.lease_device_id', $deviceA)
            ->assertJsonPath('error.details.leased_by_user_id', $cashierA);

        DB::table('transactions')->where('id', $orderId)->update([
            'lease_expires_at' => now()->subSecond(),
        ]);

        $this->postJson("/api/v1/parked-orders/{$orderId}/lease/acquire", [
            'device_id' => $deviceB,
        ])->assertOk()
            ->assertJsonPath('data.lease.device_id', $deviceB)
            ->assertJsonPath('data.lease.user_id', $cashierB);
    }

    public function test_owner_can_refresh_update_revision_conflict_and_release_lease(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->terminalContext();
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $cashierId);
        $productId = $this->product($businessId, $outletId, 'Kopi Susu', 15000);
        $order = $this->parkedOrder($token, $outletId, $deviceId, $cashierId, $shiftId, $productId)
            ->assertCreated()
            ->assertJsonPath('data.revision', 1);
        $orderId = $order->json('data.id');

        $this->withToken($token)->postJson("/api/v1/parked-orders/{$orderId}/lease/acquire", [
            'device_id' => $deviceId,
        ])->assertOk();

        $firstExpiry = DB::table('transactions')->where('id', $orderId)->value('lease_expires_at');
        $this->travel(10)->seconds();

        $this->withToken($token)->postJson("/api/v1/parked-orders/{$orderId}/lease/refresh", [
            'device_id' => $deviceId,
        ])->assertOk()
            ->assertJsonPath('data.lease.device_id', $deviceId);
        $this->assertTrue(DB::table('transactions')->where('id', $orderId)->value('lease_expires_at') > $firstExpiry);

        $this->withToken($token)->putJson("/api/v1/parked-orders/{$orderId}", [
            'device_id' => $deviceId,
            'expected_revision' => 1,
            'notes' => 'Tambah gula',
            'items' => [
                ['product_id' => $productId, 'quantity' => 2],
            ],
        ])->assertOk()
            ->assertJsonPath('data.revision', 2)
            ->assertJsonPath('data.grand_total', 30000)
            ->assertJsonPath('data.items.0.quantity', 2);

        $this->withToken($token)->putJson("/api/v1/parked-orders/{$orderId}", [
            'device_id' => $deviceId,
            'expected_revision' => 1,
            'items' => [
                ['product_id' => $productId, 'quantity' => 1],
            ],
        ])->assertStatus(409)
            ->assertJsonPath('error.code', 'CONFLICT_REVISION')
            ->assertJsonPath('error.details.current_revision', 2);

        $this->withToken($token)->postJson("/api/v1/parked-orders/{$orderId}/lease/release", [
            'device_id' => $deviceId,
        ])->assertOk()
            ->assertJsonPath('data.lease.device_id', null)
            ->assertJsonPath('data.lease.user_id', null);

        $this->assertDatabaseHas('audit_logs', [
            'business_id' => $businessId,
            'actor_id' => $cashierId,
            'action' => 'parked_order.update',
            'entity_type' => 'transaction',
            'entity_id' => $orderId,
        ]);
    }

    public function test_update_requires_active_owner_lease_and_existing_held_listing_still_works(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->terminalContext();
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $cashierId);
        $productId = $this->product($businessId, $outletId, 'Teh Tawar', 5000);
        $orderId = $this->parkedOrder($token, $outletId, $deviceId, $cashierId, $shiftId, $productId)->json('data.id');

        $this->withToken($token)->putJson("/api/v1/parked-orders/{$orderId}", [
            'device_id' => $deviceId,
            'expected_revision' => 1,
            'items' => [
                ['product_id' => $productId, 'quantity' => 2],
            ],
        ])->assertStatus(409)
            ->assertJsonPath('error.code', 'ORDER_LOCK_REQUIRED');

        $this->withToken($token)->getJson('/api/v1/parked-orders')
            ->assertOk()
            ->assertJsonCount(1, 'data.parked_orders')
            ->assertJsonPath('data.parked_orders.0.id', $orderId)
            ->assertJsonPath('data.parked_orders.0.status', 'held')
            ->assertJsonPath('data.parked_orders.0.revision', 1);

        $this->withToken($token)->getJson('/api/v1/transactions?status=held')
            ->assertOk()
            ->assertJsonCount(1, 'data.transactions')
            ->assertJsonPath('data.transactions.0.id', $orderId)
            ->assertJsonPath('data.transactions.0.status', 'held');
    }

    private function parkedOrder(string $token, string $outletId, string $deviceId, string $cashierId, string $shiftId, string $productId): TestResponse
    {
        return $this->withToken($token)->postJson('/api/v1/parked-orders', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'items' => [
                ['product_id' => $productId, 'quantity' => 1],
            ],
        ]);
    }

    private function terminalContext(string $role = 'cashier'): array
    {
        $businessId = $this->business();
        $outletId = $this->outlet($businessId);
        $deviceId = $this->device($businessId, $outletId, 'tablet-a');
        $cashierId = $this->user($businessId, $role.'@example.test', $role);

        return [$businessId, $outletId, $deviceId, $cashierId, $this->tokenFor($cashierId)];
    }

    private function tokenFor(string $userId): string
    {
        DB::table('personal_access_tokens')->insert([
            'id' => (string) Str::uuid(),
            'tokenable_type' => 'App\\Models\\User',
            'tokenable_id' => $userId,
            'name' => 'test',
            'token' => hash('sha256', 'plain-'.$userId),
            'abilities' => json_encode(['*']),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return 'plain-'.$userId;
    }

    private function business(): string
    {
        $id = (string) Str::uuid();
        DB::table('businesses')->insert([
            'id' => $id,
            'name' => 'Kedai Test',
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function outlet(string $businessId): string
    {
        $id = (string) Str::uuid();
        DB::table('outlets')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'name' => 'Outlet Test',
            'service_charge_rate' => 0,
            'tax_rate' => 0,
            'timezone' => 'Asia/Jakarta',
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function device(string $businessId, string $outletId, string $uuid): string
    {
        $id = (string) Str::uuid();
        DB::table('devices')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'device_uuid' => $uuid,
            'name' => $uuid,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function user(string $businessId, string $email, string $role): string
    {
        $id = (string) Str::uuid();
        DB::table('users')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'name' => $role.' user',
            'email' => $email,
            'password' => Hash::make('password'),
            'role' => $role,
            'pin_hash' => Hash::make('1234'),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function product(string $businessId, string $outletId, string $name, int $price): string
    {
        $categoryId = (string) Str::uuid();
        DB::table('product_categories')->insert([
            'id' => $categoryId,
            'business_id' => $businessId,
            'name' => 'Kategori',
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        $id = (string) Str::uuid();
        DB::table('products')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'product_category_id' => $categoryId,
            'name' => $name,
            'price' => $price,
            'track_stock' => true,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function shift(string $businessId, string $outletId, string $deviceId, string $cashierId): string
    {
        $id = (string) Str::uuid();
        DB::table('shift_sessions')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'status' => 'open',
            'opening_cash' => 0,
            'opened_at' => now(),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }
}
