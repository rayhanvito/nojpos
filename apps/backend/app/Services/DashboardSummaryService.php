<?php

namespace App\Services;

use App\Models\User;
use Carbon\CarbonImmutable;
use Illuminate\Database\Query\Builder;
use Illuminate\Support\Collection;
use Illuminate\Support\Facades\DB;

class DashboardSummaryService
{
    private const PAID_STATUSES = ['paid', 'partially_refunded', 'refunded'];

    private const LOW_STOCK_DEFAULT_THRESHOLD = 5;

    public function __construct(private readonly BusinessClock $clock) {}

    /**
     * @param  array<string, mixed>  $filters
     * @return array{data:array<string, mixed>, meta:array<string, mixed>}
     */
    public function summary(User $user, array $filters): array
    {
        $scope = $this->resolveScope($user, $filters);
        $generatedAt = now()->toISOString();
        $todayTransactions = $this->paidTransactions($scope['business_id'], $scope['outlet_id'], $scope['day_start'], $scope['day_end'])->get();
        $salesToday = (int) $todayTransactions->sum(fn (object $row): int => (int) $row->grand_total);
        $transactionCount = $todayTransactions->count();
        $lowStockRows = $this->lowStockRows($scope['business_id'], $scope['outlet_id']);
        $cashDifference = $this->cashDifference($scope['business_id'], $scope['outlet_id'], $scope['day_start'], $scope['day_end']);

        $data = [
            'meta' => [
                'date' => $scope['date'],
                'range' => $scope['range'],
                'timezone' => $scope['timezone'],
                'business_id' => $scope['business_id'],
                'outlet_id' => $scope['outlet_id'],
                'outlet_name' => $scope['outlet_name'],
                'generated_at' => $generatedAt,
                'currency' => 'IDR',
                'money_format' => 'integer_rupiah',
                'data_status' => 'partial',
                'is_empty_today' => $transactionCount === 0,
            ],
            'kpis' => [
                'sales_today' => $this->kpi($salesToday, 'money', 'Penjualan hari ini', null, false),
                'transaction_count' => $this->kpi($transactionCount, 'integer', 'Jumlah transaksi', null, false),
                'average_transaction' => $this->kpi($transactionCount === 0 ? 0 : intdiv($salesToday, $transactionCount), 'money', 'Rata-rata transaksi', null, false),
                'gross_profit_estimate' => $this->kpi(0, 'money', 'Gross profit estimasi', null, true),
                'low_stock_count' => $this->kpi($lowStockRows->count(), 'integer', 'Stok hampir habis', $lowStockRows->isEmpty() ? null : 'Butuh cek', false),
                'cash_difference' => $this->kpi($cashDifference, 'money', 'Selisih kas', $cashDifference === 0 ? null : 'Perlu konfirmasi', true),
            ],
            'alerts' => $this->alerts($scope, $lowStockRows, $cashDifference),
            'sales_last_7_days' => $this->salesLastSevenDays($scope),
            'payment_methods' => $this->paymentMethods($scope, $salesToday),
            'top_products' => $this->topProducts($scope),
            'low_stock_items' => $this->lowStockItems($lowStockRows),
            'recent_transactions' => $this->recentTransactions($scope),
            'cashier_performance' => $this->cashierPerformance($scope),
            'branch_highlights' => $this->branchHighlights($scope),
            'data_notes' => $this->dataNotes(),
        ];

        return [
            'data' => $data,
            'meta' => [
                'request_id' => request()?->headers->get('X-Request-Id'),
                'generated_at' => $generatedAt,
            ],
        ];
    }

