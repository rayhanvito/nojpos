<?php

namespace Tests\Feature;

use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class EndToEndOperationalFlowTest extends TestCase
{
    use RefreshDatabase;

    public function test_cashier_operational_flow_from_login_to_receipt_refund_void_reports_inventory_and_close(): void
    {
        config()->set('nojpos.checkout.require_quote_for_checkout', true);

        $ctx = $this->context();
        $other = $this->context('Other Tenant', 'other-owner@example.test', 'other-cashier@example.test', 'other-terminal');

        $login = $this->postJson('/api/v1/auth/login', [
            'email' => $ctx['owner_email'],
            'password' => 'password',
            'device_uuid' => $ctx['device_uuid'],
        ])->assertOk()
            ->assertJsonPath('data.business.id', $ctx['business'])
            ->assertJsonPath('data.outlets.0.id', $ctx['outlet'])
            ->assertJsonPath('data.device.id', $ctx['device']);

        $token = $login->json('data.token');

        $this->withToken($token)->postJson('/api/v1/auth/pin-switch', [
            'pin' => '1234',
            'device_id' => $ctx['device'],
            'outlet_id' => $ctx['outlet'],
        ])->assertOk()
            ->assertJsonPath('data.cashier.id', $ctx['cashier'])
            ->assertJsonPath('data.terminal_session.status', 'active');

        $this->withToken($token)->getJson('/api/v1/outlets/'.$ctx['outlet'].'/store-state')
            ->assertOk()
            ->assertJsonPath('data.status', 'open');

        $openShiftPayload = [
            'outlet_id' => $ctx['outlet'],
            'device_id' => $ctx['device'],
            'cashier_id' => $ctx['cashier'],
            'opening_cash' => 100000,
        ];
        $openShiftKey = (string) Str::uuid();
        $openShift = $this->withToken($token)->withHeaders(['Idempotency-Key' => $openShiftKey])
            ->postJson('/api/v1/shifts/open', $openShiftPayload)
            ->assertCreated()
            ->assertJsonPath('data.status', 'open');

        $this->withToken($token)->withHeaders(['Idempotency-Key' => $openShiftKey])
            ->postJson('/api/v1/shifts/open', $openShiftPayload)
            ->assertCreated()
            ->assertExactJson($openShift->json());

        $shiftId = $openShift->json('data.id');
        $this->assertSame(1, DB::table('shift_sessions')->where('business_id', $ctx['business'])->where('status', 'open')->count());

        $this->withToken($token)->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, $shiftId, [
            ['product_id' => $other['product_a'], 'quantity' => 1, 'unit_price' => 1],
        ]))->assertNotFound()
            ->assertJsonPath('error.code', 'PRODUCT_NOT_FOUND');

        $quote = $this->withToken($token)->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, $shiftId, [
            ['product_id' => $ctx['product_a'], 'quantity' => 2, 'unit_price' => 1],
        ]))->assertOk()
            ->assertJsonPath('data.subtotal', 24000)
            ->assertJsonPath('data.grand_total', 24000);

        $checkoutPayload = $this->checkoutPayload($ctx, $shiftId, $quote, [
            ['product_id' => $ctx['product_a'], 'quantity' => 2, 'unit_price' => 1],
        ]);
        $checkoutKey = $quote->json('data.checkout_idempotency_key');
        $checkout = $this->withToken($token)->withHeaders(['Idempotency-Key' => $checkoutKey])
            ->postJson('/api/v1/transactions', $checkoutPayload)
            ->assertCreated()
            ->assertJsonPath('data.status', 'paid')
            ->assertJsonPath('data.grand_total', 24000);

        $this->withToken($token)->withHeaders(['Idempotency-Key' => $checkoutKey])
            ->postJson('/api/v1/transactions', $checkoutPayload)
            ->assertCreated()
            ->assertExactJson($checkout->json());

        $saleId = $checkout->json('data.id');
        $saleItemId = DB::table('transaction_items')->where('transaction_id', $saleId)->value('id');

        $this->assertSame(1, DB::table('transactions')->where('business_id', $ctx['business'])->where('id', $saleId)->count());
        $this->assertSame(1, DB::table('payments')->where('business_id', $ctx['business'])->where('transaction_id', $saleId)->where('is_cash', true)->where('status', 'confirmed')->count());
        $this->assertSame(8, $this->stock($ctx, $ctx['product_a']));

        $this->withToken($token)->getJson('/api/v1/transactions/'.$saleId)
            ->assertOk()
            ->assertJsonPath('data.receipt.status', 'paid')
            ->assertJsonPath('data.receipt.items.0.name', 'E2E Kopi Susu')
            ->assertJsonPath('data.receipt.items.0.quantity', 2)
            ->assertJsonPath('data.receipt.items.0.unit_price', 12000)
            ->assertJsonPath('data.receipt.totals.grand_total', 24000)
            ->assertJsonPath('data.receipt.cashier.id', $ctx['cashier']);

        $refundPayload = [
            'reason' => 'Retur satu item E2E',
            'refund_method' => 'cash',
            'authorization_pin' => '1234',
            'lines' => [[
                'transaction_item_id' => $saleItemId,
                'quantity' => 1,
                'restock' => true,
                'non_restock_reason' => null,
            ]],
        ];
        $refundKey = (string) Str::uuid();
        $refund = $this->withToken($token)->withHeaders(['Idempotency-Key' => $refundKey])
            ->postJson('/api/v1/transactions/'.$saleId.'/refund', $refundPayload)
            ->assertCreated()
            ->assertJsonPath('data.transaction_status', 'partially_refunded')
            ->assertJsonPath('data.total_refund_amount', 12000);

        $this->withToken($token)->withHeaders(['Idempotency-Key' => $refundKey])
            ->postJson('/api/v1/transactions/'.$saleId.'/refund', $refundPayload)
            ->assertCreated()
            ->assertExactJson($refund->json());

        $this->assertSame(9, $this->stock($ctx, $ctx['product_a']));
        $this->assertSame(1, DB::table('refunds')->where('business_id', $ctx['business'])->where('transaction_id', $saleId)->count());
        $this->assertSame(1, DB::table('stock_movements')->where('business_id', $ctx['business'])->where('transaction_id', $saleId)->where('type', 'refund_reversal')->count());

        $voidQuote = $this->withToken($token)->postJson('/api/v1/transactions/quote', $this->quotePayload($ctx, $shiftId, [
            ['product_id' => $ctx['product_b'], 'quantity' => 1, 'unit_price' => 1],
        ]))->assertOk()
            ->assertJsonPath('data.grand_total', 15000);

        $voidCheckoutPayload = $this->checkoutPayload($ctx, $shiftId, $voidQuote, [
            ['product_id' => $ctx['product_b'], 'quantity' => 1, 'unit_price' => 1],
        ]);
        $voidCheckout = $this->withToken($token)->withHeaders(['Idempotency-Key' => $voidQuote->json('data.checkout_idempotency_key')])
            ->postJson('/api/v1/transactions', $voidCheckoutPayload)
            ->assertCreated()
            ->assertJsonPath('data.status', 'paid');
        $voidSaleId = $voidCheckout->json('data.id');

        $voidPayload = [
            'transaction_id' => $voidSaleId,
            'shift_id' => $shiftId,
            'reason' => 'Salah input E2E',
        ];
        $voidKey = (string) Str::uuid();
        $void = $this->withToken($token)->withHeaders(['Idempotency-Key' => $voidKey])
            ->postJson('/api/v1/voids', $voidPayload)
            ->assertOk()
            ->assertJsonPath('data.status', 'voided');

        $this->withToken($token)->withHeaders(['Idempotency-Key' => $voidKey])
            ->postJson('/api/v1/voids', $voidPayload)
            ->assertOk()
            ->assertExactJson($void->json());

        $this->assertSame(5, $this->stock($ctx, $ctx['product_b']));
        $this->assertSame(1, DB::table('stock_movements')->where('business_id', $ctx['business'])->where('transaction_id', $voidSaleId)->where('type', 'void')->count());
        $this->assertSame(1, DB::table('cash_movements')->where('business_id', $ctx['business'])->where('shift_id', $shiftId)->where('type', 'cash_out')->where('amount', 15000)->where('reason', 'Void: Salah input E2E')->count());
        $this->withToken($token)->getJson('/api/v1/transactions/'.$voidSaleId)
            ->assertOk()
            ->assertJsonPath('data.receipt.status', 'voided')
            ->assertJsonPath('data.receipt.items.0.name', 'E2E Roti Bakar');

        $summary = $this->withToken($token)->getJson('/api/v1/reports/sales-summary?date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonPath('data.total_sales', 24000)
            ->assertJsonPath('data.refund_total', 12000)
            ->assertJsonPath('data.net_sales', 12000)
            ->assertJsonPath('data.total_void', 15000);
        $this->assertIsArray($summary->json('data.payment_totals'));

        $this->withToken($token)->getJson('/api/v1/reports/void-refund-audit?date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonPath('data.rows.0.type', 'void')
            ->assertJsonPath('data.rows.0.amount', 15000)
            ->assertJsonPath('data.refund_rows.0.type', 'refund')
            ->assertJsonPath('data.refund_rows.0.refund_amount', 12000);

        $this->withToken($token)->getJson('/api/v1/inventory?outlet_id='.$ctx['outlet'])
            ->assertOk()
            ->assertJsonPath('data.items.0.product_id', $ctx['product_a'])
            ->assertJsonPath('data.items.0.stock_on_hand', 9);

        foreach (['auth.login', 'auth.pin_switch', 'terminal_session.started', 'shift.open', 'refund.create', 'transaction.void'] as $action) {
            $this->assertSame(1, DB::table('audit_logs')->where('business_id', $ctx['business'])->where('action', $action)->count(), 'Expected one audit event for '.$action);
        }
        $this->assertSame(1, DB::table('audit_logs')->where('business_id', $ctx['business'])->where('action', 'transaction.create')->where('entity_id', $saleId)->count());
        $this->assertSame(1, DB::table('audit_logs')->where('business_id', $ctx['business'])->where('action', 'transaction.create')->where('entity_id', $voidSaleId)->count());

        $expectedCash = 100000 + 24000 - 12000;
        $closeShift = $this->withToken($token)->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/shifts/'.$shiftId.'/close', [
                'actual_cash' => $expectedCash,
                'pin' => '1234',
            ])->assertOk()
            ->assertJsonPath('data.expected_cash', $expectedCash)
            ->assertJsonPath('data.status', 'closed');

        $this->assertSame('closed', $closeShift->json('data.status'));

        $ownerCloseToken = $this->postJson('/api/v1/auth/login', [
            'email' => $ctx['owner_email'],
            'password' => 'password',
            'device_uuid' => $ctx['device_uuid'],
        ])->assertOk()->json('data.token');

        $this->withToken($ownerCloseToken)->withHeaders(['Idempotency-Key' => (string) Str::uuid()])
            ->postJson('/api/v1/outlets/'.$ctx['outlet'].'/store/close', [
                'reason' => 'Tutup selesai E2E',
                'authorization_pin' => '1234',
            ])->assertOk()
            ->assertJsonPath('data.status', 'closed');
    }

    /**
     * @return array<string, string>
     */
    private function context(string $businessName = 'E2E Tenant', string $ownerEmail = 'owner-e2e@example.test', string $cashierEmail = 'cashier-e2e@example.test', string $deviceUuid = 'e2e-terminal'): array
    {
        $business = (string) Str::uuid();
        $outlet = (string) Str::uuid();
        $device = (string) Str::uuid();
        $owner = (string) Str::uuid();
        $cashier = (string) Str::uuid();
        $category = (string) Str::uuid();
        $productA = (string) Str::uuid();
        $productB = (string) Str::uuid();
        $now = now();

        DB::table('businesses')->insert(['id' => $business, 'name' => $businessName, 'created_at' => $now, 'updated_at' => $now]);
        DB::table('outlets')->insert([
            'id' => $outlet,
            'business_id' => $business,
            'name' => 'Outlet '.$businessName,
            'timezone' => 'Asia/Jakarta',
            'service_charge_rate' => 0,
            'tax_rate' => 0,
            'store_open_close_enabled' => true,
            'receipt_header_name' => $businessName,
            'receipt_footer_note' => 'Terima kasih',
            'created_at' => $now,
            'updated_at' => $now,
        ]);
        DB::table('devices')->insert(['id' => $device, 'business_id' => $business, 'outlet_id' => $outlet, 'device_uuid' => $deviceUuid, 'name' => $deviceUuid, 'created_at' => $now, 'updated_at' => $now]);
        DB::table('users')->insert([
            ['id' => $owner, 'business_id' => $business, 'name' => 'Owner '.$businessName, 'email' => $ownerEmail, 'password' => Hash::make('password'), 'role' => 'owner', 'pin_hash' => Hash::make('1234'), 'created_at' => $now, 'updated_at' => $now],
            ['id' => $cashier, 'business_id' => $business, 'name' => 'Cashier '.$businessName, 'email' => $cashierEmail, 'password' => Hash::make('password'), 'role' => 'cashier', 'pin_hash' => Hash::make('1234'), 'created_at' => $now, 'updated_at' => $now],
        ]);
        DB::table('product_categories')->insert(['id' => $category, 'business_id' => $business, 'name' => 'E2E Menu', 'created_at' => $now, 'updated_at' => $now]);
        DB::table('products')->insert([
            ['id' => $productA, 'business_id' => $business, 'outlet_id' => $outlet, 'product_category_id' => $category, 'name' => 'E2E Kopi Susu', 'barcode' => 'E2E-A-'.$productA, 'price' => 12000, 'track_stock' => true, 'created_at' => $now, 'updated_at' => $now],
            ['id' => $productB, 'business_id' => $business, 'outlet_id' => $outlet, 'product_category_id' => $category, 'name' => 'E2E Roti Bakar', 'barcode' => 'E2E-B-'.$productB, 'price' => 15000, 'track_stock' => true, 'created_at' => $now, 'updated_at' => $now],
        ]);
        DB::table('payment_method_configs')->insert(['id' => (string) Str::uuid(), 'business_id' => $business, 'outlet_id' => $outlet, 'method' => 'cash', 'is_cash' => true, 'created_at' => $now, 'updated_at' => $now]);
        $this->stockMovement($business, $outlet, $productA, 10);
        $this->stockMovement($business, $outlet, $productB, 5);

        return [
            'business' => $business,
            'outlet' => $outlet,
            'device' => $device,
            'device_uuid' => $deviceUuid,
            'owner' => $owner,
            'owner_email' => $ownerEmail,
            'cashier' => $cashier,
            'cashier_email' => $cashierEmail,
            'category' => $category,
            'product_a' => $productA,
            'product_b' => $productB,
        ];
    }

    private function quotePayload(array $ctx, string $shiftId, array $items): array
    {
        return [
            'outlet_id' => $ctx['outlet'],
            'device_id' => $ctx['device'],
            'cashier_id' => $ctx['cashier'],
            'shift_id' => $shiftId,
            'items' => $items,
        ];
    }

    private function checkoutPayload(array $ctx, string $shiftId, object $quote, array $items): array
    {
        return $this->quotePayload($ctx, $shiftId, $items) + [
            'quote_id' => $quote->json('data.quote_id'),
            'checkout_token' => $quote->json('data.checkout_token'),
            'payments' => [['method' => 'cash', 'amount' => $quote->json('data.grand_total')]],
        ];
    }

    private function stockMovement(string $businessId, string $outletId, string $productId, int $quantity): void
    {
        DB::table('stock_movements')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'product_id' => $productId,
            'transaction_id' => null,
            'type' => 'purchase',
            'quantity_delta' => $quantity,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
    }

    private function stock(array $ctx, string $productId): int
    {
        return (int) DB::table('stock_movements')
            ->where('business_id', $ctx['business'])
            ->where('outlet_id', $ctx['outlet'])
            ->where('product_id', $productId)
            ->sum('quantity_delta');
    }
}
