<?php

namespace App\Services;

use App\Models\User;
use Carbon\CarbonImmutable;
use Illuminate\Database\Query\Builder;
use Illuminate\Support\Collection;
use Illuminate\Support\Facades\DB;

class ReportService
{
    public function __construct(private readonly BusinessClock $clock) {}

    /**
     * @param  array<string, mixed>  $filters
     * @return array{data:array<string, mixed>, meta:array<string, mixed>}
     */
    public function salesSummary(User $user, array $filters): array
    {
        $scope = $this->resolveTransactionScope($user, $filters);

        /** @var Collection<int, object> $paidRows */
        $paidRows = $this->transactionBaseQuery($scope)
            ->where('transactions.status', 'paid')
            ->select([
                'transactions.id',
                'transactions.grand_total',
                'transactions.discount_total',
                'transactions.service_charge_total',
                'transactions.tax_total',
                'transactions.rounding_total',
                'transactions.created_at',
            ])
            ->get();

        $transactionCount = $paidRows->count();
        $totalSales = (int) $paidRows->sum(fn (object $row): int => (int) $row->grand_total);

        $data = [
            'total_sales' => $totalSales,
            'transaction_count' => $transactionCount,
            'payment_totals' => $this->paymentTotals($scope['business_id'], $paidRows->pluck('id')),
            'total_discount' => (int) $paidRows->sum(fn (object $row): int => (int) $row->discount_total),
            'total_service_charge' => (int) $paidRows->sum(fn (object $row): int => (int) $row->service_charge_total),
            'total_tax' => (int) $paidRows->sum(fn (object $row): int => (int) $row->tax_total),
            'total_rounding' => (int) $paidRows->sum(fn (object $row): int => (int) $row->rounding_total),
            'total_void' => (int) $this->transactionBaseQuery($scope)->where('transactions.status', 'voided')->sum('transactions.grand_total'),
            'average_transaction_value' => $transactionCount === 0 ? 0 : intdiv($totalSales, $transactionCount),
            'chart' => $this->chartSeries($paidRows, $scope['bucket'], $scope['timezone']),
        ];

        return ['data' => $data, 'meta' => $this->meta($scope)];
    }

    /**
     * @param  array<string, mixed>  $filters
     * @return array{data:array<string, mixed>, meta:array<string, mixed>}
     */
    public function soldProducts(User $user, array $filters): array
    {
        $scope = $this->resolveTransactionScope($user, $filters);

        $rows = DB::table('transaction_items as items')
            ->join('transactions as transactions', function ($join) use ($scope): void {
                $join->on('transactions.id', '=', 'items.transaction_id')
                    ->where('transactions.business_id', '=', $scope['business_id']);
            })
            ->join('outlets as outlets', function ($join) use ($scope): void {
                $join->on('outlets.id', '=', 'transactions.outlet_id')
                    ->where('outlets.business_id', '=', $scope['business_id']);
            })
            ->leftJoin('products as products', function ($join) use ($scope): void {
                $join->on('products.id', '=', 'items.product_id')
                    ->where('products.business_id', '=', $scope['business_id']);
            })
            ->leftJoin('product_categories as categories', function ($join) use ($scope): void {
                $join->on('categories.id', '=', 'products.product_category_id')
                    ->where('categories.business_id', '=', $scope['business_id']);
            })
            ->where('items.business_id', $scope['business_id'])
            ->where('transactions.status', 'paid')
            ->where('transactions.created_at', '>=', $scope['start'])
            ->where('transactions.created_at', '<', $scope['end'])
            ->when($scope['outlet_id'], fn (Builder $query, string $outletId): Builder => $query->where('transactions.outlet_id', $outletId))
            ->when($scope['cashier_id'], fn (Builder $query, string $cashierId): Builder => $query->where('transactions.cashier_id', $cashierId))
            ->when($scope['shift_id'], fn (Builder $query, string $shiftId): Builder => $query->where('transactions.shift_id', $shiftId))
            ->selectRaw('items.product_id, items.name as product_name, categories.name as category_name, transactions.outlet_id, outlets.name as outlet_name, sum(items.quantity) as quantity_sold, sum(items.subtotal) as gross_sales, sum(items.discount) as discount_total')
            ->groupBy('items.product_id', 'items.name', 'categories.name', 'transactions.outlet_id', 'outlets.name')
            ->orderByDesc('quantity_sold')
            ->paginate(50);

        $rows->getCollection()->transform(fn (object $row): array => [
            'product_id' => $row->product_id,
            'product_name' => $row->product_name,
            'category_name' => $row->category_name,
            'quantity_sold' => (int) $row->quantity_sold,
            'gross_sales' => (int) $row->gross_sales,
            'discount_total' => (int) $row->discount_total,
            'net_sales' => (int) $row->gross_sales - (int) $row->discount_total,
            'outlet_id' => $row->outlet_id,
            'outlet_name' => $row->outlet_name,
        ]);

        return [
            'data' => ['rows' => $rows->items()],
            'meta' => array_merge($this->meta($scope), ['pagination' => [
                'total' => $rows->total(),
                'per_page' => $rows->perPage(),
                'current_page' => $rows->currentPage(),
            ]]),
        ];
    }

