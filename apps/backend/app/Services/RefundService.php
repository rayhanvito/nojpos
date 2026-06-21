<?php

namespace App\Services;

use App\Models\User;
use App\Support\Nojpos;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;

class RefundService
{
    /**
     * @param  array<string, mixed>  $data
     * @return array<string, mixed>
     */
    public function create(User $actor, string $transactionId, array $data, ?string $idempotencyKey): array
    {
        return DB::transaction(function () use ($actor, $transactionId, $data, $idempotencyKey): array {
            $businessId = (string) $actor->business_id;
            $transaction = DB::table('transactions')
                ->where('business_id', $businessId)
                ->where('id', $transactionId)
                ->lockForUpdate()
                ->first();

            if (! $transaction) {
                throw new RefundOperationException('NOT_FOUND', 'Transaction not found.', [], 404);
            }

            $this->assertRefundableStatus($transaction);
            $this->assertAuthorization($actor, $transaction, $data);

            $refundMethod = (string) $data['refund_method'];
            $cashShift = $refundMethod === 'cash' ? $this->cashShiftForRefund($transaction) : null;
            if ($refundMethod === 'original_method') {
                $this->assertOriginalMethodSupported($transaction);
            }

            $lines = $this->calculateLines($businessId, $transaction, $data['lines']);
            $totalAmount = array_sum(array_map(fn (array $line): int => $line['amount'], $lines));
            $alreadyRefunded = $this->refundedAmount($businessId, $transaction->id);
            $remainingBefore = max(0, (int) $transaction->grand_total - $alreadyRefunded);

            if ($totalAmount <= 0 || $totalAmount > $remainingBefore) {
                throw new RefundOperationException('REFUND_EXCEEDS_PAID', 'Refund amount exceeds remaining refundable amount.', [
                    'remaining_refundable_amount' => $remainingBefore,
                    'requested_amount' => $totalAmount,
                ], 422);
            }

            $now = now();
            $refundId = (string) Str::uuid();
            $statusAfter = ($remainingBefore - $totalAmount) === 0 ? 'refunded' : 'partially_refunded';
            $before = (array) $transaction;

            DB::table('refunds')->insert([
                'id' => $refundId,
                'business_id' => $businessId,
                'outlet_id' => $transaction->outlet_id,
                'transaction_id' => $transaction->id,
                'shift_id' => $cashShift?->id ?? $transaction->shift_id,
                'actor_id' => $actor->id,
                'authorizer_id' => $actor->id,
                'refund_method' => $refundMethod,
                'total_amount' => $totalAmount,
                'reason' => $data['reason'],
                'notes' => $data['notes'] ?? null,
                'status' => 'finalized',
                'idempotency_key' => $idempotencyKey,
                'finalized_at' => $now,
                'created_at' => $now,
                'updated_at' => $now,
            ]);

            $stockMovements = [];
            foreach ($lines as $line) {
                $lineId = (string) Str::uuid();
                DB::table('refund_lines')->insert([
                    'id' => $lineId,
                    'business_id' => $businessId,
                    'refund_id' => $refundId,
                    'transaction_item_id' => $line['transaction_item_id'],
                    'product_id' => $line['product_id'],
                    'quantity' => $line['quantity'],
                    'amount' => $line['amount'],
                    'restock' => $line['restock'],
                    'non_restock_reason' => $line['non_restock_reason'],
                    'created_at' => $now,
                    'updated_at' => $now,
                ]);

                if ($line['restock'] && $line['track_stock']) {
                    $movementId = (string) Str::uuid();
                    $beforeQty = $this->stockOnHand($businessId, $transaction->outlet_id, $line['product_id']);
                    DB::table('stock_movements')->insert([
                        'id' => $movementId,
                        'business_id' => $businessId,
                        'outlet_id' => $transaction->outlet_id,
                        'product_id' => $line['product_id'],
                        'transaction_id' => $transaction->id,
                        'refund_id' => $refundId,
                        'actor_id' => $actor->id,
                        'type' => 'refund_reversal',
                        'quantity_delta' => $line['quantity'],
                        'before_quantity' => $beforeQty,
                        'after_quantity' => $beforeQty + $line['quantity'],
                        'reason' => $data['reason'],
                        'created_at' => $now,
                        'updated_at' => $now,
                    ]);
                    $stockMovements[] = [
                        'movement_id' => $movementId,
                        'product_id' => $line['product_id'],
                        'quantity_delta' => $line['quantity'],
                        'type' => 'refund_reversal',
                    ];
                }
            }

            $cashMovement = null;
            if ($refundMethod === 'cash') {
                $cashMovementId = (string) Str::uuid();
                DB::table('cash_movements')->insert([
                    'id' => $cashMovementId,
                    'business_id' => $businessId,
                    'outlet_id' => $transaction->outlet_id,
                    'shift_id' => $cashShift->id,
                    'actor_id' => $actor->id,
                    'refund_id' => $refundId,
                    'type' => 'cash_out',
                    'amount' => $totalAmount,
                    'reason' => 'Refund: '.$data['reason'],
                    'created_at' => $now,
                    'updated_at' => $now,
                ]);
                $cashMovement = [
                    'movement_id' => $cashMovementId,
                    'type' => 'cash_out',
                    'amount' => $totalAmount,
                ];
            }

            DB::table('transactions')
                ->where('business_id', $businessId)
                ->where('id', $transaction->id)
                ->update([
                    'status' => $statusAfter,
                    'updated_at' => $now,
                ]);

            Nojpos::audit($businessId, $actor->id, 'refund.create', 'refund', $refundId, [
                'transaction_id' => $transaction->id,
                'status' => $before['status'],
                'remaining_refundable_amount' => $remainingBefore,
            ], [
                'transaction_id' => $transaction->id,
                'status' => $statusAfter,
                'total_amount' => $totalAmount,
                'refund_method' => $refundMethod,
            ]);

            return $this->payload(
                $businessId,
                $refundId,
                $transaction->id,
                $statusAfter,
                $stockMovements,
                $cashMovement,
                $remainingBefore - $totalAmount,
            );
        });
    }

