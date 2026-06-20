<?php

namespace App\Services;

use App\Models\ShiftSession;
use App\Support\Nojpos;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;

class ShiftService
{
    /**
     * @param  array{outlet_id: string, device_id: string, cashier_id: string, opening_cash: int}  $data
     */
    public function open(string $businessId, array $data, string $actorId): ShiftSession
    {
        return DB::transaction(function () use ($businessId, $data, $actorId): ShiftSession {
            $existing = ShiftSession::query()
                ->where('business_id', $businessId)
                ->where('outlet_id', $data['outlet_id'])
                ->where('device_id', $data['device_id'])
                ->where('status', 'open')
                ->lockForUpdate()
                ->first();

            if ($existing) {
                throw new ShiftStateException('SHIFT_ALREADY_OPEN', 'A shift is already open for this device and outlet.');
            }

            $shift = ShiftSession::query()->create([
                'business_id' => $businessId,
                'outlet_id' => $data['outlet_id'],
                'device_id' => $data['device_id'],
                'cashier_id' => $data['cashier_id'],
                'status' => 'open',
                'opening_cash' => (int) $data['opening_cash'],
                'opened_at' => now(),
            ]);

            Nojpos::audit($businessId, $actorId, 'shift.open', 'shift_session', $shift->id, null, $this->snapshot($shift));

            return $shift->refresh();
        });
    }

    /**
     * @param  array{type: string, amount: int, reason: string}  $data
     * @return array{id: string, type: string, amount: int, reason: string}
     */
    public function cashMovement(string $businessId, string $shiftId, array $data, string $actorId): array
    {
        return DB::transaction(function () use ($businessId, $shiftId, $data, $actorId): array {
            $shift = ShiftSession::query()
                ->where('business_id', $businessId)
                ->where('id', $shiftId)
                ->lockForUpdate()
                ->first();

            if (! $shift) {
                throw new ShiftStateException('NOT_FOUND', 'Shift not found.', [], 404);
            }

            if ($shift->status !== 'open') {
                throw new ShiftStateException('SHIFT_ALREADY_CLOSED', 'Shift is already closed.');
            }

            $id = (string) Str::uuid();
            $movement = [
                'id' => $id,
                'business_id' => $businessId,
                'outlet_id' => $shift->outlet_id,
                'shift_id' => $shift->id,
                'actor_id' => $actorId,
                'type' => $data['type'],
                'amount' => (int) $data['amount'],
                'reason' => $data['reason'],
                'created_at' => now(),
                'updated_at' => now(),
            ];

            DB::table('cash_movements')->insert($movement);
            Nojpos::audit($businessId, $actorId, 'shift.cash_movement', 'cash_movement', $id, null, [
                'id' => $id,
                'shift_id' => $shift->id,
                'type' => $data['type'],
                'amount' => (int) $data['amount'],
                'reason' => $data['reason'],
            ]);

            return [
                'id' => $id,
                'type' => $data['type'],
                'amount' => (int) $data['amount'],
                'reason' => $data['reason'],
            ];
        });
    }

    /**
     * @param  array{actual_cash: int, pin: string, variance_reason?: string|null, approver_id?: string|null}  $data
     */
    public function close(string $businessId, string $shiftId, array $data, string $actorId): ShiftSession
    {
        return DB::transaction(function () use ($businessId, $shiftId, $data, $actorId): ShiftSession {
            $shift = ShiftSession::query()
                ->where('business_id', $businessId)
                ->where('id', $shiftId)
                ->lockForUpdate()
                ->first();

            if (! $shift) {
                throw new ShiftStateException('NOT_FOUND', 'Shift not found.', [], 404);
            }

            if ($shift->status !== 'open') {
                throw new ShiftStateException('SHIFT_ALREADY_CLOSED', 'Shift is already closed.');
            }

            $this->assertFreshPin($businessId, $actorId, $data['pin']);

            $summary = $this->summary($shift);
            $actual = (int) $data['actual_cash'];
            $difference = $actual - $summary['expected_cash'];
            $varianceReason = trim((string) ($data['variance_reason'] ?? ''));

            if ($difference !== 0 && $varianceReason === '') {
                throw new ShiftStateException(
                    'SHIFT_VARIANCE_REASON_REQUIRED',
                    'A variance reason is required when actual cash differs from expected cash.',
                    ['variance' => $difference],
                );
            }

            $approverId = $this->approvedBy($businessId, $data['approver_id'] ?? null);
            $report = [
                'shift_id' => $shift->id,
                'business_id' => $businessId,
                'outlet_id' => $shift->outlet_id,
                'device_id' => $shift->device_id,
                'cashier_id' => $shift->cashier_id,
                'opened_at' => $shift->opened_at,
                'closed_at' => now()->toISOString(),
                'opening_cash' => (int) $shift->opening_cash,
                'cash_sales' => $summary['cash_sales'],
                'cash_in' => $summary['cash_in'],
                'cash_out' => $summary['cash_out'],
                'expected_cash' => $summary['expected_cash'],
                'actual_cash' => $actual,
                'cash_difference' => $difference,
                'variance_reason' => $varianceReason === '' ? null : $varianceReason,
                'approved_by' => $approverId,
                'payment_totals' => $this->paymentTotals($shift),
            ];

            $before = $this->snapshot($shift);
            $shift->update([
                'status' => 'closed',
                'expected_cash' => $summary['expected_cash'],
                'actual_cash' => $actual,
                'cash_difference' => $difference,
                'variance_reason' => $varianceReason === '' ? null : $varianceReason,
                'approved_by' => $approverId,
                'close_report_snapshot' => json_encode($report, JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE),
                'closed_at' => now(),
            ]);

            Nojpos::audit($businessId, $actorId, 'shift.close', 'shift_session', $shift->id, $before, $report);

            return $shift->refresh();
        });
    }

