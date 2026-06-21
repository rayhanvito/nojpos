<?php

namespace App\Services;

use App\Support\Nojpos;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;

class StoreStateService
{
    /**
     * @return array<string, mixed>
     */
    public function state(string $businessId, string $outletId): array
    {
        $outlet = $this->outlet($businessId, $outletId);
        $state = $this->stateRow($businessId, $outletId);

        return $this->payload($outlet, $state, $this->blockingShifts($businessId, $outletId));
    }

    /**
     * @param  array{authorization_pin?: string|null, reason?: string|null, override_out_of_hours?: bool|null}  $data
     * @return array<string, mixed>
     */
    public function open(string $businessId, string $outletId, array $data, string $actorId, string $actorRole): array
    {
        return DB::transaction(function () use ($businessId, $outletId, $data, $actorId, $actorRole): array {
            $outlet = $this->outlet($businessId, $outletId, lock: true);
            $this->assertEnabled($outlet);
            $this->assertAuthorizationPin($businessId, $actorId, $actorRole, $data['authorization_pin'] ?? null);

            $state = $this->stateRow($businessId, $outletId, lock: true) ?? $this->createDefaultState($businessId, $outletId);
            if ($state->status === 'open') {
                throw new StoreOperationException('STORE_ALREADY_OPEN', 'Store is already open.', [], 409);
            }

            $before = $this->stateSnapshot($state);
            $now = now();
            DB::table('outlet_store_states')
                ->where('id', $state->id)
                ->update([
                    'status' => 'open',
                    'opened_at' => $now,
                    'opened_by_user_id' => $actorId,
                    'closed_at' => null,
                    'closed_by_user_id' => null,
                    'reason' => $data['reason'] ?? null,
                    'updated_at' => $now,
                ]);

            $fresh = $this->stateRow($businessId, $outletId, lock: true);
            Nojpos::audit($businessId, $actorId, 'store.open', 'outlet_store_state', $fresh->id, $before, $this->stateSnapshot($fresh));

            return $this->payload($outlet, $fresh, []);
        });
    }

    /**
     * @param  array{authorization_pin?: string|null, reason?: string|null, force?: bool|null}  $data
     * @return array<string, mixed>
     */
    public function close(string $businessId, string $outletId, array $data, string $actorId, string $actorRole): array
    {
        return DB::transaction(function () use ($businessId, $outletId, $data, $actorId, $actorRole): array {
            $outlet = $this->outlet($businessId, $outletId, lock: true);
            $this->assertEnabled($outlet);
            $this->assertAuthorizationPin($businessId, $actorId, $actorRole, $data['authorization_pin'] ?? null);

            if ((bool) ($data['force'] ?? false)) {
                throw new StoreOperationException('STORE_FORCE_CLOSE_UNSUPPORTED', 'Force close is not supported.', [], 422);
            }

            $state = $this->stateRow($businessId, $outletId, lock: true) ?? $this->createDefaultState($businessId, $outletId);
            if ($state->status === 'closed') {
                throw new StoreOperationException('STORE_ALREADY_CLOSED', 'Store is already closed.', [], 409);
            }

            $blockingShifts = $this->blockingShifts($businessId, $outletId);
            if ($blockingShifts !== []) {
                throw new StoreOperationException('STORE_CLOSE_BLOCKED_OPEN_SHIFTS', 'Store cannot close while shifts are open or pending close.', ['blocking_shifts' => $blockingShifts, 'blocking_shifts_count' => count($blockingShifts)], 409);
            }

            $before = $this->stateSnapshot($state);
            $now = now();
            DB::table('outlet_store_states')
                ->where('id', $state->id)
                ->update([
                    'status' => 'closed',
                    'closed_at' => $now,
                    'closed_by_user_id' => $actorId,
                    'reason' => $data['reason'] ?? null,
                    'updated_at' => $now,
                ]);

            $fresh = $this->stateRow($businessId, $outletId, lock: true);
            Nojpos::audit($businessId, $actorId, 'store.close', 'outlet_store_state', $fresh->id, $before, $this->stateSnapshot($fresh));

            return $this->payload($outlet, $fresh, []);
        });
    }

    public function assertStoreAllowsShiftOpen(string $businessId, string $outletId): void
    {
        $outlet = $this->outlet($businessId, $outletId);
        if (! (bool) ($outlet->store_open_close_enabled ?? false)) {
            return;
        }

        $state = $this->stateRow($businessId, $outletId);
        if (($state?->status ?? 'open') === 'closed') {
            throw new StoreOperationException('STORE_CLOSED_SHIFT_OPEN_BLOCKED', 'Store is closed. Open the store before opening a shift.', ['store_state' => $this->payload($outlet, $state, [])], 409);
        }
    }

    public function assertStoreAllowsCheckout(string $businessId, string $outletId): void
    {
        $outlet = $this->outlet($businessId, $outletId);
        if (! (bool) ($outlet->store_open_close_enabled ?? false)) {
            return;
        }

        $state = $this->stateRow($businessId, $outletId);
        if (($state?->status ?? 'open') === 'closed') {
            throw new StoreOperationException('STORE_CLOSED_CHECKOUT_BLOCKED', 'Store is closed. Open the store before checkout.', ['store_state' => $this->payload($outlet, $state, [])], 409);
        }
    }

    private function outlet(string $businessId, string $outletId, bool $lock = false): object
    {
        $query = DB::table('outlets')
            ->where('business_id', $businessId)
            ->where('id', $outletId)
            ->whereNull('deleted_at');

        if ($lock) {
            $query->lockForUpdate();
        }

        $outlet = $query->first();
        if (! $outlet) {
            throw new StoreOperationException('NOT_FOUND', 'Outlet not found.', [], 404);
        }

        return $outlet;
    }

