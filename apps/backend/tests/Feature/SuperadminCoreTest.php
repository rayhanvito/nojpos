<?php

namespace Tests\Feature;

use App\Models\User;
use Database\Seeders\SuperadminBootstrapSeeder;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class SuperadminCoreTest extends TestCase
{
    use RefreshDatabase;

    public function test_superadmin_namespace_requires_authentication_and_superadmin_role(): void
    {
        $businessId = $this->business('Tenant A');

        $this->getJson('/api/v1/superadmin/summary')
            ->assertUnauthorized()
            ->assertJsonPath('error.code', 'UNAUTHENTICATED');

        foreach (['cashier', 'owner', 'admin'] as $role) {
            $token = $this->tokenFor($this->user($businessId, "{$role}@example.test", $role));

            $this->withToken($token)->getJson('/api/v1/superadmin/summary')
                ->assertForbidden()
                ->assertJsonPath('error.code', 'FORBIDDEN');
        }
    }

    public function test_superadmin_can_manage_plans_and_changes_are_audited(): void
    {
        $token = $this->tokenFor($this->superadmin());

        $created = $this->withToken($token)->postJson('/api/v1/superadmin/plans', [
            'name' => 'Starter',
            'code' => 'starter',
            'price' => 99000,
            'billing_period' => 'monthly',
            'max_outlets' => 1,
            'max_devices' => 2,
            'max_users' => 3,
            'max_products' => 100,
        ])->assertCreated()
            ->assertJsonPath('data.code', 'starter')
            ->assertJsonPath('data.price', 99000)
            ->assertJsonPath('data.is_active', true);

        $planId = $created->json('data.id');

        $this->withToken($token)->putJson("/api/v1/superadmin/plans/{$planId}", [
            'name' => 'Starter Plus',
            'code' => 'starter-plus',
            'price' => 149000,
            'billing_period' => 'monthly',
            'max_outlets' => 2,
            'max_devices' => 4,
            'max_users' => 6,
            'max_products' => 250,
        ])->assertOk()
            ->assertJsonPath('data.name', 'Starter Plus')
            ->assertJsonPath('data.code', 'starter-plus')
            ->assertJsonPath('data.price', 149000);

        $this->withToken($token)->patchJson("/api/v1/superadmin/plans/{$planId}/status", [
            'is_active' => false,
        ])->assertOk()
            ->assertJsonPath('data.is_active', false);

        $this->assertDatabaseHas('audit_logs', [
            'business_id' => null,
            'action' => 'superadmin.plan.created',
            'entity_type' => 'plan',
            'entity_id' => $planId,
        ]);
        $this->assertNotNull(DB::table('audit_logs')
            ->where('action', 'superadmin.plan.updated')
            ->where('entity_id', $planId)
            ->value('before'));
        $this->assertNotNull(DB::table('audit_logs')
            ->where('action', 'superadmin.plan.status_updated')
            ->where('entity_id', $planId)
            ->value('after'));
    }

    public function test_superadmin_provisions_business_owner_outlet_subscription_and_idempotency_deduplicates(): void
    {
        $token = $this->tokenFor($this->superadmin());
        $planId = $this->plan('starter');
        $key = (string) Str::uuid();

        $payload = [
            'business' => ['name' => 'Kedai Baru'],
            'owner' => [
                'name' => 'Owner Baru',
                'email' => 'owner-baru@example.test',
                'password' => 'password',
            ],
            'outlet' => ['name' => 'Outlet Utama'],
            'subscription' => [
                'plan_id' => $planId,
                'status' => 'trial',
                'trial_ends_at' => now()->addDays(14)->toDateString(),
                'current_period_ends_at' => now()->addMonth()->toDateString(),
                'notes' => 'Bootstrap tenant',
            ],
        ];

        $first = $this->withToken($token)
            ->withHeaders(['Idempotency-Key' => $key])
            ->postJson('/api/v1/superadmin/businesses', $payload)
            ->assertCreated()
            ->assertJsonPath('data.business.name', 'Kedai Baru')
            ->assertJsonPath('data.owner.email', 'owner-baru@example.test')
            ->assertJsonPath('data.outlet.name', 'Outlet Utama')
            ->assertJsonPath('data.subscription.plan_id', $planId);

        $businessId = $first->json('data.business.id');

        $this->withToken($token)
            ->withHeaders(['Idempotency-Key' => $key])
            ->postJson('/api/v1/superadmin/businesses', $payload)
            ->assertCreated()
            ->assertExactJson($first->json());

        $this->assertSame(1, DB::table('businesses')->where('name', 'Kedai Baru')->count());
        $this->assertSame(1, DB::table('users')->where('email', 'owner-baru@example.test')->count());
        $this->assertSame(1, DB::table('outlets')->where('business_id', $businessId)->count());
        $this->assertSame(1, DB::table('subscriptions')->where('business_id', $businessId)->whereNull('deleted_at')->count());

        $payload['business']['name'] = 'Kedai Lain';
        $this->withToken($token)
            ->withHeaders(['Idempotency-Key' => $key])
            ->postJson('/api/v1/superadmin/businesses', $payload)
            ->assertConflict()
            ->assertJsonPath('error.code', 'IDEMPOTENCY_MISMATCH');
    }

    public function test_superadmin_assigns_single_current_subscription_with_audit_and_idempotency(): void
    {
        $superadminId = $this->superadmin();
        $token = $this->tokenFor($superadminId);
        $businessId = $this->business('Tenant Subscription');
        $starter = $this->plan('starter');
        $growth = $this->plan('growth');
        $key = (string) Str::uuid();

        $first = $this->withToken($token)
            ->withHeaders(['Idempotency-Key' => $key])
            ->postJson("/api/v1/superadmin/businesses/{$businessId}/subscription", [
                'plan_id' => $starter,
                'status' => 'active',
                'current_period_ends_at' => now()->addMonth()->toDateString(),
                'max_outlets_override' => 3,
                'notes' => 'Initial assignment',
            ])->assertCreated()
            ->assertJsonPath('data.business_id', $businessId)
            ->assertJsonPath('data.plan_id', $starter)
            ->assertJsonPath('data.status', 'active');

        $this->withToken($token)
            ->withHeaders(['Idempotency-Key' => $key])
            ->postJson("/api/v1/superadmin/businesses/{$businessId}/subscription", [
                'plan_id' => $starter,
                'status' => 'active',
                'current_period_ends_at' => now()->addMonth()->toDateString(),
                'max_outlets_override' => 3,
                'notes' => 'Initial assignment',
            ])->assertCreated()
            ->assertExactJson($first->json());

        $this->withToken($token)
            ->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson("/api/v1/superadmin/businesses/{$businessId}/subscription", [
                'plan_id' => $growth,
                'status' => 'suspended',
                'max_products_override' => 500,
                'notes' => 'Manual suspension',
            ])->assertOk()
            ->assertJsonPath('data.plan_id', $growth)
            ->assertJsonPath('data.status', 'suspended');

        $this->assertSame(1, DB::table('subscriptions')->where('business_id', $businessId)->whereNull('deleted_at')->count());
        $this->assertDatabaseHas('audit_logs', [
            'business_id' => $businessId,
            'actor_id' => $superadminId,
            'action' => 'superadmin.subscription.assigned',
            'entity_type' => 'subscription',
        ]);
        $this->assertNotNull(DB::table('audit_logs')
            ->where('business_id', $businessId)
            ->where('action', 'superadmin.subscription.updated')
            ->value('before'));
    }

    public function test_bootstrap_seeder_is_idempotent_and_superadmin_can_login(): void
    {
        $this->seed(SuperadminBootstrapSeeder::class);
        $this->seed(SuperadminBootstrapSeeder::class);

        $this->assertSame(1, DB::table('users')->where('role', 'superadmin')->whereNull('business_id')->count());
        $this->assertSame(1, DB::table('plans')->where('code', 'starter')->whereNull('deleted_at')->count());

        $this->postJson('/api/v1/auth/login', [
            'email' => env('NOJPOS_SUPERADMIN_EMAIL', 'superadmin@nojpos.test'),
            'password' => env('NOJPOS_SUPERADMIN_PASSWORD', 'password'),
            'device_uuid' => 'superadmin-web',
        ])->assertOk()
            ->assertJsonPath('data.user.role', 'superadmin')
            ->assertJsonPath('data.business', null);
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

    private function plan(string $code): string
    {
        $id = (string) Str::uuid();
        DB::table('plans')->insert([
            'id' => $id,
            'name' => Str::title($code),
            'code' => $code,
            'active_code' => $code,
            'price' => 100000,
            'billing_period' => 'monthly',
            'max_outlets' => 1,
            'max_devices' => 2,
            'max_users' => 5,
            'max_products' => 100,
            'is_active' => true,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function superadmin(): string
    {
        return $this->user(null, 'superadmin@example.test', 'superadmin');
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