    /**
     * @param  array<string, mixed>  $filters
     * @return array{business_id:string, outlet_id:?string, outlet_name:?string, timezone:string, date:string, range:string, day_start:CarbonImmutable, day_end:CarbonImmutable, chart_start:CarbonImmutable, chart_end:CarbonImmutable}
     */
    private function resolveScope(User $user, array $filters): array
    {
        if (! in_array((string) $user->role, ['owner', 'admin'], true)) {
            throw new DashboardAccessException('FORBIDDEN', 'Anda tidak memiliki akses ke dashboard toko ini.', [], 403);
        }

        $businessId = (string) $user->business_id;
        $outletId = isset($filters['outlet_id']) ? (string) $filters['outlet_id'] : null;
        $outletName = null;

        if ($outletId !== null) {
            $outlet = DB::table('outlets')
                ->where('business_id', $businessId)
                ->where('id', $outletId)
                ->whereNull('deleted_at')
                ->first(['name', 'timezone']);

            if (! $outlet) {
                throw new DashboardAccessException('FORBIDDEN', 'Anda tidak memiliki akses ke dashboard toko ini.', [], 403);
            }

            $outletName = $outlet->name;
            $timezone = $this->clock->outletTimezone($businessId, $outletId);
        } else {
            $timezone = BusinessClock::DEFAULT_TIMEZONE;
        }

        $date = (string) ($filters['date'] ?? $this->clock->localNow($timezone)->toDateString());
        [$dayStart, $dayEnd] = $this->clock->utcDayWindow($timezone, $date);
        $chartLocalStart = CarbonImmutable::parse($date, $timezone)->startOfDay()->subDays(6);
        $chartStart = $chartLocalStart->utc();

        return [
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'outlet_name' => $outletName,
            'timezone' => $timezone,
            'date' => $date,
            'range' => (string) ($filters['range'] ?? 'last_7_days'),
            'day_start' => $dayStart,
            'day_end' => $dayEnd,
            'chart_start' => $chartStart,
            'chart_end' => $dayEnd,
        ];
    }

    private function kpi(int $value, string $type, string $label, ?string $trendLabel, bool $isEstimate): array
    {
        return [
            'value' => $value,
            'type' => $type,
            'label' => $label,
            'trend_label' => $trendLabel,
            'is_estimate' => $isEstimate,
        ];
    }

    private function paidTransactions(string $businessId, ?string $outletId, CarbonImmutable $start, CarbonImmutable $end): Builder
    {
        return DB::table('transactions')
            ->where('transactions.business_id', $businessId)
            ->whereIn('transactions.status', self::PAID_STATUSES)
            ->where('transactions.created_at', '>=', $start)
            ->where('transactions.created_at', '<', $end)
            ->whereNull('transactions.deleted_at')
            ->when($outletId, fn (Builder $query, string $filterOutletId): Builder => $query->where('transactions.outlet_id', $filterOutletId));
    }

    /**
     * @return Collection<int, object>
     */
    private function lowStockRows(string $businessId, ?string $outletId): Collection
    {
        $stockTotals = DB::table('stock_movements')
            ->where('business_id', $businessId)
            ->whereNull('deleted_at')
            ->selectRaw('product_id, outlet_id, coalesce(sum(quantity_delta), 0) as remaining_stock')
            ->groupBy('product_id', 'outlet_id');

        return DB::table('products as products')
            ->joinSub($stockTotals, 'stock', fn ($join) => $join->on('stock.product_id', '=', 'products.id'))
            ->join('outlets as outlets', function ($join) use ($businessId): void {
                $join->on('outlets.id', '=', 'stock.outlet_id')
                    ->where('outlets.business_id', '=', $businessId)
                    ->whereNull('outlets.deleted_at');
            })
            ->where('products.business_id', $businessId)
            ->where('products.track_stock', true)
            ->whereNull('products.deleted_at')
            ->when($outletId, fn (Builder $query, string $filterOutletId): Builder => $query->where('stock.outlet_id', $filterOutletId))
            ->where('stock.remaining_stock', '<=', self::LOW_STOCK_DEFAULT_THRESHOLD)
            ->orderBy('stock.remaining_stock')
            ->limit(20)
            ->get([
                'products.id as product_id',
                'products.name',
                'stock.outlet_id',
                'outlets.name as outlet_name',
                'stock.remaining_stock',
            ]);
    }

