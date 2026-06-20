<?php

namespace App\Http\Requests\Settings;

use Illuminate\Foundation\Http\FormRequest;

class UpdatePaymentMethodSettingsRequest extends FormRequest
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
            'method' => ['sometimes', 'required', 'string', 'max:64'],
            'is_cash' => ['sometimes', 'boolean'],
            'active' => ['sometimes', 'boolean'],
            'outlet_id' => ['sometimes', 'nullable', 'uuid'],
        ];
    }
}