    /**
     * @param  array<string, mixed>  $filters
     * @return array{data:array<string, mixed>, meta:array<string, mixed>}
     */
    public function paymentMethods(User $user, array $filters): array
    {
        $scope = $this->resolveTransactionScope($user, $filters);

        $rows = DB::table('payments as payments')
            ->join('transactions as transactions', function ($join) use ($scope): void {
                $join->on('transactions.id', '=', 'payments.transaction_id')
                    ->where('transactions.business_id', '=', $scope['business_id']);
            })
            ->leftJoin('payment_method_configs as configs', function ($join) use ($scope): void {
                $join->on('configs.business_id', '=', 'payments.business_id')
                    ->on('configs.method', '=', 'payments.method')
                    ->where('configs.business_id', '=', $scope['business_id']);
            })
            ->where('payments.business_id', $scope['business_id'])
            ->where('transactions.status', 'paid')
            ->whereIn('payments.status', ['confirmed', 'settled'])
            ->where('transactions.created_at', '>=', $scope['start'])
            ->where('transactions.created_at', '<', $scope['end'])
            ->when($scope['outlet_id'], fn (Builder $query, string $outletId): Builder => $query->where('transactions.outlet_id', $outletId))
            ->when($scope['cashier_id'], fn (Builder $query, string $cashierId): Builder => $query->where('transactions.cashier_id', $cashierId))
            ->when($scope['shift_id'], fn (Builder $query, string $shiftId): Builder => $query->where('transactions.shift_id', $shiftId))
            ->selectRaw('payments.method, min(configs.id) as payment_method_config_id, count(distinct payments.transaction_id) as transaction_count, sum(payments.amount) as gross_amount')
            ->groupBy('payments.method')
            ->orderBy('payments.method')
            ->get()
            ->map(fn (object $row): array => [
                'method' => $row->method,
                'payment_method_config_id' => $row->payment_method_config_id,
                'transaction_count' => (int) $row->transaction_count,
                'gross_amount' => (int) $row->gross_amount,
                'net_amount' => (int) $row->gross_amount,
            ])
            ->all();

        return ['data' => ['rows' => $rows], 'meta' => $this->meta($scope)];
    }

