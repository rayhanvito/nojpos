<?php

namespace App\Services;

use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class AuditLogService
{
    public function record(
        ?string $businessId,
        ?string $actorId,
        string $action,
        ?string $entityType = null,
        ?string $entityId = null,
        ?array $before = null,
        ?array $after = null,
        array $context = [],
    ): string {
        $id = (string) Str::uuid();

        DB::table('audit_logs')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'actor_id' => $actorId,
            'approver_id' => $context['approver_id'] ?? null,
            'outlet_id' => $context['outlet_id'] ?? null,
            'device_id' => $context['device_id'] ?? null,
            'request_id' => $context['request_id'] ?? null,
            'idempotency_key' => $context['idempotency_key'] ?? null,
            'event_version' => $context['event_version'] ?? 1,
            'action' => $action,
            'entity_type' => $entityType,
            'entity_id' => $entityId,
            'before' => $before ? json_encode($before, JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE) : null,
            'after' => $after ? json_encode($after, JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE) : null,
            'created_at' => now(),
        ]);

        return $id;
    }
}
