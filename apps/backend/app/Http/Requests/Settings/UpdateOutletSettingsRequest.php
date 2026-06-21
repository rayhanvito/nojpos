<?php

namespace App\Http\Requests\Settings;

use Illuminate\Foundation\Http\FormRequest;

class UpdateOutletSettingsRequest extends FormRequest
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
            'name' => ['sometimes', 'required', 'string', 'max:255'],
            'timezone' => ['sometimes', 'required', 'timezone'],
            'service_charge_rate' => ['sometimes', 'required', 'integer', 'min:0', 'max:100'],
            'tax_rate' => ['sometimes', 'required', 'integer', 'min:0', 'max:100'],
            'receipt_config' => ['sometimes', 'required', 'array'],
            'receipt_config.paper_width' => ['sometimes', 'required', 'in:58mm,80mm'],
            'receipt_config.header_name' => ['sometimes', 'nullable', 'string', 'max:255'],
            'receipt_config.header_address' => ['sometimes', 'nullable', 'string', 'max:500'],
            'receipt_config.footer_note' => ['sometimes', 'nullable', 'string', 'max:500'],
            'receipt_config.show_logo' => ['sometimes', 'boolean'],
            'receipt_config.show_qris_info' => ['sometimes', 'boolean'],
            'operational_config' => ['sometimes', 'required', 'array'],
            'operational_config.store_open_close_enabled' => ['sometimes', 'boolean'],
        ];
    }
}
