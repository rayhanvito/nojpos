import { notFound } from 'next/navigation';

import { AdminShell } from '@/components/admin-shell';
import { DetailPreview } from '@/components/detail-preview';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { getInventoryDetailPreview, inventoryDetailPreviews, previewStateMatrix } from '@/fixtures/preview';

export function generateStaticParams() {
  return inventoryDetailPreviews.map((detail) => ({ resourceId: detail.id }));
}

export default async function InventoryDetailPage({ params }: { params: Promise<{ resourceId: string }> }) {
  const { resourceId } = await params;
  const detail = getInventoryDetailPreview(resourceId);

  if (!detail) {
    notFound();
  }

  return (
    <AdminShell title={`Inventaris · ${detail.title}`}>
      <DetailPreview detail={detail} />
      <PreviewStateBoard
        title="Inventory detail state coverage"
        description="Detail inventaris memakai state inventory yang sama tanpa mutation, transfer, atau stock calculation client-side."
        states={previewStateMatrix.inventory}
      />
    </AdminShell>
  );
}
