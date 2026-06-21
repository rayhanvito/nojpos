<?php

namespace App\Http\Requests;

use App\Support\ApiResponse;
use Illuminate\Contracts\Validation\Validator as ValidatorContract;
use Illuminate\Foundation\Http\FormRequest;
use Illuminate\Http\Exceptions\HttpResponseException;
use Illuminate\Validation\Validator;

class CheckoutTransactionRequest extends FormRequest
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
        $strictQuote = (bool) config('nojpos.checkout.require_quote_for_checkout', false);

        return [
            'outlet_id' => ['required', 'uuid'],
            'device_id' => ['required', 'uuid'],
            'cashier_id' => ['required', 'uuid'],
            'shift_id' => ['required', 'uuid'],
            'customer_id' => ['nullable', 'uuid'],
            'served_by' => ['nullable', 'uuid'],
            'quote_id' => [$strictQuote ? 'required' : 'nullable', 'uuid'],
            'checkout_token' => ['nullable', 'string', 'size:64'],
            'quote_token' => ['nullable', 'string', 'size:64'],
            'status' => ['nullable', 'in:held,unpaid'],
            'promotion_codes' => ['nullable', 'array'],
            'promotion_codes.*' => ['string', 'max:64'],
            'applied_promotion_ids' => ['nullable', 'array'],
            'applied_promotion_ids.*' => ['uuid'],
            'manual_discount' => ['nullable', 'array'],
            'manual_discount.type' => ['required_with:manual_discount', 'in:amount,percent'],
            'manual_discount.value' => ['required_with:manual_discount', 'integer', 'min:0'],
            'manual_discount.reason' => ['nullable', 'string', 'max:255'],
            'cart_discount' => ['nullable', 'integer', 'min:0'],
            'cart_discount_type' => ['nullable', 'in:amount,percent'],
            'rounding' => ['nullable', 'integer'],
            'subtotal' => ['nullable', 'integer', 'min:0'],
            'item_discount_total' => ['nullable', 'integer', 'min:0'],
            'cart_discount_total' => ['nullable', 'integer', 'min:0'],
            'discount_total' => ['nullable', 'integer', 'min:0'],
            'service_charge_total' => ['nullable', 'integer', 'min:0'],
            'tax_total' => ['nullable', 'integer', 'min:0'],
            'rounding_total' => ['nullable', 'integer'],
            'grand_total' => ['nullable', 'integer', 'min:0'],
            'notes' => ['nullable', 'string'],
            'items' => ['required', 'array', 'min:1'],
            'items.*.product_id' => ['required', 'uuid'],
            'items.*.quantity' => ['required', 'integer', 'min:1'],
            // Legacy clients still send unit_price. Checkout ignores it when quote_id is present.
            'items.*.unit_price' => ['required', 'integer', 'min:0'],
            'items.*.discount' => ['nullable', 'integer', 'min:0'],
            'items.*.discount_type' => ['nullable', 'in:amount,percent'],
            'payments' => ['nullable', 'array'],
            'payments.*.method' => ['required', 'string'],
            'payments.*.amount' => ['required', 'integer', 'min:0'],
            'payments.*.reference' => ['nullable', 'string'],
        ];
    }

    public function withValidator(Validator $validator): void
    {
        $validator->after(function (Validator $validator): void {
            if (! (bool) config('nojpos.checkout.require_quote_for_checkout', false)) {
                return;
            }

            if (! $this->filled('quote_id')) {
                $validator->errors()->add('quote_id', 'A checkout quote is required.');
            }
        });
    }

    protected function failedValidation(ValidatorContract $validator): void
    {
        if ((bool) config('nojpos.checkout.require_quote_for_checkout', false) && $validator->errors()->has('quote_id')) {
            throw new HttpResponseException(ApiResponse::error(
                'QUOTE_REQUIRED',
                'Checkout quote is required. Please request a quote before checkout.',
                ['quote_id' => $validator->errors()->get('quote_id')],
                422,
            ));
        }

        parent::failedValidation($validator);
    }
}
