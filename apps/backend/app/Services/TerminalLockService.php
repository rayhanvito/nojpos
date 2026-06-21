<?php

namespace App\Services;

use App\Models\BusinessSecuritySetting;
use App\Models\User;
use App\Support\Nojpos;
use Carbon\CarbonImmutable;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;

class TerminalLockService
{
    public function __construct(private readonly TerminalSessionService $terminalSessions) {}

    /**
     * @param  array{outlet_id:string, device_id:string, reason:string, cashier_id?:string|null, shift_id?:string|null}  $data
     */
    public function lock(User $actor, array $data): array
    {
        return DB::transaction(function () use ($actor, $data): array {
            $businessId = (string) $actor->business_id;
            $this->assertOutletAndDevice($businessId, $data['outlet_id'], $data['device_id']);
            $shift = $this->activeShift($businessId, $data['outlet_id'], $data['device_id'], $data['cashier_id'] ?? null, $data['shift_id'] ?? null);
            $state = $this->stateForUpdate($businessId, $data['outlet_id'], $data['device_id'], $actor->id, $shift?->cashier_id, $shift?->id);

            if ((bool) $state->locked) {
                throw new TerminalLockException('TERMINAL_ALREADY_LOCKED', 'Terminal is already locked.', [], 409);
            }

            $before = $this->stateSnapshot($state);
            $now = now();
            DB::table('terminal_lock_states')
                ->where('id', $state->id)
                ->update([
                    'current_user_id' => $actor->id,
                    'current_cashier_id' => $shift?->cashier_id,
                    'shift_id' => $shift?->id,
                    'locked' => true,
                    'locked_at' => $now,
                    'lock_reason' => $data['reason'],
                    'unlocked_at' => null,
                    'unlocked_by_user_id' => null,
                    'updated_at' => $now,
                ]);

            $this->terminalSessions->lock($actor, $data['outlet_id'], $data['device_id']);

            $fresh = DB::table('terminal_lock_states')->where('id', $state->id)->first();
            Nojpos::audit($businessId, $actor->id, 'terminal.lock', 'terminal_lock_state', $fresh->id, $before, $this->stateSnapshot($fresh));

            return $this->payload($businessId, $fresh);
        });
    }

    /**
     * @param  array{outlet_id:string, device_id:string, staff_id:string, pin:string, mode?:string|null}  $data
     */
    public function unlock(User $actor, array $data): array
    {
        $result = DB::transaction(function () use ($actor, $data): array {
            $businessId = (string) $actor->business_id;
            $this->assertOutletAndDevice($businessId, $data['outlet_id'], $data['device_id']);
            $state = $this->stateForUpdate($businessId, $data['outlet_id'], $data['device_id']);

            if (! (bool) $state->locked) {
                throw new TerminalLockException('TERMINAL_NOT_LOCKED', 'Terminal is not locked.', [], 409);
            }

            $staff = $this->staffForUnlock($businessId, $data['outlet_id'], $data['staff_id']);
            $attempt = $this->pinAttemptForUpdate($businessId, $staff->id);
            if ($attempt && $attempt->locked_until && now()->lessThan($attempt->locked_until)) {
                throw new TerminalLockException('TERMINAL_LOCKED_OUT', 'PIN is temporarily locked.', [
                    'lockout_until' => CarbonImmutable::parse($attempt->locked_until)->toISOString(),
                    'failed_attempts_remaining' => 0,
                ], 409);
            }

            if (! $staff->pin_hash || ! Hash::check($data['pin'], $staff->pin_hash)) {
                $freshAttempt = $this->recordPinFailure($businessId, $staff->id, $attempt);
                $settings = $this->settings($businessId);
                $remaining = max(0, (int) $settings->pin_lockout_max_attempts - (int) $freshAttempt->attempts);
                $details = ['failed_attempts_remaining' => $remaining];
                if ($freshAttempt->locked_until) {
                    $details['lockout_until'] = CarbonImmutable::parse($freshAttempt->locked_until)->toISOString();
                }

                return ['__terminal_lock_error' => new TerminalLockException('INVALID_PIN', 'PIN is invalid.', $details, 422)];
            }

            $this->resetPinAttempt($businessId, $staff->id, $attempt);
            $before = $this->stateSnapshot($state);
            $sameStaff = $state->current_cashier_id === $staff->id || $state->current_user_id === $staff->id;
            $mode = $sameStaff ? 'resume_current' : ($data['mode'] ?? 'resume_current');
            $now = now();

            DB::table('terminal_lock_states')
                ->where('id', $state->id)
                ->update([
                    'locked' => false,
                    'failed_attempts' => 0,
                    'lockout_until' => null,
                    'unlocked_at' => $now,
                    'unlocked_by_user_id' => $staff->id,
                    'updated_at' => $now,
                ]);

            $this->terminalSessions->unlock($actor, $data['outlet_id'], $data['device_id'], $staff->id);

            $fresh = DB::table('terminal_lock_states')->where('id', $state->id)->first();
            Nojpos::audit($businessId, $staff->id, 'terminal.unlock', 'terminal_lock_state', $fresh->id, $before, [
                ...$this->stateSnapshot($fresh),
                'mode' => $mode,
                'same_staff' => $sameStaff,
                'unlocked_by_user_id' => $staff->id,
            ]);

            return [
                ...$this->payload($businessId, $fresh),
                'unlocked_by' => $this->userPayload($staff),
                'mode' => $mode,
                'same_staff' => $sameStaff,
                'handover_required' => ! $sameStaff,
                'shift_preserved' => $fresh->shift_id !== null,
                'cart_preserved' => true,
            ];
        });

        if (($result['__terminal_lock_error'] ?? null) instanceof TerminalLockException) {
            throw $result['__terminal_lock_error'];
        }

        return $result;
    }

