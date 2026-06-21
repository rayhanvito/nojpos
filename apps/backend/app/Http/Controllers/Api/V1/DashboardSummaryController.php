<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\DashboardSummaryRequest;
use App\Services\DashboardAccessException;
use App\Services\DashboardSummaryService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;

class DashboardSummaryController extends Controller
{
    public function __construct(private readonly DashboardSummaryService $dashboard) {}

    public function __invoke(DashboardSummaryRequest $request): JsonResponse
    {
        try {
            $result = $this->dashboard->summary($request->user(), $request->validated());
        } catch (DashboardAccessException $error) {
            return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
        }

        return ApiResponse::success($result['data'], $result['meta']);
    }
}
