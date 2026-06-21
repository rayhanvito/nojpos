<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class OpenStoreRequest extends FormRequest
{
    public function authorize(): bool
    {
        return true;
    }

    /**
     * @return array<string, array<int, string>>
     */
    public function rules(): array
    {
        return [
            'authorization_pin' => ['nullable', 'string'],
            'reason' => ['nullable', 'string', 'max:255'],
            'override_out_of_hours' => ['nullable', 'boolean'],
        ];
    }
}