    private function cashDifference(string $businessId, ?string $outletId, CarbonImmutable $start, CarbonImmutable $end): int
    {
        return (int) DB::table('shift_sessions')
            ->where('business_id', $businessId)
            ->where('status', 'closed')
            ->where('opened_at', '>=', $start)
            ->where('opened_at', '<', $end)
            ->whereNull('deleted_at')
            ->when($outletId, fn (Builder $query, string $filterOutletId): Builder => $query->where('outlet_id', $filterOutletId))
            ->sum('cash_difference');
    }

    /**
     * @param  array<string, mixed>  $scope
     * @param  Collection<int, object>  $lowStockRows
     * @return array<int, array<string, mixed>>
     */
    private function alerts(array $scope, Collection $lowStockRows, int $cashDifference): array
    {
        $alerts = [];
        $lowStockCount = $lowStockRows->count();
        $outStockCount = $lowStockRows->filter(fn (object $row): bool => (int) $row->remaining_stock <= 0)->count();

        if ($outStockCount > 0) {
            $alerts[] = [
                'id' => 'out-of-stock-'.$scope['date'],
                'type' => 'out_of_stock',
                'severity' => 'danger',
                'title' => $outStockCount.' produk habis',
                'message' => 'Cek stok produk habis sebelum transaksi berikutnya.',
                'outlet_id' => $scope['outlet_id'],
                'outlet_name' => $scope['outlet_name'],
                'entity_type' => 'inventory',
                'entity_id' => null,
                'action_label' => 'Cek inventori',
                'action_path' => '/inventory',
            ];
        } elseif ($lowStockCount > 0) {
            $alerts[] = [
                'id' => 'low-stock-'.$scope['date'],
                'type' => 'low_stock',
                'severity' => 'warning',
                'title' => $lowStockCount.' produk stok menipis',
                'message' => 'Cek stok sebelum jam ramai agar kasir tidak menjual item yang kosong.',
                'outlet_id' => $scope['outlet_id'],
                'outlet_name' => $scope['outlet_name'],
                'entity_type' => 'inventory',
                'entity_id' => null,
                'action_label' => 'Cek inventori',
                'action_path' => '/inventory',
            ];
        }

        $openShiftCount = (int) DB::table('shift_sessions')
            ->where('business_id', $scope['business_id'])
            ->where('status', 'open')
            ->whereNull('deleted_at')
            ->when($scope['outlet_id'], fn (Builder $query, string $filterOutletId): Builder => $query->where('outlet_id', $filterOutletId))
            ->count();

        if ($openShiftCount > 0) {
            $alerts[] = [
                'id' => 'unclosed-shift-'.$scope['date'],
                'type' => 'unclosed_shift',
                'severity' => 'info',
                'title' => $openShiftCount.' shift masih terbuka',
                'message' => 'Pastikan shift ditutup dari aplikasi kasir saat operasional selesai.',
                'outlet_id' => $scope['outlet_id'],
                'outlet_name' => $scope['outlet_name'],
                'entity_type' => 'shift',
                'entity_id' => null,
                'action_label' => 'Cek shift',
                'action_path' => '/reports/shifts',
            ];
        }

        if ($cashDifference !== 0) {
            $alerts[] = [
                'id' => 'cash-difference-'.$scope['date'],
                'type' => 'cash_difference',
                'severity' => 'warning',
                'title' => 'Ada selisih kas',
                'message' => 'Selisih kas terbaca dari shift yang sudah ditutup.',
                'outlet_id' => $scope['outlet_id'],
                'outlet_name' => $scope['outlet_name'],
                'entity_type' => 'cash',
                'entity_id' => null,
                'action_label' => 'Cek laporan kas',
                'action_path' => '/reports',
            ];
        }

        return array_slice($alerts, 0, 5);
    }