    /**
     * @param  array<string, mixed>  $filters
     * @return array{data:array<string, mixed>, meta:array<string, mixed>}
     */
    public function cashierShifts(User $user, array $filters): array
    {
        $scope = $this->resolveTransactionScope($user, $filters);
        $moveTotals = DB::table('cash_movements')
            ->where('business_id', $scope['business_id'])
            ->selectRaw("shift_id, coalesce(sum(case when type='cash_in' then amount else 0 end),0) as cash_in, coalesce(sum(case when type='cash_out' then amount else 0 end),0) as cash_out")
            ->groupBy('shift_id');

        $rows = DB::table('shift_sessions as shifts')
            ->join('users as users', function ($join) use ($scope): void {
                $join->on('users.id', '=', 'shifts.cashier_id')
                    ->where('users.business_id', '=', $scope['business_id']);
            })
            ->join('outlets as outlets', function ($join) use ($scope): void {
                $join->on('outlets.id', '=', 'shifts.outlet_id')
                    ->where('outlets.business_id', '=', $scope['business_id']);
            })
            ->leftJoinSub($moveTotals, 'moves', fn ($join) => $join->on('moves.shift_id', '=', 'shifts.id'))
            ->where('shifts.business_id', $scope['business_id'])
            ->where('shifts.opened_at', '>=', $scope['start'])
            ->where('shifts.opened_at', '<', $scope['end'])
            ->when($scope['outlet_id'], fn (Builder $query, string $outletId): Builder => $query->where('shifts.outlet_id', $outletId))
            ->when($scope['cashier_id'], fn (Builder $query, string $cashierId): Builder => $query->where('shifts.cashier_id', $cashierId))
            ->when($scope['shift_id'], fn (Builder $query, string $shiftId): Builder => $query->where('shifts.id', $shiftId))
            ->select([
                'shifts.id',
                'shifts.cashier_id',
                'shifts.outlet_id',
                'shifts.opened_at',
                'shifts.closed_at',
                'shifts.opening_cash',
                'shifts.expected_cash',
                'shifts.actual_cash',
                'shifts.cash_difference',
                'shifts.status',
                'users.name as cashier_name',
                'outlets.name as outlet_name',
                DB::raw('coalesce(moves.cash_in,0) as cash_in'),
                DB::raw('coalesce(moves.cash_out,0) as cash_out'),
            ])
            ->orderByDesc('shifts.opened_at')
            ->paginate(50);

        $rows->getCollection()->transform(fn (object $shift): array => [
            'shift_id' => $shift->id,
            'shift_number' => $shift->id,
            'cashier' => ['id' => $shift->cashier_id, 'name' => $shift->cashier_name],
            'outlet' => ['id' => $shift->outlet_id, 'name' => $shift->outlet_name],
            'opened_at' => $shift->opened_at,
            'closed_at' => $shift->closed_at,
            'opening_cash' => (int) $shift->opening_cash,
            'expected_cash' => (int) ($shift->expected_cash ?? 0),
            'declared_closing_cash' => (int) ($shift->actual_cash ?? 0),
            'variance' => (int) ($shift->cash_difference ?? 0),
            'cash_in' => (int) $shift->cash_in,
            'cash_out' => (int) $shift->cash_out,
            'refunds' => 0,
            'status' => $shift->status,
        ]);

        return [
            'data' => ['rows' => $rows->items()],
            'meta' => array_merge($this->meta($scope), ['pagination' => [
                'total' => $rows->total(),
                'per_page' => $rows->perPage(),
                'current_page' => $rows->currentPage(),
            ]]),
        ];
    }

    /**
     * @param  array<string, mixed>  $filters
     * @return array{data:array<string, mixed>, meta:array<string, mixed>}
     */
    public function voidRefundAudit(User $user, array $filters): array
    {
        $scope = $this->resolveTransactionScope($user, $filters);
        $voidReasons = DB::table('cash_movements')
            ->where('business_id', $scope['business_id'])
            ->where('type', 'cash_out')
            ->where('reason', 'like', 'Void:%')
            ->select('shift_id', 'reason');

        $rows = $this->transactionBaseQuery($scope)
            ->leftJoin('users as users', function ($join) use ($scope): void {
                $join->on('users.id', '=', 'transactions.cashier_id')
                    ->where('users.business_id', '=', $scope['business_id']);
            })
            ->leftJoin('audit_logs as audit', function ($join) use ($scope): void {
                $join->on('audit.entity_id', '=', 'transactions.id')
                    ->where('audit.business_id', '=', $scope['business_id'])
                    ->whereIn('audit.action', ['void', 'transaction.void']);
            })
            ->leftJoinSub($voidReasons, 'void_reasons', fn ($join) => $join->on('void_reasons.shift_id', '=', 'transactions.shift_id'))
            ->where('transactions.status', 'voided')
            ->select([
                'transactions.number',
                'transactions.created_at',
                'transactions.updated_at',
                'transactions.grand_total',
                'users.name as cashier_name',
                'audit.created_at as voided_at',
                'void_reasons.reason',
            ])
            ->orderByDesc('transactions.updated_at')
            ->paginate(50);

        $rows->getCollection()->transform(fn (object $row): array => [
            'transaction_number' => $row->number,
            'original_order_time' => $row->created_at,
            'void_or_refund_time' => $row->voided_at ?? $row->updated_at,
            'type' => 'void',
            'cashier' => $row->cashier_name,
            'authorizer' => null,
            'reason' => $this->normalizeVoidReason($row->reason),
            'order_type' => null,
            'amount' => (int) $row->grand_total,
            'refund_amount' => null,
        ]);

        return [
            'data' => ['rows' => $rows->items(), 'refund_rows' => []],
            'meta' => array_merge($this->meta($scope), ['pagination' => [
                'total' => $rows->total(),
                'per_page' => $rows->perPage(),
                'current_page' => $rows->currentPage(),
            ]]),
        ];
    }

