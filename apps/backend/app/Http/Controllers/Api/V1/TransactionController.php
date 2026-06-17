<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Support\ApiResponse;
use App\Support\Nojpos;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class TransactionController extends Controller
{
    public function index(Request $request): JsonResponse
    {
        $transactions = DB::table('transactions')
            ->where('business_id', $request->user()->business_id)
            ->orderByDesc('created_at')
            ->get()
            ->map(fn ($transaction): array => $this->transactionPayload((array) $transaction))
            ->values()
            ->all();

        return ApiResponse::success(['transactions' => $transactions]);
    }

    public function store(Request $request): JsonResponse
    {
        $data = $request->validate([
            'outlet_id' => ['required', 'uuid'],
            'device_id' => ['required', 'uuid'],
            'cashier_id' => ['required', 'uuid'],
            'shift_id' => ['required', 'uuid'],
            'cart_discount' => ['nullable', 'integer', 'min:0'],
            'rounding' => ['nullable', 'integer'],
            'notes' => ['nullable', 'string'],
            'items' => ['required', 'array', 'min:1'],
            'items.*.product_id' => ['required', 'uuid'],
            'items.*.quantity' => ['required', 'integer', 'min:1'],
            'items.*.unit_price' => ['required', 'integer', 'min:0'],
            'items.*.discount' => ['nullable', 'integer', 'min:0'],
            'payments' => ['required', 'array', 'min:1'],
            'payments.*.method' => ['required', 'string'],
            'payments.*.amount' => ['required', 'integer', 'min:0'],
            'payments.*.reference' => ['nullable', 'string'],
        ]);

        $businessId = $request->user()->business_id;
        if (! $this->operationalContextIsValid($businessId, $data)) {
            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        $shift = DB::table('shift_sessions')
            ->where('id', $data['shift_id'])
            ->where('business_id', $businessId)
            ->where('status', 'open')
            ->first();

        if (! $shift) {
            return ApiResponse::error('SHIFT_NOT_OPEN', 'Shift must be open before checkout.', [], 422);
        }

        $outlet = DB::table('outlets')->where('id', $data['outlet_id'])->where('business_id', $businessId)->first();

        $subtotal = 0;
        $itemDiscountTotal = 0;
        $items = [];

        foreach ($data['items'] as $item) {
            $product = DB::table('products')
                ->where('id', $item['product_id'])
                ->where('business_id', $businessId)
                ->first();

            if (! $product) {
                return ApiResponse::error('PRODUCT_NOT_FOUND', 'Product is not available for this business.', [], 404);
            }

            $quantity = (int) $item['quantity'];
            $unitPrice = (int) $item['unit_price'];
            $discount = (int) ($item['discount'] ?? 0);
            $lineSubtotal = $quantity * $unitPrice;
            $subtotal += $lineSubtotal;
            $itemDiscountTotal += $discount;

            $items[] = [
                'product' => $product,
                'quantity' => $quantity,
                'unit_price' => $unitPrice,
                'discount' => $discount,
                'subtotal' => $lineSubtotal,
            ];
        }

        $cartDiscount = (int) ($data['cart_discount'] ?? 0);
        $discountTotal = $itemDiscountTotal + $cartDiscount;
        $taxableBase = max(0, $subtotal - $discountTotal);
        $serviceCharge = intdiv($taxableBase * (int) $outlet->service_charge_rate, 100);
        $tax = intdiv(($taxableBase + $serviceCharge) * (int) $outlet->tax_rate, 100);
        $rounding = (int) ($data['rounding'] ?? 0);
        $grandTotal = $taxableBase + $serviceCharge + $tax + $rounding;

        $paymentRows = [];
        $confirmedTotal = 0;
        $hasConfirmedPayment = false;
        foreach ($data['payments'] as $payment) {
            $isCash = $this->isCashMethod($businessId, $data['outlet_id'], $payment['method']);
            $status = $isCash ? 'confirmed' : 'pending';
            if ($status === 'confirmed') {
                $confirmedTotal += (int) $payment['amount'];
                $hasConfirmedPayment = true;
            }
            $paymentRows[] = [
                'method' => $payment['method'],
                'reference' => $payment['reference'] ?? null,
                'amount' => (int) $payment['amount'],
                'status' => $status,
                'is_cash' => $isCash,
            ];
        }

        $status = ($confirmedTotal >= $grandTotal || $hasConfirmedPayment) ? 'paid' : 'pending';
        $transactionId = (string) Str::uuid();
        $now = now();

        DB::transaction(function () use ($businessId, $data, $request, $transactionId, $subtotal, $discountTotal, $serviceCharge, $tax, $rounding, $grandTotal, $status, $items, $paymentRows, $now): void {
            DB::table('transactions')->insert([
                'id' => $transactionId,
                'business_id' => $businessId,
                'outlet_id' => $data['outlet_id'],
                'device_id' => $data['device_id'],
                'cashier_id' => $data['cashier_id'],
                'shift_id' => $data['shift_id'],
                'number' => 'TRX-'.now()->format('YmdHis').'-'.Str::upper(Str::random(4)),
                'status' => $status,
                'subtotal' => $subtotal,
                'discount_total' => $discountTotal,
                'service_charge_total' => $serviceCharge,
                'tax_total' => $tax,
                'rounding_total' => $rounding,
                'grand_total' => $grandTotal,
                'notes' => $data['notes'] ?? null,
                'created_at' => $now,
                'updated_at' => $now,
            ]);

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

                if ($status === 'paid' && (bool) $item['product']->track_stock) {
                    DB::table('stock_movements')->insert([
                        'id' => (string) Str::uuid(),
                        'business_id' => $businessId,
                        'outlet_id' => $data['outlet_id'],
                        'product_id' => $item['product']->id,
                        'transaction_id' => $transactionId,
                        'type' => 'sale',
                        'quantity_delta' => -1 * $item['quantity'],
                        'created_at' => $now,
                        'updated_at' => $now,
                    ]);
                }
            }

            foreach ($paymentRows as $payment) {
                DB::table('payments')->insert([
                    'id' => (string) Str::uuid(),
                    'business_id' => $businessId,
                    'transaction_id' => $transactionId,
                    'method' => $payment['method'],
                    'reference' => $payment['reference'],
                    'amount' => $payment['amount'],
                    'status' => $payment['status'],
                    'is_cash' => $payment['is_cash'],
                    'confirmed_by' => $payment['status'] === 'confirmed' ? $request->user()->id : null,
                    'confirmed_at' => $payment['status'] === 'confirmed' ? $now : null,
                    'created_at' => $now,
                    'updated_at' => $now,
                ]);
            }
        });

        Nojpos::audit($businessId, $request->user()->id, 'transaction.create', 'transaction', $transactionId);

        $transaction = (array) DB::table('transactions')->where('id', $transactionId)->first();

        return ApiResponse::success($this->transactionPayload($transaction), [], 201);
    }

    private function operationalContextIsValid(string $businessId, array $data): bool
    {
        foreach (['outlets' => 'outlet_id', 'devices' => 'device_id', 'users' => 'cashier_id', 'shift_sessions' => 'shift_id'] as $table => $field) {
            if (! DB::table($table)->where('id', $data[$field])->where('business_id', $businessId)->exists()) {
                return false;
            }
        }

        return true;
    }

    private function isCashMethod(string $businessId, string $outletId, string $method): bool
    {
        $config = DB::table('payment_method_configs')
            ->where('business_id', $businessId)
            ->where('method', $method)
            ->where(function ($query) use ($outletId): void {
                $query->whereNull('outlet_id')->orWhere('outlet_id', $outletId);
            })
            ->first();

        return $config ? (bool) $config->is_cash : $method === 'cash';
    }

    private function transactionPayload(array $transaction): array
    {
        return [
            'id' => $transaction['id'],
            'business_id' => $transaction['business_id'],
            'outlet_id' => $transaction['outlet_id'],
            'device_id' => $transaction['device_id'],
            'cashier_id' => $transaction['cashier_id'],
            'shift_id' => $transaction['shift_id'],
            'number' => $transaction['number'],
            'status' => $transaction['status'],
            'subtotal' => (int) $transaction['subtotal'],
            'discount_total' => (int) $transaction['discount_total'],
            'service_charge_total' => (int) $transaction['service_charge_total'],
            'tax_total' => (int) $transaction['tax_total'],
            'rounding_total' => (int) $transaction['rounding_total'],
            'grand_total' => (int) $transaction['grand_total'],
            'notes' => $transaction['notes'] ?? null,
        ];
    }
}
