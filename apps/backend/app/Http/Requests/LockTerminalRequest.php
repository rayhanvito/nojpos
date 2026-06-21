<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class LockTerminalRequest extends FormRequest
{
    public function authorize(): bool
    {
        return true;
    }

    public function rules(): array
    {
        return [
            'outlet_id' => ['required', 'uuid'],
            'device_id' => ['required', 'uuid'],
            'cashier_id' => ['nullable', 'uuid'],
            'shift_id' => ['nullable', 'uuid'],
            'reason' => ['required', 'in:manual,idle_timeout,session_timeout'],
        ];
    }
}
