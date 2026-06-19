<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\Plan;
use App\Models\Scopes\BusinessScope;
use App\Models\Subscription;
use App\Services\Subscription\SubscriptionLimitService;
use App\Services\Subscription\SubscriptionUsageService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class SubscriptionController extends Controller
{
    public function show(
        Request $request,
        SubscriptionUsageService $usageService,
        SubscriptionLimitService $limitService,
    ): JsonResponse {
        $businessId = $request->user()?->business_id;
        $subscription = $businessId
            ? Subscription::withoutGlobalScope(BusinessScope::class)
                ->where('business_id', $businessId)
                ->first()
            : null;
        $plan = $subscription?->plan_id
            ? Plan::query()->where('id', $subscription->plan_id)->first()
            : null;

        return ApiResponse::success([
            'status' => $subscription?->status,
            'plan' => $plan ? [
                'name' => $plan->name,
                'code' => $plan->code,
            ] : null,
            'limits' => $limitService->limits($subscription, $plan),
            'usage' => $usageService->usage($businessId),
            'enforcement' => $limitService->enforcement(),
            'trial_ends_at' => $subscription?->trial_ends_at?->toDateTimeString(),
            'current_period_ends_at' => $subscription?->current_period_ends_at?->toDateTimeString(),
        ]);
    }
}
