<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\ReportRequest;
use App\Services\ReportAccessException;
use App\Services\ReportService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;

class ReportController extends Controller
{
    public function __construct(private readonly ReportService $reports) {}

    public function salesSummary(ReportRequest $request): JsonResponse
    {
        return $this->reportResponse($request, fn (array $filters): array => $this->reports->salesSummary($request->user(), $filters));
    }

    public function soldProducts(ReportRequest $request): JsonResponse
    {
        return $this->reportResponse($request, fn (array $filters): array => $this->reports->soldProducts($request->user(), $filters));
    }

    public function paymentMethods(ReportRequest $request): JsonResponse
    {
        return $this->reportResponse($request, fn (array $filters): array => $this->reports->paymentMethods($request->user(), $filters));
    }

    public function cashierShifts(ReportRequest $request): JsonResponse
    {
        return $this->reportResponse($request, fn (array $filters): array => $this->reports->cashierShifts($request->user(), $filters));
    }

    public function voidRefundAudit(ReportRequest $request): JsonResponse
    {
        return $this->reportResponse($request, fn (array $filters): array => $this->reports->voidRefundAudit($request->user(), $filters));
    }

    public function topTen(ReportRequest $request): JsonResponse
    {
        return $this->reportResponse($request, fn (array $filters): array => $this->reports->topTen($request->user(), $filters));
    }

    /**
     * @param  callable(array<string, mixed>): array{data:array<string, mixed>, meta:array<string, mixed>}  $callback
     */
    private function reportResponse(ReportRequest $request, callable $callback): JsonResponse
    {
        $filters = $request->validated();

        if (isset($filters['export'])) {
            return ApiResponse::error(
                'REPORT_EXPORT_NOT_IMPLEMENTED',
                'Report export is not available yet. Use the paginated JSON report endpoint for now.',
                ['export' => $filters['export']],
                501,
            );
        }

        try {
            $result = $callback($filters);
        } catch (ReportAccessException $error) {
            return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
        }

        return ApiResponse::success($result['data'], $result['meta']);
    }
}
