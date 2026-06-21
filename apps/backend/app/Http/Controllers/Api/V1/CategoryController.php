<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\StoreCategoryRequest;
use App\Http\Requests\UpdateCategoryRequest;
use App\Services\CategoryService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class CategoryController extends Controller
{
    public function __construct(private readonly CategoryService $categories) {}

    public function index(Request $request): JsonResponse
    {
        $categories = DB::table('product_categories')
            ->where('business_id', $request->user()->business_id)
            ->whereNull('deleted_at')
            ->when($request->filled('search'), fn ($query) => $query->where('name', 'like', '%'.$request->query('search').'%'))
            ->orderBy('name')
            ->get(['id', 'business_id', 'name'])
            ->values()
            ->all();

        return ApiResponse::success(['categories' => $categories]);
    }

    public function store(StoreCategoryRequest $request): JsonResponse
    {
        if (! $this->canManage($request)) {
            return ApiResponse::error('FORBIDDEN', 'Only owner or admin can manage categories.', [], 403);
        }

        return ApiResponse::success($this->categories->create($request, $request->validated()), [], 201);
    }

    public function update(UpdateCategoryRequest $request, string $category): JsonResponse
    {
        if (! $this->canManage($request)) {
            return ApiResponse::error('FORBIDDEN', 'Only owner or admin can manage categories.', [], 403);
        }

        $payload = $this->categories->update($request, $category, $request->validated());

        if (! $payload) {
            return ApiResponse::error('NOT_FOUND', 'Category not found.', [], 404);
        }

        return ApiResponse::success($payload);
    }

    public function destroy(Request $request, string $category): JsonResponse
    {
        if (! $this->canManage($request)) {
            return ApiResponse::error('FORBIDDEN', 'Only owner or admin can manage categories.', [], 403);
        }

        $payload = $this->categories->archive($request, $category);

        if (! $payload) {
            return ApiResponse::error('NOT_FOUND', 'Category not found.', [], 404);
        }

        return ApiResponse::success($payload);
    }

    private function canManage(Request $request): bool
    {
        return in_array($request->user()->role, ['owner', 'admin'], true);
    }
}
