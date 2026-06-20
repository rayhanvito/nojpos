<?php

namespace Tests\Feature;

use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class SettingsAggregateTest extends TestCase
{
    use RefreshDatabase;

    public function test_owner_can_get_settings_aggregate_without_sensitive_pin_or_secret_fields(): void
    {
        [$businessId, $outletId, $ownerId, $token] = $this->businessWithUser('owner');
        $configId = $this->paymentMethod($businessId, $outletId, 'cash', true);

        $response = $this->withToken($token)->getJson('/api/v1/settings')->assertOk();

        $response->assertJsonPath('data.business.id', $businessId)
            ->assertJsonPath('data.business.currency', 'IDR')
            ->assertJsonPath('data.outlets.0.id', $outletId)
            ->assertJsonPath('data.outlets.0.receipt_config.paper_width', '58mm')
            ->assertJsonPath('data.payment_methods.0.id', $configId)
            ->assertJsonPath('data.payment_methods.0.active', true)
            ->assertJsonPath('data.payment_methods.0.is_cash', true)
            ->assertJsonPath('data.security.pin_policy.max_attempts', 5)
            ->assertJsonPath('data.security.terminal_policy.idle_lock_timeout_seconds', 180)
            ->assertJsonPath('data.permissions.current_user.id', $ownerId)
            ->assertJsonPath('data.permissions.current_user.can_update_security_settings', true)
            ->assertJsonStructure(['data' => ['business', 'outlets', 'payment_methods', 'security', 'permissions'], 'meta' => ['config_version', 'server_timestamp', 'scope']]);

        $encoded = json_encode($response->json(), JSON_THROW_ON_ERROR);
        $this->assertStringNotContainsString('pin_hash', $encoded);
        $this->assertStringNotContainsString('password', $encoded);
        $this->assertStringNotContainsString('token', $encoded);
    }

    public function test_cashier_can_get_pos_safe_settings_but_inactive_payment_methods_are_hidden(): void
    {
        [$businessId, $outletId, $cashierId, $token] = $this->businessWithUser('cashier');
        $this->paymentMethod($businessId, $outletId, 'cash', true);
        $inactiveId = $this->paymentMethod($businessId, $outletId, 'qris', false, now());

        $response = $this->withToken($token)->getJson('/api/v1/settings')->assertOk();

        $response->assertJsonPath('meta.scope', 'pos_safe')
            ->assertJsonPath('data.permissions.current_user.id', $cashierId)
            ->assertJsonPath('data.permissions.current_user.can_update_settings', false)
            ->assertJsonCount(1, 'data.payment_methods');

        $encoded = json_encode($response->json('data.payment_methods'), JSON_THROW_ON_ERROR);
        $this->assertStringNotContainsString($inactiveId, $encoded);
    }

    public function test_cashier_cannot_update_settings(): void
    {
        [$businessId, $outletId, $cashierId, $token] = $this->businessWithUser('cashier');

        $this->withToken($token)->withHeader('Idempotency-Key', (string) Str::uuid())->patchJson('/api/v1/settings/security', [
            'pin_policy' => ['max_attempts' => 4],
        ])->assertStatus(403)
            ->assertJsonPath('error.code', 'FORBIDDEN');

        $this->assertDatabaseMissing('business_security_settings', [
            'business_id' => $businessId,
            'updated_by' => $cashierId,
        ]);
        $this->assertDatabaseHas('outlets', [
            'id' => $outletId,
            'business_id' => $businessId,
        ]);
    }

    public function test_admin_can_update_security_settings_and_audit_is_recorded_without_pin_hash(): void
    {
        [$businessId, $outletId, $adminId, $token] = $this->businessWithUser('admin');

        $response = $this->withToken($token)->withHeader('Idempotency-Key', (string) Str::uuid())->patchJson('/api/v1/settings/security', [
            'sensitive_actions' => [
                'void' => ['requires_pin' => false],
                'settings_change' => ['requires_pin' => true],
            ],
            'pin_policy' => [
                'max_attempts' => 3,
                'lockout_minutes' => 10,
            ],
            'terminal_policy' => [
                'idle_lock_timeout_seconds' => 240,
                'session_timeout_seconds' => 1200,
            ],
        ])->assertOk();

        $response->assertJsonPath('data.security.sensitive_actions.void.requires_pin', false)
            ->assertJsonPath('data.security.sensitive_actions.settings_change.requires_pin', true)
            ->assertJsonPath('data.security.pin_policy.max_attempts', 3)
            ->assertJsonPath('data.security.pin_policy.lockout_minutes', 10)
            ->assertJsonPath('data.security.terminal_policy.idle_lock_timeout_seconds', 240)
            ->assertJsonPath('data.security.terminal_policy.session_timeout_seconds', 1200);

        $this->assertDatabaseHas('business_security_settings', [
            'business_id' => $businessId,
            'updated_by' => $adminId,
            'pin_required_void' => false,
            'pin_required_settings_change' => true,
            'pin_lockout_max_attempts' => 3,
            'pin_lockout_decay_minutes' => 10,
            'idle_lock_timeout_seconds' => 240,
            'terminal_session_timeout_seconds' => 1200,
        ]);

        $this->assertDatabaseHas('audit_logs', [
            'business_id' => $businessId,
            'actor_id' => $adminId,
            'action' => 'settings.update',
            'entity_type' => 'business_security_setting',
        ]);

        $audit = DB::table('audit_logs')->where('business_id', $businessId)->where('action', 'settings.update')->first();
        $encodedAudit = json_encode($audit, JSON_THROW_ON_ERROR);
        $this->assertStringNotContainsString('pin_hash', $encodedAudit);
        $this->assertStringNotContainsString('1234', $encodedAudit);
        $this->assertDatabaseHas('outlets', ['id' => $outletId]);
    }

    public function test_owner_can_update_outlet_and_payment_method_settings_with_business_scope(): void
    {
        [$businessId, $outletId, $ownerId, $token] = $this->businessWithUser('owner');
        $configId = $this->paymentMethod($businessId, $outletId, 'cash', true);

        $this->withToken($token)->withHeader('Idempotency-Key', (string) Str::uuid())->patchJson('/api/v1/settings/outlets/'.$outletId, [
            'name' => 'Outlet Baru',
            'timezone' => 'Asia/Makassar',
            'service_charge_rate' => 5,
            'tax_rate' => 11,
            'receipt_config' => [
                'paper_width' => '80mm',
                'header_name' => 'Kedai Baru',
                'show_logo' => true,
            ],
        ])->assertOk()
            ->assertJsonPath('data.outlets.0.name', 'Outlet Baru')
            ->assertJsonPath('data.outlets.0.transaction_config.service_charge_rate', 5)
            ->assertJsonPath('data.outlets.0.receipt_config.paper_width', '80mm');

        $this->withToken($token)->withHeader('Idempotency-Key', (string) Str::uuid())->patchJson('/api/v1/settings/payment-methods/'.$configId, [
            'method' => 'edc_bca',
            'is_cash' => false,
            'active' => false,
        ])->assertOk()
            ->assertJsonPath('data.payment_methods.0.method', 'edc_bca')
            ->assertJsonPath('data.payment_methods.0.active', false);

        $this->assertDatabaseHas('outlets', [
            'id' => $outletId,
            'business_id' => $businessId,
            'name' => 'Outlet Baru',
            'service_charge_rate' => 5,
            'tax_rate' => 11,
            'receipt_paper_width' => '80mm',
            'receipt_header_name' => 'Kedai Baru',
            'receipt_show_logo' => true,
        ]);
        $this->assertDatabaseHas('payment_method_configs', [
            'id' => $configId,
            'business_id' => $businessId,
            'method' => 'edc_bca',
            'is_cash' => false,
        ]);
        $this->assertSame(2, DB::table('audit_logs')->where('business_id', $businessId)->where('actor_id', $ownerId)->where('action', 'settings.update')->count());
    }

    public function test_cross_business_outlet_and_payment_config_updates_are_forbidden(): void
    {
        [$businessA, , , $tokenA] = $this->businessWithUser('owner', 'a-owner@example.test');
        [$businessB, $outletB] = $this->businessWithUser('owner', 'b-owner@example.test');
        $configB = $this->paymentMethod($businessB, $outletB, 'cash', true);

        $this->withToken($tokenA)->withHeader('Idempotency-Key', (string) Str::uuid())->patchJson('/api/v1/settings/outlets/'.$outletB, [
            'name' => 'Should Fail',
        ])->assertStatus(403)
            ->assertJsonPath('error.code', 'FORBIDDEN');

        $this->withToken($tokenA)->withHeader('Idempotency-Key', (string) Str::uuid())->patchJson('/api/v1/settings/payment-methods/'.$configB, [
            'active' => false,
        ])->assertStatus(403)
            ->assertJsonPath('error.code', 'FORBIDDEN');

        $this->assertDatabaseHas('businesses', ['id' => $businessA]);
        $this->assertDatabaseHas('payment_method_configs', ['id' => $configB, 'business_id' => $businessB, 'deleted_at' => null]);
    }

    public function test_settings_updates_are_explicitly_validated(): void
    {
        [, $outletId, , $token] = $this->businessWithUser('owner');

        $outletValidation = $this->withToken($token)->withHeader('Idempotency-Key', (string) Str::uuid())->patchJson('/api/v1/settings/outlets/'.$outletId, [
            'tax_rate' => 101,
            'receipt_config' => ['paper_width' => 'A4'],
        ])->assertStatus(422)
            ->assertJsonPath('error.code', 'VALIDATION_ERROR')
            ->assertJsonPath('error.details.tax_rate.0', 'The tax rate field must not be greater than 100.');
        $this->assertArrayHasKey('receipt_config.paper_width', $outletValidation->json('error.details'));

        $securityValidation = $this->withToken($token)->withHeader('Idempotency-Key', (string) Str::uuid())->patchJson('/api/v1/settings/security', [
            'pin_policy' => ['max_attempts' => 0],
            'terminal_policy' => ['idle_lock_timeout_seconds' => 10],
        ])->assertStatus(422)
            ->assertJsonPath('error.code', 'VALIDATION_ERROR');
        $this->assertArrayHasKey('pin_policy.max_attempts', $securityValidation->json('error.details'));
        $this->assertArrayHasKey('terminal_policy.idle_lock_timeout_seconds', $securityValidation->json('error.details'));
    }

    public function test_existing_me_outlets_and_quote_endpoints_still_work_after_settings_contract(): void
    {
        [$businessId, $outletId, $cashierId, $token] = $this->businessWithUser('cashier');
        $categoryId = $this->category($businessId, 'Minuman');
        $productId = $this->product($businessId, $outletId, $categoryId, 'Teh', 10000);
        $deviceId = $this->device($businessId, $outletId, 'tablet-quote');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $cashierId);

        $this->withToken($token)->getJson('/api/v1/me')->assertOk()
            ->assertJsonPath('data.user.id', $cashierId)
            ->assertJsonPath('data.outlets.0.id', $outletId);

        $this->withToken($token)->getJson('/api/v1/outlets')->assertOk()
            ->assertJsonPath('data.outlets.0.receipt_config.paper_width', '58mm');

        $this->withToken($token)->postJson('/api/v1/transactions/quote', [
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'items' => [
                ['product_id' => $productId, 'quantity' => 2, 'unit_price' => 1],
            ],
        ])->assertOk()
            ->assertJsonPath('data.subtotal', 20000)
            ->assertJsonPath('data.rounding_total', 0);
    }

    /**
     * @return array{0:string, 1:string, 2:string, 3:string}
     */
    private function businessWithUser(string $role, string $email = 'user@example.test'): array
    {
        $businessId = $this->business('Business '.$email);
        $outletId = $this->outlet($businessId);
        $userId = $this->user($businessId, $email, $role);

        return [$businessId, $outletId, $userId, $this->tokenFor($userId)];
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

    private function business(string $name): string
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

    private function outlet(string $businessId): string
    {
        $id = (string) Str::uuid();
        DB::table('outlets')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'name' => 'Outlet',
            'service_charge_rate' => 0,
            'tax_rate' => 0,
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

    private function paymentMethod(string $businessId, string $outletId, string $method, bool $isCash, mixed $deletedAt = null): string
    {
        $id = (string) Str::uuid();
        DB::table('payment_method_configs')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'method' => $method,
            'is_cash' => $isCash,
            'deleted_at' => $deletedAt,
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

    private function product(string $businessId, string $outletId, string $categoryId, string $name, int $price): string
    {
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
