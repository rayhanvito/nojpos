<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\Product;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class ProductController extends Controller
{
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
            $query->where('name', 'like', '%'.$request->query('search').'%');
        }

        if ($request->filled('barcode')) {
            $query->where('barcode', $request->query('barcode'));
        }

        if ($request->filled('category')) {
            $categoryIds = DB::table('product_categories')
                ->where('business_id', $request->user()->business_id)
                ->where('name', $request->query('category'))
                ->pluck('id');

            $query->whereIn('product_category_id', $categoryIds);
        }

        return ApiResponse::success([
            'categories' => DB::table('product_categories')
                ->where('business_id', $request->user()->business_id)
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

    public function store(Request $request): JsonResponse
    {
        if (! $this->canManage($request)) {
            return ApiResponse::error('FORBIDDEN', 'Only owner or admin can manage products.', [], 403);
        }

        $data = $request->validate($this->rules());
        $businessId = $request->user()->business_id;

        if (! $this->relationsAreValid($businessId, $data)) {
            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        $product = Product::query()->create([
            'business_id' => $businessId,
            'outlet_id' => $data['outlet_id'] ?? null,
            'product_category_id' => $data['product_category_id'] ?? null,
            'name' => $data['name'],
            'barcode' => $data['barcode'] ?? null,
            'price' => $data['price'],
            'track_stock' => $data['track_stock'] ?? false,
        ]);

        return ApiResponse::success($this->productPayload($product->id), [], 201);
    }

    public function update(Request $request, string $product): JsonResponse
    {
        if (! $this->canManage($request)) {
            return ApiResponse::error('FORBIDDEN', 'Only owner or admin can manage products.', [], 403);
        }

        $data = $request->validate($this->rules(false));
        $businessId = $request->user()->business_id;

        $existing = Product::query()
            ->where('business_id', $businessId)
            ->where('id', $product)
            ->first();

        if (! $existing) {
            return ApiResponse::error('NOT_FOUND', 'Product not found.', [], 404);
        }

        if (! $this->relationsAreValid($businessId, $data)) {
            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        $existing->update(array_filter([
            'outlet_id' => $data['outlet_id'] ?? null,
            'product_category_id' => $data['product_category_id'] ?? null,
            'name' => $data['name'] ?? null,
            'barcode' => array_key_exists('barcode', $data) ? $data['barcode'] : null,
            'price' => $data['price'] ?? null,
            'track_stock' => array_key_exists('track_stock', $data) ? $data['track_stock'] : null,
        ], fn ($value): bool => $value !== null));

        if (array_key_exists('barcode', $data) && $data['barcode'] === null) {
            $existing->update(['barcode' => null]);
        }

        return ApiResponse::success($this->productPayload($product));
    }

    private function rules(bool $creating = true): array
    {
        $required = $creating ? 'required' : 'sometimes';

        return [
            'outlet_id' => [$creating ? 'required' : 'sometimes', 'uuid'],
            'product_category_id' => ['nullable', 'uuid'],
            'name' => [$required, 'string', 'max:255'],
            'barcode' => ['nullable', 'string', 'max:100'],
            'price' => [$required, 'integer', 'min:0'],
            'track_stock' => ['sometimes', 'boolean'],
        ];
    }

    private function canManage(Request $request): bool
    {
        return in_array($request->user()->role, ['owner', 'admin'], true);
    }

    private function relationsAreValid(string $businessId, array $data): bool
    {
        if (($data['outlet_id'] ?? null) && ! $this->belongsToBusiness('outlets', $businessId, $data['outlet_id'])) {
            return false;
        }

        if (($data['product_category_id'] ?? null) && ! $this->belongsToBusiness('product_categories', $businessId, $data['product_category_id'])) {
            return false;
        }

        return true;
    }

    private function belongsToBusiness(string $table, string $businessId, string $id): bool
    {
        return DB::table($table)
            ->where('business_id', $businessId)
            ->where('id', $id)
            ->exists();
    }

    private function productPayload(string $id): array
    {
        $product = Product::query()
            ->where('id', $id)
            ->firstOrFail(['id', 'business_id', 'outlet_id', 'product_category_id', 'name', 'barcode', 'price', 'track_stock']);

        return [
            'id' => $product->id,
            'business_id' => $product->business_id,
            'outlet_id' => $product->outlet_id,
            'product_category_id' => $product->product_category_id,
            'name' => $product->name,
            'barcode' => $product->barcode,
            'price' => (int) $product->price,
            'track_stock' => (bool) $product->track_stock,
        ];
    }
}
