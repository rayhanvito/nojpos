<?php

namespace App\Services;

use Carbon\CarbonImmutable;
use DateTimeZone;
use Illuminate\Support\Facades\DB;

class BusinessClock
{
    public const DEFAULT_TIMEZONE = 'Asia/Jakarta';

    /**
     * @return array{0: CarbonImmutable, 1: CarbonImmutable}
     */
    public function utcDayWindow(string $timezone, CarbonImmutable|string $localDate): array
    {
        $localStart = CarbonImmutable::parse($localDate, $this->normalizeTimezone($timezone))->startOfDay();

        return [$localStart->utc(), $localStart->addDay()->utc()];
    }

    public function documentDate(string $timezone, CarbonImmutable|string $utcTimestamp): string
    {
        return CarbonImmutable::parse($utcTimestamp, 'UTC')
            ->setTimezone($this->normalizeTimezone($timezone))
            ->format('Ymd');
    }

    public function documentTimestamp(string $timezone, CarbonImmutable|string $utcTimestamp): string
    {
        return CarbonImmutable::parse($utcTimestamp, 'UTC')
            ->setTimezone($this->normalizeTimezone($timezone))
            ->format('YmdHis');
    }

    public function outletTimezone(string $businessId, string $outletId): string
    {
        $timezone = DB::table('outlets')
            ->where('business_id', $businessId)
            ->where('id', $outletId)
            ->value('timezone');

        return $this->normalizeTimezone(is_string($timezone) ? $timezone : self::DEFAULT_TIMEZONE);
    }

    public function localNow(string $timezone): CarbonImmutable
    {
        return CarbonImmutable::now('UTC')->setTimezone($this->normalizeTimezone($timezone));
    }

    private function normalizeTimezone(string $timezone): string
    {
        try {
            new DateTimeZone($timezone);

            return $timezone;
        } catch (\Exception) {
            return self::DEFAULT_TIMEZONE;
        }
    }
}
