import { notFound } from 'next/navigation';

import { AdminShell } from '@/components/admin-shell';
import { DetailPreview } from '@/components/detail-preview';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { getOutletDetailPreview, outletDetailPreviews, previewStateMatrix } from '@/fixtures/preview';

export function generateStaticParams() {
  return outletDetailPreviews.map((detail) => ({ outletId: detail.id }));
}

export default async function OutletDetailPage({ params }: { params: Promise<{ outletId: string }> }) {
  const { outletId } = await params;
  const detail = getOutletDetailPreview(outletId);

  if (!detail) {
    notFound();
  }

  return (
    <AdminShell title={`Outlet · ${detail.title}`}>
      <DetailPreview detail={detail} />
      <PreviewStateBoard title="Outlet detail state coverage" description="Detail outlet memakai state outlet sebelum open/close outlet dan terminal lock contract tersedia." states={previewStateMatrix.outlets} />
    </AdminShell>
  );
}
