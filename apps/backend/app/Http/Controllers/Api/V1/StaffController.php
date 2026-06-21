<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\StoreStaffRequest;
use App\Http\Requests\UpdateStaffRequest;
use App\Models\User;
use App\Services\StaffService;
use App\Services\StaffServiceException;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class StaffController extends Controller
{
    public function __construct(private readonly StaffService $staffService) {}

    public function index(Request $request): JsonResponse
    {
        $data = $request->validate([
            'outlet_id' => ['nullable', 'uuid'],
            'search' => ['nullable', 'string'],
        ]);

        $businessId = $request->user()->business_id;
        if (($data['outlet_id'] ?? null) && ! DB::table('outlets')->where('business_id', $businessId)->where('id', $data['outlet_id'])->exists()) {
            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        $staff = User::query()
            ->where('business_id', $businessId)
            ->whereIn('role', ['owner', 'admin', 'cashier'])
            ->when($data['search'] ?? null, function ($query, string $search): void {
                $query->where(function ($query) use ($search): void {
                    $query->where('name', 'like', '%'.$search.'%')
                        ->orWhere('email', 'like', '%'.$search.'%');
                });
            })
            ->orderBy('name')
            ->get()
            ->map(fn (User $user): array => $this->staffService->payload($user))
            ->values()
            ->all();

        return ApiResponse::success(['staff' => $staff]);
    }

    public function store(StoreStaffRequest $request): JsonResponse
    {
        try {
            return ApiResponse::success([
                'staff' => $this->staffService->create($request, $request->validated()),
            ], [], 201);
        } catch (StaffServiceException $exception) {
            return ApiResponse::error($exception->errorCode, $exception->getMessage(), $exception->details, $exception->status);
        }
    }

    public function update(UpdateStaffRequest $request, User $staff): JsonResponse
    {
        try {
            return ApiResponse::success([
                'staff' => $this->staffService->update($request, $staff, $request->validated()),
            ]);
        } catch (StaffServiceException $exception) {
            return ApiResponse::error($exception->errorCode, $exception->getMessage(), $exception->details, $exception->status);
        }
    }

    public function destroy(Request $request, User $staff): JsonResponse
    {
        try {
            $this->staffService->deactivate($request, $staff);
        } catch (StaffServiceException $exception) {
            return ApiResponse::error($exception->errorCode, $exception->getMessage(), $exception->details, $exception->status);
        }

        return ApiResponse::success(['deleted' => true]);
    }
}
