<?php

namespace App\Support;

use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class Nojpos
{
    public static function audit(?string $businessId, ?string $actorId, string $action, ?string $entityType = null, ?string $entityId = null, ?array $before = null, ?array $after = null): void
    {
        DB::table('audit_logs')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'actor_id' => $actorId,
            'action' => $action,
            'entity_type' => $entityType,
            'entity_id' => $entityId,
            'before' => $before ? json_encode($before) : null,
            'after' => $after ? json_encode($after) : null,
            'created_at' => now(),
        ]);
    }
}
