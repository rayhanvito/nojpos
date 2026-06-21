<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\CloseStoreRequest;
use App\Http\Requests\OpenStoreRequest;
use App\Services\StoreOperationException;
use App\Services\StoreStateService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class OutletStoreController extends Controller
{
    public function __construct(private readonly StoreStateService $stores) {}

    public function show(Request $request, string $outlet): JsonResponse
    {
        try {
            return ApiResponse::success($this->stores->state($request->user()->business_id, $outlet));
        } catch (StoreOperationException $error) {
            return $this->storeError($error);
        }
    }

    public function open(OpenStoreRequest $request, string $outlet): JsonResponse
    {
        try {
            return ApiResponse::success($this->stores->open(
                $request->user()->business_id,
                $outlet,
                $request->validated(),
                $request->user()->id,
                $request->user()->role,
            ));
        } catch (StoreOperationException $error) {
            return $this->storeError($error);
        }
    }

    public function close(CloseStoreRequest $request, string $outlet): JsonResponse
    {
        try {
            return ApiResponse::success($this->stores->close(
                $request->user()->business_id,
                $outlet,
                $request->validated(),
                $request->user()->id,
                $request->user()->role,
            ));
        } catch (StoreOperationException $error) {
            return $this->storeError($error);
        }
    }

    private function storeError(StoreOperationException $error): JsonResponse
    {
        return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
    }
}
