import { notFound } from 'next/navigation';

import { AdminShell } from '@/components/admin-shell';
import { DetailPreview } from '@/components/detail-preview';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { auditDetailPreviews, getAuditDetailPreview, previewStateMatrix } from '@/fixtures/preview';

export function generateStaticParams() {
  return auditDetailPreviews.map((detail) => ({ eventId: detail.id }));
}

export default async function AuditDetailPage({ params }: { params: Promise<{ eventId: string }> }) {
  const { eventId } = await params;
  const detail = getAuditDetailPreview(eventId);

  if (!detail) {
    notFound();
  }

  return (
    <AdminShell title={`Audit · ${detail.title}`}>
      <DetailPreview detail={detail} />
      <PreviewStateBoard
        title="Audit detail state coverage"
        description="Detail audit memakai state audit yang sama dan tetap menahan isi event sampai contract backend disetujui."
        states={previewStateMatrix.audit}
      />
    </AdminShell>
  );
}
