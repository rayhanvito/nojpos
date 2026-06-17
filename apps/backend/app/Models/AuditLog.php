<?php

namespace App\Models;

use App\Models\Concerns\UsesUuid;
use Illuminate\Database\Eloquent\Model;

class AuditLog extends Model
{
    use UsesUuid;

    public $timestamps = false;

    protected $guarded = [];
}
