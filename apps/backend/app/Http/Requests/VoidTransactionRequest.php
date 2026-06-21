<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class VoidTransactionRequest extends FormRequest
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
            'transaction_id' => ['required', 'uuid'],
            'shift_id' => ['required', 'uuid'],
            'reason' => ['required', 'string', 'max:500'],
        ];
    }
}
