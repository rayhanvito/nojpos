<?php

namespace App\Services;

use RuntimeException;

final class DashboardAccessException extends RuntimeException
{
    public function __construct(
        public readonly string $errorCode,
        string $message,
        public readonly array $details = [],
        public readonly int $status = 403,
    ) {
        parent::__construct($message);
    }
}
