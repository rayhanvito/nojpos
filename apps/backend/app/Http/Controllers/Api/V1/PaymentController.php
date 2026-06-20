<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\Payment;
use App\Services\PaymentService;
use App\Services\PaymentStateException;
use App\Services\TerminalContextException;
use App\Services\TerminalContextService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class PaymentController extends Controller
{
    public function __construct(
        private readonly TerminalContextService $terminalContext,
        private readonly PaymentService $payments,
    ) {}

    public function store(Request $request): JsonResponse
    {
        if ($request->boolean('confirm')) {
            return $this->confirm($request);
        }

        $data = $request->validate([
            'transaction_id' => ['required', 'uuid'],
            'method' => ['required', 'string'],
            'amount' => ['required', 'integer', 'min:0'],
            'reference' => ['nullable', 'string'],
        ]);

        $transaction = DB::table('transactions')
            ->where('id', $data['transaction_id'])
            ->where('business_id', $request->user()->business_id)
            ->first();

        if (! $transaction) {
            return ApiResponse::error('NOT_FOUND', 'Transaction not found.', [], 404);
        }

        try {
            $this->terminalContext->assertTransactionContext($request, $transaction, requireOpenShift: true);
        } catch (TerminalContextException $error) {
            return $this->terminalContextError($error);
        }

        try {
            $payment = $this->payments->create(
                $transaction->business_id,
                $transaction->id,
                $data,
                $request->user()->id,
            );
        } catch (PaymentStateException $error) {
            return $this->paymentStateError($error);
        }

        return ApiResponse::success($this->paymentPayload($payment), [], 201);
    }

    private function confirm(Request $request): JsonResponse
    {
        $data = $request->validate([
            'payment_id' => ['required', 'uuid'],
            'confirm' => ['required', 'boolean'],
        ]);

        $payment = DB::table('payments')
            ->where('business_id', $request->user()->business_id)
            ->where('id', $data['payment_id'])
            ->first();

        if (! $payment) {
            return ApiResponse::error('NOT_FOUND', 'Payment not found.', [], 404);
        }

        $transaction = DB::table('transactions')
            ->where('business_id', $request->user()->business_id)
            ->where('id', $payment->transaction_id)
            ->first();

        if (! $transaction) {
            return ApiResponse::error('NOT_FOUND', 'Transaction not found.', [], 404);
        }

        try {
            $this->terminalContext->assertTransactionContext($request, $transaction, requireOpenShift: true);
        } catch (TerminalContextException $error) {
            return $this->terminalContextError($error);
        }

        try {
            $confirmed = $this->payments->confirm(
                $request->user()->business_id,
                $payment->id,
                $request->user()->id,
            );
        } catch (PaymentStateException $error) {
            return $this->paymentStateError($error);
        }

        return ApiResponse::success($this->paymentPayload($confirmed));
    }

    private function terminalContextError(TerminalContextException $error): JsonResponse
    {
        return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
    }

    private function paymentStateError(PaymentStateException $error): JsonResponse
    {
        return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
    }

    private function paymentPayload(Payment $payment): array
    {
        return [
            'id' => $payment->id,
            'business_id' => $payment->business_id,
            'transaction_id' => $payment->transaction_id,
            'method' => $payment->method,
            'reference' => $payment->reference,
            'amount' => (int) $payment->amount,
            'status' => $payment->status,
            'is_cash' => (bool) $payment->is_cash,
            'confirmed_by' => $payment->confirmed_by,
            'confirmed_at' => $payment->confirmed_at,
        ];
    }
}
