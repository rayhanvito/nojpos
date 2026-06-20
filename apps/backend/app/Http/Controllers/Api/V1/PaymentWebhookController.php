<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Services\PaymentService;
use App\Services\PaymentWebhookVerifier;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class PaymentWebhookController extends Controller
{
    public function __construct(private readonly PaymentWebhookVerifier $verifier, private readonly PaymentService $payments) {}

    public function store(Request $request, string $provider): JsonResponse
    {
        if (! $this->verifier->verify($request, $provider)) {
            return ApiResponse::error('WEBHOOK_SIGNATURE_INVALID', 'Invalid webhook signature.', [], 403);
        }

        $data = $request->validate(['event_id' => ['required', 'string', 'max:191'], 'payment_ref' => ['required', 'uuid'], 'outcome' => ['required', 'in:confirmed,failed,declined']]);
        $existing = DB::table('payment_webhook_events')->where('provider', $provider)->where('event_id', $data['event_id'])->first();
        if ($existing) {
            return ApiResponse::success(['accepted' => true]);
        }

        $payment = $this->payments->settleWebhook($provider, $data['payment_ref'], $data['outcome']);
        DB::table('payment_webhook_events')->insert(['id' => (string) Str::uuid(), 'business_id' => $payment?->business_id, 'provider' => $provider, 'event_id' => $data['event_id'], 'payment_id' => $payment?->id, 'outcome' => $data['outcome'], 'received_at' => now(), 'created_at' => now(), 'updated_at' => now()]);

        return ApiResponse::success(['accepted' => true]);
    }
}
