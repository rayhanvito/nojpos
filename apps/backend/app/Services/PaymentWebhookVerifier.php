<?php

namespace App\Services;

use Illuminate\Http\Request;

class PaymentWebhookVerifier
{
    public function verify(Request $request, string $provider): bool
    {
        $secret = config("services.payment_webhooks.{$provider}.secret");
        $signature = (string) $request->header('X-Payment-Signature');

        return is_string($secret) && $secret !== '' && $signature !== ''
            && hash_equals(hash_hmac('sha256', $request->getContent(), $secret), $signature);
    }
}
