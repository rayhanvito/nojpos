<?php

namespace App\Services;

use App\Models\User;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;

class StaffService
{
    public function __construct(private readonly AuditLogService $auditLog) {}

    public function create(Request $request, array $data): array
    {
        $actor = $request->user();
        $this->assertCanManageStaff((string) $actor->role);
        $this->assertCanAssignRole((string) $actor->role, (string) $data['role']);

        return DB::transaction(function () use ($request, $data, $actor): array {
            $staff = User::query()->create([
                'business_id' => $actor->business_id,
                'name' => $data['name'],
                'email' => $data['email'] ?? $this->generatedEmail((string) $actor->business_id),
                'role' => $data['role'],
                'password' => $data['password'] ?? bin2hex(random_bytes(16)),
                'pin_hash' => Hash::make($data['pin']),
            ]);

            $payload = $this->payload($staff);
            $this->auditLog->record(
                (string) $actor->business_id,
                (string) $actor->id,
                'staff.created',
                'user',
                $staff->id,
                null,
                $this->auditPayload($staff),
                $this->auditContext($request),
            );
            $this->auditLog->record(
                (string) $actor->business_id,
                (string) $actor->id,
                'staff.pin_updated',
                'user',
                $staff->id,
                null,
                ['id' => $staff->id, 'pin_set' => true],
                $this->auditContext($request),
            );

            return $payload;
        });
    }

    public function update(Request $request, User $staff, array $data): array
    {
        $actor = $request->user();
        $this->assertCanManageStaff((string) $actor->role);

        return DB::transaction(function () use ($request, $staff, $data, $actor): array {
            $fresh = User::query()
                ->where('business_id', $actor->business_id)
                ->where('id', $staff->id)
                ->whereIn('role', ['owner', 'admin', 'cashier'])
                ->lockForUpdate()
                ->first();

            if (! $fresh) {
                throw StaffServiceException::forbidden();
            }

            if ($fresh->role === 'owner' && $fresh->id !== $actor->id) {
                throw StaffServiceException::forbidden('Owner accounts can only be edited by themselves.');
            }

            if (array_key_exists('role', $data)) {
                if ($fresh->role === 'owner') {
                    throw StaffServiceException::forbidden('Owner role cannot be changed from staff management.');
                }
                $this->assertCanAssignRole((string) $actor->role, (string) $data['role']);
            }

            $before = $this->auditPayload($fresh);
            $pinChanged = false;

            if (array_key_exists('name', $data)) {
                $fresh->name = $data['name'];
            }
            if (array_key_exists('email', $data)) {
                $fresh->email = $data['email'] ?: $this->generatedEmail((string) $actor->business_id);
            }
            if (array_key_exists('role', $data)) {
                $fresh->role = $data['role'];
            }
            if (! empty($data['password'])) {
                $fresh->password = $data['password'];
            }
            if (! empty($data['pin'])) {
                $fresh->pin_hash = Hash::make($data['pin']);
                $pinChanged = true;
            }

            $fresh->save();
            $fresh->refresh();

            $this->auditLog->record(
                (string) $actor->business_id,
                (string) $actor->id,
                'staff.updated',
                'user',
                $fresh->id,
                $before,
                $this->auditPayload($fresh),
                $this->auditContext($request),
            );

            if ($pinChanged) {
                $this->auditLog->record(
                    (string) $actor->business_id,
                    (string) $actor->id,
                    'staff.pin_updated',
                    'user',
                    $fresh->id,
                    ['id' => $fresh->id, 'pin_set' => $before['pin_set']],
                    ['id' => $fresh->id, 'pin_set' => true],
                    $this->auditContext($request),
                );
            }

            return $this->payload($fresh);
        });
    }

    public function deactivate(Request $request, User $staff): void
    {
        $actor = $request->user();
        $this->assertCanManageStaff((string) $actor->role);

        DB::transaction(function () use ($request, $staff, $actor): void {
            $fresh = User::query()
                ->where('business_id', $actor->business_id)
                ->where('id', $staff->id)
                ->whereIn('role', ['admin', 'cashier'])
                ->lockForUpdate()
                ->first();

            if (! $fresh) {
                throw StaffServiceException::forbidden('Only admin and cashier staff in the current business can be deleted.');
            }

            if ($fresh->id === $actor->id) {
                throw StaffServiceException::forbidden('You cannot delete your own account from cashier settings.');
            }

            $before = $this->auditPayload($fresh);
            $fresh->delete();

            $this->auditLog->record(
                (string) $actor->business_id,
                (string) $actor->id,
                'staff.deactivated',
                'user',
                $fresh->id,
                $before,
                $before + ['deactivated' => true],
                $this->auditContext($request),
            );
        });
    }

    private function assertCanManageStaff(string $actorRole): void
    {
        if (! in_array($actorRole, ['owner', 'admin'], true)) {
            throw StaffServiceException::forbidden('You are not allowed to manage staff.');
        }
    }

    private function assertCanAssignRole(string $actorRole, string $targetRole): void
    {
        $rank = ['cashier' => 1, 'admin' => 2, 'owner' => 3, 'superadmin' => 4];
        if (! isset($rank[$actorRole], $rank[$targetRole]) || $rank[$targetRole] > $rank[$actorRole]) {
            throw StaffServiceException::forbidden('You cannot assign a role higher than your own.');
        }
    }

    public function payload(User $user): array
    {
        return [
            'id' => $user->id,
            'business_id' => $user->business_id,
            'name' => $user->name,
            'email' => $this->publicEmail((string) $user->email),
            'role' => $user->role,
        ];
    }

    private function auditPayload(User $user): array
    {
        return [
            'id' => $user->id,
            'business_id' => $user->business_id,
            'role' => $user->role,
            'has_email' => ! $this->isGeneratedEmail((string) $user->email),
            'pin_set' => filled($user->pin_hash),
        ];
    }

    private function generatedEmail(string $businessId): string
    {
        return uniqid('staff-', true).'@'.$businessId.'.staff.local';
    }

    private function publicEmail(string $email): string
    {
        return $this->isGeneratedEmail($email) ? '' : $email;
    }

    private function isGeneratedEmail(string $email): bool
    {
        return str_contains($email, '.staff.local') || str_ends_with($email, '@staff.local');
    }

    private function auditContext(Request $request): array
    {
        return [
            'request_id' => $request->header('X-Request-Id'),
            'idempotency_key' => $request->attributes->get('idempotency_key') ?: $request->header('Idempotency-Key'),
        ];
    }
}
