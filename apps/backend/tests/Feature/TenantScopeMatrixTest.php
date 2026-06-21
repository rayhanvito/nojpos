<?php

namespace Tests\Feature;

use App\Models\AuditLog;
use App\Models\AttendanceRecord;
use App\Models\CashMovement;
use App\Models\Customer;
use App\Models\Device;
use App\Models\IdempotencyKey;
use App\Models\InventoryPurchase;
use App\Models\InventoryPurchaseItem;
use App\Models\Outlet;
use App\Models\Payment;
use App\Models\PaymentMethodConfig;
use App\Models\Product;
use App\Models\ProductCategory;
use App\Models\ShiftSession;
use App\Models\StockMovement;
use App\Models\Transaction;
use App\Models\TransactionItem;
use App\Models\User;
use App\Models\Concerns\BelongsToBusiness;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class TenantScopeMatrixTest extends TestCase
{
    use RefreshDatabase;

    protected function setUp(): void
    {
        parent::setUp();

        config()->set('nojpos.checkout.require_quote_for_checkout', false);
    }

    /**
     * Every new tenant-owned endpoint must add a row to test_domain_endpoints_reject_cross_tenant_resources.
     */
    public function test_all_business_owned_models_are_scoped_by_default(): void
    {
        $businessA = $this->business('Tenant A');
        $businessB = $this->business('Tenant B');
        $userA = $this->user($businessA, 'owner-a@example.test', 'owner');
        $tokenA = $this->tokenFor($userA);

        $models = [
            AuditLog::class => 'audit_logs',
            AttendanceRecord::class => 'attendance_records',
            CashMovement::class => 'cash_movements',
            Customer::class => 'customers',
            Device::class => 'devices',
            IdempotencyKey::class => 'idempotency_keys',
            InventoryPurchase::class => 'inventory_purchases',
            InventoryPurchaseItem::class => 'inventory_purchase_items',
            Outlet::class => 'outlets',
            Payment::class => 'payments',
            PaymentMethodConfig::class => 'payment_method_configs',
            Product::class => 'products',
            ProductCategory::class => 'product_categories',
            ShiftSession::class => 'shift_sessions',
            StockMovement::class => 'stock_movements',
            Transaction::class => 'transactions',
            TransactionItem::class => 'transaction_items',
            User::class => 'users',
        ];

        foreach ($models as $model => $table) {
            $this->assertContains(
                BelongsToBusiness::class,
                class_uses_recursive($model),
                "{$model} must use BelongsToBusiness.",
            );

            DB::table($table)->insert($this->minimalTenantRow($table, $businessA));
            DB::table($table)->insert($this->minimalTenantRow($table, $businessB));
        }

        $this->withToken($tokenA)->getJson('/api/v1/me')->assertOk();
        $this->assertSame(1, Product::query()->count());
        $this->assertSame(1, Transaction::query()->count());
        $this->assertSame(2, User::query()->whereIn('role', ['owner', 'admin', 'cashier'])->count());
    }

    public function test_domain_endpoints_reject_cross_tenant_resources(): void
    {
        $ctxA = $this->tenantContext('Tenant A', 'owner-a@example.test', 'owner');
        $ctxB = $this->tenantContext('Tenant B', 'owner-b@example.test', 'owner');

        $this->paymentMethod($ctxA['business'], $ctxA['outlet'], 'cash', true);
        $this->paymentMethod($ctxB['business'], $ctxB['outlet'], 'cash', true);
        $categoryA = $this->category($ctxA['business'], 'Kategori A');
        $categoryB = $this->category($ctxB['business'], 'Kategori B');
        $productA = $this->product($ctxA['business'], $ctxA['outlet'], $categoryA, 'Produk A', 10000);
        $productB = $this->product($ctxB['business'], $ctxB['outlet'], $categoryB, 'Produk B', 10000);
        $customerB = $this->customer($ctxB['business'], 'Pelanggan B');
        $shiftA = $this->shift($ctxA['business'], $ctxA['outlet'], $ctxA['device'], $ctxA['user'], 'open');
        $shiftB = $this->shift($ctxB['business'], $ctxB['outlet'], $ctxB['device'], $ctxB['user'], 'open');
        $transactionB = $this->transaction($ctxB['business'], $ctxB['outlet'], $ctxB['device'], $ctxB['user'], $shiftB, $productB);
        $paymentB = $this->payment($ctxB['business'], $transactionB, 'cash', 10000);
        $this->attendance($ctxB['business'], $ctxB['outlet'], $ctxB['user']);
        $this->stockMovement($ctxB['business'], $ctxB['outlet'], $productB, 'adjustment', 3);

        $tokenA = $ctxA['token'];

        $this->assertTenantIsolated('GET', '/api/v1/inventory?outlet_id='.$ctxB['outlet'], $tokenA);
        $this->assertTenantIsolated('GET', '/api/v1/attendance?staff_id='.$ctxB['user'], $tokenA);
        $this->assertTenantIsolated('GET', '/api/v1/reports/sales-summary?shift_id='.$shiftB, $tokenA);
        $this->assertTenantIsolated('POST', '/api/v1/shifts/open', $tokenA, [
            'outlet_id' => $ctxB['outlet'],
            'device_id' => $ctxA['device'],
            'cashier_id' => $ctxA['user'],
            'opening_cash' => 100000,
        ], ['Idempotency-Key' => (string) Str::uuid()]);
        $this->assertTenantIsolated('POST', '/api/v1/transactions/quote', $tokenA, $this->checkoutPayload($ctxB['outlet'], $ctxB['device'], $ctxB['user'], $shiftB, $productB));
        $this->assertTenantIsolated('POST', '/api/v1/transactions', $tokenA, $this->checkoutPayload($ctxB['outlet'], $ctxB['device'], $ctxB['user'], $shiftB, $productB, true), ['Idempotency-Key' => (string) Str::uuid()]);
        $this->assertTenantIsolated('POST', '/api/v1/payments', $tokenA, [
            'transaction_id' => $transactionB,
            'method' => 'cash',
            'amount' => 10000,
        ], ['Idempotency-Key' => (string) Str::uuid()]);
        $this->assertTenantIsolated('POST', '/api/v1/payments', $tokenA, [
            'payment_id' => $paymentB,
            'confirm' => true,
        ], ['Idempotency-Key' => (string) Str::uuid()]);
        $this->assertTenantIsolated('POST', '/api/v1/voids', $tokenA, [
            'transaction_id' => $transactionB,
            'shift_id' => $shiftA,
            'reason' => 'Cross tenant attempt',
        ], ['Idempotency-Key' => (string) Str::uuid()]);
        $this->assertTenantIsolated('POST', '/api/v1/products', $tokenA, [
            'outlet_id' => $ctxB['outlet'],
            'product_category_id' => $categoryA,
            'name' => 'Spoof Product',
            'price' => 1000,
        ], ['Idempotency-Key' => (string) Str::uuid()]);
        $this->assertTenantIsolated('PUT', '/api/v1/products/'.$productB, $tokenA, [
            'outlet_id' => $ctxA['outlet'],
            'product_category_id' => $categoryA,
            'name' => 'Spoof Product',
            'price' => 1000,
        ], ['Idempotency-Key' => (string) Str::uuid()]);
        $this->assertTenantIsolated('DELETE', '/api/v1/products/'.$productB, $tokenA, [], ['Idempotency-Key' => (string) Str::uuid()]);
        $this->assertTenantIsolated('DELETE', '/api/v1/categories/'.$categoryB, $tokenA, [], ['Idempotency-Key' => (string) Str::uuid()]);
        $this->assertTenantIsolated('PUT', '/api/v1/customers/'.$customerB, $tokenA, [
            'name' => 'Spoof Customer',
        ], ['Idempotency-Key' => (string) Str::uuid()]);
        $this->assertTenantIsolated('DELETE', '/api/v1/customers/'.$customerB, $tokenA, [], ['Idempotency-Key' => (string) Str::uuid()]);
        $this->assertTenantIsolated('PUT', '/api/v1/staff/'.$ctxB['user'], $tokenA, [
            'name' => 'Spoof Staff',
        ], ['Idempotency-Key' => (string) Str::uuid()]);
        $this->assertTenantIsolated('DELETE', '/api/v1/staff/'.$ctxB['user'], $tokenA, [], ['Idempotency-Key' => (string) Str::uuid()]);
        $this->assertTenantIsolated('POST', '/api/v1/inventory/purchases', $tokenA, [
            'outlet_id' => $ctxA['outlet'],
            'items' => [
                ['product_id' => $productB, 'quantity' => 1, 'unit_cost' => 1000],
            ],
        ], ['Idempotency-Key' => (string) Str::uuid()]);

        $this->withToken($tokenA)->getJson('/api/v1/products?search=Produk B')
            ->assertOk()
            ->assertJsonCount(0, 'data.products');
        $this->withToken($tokenA)->getJson('/api/v1/categories?search=Kategori B')
            ->assertOk()
            ->assertJsonCount(0, 'data.categories');
        $this->withToken($tokenA)->getJson('/api/v1/customers?search=Pelanggan B')
            ->assertOk()
            ->assertJsonCount(0, 'data.customers');
        $this->withToken($tokenA)->getJson('/api/v1/staff?search=owner-b')
            ->assertOk()
            ->assertJsonCount(0, 'data.staff');
    }

    public function test_write_payload_cannot_spoof_business_id(): void
    {
        $ctxA = $this->tenantContext('Tenant A', 'owner-a@example.test', 'owner');
        $ctxB = $this->tenantContext('Tenant B', 'owner-b@example.test', 'owner');

        $created = $this->withToken($ctxA['token'])->postJson('/api/v1/customers', [
            'business_id' => $ctxB['business'],
            'name' => 'No Spoof',
            'phone' => '0812',
            'group' => 'VIP',
        ])->assertCreated();

        $created->assertJsonPath('data.business_id', $ctxA['business']);
        $this->assertDatabaseHas('customers', [
            'id' => $created->json('data.id'),
            'business_id' => $ctxA['business'],
        ]);
        $this->assertDatabaseMissing('customers', [
            'id' => $created->json('data.id'),
            'business_id' => $ctxB['business'],
        ]);
    }

    public function test_superadmin_is_not_cross_tenant_by_default(): void
    {
        $ctxB = $this->tenantContext('Tenant B', 'owner-b@example.test', 'owner');
        $superadmin = $this->user(null, 'superadmin@example.test', 'superadmin');

        $this->product($ctxB['business'], $ctxB['outlet'], $this->category($ctxB['business'], 'Kategori B'), 'Produk B', 10000);

        $this->withToken($this->tokenFor($superadmin))->getJson('/api/v1/products')
            ->assertOk()
            ->assertJsonCount(0, 'data.products');
    }

    private function assertTenantIsolated(string $method, string $uri, string $token, array $payload = [], array $headers = []): void
    {
        $response = $this->withToken($token)->withHeaders($headers)->json($method, $uri, $payload);

        $this->assertContains(
            $response->status(),
            [403, 404],
            "{$method} {$uri} must reject cross-tenant resource access.",
        );
    }

    private function minimalTenantRow(string $table, string $businessId): array
    {
        $id = (string) Str::uuid();
        $now = now();

        return match ($table) {
            'audit_logs' => ['id' => $id, 'business_id' => $businessId, 'actor_id' => null, 'action' => 'test', 'entity_type' => null, 'entity_id' => null, 'before' => null, 'after' => null, 'created_at' => $now],
            'attendance_records' => ['id' => $id, 'business_id' => $businessId, 'outlet_id' => (string) Str::uuid(), 'staff_id' => (string) Str::uuid(), 'created_by' => (string) Str::uuid(), 'clock_in_at' => $now, 'clock_out_at' => null, 'created_at' => $now, 'updated_at' => $now],
            'cash_movements' => ['id' => $id, 'business_id' => $businessId, 'outlet_id' => (string) Str::uuid(), 'shift_id' => (string) Str::uuid(), 'actor_id' => (string) Str::uuid(), 'type' => 'cash_in', 'amount' => 1, 'reason' => null, 'created_at' => $now, 'updated_at' => $now],
            'customers' => ['id' => $id, 'business_id' => $businessId, 'name' => 'Customer '.$id, 'phone' => null, 'group' => 'Tanpa Grup', 'created_at' => $now, 'updated_at' => $now],
            'devices' => ['id' => $id, 'business_id' => $businessId, 'outlet_id' => (string) Str::uuid(), 'device_uuid' => 'device-'.$id, 'name' => 'Device', 'created_at' => $now, 'updated_at' => $now],
            'idempotency_keys' => ['id' => $id, 'business_id' => $businessId, 'endpoint' => 'POST test/'.$id, 'key' => (string) Str::uuid(), 'request_hash' => hash('sha256', $id), 'response_snapshot' => '{}', 'status' => 200, 'created_at' => $now, 'updated_at' => $now],
            'inventory_purchases' => ['id' => $id, 'business_id' => $businessId, 'outlet_id' => (string) Str::uuid(), 'number' => 'PO-'.$id, 'supplier_name' => null, 'notes' => null, 'total' => 1, 'purchased_at' => $now, 'created_at' => $now, 'updated_at' => $now],
            'inventory_purchase_items' => ['id' => $id, 'business_id' => $businessId, 'purchase_id' => (string) Str::uuid(), 'product_id' => (string) Str::uuid(), 'quantity' => 1, 'unit_cost' => 1, 'subtotal' => 1, 'created_at' => $now, 'updated_at' => $now],
            'outlets' => ['id' => $id, 'business_id' => $businessId, 'name' => 'Outlet '.$id, 'service_charge_rate' => 0, 'tax_rate' => 0, 'created_at' => $now, 'updated_at' => $now],
            'payments' => ['id' => $id, 'business_id' => $businessId, 'transaction_id' => (string) Str::uuid(), 'method' => 'cash', 'reference' => null, 'amount' => 1, 'status' => 'confirmed', 'is_cash' => true, 'confirmed_by' => null, 'confirmed_at' => $now, 'created_at' => $now, 'updated_at' => $now],
            'payment_method_configs' => ['id' => $id, 'business_id' => $businessId, 'outlet_id' => null, 'method' => 'cash-'.$id, 'is_cash' => true, 'created_at' => $now, 'updated_at' => $now],
            'products' => ['id' => $id, 'business_id' => $businessId, 'outlet_id' => null, 'product_category_id' => null, 'name' => 'Product '.$id, 'barcode' => null, 'price' => 1, 'track_stock' => false, 'created_at' => $now, 'updated_at' => $now],
            'product_categories' => ['id' => $id, 'business_id' => $businessId, 'name' => 'Category '.$id, 'created_at' => $now, 'updated_at' => $now],
            'shift_sessions' => ['id' => $id, 'business_id' => $businessId, 'outlet_id' => (string) Str::uuid(), 'device_id' => (string) Str::uuid(), 'cashier_id' => (string) Str::uuid(), 'status' => 'open', 'opening_cash' => 0, 'expected_cash' => null, 'actual_cash' => null, 'cash_difference' => null, 'opened_at' => $now, 'closed_at' => null, 'created_at' => $now, 'updated_at' => $now],
            'stock_movements' => ['id' => $id, 'business_id' => $businessId, 'outlet_id' => (string) Str::uuid(), 'product_id' => (string) Str::uuid(), 'transaction_id' => null, 'purchase_id' => null, 'type' => 'adjustment', 'quantity_delta' => 1, 'created_at' => $now, 'updated_at' => $now],
            'transactions' => ['id' => $id, 'business_id' => $businessId, 'outlet_id' => (string) Str::uuid(), 'device_id' => (string) Str::uuid(), 'cashier_id' => (string) Str::uuid(), 'shift_id' => (string) Str::uuid(), 'number' => 'TRX-'.$id, 'status' => 'paid', 'subtotal' => 1, 'discount_total' => 0, 'item_discount_total' => 0, 'cart_discount_total' => 0, 'service_charge_total' => 0, 'tax_total' => 0, 'rounding_total' => 0, 'grand_total' => 1, 'created_at' => $now, 'updated_at' => $now],
            'transaction_items' => ['id' => $id, 'business_id' => $businessId, 'transaction_id' => (string) Str::uuid(), 'product_id' => (string) Str::uuid(), 'name' => 'Item', 'quantity' => 1, 'unit_price' => 1, 'discount' => 0, 'subtotal' => 1, 'created_at' => $now, 'updated_at' => $now],
            'users' => ['id' => $id, 'business_id' => $businessId, 'name' => 'User '.$id, 'email' => $id.'@example.test', 'password' => Hash::make('password'), 'role' => 'cashier', 'pin_hash' => Hash::make('1234'), 'created_at' => $now, 'updated_at' => $now],
        };
    }

    private function tenantContext(string $name, string $email, string $role): array
    {
        $business = $this->business($name);
        $outlet = $this->outlet($business);
        $device = $this->device($business, $outlet, Str::slug($name).'-tablet');
        $user = $this->user($business, $email, $role);

        return compact('business', 'outlet', 'device', 'user') + ['token' => $this->tokenFor($user)];
    }

    private function checkoutPayload(string $outlet, string $device, string $cashier, string $shift, string $product, bool $withPayment = false): array
    {
        $payload = [
            'outlet_id' => $outlet,
            'device_id' => $device,
            'cashier_id' => $cashier,
            'shift_id' => $shift,
            'items' => [
                ['product_id' => $product, 'quantity' => 1, 'unit_price' => 10000],
            ],
        ];

        if ($withPayment) {
            $payload['payments'] = [
                ['method' => 'cash', 'amount' => 10000],
            ];
        }

        return $payload;
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
        DB::table('outlets')->insert(['id' => $id, 'business_id' => $businessId, 'name' => 'Outlet', 'service_charge_rate' => 0, 'tax_rate' => 0, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function device(string $businessId, string $outletId, string $uuid): string
    {
        $id = (string) Str::uuid();
        DB::table('devices')->insert(['id' => $id, 'business_id' => $businessId, 'outlet_id' => $outletId, 'device_uuid' => $uuid, 'name' => $uuid, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function user(?string $businessId, string $email, string $role): string
    {
        $id = (string) Str::uuid();
        DB::table('users')->insert(['id' => $id, 'business_id' => $businessId, 'name' => $role.' user', 'email' => $email, 'password' => Hash::make('password'), 'role' => $role, 'pin_hash' => Hash::make('1234'), 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function tokenFor(string $userId): string
    {
        DB::table('personal_access_tokens')->insert([
            'id' => (string) Str::uuid(),
            'tokenable_type' => User::class,
            'tokenable_id' => $userId,
            'name' => 'test',
            'token' => hash('sha256', 'plain-'.$userId),
            'abilities' => json_encode(['*']),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return 'plain-'.$userId;
    }

    private function category(string $businessId, string $name): string
    {
        $id = (string) Str::uuid();
        DB::table('product_categories')->insert(['id' => $id, 'business_id' => $businessId, 'name' => $name, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function product(string $businessId, string $outletId, string $categoryId, string $name, int $price): string
    {
        $id = (string) Str::uuid();
        DB::table('products')->insert(['id' => $id, 'business_id' => $businessId, 'outlet_id' => $outletId, 'product_category_id' => $categoryId, 'name' => $name, 'barcode' => null, 'price' => $price, 'track_stock' => true, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function customer(string $businessId, string $name): string
    {
        $id = (string) Str::uuid();
        DB::table('customers')->insert(['id' => $id, 'business_id' => $businessId, 'name' => $name, 'phone' => null, 'group' => 'Tanpa Grup', 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function shift(string $businessId, string $outletId, string $deviceId, string $cashierId, string $status): string
    {
        $id = (string) Str::uuid();
        DB::table('shift_sessions')->insert(['id' => $id, 'business_id' => $businessId, 'outlet_id' => $outletId, 'device_id' => $deviceId, 'cashier_id' => $cashierId, 'status' => $status, 'opening_cash' => 0, 'opened_at' => now(), 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function transaction(string $businessId, string $outletId, string $deviceId, string $cashierId, string $shiftId, string $productId): string
    {
        $id = (string) Str::uuid();
        DB::table('transactions')->insert(['id' => $id, 'business_id' => $businessId, 'outlet_id' => $outletId, 'device_id' => $deviceId, 'cashier_id' => $cashierId, 'shift_id' => $shiftId, 'number' => 'TRX-'.$id, 'status' => 'paid', 'subtotal' => 10000, 'discount_total' => 0, 'item_discount_total' => 0, 'cart_discount_total' => 0, 'service_charge_total' => 0, 'tax_total' => 0, 'rounding_total' => 0, 'grand_total' => 10000, 'created_at' => now(), 'updated_at' => now()]);
        DB::table('transaction_items')->insert(['id' => (string) Str::uuid(), 'business_id' => $businessId, 'transaction_id' => $id, 'product_id' => $productId, 'name' => 'Item', 'quantity' => 1, 'unit_price' => 10000, 'discount' => 0, 'subtotal' => 10000, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function payment(string $businessId, string $transactionId, string $method, int $amount): string
    {
        $id = (string) Str::uuid();
        DB::table('payments')->insert(['id' => $id, 'business_id' => $businessId, 'transaction_id' => $transactionId, 'method' => $method, 'reference' => null, 'amount' => $amount, 'status' => 'pending', 'is_cash' => false, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function paymentMethod(string $businessId, string $outletId, string $method, bool $isCash): void
    {
        DB::table('payment_method_configs')->insert(['id' => (string) Str::uuid(), 'business_id' => $businessId, 'outlet_id' => $outletId, 'method' => $method, 'is_cash' => $isCash, 'created_at' => now(), 'updated_at' => now()]);
    }

    private function attendance(string $businessId, string $outletId, string $staffId): void
    {
        DB::table('attendance_records')->insert(['id' => (string) Str::uuid(), 'business_id' => $businessId, 'outlet_id' => $outletId, 'staff_id' => $staffId, 'created_by' => $staffId, 'clock_in_at' => now(), 'created_at' => now(), 'updated_at' => now()]);
    }

    private function stockMovement(string $businessId, string $outletId, string $productId, string $type, int $quantity): void
    {
        DB::table('stock_movements')->insert(['id' => (string) Str::uuid(), 'business_id' => $businessId, 'outlet_id' => $outletId, 'product_id' => $productId, 'transaction_id' => null, 'purchase_id' => null, 'type' => $type, 'quantity_delta' => $quantity, 'created_at' => now(), 'updated_at' => now()]);
    }
}