    /**
     * @param  array{outlet_id:string, device_id:string, staff_id?:string|null}  $data
     */
    public function state(User $actor, array $data): array
    {
        $businessId = (string) $actor->business_id;
        $this->assertOutletAndDevice($businessId, $data['outlet_id'], $data['device_id']);
        $state = $this->stateForRead($businessId, $data['outlet_id'], $data['device_id']);
        $payload = $this->payload($businessId, $state);

        if (! empty($data['staff_id'])) {
            $staff = $this->staffForUnlock($businessId, $data['outlet_id'], $data['staff_id']);
            $attempt = $this->pinAttempt($businessId, $staff->id);
            $settings = $this->settings($businessId);
            $payload['failed_attempts_remaining'] = max(0, (int) $settings->pin_lockout_max_attempts - (int) ($attempt->attempts ?? 0));
            $payload['lockout_until'] = $attempt?->locked_until ? CarbonImmutable::parse($attempt->locked_until)->toISOString() : null;
        }

        return $payload;
    }

    private function stateForRead(string $businessId, string $outletId, string $deviceId): object
    {
        $state = DB::table('terminal_lock_states')
            ->where('business_id', $businessId)
            ->where('outlet_id', $outletId)
            ->where('device_id', $deviceId)
            ->first();

        if ($state) {
            return $state;
        }

        return $this->createState($businessId, $outletId, $deviceId);
    }

    private function stateForUpdate(string $businessId, string $outletId, string $deviceId, ?string $currentUserId = null, ?string $cashierId = null, ?string $shiftId = null): object
    {
        $state = DB::table('terminal_lock_states')
            ->where('business_id', $businessId)
            ->where('outlet_id', $outletId)
            ->where('device_id', $deviceId)
            ->lockForUpdate()
            ->first();

        if ($state) {
            return $state;
        }

        $this->createState($businessId, $outletId, $deviceId, $currentUserId, $cashierId, $shiftId);

        return DB::table('terminal_lock_states')
            ->where('business_id', $businessId)
            ->where('outlet_id', $outletId)
            ->where('device_id', $deviceId)
            ->lockForUpdate()
            ->first();
    }

