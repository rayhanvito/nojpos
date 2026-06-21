<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class UnlockTerminalRequest extends FormRequest
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
            'staff_id' => ['required', 'uuid'],
            'pin' => ['required', 'string'],
            'mode' => ['nullable', 'in:resume_current,handover'],
        ];
    }
}
