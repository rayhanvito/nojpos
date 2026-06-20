<?php

namespace App\Http\Requests\Settings;

use Illuminate\Foundation\Http\FormRequest;

class UpdateSecuritySettingsRequest extends FormRequest
{
    public function authorize(): bool
    {
        return true;
    }

    /**
     * @return array<string, mixed>
     */
    public function rules(): array
    {
        return [
            'sensitive_actions' => ['sometimes', 'required', 'array'],
            'sensitive_actions.void.requires_pin' => ['sometimes', 'boolean'],
            'sensitive_actions.refund.requires_pin' => ['sometimes', 'boolean'],
            'sensitive_actions.discount_override.requires_pin' => ['sometimes', 'boolean'],
            'sensitive_actions.cash_out_over_limit.requires_pin' => ['sometimes', 'boolean'],
            'sensitive_actions.close_shift.requires_pin' => ['sometimes', 'boolean'],
            'sensitive_actions.store_open_close.requires_pin' => ['sometimes', 'boolean'],
            'sensitive_actions.settings_change.requires_pin' => ['sometimes', 'boolean'],
            'pin_policy' => ['sometimes', 'required', 'array'],
            'pin_policy.max_attempts' => ['sometimes', 'required', 'integer', 'min:1', 'max:20'],
            'pin_policy.lockout_minutes' => ['sometimes', 'required', 'integer', 'min:1', 'max:1440'],
            'terminal_policy' => ['sometimes', 'required', 'array'],
            'terminal_policy.idle_lock_timeout_seconds' => ['sometimes', 'required', 'integer', 'min:30', 'max:86400'],
            'terminal_policy.session_timeout_seconds' => ['sometimes', 'required', 'integer', 'min:60', 'max:86400'],
        ];
    }
}
