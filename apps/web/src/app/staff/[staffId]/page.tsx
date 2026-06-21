import { notFound } from 'next/navigation';

import { AdminShell } from '@/components/admin-shell';
import { DetailPreview } from '@/components/detail-preview';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { getStaffDetailPreview, previewStateMatrix, staffDetailPreviews } from '@/fixtures/preview';

export function generateStaticParams() {
  return staffDetailPreviews.map((detail) => ({ staffId: detail.id }));
}

export default async function StaffDetailPage({ params }: { params: Promise<{ staffId: string }> }) {
  const { staffId } = await params;
  const detail = getStaffDetailPreview(staffId);

  if (!detail) {
    notFound();
  }

  return (
    <AdminShell title={`Staf · ${detail.title}`}>
      <DetailPreview detail={detail} />
      <PreviewStateBoard
        title="Staff detail state coverage"
        description="Detail staf memakai state staff yang sama tanpa auth browser, invite aktif, atau role update client-side."
        states={previewStateMatrix.staff}
      />
    </AdminShell>
  );
}
