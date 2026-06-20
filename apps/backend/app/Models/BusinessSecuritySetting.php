<?php

namespace App\Models;

use App\Models\Concerns\BelongsToBusiness;
use App\Models\Concerns\UsesUuid;
use Illuminate\Database\Eloquent\Model;

class BusinessSecuritySetting extends Model
{
    use BelongsToBusiness, UsesUuid;

    protected $guarded = [];

    protected function casts(): array
    {
        return [
            'pin_required_void' => 'boolean',
            'pin_required_refund' => 'boolean',
            'pin_required_discount_override' => 'boolean',
            'pin_required_cash_out_over_limit' => 'boolean',
            'pin_required_close_shift' => 'boolean',
            'pin_required_store_open_close' => 'boolean',
            'pin_required_settings_change' => 'boolean',
            'pin_lockout_max_attempts' => 'integer',
            'pin_lockout_decay_minutes' => 'integer',
            'idle_lock_timeout_seconds' => 'integer',
            'terminal_session_timeout_seconds' => 'integer',
        ];
    }
}
