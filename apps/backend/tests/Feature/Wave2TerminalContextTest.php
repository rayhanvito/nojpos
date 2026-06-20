<?php

namespace Tests\Feature;

use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class Wave2TerminalContextTest extends TestCase
{
    use RefreshDatabase;

    public function test_same_tenant_cashier_spoofing_is_rejected(): void
    {
        $ctx = $this->tenantContext();
        $otherCashier = $this->user($ctx['business'], 'cashier-b@example.test', 'cashier');

        $this->withToken($this->tokenFor($ctx['cashier']))
            ->postJson('/api/v1/shifts/open', [
                'outlet_id' => $ctx['outlet'],
                'device_id' => $ctx['device'],
                'cashier_id' => $otherCashier,
                'opening_cash' => 100000,
            ])->assertForbidden()
            ->assertJsonPath('error.code', 'TERMINAL_CONTEXT_MISMATCH');

        $this->assertSame(0, DB::table('shift_sessions')->where('business_id', $ctx['business'])->count());
    }

    public function test_same_tenant_device_spoofing_and_outlet_device_mismatch_are_rejected(): void
    {
        $ctx = $this->tenantContext();
        $otherOutlet = $this->outlet($ctx['business'], 'Outlet B');
        $otherDevice = $this->device($ctx['business'], $otherOutlet, 'tablet-b');

        $this->withToken($this->tokenFor($ctx['cashier']))
            ->postJson('/api/v1/shifts/open', [
                'outlet_id' => $ctx['outlet'],
                'device_id' => $otherDevice,
                'cashier_id' => $ctx['cashier'],
                'opening_cash' => 100000,
            ])->assertForbidden()
            ->assertJsonPath('error.code', 'TERMINAL_CONTEXT_MISMATCH');

        $this->assertSame(0, DB::table('shift_sessions')->where('business_id', $ctx['business'])->count());
    }

    public function test_unknown_or_revoked_device_is_rejected_as_not_enrolled(): void
    {
        $ctx = $this->tenantContext();

        $this->withToken($this->tokenFor($ctx['cashier']))
            ->postJson('/api/v1/shifts/open', [
                'outlet_id' => $ctx['outlet'],
                'device_id' => (string) Str::uuid(),
                'cashier_id' => $ctx['cashier'],
                'opening_cash' => 100000,
            ])->assertForbidden()
            ->assertJsonPath('error.code', 'DEVICE_NOT_ENROLLED');
    }

    public function test_shift_tuple_mismatch_is_rejected_for_checkout(): void
    {
        $ctx = $this->tenantContext();
        $otherCashier = $this->user($ctx['business'], 'cashier-b@example.test', 'cashier');
        $category = $this->category($ctx['business']);
        $product = $this->product($ctx['business'], $ctx['outlet'], $category, 10000);
        $shift = $this->shift($ctx['business'], $ctx['outlet'], $ctx['device'], $ctx['cashier']);

        $this->withToken($this->tokenFor($otherCashier))
            ->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/transactions', [
                'outlet_id' => $ctx['outlet'],
                'device_id' => $ctx['device'],
                'cashier_id' => $otherCashier,
                'shift_id' => $shift,
                'items' => [
                    ['product_id' => $product, 'quantity' => 1, 'unit_price' => 10000, 'discount' => 0],
                ],
                'payments' => [
                    ['method' => 'cash', 'amount' => 10000],
                ],
            ])->assertForbidden()
            ->assertJsonPath('error.code', 'TERMINAL_CONTEXT_MISMATCH');

        $this->assertSame(0, DB::table('transactions')->where('business_id', $ctx['business'])->count());
    }

    public function test_payment_from_wrong_cashier_terminal_context_is_rejected(): void
    {
        $ctx = $this->tenantContext();
        $otherCashier = $this->user($ctx['business'], 'cashier-b@example.test', 'cashier');
        $shift = $this->shift($ctx['business'], $ctx['outlet'], $ctx['device'], $ctx['cashier']);
        $transaction = $this->transaction($ctx['business'], $ctx['outlet'], $ctx['device'], $ctx['cashier'], $shift, 10000);

        $this->withToken($this->tokenFor($otherCashier))
            ->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/payments', [
                'transaction_id' => $transaction,
                'method' => 'cash',
                'amount' => 10000,
            ])->assertForbidden()
            ->assertJsonPath('error.code', 'TERMINAL_CONTEXT_MISMATCH');

        $this->assertSame(0, DB::table('payments')->where('business_id', $ctx['business'])->count());
    }

    public function test_void_rejects_transaction_that_does_not_match_shift_device_tuple(): void
    {
        $ctx = $this->tenantContext();
        $owner = $this->user($ctx['business'], 'owner@example.test', 'owner');
        $otherDevice = $this->device($ctx['business'], $ctx['outlet'], 'tablet-b');
        $shift = $this->shift($ctx['business'], $ctx['outlet'], $ctx['device'], $ctx['cashier']);
        $transaction = $this->transaction($ctx['business'], $ctx['outlet'], $otherDevice, $ctx['cashier'], $shift, 10000, 'paid');

        $this->withToken($this->tokenFor($owner))
            ->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/voids', [
                'transaction_id' => $transaction,
                'shift_id' => $shift,
                'reason' => 'Tuple mismatch test',
            ])->assertForbidden()
            ->assertJsonPath('error.code', 'TERMINAL_CONTEXT_MISMATCH');

        $this->assertDatabaseHas('transactions', [
            'id' => $transaction,
            'status' => 'paid',
        ]);
    }

    public function test_valid_terminal_context_still_succeeds_for_shift_and_checkout(): void
    {
        $ctx = $this->tenantContext();
        $category = $this->category($ctx['business']);
        $product = $this->product($ctx['business'], $ctx['outlet'], $category, 10000);

        $token = $this->tokenFor($ctx['cashier']);

        $open = $this->withToken($token)
            ->postJson('/api/v1/shifts/open', [
                'outlet_id' => $ctx['outlet'],
                'device_id' => $ctx['device'],
                'cashier_id' => $ctx['cashier'],
                'opening_cash' => 100000,
            ])->assertCreated();

        $this->withToken($token)
            ->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/transactions', [
                'outlet_id' => $ctx['outlet'],
                'device_id' => $ctx['device'],
                'cashier_id' => $ctx['cashier'],
                'shift_id' => $open->json('data.id'),
                'items' => [
                    ['product_id' => $product, 'quantity' => 1, 'unit_price' => 10000, 'discount' => 0],
                ],
                'payments' => [
                    ['method' => 'cash', 'amount' => 10000],
                ],
            ])->assertCreated()
            ->assertJsonPath('data.status', 'paid')
            ->assertJsonPath('data.cashier_id', $ctx['cashier'])
            ->assertJsonPath('data.device_id', $ctx['device'])
            ->assertJsonPath('data.outlet_id', $ctx['outlet']);
    }

    private function tenantContext(): array
    {
        $business = $this->business();
        $outlet = $this->outlet($business, 'Outlet A');
        $device = $this->device($business, $outlet, 'tablet-a');
        $cashier = $this->user($business, 'cashier-a@example.test', 'cashier');

        return compact('business', 'outlet', 'device', 'cashier');
    }

    private function business(): string
    {
        $id = (string) Str::uuid();
        DB::table('businesses')->insert([
            'id' => $id,
            'name' => 'Wave 2 Tenant',
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function outlet(string $business, string $name): string
    {
        $id = (string) Str::uuid();
        DB::table('outlets')->insert([
            'id' => $id,
            'business_id' => $business,
            'name' => $name,
            'service_charge_rate' => 0,
            'tax_rate' => 0,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function device(string $business, string $outlet, string $uuid): string
    {
        $id = (string) Str::uuid();
        DB::table('devices')->insert([
            'id' => $id,
            'business_id' => $business,
            'outlet_id' => $outlet,
            'device_uuid' => $uuid,
            'name' => $uuid,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function user(string $business, string $email, string $role): string
    {
        $id = (string) Str::uuid();
        DB::table('users')->insert([
            'id' => $id,
            'business_id' => $business,
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

    private function tokenFor(string $user): string
    {
        DB::table('personal_access_tokens')->insert([
            'id' => (string) Str::uuid(),
            'tokenable_type' => 'App\\Models\\User',
            'tokenable_id' => $user,
            'name' => 'test',
            'token' => hash('sha256', 'plain-'.$user),
            'abilities' => json_encode(['*']),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return 'plain-'.$user;
    }

    private function category(string $business): string
    {
        $id = (string) Str::uuid();
        DB::table('product_categories')->insert([
            'id' => $id,
            'business_id' => $business,
            'name' => 'Kategori',
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function product(string $business, string $outlet, string $category, int $price): string
    {
        $id = (string) Str::uuid();
        DB::table('products')->insert([
            'id' => $id,
            'business_id' => $business,
            'outlet_id' => $outlet,
            'product_category_id' => $category,
            'name' => 'Produk Test',
            'price' => $price,
            'track_stock' => true,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        DB::table('payment_method_configs')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $business,
            'outlet_id' => $outlet,
            'method' => 'cash',
            'is_cash' => true,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function shift(string $business, string $outlet, string $device, string $cashier): string
    {
        $id = (string) Str::uuid();
        DB::table('shift_sessions')->insert([
            'id' => $id,
            'business_id' => $business,
            'outlet_id' => $outlet,
            'device_id' => $device,
            'cashier_id' => $cashier,
            'status' => 'open',
            'opening_cash' => 0,
            'opened_at' => now(),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function transaction(string $business, string $outlet, string $device, string $cashier, string $shift, int $total, string $status = 'unpaid'): string
    {
        $id = (string) Str::uuid();
        DB::table('transactions')->insert([
            'id' => $id,
            'business_id' => $business,
            'outlet_id' => $outlet,
            'device_id' => $device,
            'cashier_id' => $cashier,
            'shift_id' => $shift,
            'number' => 'TST-'.Str::upper(Str::random(6)),
            'status' => $status,
            'subtotal' => $total,
            'discount_total' => 0,
            'item_discount_total' => 0,
            'cart_discount_total' => 0,
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
