<?php

namespace App\Services;

use RuntimeException;

class StaffServiceException extends RuntimeException
{
    public function __construct(
        public readonly string $errorCode,
        string $message,
        public readonly int $status = 422,
        public readonly array $details = [],
    ) {
        parent::__construct($message);
    }

    public static function forbidden(string $message = 'Resource is outside the current business scope.'): self
    {
        return new self('FORBIDDEN', $message, 403);
    }
}
