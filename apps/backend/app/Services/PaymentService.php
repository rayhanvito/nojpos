<?php

namespace App\Services;

use App\Models\Payment;
use App\Support\Nojpos;
use Carbon\CarbonImmutable;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class PaymentService
{
    private const PAYABLE_TRANSACTION_STATUSES = ['pending', 'unpaid', 'partial'];

    /**
     * @param  array<int, array{method: string, amount: int, reference?: string|null}>  $payments
     * @return array{rows: array<int, array<string, mixed>>, confirmed_total: int, has_async: bool}
     */
    public function prepareCheckoutAllocations(string $businessId, string $outletId, array $payments, int $grandTotal): array
    {
        $rows = [];
        $confirmedTotal = 0;
        $hasAsync = false;
        $allocatedTotal = 0;

        foreach ($payments as $payment) {
            $amount = (int) $payment['amount'];
            $allocatedTotal += $amount;

            if ($allocatedTotal > $grandTotal) {
                throw new PaymentStateException(
                    'PAYMENT_OVERPAY',
                    'Payment amount exceeds the remaining payable total.',
                    [
                        'remaining' => max(0, $grandTotal - ($allocatedTotal - $amount)),
                        'amount' => $amount,
                    ],
                );
            }

            $config = $this->paymentMethodConfig($businessId, $outletId, $payment['method']);
            $isCash = $config ? (bool) $config->is_cash : $payment['method'] === 'cash';
            $isAsync = ! $isCash && $config && (bool) $config->async_confirmation_enabled;
            $status = $isCash ? 'confirmed' : 'pending';

            if ($status === 'confirmed') {
                $confirmedTotal += $amount;
            }

            $rows[] = [
                'method' => $payment['method'],
                'reference' => $payment['reference'] ?? null,
                'amount' => $amount,
                'status' => $status,
                'is_cash' => $isCash,
                'provider' => $isAsync ? ($config->provider ?: $payment['method']) : null,
                'provider_reference' => $isAsync ? (string) Str::uuid() : null,
                'confirm_expires_at' => $isAsync ? now()->addMinutes(max(1, (int) $config->async_expiry_minutes)) : null,
            ];
            $hasAsync = $hasAsync || $isAsync;
        }

        return [
            'rows' => $rows,
            'confirmed_total' => $confirmedTotal,
            'has_async' => $hasAsync,
        ];
    }

    /**
     * @param  array{method: string, amount: int, reference?: string|null}  $data
     */
    public function create(string $businessId, string $transactionId, array $data, string $actorId): Payment
    {
        return DB::transaction(function () use ($businessId, $transactionId, $data, $actorId): Payment {
            $transaction = DB::table('transactions')
                ->where('business_id', $businessId)
                ->where('id', $transactionId)
                ->lockForUpdate()
                ->first();

            if (! $transaction) {
                throw new PaymentStateException('NOT_FOUND', 'Transaction not found.', [], 404);
            }

            $this->assertTransactionPayable($transaction);

            $amount = (int) $data['amount'];
            $this->assertAmountWithinRemaining(
                $transaction,
                $amount,
                includePending: true,
            );

            $isCash = $this->isCashMethod($businessId, $transaction->outlet_id, $data['method']);
            $status = $isCash ? 'confirmed' : 'pending';
            $payment = Payment::query()->create([
                'business_id' => $businessId,
                'transaction_id' => $transaction->id,
                'method' => $data['method'],
                'reference' => $data['reference'] ?? null,
                'amount' => $amount,
                'status' => $status,
                'is_cash' => $isCash,
                'confirmed_by' => $status === 'confirmed' ? $actorId : null,
                'confirmed_at' => $status === 'confirmed' ? now() : null,
            ]);

            $this->updateTransactionStatusLocked($transaction->id);

            Nojpos::audit($businessId, $actorId, 'payment.create', 'payment', $payment->id, null, $this->paymentSnapshot($payment));

            return $payment->refresh();
        });
    }

    public function confirm(string $businessId, string $paymentId, string $actorId): Payment
    {
        return DB::transaction(function () use ($businessId, $paymentId, $actorId): Payment {
            $payment = Payment::query()
                ->where('business_id', $businessId)
                ->where('id', $paymentId)
                ->lockForUpdate()
                ->first();

            if (! $payment) {
                throw new PaymentStateException('NOT_FOUND', 'Payment not found.', [], 404);
            }

            $transaction = DB::table('transactions')
                ->where('business_id', $businessId)
                ->where('id', $payment->transaction_id)
                ->lockForUpdate()
                ->first();

            if (! $transaction) {
                throw new PaymentStateException('NOT_FOUND', 'Transaction not found.', [], 404);
            }

            if ($payment->status === 'confirmed') {
                return $payment->refresh();
            }

            $this->assertTransactionPayable($transaction);

            if ($payment->status !== 'pending') {
                throw new PaymentStateException(
                    'PAYMENT_STATE_INVALID',
                    'Payment cannot be confirmed from its current state.',
                    ['payment_status' => $payment->status],
                );
            }

            $this->assertAmountWithinRemaining(
                $transaction,
                (int) $payment->amount,
                includePending: false,
                excludingPaymentId: $payment->id,
            );

            $before = $this->paymentSnapshot($payment);

            $payment->update([
                'status' => 'confirmed',
                'confirmed_by' => $actorId,
                'confirmed_at' => now(),
            ]);

            $this->updateTransactionStatusLocked($transaction->id);

            Nojpos::audit($businessId, $actorId, 'payment.confirm', 'payment', $payment->id, $before, $this->paymentSnapshot($payment->refresh()));

            return $payment->refresh();
        });
    }

    public function settleWebhook(string $provider, string $reference, string $outcome): ?Payment
    {
        return DB::transaction(function () use ($provider, $reference, $outcome): ?Payment {
            $payment = Payment::query()->where('provider', $provider)->where('provider_reference', $reference)->lockForUpdate()->first();
            if (! $payment) {
                return null;
            }
            $transaction = DB::table('transactions')->where('business_id', $payment->business_id)->where('id', $payment->transaction_id)->lockForUpdate()->first();
            if (! $transaction || $payment->status !== 'pending') {
                return $payment->refresh();
            }
            if ($payment->confirm_expires_at && CarbonImmutable::parse($payment->confirm_expires_at)->isPast()) {
                return $this->failLocked($payment, $transaction, 'expired');
            }
            if ($outcome !== 'confirmed') {
                return $this->failLocked($payment, $transaction, $outcome === 'declined' ? 'declined' : 'failed');
            }
            $before = $this->paymentSnapshot($payment);
            $payment->update(['status' => 'confirmed', 'confirmed_at' => now()]);
            DB::table('transactions')->where('business_id', $payment->business_id)->where('id', $transaction->id)->update(['status' => 'paid', 'updated_at' => now()]);
            $this->applyStockLocked($transaction);
            Nojpos::audit($payment->business_id, null, 'payment.confirm', 'payment', $payment->id, $before, $this->paymentSnapshot($payment->refresh()));

            return $payment->refresh();
        });
    }

    public function expireTransaction(string $businessId, string $transactionId): void
    {
        DB::transaction(function () use ($businessId, $transactionId): void {
            $transaction = DB::table('transactions')->where('business_id', $businessId)->where('id', $transactionId)->lockForUpdate()->first();
            if (! $transaction || $transaction->status !== 'payment_pending') {
                return;
            }
            $payment = Payment::query()->where('business_id', $businessId)->where('transaction_id', $transactionId)->where('status', 'pending')->lockForUpdate()->first();
            if ($payment && $payment->confirm_expires_at && CarbonImmutable::parse($payment->confirm_expires_at)->isPast()) {
                $this->failLocked($payment, $transaction, 'expired');
            }
        });
    }

    private function assertTransactionPayable(object $transaction): void
    {
        if (! in_array($transaction->status, self::PAYABLE_TRANSACTION_STATUSES, true)) {
            throw new PaymentStateException(
                'PAYMENT_STATE_INVALID',
                'Transaction cannot accept payment in its current state.',
                ['transaction_status' => $transaction->status],
            );
        }
    }

    private function assertAmountWithinRemaining(
        object $transaction,
        int $amount,
        bool $includePending,
        ?string $excludingPaymentId = null,
    ): void {
        $query = DB::table('payments')
            ->where('business_id', $transaction->business_id)
            ->where('transaction_id', $transaction->id)
            ->whereNull('deleted_at')
            ->whereIn('status', $includePending ? ['pending', 'confirmed'] : ['confirmed']);

        if ($excludingPaymentId) {
            $query->where('id', '!=', $excludingPaymentId);
        }

        $allocated = (int) $query->sum('amount');
        $remaining = (int) $transaction->grand_total - $allocated;

        if ($amount > $remaining) {
            throw new PaymentStateException(
                'PAYMENT_OVERPAY',
                'Payment amount exceeds the remaining payable total.',
                [
                    'remaining' => max(0, $remaining),
                    'amount' => $amount,
                ],
            );
        }
    }

    private function isCashMethod(string $businessId, string $outletId, string $method): bool
    {
        $config = $this->paymentMethodConfig($businessId, $outletId, $method);

        return $config ? (bool) $config->is_cash : $method === 'cash';
    }

    private function paymentMethodConfig(string $businessId, string $outletId, string $method): ?object
    {
        return DB::table('payment_method_configs')
            ->where('business_id', $businessId)
            ->where('method', $method)
            ->where(function ($query) use ($outletId): void {
                $query->whereNull('outlet_id')->orWhere('outlet_id', $outletId);
            })
            ->first();
    }

    private function updateTransactionStatusLocked(string $transactionId): void
    {
        $transaction = DB::table('transactions')
            ->where('id', $transactionId)
            ->lockForUpdate()
            ->first();

        if (! $transaction) {
            return;
        }

        $confirmedTotal = (int) DB::table('payments')
            ->where('business_id', $transaction->business_id)
            ->where('transaction_id', $transactionId)
            ->whereNull('deleted_at')
            ->where('status', 'confirmed')
            ->sum('amount');

        $status = match (true) {
            $confirmedTotal >= (int) $transaction->grand_total => 'paid',
            $confirmedTotal > 0 => 'partial',
            default => 'unpaid',
        };

        DB::table('transactions')
            ->where('business_id', $transaction->business_id)
            ->where('id', $transactionId)
            ->update([
                'status' => $status,
                'updated_at' => now(),
            ]);
    }

    private function paymentSnapshot(Payment $payment): array
    {
        return [
            'id' => $payment->id,
            'business_id' => $payment->business_id,
            'transaction_id' => $payment->transaction_id,
            'method' => $payment->method,
            'amount' => (int) $payment->amount,
            'status' => $payment->status,
            'is_cash' => (bool) $payment->is_cash,
            'confirmed_by' => $payment->confirmed_by,
            'confirmed_at' => $payment->confirmed_at,
            'confirm_expires_at' => $payment->confirm_expires_at,
            'failure_reason' => $payment->failure_reason,
        ];
    }

    private function failLocked(Payment $payment, object $transaction, string $status): Payment
    {
        $before = $this->paymentSnapshot($payment);
        $payment->update(['status' => $status, 'failed_at' => now(), 'failure_reason' => $status]);
        DB::table('transactions')->where('business_id', $payment->business_id)->where('id', $transaction->id)->update(['status' => 'payment_failed', 'updated_at' => now()]);
        Nojpos::audit($payment->business_id, null, $status === 'expired' ? 'payment.expire' : 'payment.fail', 'payment', $payment->id, $before, $this->paymentSnapshot($payment->refresh()));

        return $payment->refresh();
    }

    private function applyStockLocked(object $transaction): void
    {
        foreach (DB::table('transaction_items')->where('business_id', $transaction->business_id)->where('transaction_id', $transaction->id)->get() as $item) {
            $product = DB::table('products')->where('business_id', $transaction->business_id)->where('id', $item->product_id)->lockForUpdate()->first();
            if ($product && (bool) $product->track_stock && ! DB::table('stock_movements')->where('transaction_id', $transaction->id)->where('product_id', $item->product_id)->where('type', 'sale')->exists()) {
                DB::table('stock_movements')->insert(['id' => (string) Str::uuid(), 'business_id' => $transaction->business_id, 'outlet_id' => $transaction->outlet_id, 'product_id' => $item->product_id, 'transaction_id' => $transaction->id, 'type' => 'sale', 'quantity_delta' => -$item->quantity, 'created_at' => now(), 'updated_at' => now()]);
            }
        }
    }
}