    private function assertRefundableStatus(object $transaction): void
    {
        if ($transaction->status === 'refunded') {
            throw new RefundOperationException('REFUND_ALREADY_COMPLETE', 'Transaction is already fully refunded.', [], 409);
        }

        if (! in_array($transaction->status, ['paid', 'partially_refunded'], true)) {
            throw new RefundOperationException('REFUND_STATE_INVALID', 'Only paid transactions can be refunded.', [
                'transaction_status' => $transaction->status,
            ], 422);
        }
    }

    /**
     * @param  array<string, mixed>  $data
     */
    private function assertAuthorization(User $actor, object $transaction, array $data): void
    {
        if (! in_array((string) $actor->role, ['owner', 'admin'], true)) {
            throw new RefundOperationException('FORBIDDEN', 'User is not allowed to refund transactions.', [], 403);
        }

        if ($actor->role !== 'admin' || ! $this->refundRequiresPin((string) $actor->business_id)) {
            return;
        }

        $pin = (string) ($data['authorization_pin'] ?? '');
        if ($pin === '') {
            throw new RefundOperationException('PIN_REQUIRED', 'Authorization PIN is required for refunds.', [], 422);
        }

        if (! Hash::check($pin, $actor->pin_hash ?? '')) {
            throw new RefundOperationException('INVALID_PIN', 'Authorization PIN is invalid.', [], 422);
        }
    }

    private function refundRequiresPin(string $businessId): bool
    {
        $setting = DB::table('business_security_settings')
            ->where('business_id', $businessId)
            ->first();

        return $setting ? (bool) $setting->pin_required_refund : true;
    }

    private function cashShiftForRefund(object $transaction): object
    {
        $shift = DB::table('shift_sessions')
            ->where('business_id', $transaction->business_id)
            ->where('outlet_id', $transaction->outlet_id)
            ->where('status', 'open')
            ->lockForUpdate()
            ->latest('opened_at')
            ->first();

        if (! $shift) {
            throw new RefundOperationException('SHIFT_NOT_OPEN', 'An open shift is required for cash refunds.', [], 422);
        }

        return $shift;
    }

    private function assertOriginalMethodSupported(object $transaction): void
    {
        $hasNonCash = DB::table('payments')
            ->where('business_id', $transaction->business_id)
            ->where('transaction_id', $transaction->id)
            ->whereIn('status', ['confirmed', 'settled'])
            ->where('is_cash', false)
            ->exists();

        if (! $hasNonCash) {
            throw new RefundOperationException('REFUND_METHOD_UNSUPPORTED', 'Original-method refund is only available for non-cash payments.', [], 422);
        }
    }

