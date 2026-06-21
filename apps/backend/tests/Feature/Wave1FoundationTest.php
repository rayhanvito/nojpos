<?php

namespace Tests\Feature;

use App\Services\AuditLogService;
use App\Services\IdempotencyService;
use App\Support\ApiResponse;
use App\Support\Nojpos;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Facades\Route;
use Illuminate\Support\Str;
use RuntimeException;
use Tests\TestCase;

class Wave1FoundationTest extends TestCase
{
    use RefreshDatabase;

    protected function setUp(): void
    {
        parent::setUp();

        Route::middleware(['api', 'auth:sanctum', 'idempotency'])
            ->post('/api/v1/__wave1/effects', function (Request $request) {
                $user = $request->user();
                $effectId = (string) Str::uuid();
                $method = (string) $request->input('method', 'cash');

                return DB::transaction(function () use ($request, $user, $effectId, $method) {
                    DB::table('payment_method_configs')->insert([
                        'id' => $effectId,
                        'business_id' => $user->business_id,
                        'outlet_id' => $request->input('outlet_id'),
                        'method' => $method,
                        'is_cash' => $method === 'cash',
                        'created_at' => now(),
                        'updated_at' => now(),
                    ]);

                    Nojpos::audit(
                        $user->business_id,
                        $user->id,
                        'test.effect',
                        'payment_method_config',
                        $effectId,
                        null,
                        ['method' => $method],
                    );

                    return ApiResponse::success(['effect_id' => $effectId, 'method' => $method], [], 201);
                });
            });
    }

    public function test_completed_replay_returns_stored_response_and_exactly_one_effect_and_audit(): void
    {
        [$businessId, $outletId, $token] = $this->tenantContext();
        $key = (string) Str::uuid();
        $payload = ['outlet_id' => $outletId, 'method' => 'cash'];

        $first = $this->withToken($token)
            ->withHeaders(['Idempotency-Key' => $key, 'X-Request-Id' => 'req-1'])
            ->postJson('/api/v1/__wave1/effects', $payload)
            ->assertCreated();

        $second = $this->withToken($token)
            ->withHeaders(['Idempotency-Key' => $key, 'X-Request-Id' => 'req-2'])
            ->postJson('/api/v1/__wave1/effects', $payload)
            ->assertCreated();

        $second->assertExactJson($first->json());

        $this->assertSame(1, DB::table('payment_method_configs')->where('business_id', $businessId)->count());
        $this->assertSame(1, DB::table('audit_logs')->where('business_id', $businessId)->where('action', 'test.effect')->count());
        $this->assertDatabaseHas('audit_logs', [
            'business_id' => $businessId,
            'action' => 'test.effect',
            'idempotency_key' => $key,
            'request_id' => 'req-1',
            'event_version' => 1,
        ]);
        $this->assertDatabaseHas('idempotency_keys', [
            'business_id' => $businessId,
            'endpoint' => 'POST api/v1/__wave1/effects',
            'key' => $key,
            'state' => 'completed',
            'status' => 201,
        ]);
    }

    public function test_same_key_different_body_returns_409_idempotency_mismatch(): void
    {
        [$businessId, $outletId, $token] = $this->tenantContext();
        $key = (string) Str::uuid();

        $this->withToken($token)
            ->withHeaders(['Idempotency-Key' => $key])
            ->postJson('/api/v1/__wave1/effects', ['outlet_id' => $outletId, 'method' => 'cash'])
            ->assertCreated();

        $this->withToken($token)
            ->withHeaders(['Idempotency-Key' => $key])
            ->postJson('/api/v1/__wave1/effects', ['outlet_id' => $outletId, 'method' => 'qris'])
            ->assertConflict()
            ->assertJsonPath('error.code', 'IDEMPOTENCY_MISMATCH');

        $this->assertSame(1, DB::table('payment_method_configs')->where('business_id', $businessId)->count());
        $this->assertSame(1, DB::table('audit_logs')->where('business_id', $businessId)->where('action', 'test.effect')->count());
    }