    /**
     * @param  array<string, mixed>  $scope
     * @return array<int, array<string, mixed>>
     */
    private function salesLastSevenDays(array $scope): array
    {
        $rows = $this->paidTransactions($scope['business_id'], $scope['outlet_id'], $scope['chart_start'], $scope['chart_end'])
            ->get(['transactions.id', 'transactions.grand_total', 'transactions.created_at']);

        $grouped = $rows->groupBy(fn (object $row): string => CarbonImmutable::parse($row->created_at, 'UTC')->setTimezone($scope['timezone'])->toDateString());
        $anchor = CarbonImmutable::parse($scope['date'], $scope['timezone'])->startOfDay()->subDays(6);
        $series = [];

        for ($index = 0; $index < 7; $index++) {
            $date = $anchor->addDays($index);
            $key = $date->toDateString();
            $transactions = $grouped->get($key, collect());
            $count = $transactions->count();
            $sales = (int) $transactions->sum(fn (object $row): int => (int) $row->grand_total);
            $series[] = [
                'date' => $key,
                'label' => $this->dayLabel($date),
                'sales' => $sales,
                'transaction_count' => $count,
                'average_transaction' => $count === 0 ? 0 : intdiv($sales, $count),
            ];
        }

        return $series;
    }

    /**
     * @param  array<string, mixed>  $scope
     * @return array<int, array<string, mixed>>
     */
    private function paymentMethods(array $scope, int $salesToday): array
    {
        return DB::table('payments as payments')
            ->join('transactions as transactions', function ($join) use ($scope): void {
                $join->on('transactions.id', '=', 'payments.transaction_id')
                    ->where('transactions.business_id', '=', $scope['business_id']);
            })
            ->where('payments.business_id', $scope['business_id'])
            ->whereIn('transactions.status', self::PAID_STATUSES)
            ->whereIn('payments.status', ['confirmed', 'settled'])
            ->where('transactions.created_at', '>=', $scope['day_start'])
            ->where('transactions.created_at', '<', $scope['day_end'])
            ->whereNull('payments.deleted_at')
            ->whereNull('transactions.deleted_at')
            ->when($scope['outlet_id'], fn (Builder $query, string $filterOutletId): Builder => $query->where('transactions.outlet_id', $filterOutletId))
            ->selectRaw('payments.method, count(distinct payments.transaction_id) as transaction_count, coalesce(sum(payments.amount), 0) as amount')
            ->groupBy('payments.method')
            ->orderByDesc('amount')
            ->get()
            ->map(fn (object $row): array => [
                'method' => (string) $row->method,
                'label' => $this->paymentMethodLabel((string) $row->method),
                'amount' => (int) $row->amount,
                'transaction_count' => (int) $row->transaction_count,
                'share_percent' => $salesToday === 0 ? 0.0 : round(((int) $row->amount / $salesToday) * 100, 2),
            ])
            ->values()
            ->all();
    }

    /**
     * @param  array<string, mixed>  $scope
     * @return array<int, array<string, mixed>>
     */
    private function topProducts(array $scope): array
    {
        $rows = DB::table('transaction_items as items')
            ->join('transactions as transactions', function ($join) use ($scope): void {
                $join->on('transactions.id', '=', 'items.transaction_id')
                    ->where('transactions.business_id', '=', $scope['business_id']);
            })
            ->leftJoin('outlets as outlets', function ($join) use ($scope): void {
                $join->on('outlets.id', '=', 'transactions.outlet_id')
                    ->where('outlets.business_id', '=', $scope['business_id']);
            })
            ->where('items.business_id', $scope['business_id'])
            ->whereIn('transactions.status', self::PAID_STATUSES)
            ->where('transactions.created_at', '>=', $scope['day_start'])
            ->where('transactions.created_at', '<', $scope['day_end'])
            ->whereNull('items.deleted_at')
            ->whereNull('transactions.deleted_at')
            ->when($scope['outlet_id'], fn (Builder $query, string $filterOutletId): Builder => $query->where('transactions.outlet_id', $filterOutletId))
            ->selectRaw('items.product_id, items.name, transactions.outlet_id, outlets.name as outlet_name, coalesce(sum(items.quantity), 0) as qty_sold, coalesce(sum(items.subtotal - items.discount), 0) as sales')
            ->groupBy('items.product_id', 'items.name', 'transactions.outlet_id', 'outlets.name')
            ->orderByDesc('qty_sold')
            ->limit(5)
            ->get();
        $topQty = max(1, (int) $rows->max('qty_sold'));

        return $rows->map(fn (object $row): array => [
            'product_id' => $row->product_id,
            'name' => (string) $row->name,
            'qty_sold' => (int) $row->qty_sold,
            'sales' => (int) $row->sales,
            'share_percent' => round(((int) $row->qty_sold / $topQty) * 100, 2),
            'outlet_id' => $scope['outlet_id'] === null ? $row->outlet_id : $scope['outlet_id'],
            'outlet_name' => $scope['outlet_id'] === null ? $row->outlet_name : $scope['outlet_name'],
        ])->values()->all();
    }

