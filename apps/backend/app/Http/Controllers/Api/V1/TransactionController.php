<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\CheckoutTransactionRequest;
use App\Services\BusinessClock;
use App\Services\IdempotencyService;
use App\Services\PaymentService;
use App\Services\PaymentStateException;
use App\Services\StoreOperationException;
use App\Services\StoreStateService;
use App\Services\TerminalContextException;
use App\Services\TerminalContextService;
use App\Services\TransactionQuoteService;
use App\Support\ApiResponse;
use App\Support\Nojpos;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class TransactionController extends Controller
{
    public function __construct(
        private readonly TransactionQuoteService $quoteService,
        private readonly TerminalContextService $terminalContext,
        private readonly PaymentService $payments,
        private readonly StoreStateService $stores,
        private readonly IdempotencyService $idempotency,
        private readonly BusinessClock $clock,
    ) {}

    public function index(Request $request): JsonResponse
    {
        $status = $request->query('status');
        $transactions = DB::table('transactions')
            ->where('business_id', $request->user()->business_id)
            ->when($status && $status !== 'pending', fn ($query) => $query->where('status', $status))
            ->when($status === 'pending', function ($query): void {
                $query->whereExists(function ($query): void {
                    $query->selectRaw('1')
                        ->from('payments')
                        ->whereColumn('payments.transaction_id', 'transactions.id')
                        ->where('payments.status', 'pending');
                });
            })
            ->orderByDesc('created_at')
            ->get()
            ->map(fn ($transaction): array => $this->transactionPayload((array) $transaction))
            ->values()
            ->all();

        return ApiResponse::success(['transactions' => $transactions]);
    }

    public function recovery(Request $request, string $key): JsonResponse
    {
        $recovery = $this->idempotency->checkoutRecovery(
            $request->user()->business_id,
            $key,
        );

        return ApiResponse::success([
            'status' => $recovery['status'],
            'transaction' => $recovery['response']['data'] ?? null,
        ]);
    }

    public function show(Request $request, string $transaction): JsonResponse
    {
        $this->payments->expireTransaction($request->user()->business_id, $transaction);
        $row = DB::table('transactions')->where('business_id', $request->user()->business_id)->where('id', $transaction)->first();
        if (! $row) {
            return ApiResponse::error('NOT_FOUND', 'Transaction not found.', [], 404);
        }

        return ApiResponse::success($this->transactionPayload((array) $row, true));
    }

    public function quote(Request $request): JsonResponse
    {
        $data = $request->validate($this->transactionValidationRules(requirePayments: false, includeContext: true));
        $businessId = $request->user()->business_id;

        try {
            $this->terminalContext->fromPayload($request, $data, requireOpenShift: true);
        } catch (TerminalContextException $error) {
            return $this->terminalContextError($error);
        }

        if (! $this->optionalRelationsAreValid($businessId, $data)) {
            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        try {
            $quote = $this->quoteService->createSnapshot($businessId, $data);
        } catch (\RuntimeException $error) {
            if ($error->getMessage() === 'PRODUCT_NOT_FOUND') {
                return ApiResponse::error('PRODUCT_NOT_FOUND', 'Product is not available for this business.', [], 404);
            }

            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        return ApiResponse::success($quote);
    }

    public function store(CheckoutTransactionRequest $request): JsonResponse
    {
        $data = $request->validated();

        $requestedStatus = $data['status'] ?? null;
        $isHeldOrUnpaid = in_array($requestedStatus, ['held', 'unpaid'], true);
        if (! $isHeldOrUnpaid && empty($data['payments'])) {
            return ApiResponse::error('VALIDATION_ERROR', 'The given data was invalid.', ['payments' => ['The payments field is required.']], 422);
        }

        $businessId = $request->user()->business_id;
        try {
            $this->terminalContext->fromPayload($request, $data, requireOpenShift: true);
        } catch (TerminalContextException $error) {
            return $this->terminalContextError($error);
        }

        if (! $this->optionalRelationsAreValid($businessId, $data)) {
            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        try {
            $this->stores->assertStoreAllowsCheckout($businessId, $data['outlet_id']);
        } catch (StoreOperationException $error) {
            return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
        }

        try {
            $quote = $this->quoteService->quoteForCheckout(
                $businessId,
                $data,
                $request->attributes->get('idempotency_key') ?: $request->header('Idempotency-Key'),
            );
        } catch (\RuntimeException $error) {
            if ($error->getMessage() === 'PRODUCT_NOT_FOUND') {
                return ApiResponse::error('PRODUCT_NOT_FOUND', 'Product is not available for this business.', [], 404);
            }

            if ($error->getMessage() === 'QUOTE_STALE') {
                return ApiResponse::error('QUOTE_STALE', 'Checkout quote is stale. Please request a new quote.', [], 409);
            }

            if ($error->getMessage() === 'QUOTE_REQUIRED') {
                return ApiResponse::error('QUOTE_REQUIRED', 'Checkout quote is required. Please request a quote before checkout.', [], 422);
            }

            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        if (empty($data['quote_id']) && ! $this->clientBreakdownMatchesQuote($data, $quote)) {
            return ApiResponse::error('TOTAL_MISMATCH', 'Checkout total does not match server quote.', [], 422);
        }

        $subtotal = $quote['subtotal'];
        $itemDiscountTotal = $quote['item_discount_total'];
        $cartDiscount = $quote['cart_discount_total'];
        $promotionDiscountTotal = $quote['promotion_discount_total'] ?? 0;
        $discountTotal = $quote['discount_total'];
        $serviceCharge = $quote['service_charge_total'];
        $tax = $quote['tax_total'];
        $rounding = $quote['rounding_total'];
        $grandTotal = $quote['grand_total'];
        $items = $quote['items'];

        try {
            $paymentAllocation = $this->payments->prepareCheckoutAllocations(
                $businessId,
                $data['outlet_id'],
                $data['payments'] ?? [],
                $grandTotal,
            );
        } catch (PaymentStateException $error) {
            return $this->paymentStateError($error);
        }

        $paymentRows = $paymentAllocation['rows'];
        $confirmedTotal = $paymentAllocation['confirmed_total'];
        $status = $isHeldOrUnpaid ? $requestedStatus : ($paymentAllocation['has_async'] ? 'payment_pending' : $this->transactionStatus($confirmedTotal, $grandTotal));
        $transactionId = (string) Str::uuid();
        $now = now();

        DB::transaction(function () use ($businessId, $data, $request, $transactionId, $subtotal, $itemDiscountTotal, $cartDiscount, $promotionDiscountTotal, $discountTotal, $serviceCharge, $tax, $rounding, $grandTotal, $status, $items, $paymentRows, $now): void {
            DB::table('transactions')->insert([
                'id' => $transactionId,
                'business_id' => $businessId,
                'outlet_id' => $data['outlet_id'],
                'device_id' => $data['device_id'],
                'cashier_id' => $data['cashier_id'],
                'shift_id' => $data['shift_id'],
                'customer_id' => $data['customer_id'] ?? null,
                'served_by' => $data['served_by'] ?? null,
                'number' => 'TRX-'.$this->clock->documentTimestamp($this->clock->outletTimezone($businessId, $data['outlet_id']), $now).'-'.Str::upper(Str::random(4)),
                'status' => $status,
                'subtotal' => $subtotal,
                'discount_total' => $discountTotal,
                'item_discount_total' => $itemDiscountTotal,
                'cart_discount_total' => $cartDiscount,
                'promotion_discount_total' => $promotionDiscountTotal,
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
                    'provider' => $payment['provider'],
                    'provider_reference' => $payment['provider_reference'],
                    'confirm_expires_at' => $payment['confirm_expires_at'],
                    'confirmed_by' => $payment['status'] === 'confirmed' ? $request->user()->id : null,
                    'confirmed_at' => $payment['status'] === 'confirmed' ? $now : null,
                    'created_at' => $now,
                    'updated_at' => $now,
                ]);
            }

            if (! empty($data['quote_id'])) {
                $this->quoteService->markUsed($businessId, $data['quote_id'], $transactionId);
            }

            Nojpos::audit($businessId, $request->user()->id, 'transaction.create', 'transaction', $transactionId);
            if ($status === 'payment_pending') {
                Nojpos::audit($businessId, $request->user()->id, 'transaction.payment_pending', 'transaction', $transactionId);
            }
        });

        $transaction = (array) DB::table('transactions')->where('id', $transactionId)->first();

        return ApiResponse::success($this->transactionPayload($transaction), [], 201);
    }

    private function terminalContextError(TerminalContextException $error): JsonResponse
    {
        return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
    }

    private function paymentStateError(PaymentStateException $error): JsonResponse
    {
        return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
    }

    private function optionalRelationsAreValid(string $businessId, array $data): bool
    {
        if (($data['customer_id'] ?? null) && ! DB::table('customers')->where('id', $data['customer_id'])->where('business_id', $businessId)->exists()) {
            return false;
        }

        if (($data['served_by'] ?? null) && ! DB::table('users')->where('id', $data['served_by'])->where('business_id', $businessId)->exists()) {
            return false;
        }

        return true;
    }

    private function transactionValidationRules(bool $requirePayments, bool $includeContext): array
    {
        return [
            'outlet_id' => ['required', 'uuid'],
            'device_id' => ['required', 'uuid'],
            'cashier_id' => ['required', 'uuid'],
            'shift_id' => ['required', 'uuid'],
            'customer_id' => ['nullable', 'uuid'],
            'served_by' => ['nullable', 'uuid'],
            'quote_id' => ['nullable', 'uuid'],
            'checkout_token' => ['nullable', 'string', 'size:64'],
            'quote_token' => ['nullable', 'string', 'size:64'],
            'status' => ['nullable', 'in:held,unpaid'],
            'promotion_codes' => ['nullable', 'array'],
            'promotion_codes.*' => ['string', 'max:64'],
            'applied_promotion_ids' => ['nullable', 'array'],
            'applied_promotion_ids.*' => ['uuid'],
            'manual_discount' => ['nullable', 'array'],
            'manual_discount.type' => ['required_with:manual_discount', 'in:amount,percent'],
            'manual_discount.value' => ['required_with:manual_discount', 'integer', 'min:0'],
            'manual_discount.reason' => ['nullable', 'string', 'max:255'],
            'cart_discount' => ['nullable', 'integer', 'min:0'],
            'cart_discount_type' => ['nullable', 'in:amount,percent'],
            'rounding' => ['nullable', 'integer'],
            'subtotal' => ['nullable', 'integer', 'min:0'],
            'item_discount_total' => ['nullable', 'integer', 'min:0'],
            'cart_discount_total' => ['nullable', 'integer', 'min:0'],
            'discount_total' => ['nullable', 'integer', 'min:0'],
            'service_charge_total' => ['nullable', 'integer', 'min:0'],
            'tax_total' => ['nullable', 'integer', 'min:0'],
            'rounding_total' => ['nullable', 'integer'],
            'grand_total' => ['nullable', 'integer', 'min:0'],
            'notes' => ['nullable', 'string'],
            'items' => ['required', 'array', 'min:1'],
            'items.*.product_id' => ['required', 'uuid'],
            'items.*.quantity' => ['required', 'integer', 'min:1'],
            'items.*.unit_price' => ['required', 'integer', 'min:0'],
            'items.*.discount' => ['nullable', 'integer', 'min:0'],
            'items.*.discount_type' => ['nullable', 'in:amount,percent'],
            'payments' => [$requirePayments ? 'required' : 'nullable', 'array'],
            'payments.*.method' => ['required', 'string'],
            'payments.*.amount' => ['required', 'integer', 'min:0'],
            'payments.*.reference' => ['nullable', 'string'],
        ];
    }

    private function clientBreakdownMatchesQuote(array $data, array $quote): bool
    {
        foreach ([
            'subtotal',
            'item_discount_total',
            'cart_discount_total',
            'promotion_discount_total',
            'discount_total',
            'service_charge_total',
            'tax_total',
            'rounding_total',
            'grand_total',
        ] as $field) {
            if (array_key_exists($field, $data) && (int) $data[$field] !== (int) $quote[$field]) {
                return false;
            }
        }

        return true;
    }

    private function transactionStatus(int $confirmedTotal, int $grandTotal): string
    {
        if ($confirmedTotal >= $grandTotal) {
            return 'paid';
        }

        return $confirmedTotal > 0 ? 'partial' : 'unpaid';
    }

    private function transactionPayload(array $transaction, bool $includeServerTime = false): array
    {
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
            'created_at' => $transaction['created_at'] ?? null,
            'updated_at' => $transaction['updated_at'] ?? null,
            'cashier' => $this->relatedUserPayload($transaction['cashier_id'], $transaction['business_id']),
            'customer' => $this->relatedCustomerPayload($transaction['customer_id'] ?? null, $transaction['business_id']),
            'items' => $this->transactionItems($transaction['id']),
            'payments' => $this->transactionPayments($transaction['id']),
            'receipt' => $this->receiptPayload($transaction),
            'server_time' => $includeServerTime ? now()->toIso8601String() : null,
            'allowed_next_actions' => $transaction['status'] === 'payment_pending' ? ['poll'] : [],
        ];
    }

    private function relatedUserPayload(?string $userId, ?string $businessId = null): ?array
    {
        if (! $userId) {
            return null;
        }

        $query = DB::table('users')->where('id', $userId);
        if ($businessId) {
            $query->where('business_id', $businessId);
        }

        $user = $query->first();
        if (! $user) {
            return null;
        }

        return [
            'id' => $user->id,
            'name' => $user->name,
            'email' => $user->email,
            'role' => $user->role,
        ];
    }

    private function relatedOutletPayload(?string $outletId, string $businessId): ?array
    {
        if (! $outletId) {
            return null;
        }

        $outlet = DB::table('outlets')
            ->where('id', $outletId)
            ->where('business_id', $businessId)
            ->first();

        if (! $outlet) {
            return null;
        }

        return [
            'id' => $outlet->id,
            'name' => $outlet->name,
            'timezone' => $outlet->timezone ?? 'Asia/Jakarta',
        ];
    }

    private function relatedCustomerPayload(?string $customerId, ?string $businessId = null): ?array
    {
        if (! $customerId) {
            return null;
        }

        $query = DB::table('customers')->where('id', $customerId);
        if ($businessId) {
            $query->where('business_id', $businessId);
        }

        $customer = $query->first();
        if (! $customer) {
            return null;
        }

        return [
            'id' => $customer->id,
            'name' => $customer->name,
            'phone' => $customer->phone,
            'group' => $customer->group,
        ];
    }

    private function receiptPayload(array $transaction): array
    {
        $items = $this->transactionItems($transaction['id']);
        $payments = $this->transactionPayments($transaction['id']);

        return [
            'transaction_id' => $transaction['id'],
            'transaction_number' => $transaction['number'],
            'number' => $transaction['number'],
            'issued_at' => $transaction['created_at'] ?? null,
            'updated_at' => $transaction['updated_at'] ?? null,
            'status' => $transaction['status'],
            'outlet' => $this->relatedOutletPayload($transaction['outlet_id'], $transaction['business_id']),
            'cashier' => $this->relatedUserPayload($transaction['cashier_id'], $transaction['business_id']),
            'customer' => $this->relatedCustomerPayload($transaction['customer_id'] ?? null, $transaction['business_id']),
            'items' => $items,
            'totals' => [
                'subtotal' => (int) $transaction['subtotal'],
                'item_discount_total' => (int) ($transaction['item_discount_total'] ?? 0),
                'cart_discount_total' => (int) ($transaction['cart_discount_total'] ?? 0),
                'promotion_discount_total' => (int) ($transaction['promotion_discount_total'] ?? 0),
                'discount_total' => (int) $transaction['discount_total'],
                'service_charge_total' => (int) $transaction['service_charge_total'],
                'tax_total' => (int) $transaction['tax_total'],
                'rounding_total' => (int) $transaction['rounding_total'],
                'grand_total' => (int) $transaction['grand_total'],
            ],
            'payments' => array_map(fn (array $payment): array => [
                'id' => $payment['id'],
                'method' => $payment['method'],
                'amount' => $payment['amount'],
                'status' => $payment['status'],
                'is_cash' => $payment['is_cash'],
                'confirmed_at' => $payment['confirmed_at'],
            ], $payments),
        ];
    }

    private function transactionItems(string $transactionId): array
    {
        return DB::table('transaction_items')
            ->where('transaction_id', $transactionId)
            ->orderBy('created_at')
            ->get()
            ->map(fn ($item): array => [
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
    }

    private function transactionPayments(string $transactionId): array
    {
        return DB::table('payments')
            ->where('transaction_id', $transactionId)
            ->orderBy('created_at')
            ->get()
            ->map(fn ($payment): array => [
                'id' => $payment->id,
                'business_id' => $payment->business_id,
                'transaction_id' => $payment->transaction_id,
                'method' => $payment->method,
                'reference' => $payment->reference,
                'payment_ref' => $payment->provider_reference,
                'provider' => $payment->provider,
                'amount' => (int) $payment->amount,
                'status' => $payment->status,
                'is_cash' => (bool) $payment->is_cash,
                'confirmed_by' => $payment->confirmed_by,
                'confirmed_at' => $payment->confirmed_at,
                'confirm_expires_at' => $payment->confirm_expires_at,
                'failure_reason' => $payment->failure_reason,
            ])
            ->values()
            ->all();
    }
}
