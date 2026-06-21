<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\StoreProductRequest;
use App\Http\Requests\UpdateProductRequest;
use App\Models\Product;
use App\Services\ProductService;
use App\Services\ProductServiceException;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class ProductController extends Controller
{
    public function __construct(private readonly ProductService $products) {}

    public function index(Request $request): JsonResponse
    {
        $query = Product::query();

        if ($request->filled('outlet_id')) {
            if (! $this->belongsToBusiness('outlets', $request->user()->business_id, $request->query('outlet_id'))) {
                return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
            }

            $query->where(function ($query) use ($request): void {
                $query->where('outlet_id', $request->query('outlet_id'))
                    ->orWhereNull('outlet_id');
            });
        }

        if ($request->filled('search')) {
            $search = $request->query('search');
            $query->where(function ($query) use ($search): void {
                $query->where('name', 'like', '%'.$search.'%')
                    ->orWhere('barcode', 'like', '%'.$search.'%');
            });
        }

        if ($request->filled('barcode')) {
            $query->where('barcode', $request->query('barcode'));
        }

        if ($request->filled('category')) {
            $categoryIds = DB::table('product_categories')
                ->where('business_id', $request->user()->business_id)
                ->whereNull('deleted_at')
                ->where('name', $request->query('category'))
                ->pluck('id');

            $query->whereIn('product_category_id', $categoryIds);
        }

        return ApiResponse::success([
            'categories' => DB::table('product_categories')
                ->where('business_id', $request->user()->business_id)
                ->whereNull('deleted_at')
                ->orderBy('name')
                ->get(['id', 'business_id', 'name'])
                ->values()
                ->all(),
            'products' => $query
                ->orderBy('name')
                ->get(['id', 'business_id', 'outlet_id', 'product_category_id', 'name', 'barcode', 'price', 'track_stock'])
                ->values()
                ->all(),
        ]);
    }

    public function store(StoreProductRequest $request): JsonResponse
    {
        if (! $this->canManage($request)) {
            return ApiResponse::error('FORBIDDEN', 'Only owner or admin can manage products.', [], 403);
        }

        try {
            return ApiResponse::success($this->products->create($request, $request->validated()), [], 201);
        } catch (ProductServiceException $exception) {
            return ApiResponse::error($exception->errorCode, $exception->getMessage(), $exception->details, $exception->status);
        }
    }

    public function update(UpdateProductRequest $request, string $product): JsonResponse
    {
        if (! $this->canManage($request)) {
            return ApiResponse::error('FORBIDDEN', 'Only owner or admin can manage products.', [], 403);
        }

        try {
            $payload = $this->products->update($request, $product, $request->validated());
        } catch (ProductServiceException $exception) {
            return ApiResponse::error($exception->errorCode, $exception->getMessage(), $exception->details, $exception->status);
        }

        if (! $payload) {
            return ApiResponse::error('NOT_FOUND', 'Product not found.', [], 404);
        }

        return ApiResponse::success($payload);
    }

    public function destroy(Request $request, string $product): JsonResponse
    {
        if (! $this->canManage($request)) {
            return ApiResponse::error('FORBIDDEN', 'Only owner or admin can manage products.', [], 403);
        }

        $payload = $this->products->archive($request, $product);

        if (! $payload) {
            return ApiResponse::error('NOT_FOUND', 'Product not found.', [], 404);
        }

        return ApiResponse::success($payload);
    }

    private function canManage(Request $request): bool
    {
        return in_array($request->user()->role, ['owner', 'admin'], true);
    }

    private function belongsToBusiness(string $table, string $businessId, string $id): bool
    {
        return DB::table($table)
            ->where('business_id', $businessId)
            ->where('id', $id)
            ->whereNull('deleted_at')
            ->exists();
    }
}
