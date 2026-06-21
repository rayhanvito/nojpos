<?php

namespace App\Services;

use RuntimeException;
use Throwable;

class TerminalSessionException extends RuntimeException
{
    public function __construct(
        public readonly string $errorCode,
        string $message,
        public readonly array $details = [],
        public readonly int $status = 403,
        ?Throwable $previous = null,
    ) {
        parent::__construct($message, 0, $previous);
    }
}
