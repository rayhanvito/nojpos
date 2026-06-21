<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\LockTerminalRequest;
use App\Http\Requests\TerminalLockStateRequest;
use App\Http\Requests\UnlockTerminalRequest;
use App\Services\TerminalLockException;
use App\Services\TerminalLockService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;

class TerminalLockController extends Controller
{
    public function __construct(private readonly TerminalLockService $terminalLocks) {}

    public function state(TerminalLockStateRequest $request): JsonResponse
    {
        try {
            return ApiResponse::success($this->terminalLocks->state($request->user(), $request->validated()));
        } catch (TerminalLockException $error) {
            return $this->terminalLockError($error);
        }
    }

    public function lock(LockTerminalRequest $request): JsonResponse
    {
        try {
            return ApiResponse::success($this->terminalLocks->lock($request->user(), $request->validated()));
        } catch (TerminalLockException $error) {
            return $this->terminalLockError($error);
        }
    }

    public function unlock(UnlockTerminalRequest $request): JsonResponse
    {
        try {
            return ApiResponse::success($this->terminalLocks->unlock($request->user(), $request->validated()));
        } catch (TerminalLockException $error) {
            return $this->terminalLockError($error);
        }
    }

    private function terminalLockError(TerminalLockException $error): JsonResponse
    {
        return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
    }
}
