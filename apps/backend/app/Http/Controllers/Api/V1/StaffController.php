<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\User;
use App\Support\ApiResponse;
use App\Support\Nojpos;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;

class StaffController extends Controller
{
    public function index(Request $request): JsonResponse
    {
        $data = $request->validate([
            'outlet_id' => ['nullable', 'uuid'],
            'search' => ['nullable', 'string'],
        ]);

        $businessId = $request->user()->business_id;
        if (($data['outlet_id'] ?? null) && ! DB::table('outlets')->where('business_id', $businessId)->where('id', $data['outlet_id'])->exists()) {
            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        $staff = User::query()
            ->where('business_id', $businessId)
            ->whereIn('role', ['owner', 'admin', 'cashier'])
            ->when($data['search'] ?? null, function ($query, string $search): void {
                $query->where(function ($query) use ($search): void {
                    $query->where('name', 'like', '%'.$search.'%')
                        ->orWhere('email', 'like', '%'.$search.'%');
                });
            })
            ->orderBy('name')
            ->get()
            ->map(fn (User $user): array => $this->staffPayload($user))
            ->values()
            ->all();

        return ApiResponse::success(['staff' => $staff]);
    }

    public function store(Request $request): JsonResponse
    {
        $actor = $request->user();
        $data = $request->validate([
            'name' => ['required', 'string', 'max:255'],
            'email' => ['nullable', 'email', 'max:255', 'unique:users,email'],
            'role' => ['required', 'in:admin,cashier'],
            'password' => ['nullable', 'string', 'min:8'],
            'pin' => ['required', 'digits_between:4,8'],
        ]);

        $staff = User::query()->create([
            'business_id' => $actor->business_id,
            'name' => $data['name'],
            'email' => $data['email'] ?? $this->generatedEmail($actor->business_id),
            'role' => $data['role'],
            'password' => $data['password'] ?? bin2hex(random_bytes(16)),
            'pin_hash' => Hash::make($data['pin']),
        ]);

        Nojpos::audit($actor->business_id, $actor->id, 'staff.create', 'user', $staff->id);

        return ApiResponse::success(['staff' => $this->staffPayload($staff)], [], 201);
    }

    public function update(Request $request, User $staff): JsonResponse
    {
        $actor = $request->user();
        if ($staff->business_id !== $actor->business_id || ! in_array($staff->role, ['owner', 'admin', 'cashier'], true)) {
            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        if ($staff->role === 'owner' && $staff->id !== $actor->id) {
            return ApiResponse::error('FORBIDDEN', 'Owner accounts can only be edited by themselves.', [], 403);
        }

        $data = $request->validate([
            'name' => ['sometimes', 'required', 'string', 'max:255'],
            'email' => ['sometimes', 'nullable', 'email', 'max:255', 'unique:users,email,'.$staff->id],
            'role' => ['sometimes', 'required', 'in:admin,cashier'],
            'password' => ['sometimes', 'nullable', 'string', 'min:8'],
            'pin' => ['sometimes', 'nullable', 'digits_between:4,8'],
        ]);

        if (array_key_exists('name', $data)) {
            $staff->name = $data['name'];
        }
        if (array_key_exists('email', $data)) {
            $staff->email = $data['email'] ?: $this->generatedEmail($actor->business_id);
        }
        if (array_key_exists('role', $data) && $staff->role !== 'owner') {
            $staff->role = $data['role'];
        }
        if (! empty($data['password'])) {
            $staff->password = $data['password'];
        }
        if (! empty($data['pin'])) {
            $staff->pin_hash = Hash::make($data['pin']);
        }
        $staff->save();

        Nojpos::audit($actor->business_id, $actor->id, 'staff.update', 'user', $staff->id);

        return ApiResponse::success(['staff' => $this->staffPayload($staff->refresh())]);
    }

    public function destroy(Request $request, User $staff): JsonResponse
    {
        $actor = $request->user();
        if ($staff->business_id !== $actor->business_id || ! in_array($staff->role, ['admin', 'cashier'], true)) {
            return ApiResponse::error('FORBIDDEN', 'Only admin and cashier staff in the current business can be deleted.', [], 403);
        }

        if ($staff->id === $actor->id) {
            return ApiResponse::error('FORBIDDEN', 'You cannot delete your own account from cashier settings.', [], 403);
        }

        $staff->delete();
        Nojpos::audit($actor->business_id, $actor->id, 'staff.delete', 'user', $staff->id);

        return ApiResponse::success(['deleted' => true]);
    }

    private function staffPayload(User $user): array
    {
        return [
            'id' => $user->id,
            'business_id' => $user->business_id,
            'name' => $user->name,
            'email' => str_ends_with($user->email, '@staff.local') ? '' : $user->email,
            'role' => $user->role,
        ];
    }

    private function generatedEmail(string $businessId): string
    {
        return uniqid('staff-', true).'@'.$businessId.'.staff.local';
    }
}
