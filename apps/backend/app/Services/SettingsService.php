<?php

namespace App\Services;

use App\Models\User;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class SettingsService
{
    public function __construct(private readonly AuditLogService $auditLog) {}

    /**
     * @return array<string, mixed>
     */
    public function aggregate(User $actor): array
    {
        $businessId = $this->requireBusiness($actor);
        $business = $this->business($businessId);
        $outlets = $this->outlets($businessId);
        $paymentMethods = $this->paymentMethods($businessId, $actor->role === 'cashier');
        $securitySetting = $this->securitySetting($businessId);
        $payload = [
            'business' => $this->businessPayload($business, $outlets[0] ?? null),
            'outlets' => array_map(fn (object $outlet): array => $this->outletPayload($outlet), $outlets),
            'payment_methods' => array_map(fn (object $config): array => $this->paymentMethodPayload($config), $paymentMethods),
            'security' => $this->securityPayload($securitySetting),
            'permissions' => $this->permissionsPayload($actor),
        ];

        return [
            'settings' => $payload,
            'meta' => [
                'config_version' => $this->configVersion($payload),
                'server_timestamp' => now()->toISOString(),
                'scope' => $actor->role === 'cashier' ? 'pos_safe' : 'admin',
            ],
        ];
    }

    /**
     * @return array<string, mixed>
     */
    public function updateBusiness(User $actor, array $data, array $context = []): array
    {
        $businessId = $this->requireBusiness($actor);

        return DB::transaction(function () use ($actor, $businessId, $data, $context): array {
            $business = DB::table('businesses')
                ->where('id', $businessId)
                ->lockForUpdate()
                ->first();

            if (! $business) {
                throw new SettingsException('NOT_FOUND', 'Business was not found.', [], 404);
            }

            $before = $this->businessPayload($business, null);
            $updates = [];
            if (array_key_exists('name', $data)) {
                $updates['name'] = $data['name'];
            }

            if ($updates !== []) {
                $updates['updated_at'] = now();
                DB::table('businesses')->where('id', $businessId)->update($updates);
            }

            $after = $this->businessPayload($this->business($businessId), null);
            $this->auditLog->record(
                businessId: $businessId,
                actorId: $actor->id,
                action: 'settings.update',
                entityType: 'business',
                entityId: $businessId,
                before: $before,
                after: $after,
                context: $this->auditContext($context),
            );

            return $this->aggregate($actor);
        });
    }

    /**
     * @return array<string, mixed>
     */
    public function updateOutlet(User $actor, string $outletId, array $data, array $context = []): array
    {
        $businessId = $this->requireBusiness($actor);

        return DB::transaction(function () use ($actor, $businessId, $outletId, $data, $context): array {
            $outlet = $this->resourceForBusiness('outlets', $outletId, $businessId, true);
            $before = $this->outletPayload($outlet);
            $updates = [];

            foreach (['name', 'timezone', 'service_charge_rate', 'tax_rate'] as $field) {
                if (array_key_exists($field, $data)) {
                    $updates[$field] = $data[$field];
                }
            }

            $operational = $data['operational_config'] ?? [];
            if (array_key_exists('store_open_close_enabled', $operational)) {
                $updates['store_open_close_enabled'] = (bool) $operational['store_open_close_enabled'];
            }

            $receipt = $data['receipt_config'] ?? [];
            $receiptFields = [
                'paper_width' => 'receipt_paper_width',
                'header_name' => 'receipt_header_name',
                'header_address' => 'receipt_header_address',
                'footer_note' => 'receipt_footer_note',
                'show_logo' => 'receipt_show_logo',
                'show_qris_info' => 'receipt_show_qris_info',
            ];
            foreach ($receiptFields as $input => $column) {
                if (array_key_exists($input, $receipt)) {
                    $updates[$column] = $receipt[$input];
                }
            }

            if ($updates !== []) {
                $updates['updated_at'] = now();
                DB::table('outlets')
                    ->where('business_id', $businessId)
                    ->where('id', $outletId)
                    ->update($updates);
            }

            $after = $this->outletPayload($this->resourceForBusiness('outlets', $outletId, $businessId));
            $this->auditLog->record(
                businessId: $businessId,
                actorId: $actor->id,
                action: 'settings.update',
                entityType: 'outlet',
                entityId: $outletId,
                before: $before,
                after: $after,
                context: $this->auditContext($context + ['outlet_id' => $outletId]),
            );

            return $this->aggregate($actor);
        });
    }

    /**
     * @return array<string, mixed>
     */
    public function updateSecurity(User $actor, array $data, array $context = []): array
    {
        $businessId = $this->requireBusiness($actor);

        return DB::transaction(function () use ($actor, $businessId, $data, $context): array {
            $setting = DB::table('business_security_settings')
                ->where('business_id', $businessId)
                ->lockForUpdate()
                ->first();

            if (! $setting) {
                $id = (string) Str::uuid();
                $now = now();
                DB::table('business_security_settings')->insert([
                    'id' => $id,
                    'business_id' => $businessId,
                    'updated_by' => $actor->id,
                    'created_at' => $now,
                    'updated_at' => $now,
                ]);
                $setting = DB::table('business_security_settings')
                    ->where('business_id', $businessId)
                    ->lockForUpdate()
                    ->first();
            }

            $before = $this->securityPayload($setting);
            $updates = $this->securityUpdates($data);

            if ($updates !== []) {
                $updates['updated_by'] = $actor->id;
                $updates['updated_at'] = now();
                DB::table('business_security_settings')
                    ->where('business_id', $businessId)
                    ->update($updates);
            }

            $afterSetting = DB::table('business_security_settings')->where('business_id', $businessId)->first();
            $after = $this->securityPayload($afterSetting);
            $this->auditLog->record(
                businessId: $businessId,
                actorId: $actor->id,
                action: 'settings.update',
                entityType: 'business_security_setting',
                entityId: $afterSetting->id,
                before: $before,
                after: $after,
                context: $this->auditContext($context),
            );

            return $this->aggregate($actor);
        });
    }

    /**
     * @return array<string, mixed>
     */
    public function updatePaymentMethod(User $actor, string $configId, array $data, array $context = []): array
    {
        $businessId = $this->requireBusiness($actor);

        return DB::transaction(function () use ($actor, $businessId, $configId, $data, $context): array {
            $config = $this->resourceForBusiness('payment_method_configs', $configId, $businessId, true);
            $before = $this->paymentMethodPayload($config);
            $updates = [];

            if (array_key_exists('outlet_id', $data) && $data['outlet_id'] !== null) {
                $this->resourceForBusiness('outlets', $data['outlet_id'], $businessId);
            }

            foreach (['method', 'is_cash', 'outlet_id'] as $field) {
                if (array_key_exists($field, $data)) {
                    $updates[$field] = $data[$field];
                }
            }

            if (array_key_exists('active', $data)) {
                $updates['deleted_at'] = $data['active'] ? null : now();
            }

            if ($updates !== []) {
                $updates['updated_at'] = now();
                DB::table('payment_method_configs')
                    ->where('business_id', $businessId)
                    ->where('id', $configId)
                    ->update($updates);
            }

            $after = $this->paymentMethodPayload($this->resourceForBusiness('payment_method_configs', $configId, $businessId));
            $this->auditLog->record(
                businessId: $businessId,
                actorId: $actor->id,
                action: 'settings.update',
                entityType: 'payment_method_config',
                entityId: $configId,
                before: $before,
                after: $after,
                context: $this->auditContext($context + ['outlet_id' => $after['outlet_id']]),
            );

            return $this->aggregate($actor);
        });
    }

    private function requireBusiness(User $actor): string
    {
        if (! $actor->business_id) {
            throw new SettingsException('FORBIDDEN', 'Settings are only available for business users.', [], 403);
        }

        return $actor->business_id;
    }

    private function business(string $businessId): object
    {
        $business = DB::table('businesses')->where('id', $businessId)->first();
        if (! $business) {
            throw new SettingsException('NOT_FOUND', 'Business was not found.', [], 404);
        }

        return $business;
    }

    /**
     * @return array<int, object>
     */
    private function outlets(string $businessId): array
    {
        return DB::table('outlets')
            ->where('business_id', $businessId)
            ->whereNull('deleted_at')
            ->orderBy('name')
            ->get()
            ->all();
    }

    /**
     * @return array<int, object>
     */
    private function paymentMethods(string $businessId, bool $posSafeOnly): array
    {
        return DB::table('payment_method_configs')
            ->where('business_id', $businessId)
            ->when($posSafeOnly, fn ($query) => $query->whereNull('deleted_at'))
            ->orderBy('outlet_id')
            ->orderByDesc('is_cash')
            ->orderBy('method')
            ->get()
            ->all();
    }

    private function securitySetting(string $businessId): object
    {
        $setting = DB::table('business_security_settings')->where('business_id', $businessId)->first();

        return $setting ?: (object) [
            'id' => null,
            'business_id' => $businessId,
            'pin_required_void' => true,
            'pin_required_refund' => true,
            'pin_required_discount_override' => true,
            'pin_required_cash_out_over_limit' => true,
            'pin_required_close_shift' => true,
            'pin_required_store_open_close' => true,
            'pin_required_settings_change' => false,
            'pin_lockout_max_attempts' => 5,
            'pin_lockout_decay_minutes' => 15,
            'idle_lock_timeout_seconds' => 180,
            'terminal_session_timeout_seconds' => 900,
            'updated_by' => null,
            'created_at' => null,
            'updated_at' => null,
        ];
    }

    /**
     * @return array<string, mixed>
     */
    private function businessPayload(object $business, ?object $defaultOutlet): array
    {
        return [
            'id' => $business->id,
            'name' => $business->name,
            'timezone' => $defaultOutlet->timezone ?? 'Asia/Jakarta',
            'currency' => 'IDR',
            'default_outlet_id' => $defaultOutlet->id ?? null,
            'updated_at' => $business->updated_at ?? null,
        ];
    }

    /**
     * @return array<string, mixed>
     */
    private function outletPayload(object $outlet): array
    {
        return [
            'id' => $outlet->id,
            'business_id' => $outlet->business_id,
            'name' => $outlet->name,
            'timezone' => $outlet->timezone ?? 'Asia/Jakarta',
            'transaction_config' => [
                'service_charge_rate' => (int) $outlet->service_charge_rate,
                'tax_rate' => (int) $outlet->tax_rate,
                'rounding_policy' => 'none',
            ],
            'receipt_config' => [
                'paper_width' => $outlet->receipt_paper_width ?? '58mm',
                'header_name' => $outlet->receipt_header_name ?? null,
                'header_address' => $outlet->receipt_header_address ?? null,
                'footer_note' => $outlet->receipt_footer_note ?? null,
                'show_logo' => (bool) ($outlet->receipt_show_logo ?? false),
                'show_qris_info' => (bool) ($outlet->receipt_show_qris_info ?? true),
            ],
            'operational_config' => [
                'cash_out_limit' => 500000,
                'shift_auto_expiry_time' => null,
                'store_open_close_enabled' => (bool) ($outlet->store_open_close_enabled ?? false),
            ],
            'updated_at' => $outlet->updated_at ?? null,
        ];
    }

    /**
     * @return array<string, mixed>
     */
    private function paymentMethodPayload(object $config): array
    {
        return [
            'id' => $config->id,
            'business_id' => $config->business_id,
            'outlet_id' => $config->outlet_id,
            'method' => $config->method,
            'enabled' => $config->deleted_at === null,
            'active' => $config->deleted_at === null,
            'is_cash' => (bool) $config->is_cash,
            'config' => [],
            'updated_at' => $config->updated_at ?? null,
        ];
    }

    /**
     * @return array<string, mixed>
     */
    private function securityPayload(object $setting): array
    {
        return [
            'id' => $setting->id,
            'pin_policy' => [
                'scope' => 'staff_business',
                'max_attempts' => (int) $setting->pin_lockout_max_attempts,
                'lockout_minutes' => (int) $setting->pin_lockout_decay_minutes,
                'owner_admin_can_clear' => true,
            ],
            'terminal_policy' => [
                'idle_lock_timeout_seconds' => (int) $setting->idle_lock_timeout_seconds,
                'session_timeout_seconds' => (int) $setting->terminal_session_timeout_seconds,
            ],
            'sensitive_actions' => [
                'void' => ['requires_pin' => (bool) $setting->pin_required_void, 'roles' => ['cashier_with_approval', 'admin', 'owner']],
                'refund' => ['requires_pin' => (bool) $setting->pin_required_refund, 'roles' => ['admin', 'owner']],
                'discount_override' => ['requires_pin' => (bool) $setting->pin_required_discount_override, 'roles' => ['cashier_with_approval', 'admin', 'owner']],
                'cash_out_over_limit' => ['requires_pin' => (bool) $setting->pin_required_cash_out_over_limit, 'roles' => ['admin', 'owner']],
                'close_shift' => ['requires_pin' => (bool) $setting->pin_required_close_shift, 'roles' => ['cashier', 'admin', 'owner']],
                'store_open_close' => ['requires_pin' => (bool) $setting->pin_required_store_open_close, 'roles' => ['admin', 'owner']],
                'settings_change' => ['requires_pin' => (bool) $setting->pin_required_settings_change, 'roles' => ['admin', 'owner']],
            ],
            'updated_by' => $setting->updated_by,
            'updated_at' => $setting->updated_at,
        ];
    }

    /**
     * @return array<string, mixed>
     */
    private function permissionsPayload(User $actor): array
    {
        $canUpdate = in_array($actor->role, ['owner', 'admin'], true);

        return [
            'current_user' => [
                'id' => $actor->id,
                'role' => $actor->role,
                'can_view_settings' => true,
                'can_update_settings' => $canUpdate,
                'can_update_security_settings' => $canUpdate,
            ],
            'roles' => [
                'owner' => ['can_view_settings' => true, 'can_update_settings' => true, 'can_update_security_settings' => true],
                'admin' => ['can_view_settings' => true, 'can_update_settings' => true, 'can_update_security_settings' => true],
                'cashier' => ['can_view_settings' => true, 'can_update_settings' => false, 'can_update_security_settings' => false],
            ],
        ];
    }

    /**
     * @return array<string, mixed>
     */
    private function securityUpdates(array $data): array
    {
        $updates = [];
        $actions = $data['sensitive_actions'] ?? [];
        $columns = [
            'void' => 'pin_required_void',
            'refund' => 'pin_required_refund',
            'discount_override' => 'pin_required_discount_override',
            'cash_out_over_limit' => 'pin_required_cash_out_over_limit',
            'close_shift' => 'pin_required_close_shift',
            'store_open_close' => 'pin_required_store_open_close',
            'settings_change' => 'pin_required_settings_change',
        ];

        foreach ($columns as $action => $column) {
            if (array_key_exists($action, $actions) && array_key_exists('requires_pin', $actions[$action])) {
                $updates[$column] = (bool) $actions[$action]['requires_pin'];
            }
        }

        if (array_key_exists('max_attempts', $data['pin_policy'] ?? [])) {
            $updates['pin_lockout_max_attempts'] = (int) $data['pin_policy']['max_attempts'];
        }
        if (array_key_exists('lockout_minutes', $data['pin_policy'] ?? [])) {
            $updates['pin_lockout_decay_minutes'] = (int) $data['pin_policy']['lockout_minutes'];
        }
        if (array_key_exists('idle_lock_timeout_seconds', $data['terminal_policy'] ?? [])) {
            $updates['idle_lock_timeout_seconds'] = (int) $data['terminal_policy']['idle_lock_timeout_seconds'];
        }
        if (array_key_exists('session_timeout_seconds', $data['terminal_policy'] ?? [])) {
            $updates['terminal_session_timeout_seconds'] = (int) $data['terminal_policy']['session_timeout_seconds'];
        }

        return $updates;
    }

    private function resourceForBusiness(string $table, string $id, string $businessId, bool $lock = false): object
    {
        $exists = DB::table($table)->where('id', $id)->exists();
        if (! $exists) {
            throw new SettingsException('NOT_FOUND', 'Resource was not found.', [], 404);
        }

        $query = DB::table($table)
            ->where('id', $id)
            ->where('business_id', $businessId);

        if ($lock) {
            $query->lockForUpdate();
        }

        $resource = $query->first();
        if (! $resource) {
            throw new SettingsException('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        return $resource;
    }

    /**
     * @return array<string, mixed>
     */
    private function auditContext(array $context): array
    {
        return array_filter([
            'outlet_id' => $context['outlet_id'] ?? null,
            'device_id' => $context['device_id'] ?? null,
            'request_id' => $context['request_id'] ?? null,
            'idempotency_key' => $context['idempotency_key'] ?? null,
            'event_version' => 1,
        ], fn (mixed $value): bool => $value !== null);
    }

    private function configVersion(array $payload): string
    {
        return hash('sha256', json_encode($payload, JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE));
    }
}