    private function createState(string $businessId, string $outletId, string $deviceId, ?string $currentUserId = null, ?string $cashierId = null, ?string $shiftId = null): object
    {
        $id = (string) Str::uuid();
        DB::table('terminal_lock_states')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'current_user_id' => $currentUserId,
            'current_cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'locked' => false,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return DB::table('terminal_lock_states')->where('id', $id)->first();
    }

    private function assertOutletAndDevice(string $businessId, string $outletId, string $deviceId): void
    {
        $outletExists = DB::table('outlets')
            ->where('business_id', $businessId)
            ->where('id', $outletId)
            ->whereNull('deleted_at')
            ->exists();
        $deviceExists = DB::table('devices')
            ->where('business_id', $businessId)
            ->where('id', $deviceId)
            ->where('outlet_id', $outletId)
            ->whereNull('deleted_at')
            ->exists();

        if (! $outletExists || ! $deviceExists) {
            throw new TerminalLockException('FORBIDDEN', 'Terminal context is outside the current business scope.', [], 403);
        }
    }

    private function activeShift(string $businessId, string $outletId, string $deviceId, ?string $cashierId, ?string $shiftId): ?object
    {
        $query = DB::table('shift_sessions')
            ->where('business_id', $businessId)
            ->where('outlet_id', $outletId)
            ->where('device_id', $deviceId)
            ->where('status', 'open')
            ->whereNull('deleted_at');

        if ($shiftId) {
            $query->where('id', $shiftId);
        }

        if ($cashierId) {
            $query->where('cashier_id', $cashierId);
        }

        return $query->latest('opened_at')->first();
    }

    private function staffForUnlock(string $businessId, string $outletId, string $staffId): object
    {
        $staff = DB::table('users')
            ->where('business_id', $businessId)
            ->where('id', $staffId)
            ->whereIn('role', ['cashier', 'admin', 'owner'])
            ->whereNull('deleted_at')
            ->first();

        if (! $staff) {
            throw new TerminalLockException('FORBIDDEN', 'Staff is not authorized for this outlet.', [], 403);
        }

        $hasShiftAtOutlet = DB::table('shift_sessions')
            ->where('business_id', $businessId)
            ->where('outlet_id', $outletId)
            ->where('cashier_id', $staffId)
            ->where('status', 'open')
            ->whereNull('deleted_at')
            ->exists();

        if ($staff->role === 'cashier' && ! $hasShiftAtOutlet) {
            throw new TerminalLockException('FORBIDDEN', 'Staff is not authorized for this outlet.', [], 403);
        }

        return $staff;
    }

    private function pinAttemptForUpdate(string $businessId, string $staffId): ?object
    {
        return DB::table('pin_attempts')
            ->where('business_id', $businessId)
            ->where('pin_key', $this->pinKey($staffId))
            ->lockForUpdate()
            ->first();
    }

    private function pinAttempt(string $businessId, string $staffId): ?object
    {
        return DB::table('pin_attempts')
            ->where('business_id', $businessId)
            ->where('pin_key', $this->pinKey($staffId))
            ->first();
    }

    private function recordPinFailure(string $businessId, string $staffId, ?object $attempt): object
    {
        $settings = $this->settings($businessId);
        $attempt ??= DB::table('pin_attempts')
            ->where('business_id', $businessId)
            ->where('pin_key', $this->pinKey($staffId))
            ->orderByDesc('attempts')
            ->first();

        $attempts = (int) ($attempt->attempts ?? 0) + 1;
        $values = [
            'outlet_id' => null,
            'device_id' => null,
            'pin_key' => $this->pinKey($staffId),
            'attempts' => $attempts,
            'locked_until' => $attempts >= (int) $settings->pin_lockout_max_attempts ? now()->addMinutes((int) $settings->pin_lockout_decay_minutes) : null,
            'updated_at' => now(),
        ];

        if ($attempt) {
            DB::table('pin_attempts')->where('id', $attempt->id)->update($values);
        } else {
            DB::table('pin_attempts')->insert($values + [
                'id' => (string) Str::uuid(),
                'business_id' => $businessId,
                'created_at' => now(),
            ]);
        }

        $totalAttempts = (int) DB::table('pin_attempts')
            ->where('business_id', $businessId)
            ->where('pin_key', $this->pinKey($staffId))
            ->sum('attempts');
        $lockedUntil = $totalAttempts >= (int) $settings->pin_lockout_max_attempts ? now()->addMinutes((int) $settings->pin_lockout_decay_minutes) : null;

        DB::table('pin_attempts')
            ->where('business_id', $businessId)
            ->where('pin_key', $this->pinKey($staffId))
            ->update([
                'attempts' => $totalAttempts,
                'locked_until' => $lockedUntil,
                'updated_at' => now(),
            ]);

        return DB::table('pin_attempts')
            ->where('business_id', $businessId)
            ->where('pin_key', $this->pinKey($staffId))
            ->orderByDesc('attempts')
            ->first();
    }