    /**
     * @param  array<string, mixed>  $filters
     * @return array{data:array<string, mixed>, meta:array<string, mixed>}
     */
    public function topTen(User $user, array $filters): array
    {
        $scope = $this->resolveTransactionScope($user, $filters);

        $best = DB::table('transaction_items as items')
            ->join('transactions as transactions', function ($join) use ($scope): void {
                $join->on('transactions.id', '=', 'items.transaction_id')
                    ->where('transactions.business_id', '=', $scope['business_id']);
            })
            ->where('items.business_id', $scope['business_id'])
            ->where('transactions.status', 'paid')
            ->where('transactions.created_at', '>=', $scope['start'])
            ->where('transactions.created_at', '<', $scope['end'])
            ->when($scope['outlet_id'], fn (Builder $query, string $outletId): Builder => $query->where('transactions.outlet_id', $outletId))
            ->when($scope['cashier_id'], fn (Builder $query, string $cashierId): Builder => $query->where('transactions.cashier_id', $cashierId))
            ->when($scope['shift_id'], fn (Builder $query, string $shiftId): Builder => $query->where('transactions.shift_id', $shiftId))
            ->selectRaw('items.product_id, items.name, sum(items.quantity) as quantity_sold')
            ->groupBy('items.product_id', 'items.name')
            ->orderByDesc('quantity_sold')
            ->limit(10)
            ->get()
            ->map(fn (object $row): array => [
                'product_id' => $row->product_id,
                'product_name' => $row->name,
                'quantity_sold' => (int) $row->quantity_sold,
            ])
            ->all();

        return [
            'data' => [
                'best_selling' => $best,
                'highest_order_type_value' => [],
                'low_stock' => [],
                'highest_gross_profit' => [],
                'data_availability' => ['gross_profit' => 'cost_coverage_incomplete'],
            ],
            'meta' => $this->meta($scope),
        ];
    }

