<?php

namespace App\Services;

use App\Models\IdempotencyKey;
use App\Models\Scopes\BusinessScope;
use App\Support\ApiResponse;
use Illuminate\Database\QueryException;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;
use Symfony\Component\HttpFoundation\Response;

class IdempotencyService
{
    public const SYSTEM_BUSINESS_ID = '00000000-0000-4000-8000-000000000000';

    /**
     * @return array{action: 'reserved', record: IdempotencyKey}|array{action: 'response', response: Response}
     */
    public function reserveOrReplay(Request $request): array
    {
        $businessId = $this->businessId($request);
        $endpoint = $this->endpoint($request);
        $key = (string) $request->header('Idempotency-Key');
        $hash = $this->hashPayload($request->all());

        try {
            return DB::transaction(function () use ($businessId, $endpoint, $key, $hash): array {
                $stored = $this->query()
                    ->where('business_id', $businessId)
                    ->where('endpoint', $endpoint)
                    ->where('key', $key)
                    ->lockForUpdate()
                    ->first();

                if ($stored) {
                    return $this->storedResult($stored, $hash);
                }

                $record = $this->query()->create([
                    'business_id' => $businessId,
                    'endpoint' => $endpoint,
                    'key' => $key,
                    'request_hash' => $hash,
                    'state' => 'in_progress',
                    'reserved_at' => now(),
                    'response_snapshot' => '',
                    'status' => 0,
                ]);

                return ['action' => 'reserved', 'record' => $record];
            });
        } catch (QueryException $exception) {
            if (! $this->isUniqueConstraintViolation($exception)) {
                throw $exception;
            }

            return DB::transaction(function () use ($businessId, $endpoint, $key, $hash): array {
                $stored = $this->query()
                    ->where('business_id', $businessId)
                    ->where('endpoint', $endpoint)
                    ->where('key', $key)
                    ->lockForUpdate()
                    ->firstOrFail();

                return $this->storedResult($stored, $hash);
            });
        }
    }

    public function complete(IdempotencyKey $record, Response $response): void
    {
        $this->query()
            ->where('id', $record->id)
            ->where('state', 'in_progress')
            ->update([
                'state' => 'completed',
                'response_snapshot' => $response->getContent() ?: '',
                'status' => $response->getStatusCode(),
                'completed_at' => now(),
                'updated_at' => now(),
            ]);
    }

    public function hashPayload(array $payload): string
    {
        $this->sortRecursive($payload);

        return hash('sha256', json_encode($payload, JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE | JSON_PRESERVE_ZERO_FRACTION));
    }

    public function endpoint(Request $request): string
    {
        return $request->method().' '.$request->path();
    }

    public function businessId(Request $request): string
    {
        return $request->user()?->business_id ?: self::SYSTEM_BUSINESS_ID;
    }

    /**
     * @return array{status: 'missing'|'in_progress'|'completed', response?: array}
     */
    public function checkoutRecovery(string $businessId, string $key): array
    {
        $record = $this->query()
            ->where('business_id', $businessId)
            ->where('endpoint', 'POST api/v1/transactions')
            ->where('key', $key)
            ->first();

        if (! $record) {
            return ['status' => 'missing'];
        }

        if (($record->state ?? 'completed') !== 'completed') {
            return ['status' => 'in_progress'];
        }

        $response = json_decode((string) $record->response_snapshot, true);

        return [
            'status' => 'completed',
            'response' => is_array($response) ? $response : [],
        ];
    }

    private function storedResult(IdempotencyKey $stored, string $hash): array
    {
        if ($stored->request_hash !== $hash) {
            return [
                'action' => 'response',
                'response' => ApiResponse::error(
                    'IDEMPOTENCY_MISMATCH',
                    'Idempotency key was reused with a different request payload.',
                    [],
                    409,
                ),
            ];
        }

        if (($stored->state ?? 'completed') !== 'completed') {
            return [
                'action' => 'response',
                'response' => ApiResponse::error(
                    'IDEMPOTENCY_IN_PROGRESS',
                    'A request with this idempotency key is still in progress.',
                    [],
                    409,
                ),
            ];
        }

        return [
            'action' => 'response',
            'response' => response($stored->response_snapshot, $stored->status)
                ->header('Content-Type', 'application/json'),
        ];
    }

    private function query()
    {
        return IdempotencyKey::withoutGlobalScope(BusinessScope::class);
    }

    private function isUniqueConstraintViolation(QueryException $exception): bool
    {
        $sqlState = (string) ($exception->errorInfo[0] ?? '');
        $driverCode = (string) ($exception->errorInfo[1] ?? '');
        $message = Str::lower($exception->getMessage());

        return in_array($sqlState, ['23000', '23505'], true)
            || $driverCode === '1062'
            || str_contains($message, 'unique constraint')
            || str_contains($message, 'duplicate entry');
    }

    private function sortRecursive(array &$value): void
    {
        foreach ($value as &$item) {
            if (is_array($item)) {
                $this->sortRecursive($item);
            }
        }

        if (! array_is_list($value)) {
            ksort($value);
        }
    }
}
