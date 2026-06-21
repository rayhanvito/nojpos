<?php

namespace App\Services;

use RuntimeException;

class VoidOperationException extends RuntimeException
{
    public function __construct(
        public readonly string $errorCode,
        string $message,
        public readonly array $details = [],
        public readonly int $status = 422,
    ) {
        parent::__construct($message);
    }
}