    /**
     * @param  Collection<int, object>  $rows
     * @return array<int, array<string, mixed>>
     */
    private function lowStockItems(Collection $rows): array
    {
        return $rows->take(5)->map(fn (object $row): array => [
            'product_id' => (string) $row->product_id,
            'name' => (string) $row->name,
            'sku' => null,
            'outlet_id' => (string) $row->outlet_id,
            'outlet_name' => (string) $row->outlet_name,
            'remaining_stock' => (int) $row->remaining_stock,
            'threshold' => self::LOW_STOCK_DEFAULT_THRESHOLD,
            'unit' => null,
            'status' => (int) $row->remaining_stock <= 0 ? 'out' : 'low',
        ])->values()->all();
    }

    /**
     * @param  array<string, mixed>  $scope
     * @return array<int, array<string, mixed>>
     */
    private function recentTransactions(array $scope): array
    {
        return $this->paidTransactions($scope['business_id'], $scope['outlet_id'], $scope['day_start'], $scope['day_end'])
            ->leftJoin('outlets as outlets', function ($join) use ($scope): void {
                $join->on('outlets.id', '=', 'transactions.outlet_id')
                    ->where('outlets.business_id', '=', $scope['business_id']);
            })
            ->leftJoin('users as cashiers', function ($join) use ($scope): void {
                $join->on('cashiers.id', '=', 'transactions.cashier_id')
                    ->where('cashiers.business_id', '=', $scope['business_id']);
            })
            ->leftJoin('payments as payments', function ($join) use ($scope): void {
                $join->on('payments.transaction_id', '=', 'transactions.id')
                    ->where('payments.business_id', '=', $scope['business_id'])
                    ->whereIn('payments.status', ['confirmed', 'settled'])
                    ->whereNull('payments.deleted_at');
            })
            ->select([
                'transactions.id',
                'transactions.number',
                'transactions.created_at',
                'transactions.outlet_id',
                'transactions.status',
                'transactions.grand_total',
                'outlets.name as outlet_name',
                'cashiers.name as cashier_name',
                DB::raw('min(payments.method) as payment_method'),
            ])
            ->groupBy('transactions.id', 'transactions.number', 'transactions.created_at', 'transactions.outlet_id', 'transactions.status', 'transactions.grand_total', 'outlets.name', 'cashiers.name')
            ->orderByDesc('transactions.created_at')
            ->limit(5)
            ->get()
            ->map(fn (object $row): array => [
                'transaction_id' => (string) $row->id,
                'code' => (string) $row->number,
                'time' => CarbonImmutable::parse($row->created_at, 'UTC')->setTimezone($scope['timezone'])->format('H:i'),
                'occurred_at' => CarbonImmutable::parse($row->created_at, 'UTC')->toISOString(),
                'outlet_id' => (string) $row->outlet_id,
                'outlet_name' => (string) $row->outlet_name,
                'cashier_name' => $row->cashier_name,
                'payment_method' => $row->payment_method ? $this->paymentMethodLabel((string) $row->payment_method) : null,
                'total' => (int) $row->grand_total,
                'status' => (string) $row->status,
            ])
            ->values()
            ->all();
    }

