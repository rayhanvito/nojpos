<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Models\Business;
use App\Models\Outlet;
use App\Models\User;
use App\Services\TerminalSessionException;
use App\Services\TerminalSessionService;
use App\Support\ApiResponse;
use App\Support\Nojpos;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;

class AuthController extends Controller
{
    public function __construct(private readonly TerminalSessionService $terminalSessions) {}

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
        $outlet = DB::table('outlets')
            ->where('business_id', $user->business_id)
            ->first();
        $device = null;
        if ($outlet) {
            $device = DB::table('devices')
                ->where('business_id', $user->business_id)
                ->where('device_uuid', $data['device_uuid'])
                ->first();

            if (! $device) {
                $deviceId = (string) Str::uuid();
                DB::table('devices')->insert([
                    'id' => $deviceId,
                    'business_id' => $user->business_id,
                    'outlet_id' => $outlet->id,
                    'device_uuid' => $data['device_uuid'],
                    'name' => $data['device_uuid'],
                    'created_at' => now(),
                    'updated_at' => now(),
                ]);
                $device = DB::table('devices')->where('id', $deviceId)->first();
            }
        }

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
            'device' => $device ? [
                'id' => $device->id,
                'device_uuid' => $device->device_uuid,
            ] : null,
            'outlets' => Outlet::query()
                ->where('business_id', $user->business_id)
                ->get()
                ->map(fn (Outlet $outlet): array => $this->outletPayload($outlet))
                ->values()
                ->all(),
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

        $this->rememberActingCashier($request, $cashier->id, $data['outlet_id'], $data['device_id']);

        try {
            $terminalSession = $this->terminalSessions->startFromPinSwitch($actor, $cashier, $request, $data['outlet_id'], $data['device_id']);
        } catch (TerminalSessionException $error) {
            return ApiResponse::error($error->errorCode, $error->getMessage(), $error->details, $error->status);
        }

        Nojpos::audit($actor->business_id, $actor->id, 'auth.pin_switch', 'user', $cashier->id);

        return ApiResponse::success([
            'cashier' => $this->userPayload($cashier),
            'terminal_session' => $terminalSession,
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
            'outlets' => Outlet::query()
                ->get()
                ->map(fn (Outlet $outlet): array => $this->outletPayload($outlet))
                ->values()
                ->all(),
            'permissions' => [$user->role],
        ]);
    }

    public function outlets(): JsonResponse
    {
        return ApiResponse::success([
            'outlets' => Outlet::query()
                ->get()
                ->map(fn (Outlet $outlet): array => $this->outletPayload($outlet))
                ->values()
                ->all(),
        ]);
    }

    private function outletPayload(Outlet $outlet): array
    {
        return [
            'id' => $outlet->id,
            'business_id' => $outlet->business_id,
            'name' => $outlet->name,
            'timezone' => $outlet->timezone ?? 'Asia/Jakarta',
            'receipt_config' => [
                'paper_width' => $outlet->receipt_paper_width ?? '58mm',
                'header_name' => $outlet->receipt_header_name,
                'header_address' => $outlet->receipt_header_address,
                'footer_note' => $outlet->receipt_footer_note,
                'show_logo' => (bool) $outlet->receipt_show_logo,
                'show_qris_info' => (bool) $outlet->receipt_show_qris_info,
            ],
            'payment_methods' => DB::table('payment_method_configs')
                ->where('business_id', $outlet->business_id)
                ->where(function ($query) use ($outlet): void {
                    $query->whereNull('outlet_id')->orWhere('outlet_id', $outlet->id);
                })
                ->whereNull('deleted_at')
                ->orderByDesc('is_cash')
                ->orderBy('method')
                ->get(['method', 'is_cash'])
                ->map(fn ($method): array => [
                    'method' => $method->method,
                    'is_cash' => (bool) $method->is_cash,
                ])
                ->values()
                ->all(),
        ];
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

    private function rememberActingCashier(Request $request, string $cashierId, string $outletId, string $deviceId): void
    {
        $token = $request->user()?->currentAccessToken();
        if (! $token || ! method_exists($token, 'forceFill')) {
            return;
        }

        $abilities = collect($token->abilities ?? [])
            ->reject(fn (string $ability): bool => str_starts_with($ability, 'acting_cashier:')
                || str_starts_with($ability, 'acting_outlet:')
                || str_starts_with($ability, 'acting_device:'))
            ->push('acting_cashier:'.$cashierId)
            ->push('acting_outlet:'.$outletId)
            ->push('acting_device:'.$deviceId)
            ->values()
            ->all();

        $token->forceFill(['abilities' => $abilities])->save();
    }

    private function forbidUnlessSameBusiness(string $table, string $id, string $businessId): void
    {
        if (! DB::table($table)->where('id', $id)->where('business_id', $businessId)->exists()) {
            abort(ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403));
        }
    }
}
