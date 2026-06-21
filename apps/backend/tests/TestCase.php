<?php

namespace Tests;

use Illuminate\Foundation\Testing\TestCase as BaseTestCase;
use Illuminate\Support\Str;
use Illuminate\Testing\TestResponse;

abstract class TestCase extends BaseTestCase
{
    public function postJson($uri, array $data = [], array $headers = [], $options = 0): TestResponse
    {
        if ($this->needsDefaultIdempotencyKey('POST', (string) $uri, $headers)) {
            $headers['Idempotency-Key'] = (string) Str::uuid();
        }

        return parent::postJson($uri, $data, $headers, $options);
    }

    public function putJson($uri, array $data = [], array $headers = [], $options = 0): TestResponse
    {
        if ($this->needsDefaultIdempotencyKey('PUT', (string) $uri, $headers)) {
            $headers['Idempotency-Key'] = (string) Str::uuid();
        }

        return parent::putJson($uri, $data, $headers, $options);
    }

    public function deleteJson($uri, array $data = [], array $headers = [], $options = 0): TestResponse
    {
        if ($this->needsDefaultIdempotencyKey('DELETE', (string) $uri, $headers)) {
            $headers['Idempotency-Key'] = (string) Str::uuid();
        }

        return parent::deleteJson($uri, $data, $headers, $options);
    }

    private function needsDefaultIdempotencyKey(string $method, string $uri, array $headers): bool
    {
        $headerKeys = array_change_key_case($headers, CASE_LOWER);
        $defaultHeaderKeys = array_change_key_case($this->defaultHeaders, CASE_LOWER);

        if (array_key_exists('idempotency-key', $headerKeys) || array_key_exists('idempotency-key', $defaultHeaderKeys)) {
            return false;
        }

        if ($method === 'POST' && in_array($uri, [
            '/api/v1/shifts/open',
            '/api/v1/inventory/purchases',
            '/api/v1/attendance',
            '/api/v1/products',
            '/api/v1/categories',
            '/api/v1/customers',
            '/api/v1/staff',
        ], true)) {
            return true;
        }

        if (in_array($method, ['PUT', 'DELETE'], true) && (
            str_starts_with($uri, '/api/v1/products/')
            || str_starts_with($uri, '/api/v1/categories/')
            || str_starts_with($uri, '/api/v1/customers/')
            || str_starts_with($uri, '/api/v1/staff/')
        )) {
            return true;
        }

        if ($method === 'POST' && str_ends_with($uri, '/delete') && str_starts_with($uri, '/api/v1/staff/')) {
            return true;
        }

        return false;
    }
}
