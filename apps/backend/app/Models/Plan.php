<?php

namespace App\Models;

use App\Models\Concerns\UsesUuid;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

class Plan extends Model
{
    use SoftDeletes, UsesUuid;

    protected $guarded = [];

    protected $casts = [
        'price' => 'integer',
        'max_outlets' => 'integer',
        'max_devices' => 'integer',
        'max_users' => 'integer',
        'max_products' => 'integer',
        'is_active' => 'boolean',
    ];

    protected static function booted(): void
    {
        static::saving(function (Plan $plan): void {
            $plan->active_code = $plan->deleted_at ? null : $plan->code;
        });
    }
}
