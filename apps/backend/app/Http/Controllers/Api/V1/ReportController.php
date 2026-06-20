<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Services\ReportAccessException;
use App\Services\ReportService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class ReportController extends Controller
{
    public function __construct(private readonly ReportService $reports) {}

    public function salesSummary(Request $request): JsonResponse
    {
        return $this->reportResponse($request, fn (array $filters): array => $this->reports->salesSummary($request->user(), $filters));
    }

    public function soldProducts(Request $request): JsonResponse
    {
        return $this->reportResponse($request, fn (array $filters): array => $this->reports->soldProducts($request->user(), $filters));
    }

    public function paymentMethods(Request $request): JsonResponse
    {
        return $this->reportResponse($request, fn (array $filters): array => $this->reports->paymentMethods($request->user(), $filters));
    }

    public function cashierShifts(Request $request): JsonResponse
    {
        return $this->reportResponse($request, fn (array $filters): array => $this->reports->cashierShifts($request->user(), $filters));
    }

    public function voidRefundAudit(Request $request): JsonResponse
    {
        return $this->reportResponse($request, fn (array $filters): array => $this->reports->voidRefundAudit($request->user(), $filters));
    }

    public function topTen(Request $request): JsonResponse
    {
        return $this->reportResponse($request, fn (array $filters): array => $this->reports->topTen($request->user(), $filters));
    }

    /**
     * @param  callable(array<string, mixed>): array{data:array<string, mixed>, meta:array<string, mixed>}  $callback
     */
    private function reportResponse(Request $request, callable $callback): JsonResponse
    {
        $filters = $request->validate([
            'date' => ['nullable', 'date'],
            'range' => ['nullable', 'in:day,week,month'],
            'shift_id' => ['nullable', 'uuid'],
            'outlet_id' => ['nullable', 'uuid'],
        ]);

        try {
            $result = $callback($filters);
        } catch (ReportAccessException $error) {
            return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
        }

        return ApiResponse::success($result['data'], $result['meta']);
    }
}
