<?php

namespace Tests\Feature;

use Carbon\CarbonImmutable;
use Database\Seeders\DemoSeeder;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class NojposWave1ApiTest extends TestCase
{
    use RefreshDatabase;

    public function test_demo_seeder_supports_flutter_wave_1_smoke_flow(): void
    {
        $this->seed(DemoSeeder::class);
        $this->seed(DemoSeeder::class);

        $this->assertSame(1, DB::table('businesses')->where('name', 'NojPOS Demo')->count());
        $this->assertSame(1, DB::table('outlets')->where('name', 'Outlet Demo')->count());
        $this->assertSame(8, DB::table('products')->where('business_id', '11111111-1111-4111-8111-111111111111')->count());

        $login = $this->postJson('/api/v1/auth/login', [
            'email' => 'owner@demo.nojpos.test',
            'password' => 'password',
            'device_uuid' => 'demo-tablet-001',
        ], ['Accept' => 'application/json'])->assertOk();

        $token = $login->json('data.token');
        $businessId = $login->json('data.business.id');
        $outletId = $login->json('data.outlets.0.id');
        $deviceId = DB::table('devices')->where('device_uuid', 'demo-tablet-001')->value('id');

        $this->withToken($token)->getJson('/api/v1/me')
            ->assertOk()
            ->assertJsonPath('data.user.email', 'owner@demo.nojpos.test')
            ->assertJsonPath('data.business.id', $businessId)
            ->assertJsonPath('data.outlets.0.id', $outletId)
            ->assertJsonPath('data.outlets.0.timezone', 'Asia/Jakarta')
            ->assertJsonPath('data.outlets.0.receipt_config.paper_width', '58mm')
            ->assertJsonPath('data.outlets.0.receipt_config.header_name', 'NojPOS Demo')
            ->assertJsonPath('data.outlets.0.receipt_config.show_logo', false)
            ->assertJsonPath('data.outlets.0.receipt_config.show_qris_info', true);

        $this->withToken($token)->getJson('/api/v1/products')
            ->assertOk()
            ->assertJsonCount(8, 'data.products')
            ->assertJsonPath('data.products.0.business_id', $businessId);

        $this->withToken($token)->postJson('/api/v1/auth/pin-switch', [
            'pin' => '0000',
            'device_id' => $deviceId,
            'outlet_id' => $outletId,
        ])->assertStatus(422)
            ->assertJsonPath('error.code', 'INVALID_PIN');

        $this->withToken($token)->postJson('/api/v1/auth/pin-switch', [
            'pin' => '1234',
            'device_id' => $deviceId,
            'outlet_id' => $outletId,
        ])->assertOk()
            ->assertJsonPath('data.cashier.email', 'cashier@demo.nojpos.test');
    }

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

    public function test_customers_can_be_created_searched_group_filtered_and_are_tenant_scoped(): void
    {
        $businessA = $this->business('A');
        $cashierA = $this->user($businessA, 'cashier-a@example.test', 'cashier');
        $tokenA = $this->tokenFor($cashierA);

        $businessB = $this->business('B');
        $this->customer($businessB, 'Budi Rahasia', '0812999999', 'VIP');

        $created = $this->withToken($tokenA)->postJson('/api/v1/customers', [
            'name' => 'Ani Pelanggan',
            'phone' => '0812345678',
            'group' => 'VIP',
        ])->assertCreated()
            ->assertJsonPath('data.name', 'Ani Pelanggan')
            ->assertJsonPath('data.business_id', $businessA);

        $this->assertDatabaseHas('customers', [
            'id' => $created->json('data.id'),
            'business_id' => $businessA,
            'name' => 'Ani Pelanggan',
            'phone' => '0812345678',
            'group' => 'VIP',
        ]);

        $this->withToken($tokenA)->getJson('/api/v1/customers?search=081234&group=VIP')
            ->assertOk()
            ->assertJsonCount(1, 'data.customers')
            ->assertJsonPath('data.customers.0.name', 'Ani Pelanggan');
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

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson("/api/v1/shifts/{$shiftId}/cash-movements", [
            'type' => 'cash_in',
            'amount' => 50000,
            'reason' => 'Tambah modal',
        ])->assertCreated();

        $close = $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson("/api/v1/shifts/{$shiftId}/close", [
            'actual_cash' => 550000,
            'pin' => '1234',
        ])->assertOk();

        $close->assertJsonPath('data.expected_cash', 550000)
            ->assertJsonPath('data.cash_difference', 0)
            ->assertJsonPath('data.status', 'closed');

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson("/api/v1/shifts/{$shiftId}/close", [
            'actual_cash' => 550000,
            'pin' => '1234',
        ])->assertStatus(422)
            ->assertJsonPath('error.code', 'SHIFT_ALREADY_CLOSED');

        $this->assertDatabaseHas('audit_logs', [
            'business_id' => $businessId,
            'action' => 'shift.close',
        ]);
    }

    public function test_cash_movement_requires_reason_and_retries_with_one_movement(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 100000,
        ])->json('data.id');

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson("/api/v1/shifts/{$shiftId}/cash-movements", [
            'type' => 'cash_out',
            'amount' => 10000,
        ])->assertStatus(422)
            ->assertJsonPath('error.code', 'VALIDATION_ERROR');

        $key = (string) Str::uuid();
        $created = $this->withToken($token)->withHeaders([
            'Idempotency-Key' => $key,
        ])->postJson("/api/v1/shifts/{$shiftId}/cash-movements", [
            'type' => 'cash_out',
            'amount' => 10000,
            'reason' => 'Beli es batu',
        ])->assertCreated()
            ->assertJsonPath('data.reason', 'Beli es batu');

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => $key,
        ])->postJson("/api/v1/shifts/{$shiftId}/cash-movements", [
            'type' => 'cash_out',
            'amount' => 10000,
            'reason' => 'Beli es batu',
        ])->assertCreated()
            ->assertExactJson($created->json());

        $this->assertSame(1, DB::table('cash_movements')
            ->where('business_id', $businessId)
            ->where('shift_id', $shiftId)
            ->count());
        $this->assertDatabaseHas('audit_logs', [
            'business_id' => $businessId,
            'action' => 'shift.cash_movement',
        ]);
    }

    public function test_close_shift_requires_variance_reason_and_preserves_first_close_report(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 100000,
        ])->json('data.id');

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson("/api/v1/shifts/{$shiftId}/close", [
            'actual_cash' => 99000,
            'pin' => '1234',
        ])->assertStatus(422)
            ->assertJsonPath('error.code', 'SHIFT_VARIANCE_REASON_REQUIRED');

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson("/api/v1/shifts/{$shiftId}/close", [
            'actual_cash' => 100000,
        ])->assertStatus(422)
            ->assertJsonPath('error.code', 'VALIDATION_ERROR');

        $close = $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson("/api/v1/shifts/{$shiftId}/close", [
            'actual_cash' => 99000,
            'pin' => '1234',
            'variance_reason' => 'Uang kembalian kurang Rp1.000',
        ])->assertOk()
            ->assertJsonPath('data.expected_cash', 100000)
            ->assertJsonPath('data.actual_cash', 99000)
            ->assertJsonPath('data.cash_difference', -1000)
            ->assertJsonPath('data.variance_reason', 'Uang kembalian kurang Rp1.000');

        $firstSnapshot = DB::table('shift_sessions')->where('id', $shiftId)->value('close_report_snapshot');

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson("/api/v1/shifts/{$shiftId}/close", [
            'actual_cash' => 1000000,
            'pin' => '1234',
            'variance_reason' => 'Jangan overwrite',
        ])->assertStatus(422)
            ->assertJsonPath('error.code', 'SHIFT_ALREADY_CLOSED');

        $this->assertDatabaseHas('shift_sessions', [
            'id' => $shiftId,
            'actual_cash' => 99000,
            'cash_difference' => -1000,
            'variance_reason' => 'Uang kembalian kurang Rp1.000',
        ]);
        $this->assertSame($firstSnapshot, DB::table('shift_sessions')->where('id', $shiftId)->value('close_report_snapshot'));
        $this->assertSame($close->json('data.close_report.expected_cash'), 100000);
    }

    public function test_close_shift_rolls_back_status_and_report_when_audit_fails(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 100000,
        ])->json('data.id');

        DB::unprepared(<<<'SQL'
CREATE TRIGGER fail_shift_close_audit_insert
BEFORE INSERT ON audit_logs
WHEN NEW.action = 'shift.close'
BEGIN
    SELECT RAISE(ABORT, 'forced shift close audit failure');
END;
SQL);

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson("/api/v1/shifts/{$shiftId}/close", [
            'actual_cash' => 100000,
            'pin' => '1234',
        ])->assertStatus(500);

        $this->assertDatabaseHas('shift_sessions', [
            'id' => $shiftId,
            'status' => 'open',
            'expected_cash' => null,
            'actual_cash' => null,
            'cash_difference' => null,
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
                ['method' => 'cash', 'amount' => 40293],
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

        $this->assertSame(1, DB::table('transactions')->where('business_id', $businessId)->count());
        $this->assertSame(1, DB::table('transaction_items')->where('business_id', $businessId)->count());
        $this->assertSame(1, DB::table('payments')->where('business_id', $businessId)->count());

        $payload['cart_discount'] = 2000;
        $this->withToken($token)->withHeaders(['Idempotency-Key' => $key])
            ->postJson('/api/v1/transactions', $payload)
            ->assertStatus(409)
            ->assertJsonPath('error.code', 'IDEMPOTENCY_MISMATCH');
    }

    public function test_checkout_recovery_lookup_returns_stored_result_for_the_same_tenant_key(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $key = (string) Str::uuid();
        $transactionId = (string) Str::uuid();

        DB::table('idempotency_keys')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'endpoint' => 'POST api/v1/transactions',
            'key' => $key,
            'request_hash' => hash('sha256', 'checkout-recovery'),
            'state' => 'completed',
            'reserved_at' => now(),
            'completed_at' => now(),
            'response_snapshot' => json_encode([
                'data' => [
                    'id' => $transactionId,
                    'number' => 'TRX-RECOVERY',
                    'status' => 'paid',
                    'grand_total' => 25000,
                ],
                'meta' => [],
            ]),
            'status' => 201,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        $this->withToken($token)
            ->getJson("/api/v1/transactions/recovery/{$key}")
            ->assertOk()
            ->assertJsonPath('data.status', 'completed')
            ->assertJsonPath('data.transaction.id', $transactionId);

        $this->withToken($token)
            ->getJson('/api/v1/transactions/recovery/'.Str::uuid())
            ->assertOk()
            ->assertJsonPath('data.status', 'missing');
    }

    public function test_stale_checkout_quote_is_rejected_when_outlet_tax_or_service_changes_before_commit(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $categoryId = $this->category($businessId, 'Minuman');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Es Teh', 10000);
        $this->paymentMethod($businessId, $outletId, 'cash', true);

        DB::table('outlets')->where('id', $outletId)->update([
            'service_charge_rate' => 5,
            'tax_rate' => 11,
        ]);

        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 100000,
        ])->json('data.id');

        $payload = [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'cart_discount' => 0,
            'rounding' => 0,
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 1,
                    'unit_price' => 10000,
                    'discount' => 0,
                ],
            ],
        ];

        $quote = $this->withToken($token)
            ->postJson('/api/v1/transactions/quote', $payload)
            ->assertOk();

        DB::table('outlets')->where('id', $outletId)->update([
            'service_charge_rate' => 10,
            'tax_rate' => 11,
        ]);

        $commitPayload = array_merge($payload, [
            'subtotal' => $quote->json('data.subtotal'),
            'item_discount_total' => $quote->json('data.item_discount_total'),
            'cart_discount_total' => $quote->json('data.cart_discount_total'),
            'discount_total' => $quote->json('data.discount_total'),
            'service_charge_total' => $quote->json('data.service_charge_total'),
            'tax_total' => $quote->json('data.tax_total'),
            'rounding_total' => $quote->json('data.rounding_total'),
            'grand_total' => $quote->json('data.grand_total'),
            'payments' => [
                ['method' => 'cash', 'amount' => $quote->json('data.grand_total')],
            ],
        ]);

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson('/api/v1/transactions', $commitPayload)
            ->assertStatus(422)
            ->assertJsonPath('error.code', 'TOTAL_MISMATCH')
            ->assertJsonPath('error.message', 'Checkout total does not match server quote.');

        $this->assertDatabaseMissing('transactions', [
            'business_id' => $businessId,
            'shift_id' => $shiftId,
        ]);
    }

    public function test_transaction_quote_matches_commit_for_discount_service_tax_and_rounding(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $categoryId = $this->category($businessId, 'Minuman');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Kopi Susu', 3000);
        $this->paymentMethod($businessId, $outletId, 'cash', true);

        DB::table('outlets')->where('id', $outletId)->update([
            'service_charge_rate' => 5,
            'tax_rate' => 11,
        ]);

        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 100000,
        ])->json('data.id');

        $payload = [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'cart_discount' => 500,
            'rounding' => 0,
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 2,
                    'unit_price' => 3000,
                    'discount' => 1000,
                ],
            ],
        ];

        $quote = $this->withToken($token)
            ->postJson('/api/v1/transactions/quote', $payload)
            ->assertOk()
            ->assertJsonPath('data.subtotal', 6000)
            ->assertJsonPath('data.item_discount_total', 1000)
            ->assertJsonPath('data.cart_discount_total', 500)
            ->assertJsonPath('data.discount_total', 1500)
            ->assertJsonPath('data.service_charge_total', 225)
            ->assertJsonPath('data.tax_total', 519)
            ->assertJsonPath('data.grand_total', 5244);

        $this->assertDatabaseMissing('transactions', [
            'business_id' => $businessId,
            'grand_total' => $quote->json('data.grand_total'),
        ]);

        $commitPayload = $payload;
        $commitPayload['payments'] = [
            ['method' => 'cash', 'amount' => $quote->json('data.grand_total')],
        ];

        $commit = $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson('/api/v1/transactions', $commitPayload)
            ->assertCreated()
            ->assertJsonPath('data.status', 'paid');

        $this->assertSame($quote->json('data.grand_total'), $commit->json('data.grand_total'));
        $this->assertSame($quote->json('data.service_charge_total'), $commit->json('data.service_charge_total'));
        $this->assertSame($quote->json('data.tax_total'), $commit->json('data.tax_total'));
    }

    public function test_server_owned_quote_ignores_client_price_and_rounding_and_returns_checkout_binding(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $categoryId = $this->category($businessId, 'Makanan');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Bakso', 18000);

        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 100000,
        ])->json('data.id');

        $quote = $this->withToken($token)->postJson('/api/v1/transactions/quote', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'rounding' => -9999,
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 2,
                    'unit_price' => 1,
                    'discount' => 0,
                ],
            ],
        ])->assertOk();

        $quote->assertJsonPath('data.subtotal', 36000)
            ->assertJsonPath('data.rounding_total', 0)
            ->assertJsonPath('data.grand_total', 36000)
            ->assertJsonStructure([
                'data' => [
                    'quote_id',
                    'expires_at',
                    'config_version',
                    'checkout_idempotency_key',
                    'checkout_token',
                    'items' => [
                        '*' => ['product_id', 'name', 'quantity', 'unit_price', 'discount', 'subtotal'],
                    ],
                ],
            ]);
    }

    public function test_checkout_uses_quote_reference_and_ignores_manipulated_financial_fields(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $categoryId = $this->category($businessId, 'Minuman');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Jus Alpukat', 22000);
        $this->paymentMethod($businessId, $outletId, 'cash', true);

        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 100000,
        ])->json('data.id');

        $quote = $this->withToken($token)->postJson('/api/v1/transactions/quote', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'rounding' => 5000,
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 1,
                    'unit_price' => 1,
                    'discount' => 0,
                ],
            ],
        ])->assertOk();

        $payload = [
            'quote_id' => $quote->json('data.quote_id'),
            'checkout_token' => $quote->json('data.checkout_token'),
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'subtotal' => 1,
            'rounding' => 9999,
            'rounding_total' => 9999,
            'grand_total' => 1,
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 1,
                    'unit_price' => 1,
                    'discount' => 0,
                ],
            ],
            'payments' => [
                ['method' => 'cash', 'amount' => $quote->json('data.grand_total')],
            ],
        ];

        $key = $quote->json('data.checkout_idempotency_key');
        $first = $this->withToken($token)->withHeaders(['Idempotency-Key' => $key])
            ->postJson('/api/v1/transactions', $payload)
            ->assertCreated()
            ->assertJsonPath('data.grand_total', 22000)
            ->assertJsonPath('data.items.0.unit_price', 22000)
            ->assertJsonPath('data.rounding_total', 0);

        $this->withToken($token)->withHeaders(['Idempotency-Key' => $key])
            ->postJson('/api/v1/transactions', $payload)
            ->assertCreated()
            ->assertExactJson($first->json());

        $this->assertSame(1, DB::table('transactions')->where('business_id', $businessId)->count());
    }

    public function test_stale_quote_reference_returns_conflict_without_creating_checkout_rows(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $categoryId = $this->category($businessId, 'Minuman');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Teh Tawar', 5000);
        $this->paymentMethod($businessId, $outletId, 'cash', true);

        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 100000,
        ])->json('data.id');

        $quote = $this->withToken($token)->postJson('/api/v1/transactions/quote', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 1,
                    'unit_price' => 5000,
                ],
            ],
        ])->assertOk();

        DB::table('transaction_quotes')->where('id', $quote->json('data.quote_id'))->update([
            'expires_at' => now()->subMinute(),
        ]);

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => $quote->json('data.checkout_idempotency_key'),
        ])->postJson('/api/v1/transactions', [
            'quote_id' => $quote->json('data.quote_id'),
            'checkout_token' => $quote->json('data.checkout_token'),
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 1,
                    'unit_price' => 5000,
                ],
            ],
            'payments' => [
                ['method' => 'cash', 'amount' => $quote->json('data.grand_total')],
            ],
        ])->assertStatus(409)
            ->assertJsonPath('error.code', 'QUOTE_STALE');

        $this->assertDatabaseMissing('transactions', [
            'business_id' => $businessId,
            'shift_id' => $shiftId,
        ]);
        $this->assertDatabaseMissing('payments', ['business_id' => $businessId]);
        $this->assertDatabaseMissing('stock_movements', ['business_id' => $businessId]);
        $this->assertDatabaseMissing('audit_logs', [
            'business_id' => $businessId,
            'action' => 'transaction.create',
        ]);
    }

    public function test_quote_bound_checkout_rolls_back_transaction_payment_stock_and_audit_on_write_failure(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $categoryId = $this->category($businessId, 'Makanan');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Sate Ayam', 30000);
        $this->paymentMethod($businessId, $outletId, 'cash', true);

        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 100000,
        ])->json('data.id');

        $quote = $this->withToken($token)->postJson('/api/v1/transactions/quote', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 1,
                    'unit_price' => 1,
                ],
            ],
        ])->assertOk();

        DB::unprepared(<<<'SQL'
CREATE TRIGGER fail_transaction_payment_insert
BEFORE INSERT ON payments
BEGIN
    SELECT RAISE(ABORT, 'forced payment failure');
END;
SQL);

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => $quote->json('data.checkout_idempotency_key'),
        ])->postJson('/api/v1/transactions', [
            'quote_id' => $quote->json('data.quote_id'),
            'checkout_token' => $quote->json('data.checkout_token'),
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 1,
                    'unit_price' => 1,
                ],
            ],
            'payments' => [
                ['method' => 'cash', 'amount' => $quote->json('data.grand_total')],
            ],
        ])->assertStatus(500);

        $this->assertDatabaseMissing('transactions', [
            'business_id' => $businessId,
            'shift_id' => $shiftId,
        ]);
        $this->assertDatabaseMissing('payments', ['business_id' => $businessId]);
        $this->assertDatabaseMissing('stock_movements', ['business_id' => $businessId]);
        $this->assertDatabaseMissing('audit_logs', [
            'business_id' => $businessId,
            'action' => 'transaction.create',
        ]);
    }

    public function test_transaction_enrichment_customer_served_by_and_discount_breakdown_are_stored(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $categoryId = $this->category($businessId, 'Makanan');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Nasi Campur', 50000);
        $customerId = $this->customer($businessId, 'Sari Regular', '0812111111', 'Regular');
        $servedById = $this->user($businessId, 'server@example.test', 'cashier');
        $this->paymentMethod($businessId, $outletId, 'cash', true);

        DB::table('outlets')->where('id', $outletId)->update([
            'service_charge_rate' => 10,
            'tax_rate' => 11,
        ]);

        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 100000,
        ])->json('data.id');

        $response = $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson('/api/v1/transactions', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'customer_id' => $customerId,
            'served_by' => $servedById,
            'notes' => 'Meja 7',
            'cart_discount' => 3000,
            'rounding' => 0,
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 2,
                    'unit_price' => 50000,
                    'discount' => 10000,
                ],
            ],
            'payments' => [
                ['method' => 'cash', 'amount' => 106227],
            ],
        ])->assertCreated();

        $response->assertJsonPath('data.subtotal', 100000)
            ->assertJsonPath('data.item_discount_total', 10000)
            ->assertJsonPath('data.cart_discount_total', 3000)
            ->assertJsonPath('data.discount_total', 13000)
            ->assertJsonPath('data.service_charge_total', 8700)
            ->assertJsonPath('data.tax_total', 10527)
            ->assertJsonPath('data.grand_total', 106227)
            ->assertJsonPath('data.customer_id', $customerId)
            ->assertJsonPath('data.served_by', $servedById)
            ->assertJsonPath('data.notes', 'Meja 7');

        $this->assertDatabaseHas('transactions', [
            'id' => $response->json('data.id'),
            'business_id' => $businessId,
            'customer_id' => $customerId,
            'served_by' => $servedById,
            'item_discount_total' => 10000,
            'cart_discount_total' => 3000,
            'grand_total' => 106227,
        ]);
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

    public function test_split_payment_lifecycle_confirmation_and_cash_only_expected_cash(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $categoryId = $this->category($businessId, 'Makanan');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Paket Hemat', 100000);
        $this->paymentMethod($businessId, $outletId, 'Tunai', true);
        $this->paymentMethod($businessId, $outletId, 'QRIS Statis', false);

        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 100000,
        ])->json('data.id');

        $transaction = $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson('/api/v1/transactions', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 1,
                    'unit_price' => 100000,
                ],
            ],
            'payments' => [
                ['method' => 'Tunai', 'amount' => 40000],
                ['method' => 'QRIS Statis', 'amount' => 60000, 'reference' => 'QR-123'],
            ],
        ])->assertCreated()
            ->assertJsonPath('data.status', 'partial');

        $transactionId = $transaction->json('data.id');
        $qrisPaymentId = DB::table('payments')
            ->where('transaction_id', $transactionId)
            ->where('method', 'QRIS Statis')
            ->value('id');

        $this->assertDatabaseHas('payments', [
            'transaction_id' => $transactionId,
            'method' => 'Tunai',
            'amount' => 40000,
            'status' => 'confirmed',
            'is_cash' => true,
        ]);
        $this->assertDatabaseHas('payments', [
            'transaction_id' => $transactionId,
            'method' => 'QRIS Statis',
            'amount' => 60000,
            'status' => 'pending',
            'is_cash' => false,
            'reference' => 'QR-123',
        ]);

        $confirmKey = (string) Str::uuid();
        $confirmed = $this->withToken($token)->withHeaders([
            'Idempotency-Key' => $confirmKey,
        ])->postJson('/api/v1/payments', [
            'payment_id' => $qrisPaymentId,
            'confirm' => true,
        ])->assertOk()
            ->assertJsonPath('data.status', 'confirmed')
            ->assertJsonPath('data.confirmed_by', $cashierId);

        $this->assertDatabaseHas('transactions', [
            'id' => $transactionId,
            'status' => 'paid',
        ]);

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => $confirmKey,
        ])->postJson('/api/v1/payments', [
            'payment_id' => $qrisPaymentId,
            'confirm' => true,
        ])->assertOk()
            ->assertExactJson($confirmed->json());

        $close = $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson("/api/v1/shifts/{$shiftId}/close", [
            'actual_cash' => 140000,
            'pin' => '1234',
        ])->assertOk();

        $close->assertJsonPath('data.expected_cash', 140000)
            ->assertJsonPath('data.cash_difference', 0)
            ->assertJsonPath('data.payment_totals.0.method', 'QRIS Statis')
            ->assertJsonPath('data.payment_totals.0.amount', 60000)
            ->assertJsonPath('data.payment_totals.1.method', 'Tunai')
            ->assertJsonPath('data.payment_totals.1.amount', 40000);
    }

    public function test_payment_cannot_be_added_to_voided_transaction(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $this->paymentMethod($businessId, $outletId, 'cash', true);
        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 0,
        ])->json('data.id');
        $transactionId = $this->transaction($businessId, $outletId, $deviceId, $cashierId, $shiftId, 50000);

        DB::table('transactions')->where('id', $transactionId)->update(['status' => 'voided']);

        $this->withToken($token)->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/payments', [
                'transaction_id' => $transactionId,
                'method' => 'cash',
                'amount' => 50000,
            ])->assertStatus(422)
            ->assertJsonPath('error.code', 'PAYMENT_STATE_INVALID');

        $this->assertDatabaseMissing('payments', [
            'business_id' => $businessId,
            'transaction_id' => $transactionId,
        ]);
        $this->assertDatabaseMissing('audit_logs', [
            'business_id' => $businessId,
            'action' => 'payment.create',
        ]);
    }

    public function test_payment_cannot_overpay_transaction_remaining_total(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $this->paymentMethod($businessId, $outletId, 'cash', true);
        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 0,
        ])->json('data.id');
        $transactionId = $this->transaction($businessId, $outletId, $deviceId, $cashierId, $shiftId, 50000);

        $this->withToken($token)->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/payments', [
                'transaction_id' => $transactionId,
                'method' => 'cash',
                'amount' => 50001,
            ])->assertStatus(422)
            ->assertJsonPath('error.code', 'PAYMENT_OVERPAY');

        $this->assertDatabaseMissing('payments', [
            'business_id' => $businessId,
            'transaction_id' => $transactionId,
        ]);
    }

    public function test_checkout_cannot_create_overpaid_payment_allocation(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $categoryId = $this->category($businessId, 'Makanan');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Nasi Goreng', 50000);
        $this->paymentMethod($businessId, $outletId, 'cash', true);
        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 0,
        ])->json('data.id');

        $this->withToken($token)->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/transactions', [
                'outlet_id' => $outletId,
                'device_id' => $deviceId,
                'cashier_id' => $cashierId,
                'shift_id' => $shiftId,
                'items' => [
                    [
                        'product_id' => $productId,
                        'quantity' => 1,
                        'unit_price' => 50000,
                    ],
                ],
                'payments' => [
                    ['method' => 'cash', 'amount' => 50001],
                ],
            ])->assertStatus(422)
            ->assertJsonPath('error.code', 'PAYMENT_OVERPAY');

        $this->assertDatabaseMissing('transactions', [
            'business_id' => $businessId,
            'grand_total' => 50000,
        ]);
        $this->assertDatabaseMissing('payments', [
            'business_id' => $businessId,
            'amount' => 50001,
        ]);
    }

    public function test_repeated_payment_confirmation_settles_once_without_duplicate_audit(): void
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

        $paymentId = $this->withToken($token)->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/payments', [
                'transaction_id' => $transactionId,
                'method' => 'transfer',
                'amount' => 50000,
            ])->assertCreated()
            ->json('data.id');

        $this->withToken($token)->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/payments', [
                'payment_id' => $paymentId,
                'confirm' => true,
            ])->assertOk()
            ->assertJsonPath('data.status', 'confirmed');

        $this->withToken($token)->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/payments', [
                'payment_id' => $paymentId,
                'confirm' => true,
            ])->assertOk()
            ->assertJsonPath('data.status', 'confirmed');

        $this->assertSame(1, DB::table('payments')
            ->where('id', $paymentId)
            ->where('status', 'confirmed')
            ->count());
        $this->assertSame(1, DB::table('audit_logs')
            ->where('business_id', $businessId)
            ->where('action', 'payment.confirm')
            ->where('entity_id', $paymentId)
            ->count());
    }

    public function test_payment_confirmation_rolls_back_payment_and_transaction_status_together(): void
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
        $paymentId = $this->withToken($token)->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/payments', [
                'transaction_id' => $transactionId,
                'method' => 'transfer',
                'amount' => 50000,
            ])->assertCreated()
            ->json('data.id');

        DB::unprepared(<<<'SQL'
CREATE TRIGGER fail_payment_transaction_update
BEFORE UPDATE OF status ON transactions
WHEN NEW.status = 'paid'
BEGIN
    SELECT RAISE(ABORT, 'forced transaction payment state failure');
END;
SQL);

        $this->withToken($token)->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/payments', [
                'payment_id' => $paymentId,
                'confirm' => true,
            ])->assertStatus(500);

        $this->assertDatabaseHas('payments', [
            'id' => $paymentId,
            'status' => 'pending',
            'confirmed_by' => null,
        ]);
        $this->assertDatabaseHas('transactions', [
            'id' => $transactionId,
            'status' => 'unpaid',
        ]);
        $this->assertDatabaseMissing('audit_logs', [
            'business_id' => $businessId,
            'action' => 'payment.confirm',
            'entity_id' => $paymentId,
        ]);
    }

    public function test_full_void_reverses_stock_cash_sets_voided_and_is_idempotent(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext('admin');
        $categoryId = $this->category($businessId, 'Makanan');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Sate Ayam', 50000);
        $this->paymentMethod($businessId, $outletId, 'cash', true);

        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 100000,
        ])->json('data.id');

        $transaction = $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson('/api/v1/transactions', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 2,
                    'unit_price' => 50000,
                ],
            ],
            'payments' => [
                ['method' => 'cash', 'amount' => 100000],
            ],
        ])->assertCreated();

        $key = (string) Str::uuid();
        $void = $this->withToken($token)->withHeaders([
            'Idempotency-Key' => $key,
        ])->postJson('/api/v1/voids', [
            'transaction_id' => $transaction->json('data.id'),
            'shift_id' => $shiftId,
            'reason' => 'Salah input item',
        ])->assertOk()
            ->assertJsonPath('data.transaction_id', $transaction->json('data.id'))
            ->assertJsonPath('data.status', 'voided');

        $this->assertDatabaseHas('transactions', [
            'id' => $transaction->json('data.id'),
            'status' => 'voided',
        ]);
        $this->assertDatabaseHas('stock_movements', [
            'business_id' => $businessId,
            'product_id' => $productId,
            'transaction_id' => $transaction->json('data.id'),
            'type' => 'sale',
            'quantity_delta' => -2,
        ]);
        $this->assertDatabaseHas('stock_movements', [
            'business_id' => $businessId,
            'product_id' => $productId,
            'transaction_id' => $transaction->json('data.id'),
            'type' => 'void',
            'quantity_delta' => 2,
        ]);
        $this->assertDatabaseHas('cash_movements', [
            'business_id' => $businessId,
            'shift_id' => $shiftId,
            'actor_id' => $cashierId,
            'type' => 'cash_out',
            'amount' => 100000,
            'reason' => 'Void: Salah input item',
        ]);
        $this->assertDatabaseHas('audit_logs', [
            'business_id' => $businessId,
            'actor_id' => $cashierId,
            'action' => 'void',
            'entity_type' => 'transaction',
            'entity_id' => $transaction->json('data.id'),
        ]);

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => $key,
        ])->postJson('/api/v1/voids', [
            'transaction_id' => $transaction->json('data.id'),
            'shift_id' => $shiftId,
            'reason' => 'Salah input item',
        ])->assertOk()
            ->assertExactJson($void->json());

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson("/api/v1/shifts/{$shiftId}/close", [
            'actual_cash' => 100000,
            'pin' => '1234',
        ])->assertOk()
            ->assertJsonPath('data.expected_cash', 100000)
            ->assertJsonPath('data.cash_difference', 0);
    }

    public function test_void_cross_shift_and_missing_reason_are_rejected(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext('supervisor');
        $otherDeviceId = $this->device($businessId, $outletId, 'tablet-2');
        $categoryId = $this->category($businessId, 'Makanan');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Bakso', 25000);
        $this->paymentMethod($businessId, $outletId, 'cash', true);

        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 0,
        ])->json('data.id');
        $otherShiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $otherDeviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 0,
        ])->json('data.id');

        $transaction = $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson('/api/v1/transactions', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 1,
                    'unit_price' => 25000,
                ],
            ],
            'payments' => [
                ['method' => 'cash', 'amount' => 25000],
            ],
        ])->assertCreated();

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson('/api/v1/voids', [
            'transaction_id' => $transaction->json('data.id'),
            'shift_id' => $shiftId,
        ])->assertStatus(422)
            ->assertJsonPath('error.code', 'VALIDATION_ERROR')
            ->assertJsonPath('error.details.reason.0', 'The reason field is required.');

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson('/api/v1/voids', [
            'transaction_id' => $transaction->json('data.id'),
            'shift_id' => $otherShiftId,
            'reason' => 'Shift salah',
        ])->assertStatus(422)
            ->assertJsonPath('error.code', 'VOID_SHIFT_MISMATCH');
    }

    public function test_held_transaction_is_saved_and_listed_by_status_filter(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $categoryId = $this->category($businessId, 'Makanan');
        $heldProductId = $this->product($businessId, $outletId, $categoryId, 'Nasi Bungkus', 15000);
        $paidProductId = $this->product($businessId, $outletId, $categoryId, 'Es Jeruk', 10000);
        $this->paymentMethod($businessId, $outletId, 'cash', true);

        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 0,
        ])->json('data.id');

        $held = $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson('/api/v1/transactions', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'status' => 'held',
            'notes' => 'Jadikan invoice',
            'items' => [
                [
                    'product_id' => $heldProductId,
                    'quantity' => 1,
                    'unit_price' => 15000,
                ],
            ],
        ])->assertCreated()
            ->assertJsonPath('data.status', 'held');

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson('/api/v1/transactions', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'items' => [
                [
                    'product_id' => $paidProductId,
                    'quantity' => 1,
                    'unit_price' => 10000,
                ],
            ],
            'payments' => [
                ['method' => 'cash', 'amount' => 10000],
            ],
        ])->assertCreated();

        $this->withToken($token)->getJson('/api/v1/transactions?status=held')
            ->assertOk()
            ->assertJsonCount(1, 'data.transactions')
            ->assertJsonPath('data.transactions.0.id', $held->json('data.id'))
            ->assertJsonPath('data.transactions.0.status', 'held');
    }

    public function test_cashier_sales_summary_is_limited_to_current_shift_day(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $otherDeviceId = $this->device($businessId, $outletId, 'tablet-2');
        $otherCashierId = $this->user($businessId, 'other-cashier@example.test', 'cashier');

        $shiftId = $this->shift($businessId, $outletId, $deviceId, $cashierId, 'open', now());
        $otherShiftId = $this->shift($businessId, $outletId, $otherDeviceId, $otherCashierId, 'open', now());
        $yesterdayShiftId = $this->shift($businessId, $outletId, $deviceId, $cashierId, 'closed', now()->subDay());

        $transactionId = $this->reportTransaction($businessId, $outletId, $deviceId, $cashierId, $shiftId, 100000, 'paid', 10000);
        $this->reportPayment($businessId, $transactionId, 'Tunai', 40000, true);
        $this->reportPayment($businessId, $transactionId, 'QRIS Statis', 60000, false);

        $otherTransactionId = $this->reportTransaction($businessId, $outletId, $otherDeviceId, $otherCashierId, $otherShiftId, 70000);
        $this->reportPayment($businessId, $otherTransactionId, 'Tunai', 70000, true);

        $oldTransactionId = $this->reportTransaction(
            $businessId,
            $outletId,
            $deviceId,
            $cashierId,
            $yesterdayShiftId,
            50000,
            'paid',
            0,
            now()->subDay(),
        );
        $this->reportPayment($businessId, $oldTransactionId, 'Tunai', 50000, true, 'confirmed', now()->subDay());

        $response = $this->withToken($token)->getJson('/api/v1/reports/sales-summary?date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonPath('data.total_sales', 100000)
            ->assertJsonPath('data.transaction_count', 1)
            ->assertJsonPath('data.total_discount', 10000)
            ->assertJsonPath('data.total_void', 0)
            ->assertJsonPath('data.average_transaction_value', 100000)
            ->assertJsonPath('data.payment_totals.0.method', 'QRIS Statis')
            ->assertJsonPath('data.payment_totals.0.amount', 60000)
            ->assertJsonPath('data.payment_totals.1.method', 'Tunai')
            ->assertJsonPath('data.payment_totals.1.amount', 40000);

        $this->assertIsInt($response->json('data.total_sales'));
        $this->assertIsInt($response->json('data.payment_totals.0.amount'));
        $this->assertIsInt($response->json('data.chart.0.amount'));
    }

    public function test_owner_sales_summary_aggregates_business_payment_totals_voids_and_chart(): void
    {
        [$businessId, $outletId, $deviceId, $ownerId, $token] = $this->cashierContext('owner');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId, 'open', now());

        $cashTransactionId = $this->reportTransaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 100000, 'paid', 5000);
        $this->reportPayment($businessId, $cashTransactionId, 'Tunai', 100000, true);

        $transferTransactionId = $this->reportTransaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 50000);
        $this->reportPayment($businessId, $transferTransactionId, 'Transfer', 50000, false);

        $voidTransactionId = $this->reportTransaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 30000, 'voided');
        $this->reportPayment($businessId, $voidTransactionId, 'Tunai', 30000, true);

        [$otherBusinessId, $otherOutletId, $otherDeviceId, $otherCashierId] = $this->cashierContext();
        $otherShiftId = $this->shift($otherBusinessId, $otherOutletId, $otherDeviceId, $otherCashierId, 'open', now());
        $otherTransactionId = $this->reportTransaction($otherBusinessId, $otherOutletId, $otherDeviceId, $otherCashierId, $otherShiftId, 999000);
        $this->reportPayment($otherBusinessId, $otherTransactionId, 'Tunai', 999000, true);

        $response = $this->withToken($token)->getJson('/api/v1/reports/sales-summary?date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonPath('data.total_sales', 150000)
            ->assertJsonPath('data.transaction_count', 2)
            ->assertJsonPath('data.total_discount', 5000)
            ->assertJsonPath('data.total_void', 30000)
            ->assertJsonPath('data.average_transaction_value', 75000)
            ->assertJsonPath('data.payment_totals.0.method', 'Transfer')
            ->assertJsonPath('data.payment_totals.0.amount', 50000)
            ->assertJsonPath('data.payment_totals.1.method', 'Tunai')
            ->assertJsonPath('data.payment_totals.1.amount', 100000);

        $this->assertIsInt($response->json('data.total_sales'));
        $this->assertIsInt($response->json('data.total_void'));
        $this->assertIsInt($response->json('data.chart.0.amount'));
    }

    public function test_sales_summary_uses_the_outlet_local_day_boundary(): void
    {
        [$businessId, $outletId, $deviceId, $ownerId, $token] = $this->cashierContext('owner');
        DB::table('outlets')->where('id', $outletId)->update(['timezone' => 'Asia/Jakarta']);
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId, 'open', CarbonImmutable::parse('2025-12-31T17:30:00Z'));

        $inside = $this->reportTransaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 10000, 'paid', 0, CarbonImmutable::parse('2025-12-31T17:30:00Z'));
        $outside = $this->reportTransaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 9000, 'paid', 0, CarbonImmutable::parse('2025-12-31T16:30:00Z'));
        $this->reportPayment($businessId, $inside, 'Tunai', 10000, true, 'confirmed', CarbonImmutable::parse('2025-12-31T17:30:00Z'));
        $this->reportPayment($businessId, $outside, 'Tunai', 9000, true, 'confirmed', CarbonImmutable::parse('2025-12-31T16:30:00Z'));

        $this->withToken($token)->getJson('/api/v1/reports/sales-summary?outlet_id='.$outletId.'&date=2026-01-01')
            ->assertOk()
            ->assertJsonPath('data.total_sales', 10000)
            ->assertJsonPath('data.transaction_count', 1);
    }

    public function test_attendance_filter_uses_the_outlet_local_day_boundary(): void
    {
        [$businessId, $outletId, , $staffId, $token] = $this->cashierContext();
        DB::table('outlets')->where('id', $outletId)->update(['timezone' => 'Asia/Jakarta']);
        $insideId = (string) Str::uuid();
        $outsideId = (string) Str::uuid();

        foreach ([
            [$insideId, CarbonImmutable::parse('2025-12-31T17:30:00Z')],
            [$outsideId, CarbonImmutable::parse('2025-12-31T16:30:00Z')],
        ] as [$id, $clockInAt]) {
            DB::table('attendance_records')->insert([
                'id' => $id,
                'business_id' => $businessId,
                'outlet_id' => $outletId,
                'staff_id' => $staffId,
                'created_by' => $staffId,
                'clock_in_at' => $clockInAt,
                'clock_out_at' => null,
                'created_at' => $clockInAt,
                'updated_at' => $clockInAt,
            ]);
        }

        $this->withToken($token)->getJson('/api/v1/attendance?outlet_id='.$outletId.'&date=2026-01-01')
            ->assertOk()
            ->assertJsonCount(1, 'data.attendance')
            ->assertJsonPath('data.attendance.0.id', $insideId);
    }

    public function test_generated_purchase_number_uses_the_outlet_local_date(): void
    {
        CarbonImmutable::setTestNow(CarbonImmutable::parse('2025-12-31T17:30:45Z'));

        try {
            [$businessId, $outletId, , , $token] = $this->cashierContext('owner');
            DB::table('outlets')->where('id', $outletId)->update(['timezone' => 'Asia/Jakarta']);
            $categoryId = $this->category($businessId, 'Minuman');
            $productId = $this->product($businessId, $outletId, $categoryId, 'Kopi', 10000);

            $this->withToken($token)->postJson('/api/v1/inventory/purchases', [
                'outlet_id' => $outletId,
                'items' => [['product_id' => $productId, 'quantity' => 1, 'unit_cost' => 5000]],
            ])->assertCreated()->assertJsonPath('data.number', 'PO-20260101003045');
        } finally {
            CarbonImmutable::setTestNow();
        }
    }

    public function test_generated_transaction_number_uses_the_outlet_local_date(): void
    {
        CarbonImmutable::setTestNow(CarbonImmutable::parse('2025-12-31T17:30:45Z'));

        try {
            [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
            DB::table('outlets')->where('id', $outletId)->update(['timezone' => 'Asia/Jakarta']);
            $categoryId = $this->category($businessId, 'Makanan');
            $productId = $this->product($businessId, $outletId, $categoryId, 'Nasi', 10000);
            $this->paymentMethod($businessId, $outletId, 'cash', true);
            $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
                'outlet_id' => $outletId,
                'device_id' => $deviceId,
                'cashier_id' => $cashierId,
                'opening_cash' => 0,
            ])->json('data.id');

            $response = $this->withToken($token)->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
                ->postJson('/api/v1/transactions', [
                    'outlet_id' => $outletId,
                    'device_id' => $deviceId,
                    'cashier_id' => $cashierId,
                    'shift_id' => $shiftId,
                    'items' => [['product_id' => $productId, 'quantity' => 1, 'unit_price' => 10000]],
                    'payments' => [['method' => 'cash', 'amount' => 10000]],
                ])
                ->assertCreated();

            $this->assertStringStartsWith('TRX-20260101003045-', $response->json('data.number'));
        } finally {
            CarbonImmutable::setTestNow();
        }
    }

    public function test_purchase_creates_invoice_stock_movements_and_inventory_balance(): void
    {
        [$businessId, $outletId, $deviceId, $ownerId, $token] = $this->cashierContext('owner');
        $categoryId = $this->category($businessId, 'Minuman');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Kopi Susu Botol', 18000);

        $purchase = $this->withToken($token)->postJson('/api/v1/inventory/purchases', [
            'outlet_id' => $outletId,
            'number' => 'INV-PO-001',
            'supplier_name' => 'Supplier Kopi',
            'notes' => 'Restock awal',
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 12,
                    'unit_cost' => 10000,
                ],
            ],
        ])->assertCreated()
            ->assertJsonPath('data.business_id', $businessId)
            ->assertJsonPath('data.outlet_id', $outletId)
            ->assertJsonPath('data.number', 'INV-PO-001')
            ->assertJsonPath('data.total', 120000)
            ->assertJsonPath('data.items.0.product_id', $productId)
            ->assertJsonPath('data.items.0.quantity', 12);

        $this->assertDatabaseHas('inventory_purchases', [
            'id' => $purchase->json('data.id'),
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'number' => 'INV-PO-001',
            'total' => 120000,
        ]);
        $this->assertDatabaseHas('stock_movements', [
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'product_id' => $productId,
            'purchase_id' => $purchase->json('data.id'),
            'type' => 'purchase',
            'quantity_delta' => 12,
        ]);

        $this->withToken($token)->getJson('/api/v1/inventory?outlet_id='.$outletId.'&type=purchase')
            ->assertOk()
            ->assertJsonPath('data.items.0.product_id', $productId)
            ->assertJsonPath('data.items.0.stock_on_hand', 12)
            ->assertJsonPath('data.items.0.warning', null)
            ->assertJsonPath('data.movements.0.type', 'purchase')
            ->assertJsonPath('data.movements.0.quantity_delta', 12);

        $otherBusinessId = $this->business('Bisnis Lain');
        $otherOutletId = $this->outlet($otherBusinessId);
        $this->withToken($token)->getJson('/api/v1/inventory?outlet_id='.$otherOutletId)
            ->assertStatus(403)
            ->assertJsonPath('error.code', 'FORBIDDEN');
    }

    public function test_negative_stock_is_allowed_and_reported_with_warning(): void
    {
        [$businessId, $outletId, $deviceId, $cashierId, $token] = $this->cashierContext();
        $categoryId = $this->category($businessId, 'Makanan');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Roti Bakar', 25000);
        $this->paymentMethod($businessId, $outletId, 'cash', true);

        $shiftId = $this->withToken($token)->postJson('/api/v1/shifts/open', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'opening_cash' => 0,
        ])->json('data.id');

        $this->withToken($token)->withHeaders([
            'Idempotency-Key' => (string) Str::uuid(),
        ])->postJson('/api/v1/transactions', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'items' => [
                [
                    'product_id' => $productId,
                    'quantity' => 3,
                    'unit_price' => 25000,
                ],
            ],
            'payments' => [
                ['method' => 'cash', 'amount' => 75000],
            ],
        ])->assertCreated()
            ->assertJsonPath('data.status', 'paid');

        $this->withToken($token)->getJson('/api/v1/inventory?outlet_id='.$outletId.'&search=Roti')
            ->assertOk()
            ->assertJsonPath('data.items.0.product_id', $productId)
            ->assertJsonPath('data.items.0.stock_on_hand', -3)
            ->assertJsonPath('data.items.0.warning', 'NEGATIVE_STOCK_ALLOWED');
    }

    public function test_owner_can_manage_products_and_categories_while_cashier_is_read_only(): void
    {
        [$businessId, $outletId, $deviceId, $ownerId, $ownerToken] = $this->cashierContext('admin');
        $cashierId = $this->user($businessId, 'readonly-cashier@example.test', 'cashier');
        $cashierToken = $this->tokenFor($cashierId);

        $category = $this->withToken($ownerToken)->postJson('/api/v1/categories', [
            'name' => 'Snack',
        ])->assertCreated()
            ->assertJsonPath('data.business_id', $businessId)
            ->assertJsonPath('data.name', 'Snack');

        $this->withToken($ownerToken)->putJson('/api/v1/categories/'.$category->json('data.id'), [
            'name' => 'Camilan',
        ])->assertOk()
            ->assertJsonPath('data.name', 'Camilan');

        $product = $this->withToken($ownerToken)->postJson('/api/v1/products', [
            'outlet_id' => $outletId,
            'product_category_id' => $category->json('data.id'),
            'name' => 'Keripik Singkong',
            'barcode' => '899100000001',
            'price' => 15000,
            'track_stock' => true,
        ])->assertCreated()
            ->assertJsonPath('data.business_id', $businessId)
            ->assertJsonPath('data.outlet_id', $outletId)
            ->assertJsonPath('data.name', 'Keripik Singkong')
            ->assertJsonPath('data.price', 15000);

        $this->assertIsInt($product->json('data.price'));

        $this->withToken($ownerToken)->putJson('/api/v1/products/'.$product->json('data.id'), [
            'outlet_id' => $outletId,
            'product_category_id' => $category->json('data.id'),
            'name' => 'Keripik Balado',
            'barcode' => '899100000002',
            'price' => 17000,
            'track_stock' => false,
        ])->assertOk()
            ->assertJsonPath('data.name', 'Keripik Balado')
            ->assertJsonPath('data.price', 17000)
            ->assertJsonPath('data.track_stock', false);

        $this->withToken($cashierToken)->getJson('/api/v1/products?search=Balado')
            ->assertOk()
            ->assertJsonPath('data.products.0.id', $product->json('data.id'));

        $this->app['auth']->forgetGuards();
        $this->withToken($cashierToken)->postJson('/api/v1/products', [
            'outlet_id' => $outletId,
            'product_category_id' => $category->json('data.id'),
            'name' => 'Tidak Boleh',
            'price' => 1000,
        ])->assertStatus(403)
            ->assertJsonPath('error.code', 'FORBIDDEN');

        $this->app['auth']->forgetGuards();
        $this->withToken($cashierToken)->putJson('/api/v1/categories/'.$category->json('data.id'), [
            'name' => 'Tidak Boleh',
        ])->assertStatus(403)
            ->assertJsonPath('error.code', 'FORBIDDEN');
    }

    public function test_staff_list_is_scoped_to_business_and_outlet_payload_has_receipt_config(): void
    {
        [$businessId, $outletId, $deviceId, $ownerId, $token] = $this->cashierContext('owner');
        $cashierId = $this->user($businessId, 'staff-cashier@example.test', 'cashier', '2468');
        $otherBusinessId = $this->business('Tenant Lain');
        $otherStaffId = $this->user($otherBusinessId, 'outside-staff@example.test', 'cashier', '1357');

        DB::table('outlets')->where('id', $outletId)->update([
            'receipt_paper_width' => '80mm',
            'receipt_header_name' => 'Kedai Test',
            'receipt_header_address' => 'Jl. Mawar 10',
            'receipt_footer_note' => 'Terima kasih',
            'receipt_show_logo' => true,
            'receipt_show_qris_info' => false,
        ]);

        $this->withToken($token)->getJson('/api/v1/staff?outlet_id='.$outletId)
            ->assertOk()
            ->assertJsonPath('data.staff.0.business_id', $businessId)
            ->assertJsonFragment(['id' => $cashierId])
            ->assertJsonMissing(['id' => $otherStaffId]);

        $this->withToken($token)->getJson('/api/v1/me')
            ->assertOk()
            ->assertJsonPath('data.outlets.0.receipt_config.paper_width', '80mm')
            ->assertJsonPath('data.outlets.0.receipt_config.header_name', 'Kedai Test')
            ->assertJsonPath('data.outlets.0.receipt_config.header_address', 'Jl. Mawar 10')
            ->assertJsonPath('data.outlets.0.receipt_config.footer_note', 'Terima kasih')
            ->assertJsonPath('data.outlets.0.receipt_config.show_logo', true)
            ->assertJsonPath('data.outlets.0.receipt_config.show_qris_info', false);
    }

    public function test_attendance_clock_in_out_pin_validation_history_and_audit(): void
    {
        [$businessId, $outletId, $deviceId, $ownerId, $token] = $this->cashierContext('owner');
        $staffId = $this->user($businessId, 'attendance-cashier@example.test', 'cashier', '2468');

        $this->withToken($token)->postJson('/api/v1/attendance', [
            'outlet_id' => $outletId,
            'staff_id' => $staffId,
            'pin' => '0000',
            'action' => 'clock_in',
        ])->assertStatus(422)
            ->assertJsonPath('error.code', 'INVALID_PIN');

        $clockIn = $this->withToken($token)->postJson('/api/v1/attendance', [
            'outlet_id' => $outletId,
            'staff_id' => $staffId,
            'pin' => '2468',
            'action' => 'clock_in',
        ])->assertCreated()
            ->assertJsonPath('data.business_id', $businessId)
            ->assertJsonPath('data.outlet_id', $outletId)
            ->assertJsonPath('data.staff_id', $staffId)
            ->assertJsonPath('data.status', 'clocked_in');

        $attendanceId = $clockIn->json('data.id');
        $this->assertDatabaseHas('attendance_records', [
            'id' => $attendanceId,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'staff_id' => $staffId,
        ]);
        $this->assertDatabaseHas('audit_logs', [
            'business_id' => $businessId,
            'actor_id' => $staffId,
            'action' => 'attendance.clock_in',
            'entity_type' => 'attendance',
            'entity_id' => $attendanceId,
        ]);

        $this->withToken($token)->postJson('/api/v1/attendance', [
            'outlet_id' => $outletId,
            'staff_id' => $staffId,
            'pin' => '2468',
            'action' => 'clock_out',
        ])->assertOk()
            ->assertJsonPath('data.id', $attendanceId)
            ->assertJsonPath('data.status', 'clocked_out');

        $this->assertNotNull(DB::table('attendance_records')->where('id', $attendanceId)->value('clock_out_at'));
        $this->assertDatabaseHas('audit_logs', [
            'business_id' => $businessId,
            'actor_id' => $staffId,
            'action' => 'attendance.clock_out',
            'entity_type' => 'attendance',
            'entity_id' => $attendanceId,
        ]);

        $this->withToken($token)->getJson('/api/v1/attendance?staff_id='.$staffId.'&date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonCount(1, 'data.attendance')
            ->assertJsonPath('data.attendance.0.id', $attendanceId)
            ->assertJsonPath('data.attendance.0.status', 'clocked_out');
    }

    private function cashierContext(string $role = 'cashier'): array
    {
        $businessId = $this->business();
        $outletId = $this->outlet($businessId);
        $deviceId = $this->device($businessId, $outletId, 'tablet-1');
        $cashierId = $this->user($businessId, $role.'@example.test', $role);

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

    private function customer(string $businessId, string $name, string $phone = '', string $group = 'Tanpa Grup'): string
    {
        $id = (string) Str::uuid();
        DB::table('customers')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'name' => $name,
            'phone' => $phone,
            'group' => $group,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function shift(string $businessId, string $outletId, string $deviceId, string $cashierId, string $status, mixed $openedAt): string
    {
        $id = (string) Str::uuid();
        DB::table('shift_sessions')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'status' => $status,
            'opening_cash' => 0,
            'opened_at' => $openedAt,
            'closed_at' => $status === 'closed' ? $openedAt : null,
            'created_at' => $openedAt,
            'updated_at' => $openedAt,
        ]);

        return $id;
    }

    private function reportTransaction(
        string $businessId,
        string $outletId,
        string $deviceId,
        string $cashierId,
        string $shiftId,
        int $total,
        string $status = 'paid',
        int $discountTotal = 0,
        mixed $createdAt = null,
    ): string {
        $createdAt ??= now();
        $id = (string) Str::uuid();
        DB::table('transactions')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'number' => 'RPT-'.Str::upper(Str::random(8)),
            'status' => $status,
            'subtotal' => $total + $discountTotal,
            'discount_total' => $discountTotal,
            'item_discount_total' => $discountTotal,
            'cart_discount_total' => 0,
            'service_charge_total' => 0,
            'tax_total' => 0,
            'rounding_total' => 0,
            'grand_total' => $total,
            'created_at' => $createdAt,
            'updated_at' => $createdAt,
        ]);

        return $id;
    }

    private function reportPayment(
        string $businessId,
        string $transactionId,
        string $method,
        int $amount,
        bool $isCash,
        string $status = 'confirmed',
        mixed $createdAt = null,
    ): void {
        $createdAt ??= now();
        DB::table('payments')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'transaction_id' => $transactionId,
            'method' => $method,
            'reference' => null,
            'amount' => $amount,
            'status' => $status,
            'is_cash' => $isCash,
            'confirmed_by' => null,
            'confirmed_at' => $status === 'confirmed' ? $createdAt : null,
            'created_at' => $createdAt,
            'updated_at' => $createdAt,
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