    public function test_idempotency_fingerprint_includes_user_and_terminal_context(): void
    {
        $service = app(IdempotencyService::class);
        $payload = ['outlet_id' => (string) Str::uuid(), 'device_id' => (string) Str::uuid(), 'cashier_id' => (string) Str::uuid()];
        $first = Request::create('/api/v1/__wave1/effects', 'POST', $payload);
        $first->setUserResolver(fn () => (object) ['id' => 'user-a']);
        $second = Request::create('/api/v1/__wave1/effects', 'POST', $payload);
        $second->setUserResolver(fn () => (object) ['id' => 'user-b']);

        $this->assertNotSame(
            $service->hashPayload($service->fingerprintPayload($first)),
            $service->hashPayload($service->fingerprintPayload($second)),
        );
    }

    public function test_parallel_same_key_second_request_gets_in_progress_without_second_effect(): void
    {
        [$businessId, $outletId, $token] = $this->tenantContext();
        $key = (string) Str::uuid();
        $payload = ['outlet_id' => $outletId, 'method' => 'cash'];

        DB::table('idempotency_keys')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'endpoint' => 'POST api/v1/__wave1/effects',
            'key' => $key,
            'request_hash' => app(IdempotencyService::class)->hashPayload([
                'context' => [
                    'user_id' => DB::table('personal_access_tokens')->where('token', hash('sha256', $token))->value('tokenable_id'),
                    'outlet_id' => $outletId,
                    'device_id' => null,
                    'cashier_id' => null,
                ],
                'payload' => $payload,
            ]),
            'state' => 'in_progress',
            'reserved_at' => now(),
            'response_snapshot' => '',
            'status' => 0,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        $this->withToken($token)
            ->withHeaders(['Idempotency-Key' => $key])
            ->postJson('/api/v1/__wave1/effects', $payload)
            ->assertConflict()
            ->assertJsonPath('error.code', 'IDEMPOTENCY_IN_PROGRESS');

        $this->assertSame(0, DB::table('payment_method_configs')->where('business_id', $businessId)->count());
        $this->assertSame(0, DB::table('audit_logs')->where('business_id', $businessId)->where('action', 'test.effect')->count());
    }

    public function test_audit_failure_rolls_back_mandatory_audited_mutation(): void
    {
        [$businessId, $outletId] = $this->tenantContext();

        app()->bind(AuditLogService::class, fn () => new class extends AuditLogService
        {
            public function record(
                ?string $businessId,
                ?string $actorId,
                string $action,
                ?string $entityType = null,
                ?string $entityId = null,
                ?array $before = null,
                ?array $after = null,
                array $context = [],
            ): string {
                throw new RuntimeException('audit failed');
            }
        });

        try {
            DB::transaction(function () use ($businessId, $outletId): void {
                $effectId = (string) Str::uuid();

                DB::table('payment_method_configs')->insert([
                    'id' => $effectId,
                    'business_id' => $businessId,
                    'outlet_id' => $outletId,
                    'method' => 'cash',
                    'is_cash' => true,
                    'created_at' => now(),
                    'updated_at' => now(),
                ]);

                Nojpos::audit($businessId, null, 'test.effect', 'payment_method_config', $effectId);
            });

            $this->fail('Expected audit failure to abort the transaction.');
        } catch (RuntimeException $exception) {
            $this->assertSame('audit failed', $exception->getMessage());
        }

        $this->assertSame(0, DB::table('payment_method_configs')->where('business_id', $businessId)->count());
        $this->assertSame(0, DB::table('audit_logs')->where('business_id', $businessId)->where('action', 'test.effect')->count());
    }

    private function tenantContext(): array
    {
        $businessId = (string) Str::uuid();
        $outletId = (string) Str::uuid();
        $userId = (string) Str::uuid();

        DB::table('businesses')->insert([
            'id' => $businessId,
            'name' => 'Wave 1 Tenant',
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        DB::table('outlets')->insert([
            'id' => $outletId,
            'business_id' => $businessId,
            'name' => 'Outlet',
            'service_charge_rate' => 0,
            'tax_rate' => 0,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        DB::table('users')->insert([
            'id' => $userId,
            'business_id' => $businessId,
            'name' => 'Owner',
            'email' => 'owner-'.Str::lower(Str::random(8)).'@example.test',
            'password' => Hash::make('password'),
            'role' => 'owner',
            'pin_hash' => Hash::make('1234'),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

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

        return [$businessId, $outletId, 'plain-'.$userId, $userId];
    }
}
