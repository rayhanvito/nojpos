<?php

namespace App\Services;

class TerminalContext
{
    public function __construct(
        public readonly string $businessId,
        public readonly string $outletId,
        public readonly string $deviceId,
        public readonly string $cashierId,
        public readonly ?string $shiftId = null,
    ) {}
}
