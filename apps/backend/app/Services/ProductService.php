<?php

namespace App\Services;

use App\Models\Product;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class ProductService
{
    public function __construct(private readonly AuditLogService $auditLog) {}

    public function create(Request $request, array $data): array
    {
        $businessId = $request->user()->business_id;
        $actorId = $request->user()->id;

        return DB::transaction(function () use ($request, $data, $businessId, $actorId): array {
            $this->assertRelations($businessId, $data);
            $this->assertBarcodeAvailable($businessId, $data['barcode'] ?? null, null);

            $product = Product::query()->create([
                'business_id' => $businessId,
                'outlet_id' => $data['outlet_id'] ?? null,
                'product_category_id' => $data['product_category_id'] ?? null,
                'name' => $data['name'],
                'barcode' => $data['barcode'] ?? null,
                'price' => (int) $data['price'],
                'track_stock' => (bool) ($data['track_stock'] ?? false),
            ]);

            $payload = $this->payload($product->id, $businessId);
            $this->auditLog->record(
                $businessId,
                $actorId,
                'product.created',
                'product',
                $product->id,
                null,
                $payload,
                $this->auditContext($request, $payload['outlet_id'] ?? null),
            );

            return $payload;
        });
    }

    public function update(Request $request, string $productId, array $data): ?array
    {
        $businessId = $request->user()->business_id;
        $actorId = $request->user()->id;

        return DB::transaction(function () use ($request, $productId, $data, $businessId, $actorId): ?array {
            $product = Product::query()
                ->where('business_id', $businessId)
                ->where('id', $productId)
                ->lockForUpdate()
                ->first();

            if (! $product) {
                return null;
            }

            $this->assertRelations($businessId, $data);
            $this->assertBarcodeAvailable($businessId, $data['barcode'] ?? null, $productId);

            $before = $this->payload($productId, $businessId);
            $updates = [];

            foreach (['outlet_id', 'product_category_id', 'name', 'barcode', 'price', 'track_stock'] as $field) {
                if (array_key_exists($field, $data)) {
                    $updates[$field] = $field === 'price' ? (int) $data[$field] : $data[$field];
                }
            }

            if ($updates !== []) {
                $updates['updated_at'] = now();
                $product->update($updates);
            }

            $after = $this->payload($productId, $businessId);
            $this->auditLog->record(
                $businessId,
                $actorId,
                'product.updated',
                'product',
                $productId,
                $before,
                $after,
                $this->auditContext($request, $after['outlet_id'] ?? null),
            );

            return $after;
        });
    }

    public function archive(Request $request, string $productId): ?array
    {
        $businessId = $request->user()->business_id;
        $actorId = $request->user()->id;

        return DB::transaction(function () use ($request, $productId, $businessId, $actorId): ?array {
            $product = Product::query()
                ->where('business_id', $businessId)
                ->where('id', $productId)
                ->lockForUpdate()
                ->first();

            if (! $product) {
                return null;
            }

            $before = $this->payload($productId, $businessId);
            $product->delete();
            $after = $before + ['archived' => true];

            $this->auditLog->record(
                $businessId,
                $actorId,
                'product.archived',
                'product',
                $productId,
                $before,
                $after,
                $this->auditContext($request, $before['outlet_id'] ?? null),
            );

            return $after;
        });
    }

    private function assertRelations(string $businessId, array $data): void
    {
        if (($data['outlet_id'] ?? null) && ! $this->belongsToBusiness('outlets', $businessId, $data['outlet_id'])) {
            throw ProductServiceException::forbidden();
        }

        if (($data['product_category_id'] ?? null) && ! $this->belongsToBusiness('product_categories', $businessId, $data['product_category_id'])) {
            throw ProductServiceException::forbidden();
        }
    }

    private function assertBarcodeAvailable(string $businessId, ?string $barcode, ?string $exceptProductId): void
    {
        if ($barcode === null || $barcode === '') {
            return;
        }

        $exists = Product::query()
            ->where('business_id', $businessId)
            ->where('barcode', $barcode)
            ->when($exceptProductId, fn ($query) => $query->where('id', '!=', $exceptProductId))
            ->lockForUpdate()
            ->exists();

        if ($exists) {
            throw ProductServiceException::duplicateBarcode();
        }
    }

    private function belongsToBusiness(string $table, string $businessId, string $id): bool
    {
        return DB::table($table)
            ->where('business_id', $businessId)
            ->where('id', $id)
            ->whereNull('deleted_at')
            ->exists();
    }

    private function payload(string $id, string $businessId): array
    {
        $product = Product::query()
            ->where('business_id', $businessId)
            ->where('id', $id)
            ->firstOrFail(['id', 'business_id', 'outlet_id', 'product_category_id', 'name', 'barcode', 'price', 'track_stock']);

        return [
            'id' => $product->id,
            'business_id' => $product->business_id,
            'outlet_id' => $product->outlet_id,
            'product_category_id' => $product->product_category_id,
            'name' => $product->name,
            'barcode' => $product->barcode,
            'price' => (int) $product->price,
            'track_stock' => (bool) $product->track_stock,
        ];
    }

    private function auditContext(Request $request, ?string $outletId): array
    {
        return [
            'outlet_id' => $outletId,
            'device_id' => $request->input('device_id'),
            'request_id' => $request->header('X-Request-Id'),
            'idempotency_key' => $request->attributes->get('idempotency_key') ?: $request->header('Idempotency-Key'),
        ];
    }
}
