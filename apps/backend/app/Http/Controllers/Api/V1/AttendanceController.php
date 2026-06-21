<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\AttendanceIndexRequest;
use App\Http\Requests\StoreAttendanceRequest;
use App\Services\AttendanceOperationException;
use App\Services\AttendanceService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;

class AttendanceController extends Controller
{
    public function __construct(private readonly AttendanceService $attendance) {}

    public function index(AttendanceIndexRequest $request): JsonResponse
    {
        try {
            return ApiResponse::success([
                'attendance' => $this->attendance->list($request->user(), $request->validated()),
            ]);
        } catch (AttendanceOperationException $error) {
            return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
        }
    }

    public function store(StoreAttendanceRequest $request): JsonResponse
    {
        try {
            $result = $this->attendance->record($request->user(), $request, $request->validated());

            return ApiResponse::success($result['payload'], [], $result['status']);
        } catch (AttendanceOperationException $error) {
            return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
        }
    }
}
