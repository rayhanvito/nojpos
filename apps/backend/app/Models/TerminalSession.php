<?php

namespace App\Models;

use App\Models\Concerns\BelongsToBusiness;
use App\Models\Concerns\UsesUuid;
use Illuminate\Database\Eloquent\Model;

class TerminalSession extends Model
{
    use BelongsToBusiness, UsesUuid;

    protected $guarded = [];

    protected function casts(): array
    {
        return [
            'opened_at' => 'datetime',
            'last_seen_at' => 'datetime',
            'expires_at' => 'datetime',
            'locked_at' => 'datetime',
            'revoked_at' => 'datetime',
            'metadata' => 'array',
        ];
    }
}
