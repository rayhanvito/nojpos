<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\CreateRefundRequest;
use App\Services\RefundOperationException;
use App\Services\RefundService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;

class RefundController extends Controller
{
    public function __construct(private readonly RefundService $refunds) {}

    public function store(CreateRefundRequest $request, string $transaction): JsonResponse
    {
        try {
            $refund = $this->refunds->create(
                $request->user(),
                $transaction,
                $request->validated(),
                $request->attributes->get('idempotency_key') ?: $request->header('Idempotency-Key'),
            );
        } catch (RefundOperationException $error) {
            return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
        }

        return ApiResponse::success($refund, [], 201);
    }
}
