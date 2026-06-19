<?php

namespace Database\Seeders;

use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;

class SuperadminBootstrapSeeder extends Seeder
{
    public function run(): void
    {
        $now = now();
        $email = env('NOJPOS_SUPERADMIN_EMAIL', 'superadmin@nojpos.test');
        $password = env('NOJPOS_SUPERADMIN_PASSWORD', 'password');

        DB::table('users')->updateOrInsert(
            ['email' => $email],
            [
                'id' => DB::table('users')->where('email', $email)->value('id') ?? (string) Str::uuid(),
                'business_id' => null,
                'name' => 'NojPOS Superadmin',
                'password' => Hash::make($password),
                'role' => 'superadmin',
                'pin_hash' => null,
                'created_at' => DB::table('users')->where('email', $email)->value('created_at') ?? $now,
                'updated_at' => $now,
                'deleted_at' => null,
            ],
        );

        DB::table('plans')->updateOrInsert(
            ['active_code' => 'starter'],
            [
                'id' => DB::table('plans')->where('active_code', 'starter')->value('id') ?? (string) Str::uuid(),
                'name' => 'Starter',
                'code' => 'starter',
                'price' => 0,
                'billing_period' => 'monthly',
                'max_outlets' => 1,
                'max_devices' => 2,
                'max_users' => 5,
                'max_products' => 100,
                'is_active' => true,
                'created_at' => DB::table('plans')->where('active_code', 'starter')->value('created_at') ?? $now,
                'updated_at' => $now,
                'deleted_at' => null,
            ],
        );

        $this->command?->info('Superadmin bootstrap seeded.');
        $this->command?->line('Superadmin email: '.$email);
    }
}
