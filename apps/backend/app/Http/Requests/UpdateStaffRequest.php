<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;
use Illuminate\Validation\Rule;

class UpdateStaffRequest extends FormRequest
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
        $staff = $this->route('staff');
        $staffId = is_object($staff) ? $staff->id : $staff;

        return [
            'name' => ['sometimes', 'required', 'string', 'max:255'],
            'email' => ['sometimes', 'nullable', 'email', 'max:255', Rule::unique('users', 'email')->ignore($staffId)],
            'role' => ['sometimes', 'required', 'in:admin,cashier'],
            'password' => ['sometimes', 'nullable', 'string', 'min:8'],
            'pin' => ['sometimes', 'nullable', 'digits_between:4,8'],
        ];
    }
}
