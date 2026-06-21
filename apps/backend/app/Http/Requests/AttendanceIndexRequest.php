<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class AttendanceIndexRequest extends FormRequest
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
            'outlet_id' => ['nullable', 'uuid'],
            'staff_id' => ['nullable', 'uuid'],
            'date' => ['nullable', 'date'],
        ];
    }
}
