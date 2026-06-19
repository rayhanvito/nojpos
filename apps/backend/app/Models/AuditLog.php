<?php

namespace App\Models;

use App\Models\Concerns\BelongsToBusiness;
use App\Models\Concerns\UsesUuid;
use Illuminate\Database\Eloquent\Model;

class AuditLog extends Model
{
    use BelongsToBusiness, UsesUuid;

    public $timestamps = false;

    protected $guarded = [];
}
