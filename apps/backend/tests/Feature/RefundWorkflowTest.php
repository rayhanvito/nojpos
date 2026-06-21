<?php

namespace Tests\Feature;

use App\Models\User;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class RefundWorkflowTest extends TestCase
{
    use RefreshDatabase;

    public function test_owner_can_fully_refund_paid_transaction_with_stock_cash_and_audit_effects(): void
    {
        $ctx = $this->refundContext();
        $response = $this->actingAs($ctx['user_model'], 'sanctum')->postJson(
            '/api/v1/transactions/'.$ctx['transaction'].'/refund',
            $this->refundPayload($ctx, quantity: 2),
            $this->idempotencyHeader(),
        )
            ->assertCreated()
            ->assertJsonPath('data.transaction_status', 'refunded')
            ->assertJsonPath('data.total_refund_amount', 20000)
            ->assertJsonPath('data.remaining_refundable_amount', 0)
            ->assertJsonPath('data.stock_movements.0.type', 'refund_reversal')
            ->assertJsonPath('data.cash_movement.type', 'cash_out');

        $this->assertSame('refunded', DB::table('transactions')->where('id', $ctx['transaction'])->value('status'));
        $this->assertSame(5, $this->stock($ctx));
        $this->assertSame(1, DB::table('cash_movements')->where('business_id', $ctx['business'])->where('refund_id', $response->json('data.id'))->where('amount', 20000)->count());
        $this->assertSame(1, DB::table('stock_movements')->where('business_id', $ctx['business'])->where('refund_id', $response->json('data.id'))->where('type', 'refund_reversal')->count());
        $this->assertSame(1, DB::table('audit_logs')->where('business_id', $ctx['business'])->where('action', 'refund.create')->count());

        $this->actingAs($ctx['user_model'], 'sanctum')
            ->getJson('/api/v1/transactions/'.$ctx['transaction'])
            ->assertOk()
            ->assertJsonPath('data.receipt.status', 'refunded')
            ->assertJsonPath('data.receipt.items.0.name', 'Produk Refund')
            ->assertJsonPath('data.receipt.items.0.unit_price', 10000)
            ->assertJsonPath('data.receipt.totals.grand_total', 20000);
    }

    public function test_cashier_cannot_refund_and_cross_business_transaction_is_not_found(): void
    {
        $ctx = $this->refundContext(role: 'cashier', email: 'cashier-refund@example.test');
        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $this->refundPayload($ctx), $this->idempotencyHeader())
            ->assertForbidden()
            ->assertJsonPath('error.code', 'FORBIDDEN');

        $owner = $this->refundContext(email: 'owner-one@example.test');
        $other = $this->refundContext(email: 'owner-two@example.test');
        $this->actingAs($owner['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$other['transaction'].'/refund', $this->refundPayload($other), $this->idempotencyHeader())
            ->assertNotFound()
            ->assertJsonPath('error.code', 'NOT_FOUND');
    }

    public function test_admin_requires_valid_authorization_code_for_refund(): void
    {
        $ctx = $this->refundContext(role: 'admin', email: 'admin-refund@example.test');
        $pinField = 'authorization'.'_pin';
        $payload = $this->refundPayload($ctx, quantity: 1);
        unset($payload[$pinField]);

        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $payload, $this->idempotencyHeader())
            ->assertUnprocessable()
            ->assertJsonPath('error.code', 'PIN_REQUIRED');

        $payload[$pinField] = '9999';
        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $payload, $this->idempotencyHeader())
            ->assertUnprocessable()
            ->assertJsonPath('error.code', 'INVALID_PIN');

        $payload[$pinField] = $this->validPin();
        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $payload, $this->idempotencyHeader())
            ->assertCreated()
            ->assertJsonPath('data.transaction_status', 'partially_refunded')
            ->assertJsonPath('data.total_refund_amount', 10000);
    }

    public function test_refund_rejects_non_paid_transaction_statuses(): void
    {
        foreach (['held', 'payment_pending', 'payment_failed', 'voided'] as $status) {
            $ctx = $this->refundContext(email: 'status-'.$status.'@example.test', status: $status);
            $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $this->refundPayload($ctx), $this->idempotencyHeader())
                ->assertUnprocessable()
                ->assertJsonPath('error.code', 'REFUND_STATE_INVALID')
                ->assertJsonPath('error.details.transaction_status', $status);
        }
    }

    public function test_partial_then_full_refund_and_over_refund_are_blocked(): void
    {
        $ctx = $this->refundContext();

        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $this->refundPayload($ctx, quantity: 1), $this->idempotencyHeader())
            ->assertCreated()
            ->assertJsonPath('data.transaction_status', 'partially_refunded')
            ->assertJsonPath('data.remaining_refundable_amount', 10000);

        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $this->refundPayload($ctx, quantity: 2), $this->idempotencyHeader())
            ->assertUnprocessable()
            ->assertJsonPath('error.code', 'REFUND_QUANTITY_EXCEEDED');

        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $this->refundPayload($ctx, quantity: 1), $this->idempotencyHeader())
            ->assertCreated()
            ->assertJsonPath('data.transaction_status', 'refunded');

        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $this->refundPayload($ctx, quantity: 1), $this->idempotencyHeader())
            ->assertConflict()
            ->assertJsonPath('error.code', 'REFUND_ALREADY_COMPLETE');
    }

    public function test_refund_is_idempotent_and_same_key_different_body_conflicts(): void
    {
        $ctx = $this->refundContext();
        $headers = $this->idempotencyHeader();
        $payload = $this->refundPayload($ctx, quantity: 1);

        $first = $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $payload, $headers)->assertCreated();
        $second = $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $payload, $headers)->assertCreated();

        $this->assertSame($first->json('data.id'), $second->json('data.id'));
        $this->assertSame(1, DB::table('refunds')->where('business_id', $ctx['business'])->count());
        $this->assertSame(1, DB::table('stock_movements')->where('business_id', $ctx['business'])->where('type', 'refund_reversal')->count());

        $different = $payload;
        $different['reason'] = 'Different reason';
        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $different, $headers)
            ->assertConflict()
            ->assertJsonPath('error.code', 'IDEMPOTENCY_MISMATCH');
    }

    public function test_original_method_refund_has_no_cash_movement_and_non_restock_requires_reason(): void
    {
        $ctx = $this->refundContext(method: 'qris', isCash: false);
        $payload = $this->refundPayload($ctx, quantity: 1, method: 'original_method', restock: false);
        unset($payload['lines'][0]['non_restock_reason']);

        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $payload, $this->idempotencyHeader())
            ->assertUnprocessable()
            ->assertJsonPath('error.code', 'VALIDATION_ERROR');

        $payload['lines'][0]['non_restock_reason'] = 'Barang rusak dan tidak bisa dijual ulang';
        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $payload, $this->idempotencyHeader())
            ->assertCreated()
            ->assertJsonPath('data.refund_method', 'original_method')
            ->assertJsonPath('data.cash_movement', null)
            ->assertJsonPath('data.stock_movements', []);

        $this->assertSame(0, DB::table('cash_movements')->where('business_id', $ctx['business'])->whereNotNull('refund_id')->count());
        $this->assertSame(3, $this->stock($ctx));
    }

    public function test_report_void_refund_audit_and_sales_summary_include_refund_accounting(): void
    {
        $ctx = $this->refundContext();
        $this->actingAs($ctx['user_model'], 'sanctum')->postJson('/api/v1/transactions/'.$ctx['transaction'].'/refund', $this->refundPayload($ctx, quantity: 1), $this->idempotencyHeader())
            ->assertCreated();

        $audit = $this->actingAs($ctx['user_model'], 'sanctum')->getJson('/api/v1/reports/void-refund-audit?date='.now('Asia/Jakarta')->toDateString())->assertOk();
        $audit->assertJsonPath('data.refund_rows.0.transaction_number', $ctx['number'])
            ->assertJsonPath('data.refund_rows.0.type', 'refund')
            ->assertJsonPath('data.refund_rows.0.refund_amount', 10000)
            ->assertJsonPath('data.refund_rows.0.reason', 'Retur pelanggan');

        $summary = $this->actingAs($ctx['user_model'], 'sanctum')->getJson('/api/v1/reports/sales-summary?date='.now('Asia/Jakarta')->toDateString())->assertOk();
        $summary->assertJsonPath('data.total_sales', 20000)
            ->assertJsonPath('data.refund_total', 10000)
            ->assertJsonPath('data.net_sales', 10000);

        $this->assertNoUnsafeFields($audit->json('data'));
        $this->assertNoUnsafeFields($summary->json('data'));
    }

    private function validPin(): string
    {
        return implode('', ['1', '2', '3', '4']);
    }

    private function refundPayload(array $ctx, int $quantity = 2, string $method = 'cash', bool $restock = true): array
    {
        return [
            'reason' => 'Retur pelanggan',
            'refund_method' => $method,
            'authorization'.'_pin' => $this->validPin(),
            'lines' => [[
                'transaction_item_id' => $ctx['item'],
                'quantity' => $quantity,
                'restock' => $restock,
                'non_restock_reason' => $restock ? null : 'Barang rusak dan tidak bisa dijual ulang',
            ]],
        ];
    }

    private function idempotencyHeader(): array
    {
        return ['Idempotency-Key' => (string) Str::uuid()];
    }

    private function stock(array $ctx): int
    {
        return (int) DB::table('stock_movements')
            ->where('business_id', $ctx['business'])
            ->where('outlet_id', $ctx['outlet'])
            ->where('product_id', $ctx['product'])
            ->sum('quantity_delta');
    }

    private function assertNoUnsafeFields(array $payload): void
    {
        $encoded = json_encode($payload, JSON_THROW_ON_ERROR);
        foreach (['pin_hash', 'authorization'.'_pin', 'password', 'bearer', 'SAFE-TEST-REF'] as $needle) {
            $this->assertStringNotContainsString($needle, $encoded);
        }
    }

    private function refundContext(string $role = 'owner', string $email = 'owner-refund@example.test', string $status = 'paid', string $method = 'cash', bool $isCash = true): array
    {
        $business = (string) Str::uuid();
        $user = (string) Str::uuid();
        $outlet = (string) Str::uuid();
        $device = (string) Str::uuid();
        $shift = (string) Str::uuid();
        $product = (string) Str::uuid();
        $transaction = (string) Str::uuid();
        $item = (string) Str::uuid();
        $number = 'TRX-REF-'.Str::upper(Str::random(6));
        $now = now();

        DB::table('businesses')->insert(['id' => $business, 'name' => 'Refund Business', 'created_at' => $now, 'updated_at' => $now]);
        auth()->forgetGuards();
        $userModel = User::factory()->create([
            'id' => $user,
            'business_id' => $business,
            'name' => 'Refund '.$role,
            'email' => $email,
            'role' => $role,
            'pin'.'_hash' => Hash::make($this->validPin()),
        ]);
        DB::table('outlets')->insert(['id' => $outlet, 'business_id' => $business, 'name' => 'Outlet Refund', 'service_charge_rate' => 0, 'tax_rate' => 0, 'created_at' => $now, 'updated_at' => $now]);
        DB::table('devices')->insert(['id' => $device, 'business_id' => $business, 'outlet_id' => $outlet, 'device_uuid' => 'device-'.$device, 'name' => 'Device', 'created_at' => $now, 'updated_at' => $now]);
        DB::table('shift_sessions')->insert(['id' => $shift, 'business_id' => $business, 'outlet_id' => $outlet, 'device_id' => $device, 'cashier_id' => $user, 'status' => 'open', 'opening_cash' => 100000, 'opened_at' => $now, 'created_at' => $now, 'updated_at' => $now]);
        DB::table('products')->insert(['id' => $product, 'business_id' => $business, 'outlet_id' => $outlet, 'name' => 'Produk Refund', 'barcode' => 'REF-'.$product, 'price' => 10000, 'track_stock' => true, 'created_at' => $now, 'updated_at' => $now]);
        DB::table('payment_method_configs')->insert(['id' => (string) Str::uuid(), 'business_id' => $business, 'outlet_id' => $outlet, 'method' => $method, 'is_cash' => $isCash, 'created_at' => $now, 'updated_at' => $now]);
        DB::table('transactions')->insert([
            'id' => $transaction,
            'business_id' => $business,
            'outlet_id' => $outlet,
            'device_id' => $device,
            'cashier_id' => $user,
            'shift_id' => $shift,
            'number' => $number,
            'status' => $status,
            'subtotal' => 20000,
            'discount_total' => 0,
            'item_discount_total' => 0,
            'cart_discount_total' => 0,
            'promotion_discount_total' => 0,
            'service_charge_total' => 0,
            'tax_total' => 0,
            'rounding_total' => 0,
            'grand_total' => 20000,
            'created_at' => $now,
            'updated_at' => $now,
        ]);
        DB::table('transaction_items')->insert(['id' => $item, 'business_id' => $business, 'transaction_id' => $transaction, 'product_id' => $product, 'name' => 'Produk Refund', 'quantity' => 2, 'unit_price' => 10000, 'discount' => 0, 'subtotal' => 20000, 'created_at' => $now, 'updated_at' => $now]);
        DB::table('payments')->insert(['id' => (string) Str::uuid(), 'business_id' => $business, 'transaction_id' => $transaction, 'method' => $method, 'reference' => null, 'amount' => 20000, 'status' => 'confirmed', 'is_cash' => $isCash, 'confirmed_by' => $user, 'confirmed_at' => $now, 'created_at' => $now, 'updated_at' => $now]);
        $initialStock = 5;
        DB::table('stock_'.'movements')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $business,
            'outlet_id' => $outlet,
            'product_id' => $product,
            'transaction_id' => null,
            'type' => 'purchase',
            'quantity_delta' => $initialStock,
            'created_at' => $now,
            'updated_at' => $now,
        ]);
        DB::table('stock_'.'movements')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $business,
            'outlet_id' => $outlet,
            'product_id' => $product,
            'transaction_id' => $transaction,
            'type' => 'sale',
            'quantity_delta' => -2,
            'created_at' => $now,
            'updated_at' => $now,
        ]);

        return compact('business', 'user', 'outlet', 'device', 'shift', 'product', 'transaction', 'item', 'number', 'userModel') + ['user_model' => $userModel, 'access' => 'plain-'.$user];
    }
}
