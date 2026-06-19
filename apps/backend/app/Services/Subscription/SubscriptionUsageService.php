<?php

namespace App\Services\Subscription;

use Illuminate\Support\Facades\DB;

class SubscriptionUsageService
{
    public function usage(?string $businessId): array
    {
        if (! $businessId) {
            return [
                'outlets' => 0,
                'devices' => 0,
                'users' => 0,
                'products' => 0,
            ];
        }

        return [
            'outlets' => $this->count('outlets', $businessId),
            'devices' => $this->count('devices', $businessId),
            'users' => $this->count('users', $businessId),
            'products' => $this->count('products', $businessId),
        ];
    }

    private function count(string $table, string $businessId): int
    {
        return DB::table($table)
            ->where('business_id', $businessId)
            ->whereNull('deleted_at')
            ->count();
    }
}
