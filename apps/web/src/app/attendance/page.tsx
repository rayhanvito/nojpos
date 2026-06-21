import { TenantPreviewPage } from '@/components/tenant-preview-page';
import { attendanceDetailLinks, attendanceLanes, attendanceMetrics, attendanceTable, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function AttendancePage() {
  return (
    <TenantPreviewPage
      title="Absensi"
      kicker="Kehadiran tim toko"
      description="Area absensi menyiapkan daftar catatan masuk-keluar, pengecualian jadwal, dan koreksi tanpa membuat logika payroll atau timesheet di browser."
      action="Tambah catatan absensi preview"
      lanes={attendanceLanes}
      metrics={attendanceMetrics}
      actionGroup={previewActionGroups.attendance}
      states={previewStateMatrix.attendance}
      table={attendanceTable}
      detailLinks={attendanceDetailLinks}
    />
  );
}