    private function stateRow(string $businessId, string $outletId, bool $lock = false): ?object
    {
        $query = DB::table('outlet_store_states')
            ->where('business_id', $businessId)
            ->where('outlet_id', $outletId);

        if ($lock) {
            $query->lockForUpdate();
        }

        return $query->first();
    }

    private function createDefaultState(string $businessId, string $outletId): object
    {
        $id = (string) Str::uuid();
        $now = now();
        DB::table('outlet_store_states')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'status' => 'open',
            'opened_at' => $now,
            'created_at' => $now,
            'updated_at' => $now,
        ]);

        return $this->stateRow($businessId, $outletId) ?? throw new StoreOperationException('STORE_STATE_NOT_FOUND', 'Store state could not be initialized.', [], 409);
    }

    private function assertEnabled(object $outlet): void
    {
        if (! (bool) ($outlet->store_open_close_enabled ?? false)) {
            throw new StoreOperationException('STORE_OPEN_CLOSE_DISABLED', 'Store open/close is disabled for this outlet.', [], 403);
        }
    }

    private function assertAuthorizationPin(string $businessId, string $actorId, string $actorRole, ?string $pin): void
    {
        if (! $this->requiresPin($businessId)) {
            return;
        }

        if ($pin === null || trim($pin) === '') {
            throw new StoreOperationException('PIN_REQUIRED', 'Authorization PIN is required.', [], 422);
        }

        $actor = DB::table('users')
            ->where('business_id', $businessId)
            ->where('id', $actorId)
            ->whereIn('role', ['owner', 'admin'])
            ->whereNull('deleted_at')
            ->first();

        if (! $actor || ! $actor->pin_hash || ! Hash::check($pin, $actor->pin_hash)) {
            throw new StoreOperationException('INVALID_PIN', 'PIN is invalid.', [], 422);
        }
    }

    private function requiresPin(string $businessId): bool
    {
        $setting = DB::table('business_security_settings')
            ->where('business_id', $businessId)
            ->first();

        return $setting ? (bool) $setting->pin_required_store_open_close : true;
    }

    /**
     * @return list<array<string, mixed>>
     */
    private function blockingShifts(string $businessId, string $outletId): array
    {
        return DB::table('shift_sessions')
            ->leftJoin('users', 'users.id', '=', 'shift_sessions.cashier_id')
            ->leftJoin('devices', 'devices.id', '=', 'shift_sessions.device_id')
            ->where('shift_sessions.business_id', $businessId)
            ->where('shift_sessions.outlet_id', $outletId)
            ->whereIn('shift_sessions.status', ['open', 'pending_close'])
            ->whereNull('shift_sessions.deleted_at')
            ->orderBy('shift_sessions.opened_at')
            ->select([
                'shift_sessions.id',
                'shift_sessions.status',
                'shift_sessions.cashier_id',
                'users.name as cashier_name',
                'shift_sessions.device_id',
                'devices.name as device_name',
                'shift_sessions.opened_at',
            ])
            ->get()
            ->map(fn (object $shift): array => [
                'shift_id' => $shift->id,
                'shift_number' => $shift->id,
                'status' => $shift->status,
                'cashier' => ['id' => $shift->cashier_id, 'name' => $shift->cashier_name],
                'device' => ['id' => $shift->device_id, 'name' => $shift->device_name],
                'opened_at' => $shift->opened_at,
            ])
            ->values()
            ->all();
    }

    /**
     * @return array<string, mixed>
     */
    private function payload(object $outlet, ?object $state, array $blockingShifts): array
    {
        $status = $state?->status ?? 'open';

        return [
            'outlet_id' => $outlet->id,
            'outlet_name' => $outlet->name,
            'status' => $status,
            'closed_at' => $state?->closed_at,
            'opened_at' => $state?->opened_at,
            'closed_by' => $this->userPayload($state?->closed_by_user_id ?? null),
            'opened_by' => $this->userPayload($state?->opened_by_user_id ?? null),
            'reason' => $state?->reason,
            'blocking_shifts_count' => count($blockingShifts),
            'blocking_shifts' => $blockingShifts,
            'settings' => [
                'store_open_close_enabled' => (bool) ($outlet->store_open_close_enabled ?? false),
                'pin_required_store_open_close' => $this->requiresPin($outlet->business_id),
            ],
            'server_time' => now()->toIso8601String(),
        ];
    }

    private function userPayload(?string $userId): ?array
    {
        if (! $userId) {
            return null;
        }

        $user = DB::table('users')->where('id', $userId)->first();
        if (! $user) {
            return null;
        }

        return ['id' => $user->id, 'name' => $user->name, 'role' => $user->role];
    }

    /**
     * @return array<string, mixed>
     */
    private function stateSnapshot(object $state): array
    {
        return [
            'id' => $state->id,
            'business_id' => $state->business_id,
            'outlet_id' => $state->outlet_id,
            'status' => $state->status,
            'opened_at' => $this->iso($state->opened_at),
            'opened_by_user_id' => $state->opened_by_user_id,
            'closed_at' => $this->iso($state->closed_at),
            'closed_by_user_id' => $state->closed_by_user_id,
            'reason' => $state->reason,
        ];
    }

    private function iso(mixed $timestamp): ?string
    {
        return $timestamp ? Carbon::parse($timestamp)->toIso8601String() : null;
    }
}
