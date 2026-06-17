<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\Payment;
use App\Support\ApiResponse;
use App\Support\Nojpos;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class PaymentController extends Controller
{
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

        $isCash = $this->isCashMethod($transaction->business_id, $transaction->outlet_id, $data['method']);
        $status = $isCash ? 'confirmed' : 'pending';

        $payment = Payment::query()->create([
            'business_id' => $transaction->business_id,
            'transaction_id' => $transaction->id,
            'method' => $data['method'],
            'reference' => $data['reference'] ?? null,
            'amount' => $data['amount'],
            'status' => $status,
            'is_cash' => $isCash,
            'confirmed_by' => $status === 'confirmed' ? $request->user()->id : null,
            'confirmed_at' => $status === 'confirmed' ? now() : null,
        ]);

        Nojpos::audit($transaction->business_id, $request->user()->id, 'payment.create', 'payment', $payment->id);

        return ApiResponse::success($this->paymentPayload($payment), [], 201);
    }

    private function confirm(Request $request): JsonResponse
    {
        $data = $request->validate([
            'payment_id' => ['required', 'uuid'],
            'confirm' => ['required', 'boolean'],
        ]);

        $payment = Payment::query()->where('id', $data['payment_id'])->firstOrFail();
        $payment->update([
            'status' => 'confirmed',
            'confirmed_by' => $request->user()->id,
            'confirmed_at' => now(),
        ]);

        Nojpos::audit($payment->business_id, $request->user()->id, 'payment.confirm', 'payment', $payment->id);

        return ApiResponse::success($this->paymentPayload($payment->refresh()));
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
        ];
    }
}
