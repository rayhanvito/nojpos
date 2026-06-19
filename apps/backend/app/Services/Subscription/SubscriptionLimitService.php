<?php

namespace App\Services\Subscription;

use App\Models\Plan;
use App\Models\Subscription;

class SubscriptionLimitService
{
    public function limits(?Subscription $subscription, ?Plan $plan): array
    {
        return [
            'max_outlets' => $this->effective($subscription?->max_outlets_override, $plan?->max_outlets),
            'max_devices' => $this->effective($subscription?->max_devices_override, $plan?->max_devices),
            'max_users' => $this->effective($subscription?->max_users_override, $plan?->max_users),
            'max_products' => $this->effective($subscription?->max_products_override, $plan?->max_products),
        ];
    }

    public function enforcement(): array
    {
        return [
            'max_devices' => 'informational',
        ];
    }

    private function effective(?int $override, ?int $default): ?int
    {
        return $override ?? $default;
    }
}
