<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\VoidTransactionRequest;
use App\Services\VoidOperationException;
use App\Services\VoidService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;

class VoidController extends Controller
{
    public function __construct(private readonly VoidService $voids) {}

    public function store(VoidTransactionRequest $request): JsonResponse
    {
        try {
            $void = $this->voids->void($request->user(), $request, $request->validated());
        } catch (VoidOperationException $error) {
            return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
        }

        return ApiResponse::success($void);
    }
}
