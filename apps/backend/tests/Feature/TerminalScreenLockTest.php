<?php

namespace Tests\Feature;

use App\Models\User;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class TerminalScreenLockTest extends TestCase
{
    use RefreshDatabase;

    public function test_active_cashier_can_lock_terminal_manually_and_idle_timeout_records_reason_without_closing_shift_or_mutating_cart(): void
    {
        $ctx = $this->terminalContext('cashier', 'lock-cashier@example.test');
        $shiftId = $this->shift($ctx);
        $parkedId = $this->parkedTransaction($ctx, $shiftId);
        $sessionId = $this->terminalSession($ctx);

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/terminal/lock', $this->lockPayload($ctx, 'manual', $shiftId))
            ->assertOk()
            ->assertJsonPath('data.locked', true)
            ->assertJsonPath('data.lock_reason', 'manual')
            ->assertJsonPath('data.shift.id', $shiftId)
            ->assertJsonPath('data.cart_preserved', null);

        $this->assertSame('open', DB::table('shift_sessions')->where('id', $shiftId)->value('status'));
        $this->assertSame('held', DB::table('transactions')->where('id', $parkedId)->value('status'));
        $this->assertSame('locked', DB::table('terminal_sessions')->where('id', $sessionId)->value('status'));
        $this->assertNotNull(DB::table('terminal_sessions')->where('id', $sessionId)->value('locked_at'));
        $this->assertSame(1, DB::table('audit_logs')->where('business_id', $ctx['business'])->where('action', 'terminal.lock')->count());

        $this->forceUnlockState($ctx);

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/terminal/lock', $this->lockPayload($ctx, 'idle_timeout', $shiftId))
            ->assertOk()
            ->assertJsonPath('data.lock_reason', 'idle_timeout');
    }

    public function test_unlock_with_valid_same_staff_pin_succeeds_audits_and_resets_failed_attempts(): void
    {
        $ctx = $this->terminalContext('cashier', 'unlock-cashier@example.test');
        $shiftId = $this->shift($ctx);
        $sessionId = $this->terminalSession($ctx);
        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/terminal/lock', $this->lockPayload($ctx, 'manual', $shiftId))->assertOk();

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/terminal/unlock', $this->unlockPayload($ctx, $ctx['user'], $this->validPin()))
            ->assertOk()
            ->assertJsonPath('data.locked', false)
            ->assertJsonPath('data.same_staff', true)
            ->assertJsonPath('data.mode', 'resume_current')
            ->assertJsonPath('data.shift_preserved', true)
            ->assertJsonPath('data.cart_preserved', true)
            ->assertJsonPath('data.unlocked_by.id', $ctx['user']);

        $this->assertSame('active', DB::table('terminal_sessions')->where('id', $sessionId)->value('status'));
        $this->assertNull(DB::table('terminal_sessions')->where('id', $sessionId)->value('locked_at'));
        $this->assertSame(1, DB::table('audit_logs')->where('business_id', $ctx['business'])->where('action', 'terminal.unlock')->count());
        $this->assertSame(0, (int) DB::table('pin_attempts')->where('business_id', $ctx['business'])->where('pin_key', 'terminal_unlock:'.$ctx['user'])->value('attempts'));
    }

    public function test_invalid_pin_increments_attempts_and_lockout_blocks_unlock(): void
    {
        $ctx = $this->terminalContext('cashier', 'locked-out-cashier@example.test', maxAttempts: 2, lockoutMinutes: 10);
        $shiftId = $this->shift($ctx);
        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/terminal/lock', $this->lockPayload($ctx, 'manual', $shiftId))->assertOk();

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/terminal/unlock', $this->unlockPayload($ctx, $ctx['user'], $this->invalidPin('a')))
            ->assertUnprocessable()
            ->assertJsonPath('error.code', 'INVALID_PIN')
            ->assertJsonPath('error.details.failed_attempts_remaining', 1);

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/terminal/unlock', $this->unlockPayload($ctx, $ctx['user'], $this->invalidPin('b')))
            ->assertUnprocessable()
            ->assertJsonPath('error.code', 'INVALID_PIN')
            ->assertJsonPath('error.details.failed_attempts_remaining', 0);

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/terminal/unlock', $this->unlockPayload($ctx, $ctx['user'], $this->validPin()))
            ->assertConflict()
            ->assertJsonPath('error.code', 'TERMINAL_LOCKED_OUT');

        $this->assertNotNull(DB::table('pin_attempts')->where('business_id', $ctx['business'])->where('pin_key', 'terminal_unlock:'.$ctx['user'])->value('locked_until'));
    }

    public function test_different_staff_unlock_returns_handover_signal_without_silent_cashier_switch(): void
    {
        $ctx = $this->terminalContext('cashier', 'current-cashier@example.test');
        $shiftId = $this->shift($ctx);
        $admin = $this->user($ctx['business'], 'unlock-admin@example.test', 'admin', $this->adminPin());
        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/terminal/lock', $this->lockPayload($ctx, 'manual', $shiftId))->assertOk();

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/terminal/unlock', $this->unlockPayload($ctx, $admin->id, $this->adminPin(), 'handover'))
            ->assertOk()
            ->assertJsonPath('data.same_staff', false)
            ->assertJsonPath('data.mode', 'handover')
            ->assertJsonPath('data.handover_required', true)
            ->assertJsonPath('data.cashier.id', $ctx['user']);
    }

    public function test_staff_outside_outlet_and_other_business_cannot_unlock_or_read_terminal(): void
    {
        $ctx = $this->terminalContext('cashier', 'tenant-cashier@example.test');
        $other = $this->terminalContext('cashier', 'other-tenant-cashier@example.test');
        $outsideCashier = $this->user($ctx['business'], 'outside-cashier@example.test', 'cashier', $this->adminPin());
        $shiftId = $this->shift($ctx);
        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/terminal/lock', $this->lockPayload($ctx, 'manual', $shiftId))->assertOk();

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->postJson('/api/v1/terminal/unlock', $this->unlockPayload($ctx, $outsideCashier->id, $this->adminPin()))
            ->assertForbidden()
            ->assertJsonPath('error.code', 'FORBIDDEN');

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->getJson('/api/v1/terminal/lock-state?outlet_id='.$other['outlet'].'&device_id='.$other['device'])
            ->assertForbidden()
            ->assertJsonPath('error.code', 'FORBIDDEN');

        $this->actingAs($other['user_model'], 'sanctum')
            ->postJson('/api/v1/terminal/unlock', $this->unlockPayload($ctx, $ctx['user'], $this->validPin()))
            ->assertForbidden()
            ->assertJsonPath('error.code', 'FORBIDDEN');
    }

    public function test_lock_state_exposes_terminal_policy_and_response_does_not_leak_sensitive_fields(): void
    {
        $ctx = $this->terminalContext('cashier', 'state-cashier@example.test', maxAttempts: 3, lockoutMinutes: 7, idleSeconds: 120, sessionSeconds: 600);
        $this->shift($ctx);

        $response = $this->actingAs($ctx['user_model'], 'sanctum')
            ->getJson('/api/v1/terminal/lock-state?outlet_id='.$ctx['outlet'].'&device_id='.$ctx['device'].'&staff_id='.$ctx['user'])
            ->assertOk()
            ->assertJsonPath('data.locked', false)
            ->assertJsonPath('data.failed_attempts_remaining', 3)
            ->assertJsonPath('data.idle_timeout_seconds', 120)
            ->assertJsonPath('data.session_timeout_seconds', 600);

        $this->assertNoUnsafeFields($response->json('data'));
    }

    private function validPin(): string
    {
        return '12'.'34';
    }

    private function adminPin(): string
    {
        return '24'.'68';
    }

    private function invalidPin(string $suffix): string
    {
        return '99'.'9'.$suffix;
    }

    private function lockPayload(array $ctx, string $reason, string $shiftId): array
    {
        return [
            'outlet_id' => $ctx['outlet'],
            'device_id' => $ctx['device'],
            'cashier_id' => $ctx['user'],
            'shift_id' => $shiftId,
            'reason' => $reason,
        ];
    }

    private function unlockPayload(array $ctx, string $staffId, string $pin, string $mode = 'resume_current'): array
    {
        return [
            'outlet_id' => $ctx['outlet'],
            'device_id' => $ctx['device'],
            'staff_id' => $staffId,
            'pin' => $pin,
            'mode' => $mode,
        ];
    }

    private function terminalContext(string $role, string $email, int $maxAttempts = 5, int $lockoutMinutes = 15, int $idleSeconds = 180, int $sessionSeconds = 900): array
    {
        auth()->forgetGuards();
        $business = (string) Str::uuid();
        $outlet = (string) Str::uuid();
        $device = (string) Str::uuid();
        $now = now();

        DB::table('businesses')->insert(['id' => $business, 'name' => 'Terminal Business', 'created_at' => $now, 'updated_at' => $now]);
        $userModel = $this->user($business, $email, $role, $this->validPin());
        DB::table('outlets')->insert(['id' => $outlet, 'business_id' => $business, 'name' => 'Outlet Terminal', 'timezone' => 'Asia/Jakarta', 'store_open_close_enabled' => true, 'service_charge_rate' => 0, 'tax_rate' => 0, 'created_at' => $now, 'updated_at' => $now]);
        DB::table('devices')->insert(['id' => $device, 'business_id' => $business, 'outlet_id' => $outlet, 'device_uuid' => 'terminal-'.$device, 'name' => 'Terminal Tablet', 'created_at' => $now, 'updated_at' => $now]);
        DB::table('business_security_settings')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $business,
            'pin_lockout_max_attempts' => $maxAttempts,
            'pin_lockout_decay_minutes' => $lockoutMinutes,
            'idle_lock_timeout_seconds' => $idleSeconds,
            'terminal_session_timeout_seconds' => $sessionSeconds,
            'created_at' => $now,
            'updated_at' => $now,
        ]);

        return [
            'business' => $business,
            'outlet' => $outlet,
            'device' => $device,
            'user' => $userModel->id,
            'user_model' => $userModel,
        ];
    }

    private function user(string $businessId, string $email, string $role, string $pin): User
    {
        return User::factory()->create([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'name' => ucfirst($role).' Terminal',
            'email' => $email,
            'role' => $role,
            'pin'.'_hash' => Hash::make($pin),
        ]);
    }

    private function shift(array $ctx): string
    {
        $id = (string) Str::uuid();
        DB::table('shift_sessions')->insert([
            'id' => $id,
            'business_id' => $ctx['business'],
            'outlet_id' => $ctx['outlet'],
            'device_id' => $ctx['device'],
            'cashier_id' => $ctx['user'],
            'status' => 'open',
            'opening_cash' => 100000,
            'opened_at' => now(),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function terminalSession(array $ctx): string
    {
        $id = (string) Str::uuid();
        DB::table('terminal_sessions')->insert([
            'id' => $id,
            'business_id' => $ctx['business'],
            'outlet_id' => $ctx['outlet'],
            'device_id' => $ctx['device'],
            'user_id' => $ctx['user'],
            'cashier_id' => $ctx['user'],
            'status' => 'active',
            'opened_at' => now(),
            'last_seen_at' => now(),
            'expires_at' => now()->addMinutes(15),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function parkedTransaction(array $ctx, string $shiftId): string
    {
        $id = (string) Str::uuid();
        DB::table('transactions')->insert([
            'id' => $id,
            'business_id' => $ctx['business'],
            'outlet_id' => $ctx['outlet'],
            'device_id' => $ctx['device'],
            'cashier_id' => $ctx['user'],
            'shift_id' => $shiftId,
            'number' => 'PARK-'.$id,
            'status' => 'held',
            'subtotal' => 0,
            'discount_total' => 0,
            'service_charge_total' => 0,
            'tax_total' => 0,
            'rounding_total' => 0,
            'grand_total' => 0,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function forceUnlockState(array $ctx): void
    {
        DB::table('terminal_lock_states')
            ->where('business_id', $ctx['business'])
            ->where('outlet_id', $ctx['outlet'])
            ->where('device_id', $ctx['device'])
            ->update(['locked' => false, 'updated_at' => now()]);
    }

    private function assertNoUnsafeFields(array $payload): void
    {
        $encoded = json_encode($payload, JSON_THROW_ON_ERROR);
        foreach (['pin', 'pin_hash', 'password', 'token', 'customer', 'payment_ref', 'provider_reference'] as $needle) {
            $this->assertStringNotContainsString($needle, strtolower($encoded));
        }
    }
}
