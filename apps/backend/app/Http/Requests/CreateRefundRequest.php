<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class CreateRefundRequest extends FormRequest
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
            'reason' => ['required', 'string', 'max:500'],
            'refund_method' => ['required', 'in:cash,original_method'],
            'authorization_pin' => ['nullable', 'string', 'max:64'],
            'notes' => ['nullable', 'string', 'max:1000'],
            'lines' => ['required', 'array', 'min:1'],
            'lines.*.transaction_item_id' => ['required', 'uuid'],
            'lines.*.quantity' => ['required', 'integer', 'min:1'],
            'lines.*.restock' => ['sometimes', 'boolean'],
            'lines.*.non_restock_reason' => ['nullable', 'string', 'max:255'],
        ];
    }
}