    /**
     * @param  array<string, mixed>  $scope
     * @return array<int, array<string, mixed>>
     */
    private function cashierPerformance(array $scope): array
    {
        $voidCounts = DB::table('transactions')
            ->where('business_id', $scope['business_id'])
            ->where('status', 'voided')
            ->where('created_at', '>=', $scope['day_start'])
            ->where('created_at', '<', $scope['day_end'])
            ->whereNull('deleted_at')
            ->when($scope['outlet_id'], fn (Builder $query, string $filterOutletId): Builder => $query->where('outlet_id', $filterOutletId))
            ->selectRaw('cashier_id, count(*) as void_count')
            ->groupBy('cashier_id');

        $refundCounts = DB::table('refunds as refunds')
            ->join('transactions as transactions', function ($join) use ($scope): void {
                $join->on('transactions.id', '=', 'refunds.transaction_id')
                    ->where('transactions.business_id', '=', $scope['business_id']);
            })
            ->where('refunds.business_id', $scope['business_id'])
            ->where('refunds.status', 'finalized')
            ->where('refunds.finalized_at', '>=', $scope['day_start'])
            ->where('refunds.finalized_at', '<', $scope['day_end'])
            ->whereNull('refunds.deleted_at')
            ->when($scope['outlet_id'], fn (Builder $query, string $filterOutletId): Builder => $query->where('refunds.outlet_id', $filterOutletId))
            ->selectRaw('transactions.cashier_id, count(*) as refund_count')
            ->groupBy('transactions.cashier_id');

        $cashDifferences = DB::table('shift_sessions')
            ->where('business_id', $scope['business_id'])
            ->where('status', 'closed')
            ->where('opened_at', '>=', $scope['day_start'])
            ->where('opened_at', '<', $scope['day_end'])
            ->whereNull('deleted_at')
            ->when($scope['outlet_id'], fn (Builder $query, string $filterOutletId): Builder => $query->where('outlet_id', $filterOutletId))
            ->selectRaw('cashier_id, coalesce(sum(cash_difference), 0) as cash_difference')
            ->groupBy('cashier_id');

        return $this->paidTransactions($scope['business_id'], $scope['outlet_id'], $scope['day_start'], $scope['day_end'])
            ->join('users as users', function ($join) use ($scope): void {
                $join->on('users.id', '=', 'transactions.cashier_id')
                    ->where('users.business_id', '=', $scope['business_id']);
            })
            ->leftJoinSub($voidCounts, 'voids', fn ($join) => $join->on('voids.cashier_id', '=', 'transactions.cashier_id'))
            ->leftJoinSub($refundCounts, 'refunds', fn ($join) => $join->on('refunds.cashier_id', '=', 'transactions.cashier_id'))
            ->leftJoinSub($cashDifferences, 'cash_diffs', fn ($join) => $join->on('cash_diffs.cashier_id', '=', 'transactions.cashier_id'))
            ->selectRaw('transactions.cashier_id, users.name, count(transactions.id) as transaction_count, coalesce(sum(transactions.grand_total), 0) as sales, coalesce(max(voids.void_count), 0) as void_count, coalesce(max(refunds.refund_count), 0) as refund_count, coalesce(max(cash_diffs.cash_difference), 0) as cash_difference')
            ->groupBy('transactions.cashier_id', 'users.name')
            ->orderByDesc('sales')
            ->limit(5)
            ->get()
            ->map(function (object $row): array {
                $count = (int) $row->transaction_count;
                $cashDifference = (int) $row->cash_difference;

                return [
                    'cashier_id' => (string) $row->cashier_id,
                    'name' => (string) $row->name,
                    'transaction_count' => $count,
                    'sales' => (int) $row->sales,
                    'average_transaction' => $count === 0 ? 0 : intdiv((int) $row->sales, $count),
                    'void_count' => (int) $row->void_count,
                    'refund_count' => (int) $row->refund_count,
                    'cash_difference' => $cashDifference,
                    'note' => $cashDifference === 0 ? null : 'Perlu cek selisih kas',
                ];
            })
            ->values()
            ->all();
    }

