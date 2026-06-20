<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\ShiftSession;
use App\Services\ShiftService;
use App\Services\ShiftStateException;
use App\Services\TerminalContextException;
use App\Services\TerminalContextService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class ShiftController extends Controller
{
    public function __construct(
        private readonly TerminalContextService $terminalContext,
        private readonly ShiftService $shifts,
    ) {}

    public function open(Request $request): JsonResponse
    {
        $data = $request->validate([
            'outlet_id' => ['required', 'uuid'],
            'device_id' => ['required', 'uuid'],
            'cashier_id' => ['required', 'uuid'],
            'opening_cash' => ['required', 'integer', 'min:0'],
        ]);

        $businessId = $request->user()->business_id;

        try {
            $this->terminalContext->fromPayload($request, $data);
        } catch (TerminalContextException $error) {
            return $this->terminalContextError($error);
        }

        try {
            $shift = $this->shifts->open($businessId, $data, $request->user()->id);
        } catch (ShiftStateException $error) {
            return $this->shiftStateError($error);
        }

        return ApiResponse::success($this->shiftPayload($shift), [], 201);
    }

    public function current(Request $request): JsonResponse
    {
        $shift = ShiftSession::query()
            ->where('business_id', $request->user()->business_id)
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
            'reason' => ['required', 'string', 'min:1'],
        ]);

        $session = ShiftSession::query()
            ->where('business_id', $request->user()->business_id)
            ->where('id', $shift)
            ->firstOrFail();

        try {
            $this->terminalContext->assertShiftForActor($request, $session, requireOpenShift: false);
        } catch (TerminalContextException $error) {
            return $this->terminalContextError($error);
        }

        try {
            $movement = $this->shifts->cashMovement($request->user()->business_id, $session->id, $data, $request->user()->id);
        } catch (ShiftStateException $error) {
            return $this->shiftStateError($error);
        }

        return ApiResponse::success($movement, [], 201);
    }

    public function close(Request $request, string $shift): JsonResponse
    {
        $data = $request->validate([
            'actual_cash' => ['required', 'integer', 'min:0'],
            'pin' => ['required', 'string'],
            'variance_reason' => ['nullable', 'string'],
            'approver_id' => ['nullable', 'uuid'],
        ]);

        $session = ShiftSession::query()
            ->where('business_id', $request->user()->business_id)
            ->where('id', $shift)
            ->firstOrFail();

        try {
            $this->terminalContext->assertShiftForActor($request, $session, requireOpenShift: false);
        } catch (TerminalContextException $error) {
            return $this->terminalContextError($error);
        }

        try {
            $closed = $this->shifts->close($request->user()->business_id, $session->id, $data, $request->user()->id);
        } catch (ShiftStateException $error) {
            return $this->shiftStateError($error);
        }

        return ApiResponse::success($this->shiftPayload($closed));
    }

    private function terminalContextError(TerminalContextException $error): JsonResponse
    {
        return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
    }

    private function shiftStateError(ShiftStateException $error): JsonResponse
    {
        return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
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
            'variance_reason' => $shift->variance_reason,
            'approved_by' => $shift->approved_by,
            'close_report' => $this->shifts->closeReport($shift),
            'payment_totals' => $this->shifts->paymentTotals($shift),
        ];
    }
}
