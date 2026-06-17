<?php

namespace App\Models\Concerns;

use Illuminate\Support\Str;

trait UsesUuid
{
    public function initializeUsesUuid(): void
    {
        $this->keyType = 'string';
        $this->incrementing = false;
    }

    protected static function bootUsesUuid(): void
    {
        static::creating(function ($model): void {
            if (! $model->getKey()) {
                $model->{$model->getKeyName()} = (string) Str::uuid();
            }
        });
    }
}
