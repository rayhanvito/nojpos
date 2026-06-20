<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Services\ParkedOrderException;
use App\Services\ParkedOrderService;
use App\Services\TerminalContextException;
use App\Services\TerminalContextService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class ParkedOrderController extends Controller
{
    public function __construct(
        private readonly ParkedOrderService $parkedOrders,
        private readonly TerminalContextService $terminalContext,
    ) {}

    public function index(Request $request): JsonResponse
    {
        $businessId = $request->user()->business_id;
        $orders = DB::table('transactions')
            ->where('business_id', $businessId)
            ->where('status', 'held')
            ->whereNull('deleted_at')
            ->when($request->query('outlet_id'), fn ($query, string $outletId) => $query->where('outlet_id', $outletId))
            ->orderByDesc('updated_at')
            ->get()
            ->map(fn (object $order): array => $this->parkedOrders->payload($businessId, $order->id))
            ->values()
            ->all();

        return ApiResponse::success(['parked_orders' => $orders]);
    }

    public function show(Request $request, string $transaction): JsonResponse
    {
        $businessId = $request->user()->business_id;
        if (! $this->exists($businessId, $transaction)) {
            return ApiResponse::error('NOT_FOUND', 'Parked order not found.', [], 404);
        }

        return ApiResponse::success($this->parkedOrders->payload($businessId, $transaction));
    }

    public function store(Request $request): JsonResponse
    {
        $data = $request->validate($this->orderRules(requireContext: true));

        try {
            $this->terminalContext->fromPayload($request, $data, requireOpenShift: true);
            $this->assertOptionalRelations($request->user()->business_id, $data);

            return ApiResponse::success($this->parkedOrders->create($request, $data), [], 201);
        } catch (TerminalContextException $error) {
            return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
        } catch (ParkedOrderException $error) {
            return $this->parkedOrderError($error);
        } catch (\RuntimeException $error) {
            return $this->quoteError($error);
        }
    }

    public function update(Request $request, string $transaction): JsonResponse
    {
        $data = $request->validate(array_merge($this->orderRules(requireContext: false), [
            'expected_revision' => ['required', 'integer', 'min:1'],
            'device_id' => ['required', 'uuid'],
        ]));

        try {
            $this->assertDeviceCanTouchParkedOrder($request, $transaction, $data['device_id']);
            $this->assertOptionalRelations($request->user()->business_id, $data);

            return ApiResponse::success($this->parkedOrders->update($request, $transaction, $data));
        } catch (ParkedOrderException $error) {
            return $this->parkedOrderError($error);
        } catch (\RuntimeException $error) {
            return $this->quoteError($error);
        }
    }

    public function acquireLease(Request $request, string $transaction): JsonResponse
    {
        return $this->leaseAction($request, $transaction, 'acquireLease');
    }

    public function refreshLease(Request $request, string $transaction): JsonResponse
    {
        return $this->leaseAction($request, $transaction, 'refreshLease');
    }

    public function releaseLease(Request $request, string $transaction): JsonResponse
    {
        return $this->leaseAction($request, $transaction, 'releaseLease');
    }

    public function destroy(Request $request, string $transaction): JsonResponse
    {
        $data = $request->validate([
            'expected_revision' => ['required', 'integer', 'min:1'],
            'device_id' => ['required', 'uuid'],
        ]);

        try {
            $this->assertDeviceCanTouchParkedOrder($request, $transaction, $data['device_id']);

            return ApiResponse::success($this->parkedOrders->cancel($request, $transaction, $data));
        } catch (ParkedOrderException $error) {
            return $this->parkedOrderError($error);
        }
    }

    private function leaseAction(Request $request, string $transaction, string $method): JsonResponse
    {
        $data = $request->validate([
            'device_id' => ['required', 'uuid'],
        ]);

        try {
            $this->assertDeviceCanTouchParkedOrder($request, $transaction, $data['device_id']);

            return ApiResponse::success($this->parkedOrders->{$method}($request, $transaction, $data));
        } catch (ParkedOrderException $error) {
            return $this->parkedOrderError($error);
        }
    }

    private function orderRules(bool $requireContext): array
    {
        $contextRule = $requireContext ? 'required' : 'nullable';

        return [
            'outlet_id' => [$contextRule, 'uuid'],
            'device_id' => [$contextRule, 'uuid'],
            'cashier_id' => [$contextRule, 'uuid'],
            'shift_id' => [$contextRule, 'uuid'],
            'customer_id' => ['nullable', 'uuid'],
            'served_by' => ['nullable', 'uuid'],
            'cart_discount' => ['nullable', 'integer', 'min:0'],
            'cart_discount_type' => ['nullable', 'in:amount,percent'],
            'notes' => ['nullable', 'string'],
            'items' => ['required', 'array', 'min:1'],
            'items.*.product_id' => ['required', 'uuid'],
            'items.*.quantity' => ['required', 'integer', 'min:1'],
            'items.*.unit_price' => ['nullable', 'integer', 'min:0'],
            'items.*.discount' => ['nullable', 'integer', 'min:0'],
            'items.*.discount_type' => ['nullable', 'in:amount,percent'],
        ];
    }

    private function assertDeviceCanTouchParkedOrder(Request $request, string $transactionId, string $deviceId): void
    {
        $businessId = $request->user()->business_id;
        $transaction = DB::table('transactions')
            ->where('business_id', $businessId)
            ->where('id', $transactionId)
            ->where('status', 'held')
            ->whereNull('deleted_at')
            ->first();

        if (! $transaction) {
            throw new ParkedOrderException('NOT_FOUND', 'Parked order not found.', 404);
        }

        $device = DB::table('devices')
            ->where('business_id', $businessId)
            ->where('id', $deviceId)
            ->where('outlet_id', $transaction->outlet_id)
            ->whereNull('deleted_at')
            ->first();

        if (! $device) {
            throw new ParkedOrderException('DEVICE_NOT_ENROLLED', 'Device is not enrolled for this outlet.', 403);
        }
    }

    private function assertOptionalRelations(string $businessId, array $data): void
    {
        if (($data['customer_id'] ?? null) && ! DB::table('customers')->where('business_id', $businessId)->where('id', $data['customer_id'])->exists()) {
            throw new ParkedOrderException('FORBIDDEN', 'Resource is outside the current business scope.', 403);
        }

        if (($data['served_by'] ?? null) && ! DB::table('users')->where('business_id', $businessId)->where('id', $data['served_by'])->exists()) {
            throw new ParkedOrderException('FORBIDDEN', 'Resource is outside the current business scope.', 403);
        }
    }

    private function exists(string $businessId, string $transactionId): bool
    {
        return DB::table('transactions')
            ->where('business_id', $businessId)
            ->where('id', $transactionId)
            ->where('status', 'held')
            ->whereNull('deleted_at')
            ->exists();
    }

    private function parkedOrderError(ParkedOrderException $error): JsonResponse
    {
        return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
    }

    private function quoteError(\RuntimeException $error): JsonResponse
    {
        if ($error->getMessage() === 'PRODUCT_NOT_FOUND') {
            return ApiResponse::error('PRODUCT_NOT_FOUND', 'Product is not available for this business.', [], 404);
        }

        return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
    }
}