    /**
     * @param  array<int, array<string, mixed>>  $requestedLines
     * @return array<int, array<string, mixed>>
     */
    private function calculateLines(string $businessId, object $transaction, array $requestedLines): array
    {
        $ids = array_values(array_unique(array_map(fn (array $line): string => (string) $line['transaction_item_id'], $requestedLines)));
        if (count($ids) !== count($requestedLines)) {
            throw new RefundOperationException('VALIDATION_ERROR', 'Duplicate refund lines are not allowed.', ['lines' => ['Duplicate transaction item IDs are not allowed.']], 422);
        }

        $items = DB::table('transaction_items as items')
            ->leftJoin('products as products', function ($join) use ($businessId): void {
                $join->on('products.id', '=', 'items.product_id')
                    ->where('products.business_id', '=', $businessId);
            })
            ->where('items.business_id', $businessId)
            ->where('items.transaction_id', $transaction->id)
            ->whereIn('items.id', $ids)
            ->lockForUpdate()
            ->select([
                'items.id',
                'items.product_id',
                'items.quantity',
                'items.subtotal',
                'items.discount',
                'products.track_stock',
            ])
            ->get()
            ->keyBy('id');

        if ($items->count() !== count($ids)) {
            throw new RefundOperationException('REFUND_LINE_INVALID', 'Refund line is not part of this transaction.', [], 422);
        }

        $lines = [];
        foreach ($requestedLines as $requestedLine) {
            $item = $items[(string) $requestedLine['transaction_item_id']];
            $quantity = (int) $requestedLine['quantity'];
            $restock = array_key_exists('restock', $requestedLine) ? (bool) $requestedLine['restock'] : true;
            $nonRestockReason = $requestedLine['non_restock_reason'] ?? null;

            if (! $restock && ! is_string($nonRestockReason)) {
                throw new RefundOperationException('VALIDATION_ERROR', 'Non-restock reason is required when restock is false.', [
                    'lines' => ['non_restock_reason is required when restock is false.'],
                ], 422);
            }

            $refundedQuantity = (int) DB::table('refund_lines as lines')
                ->join('refunds as refunds', function ($join) use ($businessId): void {
                    $join->on('refunds.id', '=', 'lines.refund_id')
                        ->where('refunds.business_id', '=', $businessId)
                        ->where('refunds.status', '=', 'finalized');
                })
                ->where('lines.business_id', $businessId)
                ->where('lines.transaction_item_id', $item->id)
                ->sum('lines.quantity');

            $refundableQuantity = (int) $item->quantity - $refundedQuantity;
            if ($quantity > $refundableQuantity) {
                throw new RefundOperationException('REFUND_QUANTITY_EXCEEDED', 'Refund quantity exceeds remaining line quantity.', [
                    'transaction_item_id' => $item->id,
                    'remaining_quantity' => max(0, $refundableQuantity),
                    'requested_quantity' => $quantity,
                ], 422);
            }

            $lineNet = max(0, (int) $item->subtotal - (int) $item->discount);
            $refundedAmount = (int) DB::table('refund_lines as lines')
                ->join('refunds as refunds', function ($join) use ($businessId): void {
                    $join->on('refunds.id', '=', 'lines.refund_id')
                        ->where('refunds.business_id', '=', $businessId)
                        ->where('refunds.status', '=', 'finalized');
                })
                ->where('lines.business_id', $businessId)
                ->where('lines.transaction_item_id', $item->id)
                ->sum('lines.amount');
            $remainingLineAmount = max(0, $lineNet - $refundedAmount);
            $amount = $quantity === $refundableQuantity
                ? $remainingLineAmount
                : min($remainingLineAmount, intdiv($lineNet * $quantity, max(1, (int) $item->quantity)));

            $lines[] = [
                'transaction_item_id' => $item->id,
                'product_id' => $item->product_id,
                'quantity' => $quantity,
                'amount' => $amount,
                'restock' => $restock,
                'non_restock_reason' => $restock ? null : trim((string) $nonRestockReason),
                'track_stock' => (bool) $item->track_stock,
            ];
        }

        return $lines;
    }

    private function refundedAmount(string $businessId, string $transactionId): int
    {
        return (int) DB::table('refunds')
            ->where('business_id', $businessId)
            ->where('transaction_id', $transactionId)
            ->where('status', 'finalized')
            ->sum('total_amount');
    }

    private function stockOnHand(string $businessId, string $outletId, string $productId): int
    {
        return (int) DB::table('stock_movements')
            ->where('business_id', $businessId)
            ->where('outlet_id', $outletId)
            ->where('product_id', $productId)
            ->sum('quantity_delta');
    }

    /**
     * @param  array<int, array<string, mixed>>  $stockMovements
     * @param  array<string, mixed>|null  $cashMovement
     * @return array<string, mixed>
     */
    private function payload(
        string $businessId,
        string $refundId,
        string $transactionId,
        string $transactionStatus,
        array $stockMovements,
        ?array $cashMovement,
        int $remainingRefundableAmount,
    ): array {
        $refund = DB::table('refunds')
            ->where('business_id', $businessId)
            ->where('id', $refundId)
            ->first();

        $lines = DB::table('refund_lines')
            ->where('business_id', $businessId)
            ->where('refund_id', $refundId)
            ->orderBy('created_at')
            ->get()
            ->map(fn (object $line): array => [
                'id' => $line->id,
                'transaction_item_id' => $line->transaction_item_id,
                'product_id' => $line->product_id,
                'quantity' => (int) $line->quantity,
                'amount' => (int) $line->amount,
                'restock' => (bool) $line->restock,
                'non_restock_reason' => $line->non_restock_reason,
            ])
            ->values()
            ->all();

        return [
            'id' => $refund->id,
            'transaction_id' => $transactionId,
            'transaction_status' => $transactionStatus,
            'status' => $refund->status,
            'refund_method' => $refund->refund_method,
            'total_refund_amount' => (int) $refund->total_amount,
            'reason' => $refund->reason,
            'refunded_lines' => $lines,
            'stock_movements' => $stockMovements,
            'cash_movement' => $cashMovement,
            'remaining_refundable_amount' => $remainingRefundableAmount,
            'server_time' => now()->toISOString(),
        ];
    }
}
