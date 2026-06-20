<?php

namespace App\Services;

use RuntimeException;

class TerminalContextException extends RuntimeException
{
    public function __construct(
        public readonly string $errorCode,
        string $message,
        public readonly int $status = 403,
        public readonly array $details = [],
    ) {
        parent::__construct($message);
    }
}
