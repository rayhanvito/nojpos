import { notFound } from 'next/navigation';

import { AdminShell } from '@/components/admin-shell';
import { DetailPreview } from '@/components/detail-preview';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { catalogDetailPreviews, getCatalogDetailPreview, previewStateMatrix } from '@/fixtures/preview';

export function generateStaticParams() {
  return catalogDetailPreviews.map((detail) => ({ productId: detail.id }));
}

export default async function CatalogDetailPage({ params }: { params: Promise<{ productId: string }> }) {
  const { productId } = await params;
  const detail = getCatalogDetailPreview(productId);

  if (!detail) {
    notFound();
  }

  return (
    <AdminShell title={`Katalog · ${detail.title}`}>
      <DetailPreview detail={detail} />
      <PreviewStateBoard
        title="Product detail state coverage"
        description="Detail produk memakai state catalog yang sama sebelum CRUD, publish, dan outlet visibility dihubungkan ke backend."
        states={previewStateMatrix.catalog}
      />
    </AdminShell>
  );
}
