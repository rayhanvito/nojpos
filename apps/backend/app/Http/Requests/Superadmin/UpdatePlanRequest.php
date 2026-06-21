<?php

namespace App\Http\Requests\Superadmin;

use Illuminate\Foundation\Http\FormRequest;

class UpdatePlanRequest extends FormRequest
{
    public function authorize(): bool
    {
        return $this->user()?->role === 'superadmin';
    }

    public function rules(): array
    {
        return [
            'name' => ['sometimes', 'string', 'max:255'],
            'code' => ['sometimes', 'string', 'max:100'],
            'price' => ['sometimes', 'integer', 'min:0'],
            'billing_period' => ['nullable', 'string', 'max:50'],
            'max_outlets' => ['nullable', 'integer', 'min:0'],
            'max_devices' => ['nullable', 'integer', 'min:0'],
            'max_users' => ['nullable', 'integer', 'min:0'],
            'max_products' => ['nullable', 'integer', 'min:0'],
            'is_active' => ['sometimes', 'boolean'],
        ];
    }
}
