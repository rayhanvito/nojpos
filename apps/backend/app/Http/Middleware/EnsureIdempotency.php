<?php

namespace App\Http\Middleware;

use App\Models\IdempotencyKey;
use App\Support\ApiResponse;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class EnsureIdempotency
{
    public function handle(Request $request, Closure $next): Response
    {
        $key = $request->header('Idempotency-Key');

        if (! $key) {
            return ApiResponse::error('IDEMPOTENCY_KEY_REQUIRED', 'Idempotency-Key header is required.', [], 422);
        }

        $businessId = $request->user()?->business_id;
        $endpoint = $request->method().' '.$request->path();
        $hash = hash('sha256', json_encode($request->all(), JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE));

        $stored = IdempotencyKey::query()
            ->where('business_id', $businessId)
            ->where('endpoint', $endpoint)
            ->where('key', $key)
            ->first();

        if ($stored) {
            if ($stored->request_hash !== $hash) {
                return ApiResponse::error(
                    'IDEMPOTENCY_CONFLICT',
                    'Idempotency key was reused with a different request payload.',
                    [],
                    409,
                );
            }

            return response($stored->response_snapshot, $stored->status)
                ->header('Content-Type', 'application/json');
        }

        $response = $next($request);

        if ($response->headers->get('Content-Type') && str_contains($response->headers->get('Content-Type'), 'application/json')) {
            IdempotencyKey::query()->create([
                'business_id' => $businessId,
                'endpoint' => $endpoint,
                'key' => $key,
                'request_hash' => $hash,
                'response_snapshot' => $response->getContent(),
                'status' => $response->getStatusCode(),
            ]);
        }

        return $response;
    }
}
