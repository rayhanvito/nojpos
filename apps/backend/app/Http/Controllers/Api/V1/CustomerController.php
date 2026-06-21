<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\StoreCustomerRequest;
use App\Http\Requests\UpdateCustomerRequest;
use App\Models\Customer;
use App\Services\CustomerService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class CustomerController extends Controller
{
    public function __construct(private readonly CustomerService $customers) {}

    public function index(Request $request): JsonResponse
    {
        $query = Customer::query()->where('business_id', $request->user()->business_id);

        if ($request->filled('search')) {
            $search = $request->query('search');
            $query->where(function ($query) use ($search): void {
                $query->where('name', 'like', '%'.$search.'%')
                    ->orWhere('phone', 'like', '%'.$search.'%');
            });
        }

        if ($request->filled('group')) {
            $query->where('group', $request->query('group'));
        }

        return ApiResponse::success([
            'customers' => $query
                ->orderBy('name')
                ->get(['id', 'business_id', 'name', 'phone', 'group'])
                ->values()
                ->all(),
        ]);
    }

    public function store(StoreCustomerRequest $request): JsonResponse
    {
        return ApiResponse::success(
            $this->customers->create($request, $request->validated()),
            [],
            201,
        );
    }

    public function update(UpdateCustomerRequest $request, Customer $customer): JsonResponse
    {
        $payload = $this->customers->update($request, $customer, $request->validated());

        if (! $payload) {
            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        return ApiResponse::success($payload);
    }

    public function destroy(Request $request, Customer $customer): JsonResponse
    {
        $payload = $this->customers->archive($request, $customer);

        if (! $payload) {
            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        return ApiResponse::success(['archived' => true]);
    }
}
