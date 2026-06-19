<?php

namespace App\Http\Controllers\Api\V1\Superadmin;

use App\Http\Controllers\Controller;
use App\Models\Business;
use App\Services\Superadmin\ProvisionBusinessService;
use App\Support\ApiResponse;
use App\Support\Nojpos;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class BusinessController extends Controller
{
    public function index(): JsonResponse
    {
        return ApiResponse::success([
            'businesses' => Business::query()
                ->orderBy('name')
                ->get()
                ->map(fn (Business $business): array => $this->payload($business))
                ->values()
                ->all(),
        ]);
    }

    public function store(Request $request, ProvisionBusinessService $service): JsonResponse
    {
        $data = $request->validate([
            'business.name' => ['required', 'string', 'max:255'],
            'business.status' => ['sometimes', 'string', 'in:active,suspended'],
            'owner.name' => ['required', 'string', 'max:255'],
            'owner.email' => ['required', 'email', 'max:255', 'unique:users,email'],
            'owner.password' => ['required', 'string', 'min:8'],
            'outlet.name' => ['required', 'string', 'max:255'],
            'subscription' => ['sometimes', 'array'],
            'subscription.plan_id' => ['nullable', 'uuid', 'exists:plans,id'],
            'subscription.status' => ['sometimes', 'string', 'in:trial,active,suspended,expired'],
            'subscription.trial_ends_at' => ['nullable', 'date'],
            'subscription.current_period_ends_at' => ['nullable', 'date'],
            'subscription.max_outlets_override' => ['nullable', 'integer', 'min:0'],
            'subscription.max_devices_override' => ['nullable', 'integer', 'min:0'],
            'subscription.max_users_override' => ['nullable', 'integer', 'min:0'],
            'subscription.max_products_override' => ['nullable', 'integer', 'min:0'],
            'subscription.notes' => ['nullable', 'string'],
        ]);

        return ApiResponse::success($service->provision($data, $request->user()->id), [], 201);
    }

    public function show(Business $business): JsonResponse
    {
        return ApiResponse::success($this->payload($business) + [
            'owner' => DB::table('users')
                ->where('business_id', $business->id)
                ->where('role', 'owner')
                ->whereNull('deleted_at')
                ->first(['id', 'business_id', 'name', 'email', 'role']),
            'outlets' => DB::table('outlets')
                ->where('business_id', $business->id)
                ->whereNull('deleted_at')
                ->orderBy('name')
                ->get(['id', 'business_id', 'name'])
                ->values()
                ->all(),
        ]);
    }

    public function update(Request $request, Business $business): JsonResponse
    {
        $data = $request->validate([
            'name' => ['sometimes', 'string', 'max:255'],
            'status' => ['sometimes', 'string', 'in:active,suspended'],
        ]);

        $before = $this->payload($business);
        $business->update($data);

        Nojpos::audit($business->id, $request->user()->id, 'superadmin.business.updated', 'business', $business->id, $before, $this->payload($business));

        return ApiResponse::success($this->payload($business));
    }

    private function payload(Business $business): array
    {
        return [
            'id' => $business->id,
            'name' => $business->name,
            'status' => $business->status ?? 'active',
        ];
    }
}
