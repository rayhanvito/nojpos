<?php

namespace App\Http\Controllers\Api\V1\Superadmin;

use App\Http\Controllers\Controller;
use App\Http\Requests\Superadmin\StorePlanRequest;
use App\Http\Requests\Superadmin\UpdatePlanRequest;
use App\Http\Requests\Superadmin\UpdatePlanStatusRequest;
use App\Models\Plan;
use App\Services\Superadmin\PlanService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;

class PlanController extends Controller
{
    public function __construct(private readonly PlanService $plans) {}

    public function index(): JsonResponse
    {
        return ApiResponse::success([
            'plans' => Plan::query()
                ->orderBy('name')
                ->get()
                ->map(fn (Plan $plan): array => $this->plans->payload($plan))
                ->values()
                ->all(),
        ]);
    }

    public function store(StorePlanRequest $request): JsonResponse
    {
        $plan = $this->plans->create(
            $request->user()->id,
            $request->validated(),
            $request->attributes->get('idempotency_key'),
        );

        return ApiResponse::success($this->plans->payload($plan), [], 201);
    }

    public function update(UpdatePlanRequest $request, Plan $plan): JsonResponse
    {
        $plan = $this->plans->update(
            $request->user()->id,
            $plan,
            $request->validated(),
            $request->attributes->get('idempotency_key'),
        );

        return ApiResponse::success($this->plans->payload($plan));
    }

    public function status(UpdatePlanStatusRequest $request, Plan $plan): JsonResponse
    {
        $plan = $this->plans->updateStatus(
            $request->user()->id,
            $plan,
            (bool) $request->validated('is_active'),
            $request->attributes->get('idempotency_key'),
        );

        return ApiResponse::success($this->plans->payload($plan));
    }
}
