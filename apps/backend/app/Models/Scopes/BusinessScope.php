<?php

namespace App\Models\Scopes;

use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Scope;
use Illuminate\Support\Facades\Auth;

class BusinessScope implements Scope
{
    public function apply(Builder $builder, Model $model): void
    {
        $user = Auth::hasUser() ? Auth::user() : null;

        if (! $user) {
            return;
        }

        if ($user->role === 'superadmin' || ! $user->business_id) {
            $builder->whereRaw('1 = 0');

            return;
        }

        $builder->where($model->getTable().'.business_id', $user->business_id);
    }
}
