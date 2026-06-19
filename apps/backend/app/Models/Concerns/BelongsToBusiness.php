<?php

namespace App\Models\Concerns;

use App\Models\Scopes\BusinessScope;
use Illuminate\Support\Facades\Auth;

trait BelongsToBusiness
{
    protected static function bootBelongsToBusiness(): void
    {
        static::addGlobalScope(new BusinessScope);

        static::creating(function ($model): void {
            $user = Auth::hasUser() ? Auth::user() : null;
            if (! $user || $user->role === 'superadmin' || ! $user->business_id) {
                return;
            }

            $model->business_id = $user->business_id;
        });
    }
}
