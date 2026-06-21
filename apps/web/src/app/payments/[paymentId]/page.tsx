import { notFound } from 'next/navigation';

import { AdminShell } from '@/components/admin-shell';
import { DetailPreview } from '@/components/detail-preview';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { getPaymentDetailPreview, paymentDetailPreviews, previewStateMatrix } from '@/fixtures/preview';

export function generateStaticParams() {
  return paymentDetailPreviews.map((detail) => ({ paymentId: detail.id }));
}

export default async function PaymentDetailPage({ params }: { params: Promise<{ paymentId: string }> }) {
  const { paymentId } = await params;
  const detail = getPaymentDetailPreview(paymentId);

  if (!detail) {
    notFound();
  }

  return (
    <AdminShell title={`Payment · ${detail.title}`}>
      <DetailPreview detail={detail} />
      <PreviewStateBoard title="Payment detail state coverage" description="Detail payment tetap read-only; refund, void, retry, dan gateway settlement menunggu contract backend." states={previewStateMatrix.payments} />
    </AdminShell>
  );
}
