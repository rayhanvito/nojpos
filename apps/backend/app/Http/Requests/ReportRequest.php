<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;
use Illuminate\Validation\Rule;
use Illuminate\Validation\Validator;

class ReportRequest extends FormRequest
{
    private const MAX_CUSTOM_RANGE_DAYS = 31;

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
            'date_from' => ['nullable', 'date_format:Y-m-d', 'required_with:date_to'],
            'date_to' => ['nullable', 'date_format:Y-m-d', 'required_with:date_from', 'after_or_equal:date_from'],
            'range' => ['nullable', Rule::in(['day', 'week', 'month'])],
            'shift_id' => ['nullable', 'uuid'],
            'outlet_id' => ['nullable', 'uuid'],
            'page' => ['nullable', 'integer', 'min:1'],
            'per_page' => ['nullable', 'integer', 'min:1', 'max:100'],
            'export' => ['nullable', Rule::in(['csv', 'xlsx'])],
        ];
    }

    public function withValidator(Validator $validator): void
    {
        $validator->after(function (Validator $validator): void {
            if ($this->filled('date') && ($this->filled('date_from') || $this->filled('date_to'))) {
                $validator->errors()->add('date', 'Use either date or date_from/date_to, not both.');
            }

            $dateFrom = $this->input('date_from');
            $dateTo = $this->input('date_to');
            if (! is_string($dateFrom) || ! is_string($dateTo)) {
                return;
            }

            $start = strtotime($dateFrom.' 00:00:00 UTC');
            $end = strtotime($dateTo.' 00:00:00 UTC');
            if ($start === false || $end === false) {
                return;
            }

            $days = (int) floor(($end - $start) / 86400) + 1;
            if ($days > self::MAX_CUSTOM_RANGE_DAYS) {
                $validator->errors()->add('date_to', 'Report custom range cannot exceed '.self::MAX_CUSTOM_RANGE_DAYS.' days.');
            }
        });
    }
}
