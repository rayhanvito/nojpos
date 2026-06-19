<?php

namespace App\Models;

use App\Models\Concerns\BelongsToBusiness;
use App\Models\Concerns\UsesUuid;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

class InventoryPurchaseItem extends Model
{
    use BelongsToBusiness, SoftDeletes, UsesUuid;

    protected $guarded = [];
}
