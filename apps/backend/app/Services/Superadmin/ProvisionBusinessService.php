<?php

namespace App\Services\Superadmin;

use App\Models\Business;
use App\Models\Scopes\BusinessScope;
use App\Models\Subscription;
use App\Support\Nojpos;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;

class ProvisionBusinessService
{
    public function provision(array $data, string $actorId): array
    {
        return DB::transaction(function () use ($data, $actorId): array {
            $now = now();
            $business = Business::query()->create([
                'name' => $data['business']['name'],
                'status' => $data['business']['status'] ?? 'active',
            ]);

            $ownerId = (string) Str::uuid();
            DB::table('users')->insert([
                'id' => $ownerId,
                'business_id' => $business->id,
                'name' => $data['owner']['name'],
                'email' => $data['owner']['email'],
                'password' => Hash::make($data['owner']['password']),
                'role' => 'owner',
                'pin_hash' => null,
                'created_at' => $now,
                'updated_at' => $now,
            ]);

            $outletId = (string) Str::uuid();
            DB::table('outlets')->insert([
                'id' => $outletId,
                'business_id' => $business->id,
                'name' => $data['outlet']['name'],
                'service_charge_rate' => 0,
                'tax_rate' => 0,
                'created_at' => $now,
                'updated_at' => $now,
            ]);

            $subscription = null;
            if (isset($data['subscription'])) {
                $subscription = Subscription::withoutGlobalScope(BusinessScope::class)->create($this->subscriptionAttributes(
                    $business->id,
                    $data['subscription'],
                    $actorId,
                ));
            }

            $payload = [
                'business' => [
                    'id' => $business->id,
                    'name' => $business->name,
                    'status' => $business->status,
                ],
                'owner' => [
                    'id' => $ownerId,
                    'business_id' => $business->id,
                    'name' => $data['owner']['name'],
                    'email' => $data['owner']['email'],
                    'role' => 'owner',
                ],
                'outlet' => [
                    'id' => $outletId,
                    'business_id' => $business->id,
                    'name' => $data['outlet']['name'],
                ],
                'subscription' => $subscription ? $this->subscriptionPayload($subscription) : null,
            ];

            Nojpos::audit($business->id, $actorId, 'superadmin.business.provisioned', 'business', $business->id, null, $payload);
            if ($subscription) {
                Nojpos::audit($business->id, $actorId, 'superadmin.subscription.assigned', 'subscription', $subscription->id, null, $this->subscriptionPayload($subscription));
            }

            return $payload;
        });
    }

    private function subscriptionAttributes(string $businessId, array $data, string $actorId): array
    {
        return [
            'business_id' => $businessId,
            'plan_id' => $data['plan_id'] ?? null,
            'status' => $data['status'] ?? 'trial',
            'trial_ends_at' => $data['trial_ends_at'] ?? null,
            'current_period_ends_at' => $data['current_period_ends_at'] ?? null,
            'max_outlets_override' => $data['max_outlets_override'] ?? null,
            'max_devices_override' => $data['max_devices_override'] ?? null,
            'max_users_override' => $data['max_users_override'] ?? null,
            'max_products_override' => $data['max_products_override'] ?? null,
            'notes' => $data['notes'] ?? null,
            'created_by' => $actorId,
            'updated_by' => $actorId,
        ];
    }

    private function subscriptionPayload(Subscription $subscription): array
    {
        return [
            'id' => $subscription->id,
            'business_id' => $subscription->business_id,
            'plan_id' => $subscription->plan_id,
            'status' => $subscription->status,
            'trial_ends_at' => $subscription->trial_ends_at?->toDateTimeString(),
            'current_period_ends_at' => $subscription->current_period_ends_at?->toDateTimeString(),
            'max_outlets_override' => $subscription->max_outlets_override,
            'max_devices_override' => $subscription->max_devices_override,
            'max_users_override' => $subscription->max_users_override,
            'max_products_override' => $subscription->max_products_override,
            'notes' => $subscription->notes,
            'created_by' => $subscription->created_by,
            'updated_by' => $subscription->updated_by,
        ];
    }
}
