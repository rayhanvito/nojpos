<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class StoreStaffRequest extends FormRequest
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
            'name' => ['required', 'string', 'max:255'],
            'email' => ['nullable', 'email', 'max:255', 'unique:users,email'],
            'role' => ['required', 'in:admin,cashier'],
            'password' => ['nullable', 'string', 'min:8'],
            'pin' => ['required', 'digits_between:4,8'],
        ];
    }
}
