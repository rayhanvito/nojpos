<?php

namespace App\Http\Controllers\Api\V1\Superadmin;

use App\Http\Controllers\Controller;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Support\Facades\DB;

class SummaryController extends Controller
{
    public function __invoke(): JsonResponse
    {
        return ApiResponse::success([
            'businesses_total' => DB::table('businesses')->whereNull('deleted_at')->count(),
            'businesses_active' => DB::table('businesses')->whereNull('deleted_at')->where('status', 'active')->count(),
            'plans_active' => DB::table('plans')->whereNull('deleted_at')->where('is_active', true)->count(),
            'subscriptions_by_status' => DB::table('subscriptions')
                ->whereNull('deleted_at')
                ->select('status', DB::raw('count(*) as total'))
                ->groupBy('status')
                ->pluck('total', 'status')
                ->all(),
        ]);
    }
}
