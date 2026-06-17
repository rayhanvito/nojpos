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
            'products' => $query
                ->orderBy('name')
                ->get(['id', 'business_id', 'outlet_id', 'product_category_id', 'name', 'barcode', 'price', 'track_stock'])
                ->values()
                ->all(),
        ]);
    }
}
