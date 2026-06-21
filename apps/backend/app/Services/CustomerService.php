<?php

namespace App\Services;

use App\Models\Customer;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class CustomerService
{
    public function __construct(private readonly AuditLogService $auditLog) {}

    public function create(Request $request, array $data): array
    {
        $businessId = $request->user()->business_id;
        $actorId = $request->user()->id;

        return DB::transaction(function () use ($request, $data, $businessId, $actorId): array {
            $customer = Customer::query()->create([
                'business_id' => $businessId,
                'name' => $data['name'],
                'phone' => $data['phone'] ?? null,
                'group' => $data['group'] ?? 'Tanpa Grup',
            ]);

            $payload = $this->payload($customer->id, $businessId);
            $this->auditLog->record(
                $businessId,
                $actorId,
                'customer.created',
                'customer',
                $customer->id,
                null,
                $this->auditPayload($payload),
                $this->auditContext($request),
            );

            return $payload;
        });
    }

    public function update(Request $request, Customer $customer, array $data): ?array
    {
        $businessId = $request->user()->business_id;
        $actorId = $request->user()->id;

        return DB::transaction(function () use ($request, $customer, $data, $businessId, $actorId): ?array {
            $fresh = Customer::query()
                ->where('business_id', $businessId)
                ->where('id', $customer->id)
                ->lockForUpdate()
                ->first();

            if (! $fresh) {
                return null;
            }

            $before = $this->payload($fresh->id, $businessId);
            $updates = [];
            foreach (['name', 'phone', 'group'] as $field) {
                if (array_key_exists($field, $data)) {
                    $updates[$field] = $field === 'group' ? ($data[$field] ?: 'Tanpa Grup') : $data[$field];
                }
            }

            if ($updates !== []) {
                $updates['updated_at'] = now();
                $fresh->update($updates);
            }

            $after = $this->payload($fresh->id, $businessId);
            $this->auditLog->record(
                $businessId,
                $actorId,
                'customer.updated',
                'customer',
                $fresh->id,
                $this->auditPayload($before),
                $this->auditPayload($after),
                $this->auditContext($request),
            );

            return $after;
        });
    }

    public function archive(Request $request, Customer $customer): ?array
    {
        $businessId = $request->user()->business_id;
        $actorId = $request->user()->id;

        return DB::transaction(function () use ($request, $customer, $businessId, $actorId): ?array {
            $fresh = Customer::query()
                ->where('business_id', $businessId)
                ->where('id', $customer->id)
                ->lockForUpdate()
                ->first();

            if (! $fresh) {
                return null;
            }

            $before = $this->payload($fresh->id, $businessId);
            $fresh->delete();
            $after = $this->auditPayload($before) + ['archived' => true];

            $this->auditLog->record(
                $businessId,
                $actorId,
                'customer.archived',
                'customer',
                $fresh->id,
                $this->auditPayload($before),
                $after,
                $this->auditContext($request),
            );

            return $before + ['archived' => true];
        });
    }

    private function payload(string $id, string $businessId): array
    {
        $customer = Customer::query()
            ->where('business_id', $businessId)
            ->where('id', $id)
            ->firstOrFail(['id', 'business_id', 'name', 'phone', 'group']);

        return [
            'id' => $customer->id,
            'business_id' => $customer->business_id,
            'name' => $customer->name,
            'phone' => $customer->phone,
            'group' => $customer->group,
        ];
    }

    private function auditPayload(array $payload): array
    {
        return [
            'id' => $payload['id'],
            'business_id' => $payload['business_id'],
            'group' => $payload['group'],
            'has_phone' => filled($payload['phone'] ?? null),
        ];
    }

    private function auditContext(Request $request): array
    {
        return [
            'request_id' => $request->header('X-Request-Id'),
            'idempotency_key' => $request->attributes->get('idempotency_key') ?: $request->header('Idempotency-Key'),
        ];
    }
}
