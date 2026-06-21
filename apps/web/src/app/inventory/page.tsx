import { TenantPreviewPage } from '@/components/tenant-preview-page';
import { previewActionGroups, previewStateMatrix } from '@/fixtures/preview';
import { getInventoryPageModel } from '@/lib/server/inventory';

export default async function InventoryPage() {
  const model = await getInventoryPageModel();

  return (
    <TenantPreviewPage
      title="Inventaris"
      kicker="Stok toko read-only"
      description={`${model.description} ${model.detail}`}
      action="Filter inventory"
      lanes={model.data.lanes}
      metrics={model.data.metrics}
      actionGroup={{
        ...previewActionGroups.inventory,
        title: 'Aksi inventaris tetap dikunci',
        description: 'Stock adjustment, purchase, count, waste, transfer, dan export belum aktif dari Web Admin.',
      }}
      states={[
        { tone: model.state === 'real' ? 'empty' : model.state === 'forbidden' ? 'forbidden' : model.state === 'error' ? 'error' : 'unavailable', title: model.title, message: `${model.sourceLabel}: ${model.detail}` },
        { tone: 'loading', title: 'Boundary BFF', message: 'Halaman membaca data melalui server-side helper/BFF, bukan langsung ke Laravel.' },
        { tone: 'forbidden', title: 'Aksi stok sensitif', message: 'Adjustment, transfer, purchase, count, waste, dan export tetap disabled.' },
        ...previewStateMatrix.inventory,
      ]}
      table={model.data.table}
      detailLinks={model.data.detailLinks}
    />
  );
}
