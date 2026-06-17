<?php

namespace App\Models\Concerns;

use App\Models\Scopes\BusinessScope;

trait BelongsToBusiness
{
    protected static function bootBelongsToBusiness(): void
    {
        static::addGlobalScope(new BusinessScope);
    }
}
