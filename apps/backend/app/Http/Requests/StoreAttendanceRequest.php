<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class StoreAttendanceRequest extends FormRequest
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
            'outlet_id' => ['required', 'uuid'],
            'staff_id' => ['required', 'uuid'],
            'pin' => ['required', 'string', 'max:64'],
            'action' => ['required', 'in:clock_in,clock_out'],
        ];
    }
}
