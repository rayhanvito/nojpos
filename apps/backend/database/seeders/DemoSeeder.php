<?php

namespace Database\Seeders;

use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;

class DemoSeeder extends Seeder
{
    private const BUSINESS_ID = '11111111-1111-4111-8111-111111111111';

    private const OUTLET_ID = '22222222-2222-4222-8222-222222222222';

    private const DEVICE_ID = '33333333-3333-4333-8333-333333333333';

    private const OWNER_ID = '44444444-4444-4444-8444-444444444444';

    private const CASHIER_ID = '55555555-5555-4555-8555-555555555555';

    private const CASH_PAYMENT_METHOD_ID = '66666666-6666-4666-8666-666666666666';

    private const QRIS_PAYMENT_METHOD_ID = '66666666-6666-4666-8666-666666666667';

    private const TRANSFER_PAYMENT_METHOD_ID = '66666666-6666-4666-8666-666666666668';

    private const EWALLET_PAYMENT_METHOD_ID = '66666666-6666-4666-8666-666666666669';

    private const OWNER_EMAIL = 'owner@demo.nojpos.test';

    private const CASHIER_EMAIL = 'cashier@demo.nojpos.test';

    private const PASSWORD = 'password';

    private const CASHIER_PIN = '1234';

    public function run(): void
    {
        $now = now();

        DB::table('businesses')->updateOrInsert(
            ['id' => self::BUSINESS_ID],
            [
                'name' => 'NojPOS Demo',
                'created_at' => $now,
                'updated_at' => $now,
                'deleted_at' => null,
            ],
        );

        DB::table('outlets')->updateOrInsert(
            ['id' => self::OUTLET_ID],
            [
                'business_id' => self::BUSINESS_ID,
                'name' => 'Outlet Demo',
                'service_charge_rate' => 5,
                'tax_rate' => 11,
                'receipt_paper_width' => '58mm',
                'receipt_header_name' => 'NojPOS Demo',
                'receipt_header_address' => 'Jl. Demo No. 1',
                'receipt_footer_note' => 'Terima kasih sudah berbelanja',
                'receipt_show_logo' => false,
                'receipt_show_qris_info' => true,
                'created_at' => $now,
                'updated_at' => $now,
                'deleted_at' => null,
            ],
        );

        DB::table('devices')->updateOrInsert(
            ['id' => self::DEVICE_ID],
            [
                'business_id' => self::BUSINESS_ID,
                'outlet_id' => self::OUTLET_ID,
                'device_uuid' => 'demo-tablet-001',
                'name' => 'Demo Android Emulator',
                'created_at' => $now,
                'updated_at' => $now,
                'deleted_at' => null,
            ],
        );

        DB::table('users')->updateOrInsert(
            ['email' => self::OWNER_EMAIL],
            [
                'id' => self::OWNER_ID,
                'business_id' => self::BUSINESS_ID,
                'name' => 'Owner Demo',
                'password' => Hash::make(self::PASSWORD),
                'role' => 'owner',
                'pin_hash' => null,
                'created_at' => $now,
                'updated_at' => $now,
                'deleted_at' => null,
            ],
        );

        DB::table('users')->updateOrInsert(
            ['email' => self::CASHIER_EMAIL],
            [
                'id' => self::CASHIER_ID,
                'business_id' => self::BUSINESS_ID,
                'name' => 'Cashier Demo',
                'password' => Hash::make(self::PASSWORD),
                'role' => 'cashier',
                'pin_hash' => Hash::make(self::CASHIER_PIN),
                'created_at' => $now,
                'updated_at' => $now,
                'deleted_at' => null,
            ],
        );

        $paymentMethods = [
            [self::CASH_PAYMENT_METHOD_ID, 'Tunai', true],
            [self::QRIS_PAYMENT_METHOD_ID, 'QRIS Statis', false],
            [self::TRANSFER_PAYMENT_METHOD_ID, 'Transfer', false],
            [self::EWALLET_PAYMENT_METHOD_ID, 'E-wallet', false],
        ];

        foreach ($paymentMethods as [$id, $method, $isCash]) {
            DB::table('payment_method_configs')->updateOrInsert(
                ['id' => $id],
                [
                    'business_id' => self::BUSINESS_ID,
                    'outlet_id' => self::OUTLET_ID,
                    'method' => $method,
                    'is_cash' => $isCash,
                    'created_at' => $now,
                    'updated_at' => $now,
                    'deleted_at' => null,
                ],
            );
        }

        $categoryIds = [
            'Minuman' => '77777777-7777-4777-8777-777777777771',
            'Makanan' => '77777777-7777-4777-8777-777777777772',
        ];

        foreach ($categoryIds as $name => $id) {
            DB::table('product_categories')->updateOrInsert(
                ['id' => $id],
                [
                    'business_id' => self::BUSINESS_ID,
                    'name' => $name,
                    'created_at' => $now,
                    'updated_at' => $now,
                    'deleted_at' => null,
                ],
            );
        }

        $products = [
            ['88888888-8888-4888-8888-888888888881', 'Makanan', 'Nasi Goreng Demo', 'DEMO-NASI', 25000],
            ['88888888-8888-4888-8888-888888888882', 'Makanan', 'Mie Ayam Demo', 'DEMO-MIE', 22000],
            ['88888888-8888-4888-8888-888888888883', 'Makanan', 'Ayam Geprek Demo', 'DEMO-AYAM', 28000],
            ['88888888-8888-4888-8888-888888888884', 'Makanan', 'Roti Bakar Demo', 'DEMO-ROTI', 18000],
            ['88888888-8888-4888-8888-888888888885', 'Minuman', 'Es Teh Demo', 'DEMO-ESTEH', 8000],
            ['88888888-8888-4888-8888-888888888886', 'Minuman', 'Kopi Susu Demo', 'DEMO-KOPI', 18000],
            ['88888888-8888-4888-8888-888888888887', 'Minuman', 'Jus Alpukat Demo', 'DEMO-JUS', 20000],
            ['88888888-8888-4888-8888-888888888888', 'Minuman', 'Air Mineral Demo', 'DEMO-AIR', 6000],
        ];

        foreach ($products as [$id, $category, $name, $barcode, $price]) {
            DB::table('products')->updateOrInsert(
                ['id' => $id],
                [
                    'business_id' => self::BUSINESS_ID,
                    'outlet_id' => self::OUTLET_ID,
                    'product_category_id' => $categoryIds[$category],
                    'name' => $name,
                    'barcode' => $barcode,
                    'price' => $price,
                    'track_stock' => true,
                    'created_at' => $now,
                    'updated_at' => $now,
                    'deleted_at' => null,
                ],
            );
        }

        $this->command?->info('Demo tenant seeded.');
        $this->command?->line('Owner email: '.self::OWNER_EMAIL);
        $this->command?->line('Owner password: '.self::PASSWORD);
        $this->command?->line('Cashier email: '.self::CASHIER_EMAIL);
        $this->command?->line('Cashier PIN: '.self::CASHIER_PIN);
        $this->command?->line('Device UUID: demo-tablet-001');
    }
}
