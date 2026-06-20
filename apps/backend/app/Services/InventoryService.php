<?php

namespace App\Services;

use App\Support\Nojpos;
use Illuminate\Support\Collection;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class InventoryService
{
    public function __construct(private readonly BusinessClock $clock) {}

    /**
     * @param  array<string, mixed>  $filters
     * @return array<string, mixed>
     */
    public function overview(string $businessId, array $filters): array
    {
        $this->assertOutletBelongsToBusiness($businessId, $filters['outlet_id'] ?? null);
        $this->assertProductBelongsToBusiness($businessId, $filters['product_id'] ?? null);

        $outlets = DB::table('outlets')
            ->where('business_id', $businessId)
            ->when($filters['outlet_id'] ?? null, fn ($query, $outletId) => $query->where('id', $outletId))
            ->whereNull('deleted_at')
            ->get(['id', 'name'])
            ->keyBy('id');

        $products = DB::table('products')
            ->leftJoin('product_categories', 'product_categories.id', '=', 'products.product_category_id')
            ->where('products.business_id', $businessId)
            ->whereNull('products.deleted_at')
            ->when($filters['outlet_id'] ?? null, function ($query, $outletId): void {
                $query->where(function ($query) use ($outletId): void {
                    $query->whereNull('products.outlet_id')->orWhere('products.outlet_id', $outletId);
                });
            })
            ->when($filters['product_id'] ?? null, fn ($query, $productId) => $query->where('products.id', $productId))
            ->when($filters['search'] ?? null, fn ($query, $search) => $query->where('products.name', 'like', '%'.$search.'%'))
            ->orderBy('products.name')
            ->get([
                'products.id',
                'products.outlet_id',
                'products.product_category_id',
                'products.name',
                'products.barcode',
                'products.track_stock',
                'product_categories.name as category_name',
                'products.updated_at',
            ]);

        $items = [];
        foreach ($products as $product) {
            $candidateOutlets = $this->candidateOutletsForProduct($product, $outlets, $filters['outlet_id'] ?? null);
            foreach ($candidateOutlets as $outletId => $outletName) {
                $onHand = $this->stockOnHand($businessId, $outletId, $product->id);
                $inTransitOut = $this->inTransitQuantity($businessId, $outletId, $product->id, 'source');
                $inTransitIn = $this->inTransitQuantity($businessId, $outletId, $product->id, 'destination');

                if (($filters['low_stock'] ?? false) && $onHand > 0) {
                    continue;
                }

                $items[] = [
                    'product_id' => $product->id,
                    'product_name' => $product->name,
                    'sku' => $product->barcode,
                    'category_name' => $product->category_name,
                    'outlet_id' => $outletId,
                    'outlet_name' => $outletName,
                    'on_hand_quantity' => $onHand,
                    'reserved_quantity' => 0,
                    'in_transit_out_quantity' => $inTransitOut,
                    'in_transit_in_quantity' => $inTransitIn,
                    'available_quantity' => $onHand,
                    'low_stock_threshold' => null,
                    'updated_at' => $product->updated_at,
                    'business_id' => $businessId,
                    'product_category_id' => $product->product_category_id,
                    'name' => $product->name,
                    'track_stock' => (bool) $product->track_stock,
                    'stock_on_hand' => $onHand,
                    'warning' => $onHand < 0 ? 'NEGATIVE_STOCK_ALLOWED' : null,
                ];
            }
        }

        return [
            'items' => array_values($items),
            'page' => 1,
            'per_page' => count($items),
        ];
    }

    /**
     * @param  array<string, mixed>  $filters
     * @return array<string, mixed>
     */
    public function movements(string $businessId, array $filters): array
    {
        $this->assertOutletBelongsToBusiness($businessId, $filters['outlet_id'] ?? null);
        $this->assertProductBelongsToBusiness($businessId, $filters['product_id'] ?? null);

        $rows = DB::table('stock_movements')
            ->join('products', 'products.id', '=', 'stock_movements.product_id')
            ->join('outlets', 'outlets.id', '=', 'stock_movements.outlet_id')
            ->leftJoin('users', 'users.id', '=', 'stock_movements.actor_id')
            ->where('stock_movements.business_id', $businessId)
            ->whereNull('stock_movements.deleted_at')
            ->when($filters['outlet_id'] ?? null, fn ($query, $outletId) => $query->where('stock_movements.outlet_id', $outletId))
            ->when($filters['product_id'] ?? null, fn ($query, $productId) => $query->where('stock_movements.product_id', $productId))
            ->when($filters['type'] ?? null, fn ($query, $type) => $query->where('stock_movements.type', $type))
            ->when($filters['from'] ?? null, fn ($query, $from) => $query->where('stock_movements.created_at', '>=', $from))
            ->when($filters['to'] ?? null, fn ($query, $to) => $query->where('stock_movements.created_at', '<=', $to))
            ->orderByDesc('stock_movements.created_at')
            ->limit((int) ($filters['per_page'] ?? 100))
            ->get([
                'stock_movements.id',
                'stock_movements.product_id',
                'products.name as product_name',
                'stock_movements.outlet_id',
                'outlets.name as outlet_name',
                'stock_movements.type',
                'stock_movements.quantity_delta',
                'stock_movements.before_quantity',
                'stock_movements.after_quantity',
                'stock_movements.transaction_id',
                'stock_movements.purchase_id',
                'stock_movements.inventory_count_id',
                'stock_movements.waste_id',
                'stock_movements.inventory_transfer_id',
                'stock_movements.reason',
                'stock_movements.actor_id',
                'users.name as actor_name',
                'stock_movements.created_at',
            ]);

        return [
            'items' => $rows->map(fn (object $row): array => [
                'movement_id' => $row->id,
                'product_id' => $row->product_id,
                'product_name' => $row->product_name,
                'outlet_id' => $row->outlet_id,
                'outlet_name' => $row->outlet_name,
                'type' => $row->type,
                'quantity_delta' => (int) $row->quantity_delta,
                'before_quantity' => $row->before_quantity === null ? null : (int) $row->before_quantity,
                'after_quantity' => $row->after_quantity === null ? null : (int) $row->after_quantity,
                'reference_type' => $this->movementReferenceType($row),
                'reference_id' => $this->movementReferenceId($row),
                'reason' => $row->reason,
                'actor' => $row->actor_id ? ['id' => $row->actor_id, 'name' => $row->actor_name] : null,
                'created_at' => $row->created_at,
            ])->values()->all(),
        ];
    }

    /**
     * @param  array<string, mixed>  $data
     * @return array<string, mixed>
     */
    public function createCount(string $businessId, string $actorId, array $data): array
    {
        $this->assertOutletBelongsToBusiness($businessId, $data['outlet_id']);
        $this->assertProductsBelongToBusiness($businessId, collect($data['lines'])->pluck('product_id')->all());

        return DB::transaction(function () use ($businessId, $actorId, $data): array {
            $now = now();
            $outletId = $data['outlet_id'];
            $countId = (string) Str::uuid();
            $number = 'CNT-'.$this->documentSuffix($businessId, $outletId, $now);

            DB::table('inventory_counts')->insert([
                'id' => $countId,
                'business_id' => $businessId,
                'outlet_id' => $outletId,
                'actor_id' => $actorId,
                'number' => $number,
                'status' => 'completed',
                'notes' => $data['notes'] ?? null,
                'counted_at' => $data['counted_at'] ?? $now,
                'created_at' => $now,
                'updated_at' => $now,
            ]);

            foreach ($data['lines'] as $line) {
                $product = $this->lockProduct($businessId, $line['product_id']);
                $before = $this->stockOnHand($businessId, $outletId, $product->id);
                $counted = (int) $line['counted_quantity'];
                $delta = $counted - $before;

                if ($delta !== 0 && empty($line['reason'])) {
                    throw new InventoryOperationException('VALIDATION_ERROR', 'Reason is required for non-zero stock count variance.', ['reason' => ['Reason is required for adjusted lines.']], 422);
                }

                DB::table('inventory_count_lines')->insert([
                    'id' => (string) Str::uuid(),
                    'business_id' => $businessId,
                    'inventory_count_id' => $countId,
                    'product_id' => $product->id,
                    'system_quantity' => $before,
                    'counted_quantity' => $counted,
                    'delta_quantity' => $delta,
                    'reason' => $line['reason'] ?? null,
                    'created_at' => $now,
                    'updated_at' => $now,
                ]);

                if ($delta !== 0) {
                    $this->insertMovement($businessId, $outletId, $product->id, 'adjustment', $delta, $actorId, [
                        'inventory_count_id' => $countId,
                        'reason' => $line['reason'] ?? null,
                        'before_quantity' => $before,
                        'after_quantity' => $counted,
                    ]);
                }
            }

            Nojpos::audit($businessId, $actorId, 'inventory.count.finalize', 'inventory_count', $countId, null, ['id' => $countId, 'outlet_id' => $outletId, 'line_count' => count($data['lines'])]);

            return $this->countPayload($businessId, $countId);
        });
    }

    /**
     * @param  array<string, mixed>  $data
     * @return array<string, mixed>
     */
    public function createWaste(string $businessId, string $actorId, array $data): array
    {
        $this->assertOutletBelongsToBusiness($businessId, $data['outlet_id']);
        $this->assertProductBelongsToBusiness($businessId, $data['product_id']);

        return DB::transaction(function () use ($businessId, $actorId, $data): array {
            $now = now();
            $before = $this->stockOnHand($businessId, $data['outlet_id'], $data['product_id']);
            $quantity = (int) $data['quantity'];

            if ($before < $quantity) {
                throw new InventoryOperationException('STOCK_INSUFFICIENT', 'Stock is not sufficient for this inventory action.', ['available_quantity' => $before, 'requested_quantity' => $quantity], 409);
            }

            $wasteId = (string) Str::uuid();
            DB::table('inventory_waste_records')->insert([
                'id' => $wasteId,
                'business_id' => $businessId,
                'outlet_id' => $data['outlet_id'],
                'product_id' => $data['product_id'],
                'actor_id' => $actorId,
                'number' => 'WST-'.$this->documentSuffix($businessId, $data['outlet_id'], $now),
                'quantity' => $quantity,
                'reason' => $data['reason'],
                'status' => 'completed',
                'occurred_at' => $data['occurred_at'] ?? $now,
                'created_at' => $now,
                'updated_at' => $now,
            ]);

            $this->insertMovement($businessId, $data['outlet_id'], $data['product_id'], 'waste', -1 * $quantity, $actorId, [
                'waste_id' => $wasteId,
                'reason' => $data['reason'],
                'before_quantity' => $before,
                'after_quantity' => $before - $quantity,
            ]);

            Nojpos::audit($businessId, $actorId, 'inventory.waste.finalize', 'inventory_waste', $wasteId, ['quantity' => $before], ['quantity' => $before - $quantity, 'reason' => $data['reason']]);

            return $this->wastePayload($businessId, $wasteId);
        });
    }

    /**
     * @param  array<string, mixed>  $data
     * @return array<string, mixed>
     */
    public function createTransfer(string $businessId, string $actorId, array $data): array
    {
        if ($data['source_outlet_id'] === $data['destination_outlet_id']) {
            throw new InventoryOperationException('VALIDATION_ERROR', 'Source and destination outlets must be different.', ['destination_outlet_id' => ['Destination must be different from source.']], 422);
        }

        $this->assertOutletBelongsToBusiness($businessId, $data['source_outlet_id']);
        $this->assertOutletBelongsToBusiness($businessId, $data['destination_outlet_id']);
        $this->assertProductsBelongToBusiness($businessId, collect($data['lines'])->pluck('product_id')->all());

        return DB::transaction(function () use ($businessId, $actorId, $data): array {
            $now = now();
            $transferId = (string) Str::uuid();
            DB::table('inventory_transfers')->insert([
                'id' => $transferId,
                'business_id' => $businessId,
                'source_outlet_id' => $data['source_outlet_id'],
                'destination_outlet_id' => $data['destination_outlet_id'],
                'actor_id' => $actorId,
                'number' => 'TRF-'.$this->documentSuffix($businessId, $data['source_outlet_id'], $now),
                'status' => 'requested',
                'notes' => $data['notes'] ?? null,
                'created_at' => $now,
                'updated_at' => $now,
            ]);

            foreach ($data['lines'] as $line) {
                DB::table('inventory_transfer_lines')->insert([
                    'id' => (string) Str::uuid(),
                    'business_id' => $businessId,
                    'inventory_transfer_id' => $transferId,
                    'product_id' => $line['product_id'],
                    'requested_quantity' => (int) $line['quantity'],
                    'sent_quantity' => 0,
                    'received_quantity' => 0,
                    'created_at' => $now,
                    'updated_at' => $now,
                ]);
            }

            Nojpos::audit($businessId, $actorId, 'inventory.transfer.create', 'inventory_transfer', $transferId, null, ['id' => $transferId, 'status' => 'requested']);

            return $this->transferPayload($businessId, $transferId);
        });
    }

    /**
     * @return array<string, mixed>
     */
    public function sendTransfer(string $businessId, string $actorId, string $transferId): array
    {
        return DB::transaction(function () use ($businessId, $actorId, $transferId): array {
            $transfer = $this->lockTransfer($businessId, $transferId);
            if ($transfer->status !== 'requested') {
                throw new InventoryOperationException('TRANSFER_STATE_INVALID', 'Transfer cannot be sent from its current state.', ['status' => $transfer->status], 409);
            }

            $lines = $this->transferLines($businessId, $transferId);
            foreach ($lines as $line) {
                $this->lockProduct($businessId, $line->product_id);
                $before = $this->stockOnHand($businessId, $transfer->source_outlet_id, $line->product_id);
                $quantity = (int) $line->requested_quantity;
                if ($before < $quantity) {
                    throw new InventoryOperationException('STOCK_INSUFFICIENT', 'Stock is not sufficient for this transfer.', ['product_id' => $line->product_id, 'available_quantity' => $before, 'requested_quantity' => $quantity], 409);
                }

                DB::table('inventory_transfer_lines')
                    ->where('business_id', $businessId)
                    ->where('id', $line->id)
                    ->update(['sent_quantity' => $quantity, 'updated_at' => now()]);

                $this->insertMovement($businessId, $transfer->source_outlet_id, $line->product_id, 'transfer_out', -1 * $quantity, $actorId, [
                    'inventory_transfer_id' => $transferId,
                    'before_quantity' => $before,
                    'after_quantity' => $before - $quantity,
                ]);
            }

            DB::table('inventory_transfers')
                ->where('business_id', $businessId)
                ->where('id', $transferId)
                ->update(['status' => 'in_transit', 'sent_at' => now(), 'updated_at' => now()]);

            Nojpos::audit($businessId, $actorId, 'inventory.transfer.dispatch', 'inventory_transfer', $transferId, ['status' => 'requested'], ['status' => 'in_transit']);

            return $this->transferPayload($businessId, $transferId);
        });
    }

    /**
     * @return array<string, mixed>
     */
    public function receiveTransfer(string $businessId, string $actorId, string $transferId): array
    {
        return DB::transaction(function () use ($businessId, $actorId, $transferId): array {
            $transfer = $this->lockTransfer($businessId, $transferId);
            if ($transfer->status !== 'in_transit') {
                throw new InventoryOperationException('TRANSFER_STATE_INVALID', 'Transfer cannot be received from its current state.', ['status' => $transfer->status], 409);
            }

            $lines = $this->transferLines($businessId, $transferId);
            foreach ($lines as $line) {
                $quantity = (int) $line->sent_quantity - (int) $line->received_quantity;
                if ($quantity <= 0) {
                    continue;
                }

                $this->lockProduct($businessId, $line->product_id);
                $before = $this->stockOnHand($businessId, $transfer->destination_outlet_id, $line->product_id);
                DB::table('inventory_transfer_lines')
                    ->where('business_id', $businessId)
                    ->where('id', $line->id)
                    ->update(['received_quantity' => (int) $line->sent_quantity, 'updated_at' => now()]);

                $this->insertMovement($businessId, $transfer->destination_outlet_id, $line->product_id, 'transfer_in', $quantity, $actorId, [
                    'inventory_transfer_id' => $transferId,
                    'before_quantity' => $before,
                    'after_quantity' => $before + $quantity,
                ]);
            }

            DB::table('inventory_transfers')
                ->where('business_id', $businessId)
                ->where('id', $transferId)
                ->update(['status' => 'received', 'received_at' => now(), 'updated_at' => now()]);

            Nojpos::audit($businessId, $actorId, 'inventory.transfer.receive', 'inventory_transfer', $transferId, ['status' => 'in_transit'], ['status' => 'received']);

            return $this->transferPayload($businessId, $transferId);
        });
    }

    /**
     * @return array<string, mixed>
     */
    public function cancelTransfer(string $businessId, string $actorId, string $transferId): array
    {
        return DB::transaction(function () use ($businessId, $actorId, $transferId): array {
            $transfer = $this->lockTransfer($businessId, $transferId);
            if ($transfer->status !== 'requested') {
                throw new InventoryOperationException('TRANSFER_STATE_INVALID', 'Only requested transfers can be cancelled.', ['status' => $transfer->status], 409);
            }

            DB::table('inventory_transfers')
                ->where('business_id', $businessId)
                ->where('id', $transferId)
                ->update(['status' => 'cancelled', 'cancelled_at' => now(), 'updated_at' => now()]);

            Nojpos::audit($businessId, $actorId, 'inventory.transfer.cancel', 'inventory_transfer', $transferId, ['status' => 'requested'], ['status' => 'cancelled']);

            return $this->transferPayload($businessId, $transferId);
        });
    }

    /**
     * @param  array<string, mixed>  $filters
     * @return array<string, mixed>
     */
    public function transfers(string $businessId, array $filters): array
    {
        $this->assertOutletBelongsToBusiness($businessId, $filters['outlet_id'] ?? null);

        $rows = DB::table('inventory_transfers')
            ->join('outlets as source', 'source.id', '=', 'inventory_transfers.source_outlet_id')
            ->join('outlets as destination', 'destination.id', '=', 'inventory_transfers.destination_outlet_id')
            ->where('inventory_transfers.business_id', $businessId)
            ->whereNull('inventory_transfers.deleted_at')
            ->when($filters['status'] ?? null, fn ($query, $status) => $query->where('inventory_transfers.status', $status))
            ->when($filters['outlet_id'] ?? null, function ($query, $outletId): void {
                $query->where(function ($query) use ($outletId): void {
                    $query->where('inventory_transfers.source_outlet_id', $outletId)
                        ->orWhere('inventory_transfers.destination_outlet_id', $outletId);
                });
            })
            ->orderByDesc('inventory_transfers.created_at')
            ->get([
                'inventory_transfers.id',
                'inventory_transfers.number',
                'inventory_transfers.source_outlet_id',
                'source.name as source_outlet_name',
                'inventory_transfers.destination_outlet_id',
                'destination.name as destination_outlet_name',
                'inventory_transfers.status',
                'inventory_transfers.notes',
                'inventory_transfers.sent_at',
                'inventory_transfers.received_at',
                'inventory_transfers.created_at',
            ]);

        return [
            'items' => $rows->map(fn (object $row): array => $this->transferRowPayload($businessId, $row))->values()->all(),
        ];
    }

    /**
     * @return array<string, mixed>
     */
    public function purchase(string $businessId, string $actorId, array $data): array
    {
        $this->assertOutletBelongsToBusiness($businessId, $data['outlet_id']);
        $this->assertProductsBelongToBusiness($businessId, collect($data['items'])->pluck('product_id')->all());

        return DB::transaction(function () use ($businessId, $actorId, $data): array {
            $now = now();
            $purchaseId = (string) Str::uuid();
            $total = collect($data['items'])->sum(fn (array $item): int => (int) $item['quantity'] * (int) $item['unit_cost']);
            $number = $data['number'] ?? 'PO-'.$this->clock->documentTimestamp($this->clock->outletTimezone($businessId, $data['outlet_id']), $now);

            DB::table('inventory_purchases')->insert([
                'id' => $purchaseId,
                'business_id' => $businessId,
                'outlet_id' => $data['outlet_id'],
                'number' => $number,
                'supplier_name' => $data['supplier_name'] ?? null,
                'notes' => $data['notes'] ?? null,
                'total' => $total,
                'purchased_at' => $now,
                'created_at' => $now,
                'updated_at' => $now,
            ]);

            foreach ($data['items'] as $item) {
                $before = $this->stockOnHand($businessId, $data['outlet_id'], $item['product_id']);
                $subtotal = (int) $item['quantity'] * (int) $item['unit_cost'];
                DB::table('inventory_purchase_items')->insert([
                    'id' => (string) Str::uuid(),
                    'business_id' => $businessId,
                    'purchase_id' => $purchaseId,
                    'product_id' => $item['product_id'],
                    'quantity' => $item['quantity'],
                    'unit_cost' => $item['unit_cost'],
                    'subtotal' => $subtotal,
                    'created_at' => $now,
                    'updated_at' => $now,
                ]);

                $this->insertMovement($businessId, $data['outlet_id'], $item['product_id'], 'purchase', (int) $item['quantity'], $actorId, [
                    'purchase_id' => $purchaseId,
                    'before_quantity' => $before,
                    'after_quantity' => $before + (int) $item['quantity'],
                ]);
            }

            Nojpos::audit($businessId, $actorId, 'inventory.purchase.finalize', 'inventory_purchase', $purchaseId, null, ['id' => $purchaseId, 'total' => $total]);

            return $this->purchasePayload($businessId, $purchaseId);
        });
    }

    private function assertOutletBelongsToBusiness(string $businessId, ?string $outletId): void
    {
        if (! $outletId) {
            return;
        }

        if (! DB::table('outlets')->where('business_id', $businessId)->where('id', $outletId)->whereNull('deleted_at')->exists()) {
            throw new InventoryOperationException('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }
    }

    private function assertProductBelongsToBusiness(string $businessId, ?string $productId): void
    {
        if (! $productId) {
            return;
        }

        if (! DB::table('products')->where('business_id', $businessId)->where('id', $productId)->whereNull('deleted_at')->exists()) {
            throw new InventoryOperationException('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }
    }

    /**
     * @param  array<int, string>  $productIds
     */
    private function assertProductsBelongToBusiness(string $businessId, array $productIds): void
    {
        $ids = collect($productIds)->unique()->values();
        $count = DB::table('products')
            ->where('business_id', $businessId)
            ->whereIn('id', $ids->all())
            ->whereNull('deleted_at')
            ->count();

        if ($count !== $ids->count()) {
            throw new InventoryOperationException('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }
    }

    private function lockProduct(string $businessId, string $productId): object
    {
        $product = DB::table('products')
            ->where('business_id', $businessId)
            ->where('id', $productId)
            ->whereNull('deleted_at')
            ->lockForUpdate()
            ->first();

        if (! $product) {
            throw new InventoryOperationException('NOT_FOUND', 'Product not found.', [], 404);
        }

        return $product;
    }

    private function lockTransfer(string $businessId, string $transferId): object
    {
        $transfer = DB::table('inventory_transfers')
            ->where('business_id', $businessId)
            ->where('id', $transferId)
            ->whereNull('deleted_at')
            ->lockForUpdate()
            ->first();

        if (! $transfer) {
            throw new InventoryOperationException('NOT_FOUND', 'Inventory transfer not found.', [], 404);
        }

        return $transfer;
    }

    private function stockOnHand(string $businessId, string $outletId, string $productId): int
    {
        return (int) DB::table('stock_movements')
            ->where('business_id', $businessId)
            ->where('outlet_id', $outletId)
            ->where('product_id', $productId)
            ->whereNull('deleted_at')
            ->sum('quantity_delta');
    }

    private function inTransitQuantity(string $businessId, string $outletId, string $productId, string $side): int
    {
        $outletColumn = $side === 'source' ? 'inventory_transfers.source_outlet_id' : 'inventory_transfers.destination_outlet_id';

        return (int) DB::table('inventory_transfer_lines')
            ->join('inventory_transfers', 'inventory_transfers.id', '=', 'inventory_transfer_lines.inventory_transfer_id')
            ->where('inventory_transfer_lines.business_id', $businessId)
            ->where('inventory_transfer_lines.product_id', $productId)
            ->where('inventory_transfers.status', 'in_transit')
            ->where($outletColumn, $outletId)
            ->whereNull('inventory_transfer_lines.deleted_at')
            ->sum(DB::raw('inventory_transfer_lines.sent_quantity - inventory_transfer_lines.received_quantity'));
    }

    /**
     * @return Collection<int, object>
     */
    private function transferLines(string $businessId, string $transferId): Collection
    {
        return DB::table('inventory_transfer_lines')
            ->where('business_id', $businessId)
            ->where('inventory_transfer_id', $transferId)
            ->whereNull('deleted_at')
            ->orderBy('created_at')
            ->get();
    }

    /**
     * @param  array<string, mixed>  $extra
     */
    private function insertMovement(string $businessId, string $outletId, string $productId, string $type, int $delta, string $actorId, array $extra = []): void
    {
        DB::table('stock_movements')->insert([
            'id' => (string) Str::uuid(),
            'business_id' => $businessId,
            'outlet_id' => $outletId,
            'product_id' => $productId,
            'transaction_id' => $extra['transaction_id'] ?? null,
            'purchase_id' => $extra['purchase_id'] ?? null,
            'inventory_count_id' => $extra['inventory_count_id'] ?? null,
            'waste_id' => $extra['waste_id'] ?? null,
            'inventory_transfer_id' => $extra['inventory_transfer_id'] ?? null,
            'actor_id' => $actorId,
            'type' => $type,
            'quantity_delta' => $delta,
            'before_quantity' => $extra['before_quantity'] ?? null,
            'after_quantity' => $extra['after_quantity'] ?? null,
            'reason' => $extra['reason'] ?? null,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
    }

    private function documentSuffix(string $businessId, string $outletId, mixed $now): string
    {
        return $this->clock->documentTimestamp($this->clock->outletTimezone($businessId, $outletId), $now).'-'.Str::upper(Str::random(4));
    }

    /**
     * @return array<string, string>
     */
    private function candidateOutletsForProduct(object $product, Collection $outlets, ?string $filterOutletId): array
    {
        if ($filterOutletId && $outlets->has($filterOutletId)) {
            return [$filterOutletId => $outlets[$filterOutletId]->name];
        }

        if ($product->outlet_id && $outlets->has($product->outlet_id)) {
            return [$product->outlet_id => $outlets[$product->outlet_id]->name];
        }

        return $outlets->mapWithKeys(fn (object $outlet): array => [$outlet->id => $outlet->name])->all();
    }

    private function movementReferenceType(object $row): ?string
    {
        return match (true) {
            (bool) $row->transaction_id => 'transaction',
            (bool) $row->purchase_id => 'inventory_purchase',
            (bool) $row->inventory_count_id => 'inventory_count',
            (bool) $row->waste_id => 'inventory_waste',
            (bool) $row->inventory_transfer_id => 'inventory_transfer',
            default => null,
        };
    }

    private function movementReferenceId(object $row): ?string
    {
        return $row->transaction_id
            ?: $row->purchase_id
            ?: $row->inventory_count_id
            ?: $row->waste_id
            ?: $row->inventory_transfer_id;
    }

    /**
     * @return array<string, mixed>
     */
    private function countPayload(string $businessId, string $countId): array
    {
        $count = DB::table('inventory_counts')->where('business_id', $businessId)->where('id', $countId)->first();
        $lines = DB::table('inventory_count_lines')
            ->join('products', 'products.id', '=', 'inventory_count_lines.product_id')
            ->where('inventory_count_lines.business_id', $businessId)
            ->where('inventory_count_lines.inventory_count_id', $countId)
            ->orderBy('inventory_count_lines.created_at')
            ->get([
                'inventory_count_lines.id',
                'inventory_count_lines.product_id',
                'products.name as product_name',
                'inventory_count_lines.system_quantity',
                'inventory_count_lines.counted_quantity',
                'inventory_count_lines.delta_quantity',
                'inventory_count_lines.reason',
            ]);

        return [
            'id' => $count->id,
            'business_id' => $count->business_id,
            'outlet_id' => $count->outlet_id,
            'number' => $count->number,
            'status' => $count->status,
            'notes' => $count->notes,
            'counted_at' => $count->counted_at,
            'lines' => $lines->map(fn (object $line): array => [
                'id' => $line->id,
                'product_id' => $line->product_id,
                'product_name' => $line->product_name,
                'system_quantity' => (int) $line->system_quantity,
                'counted_quantity' => (int) $line->counted_quantity,
                'delta_quantity' => (int) $line->delta_quantity,
                'reason' => $line->reason,
            ])->values()->all(),
        ];
    }

    /**
     * @return array<string, mixed>
     */
    private function wastePayload(string $businessId, string $wasteId): array
    {
        $waste = DB::table('inventory_waste_records')
            ->join('products', 'products.id', '=', 'inventory_waste_records.product_id')
            ->where('inventory_waste_records.business_id', $businessId)
            ->where('inventory_waste_records.id', $wasteId)
            ->first(['inventory_waste_records.*', 'products.name as product_name']);

        return [
            'id' => $waste->id,
            'business_id' => $waste->business_id,
            'outlet_id' => $waste->outlet_id,
            'product_id' => $waste->product_id,
            'product_name' => $waste->product_name,
            'number' => $waste->number,
            'quantity' => (int) $waste->quantity,
            'reason' => $waste->reason,
            'status' => $waste->status,
            'occurred_at' => $waste->occurred_at,
        ];
    }

    /**
     * @return array<string, mixed>
     */
    private function transferPayload(string $businessId, string $transferId): array
    {
        $row = DB::table('inventory_transfers')
            ->join('outlets as source', 'source.id', '=', 'inventory_transfers.source_outlet_id')
            ->join('outlets as destination', 'destination.id', '=', 'inventory_transfers.destination_outlet_id')
            ->where('inventory_transfers.business_id', $businessId)
            ->where('inventory_transfers.id', $transferId)
            ->first([
                'inventory_transfers.*',
                'source.name as source_outlet_name',
                'destination.name as destination_outlet_name',
            ]);

        return $this->transferRowPayload($businessId, $row);
    }

    /**
     * @return array<string, mixed>
     */
    private function transferRowPayload(string $businessId, object $row): array
    {
        $lines = DB::table('inventory_transfer_lines')
            ->join('products', 'products.id', '=', 'inventory_transfer_lines.product_id')
            ->where('inventory_transfer_lines.business_id', $businessId)
            ->where('inventory_transfer_lines.inventory_transfer_id', $row->id)
            ->whereNull('inventory_transfer_lines.deleted_at')
            ->orderBy('inventory_transfer_lines.created_at')
            ->get([
                'inventory_transfer_lines.id',
                'inventory_transfer_lines.product_id',
                'products.name as product_name',
                'inventory_transfer_lines.requested_quantity',
                'inventory_transfer_lines.sent_quantity',
                'inventory_transfer_lines.received_quantity',
            ]);

        $linePayloads = $lines->map(function (object $line): array {
            $inTransit = max(0, (int) $line->sent_quantity - (int) $line->received_quantity);

            return [
                'id' => $line->id,
                'product_id' => $line->product_id,
                'product_name' => $line->product_name,
                'requested_quantity' => (int) $line->requested_quantity,
                'sent_quantity' => (int) $line->sent_quantity,
                'received_quantity' => (int) $line->received_quantity,
                'in_transit_quantity' => $inTransit,
            ];
        })->values()->all();

        $sentAt = $row->sent_at ?? null;

        return [
            'id' => $row->id,
            'transfer_id' => $row->id,
            'business_id' => $row->business_id ?? $businessId,
            'number' => $row->number,
            'source_outlet' => ['id' => $row->source_outlet_id, 'name' => $row->source_outlet_name],
            'destination_outlet' => ['id' => $row->destination_outlet_id, 'name' => $row->destination_outlet_name],
            'status' => $row->status,
            'notes' => $row->notes,
            'sent_at' => $sentAt,
            'received_at' => $row->received_at ?? null,
            'created_at' => $row->created_at,
            'age_seconds' => $sentAt ? max(0, now()->diffInSeconds($sentAt, true)) : 0,
            'lines' => $linePayloads,
        ];
    }

    /**
     * @return array<string, mixed>
     */
    private function purchasePayload(string $businessId, string $id): array
    {
        $purchase = DB::table('inventory_purchases')->where('business_id', $businessId)->where('id', $id)->first();
        $items = DB::table('inventory_purchase_items')
            ->where('business_id', $businessId)
            ->where('purchase_id', $id)
            ->orderBy('created_at')
            ->get(['id', 'product_id', 'quantity', 'unit_cost', 'subtotal'])
            ->map(fn (object $row): array => [
                'id' => $row->id,
                'product_id' => $row->product_id,
                'quantity' => (int) $row->quantity,
                'unit_cost' => (int) $row->unit_cost,
                'subtotal' => (int) $row->subtotal,
            ])
            ->values()
            ->all();

        return [
            'id' => $purchase->id,
            'business_id' => $purchase->business_id,
            'outlet_id' => $purchase->outlet_id,
            'number' => $purchase->number,
            'supplier_name' => $purchase->supplier_name,
            'notes' => $purchase->notes,
            'total' => (int) $purchase->total,
            'purchased_at' => $purchase->purchased_at,
            'items' => $items,
        ];
    }
}