    /**
     * @param  array<string, mixed>  $scope
     * @return array<int, array<string, mixed>>
     */
    private function branchHighlights(array $scope): array
    {
        if ($scope['outlet_id'] !== null) {
            return [];
        }

        $outlets = DB::table('outlets')
            ->where('business_id', $scope['business_id'])
            ->whereNull('deleted_at')
            ->orderBy('name')
            ->get(['id', 'name']);

        if ($outlets->count() <= 1) {
            return [];
        }

        $lowStockByOutlet = $this->lowStockRows($scope['business_id'], null)
            ->groupBy('outlet_id')
            ->map(fn (Collection $rows): int => $rows->count());

        return $outlets->map(function (object $outlet) use ($scope, $lowStockByOutlet): array {
            $sales = (int) $this->paidTransactions($scope['business_id'], (string) $outlet->id, $scope['day_start'], $scope['day_end'])->sum('grand_total');
            $transactionCount = (int) $this->paidTransactions($scope['business_id'], (string) $outlet->id, $scope['day_start'], $scope['day_end'])->count();
            $openShiftCount = (int) DB::table('shift_sessions')
                ->where('business_id', $scope['business_id'])
                ->where('outlet_id', $outlet->id)
                ->where('status', 'open')
                ->whereNull('deleted_at')
                ->count();
            $cashDifference = $this->cashDifference($scope['business_id'], (string) $outlet->id, $scope['day_start'], $scope['day_end']);
            $lowStockCount = (int) ($lowStockByOutlet[(string) $outlet->id] ?? 0);
            [$status, $severity, $summary] = $this->branchStatus($sales, $lowStockCount, $openShiftCount, $cashDifference);

            return [
                'outlet_id' => (string) $outlet->id,
                'name' => (string) $outlet->name,
                'status' => $status,
                'summary' => $summary,
                'sales_today' => $sales,
                'transaction_count' => $transactionCount,
                'low_stock_count' => $lowStockCount,
                'open_shift_count' => $openShiftCount,
                'cash_difference' => $cashDifference,
                'severity' => $severity,
            ];
        })->values()->all();
    }

    /**
     * @return array{0:string, 1:string, 2:string}
     */
    private function branchStatus(int $sales, int $lowStockCount, int $openShiftCount, int $cashDifference): array
    {
        if ($cashDifference !== 0) {
            return ['cash_issue', 'warning', 'Ada selisih kas yang perlu dicek.'];
        }

        if ($openShiftCount > 0) {
            return ['unclosed_shift', 'info', 'Masih ada shift terbuka.'];
        }

        if ($lowStockCount > 0) {
            return ['needs_stock', 'warning', 'Ada stok yang perlu dicek.'];
        }

        if ($sales === 0) {
            return ['inactive', 'info', 'Belum ada transaksi hari ini.'];
        }

        return ['normal', 'success', 'Penjualan berjalan normal.'];
    }

    /**
     * @return array<int, array<string, string>>
     */
    private function dataNotes(): array
    {
        return [
            [
                'key' => 'gross_profit_estimate',
                'message' => 'Gross profit masih estimasi karena data modal/COGS produk belum lengkap.',
                'severity' => 'info',
            ],
            [
                'key' => 'cash_difference',
                'message' => 'Selisih kas hanya membaca data shift yang sudah tersedia dan tidak memicu rekonsiliasi kas.',
                'severity' => 'warning',
            ],
        ];
    }

    private function dayLabel(CarbonImmutable $date): string
    {
        return match ((int) $date->format('N')) {
            1 => 'Sen',
            2 => 'Sel',
            3 => 'Rab',
            4 => 'Kam',
            5 => 'Jum',
            6 => 'Sab',
            default => 'Min',
        };
    }

    private function paymentMethodLabel(string $method): string
    {
        return match ($method) {
            'cash' => 'Tunai',
            'qris' => 'QRIS',
            'card' => 'Kartu',
            'transfer' => 'Transfer',
            default => strtoupper(str_replace('_', ' ', $method)),
        };
    }
}
