<?php

namespace Tests\Feature;

use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class AsyncPaymentTest extends TestCase
{
    use RefreshDatabase;

    public function test_signed_webhook_settles_pending_payment_once(): void
    {
        [$business, $user, $transaction, $payment] = $this->pendingPayment();
        config(['services.payment_webhooks.qris.secret' => 'test-webhook-secret']);
        $body = ['event_id' => 'evt-'.Str::uuid(), 'payment_ref' => $payment['provider_reference'], 'outcome' => 'confirmed'];
        $signature = hash_hmac('sha256', json_encode($body), 'test-webhook-secret');

        $this->postJson('/api/v1/payment-webhooks/qris', $body, ['X-Payment-Signature' => $signature])->assertOk();
        $this->postJson('/api/v1/payment-webhooks/qris', $body, ['X-Payment-Signature' => $signature])->assertOk();

        $this->assertDatabaseHas('transactions', ['id' => $transaction, 'business_id' => $business, 'status' => 'paid']);
        $this->assertDatabaseHas('payments', ['id' => $payment['id'], 'status' => 'confirmed']);
        $this->assertSame(1, DB::table('stock_movements')->where('transaction_id', $transaction)->count());
        $this->assertSame(1, DB::table('audit_logs')->where('entity_id', $payment['id'])->where('action', 'payment.confirm')->count());
    }

    public function test_invalid_signature_is_rejected_without_settlement(): void
    {
        [, , $transaction, $payment] = $this->pendingPayment();
        $this->postJson('/api/v1/payment-webhooks/qris', ['event_id' => 'evt-'.Str::uuid(), 'payment_ref' => $payment['provider_reference'], 'outcome' => 'confirmed'], ['X-Payment-Signature' => 'invalid'])->assertForbidden();
        $this->assertDatabaseHas('transactions', ['id' => $transaction, 'status' => 'payment_pending']);
        $this->assertDatabaseHas('payments', ['id' => $payment['id'], 'status' => 'pending']);
    }

    public function test_polling_expires_pending_payment_without_stock_effect(): void
    {
        [$business, $user, $transaction, $payment] = $this->pendingPayment(expiresAt: now()->subSecond());
        $this->withToken($this->tokenFor($user))->getJson('/api/v1/transactions/'.$transaction)->assertOk()->assertJsonPath('data.status', 'payment_failed')->assertJsonPath('data.payments.0.status', 'expired');
        $this->assertSame(0, DB::table('stock_movements')->where('transaction_id', $transaction)->count());
        $this->assertDatabaseHas('payments', ['id' => $payment['id'], 'status' => 'expired']);
    }

    public function test_tenant_cannot_poll_other_business_transaction(): void
    {
        [, , $transaction] = $this->pendingPayment();
        [$otherBusiness, $otherUser] = $this->businessUser();
        $this->withToken($this->tokenFor($otherUser))->getJson('/api/v1/transactions/'.$transaction)->assertNotFound();
    }

    public function test_checkout_keeps_cash_synchronous_and_async_qris_pending(): void
    {
        [$business, $user, $transaction] = $this->pendingPayment();
        $token = $this->tokenFor($user);
        $pending = $this->withToken($token)->getJson('/api/v1/transactions/'.$transaction)->assertOk();
        $pending->assertJsonPath('data.status', 'payment_pending')
            ->assertJsonPath('data.payments.0.status', 'pending')
            ->assertJsonStructure(['data' => ['payments' => [['confirm_expires_at', 'payment_ref']]]]);

        DB::table('payments')->where('transaction_id', $transaction)->update(['method' => 'cash', 'is_cash' => true, 'status' => 'confirmed', 'confirm_expires_at' => null]);
        DB::table('transactions')->where('id', $transaction)->update(['status' => 'paid']);
        $cash = $this->withToken($token)->getJson('/api/v1/transactions/'.$transaction)->assertOk();
        $cash->assertJsonPath('data.status', 'paid')->assertJsonPath('data.payments.0.status', 'confirmed');
        $this->assertSame($business, $cash->json('data.business_id'));
    }

    public function test_webhook_after_expiry_keeps_payment_failed(): void
    {
        [, , $transaction, $payment] = $this->pendingPayment(expiresAt: now()->subSecond());
        config(['services.payment_webhooks.qris.secret' => 'test-webhook-secret']);
        $body = ['event_id' => 'evt-'.Str::uuid(), 'payment_ref' => $payment['provider_reference'], 'outcome' => 'confirmed'];
        $this->postJson('/api/v1/payment-webhooks/qris', $body, ['X-Payment-Signature' => hash_hmac('sha256', json_encode($body), 'test-webhook-secret')])->assertOk();
        $this->assertDatabaseHas('transactions', ['id' => $transaction, 'status' => 'payment_failed']);
        $this->assertDatabaseHas('payments', ['id' => $payment['id'], 'status' => 'expired']);
        $this->assertSame(0, DB::table('stock_movements')->where('transaction_id', $transaction)->count());
    }

    private function pendingPayment($expiresAt = null): array
    {
        [$business, $user] = $this->businessUser();
        $outlet = (string) Str::uuid();
        $device = (string) Str::uuid();
        $shift = (string) Str::uuid();
        $product = (string) Str::uuid();
        $transaction = (string) Str::uuid();
        $paymentId = (string) Str::uuid();
        DB::table('outlets')->insert(['id' => $outlet, 'business_id' => $business, 'name' => 'Test outlet', 'service_charge_rate' => 0, 'tax_rate' => 0, 'created_at' => now(), 'updated_at' => now()]);
        DB::table('devices')->insert(['id' => $device, 'business_id' => $business, 'outlet_id' => $outlet, 'device_uuid' => 'device-'.Str::uuid(), 'name' => 'device', 'created_at' => now(), 'updated_at' => now()]);
        DB::table('shift_sessions')->insert(['id' => $shift, 'business_id' => $business, 'outlet_id' => $outlet, 'device_id' => $device, 'cashier_id' => $user, 'opening_cash' => 0, 'status' => 'open', 'opened_at' => now(), 'created_at' => now(), 'updated_at' => now()]);
        DB::table('products')->insert(['id' => $product, 'business_id' => $business, 'outlet_id' => $outlet, 'name' => 'Product', 'price' => 10000, 'track_stock' => true, 'created_at' => now(), 'updated_at' => now()]);
        DB::table('transactions')->insert(['id' => $transaction, 'business_id' => $business, 'outlet_id' => $outlet, 'device_id' => $device, 'cashier_id' => $user, 'shift_id' => $shift, 'number' => 'TRX-'.Str::uuid(), 'status' => 'payment_pending', 'subtotal' => 10000, 'discount_total' => 0, 'service_charge_total' => 0, 'tax_total' => 0, 'rounding_total' => 0, 'grand_total' => 10000, 'created_at' => now(), 'updated_at' => now()]);
        DB::table('transaction_items')->insert(['id' => (string) Str::uuid(), 'business_id' => $business, 'transaction_id' => $transaction, 'product_id' => $product, 'name' => 'Product', 'quantity' => 1, 'unit_price' => 10000, 'discount' => 0, 'subtotal' => 10000, 'created_at' => now(), 'updated_at' => now()]);
        $payment = ['id' => $paymentId, 'provider_reference' => (string) Str::uuid()];
        DB::table('payments')->insert(['id' => $paymentId, 'business_id' => $business, 'transaction_id' => $transaction, 'method' => 'qris', 'amount' => 10000, 'status' => 'pending', 'is_cash' => false, 'provider' => 'qris', 'provider_reference' => $payment['provider_reference'], 'confirm_expires_at' => $expiresAt ?? now()->addMinutes(10), 'created_at' => now(), 'updated_at' => now()]);

        return [$business, $user, $transaction, $payment];
    }

    private function businessUser(): array
    {
        $business = (string) Str::uuid();
        $user = (string) Str::uuid();
        DB::table('businesses')->insert(['id' => $business, 'name' => 'Business', 'created_at' => now(), 'updated_at' => now()]);
        DB::table('users')->insert(['id' => $user, 'business_id' => $business, 'name' => 'Cashier', 'email' => Str::uuid().'@test.local', 'password' => Hash::make('password'), 'role' => 'cashier', 'pin_hash' => Hash::make('1234'), 'created_at' => now(), 'updated_at' => now()]);

        return [$business, $user];
    }

    private function tokenFor(string $user): string
    {
        DB::table('personal_access_tokens')->insert(['id' => (string) Str::uuid(), 'tokenable_type' => 'App\\Models\\User', 'tokenable_id' => $user, 'name' => 'test', 'token' => hash('sha256', 'plain-'.$user), 'abilities' => json_encode(['*']), 'created_at' => now(), 'updated_at' => now()]);

        return 'plain-'.$user;
    }
}
