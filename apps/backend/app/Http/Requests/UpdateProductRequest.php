<?php

namespace App\Http\Requests;

use Illuminate\Foundation\Http\FormRequest;

class UpdateProductRequest extends FormRequest
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
            'outlet_id' => ['sometimes', 'uuid'],
            'product_category_id' => ['nullable', 'uuid'],
            'name' => ['sometimes', 'string', 'max:255'],
            'barcode' => ['nullable', 'string', 'max:100'],
            'price' => ['sometimes', 'integer', 'min:0'],
            'track_stock' => ['sometimes', 'boolean'],
        ];
    }
}
