<?php

namespace App\Http\Controllers\Api\V1\Superadmin;

use App\Http\Controllers\Controller;
use App\Models\Plan;
use App\Support\ApiResponse;
use App\Support\Nojpos;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class PlanController extends Controller
{
    public function index(): JsonResponse
    {
        return ApiResponse::success([
            'plans' => Plan::query()
                ->orderBy('name')
                ->get()
                ->map(fn (Plan $plan): array => $this->payload($plan))
                ->values()
                ->all(),
        ]);
    }

    public function store(Request $request): JsonResponse
    {
        $data = $request->validate($this->rules());
        $plan = Plan::query()->create($data + ['is_active' => $data['is_active'] ?? true]);

        Nojpos::audit(null, $request->user()->id, 'superadmin.plan.created', 'plan', $plan->id, null, $this->payload($plan));

        return ApiResponse::success($this->payload($plan), [], 201);
    }

    public function update(Request $request, Plan $plan): JsonResponse
    {
        $data = $request->validate($this->rules(false));
        $before = $this->payload($plan);
        $plan->update($data);

        Nojpos::audit(null, $request->user()->id, 'superadmin.plan.updated', 'plan', $plan->id, $before, $this->payload($plan));

        return ApiResponse::success($this->payload($plan));
    }

    public function status(Request $request, Plan $plan): JsonResponse
    {
        $data = $request->validate(['is_active' => ['required', 'boolean']]);
        $before = $this->payload($plan);
        $plan->update(['is_active' => $data['is_active']]);

        Nojpos::audit(null, $request->user()->id, 'superadmin.plan.status_updated', 'plan', $plan->id, $before, $this->payload($plan));

        return ApiResponse::success($this->payload($plan));
    }

    private function rules(bool $creating = true): array
    {
        $required = $creating ? 'required' : 'sometimes';

        return [
            'name' => [$required, 'string', 'max:255'],
            'code' => [$required, 'string', 'max:100'],
            'price' => [$required, 'integer', 'min:0'],
            'billing_period' => ['nullable', 'string', 'max:50'],
            'max_outlets' => ['nullable', 'integer', 'min:0'],
            'max_devices' => ['nullable', 'integer', 'min:0'],
            'max_users' => ['nullable', 'integer', 'min:0'],
            'max_products' => ['nullable', 'integer', 'min:0'],
            'is_active' => ['sometimes', 'boolean'],
        ];
    }

    private function payload(Plan $plan): array
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
}
