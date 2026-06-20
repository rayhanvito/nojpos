<?php

namespace App\Support;

use App\Services\AuditLogService;

class Nojpos
{
    public static function audit(?string $businessId, ?string $actorId, string $action, ?string $entityType = null, ?string $entityId = null, ?array $before = null, ?array $after = null): void
    {
        $request = request();

        app(AuditLogService::class)->record(
            $businessId,
            $actorId,
            $action,
            $entityType,
            $entityId,
            $before,
            $after,
            [
                'approver_id' => $request?->input('approver_id'),
                'outlet_id' => $request?->input('outlet_id'),
                'device_id' => $request?->input('device_id'),
                'request_id' => $request?->header('X-Request-Id') ?: $request?->header('X-Correlation-Id'),
                'idempotency_key' => $request?->attributes->get('idempotency_key') ?: $request?->header('Idempotency-Key'),
                'event_version' => 1,
            ],
        );
    }
}
