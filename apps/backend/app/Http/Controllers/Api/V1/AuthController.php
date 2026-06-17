<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\Business;
use App\Models\Outlet;
use App\Models\User;
use App\Support\ApiResponse;
use App\Support\Nojpos;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;

class AuthController extends Controller
{
    public function login(Request $request): JsonResponse
    {
        $data = $request->validate([
            'email' => ['required', 'email'],
            'password' => ['required', 'string'],
            'device_uuid' => ['required', 'string'],
        ]);

        $user = User::query()->where('email', $data['email'])->first();

        if (! $user || ! Hash::check($data['password'], $user->password)) {
            return ApiResponse::error('INVALID_CREDENTIALS', 'Email or password is invalid.', [], 422);
        }

        $plainToken = Str::random(48);
        DB::table('personal_access_tokens')->insert([
            'id' => (string) Str::uuid(),
            'tokenable_type' => User::class,
            'tokenable_id' => $user->id,
            'name' => $data['device_uuid'],
            'token' => hash('sha256', $plainToken),
            'abilities' => json_encode(['*']),
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        Nojpos::audit($user->business_id, $user->id, 'auth.login');

        return ApiResponse::success([
            'token' => $plainToken,
            'user' => $this->userPayload($user),
            'business' => Business::query()->where('id', $user->business_id)->first()?->only(['id', 'name']),
            'outlets' => Outlet::query()->get(['id', 'business_id', 'name'])->values()->all(),
        ]);
    }

    public function pinSwitch(Request $request): JsonResponse
    {
        $data = $request->validate([
            'pin' => ['required', 'string'],
            'device_id' => ['required', 'uuid'],
            'outlet_id' => ['required', 'uuid'],
        ]);

        $actor = $request->user();
        $this->forbidUnlessSameBusiness('devices', $data['device_id'], $actor->business_id);
        $this->forbidUnlessSameBusiness('outlets', $data['outlet_id'], $actor->business_id);

        $attempt = DB::table('pin_attempts')
            ->where('business_id', $actor->business_id)
            ->where('device_id', $data['device_id'])
            ->where('outlet_id', $data['outlet_id'])
            ->first();

        if ($attempt && $attempt->locked_until && now()->lessThan($attempt->locked_until)) {
            return ApiResponse::error('PIN_LOCKED', 'PIN switch is temporarily locked.', [], 429);
        }

        $cashier = User::query()
            ->where('business_id', $actor->business_id)
            ->where('role', 'cashier')
            ->get()
            ->first(fn (User $candidate): bool => Hash::check($data['pin'], $candidate->pin_hash ?? ''));

        if (! $cashier) {
            $attempts = ($attempt->attempts ?? 0) + 1;
            DB::table('pin_attempts')->updateOrInsert(
                [
                    'business_id' => $actor->business_id,
                    'device_id' => $data['device_id'],
                    'outlet_id' => $data['outlet_id'],
                ],
                [
                    'id' => $attempt->id ?? (string) Str::uuid(),
                    'pin_key' => 'cashier',
                    'attempts' => $attempts,
                    'locked_until' => $attempts >= 5 ? now()->addMinutes(15) : null,
                    'created_at' => $attempt->created_at ?? now(),
                    'updated_at' => now(),
                ],
            );

            return ApiResponse::error('INVALID_PIN', 'PIN is invalid.', [], 422);
        }

        DB::table('pin_attempts')->updateOrInsert(
            [
                'business_id' => $actor->business_id,
                'device_id' => $data['device_id'],
                'outlet_id' => $data['outlet_id'],
            ],
            [
                'id' => $attempt->id ?? (string) Str::uuid(),
                'pin_key' => 'cashier',
                'attempts' => 0,
                'locked_until' => null,
                'created_at' => $attempt->created_at ?? now(),
                'updated_at' => now(),
            ],
        );

        Nojpos::audit($actor->business_id, $actor->id, 'auth.pin_switch', 'user', $cashier->id);

        return ApiResponse::success([
            'cashier' => $this->userPayload($cashier),
        ]);
    }

    public function logout(Request $request): JsonResponse
    {
        $request->user()->currentAccessToken()?->delete();

        return ApiResponse::success(['logged_out' => true]);
    }

    public function me(Request $request): JsonResponse
    {
        $user = $request->user();

        return ApiResponse::success([
            'user' => $this->userPayload($user),
            'business' => Business::query()->where('id', $user->business_id)->first()?->only(['id', 'name']),
            'permissions' => [$user->role],
        ]);
    }

    public function outlets(): JsonResponse
    {
        return ApiResponse::success([
            'outlets' => Outlet::query()->get(['id', 'business_id', 'name'])->values()->all(),
        ]);
    }

    private function userPayload(User $user): array
    {
        return [
            'id' => $user->id,
            'business_id' => $user->business_id,
            'name' => $user->name,
            'email' => $user->email,
            'role' => $user->role,
        ];
    }

    private function forbidUnlessSameBusiness(string $table, string $id, string $businessId): void
    {
        if (! DB::table($table)->where('id', $id)->where('business_id', $businessId)->exists()) {
            abort(ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403));
        }
    }
}
