<?php

namespace App\Models;

use App\Models\Concerns\BelongsToBusiness;
use App\Models\Concerns\UsesUuid;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

class Subscription extends Model
{
    use BelongsToBusiness, SoftDeletes, UsesUuid;

    protected $guarded = [];

    protected $casts = [
        'trial_ends_at' => 'datetime',
        'current_period_ends_at' => 'datetime',
        'max_outlets_override' => 'integer',
        'max_devices_override' => 'integer',
        'max_users_override' => 'integer',
        'max_products_override' => 'integer',
    ];

    protected static function booted(): void
    {
        static::saving(function (Subscription $subscription): void {
            $subscription->active_business_id = $subscription->deleted_at ? null : $subscription->business_id;
        });
    }
}
