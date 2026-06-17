<?php

namespace Tests\Feature;

use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class NojposWave1ApiTest extends TestCase
{
    use RefreshDatabase;

    public function test_login_returns_sanctum_token_session_outlets_and_audit_log(): void
    {
        $businessId = $this->business('Kedai Nusantara');
        $outletId = $this->outlet($businessId, 'Outlet Utama');
        $this->device($businessId, $outletId, 'tablet-1');
        $this->user($businessId, 'owner@example.test', 'owner');

        $response = $this->postJson('/api/v1/auth/login', [
            'email' => 'owner@example.test',
            'password' => 'password',
            'device_uuid' => 'tablet-1',
        ], ['Accept' => 'application/json']);

        $response->assertOk()
            ->assertJsonPath('data.user.email', 'owner@example.test')
            ->assertJsonPath('data.business.id', $businessId)
            ->assertJsonPath('data.outlets.0.id', $outletId)
            ->assertJsonStructure(['data' => ['token'], 'meta']);

        $this->assertDatabaseHas('audit_logs', [
            'business_id' => $businessId,
            'action' => 'auth.login',
            'actor_id' => $response->json('data.user.id'),
        ]);
    }

    public function test_pin_switch_is_server_verified_rate_limited_and_audited(): void
    {
        $businessId = $this->business();
        $outletId = $this->outlet($businessId);
        $deviceId = $this->device($businessId, $outletId, 'tablet-1');
        $owner = $this->user($businessId, 'owner@example.test', 'owner');
        $cashier = $this->user($businessId, 'cashier@example.test', 'cashier', '1234');
        $token = $this->tokenFor($owner);

        for ($i = 0; $i < 5; $i++) {
            $this->withToken($token)->postJson('/api/v1/auth/pin-switch', [
                'pin' => '9999',
                'device_id' => $deviceId,
                'outlet_id' => $outletId,
            ])->assertStatus(422);
        }

        $this->withToken($token)->postJson('/api/v1/auth/pin-switch', [
            'pin' => '1234',
            'device_id' => $deviceId,
            'outlet_id' => $outletId,
        ])->assertStatus(429)
            ->assertJsonPath('error.code', 'PIN_LOCKED');

        DB::table('pin_attempts')->where('business_id', $businessId)->update([
            'locked_until' => now()->subMinute(),
            'attempts' => 0,
        ]);

        $this->withToken($token)->postJson('/api/v1/auth/pin-switch', [
            'pin' => '1234',
            'device_id' => $deviceId,
            'outlet_id' => $outletId,
        ])->assertOk()
            ->assertJsonPath('data.cashier.id', $cashier);

        $this->assertDatabaseHas('audit_logs', [
            'business_id' => $businessId,
            'action' => 'auth.pin_switch',
            'actor_id' => $owner,
        ]);
    }

    public function test_tenant_scope_blocks_cross_business_operational_write(): void
    {
        $businessA = $this->business('A');
        $outletA = $this->outlet($businessA);
        $deviceA = $this->device($businessA, $outletA, 'tablet-a');
        $userA = $this->user($businessA, 'a@example.test', 'cashier');

        $businessB = $this->business('B');
        $outletB = $this->outlet($businessB);

        $this->withToken($this->tokenFor($userA))->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletB,
            'device_id' => $deviceA,
            'cashier_id' => $userA,
            'opening_cash' => 100000,
        ])->assertStatus(403)
            ->assertJsonPath('error.code', 'FORBIDDEN');
    }

    public function test_products_are_scoped_and_filterable(): void
    {
        $businessA = $this->business('A');
        $outletA = $this->outlet($businessA);
        $userA = $this->user($businessA, 'a@example.test', 'cashier');
        $categoryA = $this->category($businessA, 'Minuman');
        $this->product($businessA, $outletA, $categoryA, 'Teh Manis', 8000, 'TEH-1');

        $businessB = $this->business('B');
        $outletB = $this->outlet($businessB);
        $categoryB = $this->category($businessB, 'Makanan');
        $this->product($businessB, $outletB, $categoryB, 'Nasi Rahasia', 12000, 'NASI-1');

        $this->withToken($this->tokenFor($userA))->getJson('/api/v1/products?search=Teh&category=Minuman')
            ->assertOk()
            ->assertJsonCount(1, 'data.products')
            ->assertJsonPath('data.products.0.name', 'Teh Manis');
    }

    public function test_shift_lifecycle_cash_movements_expected_cash_and_double_close_guard(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();

        $open = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 500000,
        ])->assertCreated();

        $shiftId = $open->json('data.id');

        $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 1,
        ])->assertStatus(422)
            ->assertJsonPath('error.code', 'SHIFT_ALREADY_OPEN');

        $this->withToken($token)->postJson("/api/v1/shifts/{$shiftId}/cash-movements", [
            'type' => 'cash_in',
            'amount' => 50000,
            'reason' => 'Tambah modal',
        ])->assertCreated();

        $close = $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson("/api/v1/shifts/{$shiftId}/close", [
            'actual_cash' => 550000,
        ])->assertOk();

        $close->assertJsonPath('data.expected_cash', 550000)
            ->assertJsonPath('data.cash_difference', 0)
            ->assertJsonPath('data.status', 'closed');

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson("/api/v1/shifts/{$shiftId}/close", [
            'actual_cash' => 550000,
        ])->assertStatus(422)
            ->assertJsonPath('error.code', 'SHIFT_ALREADY_CLOSED');

        $this->assertDatabaseHas('audit_logs', [
            'business_id' => $businessId,
            'action' => 'shift.close',
        ]);
    }

    public function test_transaction_checkout_calculates_totals_creates_stock_movement_and_is_idempotent(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $categoryId = $this->category($businessId, 'Makanan');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Mie Ayam', 18000);
        $this->paymentMethod($businessId, $outletId, 'cash', true);

        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 100000,
        ])->json('data.id');

        DB::table('outlets')->where('id', $outletId)->update([
            'service_charge_rate' => 10,
            'tax_rate' => 11,
        ]);

        $payload = [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'cart_discount' => 1000,
            'rounding' => 0,
            'notes' => 'No spicy',
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 2,
                    'unit_price' => 18000,
                    'discount' => 2000,
                ],
            ],
            'payments' => [
                ['method' => 'cash', 'amount' => 40000],
            ],
        ];

        $key = (string) Str::uuid();
        $first = $this->withToken($token)->withHeaders(['Idempotency-Key' => $key])
            ->postJson('/api/v1/transactions', $payload)
            ->assertCreated();

        $first->assertJsonPath('data.subtotal', 36000)
            ->assertJsonPath('data.discount_total', 3000)
            ->assertJsonPath('data.service_charge_total', 3300)
            ->assertJsonPath('data.tax_total', 3993)
            ->assertJsonPath('data.grand_total', 40293)
            ->assertJsonPath('data.status', 'paid');

        $this->assertDatabaseHas('stock_movements', [
            'business_id' => $businessId,
            'product_id' => $productId,
            'type' => 'sale',
            'quantity_delta' => -2,
        ]);

        $this->withToken($token)->withHeaders(['Idempotency-Key' => $key])
            ->postJson('/api/v1/transactions', $payload)
            ->assertCreated()
            ->assertExactJson($first->json());

        $payload['cart_discount'] = 2000;
        $this->withToken($token)->withHeaders(['Idempotency-Key' => $key])
            ->postJson('/api/v1/transactions', $payload)
            ->assertStatus(409)
            ->assertJsonPath('error.code', 'IDEMPOTENCY_CONFLICT');
    }

    public function test_non_cash_payment_starts_pending_and_can_be_confirmed(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $this->paymentMethod($businessId, $outletId, 'transfer', false);
        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 0,
        ])->json('data.id');

        $transactionId = $this->transaction($businessId, $outletId, $deviceId, $cashierId, $shiftId, 50000);

        $payment = $this->withToken($token)->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/payments', [
                'transaction_id' => $transactionId,
                'method' => 'transfer',
                'amount' => 50000,
                'reference' => 'BCA-123',
            ])->assertCreated();

        $paymentId = $payment->json('data.id');
        $payment->assertJsonPath('data.status', 'pending');

        $this->withToken($token)->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/payments', [
                'payment_id' => $paymentId,
                'confirm' => true,
            ])->assertOk()
            ->assertJsonPath('data.status', 'confirmed')
            ->assertJsonPath('data.confirmed_by', $cashierId);
    }

    private function cashierContext(): array
    {
        $businessId = $this->business();
        $outletId = $this->outlet($businessId);
        $deviceId = $this->device($businessId, $outletId, 'tablet-1');
        $cashierId = $this->user($businessId, 'cashier@example.test', 'cashier');

        return [$businessId, $outletId, $deviceId, $cashierId, $this->tokenFor($cashierId)];
    }

    private function tokenFor(string $userId): string
    {
        return DB::table('personal_access_tokens')->insertGetId([
            'id' => (string) Str::uuid(),
            'tokenable_type' => 'App\\Models\\User',
            'tokenable_id' => $userId,
            'name' => 'test',
            'token' => hash('sha256', 'plain-'.$userId),
            'abilities' => json_encode(['*']),
            'created_at' => now(),
            'updated_at' => now(),
        ]) ? 'plain-'.$userId : '';
    }

    private function business(string $name = 'Kedai Nusantara'): string
    {
        $id = (string) Str::uuid();
        DB::table('businesses')->insert([
            'id' => $id,
            'name' => $name,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function outlet(string $businessId, string $name = 'Outlet'): string
    {
        $id = (string) Str::uuid();
        DB::table('outlets')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'name' => $name,
            'service_charge_rate' => 0,
            'tax_rate' => 0,
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

    private function user(string $businessId, string $email, string $role, string $pin = '1234'): string
    {
        $id = (string) Str::uuid();
        DB::table('users')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'name' => $role.' user',
            'email' => $email,
            'password' => Hash::make('password'),
            'role' => $role,
            'pin_hash' => Hash::make($pin),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function category(string $businessId, string $name): string
    {
        $id = (string) Str::uuid();
        DB::table('product_categories')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'name' => $name,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function product(string $businessId, string $outletId, string $categoryId, string $name, int $price, ?string $barcode = null): string
    {
        $id = (string) Str::uuid();
        DB::table('products')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'product_category_id' => $categoryId,
            'name' => $name,
            'barcode' => $barcode,
            'price' => $price,
            'track_stock' => true,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function paymentMethod(string $businessId, string $outletId, string $method, bool $isCash): void
    {
        DB::table('payment_method_configs')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'method' => $method,
            'is_cash' => $isCash,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
    }

    private function transaction(string $businessId, string $outletId, string $deviceId, string $cashierId, string $shiftId, int $total): string
    {
        $id = (string) Str::uuid();
        DB::table('transactions')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'number' => 'TST-001',
            'status' => 'pending',
            'subtotal' => $total,
            'discount_total' => 0,
            'service_charge_total' => 0,
            'tax_total' => 0,
            'rounding_total' => 0,
            'grand_total' => $total,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }
}
