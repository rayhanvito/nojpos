<?php

namespace App\Services;

use App\Models\ProductCategory;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;

class CategoryService
{
    public function __construct(private readonly AuditLogService $auditLog) {}

    public function create(Request $request, array $data): array
    {
        $businessId = $request->user()->business_id;
        $actorId = $request->user()->id;

        return DB::transaction(function () use ($request, $data, $businessId, $actorId): array {
            $category = ProductCategory::query()->create([
                'business_id' => $businessId,
                'name' => $data['name'],
            ]);

            $payload = $this->payload($category->id, $businessId);
            $this->auditLog->record(
                $businessId,
                $actorId,
                'category.created',
                'product_category',
                $category->id,
                null,
                $payload,
                $this->auditContext($request),
            );

            return $payload;
        });
    }

    public function update(Request $request, string $categoryId, array $data): ?array
    {
        $businessId = $request->user()->business_id;
        $actorId = $request->user()->id;

        return DB::transaction(function () use ($request, $categoryId, $data, $businessId, $actorId): ?array {
            $category = ProductCategory::query()
                ->where('business_id', $businessId)
                ->where('id', $categoryId)
                ->lockForUpdate()
                ->first();

            if (! $category) {
                return null;
            }

            $before = $this->payload($categoryId, $businessId);
            $category->update(['name' => $data['name']]);
            $after = $this->payload($categoryId, $businessId);

            $this->auditLog->record(
                $businessId,
                $actorId,
                'category.updated',
                'product_category',
                $categoryId,
                $before,
                $after,
                $this->auditContext($request),
            );

            return $after;
        });
    }

    public function archive(Request $request, string $categoryId): ?array
    {
        $businessId = $request->user()->business_id;
        $actorId = $request->user()->id;

        return DB::transaction(function () use ($request, $categoryId, $businessId, $actorId): ?array {
            $category = ProductCategory::query()
                ->where('business_id', $businessId)
                ->where('id', $categoryId)
                ->lockForUpdate()
                ->first();

            if (! $category) {
                return null;
            }

            $before = $this->payload($categoryId, $businessId);
            $category->delete();
            $after = $before + ['archived' => true];

            $this->auditLog->record(
                $businessId,
                $actorId,
                'category.archived',
                'product_category',
                $categoryId,
                $before,
                $after,
                $this->auditContext($request),
            );

            return $after;
        });
    }

    private function payload(string $id, string $businessId): array
    {
        $category = ProductCategory::query()
            ->where('business_id', $businessId)
            ->where('id', $id)
            ->firstOrFail(['id', 'business_id', 'name']);

        return [
            'id' => $category->id,
            'business_id' => $category->business_id,
            'name' => $category->name,
        ];
    }

    private function auditContext(Request $request): array
    {
        return [
            'request_id' => $request->header('X-Request-Id'),
            'idempotency_key' => $request->attributes->get('idempotency_key') ?: $request->header('Idempotency-Key'),
        ];
    }
}
