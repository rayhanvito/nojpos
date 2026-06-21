<?php

namespace App\Services;

use App\Models\User;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;

class AttendanceService
{
    public function __construct(
        private readonly AuditLogService $auditLog,
        private readonly BusinessClock $clock,
    ) {}

    /**
     * @param  array<string, mixed>  $filters
     * @return array<int, array<string, mixed>>
     */
    public function list(User $actor, array $filters): array
    {
        $businessId = (string) $actor->business_id;
        $outletId = $filters['outlet_id'] ?? null;
        $staffId = $filters['staff_id'] ?? null;

        if ($outletId && ! $this->belongsToBusiness('outlets', $businessId, (string) $outletId)) {
            throw new AttendanceOperationException('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        if ($staffId && ! $this->belongsToBusiness('users', $businessId, (string) $staffId)) {
            throw new AttendanceOperationException('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        $query = DB::table('attendance_records')
            ->where('business_id', $businessId)
            ->when($outletId, fn ($query, $id) => $query->where('outlet_id', $id))
            ->when($staffId, fn ($query, $id) => $query->where('staff_id', $id));

        if ($filters['date'] ?? null) {
            $timezone = $outletId
                ? $this->clock->outletTimezone($businessId, (string) $outletId)
                : BusinessClock::DEFAULT_TIMEZONE;
            [$start, $end] = $this->clock->utcDayWindow($timezone, (string) $filters['date']);
            $query->where('clock_in_at', '>=', $start)->where('clock_in_at', '<', $end);
        }

        return $query
            ->orderByDesc('clock_in_at')
            ->get()
            ->map(fn (object $record): array => $this->payload($record))
            ->values()
            ->all();
    }

    /**
     * @param  array<string, mixed>  $data
     * @return array{payload: array<string, mixed>, status: int}
     */
    public function record(User $actor, Request $request, array $data): array
    {
        $businessId = (string) $actor->business_id;
        $outletId = (string) $data['outlet_id'];
        $staffId = (string) $data['staff_id'];

        $staff = DB::table('users')
            ->where('business_id', $businessId)
            ->where('id', $staffId)
            ->first();

        if (! $this->belongsToBusiness('outlets', $businessId, $outletId) || ! $staff) {
            throw new AttendanceOperationException('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        if (! Hash::check((string) $data['pin'], $staff->pin_hash ?? '')) {
            throw new AttendanceOperationException('INVALID_PIN', 'PIN is invalid.', [], 422);
        }

        return $data['action'] === 'clock_in'
            ? $this->clockIn($businessId, (string) $actor->id, $staffId, $outletId, $request)
            : $this->clockOut($businessId, (string) $actor->id, $staffId, $outletId, $request);
    }

    /**
     * @return array{payload: array<string, mixed>, status: int}
     */
    private function clockIn(string $businessId, string $createdBy, string $staffId, string $outletId, Request $request): array
    {
        return DB::transaction(function () use ($businessId, $createdBy, $staffId, $outletId, $request): array {
            $openRecord = DB::table('attendance_records')
                ->where('business_id', $businessId)
                ->where('outlet_id', $outletId)
                ->where('staff_id', $staffId)
                ->whereNull('clock_out_at')
                ->lockForUpdate()
                ->first();

            if ($openRecord) {
                throw new AttendanceOperationException('ATTENDANCE_ALREADY_OPEN', 'Staff is already clocked in.', [], 422);
            }

            $now = now();
            $id = (string) Str::uuid();

            DB::table('attendance_records')->insert([
                'id' => $id,
                'business_id' => $businessId,
                'outlet_id' => $outletId,
                'staff_id' => $staffId,
                'created_by' => $createdBy,
                'clock_in_at' => $now,
                'clock_out_at' => null,
                'created_at' => $now,
                'updated_at' => $now,
            ]);

            $record = DB::table('attendance_records')
                ->where('business_id', $businessId)
                ->where('id', $id)
                ->first();

            $payload = $this->payload($record);
            $this->auditLog->record(
                $businessId,
                $staffId,
                'attendance.clock_in',
                'attendance',
                $id,
                null,
                $payload,
                $this->auditContext($outletId, $request),
            );

            return ['payload' => $payload, 'status' => 201];
        });
    }

    /**
     * @return array{payload: array<string, mixed>, status: int}
     */
    private function clockOut(string $businessId, string $createdBy, string $staffId, string $outletId, Request $request): array
    {
        return DB::transaction(function () use ($businessId, $createdBy, $staffId, $outletId, $request): array {
            $record = DB::table('attendance_records')
                ->where('business_id', $businessId)
                ->where('outlet_id', $outletId)
                ->where('staff_id', $staffId)
                ->whereNull('clock_out_at')
                ->orderByDesc('clock_in_at')
                ->lockForUpdate()
                ->first();

            if (! $record) {
                throw new AttendanceOperationException('ATTENDANCE_NOT_OPEN', 'Staff is not clocked in.', [], 422);
            }

            $before = $this->payload($record);
            $now = now();

            DB::table('attendance_records')
                ->where('business_id', $businessId)
                ->where('id', $record->id)
                ->update([
                    'created_by' => $createdBy,
                    'clock_out_at' => $now,
                    'updated_at' => $now,
                ]);

            $fresh = DB::table('attendance_records')
                ->where('business_id', $businessId)
                ->where('id', $record->id)
                ->first();

            $payload = $this->payload($fresh);
            $this->auditLog->record(
                $businessId,
                $staffId,
                'attendance.clock_out',
                'attendance',
                $record->id,
                $before,
                $payload,
                $this->auditContext($outletId, $request),
            );

            return ['payload' => $payload, 'status' => 200];
        });
    }

    private function belongsToBusiness(string $table, string $businessId, string $id): bool
    {
        return DB::table($table)
            ->where('business_id', $businessId)
            ->where('id', $id)
            ->exists();
    }

    /**
     * @return array<string, mixed>
     */
    private function payload(object $record): array
    {
        return [
            'id' => $record->id,
            'business_id' => $record->business_id,
            'outlet_id' => $record->outlet_id,
            'staff_id' => $record->staff_id,
            'clock_in_at' => $record->clock_in_at,
            'clock_out_at' => $record->clock_out_at,
            'status' => $record->clock_out_at ? 'clocked_out' : 'clocked_in',
        ];
    }

    /**
     * @return array<string, mixed>
     */
    private function auditContext(string $outletId, Request $request): array
    {
        return [
            'outlet_id' => $outletId,
            'request_id' => $request->header('X-Request-Id') ?: $request->header('X-Correlation-Id'),
            'idempotency_key' => $request->attributes->get('idempotency_key') ?: $request->header('Idempotency-Key'),
            'event_version' => 1,
        ];
    }
}
