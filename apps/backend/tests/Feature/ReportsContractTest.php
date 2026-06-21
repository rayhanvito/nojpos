<?php

namespace Tests\Feature;

use Carbon\CarbonImmutable;
use Illuminate\Foundation\Testing\RefreshDatabase;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;
use Tests\TestCase;

class ReportsContractTest extends TestCase
{
    use RefreshDatabase;

    public function test_owner_can_access_own_reports_cross_business_is_blocked_and_envelope_is_standard(): void
    {
        [$businessA, $outletA, $ownerA, $tokenA] = $this->businessWithUser('owner', 'owner-a@example.test');
        [$businessB, $outletB, $ownerB] = $this->businessWithUser('owner', 'owner-b@example.test');
        $deviceA = $this->device($businessA, $outletA, 'device-a');
        $shiftA = $this->shift($businessA, $outletA, $deviceA, $ownerA);
        $deviceB = $this->device($businessB, $outletB, 'device-b');
        $shiftB = $this->shift($businessB, $outletB, $deviceB, $ownerB);
        $productA = $this->product($businessA, $outletA, 'Kopi', 10000);
        $productB = $this->product($businessB, $outletB, 'Rahasia', 50000);

        $trxA = $this->transaction($businessA, $outletA, $deviceA, $ownerA, $shiftA, 'paid', 10000);
        $this->item($businessA, $trxA, $productA, 'Kopi', 1, 10000, 0);
        $this->payment($businessA, $trxA, 'cash', 10000, 'confirmed', true);
        $trxB = $this->transaction($businessB, $outletB, $deviceB, $ownerB, $shiftB, 'paid', 50000);
        $this->item($businessB, $trxB, $productB, 'Rahasia', 1, 50000, 0);
        $this->payment($businessB, $trxB, 'cash', 50000, 'confirmed', true);

        $response = $this->withToken($tokenA)->getJson('/api/v1/reports/sales-summary?date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonStructure(['data' => ['total_sales', 'transaction_count'], 'meta' => ['server_time', 'timezone', 'window']])
            ->assertJsonPath('data.total_sales', 10000)
            ->assertJsonPath('data.transaction_count', 1);
        $this->assertNoSensitiveReportData($response->json());

        $this->withToken($tokenA)->getJson('/api/v1/reports/sales-summary?shift_id='.$shiftB)
            ->assertNotFound()
            ->assertJsonPath('error.code', 'NOT_FOUND');

        $this->withToken($tokenA)->getJson('/api/v1/reports/sold-products?outlet_id='.$outletB)
            ->assertForbidden()
            ->assertJsonPath('error.code', 'FORBIDDEN');
    }

    public function test_cashier_reports_are_restricted_to_current_shift_and_outlet(): void
    {
        [$businessId, $outletId, $cashierId, $token] = $this->businessWithUser('cashier', 'cashier@example.test');
        $otherCashierId = $this->user($businessId, 'other-cashier@example.test', 'cashier');
        $deviceId = $this->device($businessId, $outletId, 'cashier-device');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $cashierId);
        $otherShiftId = $this->shift($businessId, $outletId, $deviceId, $otherCashierId);
        $productId = $this->product($businessId, $outletId, 'Es Teh', 10000);

        $ownTransaction = $this->transaction($businessId, $outletId, $deviceId, $cashierId, $shiftId, 'paid', 10000);
        $this->item($businessId, $ownTransaction, $productId, 'Es Teh', 1, 10000, 0);
        $this->payment($businessId, $ownTransaction, 'cash', 10000, 'confirmed', true);
        $otherTransaction = $this->transaction($businessId, $outletId, $deviceId, $otherCashierId, $otherShiftId, 'paid', 999000);
        $this->item($businessId, $otherTransaction, $productId, 'Es Teh', 99, 999000, 0);
        $this->payment($businessId, $otherTransaction, 'cash', 999000, 'confirmed', true);

        $this->withToken($token)->getJson('/api/v1/reports/sales-summary?date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonPath('data.total_sales', 10000)
            ->assertJsonPath('data.transaction_count', 1);

        $this->withToken($token)->getJson('/api/v1/reports/sold-products?date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonPath('data.rows.0.quantity_sold', 1)
            ->assertJsonPath('data.rows.0.gross_sales', 10000);

        $this->withToken($token)->getJson('/api/v1/reports/sales-summary?shift_id='.$otherShiftId)
            ->assertForbidden()
            ->assertJsonPath('error.code', 'FORBIDDEN');
    }

    public function test_sales_summary_accounting_excludes_unpaid_and_voided_sales(): void
    {
        [$businessId, $outletId, $ownerId, $token] = $this->businessWithUser('owner', 'accounting@example.test');
        $deviceId = $this->device($businessId, $outletId, 'accounting-device');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId);

        $paidA = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 11000, 1000, 500, 1000, -500);
        $this->payment($businessId, $paidA, 'cash', 11000, 'confirmed', true);
        $paidB = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 22000, 2000, 1000, 2000, -1000);
        $this->payment($businessId, $paidB, 'qris', 22000, 'settled', false);
        $voided = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'voided', 33000);
        $this->payment($businessId, $voided, 'cash', 33000, 'confirmed', true);
        $pending = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'payment_pending', 44000);
        $this->payment($businessId, $pending, 'qris', 44000, 'pending', false);
        $failed = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'payment_failed', 55000);
        $this->payment($businessId, $failed, 'qris', 55000, 'expired', false);

        $this->withToken($token)->getJson('/api/v1/reports/sales-summary?date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonPath('data.total_sales', 33000)
            ->assertJsonPath('data.transaction_count', 2)
            ->assertJsonPath('data.total_void', 33000)
            ->assertJsonPath('data.total_discount', 3000)
            ->assertJsonPath('data.total_service_charge', 1500)
            ->assertJsonPath('data.total_tax', 3000)
            ->assertJsonPath('data.total_rounding', -1500)
            ->assertJsonPath('data.average_transaction_value', 16500)
            ->assertJsonPath('data.payment_totals.0.method', 'cash')
            ->assertJsonPath('data.payment_totals.0.amount', 11000)
            ->assertJsonPath('data.payment_totals.1.method', 'qris')
            ->assertJsonPath('data.payment_totals.1.amount', 22000);
    }

    public function test_sold_products_aggregate_server_side_and_exclude_non_paid_statuses(): void
    {
        [$businessId, $outletId, $ownerId, $token] = $this->businessWithUser('owner', 'sold@example.test');
        $deviceId = $this->device($businessId, $outletId, 'sold-device');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId);
        $productId = $this->product($businessId, $outletId, 'Nasi Goreng', 25000);

        $paidA = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 48000);
        $this->item($businessId, $paidA, $productId, 'Nasi Goreng', 2, 50000, 2000);
        $this->payment($businessId, $paidA, 'cash', 48000, 'confirmed', true);
        $paidB = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 24000);
        $this->item($businessId, $paidB, $productId, 'Nasi Goreng', 1, 25000, 1000);
        $this->payment($businessId, $paidB, 'cash', 24000, 'confirmed', true);
        $pending = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'payment_pending', 25000);
        $this->item($businessId, $pending, $productId, 'Nasi Goreng', 10, 250000, 0);
        $voided = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'voided', 25000);
        $this->item($businessId, $voided, $productId, 'Nasi Goreng', 10, 250000, 0);

        $this->withToken($token)->getJson('/api/v1/reports/sold-products?date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonPath('data.rows.0.product_id', $productId)
            ->assertJsonPath('data.rows.0.quantity_sold', 3)
            ->assertJsonPath('data.rows.0.gross_sales', 75000)
            ->assertJsonPath('data.rows.0.discount_total', 3000)
            ->assertJsonPath('data.rows.0.net_sales', 72000);
    }

    public function test_payment_methods_count_only_confirmed_or_settled_paid_payments_grouped_by_method(): void
    {
        [$businessId, $outletId, $ownerId, $token] = $this->businessWithUser('owner', 'payments@example.test');
        $deviceId = $this->device($businessId, $outletId, 'payment-device');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId);
        $this->paymentMethod($businessId, $outletId, 'cash', true);
        $this->paymentMethod($businessId, $outletId, 'qris', false);

        $cashPaid = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 10000);
        $this->payment($businessId, $cashPaid, 'cash', 10000, 'confirmed', true);
        $qrisPaid = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 20000);
        $this->payment($businessId, $qrisPaid, 'qris', 20000, 'settled', false);
        $pendingPaymentOnPaid = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 30000);
        $this->payment($businessId, $pendingPaymentOnPaid, 'qris', 30000, 'pending', false);
        $failedTransaction = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'payment_failed', 40000);
        $this->payment($businessId, $failedTransaction, 'qris', 40000, 'expired', false);

        $this->withToken($token)->getJson('/api/v1/reports/payment-methods?date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonCount(2, 'data.rows')
            ->assertJsonPath('data.rows.0.method', 'cash')
            ->assertJsonPath('data.rows.0.transaction_count', 1)
            ->assertJsonPath('data.rows.0.gross_amount', 10000)
            ->assertJsonPath('data.rows.1.method', 'qris')
            ->assertJsonPath('data.rows.1.transaction_count', 1)
            ->assertJsonPath('data.rows.1.gross_amount', 20000);
    }

    public function test_cashier_shifts_are_tenant_and_outlet_scoped_with_expected_cash_fields(): void
    {
        [$businessId, $outletId, $ownerId, $token] = $this->businessWithUser('owner', 'shifts@example.test');
        [$otherBusinessId, $otherOutletId, $otherOwnerId] = $this->businessWithUser('owner', 'other-shifts@example.test');
        $deviceId = $this->device($businessId, $outletId, 'shift-device');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId, 'closed', 100000, 125000, 123000, -2000);
        $this->cashMovement($businessId, $outletId, $shiftId, $ownerId, 'cash_in', 30000, 'Modal tambahan');
        $this->cashMovement($businessId, $outletId, $shiftId, $ownerId, 'cash_out', 5000, 'Petty cash');
        $otherDevice = $this->device($otherBusinessId, $otherOutletId, 'other-shift-device');
        $this->shift($otherBusinessId, $otherOutletId, $otherDevice, $otherOwnerId, 'closed', 1, 2, 3, 4);

        $this->withToken($token)->getJson('/api/v1/reports/cashier-shifts?outlet_id='.$outletId.'&date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonCount(1, 'data.rows')
            ->assertJsonPath('data.rows.0.shift_id', $shiftId)
            ->assertJsonPath('data.rows.0.cashier.id', $ownerId)
            ->assertJsonPath('data.rows.0.outlet.id', $outletId)
            ->assertJsonPath('data.rows.0.opening_cash', 100000)
            ->assertJsonPath('data.rows.0.expected_cash', 125000)
            ->assertJsonPath('data.rows.0.declared_closing_cash', 123000)
            ->assertJsonPath('data.rows.0.variance', -2000)
            ->assertJsonPath('data.rows.0.cash_in', 30000)
            ->assertJsonPath('data.rows.0.cash_out', 5000)
            ->assertJsonPath('data.rows.0.status', 'closed');
    }

    public function test_void_refund_audit_returns_void_rows_and_safe_empty_refund_rows_without_mutation(): void
    {
        [$businessId, $outletId, $ownerId, $token] = $this->businessWithUser('owner', 'void-report@example.test');
        $deviceId = $this->device($businessId, $outletId, 'void-report-device');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId);
        $voided = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'voided', 15000);
        $this->cashMovement($businessId, $outletId, $shiftId, $ownerId, 'cash_out', 15000, 'Void: Salah input');
        $this->audit($businessId, $ownerId, 'void', 'transaction', $voided);

        $this->withToken($token)->getJson('/api/v1/reports/void-refund-audit?date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonPath('data.rows.0.transaction_number', DB::table('transactions')->where('id', $voided)->value('number'))
            ->assertJsonPath('data.rows.0.type', 'void')
            ->assertJsonPath('data.rows.0.cashier', 'owner user')
            ->assertJsonPath('data.rows.0.reason', 'Salah input')
            ->assertJsonPath('data.rows.0.amount', 15000)
            ->assertJsonPath('data.refund_rows', []);
    }

    public function test_top_ten_returns_best_selling_and_marks_gross_profit_unavailable_when_cost_incomplete(): void
    {
        [$businessId, $outletId, $ownerId, $token] = $this->businessWithUser('owner', 'top10@example.test');
        $deviceId = $this->device($businessId, $outletId, 'top10-device');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId);
        $bestProduct = $this->product($businessId, $outletId, 'Best Seller', 10000);
        $otherProduct = $this->product($businessId, $outletId, 'Runner Up', 10000);
        $best = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 50000);
        $this->item($businessId, $best, $bestProduct, 'Best Seller', 5, 50000, 0);
        $other = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 20000);
        $this->item($businessId, $other, $otherProduct, 'Runner Up', 2, 20000, 0);

        $this->withToken($token)->getJson('/api/v1/reports/top-10?date='.now('Asia/Jakarta')->toDateString())
            ->assertOk()
            ->assertJsonPath('data.best_selling.0.product_id', $bestProduct)
            ->assertJsonPath('data.best_selling.0.quantity_sold', 5)
            ->assertJsonPath('data.highest_gross_profit', [])
            ->assertJsonPath('data.data_availability.gross_profit', 'cost_coverage_incomplete');
    }

    public function test_sales_summary_uses_outlet_local_day_boundary_and_reports_meta_window(): void
    {
        [$businessId, $outletId, $ownerId, $token] = $this->businessWithUser('owner', 'timezone@example.test');
        DB::table('outlets')->where('id', $outletId)->update(['timezone' => 'Asia/Makassar']);
        $deviceId = $this->device($businessId, $outletId, 'timezone-device');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId, openedAt: CarbonImmutable::parse('2025-12-31 15:00:00', 'UTC'));
        $included = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 17000, createdAt: CarbonImmutable::parse('2025-12-31 16:30:00', 'UTC'));
        $this->payment($businessId, $included, 'cash', 17000, 'confirmed', true);
        $excluded = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 99000, createdAt: CarbonImmutable::parse('2025-12-31 15:30:00', 'UTC'));
        $this->payment($businessId, $excluded, 'cash', 99000, 'confirmed', true);

        $this->withToken($token)->getJson('/api/v1/reports/sales-summary?outlet_id='.$outletId.'&date=2026-01-01')
            ->assertOk()
            ->assertJsonPath('data.total_sales', 17000)
            ->assertJsonPath('data.transaction_count', 1)
            ->assertJsonPath('meta.timezone', 'Asia/Makassar')
            ->assertJsonPath('meta.window.start', '2025-12-31T16:00:00.000000Z')
            ->assertJsonPath('meta.window.end', '2026-01-01T16:00:00.000000Z');
    }

    public function test_report_pagination_metadata_and_per_page_limit_are_explicit(): void
    {
        [$businessId, $outletId, $ownerId, $token] = $this->businessWithUser('owner', 'pagination@example.test');
        $deviceId = $this->device($businessId, $outletId, 'pagination-device');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId);

        foreach (['Produk A', 'Produk B', 'Produk C'] as $index => $name) {
            $productId = $this->product($businessId, $outletId, $name, 10000 + $index);
            $transactionId = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 10000 + $index);
            $this->item($businessId, $transactionId, $productId, $name, 3 - $index, 10000 + $index, 0);
            $this->payment($businessId, $transactionId, 'cash', 10000 + $index, 'confirmed', true);
        }

        $this->withToken($token)->getJson('/api/v1/reports/sold-products?date='.now('Asia/Jakarta')->toDateString().'&per_page=2&page=2')
            ->assertOk()
            ->assertJsonCount(1, 'data.rows')
            ->assertJsonPath('meta.pagination.total', 3)
            ->assertJsonPath('meta.pagination.per_page', 2)
            ->assertJsonPath('meta.pagination.current_page', 2)
            ->assertJsonPath('meta.pagination.last_page', 2)
            ->assertJsonPath('meta.range', 'day')
            ->assertJsonStructure(['meta' => ['generated_at', 'timezone', 'window' => ['start', 'end']]]);

        $this->withToken($token)->getJson('/api/v1/reports/sold-products?date='.now('Asia/Jakarta')->toDateString().'&per_page=101')
            ->assertUnprocessable();
    }

    public function test_report_custom_date_range_uses_outlet_timezone_and_rejects_large_ranges(): void
    {
        [$businessId, $outletId, $ownerId, $token] = $this->businessWithUser('owner', 'custom-range@example.test');
        DB::table('outlets')->where('id', $outletId)->update(['timezone' => 'Asia/Makassar']);
        $deviceId = $this->device($businessId, $outletId, 'custom-range-device');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId, openedAt: CarbonImmutable::parse('2025-12-31 16:00:00', 'UTC'));

        $includedA = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 11000, createdAt: CarbonImmutable::parse('2025-12-31 16:30:00', 'UTC'));
        $this->payment($businessId, $includedA, 'cash', 11000, 'confirmed', true);
        $includedB = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 22000, createdAt: CarbonImmutable::parse('2026-01-02 15:30:00', 'UTC'));
        $this->payment($businessId, $includedB, 'cash', 22000, 'confirmed', true);
        $excluded = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 99000, createdAt: CarbonImmutable::parse('2026-01-02 16:30:00', 'UTC'));
        $this->payment($businessId, $excluded, 'cash', 99000, 'confirmed', true);

        $this->withToken($token)->getJson('/api/v1/reports/sales-summary?outlet_id='.$outletId.'&date_from=2026-01-01&date_to=2026-01-02')
            ->assertOk()
            ->assertJsonPath('data.total_sales', 33000)
            ->assertJsonPath('data.transaction_count', 2)
            ->assertJsonPath('meta.timezone', 'Asia/Makassar')
            ->assertJsonPath('meta.range', 'custom')
            ->assertJsonPath('meta.window.start', '2025-12-31T16:00:00.000000Z')
            ->assertJsonPath('meta.window.end', '2026-01-02T16:00:00.000000Z');

        $this->withToken($token)->getJson('/api/v1/reports/sales-summary?date_from=2026-01-02&date_to=2026-01-01')
            ->assertUnprocessable();

        $this->withToken($token)->getJson('/api/v1/reports/sales-summary?date_from=2026-01-01&date_to=2026-02-15')
            ->assertUnprocessable();
    }

    public function test_report_export_is_explicitly_not_implemented_until_export_contract_exists(): void
    {
        [$businessId, $outletId, $ownerId, $token] = $this->businessWithUser('owner', 'export@example.test');
        $deviceId = $this->device($businessId, $outletId, 'export-device');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId);
        $transactionId = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 10000);
        $this->payment($businessId, $transactionId, 'cash', 10000, 'confirmed', true);

        $this->withToken($token)->getJson('/api/v1/reports/sales-summary?date='.now('Asia/Jakarta')->toDateString().'&export=csv')
            ->assertStatus(501)
            ->assertJsonPath('error.code', 'REPORT_EXPORT_NOT_IMPLEMENTED');
    }

    public function test_report_responses_do_not_expose_sensitive_fields_or_raw_payment_reference_or_customer_pii(): void
    {
        [$businessId, $outletId, $ownerId, $token] = $this->businessWithUser('owner', 'sensitive@example.test');
        $deviceId = $this->device($businessId, $outletId, 'sensitive-device');
        $shiftId = $this->shift($businessId, $outletId, $deviceId, $ownerId);
        $customerId = $this->customer($businessId, 'Pelanggan Rahasia', '08123456789');
        $transaction = $this->transaction($businessId, $outletId, $deviceId, $ownerId, $shiftId, 'paid', 12345, customerId: $customerId);
        $this->payment($businessId, $transaction, 'qris', 12345, 'confirmed', false, 'QRIS-RAW-SECRET-123');

        foreach (['sales-summary', 'sold-products', 'payment-methods', 'cashier-shifts', 'void-refund-audit', 'top-10'] as $report) {
            $response = $this->withToken($token)->getJson('/api/v1/reports/'.$report.'?date='.now('Asia/Jakarta')->toDateString())->assertOk();
            $this->assertNoSensitiveReportData($response->json());
        }
    }

    /**
     * @return array{0:string, 1:string, 2:string, 3:string}
     */
    private function businessWithUser(string $role, string $email): array
    {
        $businessId = $this->business('Business '.$email);
        $outletId = $this->outlet($businessId);
        $userId = $this->user($businessId, $email, $role);

        return [$businessId, $outletId, $userId, $this->tokenFor($userId)];
    }

    private function business(string $name): string
    {
        $id = (string) Str::uuid();
        DB::table('businesses')->insert(['id' => $id, 'name' => $name, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function outlet(string $businessId): string
    {
        $id = (string) Str::uuid();
        DB::table('outlets')->insert(['id' => $id, 'business_id' => $businessId, 'name' => 'Outlet', 'service_charge_rate' => 0, 'tax_rate' => 0, 'created_at' => now(), 'updated_at' => now()]);

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

    private function device(string $businessId, string $outletId, string $uuid): string
    {
        $id = (string) Str::uuid();
        DB::table('devices')->insert(['id' => $id, 'business_id' => $businessId, 'outlet_id' => $outletId, 'device_uuid' => $uuid, 'name' => $uuid, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function shift(
        string $businessId,
        string $outletId,
        string $deviceId,
        string $cashierId,
        string $status = 'open',
        int $openingCash = 0,
        ?int $expectedCash = null,
        ?int $actualCash = null,
        ?int $cashDifference = null,
        CarbonImmutable|string|null $openedAt = null,
    ): string {
        $id = (string) Str::uuid();
        DB::table('shift_sessions')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'status' => $status,
            'opening_cash' => $openingCash,
            'expected_cash' => $expectedCash,
            'actual_cash' => $actualCash,
            'cash_difference' => $cashDifference,
            'opened_at' => $openedAt ?? now(),
            'closed_at' => $status === 'closed' ? now() : null,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function product(string $businessId, string $outletId, string $name, int $price): string
    {
        $categoryId = (string) Str::uuid();
        DB::table('product_categories')->insert(['id' => $categoryId, 'business_id' => $businessId, 'name' => 'Kategori '.$name, 'created_at' => now(), 'updated_at' => now()]);
        $id = (string) Str::uuid();
        DB::table('products')->insert(['id' => $id, 'business_id' => $businessId, 'outlet_id' => $outletId, 'product_category_id' => $categoryId, 'name' => $name, 'price' => $price, 'track_stock' => true, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function paymentMethod(string $businessId, string $outletId, string $method, bool $isCash): string
    {
        $id = (string) Str::uuid();
        DB::table('payment_method_configs')->insert(['id' => $id, 'business_id' => $businessId, 'outlet_id' => $outletId, 'method' => $method, 'is_cash' => $isCash, 'created_at' => now(), 'updated_at' => now()]);

        return $id;
    }

    private function transaction(
        string $businessId,
        string $outletId,
        string $deviceId,
        string $cashierId,
        string $shiftId,
        string $status,
        int $grandTotal,
        int $discountTotal = 0,
        int $serviceChargeTotal = 0,
        int $taxTotal = 0,
        int $roundingTotal = 0,
        CarbonImmutable|string|null $createdAt = null,
        ?string $customerId = null,
    ): string {
        $id = (string) Str::uuid();
        $timestamp = $createdAt ?? now();
        DB::table('transactions')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'device_id' => $deviceId,
            'cashier_id' => $cashierId,
            'shift_id' => $shiftId,
            'customer_id' => $customerId,
            'number' => 'TRX-'.substr($id, 0, 8),
            'status' => $status,
            'subtotal' => max(0, $grandTotal + $discountTotal - $serviceChargeTotal - $taxTotal - $roundingTotal),
            'discount_total' => $discountTotal,
            'item_discount_total' => $discountTotal,
            'cart_discount_total' => 0,
            'service_charge_total' => $serviceChargeTotal,
            'tax_total' => $taxTotal,
            'rounding_total' => $roundingTotal,
            'grand_total' => $grandTotal,
            'created_at' => $timestamp,
            'updated_at' => $timestamp,
        ]);

        return $id;
    }

    private function item(string $businessId, string $transactionId, string $productId, string $name, int $quantity, int $subtotal, int $discount): void
    {
        DB::table('transaction_items')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'transaction_id' => $transactionId,
            'product_id' => $productId,
            'name' => $name,
            'quantity' => $quantity,
            'unit_price' => intdiv($subtotal, max(1, $quantity)),
            'discount' => $discount,
            'subtotal' => $subtotal,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
    }

    private function payment(string $businessId, string $transactionId, string $method, int $amount, string $status, bool $isCash, ?string $providerReference = null): string
    {
        $id = (string) Str::uuid();
        DB::table('payments')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'transaction_id' => $transactionId,
            'method' => $method,
            'reference' => null,
            'amount' => $amount,
            'status' => $status,
            'is_cash' => $isCash,
            'confirmed_at' => in_array($status, ['confirmed', 'settled'], true) ? now() : null,
            'provider' => $isCash ? null : $method,
            'provider_reference' => $providerReference,
            'confirm_expires_at' => $status === 'pending' ? now()->addMinutes(10) : null,
            'failed_at' => in_array($status, ['failed', 'expired', 'declined'], true) ? now() : null,
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    private function cashMovement(string $businessId, string $outletId, string $shiftId, string $actorId, string $type, int $amount, string $reason): void
    {
        DB::table('cash_movements')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'shift_id' => $shiftId,
            'actor_id' => $actorId,
            'type' => $type,
            'amount' => $amount,
            'reason' => $reason,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
    }

    private function audit(string $businessId, string $actorId, string $action, string $entityType, string $entityId): void
    {
        DB::table('audit_logs')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'actor_id' => $actorId,
            'action' => $action,
            'entity_type' => $entityType,
            'entity_id' => $entityId,
            'before' => null,
            'after' => null,
            'created_at' => now(),
        ]);
    }

    private function customer(string $businessId, string $name, string $phone): string
    {
        $id = (string) Str::uuid();
        DB::table('customers')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'name' => $name,
            'phone' => $phone,
            'group' => 'VIP',
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return $id;
    }

    /**
     * @param  array<string, mixed>  $payload
     */
    private function assertNoSensitiveReportData(array $payload): void
    {
        $json = json_encode($payload, JSON_THROW_ON_ERROR);
        foreach (['pin_hash', 'password', 'remember_token', 'token', 'QRIS-RAW-SECRET-123', '08123456789', 'Pelanggan Rahasia'] as $forbidden) {
            $this->assertStringNotContainsString($forbidden, $json);
        }
    }
}
