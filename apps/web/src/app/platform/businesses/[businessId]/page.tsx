import { notFound } from 'next/navigation';

import { DetailPreview } from '@/components/detail-preview';
import { PlatformShell } from '@/components/platform-shell';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { getPlatformBusinessDetailPreview, platformBusinessDetailPreviews, previewStateMatrix } from '@/fixtures/preview';

export function generateStaticParams() {
  return platformBusinessDetailPreviews.map((detail) => ({ businessId: detail.id }));
}

export default async function PlatformBusinessDetailPage({ params }: { params: Promise<{ businessId: string }> }) {
  const { businessId } = await params;
  const detail = getPlatformBusinessDetailPreview(businessId);

  if (!detail) {
    notFound();
  }

  return (
    <PlatformShell title={`Business · ${detail.title}`}>
      <DetailPreview detail={detail} />
      <PreviewStateBoard title="Platform business detail state coverage" description="Detail business memakai platform state tanpa tenant outlet context. Suspend tenant dan change plan tetap disabled." states={previewStateMatrix.platform} />
    </PlatformShell>
  );
}
