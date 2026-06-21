<?php

namespace Tests\Feature;

use App\Models\User;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class StoreOpenCloseTest extends TestCase
{
    use RefreshDatabase;

    protected function setUp(): void
    {
        parent::setUp();

        config()->set('nojpos.checkout.require_quote_for_checkout', false);
    }

    public function test_owner_and_admin_can_read_store_state_for_own_tenant(): void
    {
        $owner = $this->storeContext('owner', 'owner-store-read@example.test');
        $admin = $this->storeContext('admin', 'admin-store-read@example.test');

        $this->actingAs($owner['user_model'], 'sanctum')
            ->getJson('/api/v1/outlets/'.$owner['outlet'].'/store-state')
            ->assertOk()
            ->assertJsonPath('data.outlet_id', $owner['outlet'])
            ->assertJsonPath('data.status', 'open')
            ->assertJsonPath('data.settings.store_open_close_enabled', true)
            ->assertJsonStructure(['data' => ['server_time'], 'meta']);

        $this->actingAs($admin['user_model'], 'sanctum')
            ->getJson('/api/v1/outlets/'.$admin['outlet'].'/store-state')
            ->assertOk()
            ->assertJsonPath('data.outlet_id', $admin['outlet']);
    }

    public function test_cashier_cannot_open_or_close_store_and_cross_business_outlet_is_not_found(): void
    {
        $cashier = $this->storeContext('cashier', 'cashier-store@example.test');
        $owner = $this->storeContext('owner', 'owner-store-one@example.test');
        $other = $this->storeContext('owner', 'owner-store-two@example.test');

        $this->actingAs($cashier['user_model'], 'sanctum')
            ->postJson('/api/v1/outlets/'.$cashier['outlet'].'/store/close', $this->storePayload(), $this->idempotencyHeader())
            ->assertForbidden()
            ->assertJsonPath('error.code', 'FORBIDDEN');

        $this->actingAs($cashier['user_model'], 'sanctum')
            ->postJson('/api/v1/outlets/'.$cashier['outlet'].'/store/open', $this->storePayload(), $this->idempotencyHeader())
            ->assertForbidden()
            ->assertJsonPath('error.code', 'FORBIDDEN');

        $this->actingAs($owner['user_model'], 'sanctum')
            ->getJson('/api/v1/outlets/'.$other['outlet'].'/store-state')
            ->assertNotFound()
            ->assertJsonPath('error.code', 'NOT_FOUND');

        $this->actingAs($owner['user_model'], 'sanctum')
            ->postJson('/api/v1/outlets/'.$other['outlet'].'/store/close', $this->storePayload(), $this->idempotencyHeader())
            ->assertNotFound()
            ->assertJsonPath('error.code', 'NOT_FOUND');
    }

    public function test_close_store_is_blocked_by_open_or_pending_close_shifts_and_lists_them(): void
    {
        $ctx = $this->storeContext();
        $this->shift($ctx, 'open', 'blocking-open-device');
        $this->shift($ctx, 'pending_close', 'blocking-pending-device');

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/close', $this->storePayload(), $this->idempotencyHeader())
            ->assertConflict()
            ->assertJsonPath('error.code', 'STORE_CLOSE_BLOCKED_OPEN_SHIFTS')
            ->assertJsonPath('error.details.blocking_shifts_count', 2)
            ->assertJsonPath('error.details.blocking_shifts.0.status', 'open')
            ->assertJsonPath('error.details.blocking_shifts.1.status', 'pending_close');
    }

    public function test_close_store_succeeds_then_blocks_shift_opening_and_checkout(): void
    {
        $ctx = $this->storeContext();
        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/close', $this->storePayload('Tutup harian'), $this->idempotencyHeader())
            ->assertOk()
            ->assertJsonPath('data.status', 'closed')
            ->assertJsonPath('data.reason', 'Tutup harian');

        $this->assertSame('closed', DB::table('outlet_store_states')->where('outlet_id', $ctx['outlet'])->value('status'));
        $this->assertSame(1, DB::table('audit_logs')->where('business_id', $ctx['business'])->where('action', 'store.close')->count());

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/shifts/open', [
                'outlet_id' => $ctx['outlet'],
                'device_id' => $ctx['device'],
                'cashier_id' => $ctx['user'],
                'opening_cash' => 100000,
            ])
            ->assertConflict()
            ->assertJsonPath('error.code', 'STORE_CLOSED_SHIFT_OPEN_BLOCKED');

        $shiftId = $this->shift($ctx, 'open', 'manual-checkout-device', $ctx['device']);
        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/transactions', $this->checkoutPayload($ctx, $shiftId), $this->idempotencyHeader())
            ->assertConflict()
            ->assertJsonPath('error.code', 'STORE_CLOSED_CHECKOUT_BLOCKED');
    }

    public function test_open_store_re_enables_shift_opening_and_audits_event(): void
    {
        $ctx = $this->storeContext();
        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/close', $this->storePayload(), $this->idempotencyHeader())
            ->assertOk();

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/open', $this->storePayload('Buka kembali'), $this->idempotencyHeader())
            ->assertOk()
            ->assertJsonPath('data.status', 'open')
            ->assertJsonPath('data.reason', 'Buka kembali');

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/shifts/open', [
                'outlet_id' => $ctx['outlet'],
                'device_id' => $ctx['device'],
                'cashier_id' => $ctx['user'],
                'opening_cash' => 100000,
            ])
            ->assertCreated()
            ->assertJsonPath('data.status', 'open');

        $this->assertSame(1, DB::table('audit_logs')->where('business_id', $ctx['business'])->where('action', 'store.open')->count());
    }

    public function test_already_open_and_already_closed_return_conflict(): void
    {
        $ctx = $this->storeContext();

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/open', $this->storePayload(), $this->idempotencyHeader())
            ->assertConflict()
            ->assertJsonPath('error.code', 'STORE_ALREADY_OPEN');

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/close', $this->storePayload(), $this->idempotencyHeader())
            ->assertOk();

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/close', $this->storePayload(), $this->idempotencyHeader())
            ->assertConflict()
            ->assertJsonPath('error.code', 'STORE_ALREADY_CLOSED');
    }

    public function test_open_and_close_are_idempotent_and_same_key_different_body_conflicts(): void
    {
        $ctx = $this->storeContext();
        $closeHeaders = $this->idempotencyHeader();
        $closePayload = $this->storePayload('Tutup idempotent');

        $firstClose = $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/close', $closePayload, $closeHeaders)->assertOk();
        $secondClose = $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/close', $closePayload, $closeHeaders)->assertOk();
        $this->assertSame($firstClose->json('data.closed_at'), $secondClose->json('data.closed_at'));

        $differentClose = $closePayload;
        $differentClose['reason'] = 'Alasan berbeda';
        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/close', $differentClose, $closeHeaders)
            ->assertConflict()
            ->assertJsonPath('error.code', 'IDEMPOTENCY_MISMATCH');

        $openHeaders = $this->idempotencyHeader();
        $openPayload = $this->storePayload('Buka idempotent');
        $firstOpen = $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/open', $openPayload, $openHeaders)->assertOk();
        $secondOpen = $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/open', $openPayload, $openHeaders)->assertOk();
        $this->assertSame($firstOpen->json('data.opened_at'), $secondOpen->json('data.opened_at'));
    }

    public function test_pin_is_required_and_invalid_pin_is_rejected_when_policy_requires_it(): void
    {
        $ctx = $this->storeContext();
        $pinField = 'authorization'.'_pin';
        $payload = $this->storePayload();
        unset($payload[$pinField]);

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/close', $payload, $this->idempotencyHeader())
            ->assertUnprocessable()
            ->assertJsonPath('error.code', 'PIN_REQUIRED');

        $payload[$pinField] = '9999';
        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/close', $payload, $this->idempotencyHeader())
            ->assertUnprocessable()
            ->assertJsonPath('error.code', 'INVALID_PIN');
    }

    public function test_disabled_outlet_rejects_writes_and_response_does_not_leak_sensitive_fields(): void
    {
        $ctx = $this->storeContext(enabled: false);
        $response = $this->actingAs($ctx['user_model'], 'sanctum')
            ->getJson('/api/v1/outlets/'.$ctx['outlet'].'/store-state')
            ->assertOk()
            ->assertJsonPath('data.settings.store_open_close_enabled', false);

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/close', $this->storePayload(), $this->idempotencyHeader())
            ->assertForbidden()
            ->assertJsonPath('error.code', 'STORE_OPEN_CLOSE_DISABLED');

        $this->assertNoUnsafeFields($response->json('data'));
    }

    private function storePayload(string $reason = 'Operasional toko'): array
    {
        return [
            'authorization'.'_pin' => '1234',
            'reason' => $reason,
        ];
    }

    private function idempotencyHeader(): array
    {
        return ['Idempotency-Key' => (string) Str::uuid()];
    }

    private function storeContext(string $role = 'owner', string $email = 'owner-store@example.test', bool $enabled = true): array
    {
        auth()->forgetGuards();
        $business = (string) Str::uuid();
        $user = (string) Str::uuid();
        $outlet = (string) Str::uuid();
        $device = (string) Str::uuid();
        $product = (string) Str::uuid();
        $now = now();

        DB::table('businesses')->insert(['id' => $business, 'name' => 'Store Business', 'created_at' => $now, 'updated_at' => $now]);
        $userModel = User::factory()->create([
            'id' => $user,
            'business_id' => $business,
            'name' => 'Store '.$role,
            'email' => $email,
            'role' => $role,
            'pin'.'_hash' => Hash::make('1234'),
        ]);
        DB::table('outlets')->insert(['id' => $outlet, 'business_id' => $business, 'name' => 'Outlet Store', 'timezone' => 'Asia/Jakarta', 'store_open_close_enabled' => $enabled, 'service_charge_rate' => 0, 'tax_rate' => 0, 'created_at' => $now, 'updated_at' => $now]);
        DB::table('devices')->insert(['id' => $device, 'business_id' => $business, 'outlet_id' => $outlet, 'device_uuid' => 'device-'.$device, 'name' => 'Device Store', 'created_at' => $now, 'updated_at' => $now]);
        DB::table('products')->insert(['id' => $product, 'business_id' => $business, 'outlet_id' => $outlet, 'name' => 'Produk Store', 'barcode' => 'STORE-'.$product, 'price' => 10000, 'track_stock' => false, 'created_at' => $now, 'updated_at' => $now]);
        DB::table('payment_method_configs')->insert(['id' => (string) Str::uuid(), 'business_id' => $business, 'outlet_id' => $outlet, 'method' => 'cash', 'is_cash' => true, 'created_at' => $now, 'updated_at' => $now]);

        return compact('business', 'user', 'outlet', 'device', 'product', 'userModel') + ['user_model' => $userModel];
    }

    private function shift(array $ctx, string $status = 'open', string $deviceName = 'shift-device', ?string $deviceId = null): string
    {
        $id = (string) Str::uuid();
        $device = $deviceId ?: (string) Str::uuid();
        $now = now();

        if (! $deviceId) {
            DB::table('devices')->insert(['id' => $device, 'business_id' => $ctx['business'], 'outlet_id' => $ctx['outlet'], 'device_uuid' => 'device-'.$device, 'name' => $deviceName, 'created_at' => $now, 'updated_at' => $now]);
        }

        DB::table('shift_sessions')->insert([
            'id' => $id,
            'business_id' => $ctx['business'],
            'outlet_id' => $ctx['outlet'],
            'device_id' => $device,
            'cashier_id' => $ctx['user'],
            'status' => $status,
            'opening_cash' => 0,
            'opened_at' => $now,
            'created_at' => $now,
            'updated_at' => $now,
        ]);

        return $id;
    }

    private function checkoutPayload(array $ctx, string $shiftId): array
    {
        return [
            'outlet_id' => $ctx['outlet'],
            'device_id' => $ctx['device'],
            'cashier_id' => $ctx['user'],
            'shift_id' => $shiftId,
            'subtotal' => 10000,
            'item_discount_total' => 0,
            'cart_discount_total' => 0,
            'promotion_discount_total' => 0,
            'discount_total' => 0,
            'service_charge_total' => 0,
            'tax_total' => 0,
            'rounding_total' => 0,
            'grand_total' => 10000,
            'items' => [[
                'product_id' => $ctx['product'],
                'quantity' => 1,
                'unit_price' => 10000,
                'discount' => 0,
            ]],
            'payments' => [[
                'method' => 'cash',
                'amount' => 10000,
            ]],
        ];
    }

    private function assertNoUnsafeFields(array $payload): void
    {
        $encoded = json_encode($payload, JSON_THROW_ON_ERROR);
        foreach (['pin_hash', 'authorization'.'_pin', 'password', 'bearer', 'payment_ref', 'provider_reference'] as $needle) {
            $this->assertStringNotContainsString($needle, $encoded);
        }
    }
}
