<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;
use Illuminate\Validation\Rule;

class DashboardSummaryRequest extends FormRequest
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
            'date' => ['nullable', 'date_format:Y-m-d'],
            'outlet_id' => ['nullable', 'uuid'],
            'range' => ['nullable', Rule::in(['last_7_days'])],
        ];
    }

    /**
     * @return array<string, string>
     */
    public function messages(): array
    {
        return [
            'date.date_format' => 'The date field must match the format Y-m-d.',
            'outlet_id.uuid' => 'The outlet id field must be a valid UUID.',
            'range.in' => 'Range dashboard hanya mendukung last_7_days untuk kontrak awal.',
        ];
    }
}
