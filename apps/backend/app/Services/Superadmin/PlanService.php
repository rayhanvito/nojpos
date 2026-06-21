<?php

namespace App\Services\Superadmin;

use App\Models\Plan;
use App\Services\AuditLogService;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class PlanService
{
    public function __construct(private readonly AuditLogService $audit) {}

    public function create(string $actorId, array $data, ?string $idempotencyKey = null): Plan
    {
        return DB::transaction(function () use ($actorId, $data, $idempotencyKey): Plan {
            $plan = Plan::query()->create($data + ['is_active' => $data['is_active'] ?? true]);

            $this->audit->record(
                null,
                $actorId,
                'superadmin.plan.created',
                'plan',
                $plan->id,
                null,
                $this->payload($plan),
                $this->context($idempotencyKey),
            );

            return $plan->refresh();
        });
    }

    public function update(string $actorId, Plan $plan, array $data, ?string $idempotencyKey = null): Plan
    {
        return DB::transaction(function () use ($actorId, $plan, $data, $idempotencyKey): Plan {
            /** @var Plan $locked */
            $locked = Plan::query()->whereKey($plan->id)->lockForUpdate()->firstOrFail();
            $before = $this->payload($locked);

            $locked->update($data);
            $locked->refresh();

            $this->audit->record(
                null,
                $actorId,
                'superadmin.plan.updated',
                'plan',
                $locked->id,
                $before,
                $this->payload($locked),
                $this->context($idempotencyKey),
            );

            return $locked;
        });
    }

    public function updateStatus(string $actorId, Plan $plan, bool $isActive, ?string $idempotencyKey = null): Plan
    {
        return DB::transaction(function () use ($actorId, $plan, $isActive, $idempotencyKey): Plan {
            /** @var Plan $locked */
            $locked = Plan::query()->whereKey($plan->id)->lockForUpdate()->firstOrFail();
            $before = $this->payload($locked);

            $locked->update(['is_active' => $isActive]);
            $locked->refresh();

            $this->audit->record(
                null,
                $actorId,
                'superadmin.plan.status_updated',
                'plan',
                $locked->id,
                $before,
                $this->payload($locked),
                $this->context($idempotencyKey),
            );

            return $locked;
        });
    }

    public function payload(Plan $plan): array
    {
        return [
            'id' => $plan->id,
            'name' => $plan->name,
            'code' => $plan->code,
            'price' => (int) $plan->price,
            'billing_period' => $plan->billing_period,
            'max_outlets' => $plan->max_outlets,
            'max_devices' => $plan->max_devices,
            'max_users' => $plan->max_users,
            'max_products' => $plan->max_products,
            'is_active' => (bool) $plan->is_active,
        ];
    }

    private function context(?string $idempotencyKey): array
    {
        return [
            'request_id' => (string) Str::uuid(),
            'idempotency_key' => $idempotencyKey,
        ];
    }
}