    private function resetPinAttempt(string $businessId, string $staffId, ?object $attempt): void
    {
        DB::table('pin_attempts')->updateOrInsert(
            [
                'business_id' => $businessId,
                'pin_key' => $this->pinKey($staffId),
            ],
            [
                'id' => $attempt->id ?? (string) Str::uuid(),
                'outlet_id' => null,
                'device_id' => null,
                'attempts' => 0,
                'locked_until' => null,
                'created_at' => $attempt->created_at ?? now(),
                'updated_at' => now(),
            ],
        );
    }

    private function pinKey(string $staffId): string
    {
        return 'terminal_unlock:'.$staffId;
    }

    private function settings(string $businessId): object
    {
        return BusinessSecuritySetting::query()->firstOrCreate(
            ['business_id' => $businessId],
            [
                'id' => (string) Str::uuid(),
                'pin_lockout_max_attempts' => 5,
                'pin_lockout_decay_minutes' => 15,
                'idle_lock_timeout_seconds' => 180,
                'terminal_session_timeout_seconds' => 900,
            ],
        );
    }

    private function payload(string $businessId, object $state): array
    {
        $settings = $this->settings($businessId);

        return [
            'locked' => (bool) $state->locked,
            'locked_at' => $this->iso($state->locked_at),
            'lock_reason' => $state->lock_reason,
            'outlet_id' => $state->outlet_id,
            'device_id' => $state->device_id,
            'cashier' => $this->userPayloadById($state->current_cashier_id),
            'shift' => $state->shift_id ? ['id' => $state->shift_id, 'preserved' => true] : null,
            'failed_attempts_remaining' => (int) $settings->pin_lockout_max_attempts,
            'lockout_until' => null,
            'idle_timeout_seconds' => (int) $settings->idle_lock_timeout_seconds,
            'session_timeout_seconds' => (int) $settings->terminal_session_timeout_seconds,
            'unlocked_at' => $this->iso($state->unlocked_at),
            'unlocked_by' => $this->userPayloadById($state->unlocked_by_user_id),
            'server_time' => now()->toISOString(),
        ];
    }

    private function stateSnapshot(object $state): array
    {
        return [
            'id' => $state->id,
            'business_id' => $state->business_id,
            'outlet_id' => $state->outlet_id,
            'device_id' => $state->device_id,
            'current_user_id' => $state->current_user_id,
            'current_cashier_id' => $state->current_cashier_id,
            'shift_id' => $state->shift_id,
            'locked' => (bool) $state->locked,
            'locked_at' => $this->iso($state->locked_at),
            'lock_reason' => $state->lock_reason,
            'unlocked_at' => $this->iso($state->unlocked_at),
            'unlocked_by_user_id' => $state->unlocked_by_user_id,
        ];
    }

    private function userPayloadById(?string $id): ?array
    {
        if (! $id) {
            return null;
        }

        $user = DB::table('users')->where('id', $id)->first();

        return $user ? $this->userPayload($user) : null;
    }

    private function userPayload(object $user): array
    {
        return [
            'id' => $user->id,
            'name' => $user->name,
            'role' => $user->role,
        ];
    }

    private function iso(mixed $value): ?string
    {
        return $value ? CarbonImmutable::parse($value)->toISOString() : null;
    }
}