    public function closeReport(ShiftSession $shift): ?array
    {
        $snapshot = $shift->close_report_snapshot;
        if (! $snapshot) {
            return null;
        }

        if (is_array($snapshot)) {
            return $snapshot;
        }

        $decoded = json_decode((string) $snapshot, true);

        return is_array($decoded) ? $decoded : null;
    }

    /**
     * @return array{cash_sales: int, cash_in: int, cash_out: int, expected_cash: int}
     */
    public function summary(ShiftSession $shift): array
    {
        $cashSales = (int) DB::table('payments')
            ->join('transactions', 'transactions.id', '=', 'payments.transaction_id')
            ->where('transactions.shift_id', $shift->id)
            ->where('payments.business_id', $shift->business_id)
            ->where('payments.is_cash', true)
            ->where('payments.status', 'confirmed')
            ->sum('payments.amount');

        $cashIn = (int) DB::table('cash_movements')
            ->where('business_id', $shift->business_id)
            ->where('shift_id', $shift->id)
            ->where('type', 'cash_in')
            ->sum('amount');

        $cashOut = (int) DB::table('cash_movements')
            ->where('business_id', $shift->business_id)
            ->where('shift_id', $shift->id)
            ->where('type', 'cash_out')
            ->sum('amount');

        return [
            'cash_sales' => $cashSales,
            'cash_in' => $cashIn,
            'cash_out' => $cashOut,
            'expected_cash' => (int) $shift->opening_cash + $cashSales + $cashIn - $cashOut,
        ];
    }

    public function paymentTotals(ShiftSession $shift): array
    {
        return DB::table('payments')
            ->join('transactions', 'transactions.id', '=', 'payments.transaction_id')
            ->where('transactions.shift_id', $shift->id)
            ->where('payments.business_id', $shift->business_id)
            ->where('payments.status', 'confirmed')
            ->select('payments.method', 'payments.is_cash', DB::raw('sum(payments.amount) as amount'))
            ->groupBy('payments.method', 'payments.is_cash')
            ->orderBy('payments.method')
            ->get()
            ->map(fn ($row): array => [
                'method' => $row->method,
                'amount' => (int) $row->amount,
                'is_cash' => (bool) $row->is_cash,
            ])
            ->values()
            ->all();
    }

    private function approvedBy(string $businessId, ?string $approverId): ?string
    {
        if (! $approverId) {
            return null;
        }

        $approver = DB::table('users')
            ->where('business_id', $businessId)
            ->where('id', $approverId)
            ->whereNull('deleted_at')
            ->first();

        if (! $approver || ! in_array($approver->role, ['owner', 'admin', 'supervisor'], true)) {
            throw new ShiftStateException('FORBIDDEN', 'Approver is not authorized for shift close variance.', [], 403);
        }

        return $approverId;
    }

    private function assertFreshPin(string $businessId, string $actorId, string $pin): void
    {
        $actor = DB::table('users')
            ->where('business_id', $businessId)
            ->where('id', $actorId)
            ->whereNull('deleted_at')
            ->first();

        if (! $actor || ! $actor->pin_hash || ! Hash::check($pin, $actor->pin_hash)) {
            throw new ShiftStateException('INVALID_PIN', 'PIN is invalid.', [], 422);
        }
    }

    private function snapshot(ShiftSession $shift): array
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
