<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class CategoryController extends Controller
{
    public function index(Request $request): JsonResponse
    {
        $categories = DB::table('product_categories')
            ->where('business_id', $request->user()->business_id)
            ->when($request->filled('search'), fn ($query) => $query->where('name', 'like', '%'.$request->query('search').'%'))
            ->orderBy('name')
            ->get(['id', 'business_id', 'name'])
            ->values()
            ->all();

        return ApiResponse::success(['categories' => $categories]);
    }

    public function store(Request $request): JsonResponse
    {
        if (! $this->canManage($request)) {
            return ApiResponse::error('FORBIDDEN', 'Only owner or admin can manage categories.', [], 403);
        }

        $data = $request->validate([
            'name' => ['required', 'string', 'max:255'],
        ]);

        $id = (string) Str::uuid();
        DB::table('product_categories')->insert([
            'id' => $id,
            'business_id' => $request->user()->business_id,
            'name' => $data['name'],
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        return ApiResponse::success($this->payload($id), [], 201);
    }

    public function update(Request $request, string $category): JsonResponse
    {
        if (! $this->canManage($request)) {
            return ApiResponse::error('FORBIDDEN', 'Only owner or admin can manage categories.', [], 403);
        }

        $data = $request->validate([
            'name' => ['required', 'string', 'max:255'],
        ]);

        $businessId = $request->user()->business_id;
        $updated = DB::table('product_categories')
            ->where('business_id', $businessId)
            ->where('id', $category)
            ->update([
                'name' => $data['name'],
                'updated_at' => now(),
            ]);

        if (! $updated) {
            return ApiResponse::error('NOT_FOUND', 'Category not found.', [], 404);
        }

        return ApiResponse::success($this->payload($category));
    }

    private function canManage(Request $request): bool
    {
        return in_array($request->user()->role, ['owner', 'admin'], true);
    }

    private function payload(string $id): array
    {
        return (array) DB::table('product_categories')
            ->where('id', $id)
            ->first(['id', 'business_id', 'name']);
    }
}
