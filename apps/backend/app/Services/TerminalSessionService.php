<?php

namespace App\Services;

use App\Models\BusinessSecuritySetting;
use App\Models\User;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class TerminalSessionService
{
    public const STATUS_ACTIVE = 'active';
    public const STATUS_LOCKED = 'locked';
    public const STATUS_EXPIRED = 'expired';
    public const STATUS_REVOKED = 'revoked';

    public function __construct(private readonly AuditLogService $audit) {}

    public function startFromPinSwitch(User $actor, User $cashier, Request $request, string $outletId, string $deviceId): array
    {
        return DB::transaction(function () use ($actor, $cashier, $request, $outletId, $deviceId): array {
            $businessId = (string) $actor->business_id;
            $this->assertOutletAndDevice($businessId, $outletId, $deviceId);

            if ($cashier->business_id !== $businessId || $cashier->role !== 'cashier') {
                throw new TerminalSessionException('TERMINAL_CONTEXT_MISMATCH', 'Terminal actor, device, outlet, and shift context do not match.');
            }

            $tokenId = $this->tokenId($request);
            $now = now();
            $expiresAt = $now->copy()->addSeconds($this->sessionTimeoutSeconds($businessId));

            $this->revokeExistingBoundSessions($businessId, $actor->id, $tokenId, $outletId, $deviceId, $now);

            $id = (string) Str::uuid();
            DB::table('terminal_sessions')->insert([
                'id' => $id,
                'business_id' => $businessId,
                'outlet_id' => $outletId,
                'device_id' => $deviceId,
                'user_id' => $actor->id,
                'cashier_id' => $cashier->id,
                'token_id' => $tokenId,
                'status' => self::STATUS_ACTIVE,
                'opened_at' => $now,
                'last_seen_at' => $now,
                'expires_at' => $expiresAt,
                'metadata' => json_encode([
                    'source' => 'pin_switch',
                    'actor_role' => $actor->role,
                ], JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE),
                'created_at' => $now,
                'updated_at' => $now,
            ]);

            $session = DB::table('terminal_sessions')->where('id', $id)->first();

            $this->audit->record(
                $businessId,
                $actor->id,
                'terminal_session.started',
                'terminal_session',
                $id,
                null,
                $this->snapshot($session),
                [
                    'outlet_id' => $outletId,
                    'device_id' => $deviceId,
                    'request_id' => $request->headers->get('X-Request-Id'),
                    'idempotency_key' => $request->headers->get('Idempotency-Key'),
                ],
            );

            return $this->payload($session);
        });
    }

    public function activeSessionForContext(Request $request, string $businessId, string $outletId, string $deviceId, string $cashierId): ?object
    {
        $bound = $this->boundSession($request, $businessId);

        if (! $bound) {
            return null;
        }

        if ($bound->outlet_id !== $outletId || $bound->device_id !== $deviceId) {
            throw new TerminalSessionException('TERMINAL_CONTEXT_MISMATCH', 'Terminal actor, device, outlet, and shift context do not match.');
        }

        if ($bound->cashier_id !== $cashierId) {
            throw new TerminalSessionException('TERMINAL_CONTEXT_MISMATCH', 'Terminal actor, device, outlet, and shift context do not match.');
        }

        $this->assertUsable($bound);
        $this->touch($bound);

        return $bound;
    }

    public function lock(User $actor, string $outletId, string $deviceId): void
    {
        DB::table('terminal_sessions')
            ->where('business_id', $actor->business_id)
            ->where('outlet_id', $outletId)
            ->where('device_id', $deviceId)
            ->where(function ($query) use ($actor): void {
                $query->where('user_id', $actor->id)
                    ->orWhere('cashier_id', $actor->id);
            })
            ->where('status', self::STATUS_ACTIVE)
            ->update([
                'status' => self::STATUS_LOCKED,
                'locked_at' => now(),
                'updated_at' => now(),
            ]);
    }

    public function unlock(User $actor, string $outletId, string $deviceId, string $staffId): void
    {
        $now = now();
        DB::table('terminal_sessions')
            ->where('business_id', $actor->business_id)
            ->where('outlet_id', $outletId)
            ->where('device_id', $deviceId)
            ->where('cashier_id', $staffId)
            ->where('status', self::STATUS_LOCKED)
            ->update([
                'status' => self::STATUS_ACTIVE,
                'locked_at' => null,
                'last_seen_at' => $now,
                'expires_at' => $now->copy()->addSeconds($this->sessionTimeoutSeconds((string) $actor->business_id)),
                'updated_at' => $now,
            ]);
    }

    public function revoke(string $businessId, string $sessionId, string $actorId, ?Request $request = null): void
    {
        DB::transaction(function () use ($businessId, $sessionId, $actorId, $request): void {
            $session = DB::table('terminal_sessions')
                ->where('business_id', $businessId)
                ->where('id', $sessionId)
                ->lockForUpdate()
                ->first();

            if (! $session || $session->status === self::STATUS_REVOKED) {
                return;
            }

            $before = $this->snapshot($session);
            DB::table('terminal_sessions')->where('id', $sessionId)->update([
                'status' => self::STATUS_REVOKED,
                'revoked_at' => now(),
                'updated_at' => now(),
            ]);
            $fresh = DB::table('terminal_sessions')->where('id', $sessionId)->first();

            $this->audit->record(
                $businessId,
                $actorId,
                'terminal_session.revoked',
                'terminal_session',
                $sessionId,
                $before,
                $this->snapshot($fresh),
                [
                    'outlet_id' => $fresh->outlet_id,
                    'device_id' => $fresh->device_id,
                    'request_id' => $request?->headers->get('X-Request-Id'),
                ],
            );
        });
    }

    private function boundSession(Request $request, string $businessId): ?object
    {
        $tokenId = $this->tokenId($request);
        $userId = $request->user()?->id;

        if (! $userId) {
            return null;
        }

        $query = DB::table('terminal_sessions')
            ->where('business_id', $businessId)
            ->where(function ($query) use ($tokenId, $userId): void {
                if ($tokenId) {
                    $query->where('token_id', $tokenId)
                        ->orWhere(function ($query) use ($userId): void {
                            $query->whereNull('token_id')->where('user_id', $userId);
                        });
                } else {
                    $query->where('user_id', $userId);
                }
            })
            ->orderByRaw("case status when 'active' then 0 when 'locked' then 1 when 'expired' then 2 when 'revoked' then 3 else 4 end")
            ->orderByDesc('opened_at')
            ->orderByDesc('created_at');

        return $query->first();
    }

    private function assertUsable(object $session): void
    {
        if ($session->expires_at && now()->greaterThan($session->expires_at)) {
            DB::table('terminal_sessions')
                ->where('id', $session->id)
                ->where('status', self::STATUS_ACTIVE)
                ->update([
                    'status' => self::STATUS_EXPIRED,
                    'updated_at' => now(),
                ]);

            throw new TerminalSessionException('TERMINAL_SESSION_EXPIRED', 'Terminal session has expired. Please switch cashier PIN again.', [], 403);
        }

        if ($session->status === self::STATUS_LOCKED) {
            throw new TerminalSessionException('TERMINAL_SESSION_LOCKED', 'Terminal is locked. Please unlock before continuing.', [], 423);
        }

        if ($session->status === self::STATUS_REVOKED || $session->status === self::STATUS_EXPIRED) {
            throw new TerminalSessionException('TERMINAL_SESSION_INVALID', 'Terminal session is no longer valid. Please switch cashier PIN again.', [], 403);
        }

        if ($session->status !== self::STATUS_ACTIVE) {
            throw new TerminalSessionException('TERMINAL_SESSION_INVALID', 'Terminal session is no longer valid. Please switch cashier PIN again.', [], 403);
        }
    }

    private function touch(object $session): void
    {
        DB::table('terminal_sessions')->where('id', $session->id)->update([
            'last_seen_at' => now(),
            'updated_at' => now(),
        ]);
    }

    private function revokeExistingBoundSessions(string $businessId, string $userId, ?string $tokenId, string $outletId, string $deviceId, mixed $now): void
    {
        DB::table('terminal_sessions')
            ->where('business_id', $businessId)
            ->where(function ($query) use ($userId, $tokenId): void {
                if ($tokenId) {
                    $query->where('token_id', $tokenId)
                        ->orWhere(function ($query) use ($userId): void {
                            $query->whereNull('token_id')->where('user_id', $userId);
                        });
                } else {
                    $query->where('user_id', $userId);
                }
            })
            ->whereIn('status', [self::STATUS_ACTIVE, self::STATUS_LOCKED])
            ->update([
                'status' => self::STATUS_REVOKED,
                'revoked_at' => $now,
                'updated_at' => $now,
            ]);
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
            throw new TerminalSessionException('FORBIDDEN', 'Terminal context is outside the current business scope.', [], 403);
        }
    }

    private function sessionTimeoutSeconds(string $businessId): int
    {
        $settings = BusinessSecuritySetting::query()->firstOrCreate(
            ['business_id' => $businessId],
            [
                'id' => (string) Str::uuid(),
                'terminal_session_timeout_seconds' => 900,
                'idle_lock_timeout_seconds' => 180,
                'pin_lockout_max_attempts' => 5,
                'pin_lockout_decay_minutes' => 15,
            ],
        );

        return max(60, (int) $settings->terminal_session_timeout_seconds);
    }

    private function tokenId(Request $request): ?string
    {
        $token = $request->user()?->currentAccessToken();

        return is_object($token) && isset($token->id) ? (string) $token->id : null;
    }

    private function payload(object $session): array
    {
        return [
            'id' => $session->id,
            'status' => $session->status,
            'outlet_id' => $session->outlet_id,
            'device_id' => $session->device_id,
            'cashier_id' => $session->cashier_id,
            'opened_at' => $this->iso($session->opened_at),
            'last_seen_at' => $this->iso($session->last_seen_at),
            'expires_at' => $this->iso($session->expires_at),
        ];
    }

    private function snapshot(object $session): array
    {
        return [
            'id' => $session->id,
            'business_id' => $session->business_id,
            'outlet_id' => $session->outlet_id,
            'device_id' => $session->device_id,
            'user_id' => $session->user_id,
            'cashier_id' => $session->cashier_id,
            'token_id' => $session->token_id,
            'status' => $session->status,
            'opened_at' => $this->iso($session->opened_at),
            'last_seen_at' => $this->iso($session->last_seen_at),
            'expires_at' => $this->iso($session->expires_at),
            'locked_at' => $this->iso($session->locked_at),
            'revoked_at' => $this->iso($session->revoked_at),
        ];
    }

    private function iso(mixed $value): ?string
    {
        return $value ? \Carbon\CarbonImmutable::parse($value)->toISOString() : null;
    }
}
