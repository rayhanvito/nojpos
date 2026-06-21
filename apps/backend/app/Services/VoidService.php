<?php

namespace App\Services;

use App\Models\User;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class VoidService
{
    public function __construct(
        private readonly TerminalContextService $terminalContext,
        private readonly AuditLogService $auditLog,
    ) {}

    /**
     * @param  array<string, mixed>  $data
     * @return array<string, mixed>
     */
    public function void(User $actor, Request $request, array $data): array
    {
        $this->assertAuthorization($actor);

        return DB::transaction(function () use ($actor, $request, $data): array {
            $businessId = (string) $actor->business_id;
            $transaction = DB::table('transactions')
                ->where('id', $data['transaction_id'])
                ->where('business_id', $businessId)
                ->lockForUpdate()
                ->first();

            if (! $transaction) {
                throw new VoidOperationException('NOT_FOUND', 'Transaction not found.', [], 404);
            }

            if ($transaction->shift_id !== $data['shift_id']) {
                throw new VoidOperationException('VOID_SHIFT_MISMATCH', 'Transaction can only be voided from the same shift.', [], 422);
            }

            try {
                $this->terminalContext->assertTransactionContext($request, $transaction, requireOpenShift: true);
            } catch (TerminalContextException $error) {
                throw new VoidOperationException($error->errorCode, $error->getMessage(), $error->details, $error->status);
            }

            $shift = DB::table('shift_sessions')
                ->where('id', $data['shift_id'])
                ->where('business_id', $businessId)
                ->where('status', 'open')
                ->lockForUpdate()
                ->first();

            if (! $shift) {
                throw new VoidOperationException('SHIFT_NOT_OPEN', 'Shift must be open to void a transaction.', [], 422);
            }

            if ($transaction->status === 'voided') {
                throw new VoidOperationException('TRANSACTION_ALREADY_VOIDED', 'Transaction has already been voided.', [], 422);
            }

            $before = (array) $transaction;
            $now = now();

            $this->reverseSaleStock($businessId, $transaction, $actor->id, (string) $data['reason'], $now);
            $cashReversal = $this->reverseCashPayments($businessId, $transaction, $actor->id, (string) $data['reason'], $now);

            DB::table('transactions')
                ->where('business_id', $businessId)
                ->where('id', $transaction->id)
                ->update([
                    'status' => 'voided',
                    'updated_at' => $now,
                ]);

            $after = array_merge($before, ['status' => 'voided', 'updated_at' => $now]);
            $this->auditLog->record(
                $businessId,
                $actor->id,
                'transaction.void',
                'transaction',
                $transaction->id,
                $before,
                $after,
                [
                    'outlet_id' => $transaction->outlet_id,
                    'device_id' => $transaction->device_id,
                    'request_id' => $request->header('X-Request-Id') ?: $request->header('X-Correlation-Id'),
                    'idempotency_key' => $request->attributes->get('idempotency_key') ?: $request->header('Idempotency-Key'),
                    'event_version' => 1,
                ],
            );

            return [
                'transaction_id' => $transaction->id,
                'status' => 'voided',
                'reason' => $data['reason'],
                'cash_reversal' => $cashReversal,
            ];
        });
    }

    private function assertAuthorization(User $actor): void
    {
        if (! in_array((string) $actor->role, ['admin', 'supervisor', 'owner'], true)) {
            throw new VoidOperationException('FORBIDDEN', 'User is not allowed to void transactions.', [], 403);
        }
    }

    private function reverseSaleStock(string $businessId, object $transaction, string $actorId, string $reason, mixed $now): void
    {
        $saleMovements = DB::table('stock_movements')
            ->where('business_id', $businessId)
            ->where('transaction_id', $transaction->id)
            ->where('type', 'sale')
            ->lockForUpdate()
            ->get();

        foreach ($saleMovements as $movement) {
            DB::table('stock_movements')->insert([
                'id' => (string) Str::uuid(),
                'business_id' => $businessId,
                'outlet_id' => $movement->outlet_id,
                'product_id' => $movement->product_id,
                'transaction_id' => $transaction->id,
                'actor_id' => $actorId,
                'type' => 'void',
                'quantity_delta' => -1 * (int) $movement->quantity_delta,
                'reason' => $reason,
                'created_at' => $now,
                'updated_at' => $now,
            ]);
        }
    }

    private function reverseCashPayments(string $businessId, object $transaction, string $actorId, string $reason, mixed $now): int
    {
        $cashReversal = (int) DB::table('payments')
            ->where('business_id', $businessId)
            ->where('transaction_id', $transaction->id)
            ->where('is_cash', true)
            ->where('status', 'confirmed')
            ->lockForUpdate()
            ->sum('amount');

        if ($cashReversal <= 0) {
            return 0;
        }

        DB::table('cash_movements')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'outlet_id' => $transaction->outlet_id,
            'shift_id' => $transaction->shift_id,
            'actor_id' => $actorId,
            'type' => 'cash_out',
            'amount' => $cashReversal,
            'reason' => 'Void: '.$reason,
            'created_at' => $now,
            'updated_at' => $now,
        ]);

        return $cashReversal;
    }
}
