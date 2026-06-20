<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Services\TerminalContextException;
use App\Services\TerminalContextService;
use App\Support\ApiResponse;
use App\Support\Nojpos;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class VoidController extends Controller
{
    public function __construct(private readonly TerminalContextService $terminalContext) {}

    public function store(Request $request): JsonResponse
    {
        $data = $request->validate([
            'transaction_id' => ['required', 'uuid'],
            'shift_id' => ['required', 'uuid'],
            'reason' => ['required', 'string'],
        ]);

        if (! $this->canVoid($request->user()->role)) {
            return ApiResponse::error('FORBIDDEN', 'User is not allowed to void transactions.', [], 403);
        }

        $businessId = $request->user()->business_id;
        $transaction = DB::table('transactions')
            ->where('id', $data['transaction_id'])
            ->where('business_id', $businessId)
            ->first();

        if (! $transaction) {
            return ApiResponse::error('NOT_FOUND', 'Transaction not found.', [], 404);
        }

        if ($transaction->shift_id !== $data['shift_id']) {
            return ApiResponse::error('VOID_SHIFT_MISMATCH', 'Transaction can only be voided from the same shift.', [], 422);
        }

        try {
            $this->terminalContext->assertTransactionContext($request, $transaction, requireOpenShift: true);
        } catch (TerminalContextException $error) {
            return $this->terminalContextError($error);
        }

        $shiftIsOpen = DB::table('shift_sessions')
            ->where('id', $data['shift_id'])
            ->where('business_id', $businessId)
            ->where('status', 'open')
            ->exists();

        if (! $shiftIsOpen) {
            return ApiResponse::error('SHIFT_NOT_OPEN', 'Shift must be open to void a transaction.', [], 422);
        }

        if ($transaction->status === 'voided') {
            return ApiResponse::error('TRANSACTION_ALREADY_VOIDED', 'Transaction has already been voided.', [], 422);
        }

        $before = (array) $transaction;
        $now = now();
        $cashReversal = 0;

        DB::transaction(function () use ($businessId, $data, $request, $transaction, $before, $now, &$cashReversal): void {
            $saleMovements = DB::table('stock_movements')
                ->where('business_id', $businessId)
                ->where('transaction_id', $transaction->id)
                ->where('type', 'sale')
                ->get();

            foreach ($saleMovements as $movement) {
                DB::table('stock_movements')->insert([
                    'id' => (string) Str::uuid(),
                    'business_id' => $businessId,
                    'outlet_id' => $movement->outlet_id,
                    'product_id' => $movement->product_id,
                    'transaction_id' => $transaction->id,
                    'type' => 'void',
                    'quantity_delta' => -1 * (int) $movement->quantity_delta,
                    'created_at' => $now,
                    'updated_at' => $now,
                ]);
            }

            $cashReversal = (int) DB::table('payments')
                ->where('business_id', $businessId)
                ->where('transaction_id', $transaction->id)
                ->where('is_cash', true)
                ->where('status', 'confirmed')
                ->sum('amount');

            if ($cashReversal > 0) {
                DB::table('cash_movements')->insert([
                    'id' => (string) Str::uuid(),
                    'business_id' => $businessId,
                    'outlet_id' => $transaction->outlet_id,
                    'shift_id' => $transaction->shift_id,
                    'actor_id' => $request->user()->id,
                    'type' => 'cash_out',
                    'amount' => $cashReversal,
                    'reason' => 'Void: '.$data['reason'],
                    'created_at' => $now,
                    'updated_at' => $now,
                ]);
            }

            DB::table('transactions')
                ->where('id', $transaction->id)
                ->update([
                    'status' => 'voided',
                    'updated_at' => $now,
                ]);

            $after = array_merge($before, ['status' => 'voided']);
            Nojpos::audit($businessId, $request->user()->id, 'void', 'transaction', $transaction->id, $before, $after);
        });

        return ApiResponse::success([
            'transaction_id' => $transaction->id,
            'status' => 'voided',
            'reason' => $data['reason'],
            'cash_reversal' => $cashReversal,
        ]);
    }

    private function terminalContextError(TerminalContextException $error): JsonResponse
    {
        return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
    }

    private function canVoid(string $role): bool
    {
        return in_array($role, ['admin', 'supervisor', 'owner'], true);
    }
}
