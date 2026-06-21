import { notFound } from 'next/navigation';

import { AdminShell } from '@/components/admin-shell';
import { DetailPreview } from '@/components/detail-preview';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { getTransactionDetailPreview, previewStateMatrix, transactionDetailPreviews } from '@/fixtures/preview';

export function generateStaticParams() {
  return transactionDetailPreviews.map((detail) => ({ transactionId: detail.id }));
}

export default async function TransactionDetailPage({ params }: { params: Promise<{ transactionId: string }> }) {
  const { transactionId } = await params;
  const detail = getTransactionDetailPreview(transactionId);

  if (!detail) {
    notFound();
  }

  return (
    <AdminShell title={`Transaction · ${detail.title}`}>
      <DetailPreview detail={detail} />
      <PreviewStateBoard title="Transaction detail state coverage" description="Detail transaksi tetap read-only; void, refund, total, tax, dan tender status menunggu backend contract." states={previewStateMatrix.transactions} />
    </AdminShell>
  );
}
