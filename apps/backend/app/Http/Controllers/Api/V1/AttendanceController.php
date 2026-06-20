<?php

namespace App\Http\Controllers\Api\V1;

use App\Http\Controllers\Controller;
use App\Services\BusinessClock;
use App\Support\ApiResponse;
use App\Support\Nojpos;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Str;

class AttendanceController extends Controller
{
    public function __construct(private readonly BusinessClock $clock) {}

    public function index(Request $request): JsonResponse
    {
        $data = $request->validate([
            'outlet_id' => ['nullable', 'uuid'],
            'staff_id' => ['nullable', 'uuid'],
            'date' => ['nullable', 'date'],
        ]);

        $businessId = $request->user()->business_id;
        if (($data['outlet_id'] ?? null) && ! $this->belongsToBusiness('outlets', $businessId, $data['outlet_id'])) {
            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        if (($data['staff_id'] ?? null) && ! $this->belongsToBusiness('users', $businessId, $data['staff_id'])) {
            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        $query = DB::table('attendance_records')
            ->where('business_id', $businessId)
            ->when($data['outlet_id'] ?? null, fn ($query, $outletId) => $query->where('outlet_id', $outletId))
            ->when($data['staff_id'] ?? null, fn ($query, $staffId) => $query->where('staff_id', $staffId));

        if ($data['date'] ?? null) {
            $timezone = isset($data['outlet_id'])
                ? $this->clock->outletTimezone($businessId, $data['outlet_id'])
                : BusinessClock::DEFAULT_TIMEZONE;
            [$start, $end] = $this->clock->utcDayWindow($timezone, $data['date']);
            $query->where('clock_in_at', '>=', $start)->where('clock_in_at', '<', $end);
        }

        return ApiResponse::success([
            'attendance' => $query
                ->orderByDesc('clock_in_at')
                ->get()
                ->map(fn (object $record): array => $this->payload($record))
                ->values()
                ->all(),
        ]);
    }

    public function store(Request $request): JsonResponse
    {
        $data = $request->validate([
            'outlet_id' => ['required', 'uuid'],
            'staff_id' => ['required', 'uuid'],
            'pin' => ['required', 'string'],
            'action' => ['required', 'in:clock_in,clock_out'],
        ]);

        $businessId = $request->user()->business_id;
        if (! $this->belongsToBusiness('outlets', $businessId, $data['outlet_id']) || ! $this->belongsToBusiness('users', $businessId, $data['staff_id'])) {
            return ApiResponse::error('FORBIDDEN', 'Resource is outside the current business scope.', [], 403);
        }

        $staff = DB::table('users')
            ->where('business_id', $businessId)
            ->where('id', $data['staff_id'])
            ->first();

        if (! $staff || ! Hash::check($data['pin'], $staff->pin_hash ?? '')) {
            return ApiResponse::error('INVALID_PIN', 'PIN is invalid.', [], 422);
        }

        return $data['action'] === 'clock_in'
            ? $this->clockIn($businessId, $request->user()->id, $data)
            : $this->clockOut($businessId, $request->user()->id, $data);
    }

    private function clockIn(string $businessId, string $createdBy, array $data): JsonResponse
    {
        $now = now();
        $id = (string) Str::uuid();
        DB::table('attendance_records')->insert([
            'id' => $id,
            'business_id' => $businessId,
            'outlet_id' => $data['outlet_id'],
            'staff_id' => $data['staff_id'],
            'created_by' => $createdBy,
            'clock_in_at' => $now,
            'clock_out_at' => null,
            'created_at' => $now,
            'updated_at' => $now,
        ]);

        Nojpos::audit($businessId, $data['staff_id'], 'attendance.clock_in', 'attendance', $id);

        return ApiResponse::success($this->payload(DB::table('attendance_records')->where('id', $id)->first()), [], 201);
    }

    private function clockOut(string $businessId, string $createdBy, array $data): JsonResponse
    {
        $record = DB::table('attendance_records')
            ->where('business_id', $businessId)
            ->where('outlet_id', $data['outlet_id'])
            ->where('staff_id', $data['staff_id'])
            ->whereNull('clock_out_at')
            ->orderByDesc('clock_in_at')
            ->first();

        if (! $record) {
            return ApiResponse::error('ATTENDANCE_NOT_OPEN', 'Staff is not clocked in.', [], 422);
        }

        DB::table('attendance_records')->where('id', $record->id)->update([
            'created_by' => $createdBy,
            'clock_out_at' => now(),
            'updated_at' => now(),
        ]);

        Nojpos::audit($businessId, $data['staff_id'], 'attendance.clock_out', 'attendance', $record->id);

        return ApiResponse::success($this->payload(DB::table('attendance_records')->where('id', $record->id)->first()));
    }

    private function belongsToBusiness(string $table, string $businessId, string $id): bool
    {
        return DB::table($table)
            ->where('business_id', $businessId)
            ->where('id', $id)
            ->exists();
    }

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
}
