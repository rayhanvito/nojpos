import { notFound } from 'next/navigation';

import { AdminShell } from '@/components/admin-shell';
import { DetailPreview } from '@/components/detail-preview';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { customerDetailPreviews, getCustomerDetailPreview, previewStateMatrix } from '@/fixtures/preview';

export function generateStaticParams() {
  return customerDetailPreviews.map((detail) => ({ customerId: detail.id }));
}

export default async function CustomerDetailPage({ params }: { params: Promise<{ customerId: string }> }) {
  const { customerId } = await params;
  const detail = getCustomerDetailPreview(customerId);

  if (!detail) {
    notFound();
  }

  return (
    <AdminShell title={`Customer · ${detail.title}`}>
      <DetailPreview detail={detail} />
      <PreviewStateBoard title="Customer detail state coverage" description="Detail customer memakai state customer yang sama sebelum privacy, consent, import, dan export dihubungkan ke backend." states={previewStateMatrix.customers} />
    </AdminShell>
  );
}
