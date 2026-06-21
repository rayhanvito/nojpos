<?php

namespace Tests\Feature;

use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class PromotionQuoteTest extends TestCase
{
    use RefreshDatabase;

    public function test_quote_without_promo_returns_server_calculated_totals_and_token(): void
    {
        $ctx = $this->context(serviceRate: 10, taxRate: 10);

        $response = $this->withToken($ctx['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, [
            ['product_id' => $ctx['product'], 'quantity' => 2, 'unit_price' => 1, 'discount' => 1000],
        ], [
            'manual_discount' => ['type' => 'amount', 'value' => 500, 'reason' => 'Owner promo'],
        ]));

        $response->assertOk()
            ->assertJsonPath('data.subtotal', 40000)
            ->assertJsonPath('data.item_discount_total', 1000)
            ->assertJsonPath('data.cart_discount_total', 500)
            ->assertJsonPath('data.promotion_discount_total', 0)
            ->assertJsonPath('data.discount_total', 1500)
            ->assertJsonPath('data.service_charge_total', 3850)
            ->assertJsonPath('data.tax_total', 4235)
            ->assertJsonPath('data.rounding_total', 0)
            ->assertJsonPath('data.grand_total', 46585)
            ->assertJsonPath('data.totals.manual_discount_total', 1500)
            ->assertJsonStructure(['data' => ['quote_id', 'quote_hash', 'quote_revision', 'checkout_idempotency_key', 'checkout_token', 'quote_token', 'server_time']]);

        $this->assertSame(0, DB::table('transactions')->count());
        $this->assertSame(0, DB::table('payments')->count());
        $this->assertSame(0, DB::table('stock_movements')->count());
    }

    public function test_percentage_and_fixed_promotions_apply_and_stack_safely(): void
    {
        $ctx = $this->context();
        $percent = $this->promotion($ctx['business'], 'Diskon 10%', 'PCT10', 'percent', 10, stackable: true);
        $fixed = $this->promotion($ctx['business'], 'Potongan 5K', 'FIX5', 'amount', 5000, stackable: true);
        $blocked = $this->promotion($ctx['business'], 'Tidak Stack', 'NOSTACK', 'amount', 1000);

        $this->withToken($ctx['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, [
            ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1],
        ], [
            'promotion_codes' => ['PCT10', 'FIX5', 'NOSTACK'],
        ]))->assertOk()
            ->assertJsonPath('data.promotion_discount_total', 7000)
            ->assertJsonPath('data.discount_total', 7000)
            ->assertJsonPath('data.grand_total', 13000)
            ->assertJsonPath('data.applied_promotions.0.promotion_id', $percent)
            ->assertJsonPath('data.applied_promotions.0.discount_amount', 2000)
            ->assertJsonPath('data.applied_promotions.1.promotion_id', $fixed)
            ->assertJsonPath('data.applied_promotions.1.discount_amount', 5000)
            ->assertJsonPath('data.rejected_promotions.0.promotion_id', $blocked)
            ->assertJsonPath('data.rejected_promotions.0.reason_code', 'STACKING_NOT_ALLOWED')
            ->assertJsonPath('data.meta.stacking_policy', 'no_stacking_unless_stackable');
    }

    public function test_inactive_expired_outlet_and_minimum_spend_promotions_are_rejected(): void
    {
        $ctx = $this->context();
        $otherOutlet = $this->outlet($ctx['business'], 'Cabang Lain');
        $inactive = $this->promotion($ctx['business'], 'Inactive', 'OFF', 'amount', 1000, isActive: false);
        $expired = $this->promotion($ctx['business'], 'Expired', 'OLD', 'amount', 1000, endsAt: now()->subMinute());
        $outletOnly = $this->promotion($ctx['business'], 'Outlet Other', 'OTHEROUTLET', 'amount', 1000, outletIds: [$otherOutlet]);
        $minSpend = $this->promotion($ctx['business'], 'Min Spend', 'MIN50', 'amount', 1000, minimumSubtotal: 50000);

        $this->withToken($ctx['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, [
            ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1],
        ], [
            'promotion_codes' => ['OFF', 'OLD', 'OTHEROUTLET', 'MIN50'],
        ]))->assertOk()
            ->assertJsonPath('data.promotion_discount_total', 0)
            ->assertJsonPath('data.rejected_promotions.0.promotion_id', $inactive)
            ->assertJsonPath('data.rejected_promotions.0.reason_code', 'INACTIVE_OR_EXPIRED')
            ->assertJsonPath('data.rejected_promotions.1.promotion_id', $expired)
            ->assertJsonPath('data.rejected_promotions.1.reason_code', 'INACTIVE_OR_EXPIRED')
            ->assertJsonPath('data.rejected_promotions.2.promotion_id', $outletOnly)
            ->assertJsonPath('data.rejected_promotions.2.reason_code', 'OUTLET_NOT_ELIGIBLE')
            ->assertJsonPath('data.rejected_promotions.3.promotion_id', $minSpend)
            ->assertJsonPath('data.rejected_promotions.3.reason_code', 'MINIMUM_SPEND_NOT_MET');
    }

    public function test_product_and_category_scoped_promotions_only_affect_eligible_lines(): void
    {
        $ctx = $this->context();
        $otherProduct = $this->product($ctx['business'], $ctx['outlet'], $ctx['other_category'], 'Teh Botol', 10000);
        $productPromo = $this->promotion($ctx['business'], 'Produk', 'PRODUCT5', 'amount', 5000, productId: $ctx['product'], stackable: true);
        $categoryPromo = $this->promotion($ctx['business'], 'Kategori', 'CAT10', 'percent', 10, categoryId: $ctx['other_category'], stackable: true);

        $this->withToken($ctx['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, [
            ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1],
            ['product_id' => $otherProduct, 'quantity' => 1, 'unit_price' => 1],
        ], [
            'promotion_codes' => ['PRODUCT5', 'CAT10'],
        ]))->assertOk()
            ->assertJsonPath('data.promotion_discount_total', 6000)
            ->assertJsonPath('data.items.0.promo_discount_total', 5000)
            ->assertJsonPath('data.items.1.promo_discount_total', 1000)
            ->assertJsonPath('data.applied_promotions.0.affected_items.0.product_id', $ctx['product'])
            ->assertJsonPath('data.applied_promotions.1.affected_items.0.product_id', $otherProduct)
            ->assertJsonPath('data.applied_promotions.0.promotion_id', $productPromo)
            ->assertJsonPath('data.applied_promotions.1.promotion_id', $categoryPromo);
    }

    public function test_tenant_isolation_prevents_using_another_business_promotion_or_product(): void
    {
        $ctxA = $this->context('Tenant A');
        $ctxB = $this->context('Tenant B');
        $this->promotion($ctxB['business'], 'Tenant B Promo', 'BTENANT', 'amount', 9000);

        $this->withToken($ctxA['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctxA, [
            ['product_id' => $ctxA['product'], 'quantity' => 1, 'unit_price' => 1],
        ], ['promotion_codes' => ['BTENANT']]))->assertOk()
            ->assertJsonPath('data.promotion_discount_total', 0)
            ->assertJsonPath('data.rejected_promotions.0.reason_code', 'NOT_FOUND');

        $this->withToken($ctxA['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctxA, [
            ['product_id' => $ctxB['product'], 'quantity' => 1, 'unit_price' => 1],
        ]))->assertNotFound()
            ->assertJsonPath('error.code', 'PRODUCT_NOT_FOUND');
    }

    public function test_checkout_with_stale_promotion_quote_fails_and_does_not_create_rows(): void
    {
        $ctx = $this->context();
        $this->paymentMethod($ctx['business'], $ctx['outlet']);
        $promo = $this->promotion($ctx['business'], 'Promo', 'PROMO', 'amount', 5000);
        $quote = $this->withToken($ctx['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, [
            ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1],
        ], ['promotion_codes' => ['PROMO']]))->assertOk();

        DB::table('promotions')->where('id', $promo)->update(['value' => 7000, 'updated_at' => now()->addSecond()]);

        $this->withToken($ctx['token'])->withHeader('Idempotency-Key', $quote->json('data.checkout_idempotency_key'))
            ->postJson('/api/v1/transactions', $this->checkoutPayload($ctx, $quote, [
                ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 999999],
            ], ['promotion_codes' => ['PROMO']]))
            ->assertConflict()
            ->assertJsonPath('error.code', 'QUOTE_STALE');

        $this->assertSame(0, DB::table('transactions')->count());
        $this->assertSame(0, DB::table('payments')->count());
        $this->assertSame(0, DB::table('stock_movements')->count());
    }

    public function test_checkout_uses_quote_snapshot_and_ignores_manipulated_client_totals(): void
    {
        $ctx = $this->context();
        $this->paymentMethod($ctx['business'], $ctx['outlet']);
        $this->promotion($ctx['business'], 'Promo', 'PROMO', 'amount', 5000);
        $quote = $this->withToken($ctx['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, [
            ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1],
        ], ['promotion_codes' => ['PROMO']]))->assertOk();

        $response = $this->withToken($ctx['token'])->withHeader('Idempotency-Key', $quote->json('data.checkout_idempotency_key'))
            ->postJson('/api/v1/transactions', $this->checkoutPayload($ctx, $quote, [
                ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1],
            ], [
                'promotion_codes' => ['PROMO'],
                'subtotal' => 1,
                'service_charge_total' => 999999,
                'tax_total' => 999999,
                'rounding_total' => -999999,
                'grand_total' => 1,
                'discount_total' => 0,
            ]));

        $response->assertCreated()
            ->assertJsonPath('data.subtotal', 20000)
            ->assertJsonPath('data.promotion_discount_total', 5000)
            ->assertJsonPath('data.discount_total', 5000)
            ->assertJsonPath('data.service_charge_total', 0)
            ->assertJsonPath('data.tax_total', 0)
            ->assertJsonPath('data.rounding_total', 0)
            ->assertJsonPath('data.grand_total', 15000);

        $this->assertDatabaseHas('transactions', [
            'id' => $response->json('data.id'),
            'grand_total' => 15000,
            'promotion_discount_total' => 5000,
        ]);
    }

    public function test_checkout_uses_quote_items_when_client_tampers_price_quantity_and_discount(): void
    {
        $ctx = $this->context(serviceRate: 10, taxRate: 10);
        $this->paymentMethod($ctx['business'], $ctx['outlet']);

        $quote = $this->withToken($ctx['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, [
            ['product_id' => $ctx['product'], 'quantity' => 2, 'unit_price' => 1, 'discount' => 1000],
        ], [
            'manual_discount' => ['type' => 'amount', 'value' => 500, 'reason' => 'Owner promo'],
        ]))->assertOk()
            ->assertJsonPath('data.grand_total', 46585);

        $response = $this->withToken($ctx['token'])
            ->postJson('/api/v1/transactions', $this->checkoutPayload($ctx, $quote, [
                ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1, 'discount' => 999999],
            ], [
                'manual_discount' => ['type' => 'amount', 'value' => 999999, 'reason' => 'Tamper'],
                'subtotal' => 1,
                'discount_total' => 999999,
                'service_charge_total' => 1,
                'tax_total' => 1,
                'rounding_total' => -999999,
                'grand_total' => 1,
                'payments' => [['method' => 'cash', 'amount' => $quote->json('data.grand_total')]],
            ]), ['Idempotency-Key' => $quote->json('data.checkout_idempotency_key')])
            ->assertCreated()
            ->assertJsonPath('data.subtotal', 40000)
            ->assertJsonPath('data.item_discount_total', 1000)
            ->assertJsonPath('data.cart_discount_total', 500)
            ->assertJsonPath('data.service_charge_total', 3850)
            ->assertJsonPath('data.tax_total', 4235)
            ->assertJsonPath('data.rounding_total', 0)
            ->assertJsonPath('data.grand_total', 46585)
            ->assertJsonPath('data.items.0.quantity', 2)
            ->assertJsonPath('data.items.0.unit_price', 20000)
            ->assertJsonPath('data.items.0.discount', 1000);

        $this->assertDatabaseHas('transaction_items', [
            'transaction_id' => $response->json('data.id'),
            'product_id' => $ctx['product'],
            'quantity' => 2,
            'unit_price' => 20000,
            'discount' => 1000,
            'subtotal' => 40000,
        ]);
    }

    public function test_split_payment_total_is_checked_against_server_quote_total(): void
    {
        $ctx = $this->context();
        $this->paymentMethod($ctx['business'], $ctx['outlet'], 'cash', true);
        $this->paymentMethod($ctx['business'], $ctx['outlet'], 'card', false);

        $quote = $this->withToken($ctx['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, [
            ['product_id' => $ctx['product'], 'quantity' => 2, 'unit_price' => 1],
        ]))->assertOk()
            ->assertJsonPath('data.grand_total', 40000);

        $response = $this->withToken($ctx['token'])
            ->postJson('/api/v1/transactions', $this->checkoutPayload($ctx, $quote, [
                ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1],
            ], [
                'grand_total' => 1,
                'payments' => [
                    ['method' => 'cash', 'amount' => 15000],
                    ['method' => 'card', 'amount' => 25000, 'reference' => 'CARD-OK'],
                ],
            ]), ['Idempotency-Key' => $quote->json('data.checkout_idempotency_key')])
            ->assertCreated()
            ->assertJsonPath('data.grand_total', 40000);

        $this->assertSame(40000, (int) DB::table('payments')->where('transaction_id', $response->json('data.id'))->sum('amount'));
        $this->assertDatabaseHas('payments', [
            'transaction_id' => $response->json('data.id'),
            'method' => 'cash',
            'amount' => 15000,
            'status' => 'confirmed',
        ]);
        $this->assertDatabaseHas('payments', [
            'transaction_id' => $response->json('data.id'),
            'method' => 'card',
            'amount' => 25000,
            'status' => 'pending',
        ]);

        $quoteOverpay = $this->withToken($ctx['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, [
            ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1],
        ]))->assertOk();

        $this->withToken($ctx['token'])
            ->postJson('/api/v1/transactions', $this->checkoutPayload($ctx, $quoteOverpay, [
                ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1],
            ], [
                'grand_total' => 1,
                'payments' => [['method' => 'cash', 'amount' => $quoteOverpay->json('data.grand_total') + 1]],
            ]), ['Idempotency-Key' => $quoteOverpay->json('data.checkout_idempotency_key')])
            ->assertUnprocessable()
            ->assertJsonPath('error.code', 'PAYMENT_OVERPAY');
    }

    public function test_async_payment_amount_uses_server_quote_amount(): void
    {
        $ctx = $this->context();
        $this->paymentMethod($ctx['business'], $ctx['outlet'], 'qris', false, true);

        $quote = $this->withToken($ctx['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, [
            ['product_id' => $ctx['product'], 'quantity' => 2, 'unit_price' => 1],
        ]))->assertOk()
            ->assertJsonPath('data.grand_total', 40000);

        $response = $this->withToken($ctx['token'])
            ->postJson('/api/v1/transactions', $this->checkoutPayload($ctx, $quote, [
                ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1],
            ], [
                'grand_total' => 1,
                'payments' => [['method' => 'qris', 'amount' => $quote->json('data.grand_total'), 'reference' => 'QR-DISPLAY']],
            ]), ['Idempotency-Key' => $quote->json('data.checkout_idempotency_key')])
            ->assertCreated()
            ->assertJsonPath('data.status', 'payment_pending')
            ->assertJsonPath('data.grand_total', 40000)
            ->assertJsonPath('data.payments.0.amount', 40000)
            ->assertJsonPath('data.payments.0.status', 'pending');

        $this->assertDatabaseHas('payments', [
            'transaction_id' => $response->json('data.id'),
            'method' => 'qris',
            'amount' => 40000,
            'status' => 'pending',
            'provider' => 'qris',
        ]);
    }

    public function test_quote_response_does_not_expose_pin_hash_tokens_payment_references_or_customer_pii(): void
    {
        $ctx = $this->context();
        $customerId = $this->customer($ctx['business']);
        $this->promotion($ctx['business'], 'Promo', 'SAFE', 'amount', 1000);

        $response = $this->withToken($ctx['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, [
            ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1],
        ], [
            'customer_id' => $customerId,
            'promotion_codes' => ['SAFE'],
        ]))->assertOk();

        $json = json_encode($response->json());
        $this->assertIsString($json);
        $this->assertStringNotContainsString('pin_hash', $json);
        $this->assertStringNotContainsString('tokenable', $json);
        $this->assertStringNotContainsString('provider_reference', $json);
        $this->assertStringNotContainsString('081234567890', $json);
        $this->assertStringNotContainsString('secret', strtolower($json));
    }

    public function test_default_checkout_without_quote_is_rejected(): void
    {
        $ctx = $this->context();
        $this->paymentMethod($ctx['business'], $ctx['outlet']);

        $this->withToken($ctx['token'])
            ->postJson('/api/v1/transactions', $this->quotePayload($ctx, [
                ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1],
            ], [
                'payments' => [['method' => 'cash', 'amount' => 20000]],
            ]), ['Idempotency-Key' => (string) Str::uuid()])
            ->assertUnprocessable()
            ->assertJsonPath('error.code', 'QUOTE_REQUIRED');

        $this->assertSame(0, DB::table('transactions')->where('business_id', $ctx['business'])->count());
    }

    public function test_compatibility_mode_can_explicitly_allow_legacy_checkout_without_quote(): void
    {
        config()->set('nojpos.checkout.require_quote_for_checkout', false);
        $ctx = $this->context();
        $this->paymentMethod($ctx['business'], $ctx['outlet']);

        $legacyPayload = $this->quotePayload($ctx, [
            ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1],
        ], [
            'payments' => [['method' => 'cash', 'amount' => 20000]],
        ]);

        $this->withToken($ctx['token'])
            ->postJson('/api/v1/transactions', $legacyPayload, ['Idempotency-Key' => (string) Str::uuid()])
            ->assertCreated()
            ->assertJsonPath('data.grand_total', 20000);

        $this->assertSame(1, DB::table('transactions')->where('business_id', $ctx['business'])->count());
    }

    public function test_quote_bound_checkout_replay_mismatch_and_reuse_are_safe(): void
    {
        $ctx = $this->context();
        $this->paymentMethod($ctx['business'], $ctx['outlet']);

        $quote = $this->withToken($ctx['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, [
            ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 1],
        ]))->assertOk();

        $payload = $this->checkoutPayload($ctx, $quote, [
            ['product_id' => $ctx['product'], 'quantity' => 1, 'unit_price' => 999999],
        ], [
            'grand_total' => 1,
        ]);

        $key = $quote->json('data.checkout_idempotency_key');
        $first = $this->withToken($ctx['token'])
            ->postJson('/api/v1/transactions', $payload, ['Idempotency-Key' => $key])
            ->assertCreated()
            ->assertJsonPath('data.grand_total', 20000)
            ->assertJsonPath('data.items.0.unit_price', 20000);

        $this->withToken($ctx['token'])
            ->postJson('/api/v1/transactions', $payload, ['Idempotency-Key' => $key])
            ->assertCreated()
            ->assertExactJson($first->json());

        $this->withToken($ctx['token'])
            ->postJson('/api/v1/transactions', array_merge($payload, ['notes' => 'changed body']), ['Idempotency-Key' => $key])
            ->assertConflict()
            ->assertJsonPath('error.code', 'IDEMPOTENCY_MISMATCH');

        $this->withToken($ctx['token'])
            ->postJson('/api/v1/transactions', $payload, ['Idempotency-Key' => (string) Str::uuid()])
            ->assertConflict()
            ->assertJsonPath('error.code', 'QUOTE_STALE');

        $this->assertSame(1, DB::table('transactions')->where('business_id', $ctx['business'])->count());
        $this->assertSame(1, DB::table('payments')->where('business_id', $ctx['business'])->count());
        $this->assertSame(1, DB::table('stock_movements')->where('business_id', $ctx['business'])->count());
    }

    public function test_checkout_rejects_quote_from_another_tenant(): void
    {
        $ctxA = $this->context('Quote Tenant A');
        $ctxB = $this->context('Quote Tenant B');
        $this->paymentMethod($ctxA['business'], $ctxA['outlet']);

        $quoteB = $this->withToken($ctxB['token'])->postJson('/api/v1/transactions/quote', $this->quotePayload($ctxB, [
            ['product_id' => $ctxB['product'], 'quantity' => 1, 'unit_price' => 1],
        ]))->assertOk();

        $this->withToken($ctxA['token'])
            ->postJson('/api/v1/transactions', [
                'quote_id' => $quoteB->json('data.quote_id'),
                'checkout_token' => $quoteB->json('data.checkout_token'),
                'outlet_id' => $ctxA['outlet'],
                'device_id' => $ctxA['device'],
                'cashier_id' => $ctxA['user'],
                'shift_id' => $ctxA['shift'],
                'items' => [
                    ['product_id' => $ctxA['product'], 'quantity' => 1, 'unit_price' => 1],
                ],
                'payments' => [['method' => 'cash', 'amount' => 20000]],
            ], ['Idempotency-Key' => $quoteB->json('data.checkout_idempotency_key')])
            ->assertForbidden()
            ->assertJsonPath('error.code', 'FORBIDDEN');

        $this->assertSame(0, DB::table('transactions')->where('business_id', $ctxA['business'])->count());
    }

    /**
     * @return array<string, string>
     */
    private function context(string $businessName = 'Promo Tenant', int $serviceRate = 0, int $taxRate = 0): array
    {
        $business = $this->business($businessName);
        $outlet = $this->outlet($business, 'Outlet', $serviceRate, $taxRate);
        $user = $this->user($business, Str::slug($businessName).'-owner@example.test', 'owner');
        $device = $this->device($business, $outlet, Str::slug($businessName).'-tablet');
        $shift = $this->shift($business, $outlet, $device, $user);
        $category = $this->category($business, 'Makanan');
        $otherCategory = $this->category($business, 'Minuman');
        $product = $this->product($business, $outlet, $category, 'Nasi Goreng', 20000);

        return [
            'business' => $business,
            'outlet' => $outlet,
            'user' => $user,
            'device' => $device,
            'shift' => $shift,
            'category' => $category,
            'other_category' => $otherCategory,
            'product' => $product,
            'token' => $this->tokenFor($user),
        ];
    }

    private function quotePayload(array $ctx, array $items, array $extra = []): array
    {
        return array_merge([
            'outlet_id' => $ctx['outlet'],
            'device_id' => $ctx['device'],
            'cashier_id' => $ctx['user'],
            'shift_id' => $ctx['shift'],
            'items' => $items,
        ], $extra);
    }

    private function checkoutPayload(array $ctx, object $quote, array $items, array $extra = []): array
    {
        return array_merge($this->quotePayload($ctx, $items), [
            'quote_id' => $quote->json('data.quote_id'),
            'checkout_token' => $quote->json('data.checkout_token'),
            'payments' => [
                ['method' => 'cash', 'amount' => $quote->json('data.grand_total')],
            ],
        ], $extra);
    }

    private function tokenFor(string $userId): string
    {
        DB::table('personal_access_tokens')->insert([
            'id' => (string) Str::uuid(),
            'tokenable_type' => 'App\\Models\\User',
            'tokenable_id' => $userId,
            'name' => 'test',
            'token' => hash('sha256', 'plain-'.$userId),
            'abilities' => json_encode(['*']),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return 'plain-'.$userId;
    }

    private function business(string $name): string
    {
        $id = (string) Str::uuid();
        DB::table('businesses')->insert(['id' => $id, 'name' => $name, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function outlet(string $businessId, string $name = 'Outlet', int $serviceRate = 0, int $taxRate = 0): string
    {
        $id = (string) Str::uuid();
        DB::table('outlets')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'name' => $name,
            'service_charge_rate' => $serviceRate,
            'tax_rate' => $taxRate,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function user(string $businessId, string $email, string $role): string
    {
        $id = (string) Str::uuid();
        DB::table('users')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'name' => $role.' user',
            'email' => $email,
            'password' => Hash::make('password'),
            'role' => $role,
            'pin_hash' => Hash::make('1234'),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function device(string $businessId, string $outletId, string $uuid): string
    {
        $id = (string) Str::uuid();
        DB::table('devices')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'device_uuid' => $uuid,
            'name' => $uuid,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function shift(string $businessId, string $outletId, string $deviceId, string $cashierId): string
    {
        $id = (string) Str::uuid();
        DB::table('shift_sessions')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'status' => 'open',
            'opening_cash' => 0,
            'opened_at' => now(),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function category(string $businessId, string $name): string
    {
        $id = (string) Str::uuid();
        DB::table('product_categories')->insert(['id' => $id, 'business_id' => $businessId, 'name' => $name, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function product(string $businessId, string $outletId, string $categoryId, string $name, int $price): string
    {
        $id = (string) Str::uuid();
        DB::table('products')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'product_category_id' => $categoryId,
            'name' => $name,
            'price' => $price,
            'track_stock' => true,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function promotion(
        string $businessId,
        string $name,
        string $code,
        string $type,
        int $value,
        bool $isActive = true,
        mixed $endsAt = null,
        int $minimumSubtotal = 0,
        ?string $productId = null,
        ?string $categoryId = null,
        array $outletIds = [],
        bool $stackable = false,
    ): string {
        $id = (string) Str::uuid();
        DB::table('promotions')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'name' => $name,
            'code' => $code,
            'type' => $type,
            'value' => $value,
            'minimum_subtotal' => $minimumSubtotal,
            'product_id' => $productId,
            'product_category_id' => $categoryId,
            'is_active' => $isActive,
            'stackable' => $stackable,
            'starts_at' => null,
            'ends_at' => $endsAt,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        foreach ($outletIds as $outletId) {
            DB::table('promotion_outlets')->insert([
                'id' => (string) Str::uuid(),
                'business_id' => $businessId,
                'promotion_id' => $id,
                'outlet_id' => $outletId,
                'created_at' => now(),
                'updated_at' => now(),
            ]);
        }

        return $id;
    }

    private function paymentMethod(string $businessId, string $outletId, string $method = 'cash', bool $isCash = true, bool $async = false): void
    {
        DB::table('payment_method_configs')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'method' => $method,
            'is_cash' => $isCash,
            'async_confirmation_enabled' => $async,
            'provider' => $async ? $method : null,
            'async_expiry_minutes' => $async ? 15 : 10,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
    }

    private function customer(string $businessId): string
    {
        $id = (string) Str::uuid();
        DB::table('customers')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'name' => 'Sensitive Customer',
            'phone' => '081234567890',
            'group' => 'VIP',
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }
}
