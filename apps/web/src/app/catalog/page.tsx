import { TenantPreviewPage } from '@/components/tenant-preview-page';
import { previewActionGroups, previewStateMatrix } from '@/fixtures/preview';
import { getCatalogPageModel } from '@/lib/server/catalog';

export default async function CatalogPage() {
  const model = await getCatalogPageModel();

  return (
    <TenantPreviewPage
      title="Katalog Produk"
      kicker="Produk dan kategori read-only"
      description={`${model.description} ${model.detail}`}
      action="Filter katalog"
      lanes={model.data.lanes}
      metrics={model.data.metrics}
      actionGroup={{
        ...previewActionGroups.catalog,
        title: 'Aksi katalog tetap dikunci',
        description: 'Tambah, edit, hapus, import, export, dan publish belum aktif dari Web Admin.',
      }}
      states={[
        { tone: model.state === 'real' ? 'empty' : model.state === 'forbidden' ? 'forbidden' : model.state === 'error' ? 'error' : 'unavailable', title: model.title, message: `${model.sourceLabel}: ${model.detail}` },
        { tone: 'loading', title: 'Boundary BFF', message: 'Halaman membaca produk/kategori melalui server-side helper/BFF, bukan langsung ke Laravel.' },
        { tone: 'forbidden', title: 'Aksi katalog', message: 'Tambah, edit, hapus, import, export, dan publish tetap disabled sampai safe write contract disetujui.' },
        ...previewStateMatrix.catalog,
      ]}
      table={model.data.table}
      detailLinks={model.data.detailLinks}
    />
  );
}
