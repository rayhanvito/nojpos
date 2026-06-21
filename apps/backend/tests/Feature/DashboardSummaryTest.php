<?php

namespace Tests\Feature;

use Carbon\CarbonImmutable;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class DashboardSummaryTest extends TestCase
{
    use RefreshDatabase;

    public function test_owner_and_admin_can_access_dashboard_summary(): void
    {
        [$businessId, $outletId, $ownerId, $ownerToken] = $this->businessWithUser('owner', 'dashboard-owner@example.test');
        $adminId = $this->user($businessId, 'dashboard-admin@example.test', 'admin', 'Admin User');
        $adminToken = $this->tokenFor($adminId);
        $deviceId = $this->device($businessId, $outletId, 'dashboard-device');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId);
        $transactionId = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 25000);
        $this->payment($businessId, $transactionId, 'cash', 25000, 'confirmed', true);

        foreach ([$ownerToken, $adminToken] as $token) {
            $this->withToken($token)->getJson('/api/v1/dashboard/summary?date='.now('Asia/Jakarta')->toDateString())
                ->assertOk()
                ->assertJsonPath('data.meta.business_id', $businessId)
                ->assertJsonPath('data.kpis.sales_today.value', 25000)
                ->assertJsonPath('data.kpis.transaction_count.value', 1)
                ->assertJsonStructure($this->dashboardShape());
        }
    }

    public function test_cashier_and_superadmin_are_forbidden_and_guest_is_unauthenticated(): void
    {
        [$businessId, $outletId, $cashierId, $cashierToken] = $this->businessWithUser('cashier', 'dashboard-cashier@example.test');
        $superadminId = $this->user(null, 'dashboard-superadmin@example.test', 'superadmin', 'Super Admin');
        $superadminToken = $this->tokenFor($superadminId);

        $this->getJson('/api/v1/dashboard/summary')
            ->assertUnauthorized()
            ->assertJsonPath('error.code', 'UNAUTHENTICATED');

        $this->withToken($cashierToken)->getJson('/api/v1/dashboard/summary')
            ->assertForbidden()
            ->assertJsonPath('error.code', 'FORBIDDEN');

        $this->withToken($superadminToken)->getJson('/api/v1/dashboard/summary')
            ->assertForbidden()
            ->assertJsonPath('error.code', 'FORBIDDEN');
    }

    public function test_invalid_filters_return_validation_error(): void
    {
        [, , , $token] = $this->businessWithUser('owner', 'dashboard-invalid@example.test');

        $this->withToken($token)->getJson('/api/v1/dashboard/summary?date=21-06-2026')
            ->assertUnprocessable()
            ->assertJsonPath('error.code', 'VALIDATION_ERROR');

        $this->withToken($token)->getJson('/api/v1/dashboard/summary?range=month')
            ->assertUnprocessable()
            ->assertJsonPath('error.code', 'VALIDATION_ERROR');

        $this->withToken($token)->getJson('/api/v1/dashboard/summary?outlet_id=not-a-uuid')
            ->assertUnprocessable()
            ->assertJsonPath('error.code', 'VALIDATION_ERROR');
    }

    public function test_outlet_id_from_other_business_is_forbidden(): void
    {
        [, , , $tokenA] = $this->businessWithUser('owner', 'dashboard-a@example.test');
        [, $outletB] = $this->businessWithUser('owner', 'dashboard-b@example.test');

        $this->withToken($tokenA)->getJson('/api/v1/dashboard/summary?outlet_id='.$outletB)
            ->assertForbidden()
            ->assertJsonPath('error.code', 'FORBIDDEN');
    }

    public function test_empty_state_returns_stable_shape_and_zero_kpis(): void
    {
        [$businessId, $outletId, , $token] = $this->businessWithUser('owner', 'dashboard-empty@example.test');

        $this->withToken($token)->getJson('/api/v1/dashboard/summary?outlet_id='.$outletId.'&date=2026-06-21')
            ->assertOk()
            ->assertJsonStructure($this->dashboardShape())
            ->assertJsonPath('data.meta.business_id', $businessId)
            ->assertJsonPath('data.meta.outlet_id', $outletId)
            ->assertJsonPath('data.meta.is_empty_today', true)
            ->assertJsonPath('data.kpis.sales_today.value', 0)
            ->assertJsonPath('data.kpis.transaction_count.value', 0)
            ->assertJsonPath('data.kpis.average_transaction.value', 0)
            ->assertJsonCount(7, 'data.sales_last_7_days')
            ->assertJsonPath('data.payment_methods', [])
            ->assertJsonPath('data.top_products', [])
            ->assertJsonPath('data.recent_transactions', [])
            ->assertJsonPath('data.cashier_performance', []);
    }

    public function test_normal_state_returns_contract_sections_from_real_backend_data_only(): void
    {
        [$businessId, $outletId, $ownerId, $token] = $this->businessWithUser('owner', 'dashboard-normal@example.test');
        DB::table('outlets')->where('id', $outletId)->update(['name' => 'Cabang Utama', 'timezone' => 'Asia/Jakarta']);
        $deviceId = $this->device($businessId, $outletId, 'normal-device');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId, 'closed', 100000, 125000, 123000, -2000, CarbonImmutable::parse('2026-06-21 01:00:00', 'UTC'));
        $productA = $this->product($businessId, $outletId, 'Es Kopi Susu', 18000);
        $productB = $this->product($businessId, $outletId, 'Cup 16 oz', 1000);
        $this->stock($businessId, $outletId, $productB, 3);
        $transactionA = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 36000, createdAt: CarbonImmutable::parse('2026-06-21 03:00:00', 'UTC'));
        $this->item($businessId, $transactionA, $productA, 'Es Kopi Susu', 2, 36000, 0);
        $this->payment($businessId, $transactionA, 'qris', 36000, 'settled', false);
        $transactionB = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 18000, createdAt: CarbonImmutable::parse('2026-06-21 04:00:00', 'UTC'));
        $this->item($businessId, $transactionB, $productA, 'Es Kopi Susu', 1, 18000, 0);
        $this->payment($businessId, $transactionB, 'cash', 18000, 'confirmed', true);

        $this->withToken($token)->getJson('/api/v1/dashboard/summary?outlet_id='.$outletId.'&date=2026-06-21')
            ->assertOk()
            ->assertJsonStructure($this->dashboardShape())
            ->assertJsonPath('data.meta.data_status', 'partial')
            ->assertJsonPath('data.meta.is_empty_today', false)
            ->assertJsonPath('data.kpis.sales_today.value', 54000)
            ->assertJsonPath('data.kpis.transaction_count.value', 2)
            ->assertJsonPath('data.kpis.average_transaction.value', 27000)
            ->assertJsonPath('data.kpis.gross_profit_estimate.value', 0)
            ->assertJsonPath('data.kpis.gross_profit_estimate.is_estimate', true)
            ->assertJsonPath('data.kpis.low_stock_count.value', 1)
            ->assertJsonPath('data.kpis.cash_difference.value', -2000)
            ->assertJsonPath('data.payment_methods.0.method', 'qris')
            ->assertJsonPath('data.payment_methods.0.amount', 36000)
            ->assertJsonPath('data.top_products.0.product_id', $productA)
            ->assertJsonPath('data.top_products.0.qty_sold', 3)
            ->assertJsonPath('data.low_stock_items.0.product_id', $productB)
            ->assertJsonPath('data.recent_transactions.0.transaction_id', $transactionB)
            ->assertJsonPath('data.cashier_performance.0.cashier_id', $ownerId)
            ->assertJsonPath('data.cashier_performance.0.sales', 54000);
    }

    public function test_tenant_isolation_excludes_other_business_data_from_summary(): void
    {
        [$businessA, $outletA, $ownerA, $tokenA] = $this->businessWithUser('owner', 'dashboard-tenant-a@example.test');
        [$businessB, $outletB, $ownerB] = $this->businessWithUser('owner', 'dashboard-tenant-b@example.test');
        $deviceA = $this->device($businessA, $outletA, 'tenant-a-device');
        $deviceB = $this->device($businessB, $outletB, 'tenant-b-device');
        $shiftA = $this->shift($businessA, $outletA, $deviceA, $ownerA);
        $shiftB = $this->shift($businessB, $outletB, $deviceB, $ownerB);
        $productA = $this->product($businessA, $outletA, 'Produk A', 10000);
        $productB = $this->product($businessB, $outletB, 'Produk Rahasia Tenant B', 999000);

        $transactionA = $this->transaction($businessA, $outletA, $deviceA, $ownerA, $shiftA, 'paid', 10000);
        $this->item($businessA, $transactionA, $productA, 'Produk A', 1, 10000, 0);
        $this->payment($businessA, $transactionA, 'cash', 10000, 'confirmed', true);
        $transactionB = $this->transaction($businessB, $outletB, $deviceB, $ownerB, $shiftB, 'paid', 999000);
        $this->item($businessB, $transactionB, $productB, 'Produk Rahasia Tenant B', 1, 999000, 0);
        $this->payment($businessB, $transactionB, 'cash', 999000, 'confirmed', true);

        $response = $this->withToken($tokenA)->getJson('/api/v1/dashboard/summary?date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonPath('data.kpis.sales_today.value', 10000)
            ->assertJsonPath('data.kpis.transaction_count.value', 1);

        $json = json_encode($response->json(), JSON_THROW_ON_ERROR);
        $this->assertStringNotContainsString('Produk Rahasia Tenant B', $json);
        $this->assertStringNotContainsString($transactionB, $json);
        $this->assertStringNotContainsString($businessB, $json);
    }

    /**
     * @return array<string, mixed>
     */
    private function dashboardShape(): array
    {
        return [
            'data' => [
                'meta',
                'kpis' => [
                    'sales_today',
                    'transaction_count',
                    'average_transaction',
                    'gross_profit_estimate',
                    'low_stock_count',
                    'cash_difference',
                ],
                'alerts',
                'sales_last_7_days',
                'payment_methods',
                'top_products',
                'low_stock_items',
                'recent_transactions',
                'cashier_performance',
                'branch_highlights',
                'data_notes',
            ],
            'meta' => ['generated_at'],
        ];
    }

    /**
     * @return array{0:string, 1:string, 2:string, 3:string}
     */
    private function businessWithUser(string $role, string $email): array
    {
        $businessId = $this->business('Business '.$email);
        $outletId = $this->outlet($businessId);
        $userId = $this->user($businessId, $email, $role, $role.' user');

        return [$businessId, $outletId, $userId, $this->tokenFor($userId)];
    }

    private function business(string $name): string
    {
        $id = (string) Str::uuid();
        DB::table('businesses')->insert(['id' => $id, 'name' => $name, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function outlet(string $businessId): string
    {
        $id = (string) Str::uuid();
        DB::table('outlets')->insert(['id' => $id, 'business_id' => $businessId, 'name' => 'Outlet', 'service_charge_rate' => 0, 'tax_rate' => 0, 'timezone' => 'Asia/Jakarta', 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function user(?string $businessId, string $email, string $role, string $name): string
    {
        $id = (string) Str::uuid();
        DB::table('users')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'name' => $name,
            'email' => $email,
            'password' => Hash::make('password'),
            'role' => $role,
            'pin_hash' => Hash::make('1234'),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
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

    private function device(string $businessId, string $outletId, string $uuid): string
    {
        $id = (string) Str::uuid();
        DB::table('devices')->insert(['id' => $id, 'business_id' => $businessId, 'outlet_id' => $outletId, 'device_uuid' => $uuid, 'name' => $uuid, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function shift(
        string $businessId,
        string $outletId,
        string $deviceId,
        string $cashierId,
        string $status = 'open',
        int $openingCash = 0,
        ?int $expectedCash = null,
        ?int $actualCash = null,
        ?int $cashDifference = null,
        CarbonImmutable|string|null $openedAt = null,
    ): string {
        $id = (string) Str::uuid();
        DB::table('shift_sessions')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'status' => $status,
            'opening_cash' => $openingCash,
            'expected_cash' => $expectedCash,
            'actual_cash' => $actualCash,
            'cash_difference' => $cashDifference,
            'opened_at' => $openedAt ?? now(),
            'closed_at' => $status === 'closed' ? now() : null,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function product(string $businessId, string $outletId, string $name, int $price): string
    {
        $categoryId = (string) Str::uuid();
        DB::table('product_categories')->insert(['id' => $categoryId, 'business_id' => $businessId, 'name' => 'Kategori '.$name, 'created_at' => now(), 'updated_at' => now()]);
        $id = (string) Str::uuid();
        DB::table('products')->insert(['id' => $id, 'business_id' => $businessId, 'outlet_id' => $outletId, 'product_category_id' => $categoryId, 'name' => $name, 'price' => $price, 'track_stock' => true, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function stock(string $businessId, string $outletId, string $productId, int $quantityDelta): void
    {
        DB::table('stock_movements')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'product_id' => $productId,
            'transaction_id' => null,
            'type' => 'manual_test_seed',
            'quantity_delta' => $quantityDelta,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
    }

    private function transaction(
        string $businessId,
        string $outletId,
        string $deviceId,
        string $cashierId,
        string $shiftId,
        string $status,
        int $grandTotal,
        int $discountTotal = 0,
        CarbonImmutable|string|null $createdAt = null,
    ): string {
        $id = (string) Str::uuid();
        $timestamp = $createdAt ?? now();
        DB::table('transactions')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'number' => 'TRX-'.substr($id, 0, 8),
            'status' => $status,
            'subtotal' => $grandTotal + $discountTotal,
            'discount_total' => $discountTotal,
            'item_discount_total' => $discountTotal,
            'cart_discount_total' => 0,
            'service_charge_total' => 0,
            'tax_total' => 0,
            'rounding_total' => 0,
            'grand_total' => $grandTotal,
            'created_at' => $timestamp,
            'updated_at' => $timestamp,
        ]);

        return $id;
    }

    private function item(string $businessId, string $transactionId, string $productId, string $name, int $quantity, int $subtotal, int $discount): void
    {
        DB::table('transaction_items')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'transaction_id' => $transactionId,
            'product_id' => $productId,
            'name' => $name,
            'quantity' => $quantity,
            'unit_price' => intdiv($subtotal, max(1, $quantity)),
            'discount' => $discount,
            'subtotal' => $subtotal,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
    }

    private function payment(string $businessId, string $transactionId, string $method, int $amount, string $status, bool $isCash): string
    {
        $id = (string) Str::uuid();
        DB::table('payments')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'transaction_id' => $transactionId,
            'method' => $method,
            'reference' => null,
            'amount' => $amount,
            'status' => $status,
            'is_cash' => $isCash,
            'confirmed_at' => in_array($status, ['confirmed', 'settled'], true) ? now() : null,
            'provider' => $isCash ? null : $method,
            'provider_reference' => null,
            'confirm_expires_at' => $status === 'pending' ? now()->addMinutes(10) : null,
            'failed_at' => in_array($status, ['failed', 'expired', 'declined'], true) ? now() : null,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }
}
