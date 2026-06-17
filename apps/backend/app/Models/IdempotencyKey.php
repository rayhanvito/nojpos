<?php

namespace App\Models;

use App\Models\Concerns\UsesUuid;
use Illuminate\Database\Eloquent\Model;

class IdempotencyKey extends Model
{
    use UsesUuid;

    protected $guarded = [];
}
