import { ActivityLanes } from '@/components/activity-lanes';
import { AdminShell } from '@/components/admin-shell';
import { DetailLinks } from '@/components/detail-links';
import { PreviewActionPanel } from '@/components/preview-action-panel';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { PreviewTable } from '@/components/preview-table';
import { PreviewToolbar } from '@/components/preview-toolbar';
import { catalogDetailLinks, catalogLanes, catalogProductsTable, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function CatalogPage() {
  return (
    <AdminShell title="Katalog Produk">
      <PreviewToolbar action="Tambah produk preview" />
      <ActivityLanes title="Area kerja katalog" lanes={catalogLanes} />
      <PreviewActionPanel group={previewActionGroups.catalog} />
      <PreviewStateBoard
        title="Kondisi tampilan katalog"
        description="Produk, kategori, varian, dan alur publikasi punya status lengkap sebelum aksi katalog dibuka."
        states={previewStateMatrix.catalog}
      />
      <DetailLinks title="Preview detail produk" links={catalogDetailLinks} />
      <section className="resource card">
        <div className="card-header">
          <div>
            <span className="card-kicker">Data produk</span>
            <h2>Produk</h2>
          </div>
          <span className="badge neutral">Preview</span>
        </div>
        <p>Harga, kategori, visibilitas, dan urutan tampil hanya layout preview. Backend tetap sumber data utama.</p>
        <PreviewTable columns={catalogProductsTable.columns} rows={catalogProductsTable.rows} />
      </section>
    </AdminShell>
  );
}
