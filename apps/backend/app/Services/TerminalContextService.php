<?php

namespace App\Services;

use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class TerminalContextService
{
    public function __construct(private readonly TerminalSessionService $terminalSessions) {}

    public function fromPayload(Request $request, array $data, bool $requireOpenShift = false): TerminalContext
    {
        foreach (['outlet_id', 'device_id', 'cashier_id'] as $field) {
            if (empty($data[$field])) {
                $this->fail('TERMINAL_CONTEXT_REQUIRED', 'Outlet, device, and cashier context are required.', 422, [$field => ['The '.$field.' field is required.']]);
            }
        }

        $businessId = $this->businessId($request);
        $this->assertOutlet($businessId, $data['outlet_id']);
        $this->assertDevice($businessId, $data['outlet_id'], $data['device_id']);
        $this->assertCashier($request, $businessId, $data['cashier_id']);

        if (! empty($data['shift_id'])) {
            $this->assertShiftTuple($businessId, $data['outlet_id'], $data['device_id'], $data['cashier_id'], $data['shift_id'], $requireOpenShift);
        }

        return new TerminalContext($businessId, $data['outlet_id'], $data['device_id'], $data['cashier_id'], $data['shift_id'] ?? null);
    }

    public function assertShiftForActor(Request $request, object $shift, bool $requireOpenShift = true): void
    {
        $businessId = $this->businessId($request);

        if (($shift->business_id ?? null) !== $businessId) {
            $this->failMismatch();
        }

        $this->assertOutlet($businessId, $shift->outlet_id);
        $this->assertDevice($businessId, $shift->outlet_id, $shift->device_id);
        if (! $this->hasValidTerminalSession($request, $businessId, $shift->outlet_id, $shift->device_id, $shift->cashier_id)) {
            $this->assertPrivilegedOrSelf($request, $shift->cashier_id);
        }

        if ($requireOpenShift && ($shift->status ?? null) !== 'open') {
            $this->fail('SHIFT_NOT_OPEN', 'Shift must be open for this terminal action.', 422);
        }
    }

    public function assertTransactionContext(Request $request, object $transaction, bool $requireOpenShift = true, bool $allowPrivilegedActor = true): void
    {
        $businessId = $this->businessId($request);

        if (($transaction->business_id ?? null) !== $businessId) {
            $this->failMismatch();
        }

        $this->assertOutlet($businessId, $transaction->outlet_id);
        $this->assertDevice($businessId, $transaction->outlet_id, $transaction->device_id);

        if (! $this->hasValidTerminalSession($request, $businessId, $transaction->outlet_id, $transaction->device_id, $transaction->cashier_id)) {
            if ($allowPrivilegedActor) {
                $this->assertPrivilegedOrSelf($request, $transaction->cashier_id);
            } else {
                $this->assertCashier($request, $businessId, $transaction->cashier_id);
            }
        }

        $this->assertShiftTuple(
            $businessId,
            $transaction->outlet_id,
            $transaction->device_id,
            $transaction->cashier_id,
            $transaction->shift_id,
            $requireOpenShift,
        );
    }

    private function businessId(Request $request): string
    {
        $businessId = $request->user()?->business_id;

        if (! $businessId) {
            $this->fail('UNAUTHENTICATED', 'Unauthenticated.', 401);
        }

        return $businessId;
    }

    private function assertOutlet(string $businessId, string $outletId): void
    {
        $exists = DB::table('outlets')
            ->where('id', $outletId)
            ->where('business_id', $businessId)
            ->whereNull('deleted_at')
            ->exists();

        if (! $exists) {
            $this->fail('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }
    }

    private function assertDevice(string $businessId, string $outletId, string $deviceId): void
    {
        $device = DB::table('devices')
            ->where('id', $deviceId)
            ->where('business_id', $businessId)
            ->whereNull('deleted_at')
            ->first();

        if (! $device) {
            $this->fail('DEVICE_NOT_ENROLLED', 'Device is not enrolled for this business.', [], status: 403);
        }

        if ($device->outlet_id !== $outletId) {
            $this->failMismatch();
        }
    }

    private function assertCashier(Request $request, string $businessId, string $cashierId): void
    {
        $exists = DB::table('users')
            ->where('id', $cashierId)
            ->where('business_id', $businessId)
            ->whereNull('deleted_at')
            ->exists();

        if (! $exists) {
            $this->failMismatch();
        }

        if ($this->hasValidTerminalSession($request, $businessId, $this->contextOutletId($request), $this->contextDeviceId($request), $cashierId)) {
            return;
        }

        if ($request->user()->id === $cashierId) {
            return;
        }

        if ($this->hasExactTokenAbility($request, 'acting_cashier:'.$cashierId)) {
            return;
        }

        $this->failMismatch();
    }

    private function assertPrivilegedOrSelf(Request $request, string $cashierId): void
    {
        if ($this->hasValidTerminalSession($request, $this->businessId($request), $this->contextOutletId($request), $this->contextDeviceId($request), $cashierId)) {
            return;
        }

        if ($request->user()->id === $cashierId) {
            return;
        }

        if ($this->hasExactTokenAbility($request, 'acting_cashier:'.$cashierId)) {
            return;
        }

        if (in_array($request->user()->role, ['owner', 'admin', 'supervisor'], true)) {
            return;
        }

        $this->failMismatch();
    }

    private function hasExactTokenAbility(Request $request, string $ability): bool
    {
        $token = $request->user()?->currentAccessToken();
        $abilities = is_object($token) ? ($token->abilities ?? []) : [];

        return in_array($ability, is_array($abilities) ? $abilities : [], true);
    }

    private function hasValidTerminalSession(Request $request, string $businessId, ?string $outletId, ?string $deviceId, string $cashierId): bool
    {
        if (! $outletId || ! $deviceId) {
            return false;
        }

        try {
            return $this->terminalSessions->activeSessionForContext($request, $businessId, $outletId, $deviceId, $cashierId) !== null;
        } catch (TerminalSessionException $error) {
            throw new TerminalContextException($error->errorCode, $error->getMessage(), $error->status, $error->details);
        }
    }

    private function contextOutletId(Request $request): ?string
    {
        return $request->input('outlet_id')
            ?? $request->route('outlet')
            ?? $request->query('outlet_id');
    }

    private function contextDeviceId(Request $request): ?string
    {
        return $request->input('device_id')
            ?? $request->query('device_id');
    }

    private function assertShiftTuple(string $businessId, string $outletId, string $deviceId, string $cashierId, string $shiftId, bool $requireOpenShift): void
    {
        $shift = DB::table('shift_sessions')
            ->where('id', $shiftId)
            ->where('business_id', $businessId)
            ->whereNull('deleted_at')
            ->first();

        if (! $shift) {
            $this->failMismatch();
        }

        if ($shift->outlet_id !== $outletId || $shift->device_id !== $deviceId || $shift->cashier_id !== $cashierId) {
            $this->failMismatch();
        }

        if ($requireOpenShift && $shift->status !== 'open') {
            $this->fail('SHIFT_NOT_OPEN', 'Shift must be open for this terminal action.', 422);
        }
    }

    private function failMismatch(): void
    {
        $this->fail('TERMINAL_CONTEXT_MISMATCH', 'Terminal actor, device, outlet, and shift context do not match.', [], 403);
    }

    private function fail(string $code, string $message, array $details = [], int $status = 403): void
    {
        throw new TerminalContextException($code, $message, $status, $details);
    }
}
