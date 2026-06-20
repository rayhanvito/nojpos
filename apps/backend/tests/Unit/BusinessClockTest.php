<?php

namespace Tests\Unit;

use App\Services\BusinessClock;
use Carbon\CarbonImmutable;
use Tests\TestCase;

class BusinessClockTest extends TestCase
{
    public function test_jakarta_day_boundary_is_converted_to_a_utc_half_open_window(): void
    {
        $clock = app(BusinessClock::class);

        [$start, $end] = $clock->utcDayWindow('Asia/Jakarta', '2026-01-01');

        $this->assertSame('2025-12-31T17:00:00+00:00', $start->toIso8601String());
        $this->assertSame('2026-01-01T17:00:00+00:00', $end->toIso8601String());
    }

    public function test_document_date_uses_outlet_local_month_and_year_from_a_utc_timestamp(): void
    {
        $clock = app(BusinessClock::class);
        $utc = CarbonImmutable::parse('2025-12-31T17:30:00Z');

        $this->assertSame('20260101', $clock->documentDate('Asia/Jakarta', $utc));
    }
}