    /**
     * @param  array<string, mixed>  $filters
     * @return array{business_id:string, outlet_id:?string, shift_id:?string, cashier_id:?string, start:CarbonImmutable, end:CarbonImmutable, timezone:string, bucket:string}
     */
    private function resolveTransactionScope(User $user, array $filters): array
    {
        $businessId = (string) $user->business_id;
        $outletId = isset($filters['outlet_id']) ? (string) $filters['outlet_id'] : null;

        if ($outletId !== null && ! $this->outletBelongsToBusiness($businessId, $outletId)) {
            throw new ReportAccessException('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        $timezone = $outletId !== null
            ? $this->clock->outletTimezone($businessId, $outletId)
            : BusinessClock::DEFAULT_TIMEZONE;
        [$start, $end, $bucket] = $this->dateWindow($filters['date'] ?? null, $filters['range'] ?? 'day', $timezone);
        $shiftId = isset($filters['shift_id']) ? (string) $filters['shift_id'] : null;
        $cashierId = null;

        if ($this->isCashierLimited((string) $user->role)) {
            [$start, $end] = $this->clock->utcDayWindow($timezone, $filters['date'] ?? $this->clock->localNow($timezone)->toDateString());
            $bucket = 'hour';
            $currentShift = $this->currentShiftForCashier((string) $user->id, $businessId, $outletId, $start, $end);

            if ($shiftId !== null && (! $currentShift || $shiftId !== $currentShift->id)) {
                throw new ReportAccessException('FORBIDDEN', 'Cashier reports are limited to the current shift.', [], 403);
            }

            $cashierId = (string) $user->id;
            $shiftId = $currentShift?->id;
            if ($currentShift && $outletId === null) {
                $outletId = $currentShift->outlet_id;
                $timezone = $this->clock->outletTimezone($businessId, $outletId);
                [$start, $end] = $this->clock->utcDayWindow($timezone, $filters['date'] ?? $this->clock->localNow($timezone)->toDateString());
            }
        } elseif ($shiftId !== null && ! $this->shiftBelongsToBusiness($businessId, $shiftId)) {
            throw new ReportAccessException('NOT_FOUND', 'Shift not found.', [], 404);
        }

        return [
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'shift_id' => $shiftId,
            'cashier_id' => $cashierId,
            'start' => $start,
            'end' => $end,
            'timezone' => $timezone,
            'bucket' => $bucket,
        ];
    }

    /**
     * @param  array{business_id:string, outlet_id:?string, shift_id:?string, cashier_id:?string, start:CarbonImmutable, end:CarbonImmutable, timezone:string, bucket:string}  $scope
     */
    private function transactionBaseQuery(array $scope): Builder
    {
        return DB::table('transactions')
            ->where('transactions.business_id', $scope['business_id'])
            ->where('transactions.created_at', '>=', $scope['start'])
            ->where('transactions.created_at', '<', $scope['end'])
            ->when($scope['outlet_id'], fn (Builder $query, string $outletId): Builder => $query->where('transactions.outlet_id', $outletId))
            ->when($scope['cashier_id'], fn (Builder $query, string $cashierId): Builder => $query->where('transactions.cashier_id', $cashierId))
            ->when($scope['shift_id'], fn (Builder $query, string $shiftId): Builder => $query->where('transactions.shift_id', $shiftId));
    }

    private function outletBelongsToBusiness(string $businessId, string $outletId): bool
    {
        return DB::table('outlets')
            ->where('business_id', $businessId)
            ->where('id', $outletId)
            ->exists();
    }

    private function shiftBelongsToBusiness(string $businessId, string $shiftId): bool
    {
        return DB::table('shift_sessions')
            ->where('business_id', $businessId)
            ->where('id', $shiftId)
            ->exists();
    }

    private function currentShiftForCashier(string $cashierId, string $businessId, ?string $outletId, CarbonImmutable $start, CarbonImmutable $end): ?object
    {
        return DB::table('shift_sessions')
            ->where('business_id', $businessId)
            ->where('cashier_id', $cashierId)
            ->where('status', 'open')
            ->when($outletId, fn (Builder $query, string $filterOutletId): Builder => $query->where('outlet_id', $filterOutletId))
            ->where('opened_at', '>=', $start)
            ->where('opened_at', '<', $end)
            ->latest('opened_at')
            ->first();
    }

    private function dateWindow(?string $date, string $range, string $timezone): array
    {
        $anchor = CarbonImmutable::parse($date ?? $this->clock->localNow($timezone)->toDateString(), $timezone);

        return match ($range) {
            'week' => [$anchor->startOfWeek()->utc(), $anchor->startOfWeek()->addWeek()->utc(), 'day'],
            'month' => [$anchor->startOfMonth()->utc(), $anchor->startOfMonth()->addMonth()->utc(), 'day'],
            default => [...$this->clock->utcDayWindow($timezone, $anchor), 'hour'],
        };
    }

    private function isCashierLimited(string $role): bool
    {
        return ! in_array($role, ['owner', 'admin'], true);
    }

    private function paymentTotals(string $businessId, Collection $transactionIds): array
    {
        if ($transactionIds->isEmpty()) {
            return [];
        }

        return DB::table('payments')
            ->where('business_id', $businessId)
            ->whereIn('transaction_id', $transactionIds->all())
            ->whereIn('status', ['confirmed', 'settled'])
            ->select('method', DB::raw('sum(amount) as amount'))
            ->groupBy('method')
            ->orderBy('method')
            ->get()
            ->map(fn (object $row): array => [
                'method' => $row->method,
                'amount' => (int) $row->amount,
            ])
            ->values()
            ->all();
    }

    private function chartSeries(Collection $transactions, string $bucket, string $timezone): array
    {
        return $transactions
            ->groupBy(fn (object $row): string => CarbonImmutable::parse($row->created_at, 'UTC')->setTimezone($timezone)->format($bucket === 'hour' ? 'Y-m-d H:00' : 'Y-m-d'))
            ->map(fn (Collection $rows, string $label): array => [
                'label' => $label,
                'amount' => (int) $rows->sum(fn (object $row): int => (int) $row->grand_total),
                'transaction_count' => $rows->count(),
            ])
            ->sortKeys()
            ->values()
            ->all();
    }

    private function normalizeVoidReason(?string $reason): ?string
    {
        if ($reason === null) {
            return null;
        }

        return trim(preg_replace('/^Void:\s*/', '', $reason) ?? $reason);
    }

    /**
     * @param  array{business_id:string, outlet_id:?string, shift_id:?string, cashier_id:?string, start:CarbonImmutable, end:CarbonImmutable, timezone:string, bucket:string}  $scope
     * @return array<string, mixed>
     */
    private function meta(array $scope): array
    {
        return [
            'server_time' => now()->toISOString(),
            'timezone' => $scope['timezone'],
            'window' => [
                'start' => $scope['start']->toISOString(),
                'end' => $scope['end']->toISOString(),
            ],
        ];
    }
}
