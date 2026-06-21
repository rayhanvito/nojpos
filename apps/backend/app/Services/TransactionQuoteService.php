<?php

namespace App\Services;

use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;
use RuntimeException;

class TransactionQuoteService
{
    private const QUOTE_TTL_MINUTES = 10;

    /**
     * @return array<string, mixed>
     */
    public function quote(string $businessId, array $data): array
    {
        return $this->calculate($businessId, $data);
    }

    public function createSnapshot(string $businessId, array $data): array
    {
        $quote = $this->calculate($businessId, $data);
        $quoteId = (string) Str::uuid();
        $checkoutKey = (string) Str::uuid();
        $checkoutToken = Str::random(64);
        $expiresAt = now()->addMinutes(self::QUOTE_TTL_MINUTES);
        $quoteHash = $this->hashSnapshot($quote);

        DB::table('transaction_quotes')->insert([
            'id' => $quoteId,
            'business_id' => $businessId,
            'outlet_id' => $data['outlet_id'],
            'device_id' => $data['device_id'],
            'cashier_id' => $data['cashier_id'],
            'shift_id' => $data['shift_id'],
            'customer_id' => $data['customer_id'] ?? null,
            'served_by' => $data['served_by'] ?? null,
            'config_version' => $quote['config_version'],
            'quote_hash' => $quoteHash,
            'checkout_idempotency_key' => $checkoutKey,
            'checkout_token' => hash('sha256', $checkoutToken),
            'input_snapshot' => json_encode($this->canonicalInput($data), JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE),
            'quote_snapshot' => json_encode($this->publicSnapshot($quote), JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE),
            'expires_at' => $expiresAt,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return array_merge($this->publicSnapshot($quote), [
            'quote_id' => $quoteId,
            'quote_hash' => $quoteHash,
            'quote_revision' => $quote['config_version'],
            'expires_at' => $expiresAt->toISOString(),
            'server_time' => now()->toISOString(),
            'checkout_idempotency_key' => $checkoutKey,
            'checkout_token' => $checkoutToken,
            'quote_token' => $checkoutToken,
        ]);
    }

    public function quoteForCheckout(string $businessId, array $data, ?string $idempotencyKey): array
    {
        if (empty($data['quote_id'])) {
            if ((bool) config('nojpos.checkout.require_quote_for_checkout', true)) {
                throw new RuntimeException('QUOTE_REQUIRED');
            }

            return $this->calculate($businessId, $data);
        }

        $quote = DB::table('transaction_quotes')
            ->where('business_id', $businessId)
            ->where('id', $data['quote_id'])
            ->lockForUpdate()
            ->first();

        if (! $quote) {
            throw new RuntimeException('QUOTE_STALE');
        }

        if (! hash_equals((string) $quote->checkout_token, hash('sha256', (string) ($data['checkout_token'] ?? $data['quote_token'] ?? '')))) {
            throw new RuntimeException('QUOTE_STALE');
        }

        if ($idempotencyKey !== $quote->checkout_idempotency_key) {
            throw new RuntimeException('QUOTE_STALE');
        }

        if (now()->greaterThanOrEqualTo($quote->expires_at)) {
            throw new RuntimeException('QUOTE_STALE');
        }

        if (! empty($quote->used_transaction_id)) {
            throw new RuntimeException('QUOTE_STALE');
        }

        foreach (['outlet_id', 'device_id', 'cashier_id', 'shift_id'] as $field) {
            if ((string) ($data[$field] ?? '') !== (string) $quote->{$field}) {
                throw new RuntimeException('QUOTE_STALE');
            }
        }

        foreach (['customer_id', 'served_by'] as $field) {
            if (($data[$field] ?? null) !== ($quote->{$field} ?? null)) {
                throw new RuntimeException('QUOTE_STALE');
            }
        }

        $snapshot = json_decode((string) $quote->quote_snapshot, true);
        if (! is_array($snapshot)) {
            throw new RuntimeException('QUOTE_STALE');
        }

        $currentVersion = $this->configVersion($businessId, $this->canonicalInput($data));
        if (! hash_equals((string) $quote->config_version, $currentVersion)) {
            throw new RuntimeException('QUOTE_STALE');
        }

        $snapshot['items'] = array_map(function (array $item): array {
            return [
                'product' => (object) [
                    'id' => $item['product_id'],
                    'name' => $item['name'],
                    'track_stock' => (bool) ($item['track_stock'] ?? false),
                ],
                'product_id' => $item['product_id'],
                'name' => $item['name'],
                'quantity' => (int) $item['quantity'],
                'unit_price' => (int) $item['unit_price'],
                'discount' => (int) $item['discount'],
                'manual_discount' => (int) ($item['manual_discount'] ?? $item['discount'] ?? 0),
                'promo_discount_total' => (int) ($item['promo_discount_total'] ?? 0),
                'subtotal' => (int) $item['subtotal'],
                'gross_line_total' => (int) ($item['gross_line_total'] ?? $item['subtotal']),
                'net_line_total' => (int) ($item['net_line_total'] ?? max(0, (int) $item['subtotal'] - (int) $item['discount'])),
                'line_total' => (int) ($item['line_total'] ?? max(0, (int) $item['subtotal'] - (int) $item['discount'])),
                'track_stock' => (bool) ($item['track_stock'] ?? false),
            ];
        }, $snapshot['items'] ?? []);

        return $snapshot;
    }

    public function markUsed(string $businessId, string $quoteId, string $transactionId): void
    {
        DB::table('transaction_quotes')
            ->where('business_id', $businessId)
            ->where('id', $quoteId)
            ->update([
                'used_transaction_id' => $transactionId,
                'updated_at' => now(),
            ]);
    }

    private function calculate(string $businessId, array $data): array
    {
        $outlet = DB::table('outlets')
            ->where('id', $data['outlet_id'])
            ->where('business_id', $businessId)
            ->first();

        if (! $outlet) {
            throw new RuntimeException('OUTLET_NOT_FOUND');
        }

        $subtotal = 0;
        $itemDiscountTotal = 0;
        $items = [];

        foreach ($data['items'] as $item) {
            $product = DB::table('products')
                ->leftJoin('product_categories', 'product_categories.id', '=', 'products.product_category_id')
                ->where('products.id', $item['product_id'])
                ->where('products.business_id', $businessId)
                ->select('products.*', 'product_categories.name as category_name')
                ->first();

            if (! $product) {
                throw new RuntimeException('PRODUCT_NOT_FOUND');
            }

            $quantity = (int) $item['quantity'];
            $unitPrice = (int) $product->price;
            $lineSubtotal = $quantity * $unitPrice;
            $manualDiscount = $this->discountAmount(
                value: (int) ($item['discount'] ?? 0),
                type: $item['discount_type'] ?? 'amount',
                base: $lineSubtotal,
            );

            $subtotal += $lineSubtotal;
            $itemDiscountTotal += $manualDiscount;

            $items[] = [
                'product' => $product,
                'product_id' => $product->id,
                'product_category_id' => $product->product_category_id,
                'name' => $product->name,
                'category_name' => $product->category_name,
                'quantity' => $quantity,
                'unit_price' => $unitPrice,
                'manual_discount' => $manualDiscount,
                'promo_discount_total' => 0,
                'discount' => $manualDiscount,
                'subtotal' => $lineSubtotal,
                'gross_line_total' => $lineSubtotal,
                'track_stock' => (bool) $product->track_stock,
            ];
        }

        $manualDiscount = $this->manualDiscount($data, max(0, $subtotal - $itemDiscountTotal));
        $cartDiscount = $manualDiscount['amount'];
        $promotionResult = $this->applyPromotions($businessId, (string) $outlet->id, $data, $items, max(0, $subtotal - $itemDiscountTotal - $cartDiscount));
        $items = $promotionResult['items'];
        $promotionDiscountTotal = $promotionResult['promotion_discount_total'];
        $manualDiscountTotal = $itemDiscountTotal + $cartDiscount;
        $discountTotal = $manualDiscountTotal + $promotionDiscountTotal;
        $taxableBase = max(0, $subtotal - $discountTotal);
        $serviceCharge = intdiv($taxableBase * (int) $outlet->service_charge_rate, 100);
        $tax = intdiv(($taxableBase + $serviceCharge) * (int) $outlet->tax_rate, 100);
        $rounding = 0;
        $grandTotal = $taxableBase + $serviceCharge + $tax + $rounding;

        $items = array_map(function (array $item): array {
            $discount = (int) $item['manual_discount'] + (int) $item['promo_discount_total'];
            $netLineTotal = max(0, (int) $item['subtotal'] - $discount);

            return array_merge($item, [
                'discount' => $discount,
                'discount_total' => $discount,
                'net_line_total' => $netLineTotal,
                'tax_total' => 0,
                'service_total' => 0,
                'line_total' => $netLineTotal,
            ]);
        }, $items);

        return [
            'config_version' => $this->configVersion($businessId, $data),
            'subtotal' => $subtotal,
            'item_discount_total' => $itemDiscountTotal,
            'cart_discount_total' => $cartDiscount,
            'promotion_discount_total' => $promotionDiscountTotal,
            'manual_discount_total' => $manualDiscountTotal,
            'discount_total' => $discountTotal,
            'service_charge_total' => $serviceCharge,
            'tax_total' => $tax,
            'rounding_total' => $rounding,
            'grand_total' => $grandTotal,
            'items' => $items,
            'applied_promotions' => $promotionResult['applied_promotions'],
            'rejected_promotions' => $promotionResult['rejected_promotions'],
            'warnings' => $promotionResult['warnings'],
            'meta' => [
                'stacking_policy' => 'no_stacking_unless_stackable',
            ],
        ];
    }

    /**
     * @param  array<int, array<string, mixed>>  $items
     * @return array{items:array<int, array<string, mixed>>, promotion_discount_total:int, applied_promotions:array<int, array<string, mixed>>, rejected_promotions:array<int, array<string, mixed>>, warnings:array<int, string>}
     */
    private function applyPromotions(string $businessId, string $outletId, array $data, array $items, int $promotionBase): array
    {
        $requested = $this->requestedPromotions($data);
        $applied = [];
        $rejected = [];
        $warnings = [];
        $promotionDiscountTotal = 0;
        $hasNonStackablePromotion = false;

        foreach ($requested as $request) {
            $promotion = $this->promotionForRequest($businessId, $request);
            $requestLabel = $request['code'] ?? $request['id'] ?? null;

            if (! $promotion) {
                $rejected[] = $this->rejectedPromotion($requestLabel, 'NOT_FOUND', 'Promosi tidak ditemukan.');

                continue;
            }

            $eligibility = $this->promotionEligibility($businessId, $outletId, $promotion, $promotionBase);
            if (! $eligibility['eligible']) {
                $rejected[] = $this->rejectedPromotion($requestLabel, $eligibility['reason_code'], $eligibility['message'], $promotion);

                continue;
            }

            if ($hasNonStackablePromotion || (! empty($applied) && ! (bool) $promotion->stackable)) {
                $rejected[] = $this->rejectedPromotion($requestLabel, 'STACKING_NOT_ALLOWED', 'Promo tidak dapat digabung dengan promo lain.', $promotion);

                continue;
            }

            $application = $this->applyPromotionToItems($promotion, $items);
            if ($application['discount_amount'] <= 0) {
                $rejected[] = $this->rejectedPromotion($requestLabel, 'NO_ELIGIBLE_ITEMS', 'Tidak ada item yang memenuhi syarat promo.', $promotion);

                continue;
            }

            $items = $application['items'];
            $promotionDiscountTotal += $application['discount_amount'];
            $applied[] = [
                'promotion_id' => $promotion->id,
                'name' => $promotion->name,
                'code' => $promotion->code,
                'type' => $promotion->type,
                'discount_amount' => $application['discount_amount'],
                'affected_items' => $application['affected_items'],
            ];
            $hasNonStackablePromotion = ! (bool) $promotion->stackable;
        }

        return [
            'items' => $items,
            'promotion_discount_total' => $promotionDiscountTotal,
            'applied_promotions' => $applied,
            'rejected_promotions' => $rejected,
            'warnings' => $warnings,
        ];
    }

    /**
     * @return array<int, array{id?:string, code?:string}>
     */
    private function requestedPromotions(array $data): array
    {
        $requested = [];

        foreach ($data['promotion_codes'] ?? [] as $code) {
            $code = trim((string) $code);
            if ($code !== '') {
                $requested[] = ['code' => $code];
            }
        }

        foreach ($data['applied_promotion_ids'] ?? [] as $id) {
            $requested[] = ['id' => (string) $id];
        }

        return $requested;
    }

    private function promotionForRequest(string $businessId, array $request): ?object
    {
        return DB::table('promotions')
            ->where('business_id', $businessId)
            ->when(isset($request['code']), fn ($query) => $query->where('code', $request['code']))
            ->when(isset($request['id']), fn ($query) => $query->where('id', $request['id']))
            ->whereNull('deleted_at')
            ->first();
    }

    /**
     * @return array{eligible:bool, reason_code:string|null, message:string|null}
     */
    private function promotionEligibility(string $businessId, string $outletId, object $promotion, int $promotionBase): array
    {
        $now = now();
        if (! (bool) $promotion->is_active || ($promotion->starts_at && $now->lessThan($promotion->starts_at)) || ($promotion->ends_at && $now->greaterThanOrEqualTo($promotion->ends_at))) {
            return ['eligible' => false, 'reason_code' => 'INACTIVE_OR_EXPIRED', 'message' => 'Promo tidak aktif atau sudah kedaluwarsa.'];
        }

        $hasOutletRules = DB::table('promotion_outlets')
            ->where('business_id', $businessId)
            ->where('promotion_id', $promotion->id)
            ->exists();
        if ($hasOutletRules && ! DB::table('promotion_outlets')->where('business_id', $businessId)->where('promotion_id', $promotion->id)->where('outlet_id', $outletId)->exists()) {
            return ['eligible' => false, 'reason_code' => 'OUTLET_NOT_ELIGIBLE', 'message' => 'Promo tidak berlaku di outlet ini.'];
        }

        if ($promotionBase < (int) $promotion->minimum_subtotal) {
            return ['eligible' => false, 'reason_code' => 'MINIMUM_SPEND_NOT_MET', 'message' => 'Minimum belanja promo belum terpenuhi.'];
        }

        return ['eligible' => true, 'reason_code' => null, 'message' => null];
    }

    /**
     * @param  array<int, array<string, mixed>>  $items
     * @return array{items:array<int, array<string, mixed>>, discount_amount:int, affected_items:array<int, array<string, int|string>>}
     */
    private function applyPromotionToItems(object $promotion, array $items): array
    {
        $eligibleIndexes = [];
        $eligibleBase = 0;

        foreach ($items as $index => $item) {
            if ($promotion->product_id && (string) $item['product_id'] !== (string) $promotion->product_id) {
                continue;
            }

            if ($promotion->product_category_id && (string) ($item['product_category_id'] ?? '') !== (string) $promotion->product_category_id) {
                continue;
            }

            $lineBase = max(0, (int) $item['subtotal'] - (int) $item['manual_discount'] - (int) $item['promo_discount_total']);
            if ($lineBase <= 0) {
                continue;
            }

            $eligibleIndexes[] = ['index' => $index, 'base' => $lineBase];
            $eligibleBase += $lineBase;
        }

        if ($eligibleBase <= 0) {
            return ['items' => $items, 'discount_amount' => 0, 'affected_items' => []];
        }

        $discountAmount = $this->discountAmount((int) $promotion->value, (string) $promotion->type, $eligibleBase);
        $remaining = $discountAmount;
        $affected = [];
        $lastIndex = array_key_last($eligibleIndexes);

        foreach ($eligibleIndexes as $position => $eligible) {
            $lineDiscount = $position === $lastIndex
                ? $remaining
                : intdiv($discountAmount * $eligible['base'], $eligibleBase);
            $remaining -= $lineDiscount;
            $items[$eligible['index']]['promo_discount_total'] = (int) $items[$eligible['index']]['promo_discount_total'] + $lineDiscount;
            $affected[] = [
                'product_id' => $items[$eligible['index']]['product_id'],
                'discount_amount' => $lineDiscount,
            ];
        }

        return ['items' => $items, 'discount_amount' => $discountAmount, 'affected_items' => $affected];
    }

    private function rejectedPromotion(?string $requestLabel, string $reasonCode, string $message, ?object $promotion = null): array
    {
        return [
            'code' => $requestLabel,
            'promotion_id' => $promotion?->id,
            'name' => $promotion?->name,
            'reason_code' => $reasonCode,
            'message' => $message,
        ];
    }

    private function manualDiscount(array $data, int $base): array
    {
        if (isset($data['manual_discount']) && is_array($data['manual_discount'])) {
            return [
                'type' => $data['manual_discount']['type'] ?? 'amount',
                'amount' => $this->discountAmount(
                    value: (int) ($data['manual_discount']['value'] ?? 0),
                    type: $data['manual_discount']['type'] ?? 'amount',
                    base: $base,
                ),
                'reason' => $data['manual_discount']['reason'] ?? null,
            ];
        }

        return [
            'type' => $data['cart_discount_type'] ?? 'amount',
            'amount' => $this->discountAmount(
                value: (int) ($data['cart_discount'] ?? 0),
                type: $data['cart_discount_type'] ?? 'amount',
                base: $base,
            ),
            'reason' => null,
        ];
    }

    private function publicSnapshot(array $quote): array
    {
        $snapshot = $quote;
        $snapshot['items'] = array_map(fn (array $item): array => [
            'product_id' => $item['product_id'] ?? $item['product']->id,
            'name' => $item['name'] ?? $item['product']->name,
            'quantity' => (int) $item['quantity'],
            'unit_price' => (int) $item['unit_price'],
            'gross_line_total' => (int) ($item['gross_line_total'] ?? $item['subtotal']),
            'manual_discount' => (int) ($item['manual_discount'] ?? $item['discount'] ?? 0),
            'discount' => (int) $item['discount'],
            'discount_total' => (int) ($item['discount_total'] ?? $item['discount']),
            'promo_discount_total' => (int) ($item['promo_discount_total'] ?? 0),
            'net_line_total' => (int) ($item['net_line_total'] ?? max(0, (int) $item['subtotal'] - (int) $item['discount'])),
            'tax_total' => (int) ($item['tax_total'] ?? 0),
            'service_total' => (int) ($item['service_total'] ?? 0),
            'line_total' => (int) ($item['line_total'] ?? max(0, (int) $item['subtotal'] - (int) $item['discount'])),
            'subtotal' => (int) $item['subtotal'],
            'track_stock' => (bool) ($item['track_stock'] ?? $item['product']->track_stock ?? false),
        ], $quote['items'] ?? []);
        $snapshot['totals'] = [
            'subtotal' => (int) $quote['subtotal'],
            'discount_total' => (int) $quote['discount_total'],
            'promotion_discount_total' => (int) ($quote['promotion_discount_total'] ?? 0),
            'manual_discount_total' => (int) ($quote['manual_discount_total'] ?? (($quote['item_discount_total'] ?? 0) + ($quote['cart_discount_total'] ?? 0))),
            'tax_total' => (int) $quote['tax_total'],
            'service_total' => (int) $quote['service_charge_total'],
            'service_charge_total' => (int) $quote['service_charge_total'],
            'rounding_total' => (int) $quote['rounding_total'],
            'grand_total' => (int) $quote['grand_total'],
        ];

        return $snapshot;
    }

    private function canonicalInput(array $data): array
    {
        $input = [
            'outlet_id' => $data['outlet_id'],
            'device_id' => $data['device_id'],
            'cashier_id' => $data['cashier_id'],
            'shift_id' => $data['shift_id'],
            'customer_id' => $data['customer_id'] ?? null,
            'served_by' => $data['served_by'] ?? null,
            'cart_discount' => (int) ($data['cart_discount'] ?? $data['manual_discount']['value'] ?? 0),
            'cart_discount_type' => $data['cart_discount_type'] ?? $data['manual_discount']['type'] ?? 'amount',
            'manual_discount' => isset($data['manual_discount']) && is_array($data['manual_discount']) ? [
                'type' => $data['manual_discount']['type'] ?? 'amount',
                'value' => (int) ($data['manual_discount']['value'] ?? 0),
                'reason' => $data['manual_discount']['reason'] ?? null,
            ] : null,
            'promotion_codes' => array_values(array_map('strval', $data['promotion_codes'] ?? [])),
            'applied_promotion_ids' => array_values(array_map('strval', $data['applied_promotion_ids'] ?? [])),
            'items' => array_map(fn (array $item): array => [
                'product_id' => $item['product_id'],
                'quantity' => (int) $item['quantity'],
                'discount' => (int) ($item['discount'] ?? 0),
                'discount_type' => $item['discount_type'] ?? 'amount',
            ], $data['items'] ?? []),
        ];

        $this->sortRecursive($input);

        return $input;
    }

    private function configVersion(string $businessId, array $data): string
    {
        $outlet = DB::table('outlets')
            ->where('id', $data['outlet_id'])
            ->where('business_id', $businessId)
            ->first();

        if (! $outlet) {
            throw new RuntimeException('OUTLET_NOT_FOUND');
        }

        $productVersions = [];
        foreach ($data['items'] ?? [] as $item) {
            $product = DB::table('products')
                ->where('id', $item['product_id'])
                ->where('business_id', $businessId)
                ->first();

            if (! $product) {
                throw new RuntimeException('PRODUCT_NOT_FOUND');
            }

            $productVersions[] = [
                'id' => $product->id,
                'category_id' => $product->product_category_id,
                'price' => (int) $product->price,
                'track_stock' => (bool) $product->track_stock,
                'updated_at' => (string) $product->updated_at,
            ];
        }

        $promotionVersions = [];
        foreach ($this->requestedPromotions($data) as $request) {
            $promotion = $this->promotionForRequest($businessId, $request);
            $promotionVersions[] = $promotion ? [
                'id' => $promotion->id,
                'code' => $promotion->code,
                'type' => $promotion->type,
                'value' => (int) $promotion->value,
                'minimum_subtotal' => (int) $promotion->minimum_subtotal,
                'product_id' => $promotion->product_id,
                'product_category_id' => $promotion->product_category_id,
                'is_active' => (bool) $promotion->is_active,
                'is_currently_active' => $this->isPromotionCurrentlyActive($promotion),
                'stackable' => (bool) $promotion->stackable,
                'starts_at' => (string) $promotion->starts_at,
                'ends_at' => (string) $promotion->ends_at,
                'updated_at' => (string) $promotion->updated_at,
                'outlets' => DB::table('promotion_outlets')
                    ->where('business_id', $businessId)
                    ->where('promotion_id', $promotion->id)
                    ->orderBy('outlet_id')
                    ->pluck('outlet_id')
                    ->all(),
            ] : ['missing' => $request];
        }

        return hash('sha256', json_encode([
            'outlet' => [
                'id' => $outlet->id,
                'service_charge_rate' => (int) $outlet->service_charge_rate,
                'tax_rate' => (int) $outlet->tax_rate,
                'rounding_policy' => 'none',
                'updated_at' => (string) $outlet->updated_at,
            ],
            'products' => $productVersions,
            'promotions' => $promotionVersions,
        ], JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE));
    }

    private function isPromotionCurrentlyActive(object $promotion): bool
    {
        $now = now();

        return (bool) $promotion->is_active
            && (! $promotion->starts_at || $now->greaterThanOrEqualTo($promotion->starts_at))
            && (! $promotion->ends_at || $now->lessThan($promotion->ends_at));
    }

    private function hashSnapshot(array $quote): string
    {
        return hash('sha256', json_encode($this->publicSnapshot($quote), JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE));
    }

    private function discountAmount(int $value, string $type, int $base): int
    {
        if ($value <= 0 || $base <= 0) {
            return 0;
        }

        $amount = $type === 'percent'
            ? intdiv($base * min($value, 100), 100)
            : $value;

        return min($amount, $base);
    }

    private function sortRecursive(array &$value): void
    {
        foreach ($value as &$item) {
            if (is_array($item)) {
                $this->sortRecursive($item);
            }
        }

        if (! array_is_list($value)) {
            ksort($value);
        }
    }
}
