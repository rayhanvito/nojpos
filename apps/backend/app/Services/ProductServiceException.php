<?php

namespace App\Services;

use RuntimeException;

class ProductServiceException extends RuntimeException
{
    public function __construct(
        public readonly string $errorCode,
        string $message,
        public readonly int $status = 422,
        public readonly array $details = [],
    ) {
        parent::__construct($message);
    }

    public static function forbidden(): self
    {
        return new self('FORBIDDEN', 'Resource is outside the current business scope.', 403);
    }

    public static function duplicateBarcode(): self
    {
        return new self('DUPLICATE_BARCODE', 'Barcode already exists for this business.', 422);
    }
}
