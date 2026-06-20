<?php

namespace App\Services;

use RuntimeException;

class ReportAccessException extends RuntimeException
{
    /**
     * @param  array<string, mixed>  $details
     */
    public function __construct(
        public readonly string $errorCode,
        string $message,
        public readonly array $details = [],
        public readonly int $status = 403,
    ) {
        parent::__construct($message);
    }
}
