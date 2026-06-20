<?php

namespace App\Services;

use RuntimeException;

class SettingsException extends RuntimeException
{
    public function __construct(
        public readonly string $codeValue,
        string $message,
        public readonly array $details = [],
        public readonly int $status = 422,
    ) {
        parent::__construct($message);
    }
}
