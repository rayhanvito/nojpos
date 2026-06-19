<?php

namespace App\Http\Controllers\Api\V1\Superadmin;

use App\Http\Controllers\Controller;
use App\Models\Business;
use App\Models\Scopes\BusinessScope;
use App\Models\Subscription;
use App\Support\ApiResponse;
use App\Support\Nojpos;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class SubscriptionController extends Controller
{
    public function show(Business $business): JsonResponse
    {
        $subscription = Subscription::withoutGlobalScope(BusinessScope::class)
            ->where('business_id', $business->id)
            ->first();

        return ApiResponse::success([
            'subscription' => $subscription ? $this->payload($subscription) : null,
        ]);
    }

    public function store(Request $request, Business $business): JsonResponse
    {
        $data = $request->validate([
            'plan_id' => ['nullable', 'uuid', 'exists:plans,id'],
            'status' => ['required', 'string', 'in:trial,active,suspended,expired'],
            'trial_ends_at' => ['nullable', 'date'],
            'current_period_ends_at' => ['nullable', 'date'],
            'max_outlets_override' => ['nullable', 'integer', 'min:0'],
            'max_devices_override' => ['nullable', 'integer', 'min:0'],
            'max_users_override' => ['nullable', 'integer', 'min:0'],
            'max_products_override' => ['nullable', 'integer', 'min:0'],
            'notes' => ['nullable', 'string'],
        ]);

        [$subscription, $before, $created] = DB::transaction(function () use ($business, $data, $request): array {
            $subscription = Subscription::withoutGlobalScope(BusinessScope::class)
                ->where('business_id', $business->id)
                ->first();
            $before = $subscription ? $this->payload($subscription) : null;

            $attributes = $data + [
                'business_id' => $business->id,
                'created_by' => $request->user()->id,
                'updated_by' => $request->user()->id,
            ];
            $attributes['updated_by'] = $request->user()->id;

            if ($subscription) {
                unset($attributes['business_id'], $attributes['created_by']);
                $subscription->update($attributes);

                return [$subscription->refresh(), $before, false];
            }

            return [
                Subscription::withoutGlobalScope(BusinessScope::class)->create($attributes),
                null,
                true,
            ];
        });

        Nojpos::audit(
            $business->id,
            $request->user()->id,
            $created ? 'superadmin.subscription.assigned' : 'superadmin.subscription.updated',
            'subscription',
            $subscription->id,
            $before,
            $this->payload($subscription),
        );

        return ApiResponse::success($this->payload($subscription), [], $created ? 201 : 200);
    }

    private function payload(Subscription $subscription): array
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
