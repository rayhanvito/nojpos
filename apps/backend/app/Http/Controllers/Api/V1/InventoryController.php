<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\Inventory\CreateInventoryCountRequest;
use App\Http\Requests\Inventory\CreateInventoryTransferRequest;
use App\Http\Requests\Inventory\CreateInventoryWasteRequest;
use App\Services\InventoryOperationException;
use App\Services\InventoryService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class InventoryController extends Controller
{
    public function __construct(private readonly InventoryService $inventory) {}

    public function index(Request $request): JsonResponse
    {
        $data = $request->validate([
            'outlet_id' => ['nullable', 'uuid'],
            'product_id' => ['nullable', 'uuid'],
            'type' => ['nullable', 'string'],
            'search' => ['nullable', 'string'],
            'low_stock' => ['nullable', 'boolean'],
            'page' => ['nullable', 'integer', 'min:1'],
            'per_page' => ['nullable', 'integer', 'min:1', 'max:200'],
        ]);

        try {
            $businessId = $request->user()->business_id;
            $overview = $this->inventory->overview($businessId, $data);
            $overview['movements'] = $this->inventory->movements($businessId, $data)['items'];

            return ApiResponse::success($overview);
        } catch (InventoryOperationException $error) {
            return $this->inventoryError($error);
        }
    }

    public function movements(Request $request): JsonResponse
    {
        $data = $request->validate([
            'outlet_id' => ['nullable', 'uuid'],
            'product_id' => ['nullable', 'uuid'],
            'type' => ['nullable', 'string'],
            'from' => ['nullable', 'date'],
            'to' => ['nullable', 'date'],
            'page' => ['nullable', 'integer', 'min:1'],
            'per_page' => ['nullable', 'integer', 'min:1', 'max:200'],
        ]);

        try {
            return ApiResponse::success($this->inventory->movements($request->user()->business_id, $data));
        } catch (InventoryOperationException $error) {
            return $this->inventoryError($error);
        }
    }

    public function counts(CreateInventoryCountRequest $request): JsonResponse
    {
        try {
            return ApiResponse::success(
                $this->inventory->createCount($request->user()->business_id, $request->user()->id, $request->validated()),
                [],
                201,
            );
        } catch (InventoryOperationException $error) {
            return $this->inventoryError($error);
        }
    }

    public function waste(CreateInventoryWasteRequest $request): JsonResponse
    {
        try {
            return ApiResponse::success(
                $this->inventory->createWaste($request->user()->business_id, $request->user()->id, $request->validated()),
                [],
                201,
            );
        } catch (InventoryOperationException $error) {
            return $this->inventoryError($error);
        }
    }

    public function transfers(Request $request): JsonResponse
    {
        $data = $request->validate([
            'outlet_id' => ['nullable', 'uuid'],
            'status' => ['nullable', 'string'],
        ]);

        try {
            return ApiResponse::success($this->inventory->transfers($request->user()->business_id, $data));
        } catch (InventoryOperationException $error) {
            return $this->inventoryError($error);
        }
    }

    public function inTransit(Request $request): JsonResponse
    {
        $data = $request->validate([
            'outlet_id' => ['nullable', 'uuid'],
        ]);
        $data['status'] = 'in_transit';

        try {
            return ApiResponse::success($this->inventory->transfers($request->user()->business_id, $data));
        } catch (InventoryOperationException $error) {
            return $this->inventoryError($error);
        }
    }

    public function createTransfer(CreateInventoryTransferRequest $request): JsonResponse
    {
        try {
            return ApiResponse::success(
                $this->inventory->createTransfer($request->user()->business_id, $request->user()->id, $request->validated()),
                [],
                201,
            );
        } catch (InventoryOperationException $error) {
            return $this->inventoryError($error);
        }
    }

    public function sendTransfer(Request $request, string $transfer): JsonResponse
    {
        try {
            return ApiResponse::success($this->inventory->sendTransfer($request->user()->business_id, $request->user()->id, $transfer));
        } catch (InventoryOperationException $error) {
            return $this->inventoryError($error);
        }
    }

    public function receiveTransfer(Request $request, string $transfer): JsonResponse
    {
        try {
            return ApiResponse::success($this->inventory->receiveTransfer($request->user()->business_id, $request->user()->id, $transfer));
        } catch (InventoryOperationException $error) {
            return $this->inventoryError($error);
        }
    }

    public function cancelTransfer(Request $request, string $transfer): JsonResponse
    {
        try {
            return ApiResponse::success($this->inventory->cancelTransfer($request->user()->business_id, $request->user()->id, $transfer));
        } catch (InventoryOperationException $error) {
            return $this->inventoryError($error);
        }
    }

    public function purchase(Request $request): JsonResponse
    {
        $data = $request->validate([
            'outlet_id' => ['required', 'uuid'],
            'number' => ['nullable', 'string', 'max:100'],
            'supplier_name' => ['nullable', 'string', 'max:255'],
            'notes' => ['nullable', 'string'],
            'items' => ['required', 'array', 'min:1'],
            'items.*.product_id' => ['required', 'uuid'],
            'items.*.quantity' => ['required', 'integer', 'min:1'],
            'items.*.unit_cost' => ['required', 'integer', 'min:0'],
        ]);

        if (! in_array($request->user()->role, ['owner', 'admin'], true)) {
            return ApiResponse::error('FORBIDDEN', 'Only owner or admin can create inventory purchases.', [], 403);
        }

        try {
            return ApiResponse::success($this->inventory->purchase($request->user()->business_id, $request->user()->id, $data), [], 201);
        } catch (InventoryOperationException $error) {
            return $this->inventoryError($error);
        }
    }

    private function inventoryError(InventoryOperationException $error): JsonResponse
    {
        return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
    }
}
