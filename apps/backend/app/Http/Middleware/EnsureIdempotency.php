<?php

namespace App\Http\Middleware;

use App\Services\IdempotencyService;
use App\Support\ApiResponse;
use Closure;
use Illuminate\Http\Request;
use Illuminate\Support\Str;
use Symfony\Component\HttpFoundation\Response;

class EnsureIdempotency
{
    public function __construct(private readonly IdempotencyService $idempotency) {}

    public function handle(Request $request, Closure $next): Response
    {
        $key = $request->header('Idempotency-Key');

        if (! $key) {
            return ApiResponse::error('IDEMPOTENCY_KEY_REQUIRED', 'Idempotency-Key header is required.', [], 422);
        }

        if (! Str::isUuid($key)) {
            return ApiResponse::error('IDEMPOTENCY_KEY_INVALID', 'Idempotency-Key must be a UUID.', [], 422);
        }

        $reservation = $this->idempotency->reserveOrReplay($request);

        if ($reservation['action'] === 'response') {
            return $reservation['response'];
        }

        $request->attributes->set('idempotency_key', $key);
        $request->attributes->set('idempotency_record_id', $reservation['record']->id);

        $response = $next($request);

        if ($response->headers->get('Content-Type') && str_contains($response->headers->get('Content-Type'), 'application/json')) {
            $this->idempotency->complete($reservation['record'], $response);
        }

        return $response;
    }
}
