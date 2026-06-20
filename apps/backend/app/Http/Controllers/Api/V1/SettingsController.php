<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Http\Requests\Settings\UpdateBusinessSettingsRequest;
use App\Http\Requests\Settings\UpdateOutletSettingsRequest;
use App\Http\Requests\Settings\UpdatePaymentMethodSettingsRequest;
use App\Http\Requests\Settings\UpdateSecuritySettingsRequest;
use App\Services\SettingsException;
use App\Services\SettingsService;
use App\Support\ApiResponse;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;

class SettingsController extends Controller
{
    public function __construct(private readonly SettingsService $settings) {}

    public function index(Request $request): JsonResponse
    {
        try {
            $aggregate = $this->settings->aggregate($request->user());

            return ApiResponse::success($aggregate['settings'], $aggregate['meta']);
        } catch (SettingsException $exception) {
            return $this->settingsError($exception);
        }
    }

    public function updateBusiness(UpdateBusinessSettingsRequest $request): JsonResponse
    {
        try {
            $aggregate = $this->settings->updateBusiness(
                actor: $request->user(),
                data: $request->validated(),
                context: $this->requestContext($request),
            );

            return ApiResponse::success($aggregate['settings'], $aggregate['meta']);
        } catch (SettingsException $exception) {
            return $this->settingsError($exception);
        }
    }

    public function updateOutlet(UpdateOutletSettingsRequest $request, string $outlet): JsonResponse
    {
        try {
            $aggregate = $this->settings->updateOutlet(
                actor: $request->user(),
                outletId: $outlet,
                data: $request->validated(),
                context: $this->requestContext($request),
            );

            return ApiResponse::success($aggregate['settings'], $aggregate['meta']);
        } catch (SettingsException $exception) {
            return $this->settingsError($exception);
        }
    }

    public function updateSecurity(UpdateSecuritySettingsRequest $request): JsonResponse
    {
        try {
            $aggregate = $this->settings->updateSecurity(
                actor: $request->user(),
                data: $request->validated(),
                context: $this->requestContext($request),
            );

            return ApiResponse::success($aggregate['settings'], $aggregate['meta']);
        } catch (SettingsException $exception) {
            return $this->settingsError($exception);
        }
    }

    public function updatePaymentMethod(UpdatePaymentMethodSettingsRequest $request, string $config): JsonResponse
    {
        try {
            $aggregate = $this->settings->updatePaymentMethod(
                actor: $request->user(),
                configId: $config,
                data: $request->validated(),
                context: $this->requestContext($request),
            );

            return ApiResponse::success($aggregate['settings'], $aggregate['meta']);
        } catch (SettingsException $exception) {
            return $this->settingsError($exception);
        }
    }

    /**
     * @return array<string, mixed>
     */
    private function requestContext(Request $request): array
    {
        return [
            'request_id' => $request->headers->get('X-Request-Id'),
            'idempotency_key' => $request->attributes->get('idempotency_key'),
        ];
    }

    private function settingsError(SettingsException $exception): JsonResponse
    {
        return ApiResponse::error(
            $exception->codeValue,
            $exception->getMessage(),
            $exception->details,
            $exception->status,
        );
    }
}
