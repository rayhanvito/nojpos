<?php

namespace Tests\Feature;

use App\Models\User;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class SubscriptionApiTest extends TestCase
{
    use RefreshDatabase;

    public function test_owner_admin_and_cashier_can_read_own_tenant_subscription_envelope(): void
    {
        $businessId = $this->business('Tenant A');
        $planId = $this->plan('starter', [
            'max_outlets' => 1,
            'max_devices' => 2,
            'max_users' => 5,
            'max_products' => 100,
        ]);
        $subscriptionId = $this->subscription($businessId, $planId, [
            'status' => 'trial',
            'max_products_override' => 250,
            'trial_ends_at' => now()->addDays(7)->toDateString(),
            'current_period_ends_at' => now()->addMonth()->toDateString(),
        ]);
        $this->outlet($businessId);
        $this->device($businessId);
        $this->product($businessId);
        $roles = ['owner', 'admin', 'cashier'];
        $tokens = [];

        foreach ($roles as $role) {
            $userId = $this->user($businessId, "{$role}@example.test", $role);
            $tokens[$role] = $this->tokenFor($userId);
        }

        foreach ($roles as $role) {
            $this->withToken($tokens[$role])
                ->getJson('/api/v1/subscription')
                ->assertOk()
                ->assertJsonStructure(['data', 'meta'])
                ->assertJsonPath('data.status', 'trial')
                ->assertJsonPath('data.plan.name', 'Starter')
                ->assertJsonPath('data.plan.code', 'starter')
                ->assertJsonPath('data.limits.max_outlets', 1)
                ->assertJsonPath('data.limits.max_devices', 2)
                ->assertJsonPath('data.limits.max_users', 5)
                ->assertJsonPath('data.limits.max_products', 250)
                ->assertJsonPath('data.usage.outlets', 1)
                ->assertJsonPath('data.usage.devices', 1)
                ->assertJsonPath('data.usage.users', count($roles))
                ->assertJsonPath('data.usage.products', 1)
                ->assertJsonPath('data.enforcement.max_devices', 'informational')
                ->assertJsonPath('data.trial_ends_at', now()->addDays(7)->startOfDay()->toDateTimeString())
                ->assertJsonPath('data.current_period_ends_at', now()->addMonth()->startOfDay()->toDateTimeString());
        }

        $this->assertDatabaseHas('subscriptions', [
            'id' => $subscriptionId,
            'business_id' => $businessId,
        ]);
    }

    public function test_tenant_subscription_does_not_leak_other_tenant_data(): void
    {
        $tenantA = $this->business('Tenant A');
        $tenantB = $this->business('Tenant B');
        $planA = $this->plan('starter', ['max_products' => 100]);
        $planB = $this->plan('enterprise', ['max_products' => 999]);
        $this->subscription($tenantA, $planA, ['status' => 'active']);
        $this->subscription($tenantB, $planB, ['status' => 'suspended']);
        $userA = $this->user($tenantA, 'owner-a@example.test', 'owner');

        $this->withToken($this->tokenFor($userA))
            ->getJson('/api/v1/subscription')
            ->assertOk()
            ->assertJsonPath('data.status', 'active')
            ->assertJsonPath('data.plan.code', 'starter')
            ->assertJsonMissing(['code' => 'enterprise'])
            ->assertJsonMissing(['status' => 'suspended']);
    }

    public function test_superadmin_without_tenant_context_does_not_leak_subscription_data(): void
    {
        $tenant = $this->business('Tenant A');
        $plan = $this->plan('starter');
        $this->subscription($tenant, $plan, ['status' => 'active']);
        $superadmin = $this->user(null, 'superadmin@example.test', 'superadmin');

        $this->withToken($this->tokenFor($superadmin))
            ->getJson('/api/v1/subscription')
            ->assertOk()
            ->assertJsonStructure(['data', 'meta'])
            ->assertJsonPath('data.status', null)
            ->assertJsonPath('data.plan', null)
            ->assertJsonPath('data.limits.max_outlets', null)
            ->assertJsonPath('data.usage.outlets', 0)
            ->assertJsonPath('data.enforcement.max_devices', 'informational');
    }

    public function test_subscription_endpoint_is_read_only_and_does_not_accept_client_status(): void
    {
        $businessId = $this->business('Tenant A');
        $planId = $this->plan('starter');
        $this->subscription($businessId, $planId, ['status' => 'active']);
        $owner = $this->user($businessId, 'owner@example.test', 'owner');

        $this->withToken($this->tokenFor($owner))
            ->postJson('/api/v1/subscription', ['status' => 'suspended'])
            ->assertStatus(405);

        $this->assertDatabaseHas('subscriptions', [
            'business_id' => $businessId,
            'status' => 'active',
        ]);
        $this->assertDatabaseMissing('subscriptions', [
            'business_id' => $businessId,
            'status' => 'suspended',
        ]);
    }

    public function test_guest_is_unauthenticated_for_subscription_endpoint(): void
    {
        $this->getJson('/api/v1/subscription')
            ->assertUnauthorized()
            ->assertJsonPath('error.code', 'UNAUTHENTICATED');
    }

    private function business(string $name): string
    {
        $id = (string) Str::uuid();
        DB::table('businesses')->insert([
            'id' => $id,
            'name' => $name,
            'status' => 'active',
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function plan(string $code, array $overrides = []): string
    {
        $id = (string) Str::uuid();
        DB::table('plans')->insert([
            'id' => $id,
            'name' => Str::title($code),
            'code' => $code,
            'active_code' => $code,
            'price' => 100000,
            'billing_period' => 'monthly',
            'max_outlets' => $overrides['max_outlets'] ?? 2,
            'max_devices' => $overrides['max_devices'] ?? 3,
            'max_users' => $overrides['max_users'] ?? 4,
            'max_products' => $overrides['max_products'] ?? 5,
            'is_active' => true,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function subscription(string $businessId, string $planId, array $overrides = []): string
    {
        $id = (string) Str::uuid();
        DB::table('subscriptions')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'active_business_id' => $businessId,
            'plan_id' => $planId,
            'status' => $overrides['status'] ?? 'active',
            'trial_ends_at' => $overrides['trial_ends_at'] ?? null,
            'current_period_ends_at' => $overrides['current_period_ends_at'] ?? null,
            'max_outlets_override' => $overrides['max_outlets_override'] ?? null,
            'max_devices_override' => $overrides['max_devices_override'] ?? null,
            'max_users_override' => $overrides['max_users_override'] ?? null,
            'max_products_override' => $overrides['max_products_override'] ?? null,
            'notes' => null,
            'created_by' => null,
            'updated_by' => null,
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

    private function device(string $businessId): string
    {
        $outletId = DB::table('outlets')->where('business_id', $businessId)->value('id') ?? $this->outlet($businessId);
        $id = (string) Str::uuid();
        DB::table('devices')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'device_uuid' => 'device-'.$id,
            'name' => 'Device',
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function product(string $businessId): string
    {
        $id = (string) Str::uuid();
        DB::table('products')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => null,
            'product_category_id' => null,
            'name' => 'Product',
            'barcode' => null,
            'price' => 1000,
            'track_stock' => false,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function user(?string $businessId, string $email, string $role): string
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
}
