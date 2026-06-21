import { notFound } from 'next/navigation';

import { AdminShell } from '@/components/admin-shell';
import { DetailPreview } from '@/components/detail-preview';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { getPromotionDetailPreview, previewStateMatrix, promotionDetailPreviews } from '@/fixtures/preview';

export function generateStaticParams() {
  return promotionDetailPreviews.map((detail) => ({ promotionId: detail.id }));
}

export default async function PromotionDetailPage({ params }: { params: Promise<{ promotionId: string }> }) {
  const { promotionId } = await params;
  const detail = getPromotionDetailPreview(promotionId);

  if (!detail) {
    notFound();
  }

  return (
    <AdminShell title={`Promosi · ${detail.title}`}>
      <DetailPreview detail={detail} />
      <PreviewStateBoard
        title="Promotion detail state coverage"
        description="Detail promo memakai state promotions yang sama tanpa menghitung diskon, voucher, eligibility, atau total di browser."
        states={previewStateMatrix.promotions}
      />
    </AdminShell>
  );
}
