<?php

namespace App\Services;

use App\Support\Nojpos;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class ParkedOrderService
{
    private const LEASE_SECONDS = 90;

    public function __construct(
        private readonly BusinessClock $clock,
        private readonly TransactionQuoteService $quoteService,
    ) {}

    public function create(Request $request, array $data): array
    {
        $businessId = $request->user()->business_id;
        $quote = $this->quoteService->quote($businessId, $data);
        $transactionId = (string) Str::uuid();
        $now = now();

        DB::transaction(function () use ($businessId, $request, $data, $quote, $transactionId, $now): void {
            DB::table('transactions')->insert([
                'id' => $transactionId,
                'business_id' => $businessId,
                'outlet_id' => $data['outlet_id'],
                'device_id' => $data['device_id'],
                'cashier_id' => $data['cashier_id'],
                'shift_id' => $data['shift_id'],
                'customer_id' => $data['customer_id'] ?? null,
                'served_by' => $data['served_by'] ?? null,
                'number' => 'PARK-'.$this->clock->documentTimestamp($this->clock->outletTimezone($businessId, $data['outlet_id']), $now).'-'.Str::upper(Str::random(4)),
                'status' => 'held',
                'revision' => 1,
                'subtotal' => $quote['subtotal'],
                'discount_total' => $quote['discount_total'],
                'item_discount_total' => $quote['item_discount_total'],
                'cart_discount_total' => $quote['cart_discount_total'],
                'promotion_discount_total' => $quote['promotion_discount_total'] ?? 0,
                'service_charge_total' => $quote['service_charge_total'],
                'tax_total' => $quote['tax_total'],
                'rounding_total' => $quote['rounding_total'],
                'grand_total' => $quote['grand_total'],
                'notes' => $data['notes'] ?? null,
                'created_at' => $now,
                'updated_at' => $now,
            ]);

            $this->replaceItems($businessId, $transactionId, $quote['items'], $now);

            Nojpos::audit($businessId, $request->user()->id, 'parked_order.create', 'transaction', $transactionId, null, [
                'status' => 'held',
                'revision' => 1,
            ]);
        });

        return $this->payload($businessId, $transactionId);
    }

    public function update(Request $request, string $transactionId, array $data): array
    {
        $businessId = $request->user()->business_id;

        return DB::transaction(function () use ($businessId, $request, $transactionId, $data): array {
            $transaction = $this->lockedParkedOrder($businessId, $transactionId);
            $this->assertRevision($transaction, (int) $data['expected_revision']);
            $this->assertLeaseAllows($transaction, $data['device_id'], $request->user()->id);

            $quoteData = array_merge($data, [
                'outlet_id' => $transaction->outlet_id,
                'device_id' => $transaction->device_id,
                'cashier_id' => $transaction->cashier_id,
                'shift_id' => $transaction->shift_id,
            ]);
            $quote = $this->quoteService->quote($businessId, $quoteData);
            $now = now();
            $nextRevision = (int) $transaction->revision + 1;

            DB::table('transactions')
                ->where('id', $transactionId)
                ->where('business_id', $businessId)
                ->update([
                    'customer_id' => $data['customer_id'] ?? $transaction->customer_id,
                    'served_by' => $data['served_by'] ?? $transaction->served_by,
                    'revision' => $nextRevision,
                    'subtotal' => $quote['subtotal'],
                    'discount_total' => $quote['discount_total'],
                    'item_discount_total' => $quote['item_discount_total'],
                    'cart_discount_total' => $quote['cart_discount_total'],
                    'promotion_discount_total' => $quote['promotion_discount_total'] ?? 0,
                    'service_charge_total' => $quote['service_charge_total'],
                    'tax_total' => $quote['tax_total'],
                    'rounding_total' => $quote['rounding_total'],
                    'grand_total' => $quote['grand_total'],
                    'notes' => $data['notes'] ?? null,
                    'updated_at' => $now,
                ]);

            $this->replaceItems($businessId, $transactionId, $quote['items'], $now);

            Nojpos::audit($businessId, $request->user()->id, 'parked_order.update', 'transaction', $transactionId, [
                'revision' => (int) $transaction->revision,
            ], [
                'revision' => $nextRevision,
            ]);

            return $this->payload($businessId, $transactionId);
        });
    }

    public function acquireLease(Request $request, string $transactionId, array $data): array
    {
        return $this->lease($request, $transactionId, $data, 'parked_order.lease_acquire');
    }

    public function refreshLease(Request $request, string $transactionId, array $data): array
    {
        return $this->lease($request, $transactionId, $data, 'parked_order.lease_refresh', requireCurrentOwner: true);
    }

    public function releaseLease(Request $request, string $transactionId, array $data): array
    {
        $businessId = $request->user()->business_id;

        return DB::transaction(function () use ($businessId, $request, $transactionId, $data): array {
            $transaction = $this->lockedParkedOrder($businessId, $transactionId);
            $this->assertLeaseOwnedBy($transaction, $data['device_id'], $request->user()->id);

            DB::table('transactions')
                ->where('business_id', $businessId)
                ->where('id', $transactionId)
                ->update([
                    'lease_device_id' => null,
                    'leased_by_user_id' => null,
                    'lease_expires_at' => null,
                    'updated_at' => now(),
                ]);

            Nojpos::audit($businessId, $request->user()->id, 'parked_order.lease_release', 'transaction', $transactionId, $this->leaseSnapshot($transaction), null);

            return $this->payload($businessId, $transactionId);
        });
    }

    public function cancel(Request $request, string $transactionId, array $data): array
    {
        $businessId = $request->user()->business_id;

        return DB::transaction(function () use ($businessId, $request, $transactionId, $data): array {
            $transaction = $this->lockedParkedOrder($businessId, $transactionId);
            $this->assertRevision($transaction, (int) $data['expected_revision']);
            $this->assertLeaseAllows($transaction, $data['device_id'], $request->user()->id);

            DB::table('transactions')
                ->where('business_id', $businessId)
                ->where('id', $transactionId)
                ->update([
                    'status' => 'voided',
                    'lease_device_id' => null,
                    'leased_by_user_id' => null,
                    'lease_expires_at' => null,
                    'updated_at' => now(),
                ]);

            Nojpos::audit($businessId, $request->user()->id, 'parked_order.cancel', 'transaction', $transactionId, [
                'status' => $transaction->status,
                'revision' => (int) $transaction->revision,
            ], [
                'status' => 'voided',
            ]);

            return $this->payload($businessId, $transactionId);
        });
    }

    private function lease(Request $request, string $transactionId, array $data, string $auditAction, bool $requireCurrentOwner = false): array
    {
        $businessId = $request->user()->business_id;

        return DB::transaction(function () use ($businessId, $request, $transactionId, $data, $auditAction, $requireCurrentOwner): array {
            $transaction = $this->lockedParkedOrder($businessId, $transactionId);

            if ($requireCurrentOwner) {
                $this->assertLeaseOwnedBy($transaction, $data['device_id'], $request->user()->id);
            } else {
                $this->assertLeaseAllows($transaction, $data['device_id'], $request->user()->id, allowEmpty: true);
            }

            $expiresAt = now()->addSeconds(self::LEASE_SECONDS);
            DB::table('transactions')
                ->where('business_id', $businessId)
                ->where('id', $transactionId)
                ->update([
                    'lease_device_id' => $data['device_id'],
                    'leased_by_user_id' => $request->user()->id,
                    'lease_expires_at' => $expiresAt,
                    'updated_at' => now(),
                ]);

            Nojpos::audit($businessId, $request->user()->id, $auditAction, 'transaction', $transactionId, $this->leaseSnapshot($transaction), [
                'lease_device_id' => $data['device_id'],
                'leased_by_user_id' => $request->user()->id,
                'lease_expires_at' => $expiresAt->toISOString(),
            ]);

            return $this->payload($businessId, $transactionId);
        });
    }

    private function lockedParkedOrder(string $businessId, string $transactionId): object
    {
        $transaction = DB::table('transactions')
            ->where('id', $transactionId)
            ->where('business_id', $businessId)
            ->where('status', 'held')
            ->whereNull('deleted_at')
            ->lockForUpdate()
            ->first();

        if (! $transaction) {
            throw new ParkedOrderException('NOT_FOUND', 'Parked order not found.', 404);
        }

        return $transaction;
    }

    private function assertRevision(object $transaction, int $expectedRevision): void
    {
        if ((int) $transaction->revision !== $expectedRevision) {
            throw new ParkedOrderException('CONFLICT_REVISION', 'Parked order has changed. Please reload it before saving.', 409, [
                'current_revision' => (int) $transaction->revision,
            ]);
        }
    }

    private function assertLeaseAllows(object $transaction, string $deviceId, string $userId, bool $allowEmpty = false): void
    {
        if ($this->leaseIsExpired($transaction) || empty($transaction->lease_device_id)) {
            if ($allowEmpty) {
                return;
            }

            throw new ParkedOrderException('ORDER_LOCK_REQUIRED', 'Acquire a lease before editing this parked order.', 409);
        }

        if ($transaction->lease_device_id === $deviceId && $transaction->leased_by_user_id === $userId) {
            return;
        }

        throw new ParkedOrderException('ORDER_LOCKED', 'Parked order is locked by another terminal.', 409, $this->lockDetails($transaction));
    }

    private function assertLeaseOwnedBy(object $transaction, string $deviceId, string $userId): void
    {
        if ($this->leaseIsExpired($transaction) || empty($transaction->lease_device_id)) {
            throw new ParkedOrderException('ORDER_LOCK_REQUIRED', 'Acquire a lease before this action.', 409);
        }

        if ($transaction->lease_device_id !== $deviceId || $transaction->leased_by_user_id !== $userId) {
            throw new ParkedOrderException('ORDER_LOCKED', 'Parked order is locked by another terminal.', 409, $this->lockDetails($transaction));
        }
    }

    private function leaseIsExpired(object $transaction): bool
    {
        return empty($transaction->lease_expires_at) || now()->greaterThanOrEqualTo($transaction->lease_expires_at);
    }

    private function lockDetails(object $transaction): array
    {
        $remaining = empty($transaction->lease_expires_at)
            ? 0
            : max(0, now()->diffInSeconds($transaction->lease_expires_at, false));

        return [
            'lease_device_id' => $transaction->lease_device_id,
            'leased_by_user_id' => $transaction->leased_by_user_id,
            'lease_expires_at' => $transaction->lease_expires_at,
            'remaining_seconds' => $remaining,
        ];
    }

    private function leaseSnapshot(object $transaction): array
    {
        return [
            'lease_device_id' => $transaction->lease_device_id,
            'leased_by_user_id' => $transaction->leased_by_user_id,
            'lease_expires_at' => $transaction->lease_expires_at,
        ];
    }

    private function replaceItems(string $businessId, string $transactionId, array $items, mixed $now): void
    {
        DB::table('transaction_items')
            ->where('business_id', $businessId)
            ->where('transaction_id', $transactionId)
            ->delete();

        foreach ($items as $item) {
            DB::table('transaction_items')->insert([
                'id' => (string) Str::uuid(),
                'business_id' => $businessId,
                'transaction_id' => $transactionId,
                'product_id' => $item['product']->id,
                'name' => $item['product']->name,
                'quantity' => $item['quantity'],
                'unit_price' => $item['unit_price'],
                'discount' => $item['discount'],
                'subtotal' => $item['subtotal'],
                'created_at' => $now,
                'updated_at' => $now,
            ]);
        }
    }

    public function payload(string $businessId, string $transactionId): array
    {
        $transaction = (array) DB::table('transactions')
            ->where('business_id', $businessId)
            ->where('id', $transactionId)
            ->first();

        $items = DB::table('transaction_items')
            ->where('business_id', $businessId)
            ->where('transaction_id', $transactionId)
            ->orderBy('created_at')
            ->get()
            ->map(fn (object $item): array => [
                'id' => $item->id,
                'product_id' => $item->product_id,
                'name' => $item->name,
                'quantity' => (int) $item->quantity,
                'unit_price' => (int) $item->unit_price,
                'discount' => (int) $item->discount,
                'subtotal' => (int) $item->subtotal,
            ])
            ->values()
            ->all();

        return [
            'id' => $transaction['id'],
            'business_id' => $transaction['business_id'],
            'outlet_id' => $transaction['outlet_id'],
            'device_id' => $transaction['device_id'],
            'cashier_id' => $transaction['cashier_id'],
            'shift_id' => $transaction['shift_id'],
            'customer_id' => $transaction['customer_id'] ?? null,
            'served_by' => $transaction['served_by'] ?? null,
            'number' => $transaction['number'],
            'status' => $transaction['status'],
            'revision' => (int) ($transaction['revision'] ?? 1),
            'lease' => [
                'device_id' => $transaction['lease_device_id'] ?? null,
                'user_id' => $transaction['leased_by_user_id'] ?? null,
                'expires_at' => $transaction['lease_expires_at'] ?? null,
                'remaining_seconds' => empty($transaction['lease_expires_at']) ? 0 : max(0, now()->diffInSeconds($transaction['lease_expires_at'], false)),
            ],
            'subtotal' => (int) $transaction['subtotal'],
            'discount_total' => (int) $transaction['discount_total'],
            'item_discount_total' => (int) ($transaction['item_discount_total'] ?? 0),
            'cart_discount_total' => (int) ($transaction['cart_discount_total'] ?? 0),
            'promotion_discount_total' => (int) ($transaction['promotion_discount_total'] ?? 0),
            'service_charge_total' => (int) $transaction['service_charge_total'],
            'tax_total' => (int) $transaction['tax_total'],
            'rounding_total' => (int) $transaction['rounding_total'],
            'grand_total' => (int) $transaction['grand_total'],
            'notes' => $transaction['notes'] ?? null,
            'items' => $items,
            'created_at' => $transaction['created_at'] ?? null,
            'updated_at' => $transaction['updated_at'] ?? null,
        ];
    }
}
