<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class TerminalLockStateRequest extends FormRequest
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
            'staff_id' => ['nullable', 'uuid'],
        ];
    }
}
