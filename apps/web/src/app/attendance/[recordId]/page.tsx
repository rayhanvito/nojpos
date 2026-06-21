import { notFound } from 'next/navigation';

import { AdminShell } from '@/components/admin-shell';
import { DetailPreview } from '@/components/detail-preview';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { attendanceDetailPreviews, getAttendanceDetailPreview, previewStateMatrix } from '@/fixtures/preview';

export function generateStaticParams() {
  return attendanceDetailPreviews.map((detail) => ({ recordId: detail.id }));
}

export default async function AttendanceDetailPage({ params }: { params: Promise<{ recordId: string }> }) {
  const { recordId } = await params;
  const detail = getAttendanceDetailPreview(recordId);

  if (!detail) {
    notFound();
  }

  return (
    <AdminShell title={`Attendance · ${detail.title}`}>
      <DetailPreview detail={detail} />
      <PreviewStateBoard title="Attendance detail state coverage" description="Detail attendance tetap read-only sampai correction workflow, role policy, dan export payroll contract tersedia." states={previewStateMatrix.attendance} />
    </AdminShell>
  );
}
