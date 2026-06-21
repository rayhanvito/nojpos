<?php

namespace App\Http\Requests\Superadmin;

use Illuminate\Foundation\Http\FormRequest;

class UpdatePlanStatusRequest extends FormRequest
{
    public function authorize(): bool
    {
        return $this->user()?->role === 'superadmin';
    }

    public function rules(): array
    {
        return [
            'is_active' => ['required', 'boolean'],
        ];
    }
}
