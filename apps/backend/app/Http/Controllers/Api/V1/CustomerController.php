<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\Customer;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class CustomerController extends Controller
{
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

    public function store(Request $request): JsonResponse
    {
        $data = $request->validate([
            'name' => ['required', 'string', 'max:255'],
            'phone' => ['nullable', 'string', 'max:50'],
            'group' => ['nullable', 'string', 'max:100'],
        ]);

        $customer = Customer::query()->create([
            'business_id' => $request->user()->business_id,
            'name' => $data['name'],
            'phone' => $data['phone'] ?? null,
            'group' => $data['group'] ?? 'Tanpa Grup',
        ]);

        return ApiResponse::success($customer->only([
            'id',
            'business_id',
            'name',
            'phone',
            'group',
        ]), [], 201);
    }
}
