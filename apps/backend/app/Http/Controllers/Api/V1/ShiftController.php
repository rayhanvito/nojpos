<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\ShiftSession;
use App\Support\ApiResponse;
use App\Support\Nojpos;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class ShiftController extends Controller
{
    public function open(Request $request): JsonResponse
    {
        $data = $request->validate([
            'outlet_id' => ['required', 'uuid'],
            'device_id' => ['required', 'uuid'],
            'cashier_id' => ['required', 'uuid'],
            'opening_cash' => ['required', 'integer', 'min:0'],
        ]);

        $businessId = $request->user()->business_id;

        if (! $this->allBelongToBusiness($businessId, [
            'outlets' => $data['outlet_id'],
            'devices' => $data['device_id'],
            'users' => $data['cashier_id'],
        ])) {
            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        $existing = ShiftSession::query()
            ->where('outlet_id', $data['outlet_id'])
            ->where('device_id', $data['device_id'])
            ->where('status', 'open')
            ->first();

        if ($existing) {
            return ApiResponse::error('SHIFT_ALREADY_OPEN', 'A shift is already open for this device and outlet.', [], 422);
        }

        $shift = ShiftSession::query()->create([
            'business_id' => $businessId,
            'outlet_id' => $data['outlet_id'],
            'device_id' => $data['device_id'],
            'cashier_id' => $data['cashier_id'],
            'status' => 'open',
            'opening_cash' => $data['opening_cash'],
            'opened_at' => now(),
        ]);

        Nojpos::audit($businessId, $request->user()->id, 'shift.open', 'shift_session', $shift->id);

        return ApiResponse::success($this->shiftPayload($shift), [], 201);
    }

    public function current(Request $request): JsonResponse
    {
        $shift = ShiftSession::query()
            ->when($request->query('outlet_id'), fn ($query, $outlet) => $query->where('outlet_id', $outlet))
            ->when($request->query('device_id'), fn ($query, $device) => $query->where('device_id', $device))
            ->where('status', 'open')
            ->latest()
            ->first();

        return ApiResponse::success(['shift' => $shift ? $this->shiftPayload($shift) : null]);
    }

    public function cashMovement(Request $request, string $shift): JsonResponse
    {
        $data = $request->validate([
            'type' => ['required', 'in:cash_in,cash_out'],
            'amount' => ['required', 'integer', 'min:1'],
            'reason' => ['nullable', 'string'],
        ]);

        $session = ShiftSession::query()->where('id', $shift)->firstOrFail();
        if ($session->status !== 'open') {
            return ApiResponse::error('SHIFT_ALREADY_CLOSED', 'Shift is already closed.', [], 422);
        }

        $id = (string) Str::uuid();
        DB::table('cash_movements')->insert([
            'id' => $id,
            'business_id' => $session->business_id,
            'outlet_id' => $session->outlet_id,
            'shift_id' => $session->id,
            'actor_id' => $request->user()->id,
            'type' => $data['type'],
            'amount' => $data['amount'],
            'reason' => $data['reason'] ?? null,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return ApiResponse::success([
            'id' => $id,
            'type' => $data['type'],
            'amount' => $data['amount'],
            'reason' => $data['reason'] ?? null,
        ], [], 201);
    }

    public function close(Request $request, string $shift): JsonResponse
    {
        $data = $request->validate([
            'actual_cash' => ['required', 'integer', 'min:0'],
        ]);

        $session = ShiftSession::query()->where('id', $shift)->firstOrFail();
        if ($session->status !== 'open') {
            return ApiResponse::error('SHIFT_ALREADY_CLOSED', 'Shift is already closed.', [], 422);
        }

        $cashSales = (int) DB::table('payments')
            ->join('transactions', 'transactions.id', '=', 'payments.transaction_id')
            ->where('transactions.shift_id', $session->id)
            ->where('payments.business_id', $session->business_id)
            ->where('payments.is_cash', true)
            ->where('payments.status', 'confirmed')
            ->sum('payments.amount');

        $cashIn = (int) DB::table('cash_movements')
            ->where('shift_id', $session->id)
            ->where('type', 'cash_in')
            ->sum('amount');

        $cashOut = (int) DB::table('cash_movements')
            ->where('shift_id', $session->id)
            ->where('type', 'cash_out')
            ->sum('amount');

        $expected = (int) $session->opening_cash + $cashSales + $cashIn - $cashOut;
        $difference = $data['actual_cash'] - $expected;

        $session->update([
            'status' => 'closed',
            'expected_cash' => $expected,
            'actual_cash' => $data['actual_cash'],
            'cash_difference' => $difference,
            'closed_at' => now(),
        ]);

        Nojpos::audit($session->business_id, $request->user()->id, 'shift.close', 'shift_session', $session->id);

        return ApiResponse::success($this->shiftPayload($session->refresh()));
    }

    private function allBelongToBusiness(string $businessId, array $tableToIds): bool
    {
        foreach ($tableToIds as $table => $id) {
            if (! DB::table($table)->where('id', $id)->where('business_id', $businessId)->exists()) {
                return false;
            }
        }

        return true;
    }

    private function shiftPayload(ShiftSession $shift): array
    {
        return [
            'id' => $shift->id,
            'business_id' => $shift->business_id,
            'outlet_id' => $shift->outlet_id,
            'device_id' => $shift->device_id,
            'cashier_id' => $shift->cashier_id,
            'status' => $shift->status,
            'opening_cash' => (int) $shift->opening_cash,
            'expected_cash' => $shift->expected_cash === null ? null : (int) $shift->expected_cash,
            'actual_cash' => $shift->actual_cash === null ? null : (int) $shift->actual_cash,
            'cash_difference' => $shift->cash_difference === null ? null : (int) $shift->cash_difference,
        ];
    }
}
