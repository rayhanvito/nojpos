import { AdminShell } from '@/components/admin-shell';
import { DetailLinks } from '@/components/detail-links';
import { PreviewActionPanel } from '@/components/preview-action-panel';
import { PreviewMetric } from '@/components/preview-metric';
import { PreviewStateBoard } from '@/components/preview-state-board';
import { PreviewToolbar } from '@/components/preview-toolbar';
import { ReadonlyResource } from '@/components/readonly-resource';
import { inventoryDetailLinks, inventoryMetrics, inventoryResources, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function InventoryPage() {
  return (
    <AdminShell title="Inventaris">
      <PreviewToolbar action="Filter preview" />
      <div className="kpi-grid">
        {inventoryMetrics.map((metric) => <PreviewMetric key={metric.label} {...metric} />)}
      </div>
      <PreviewActionPanel group={previewActionGroups.inventory} />
      <PreviewStateBoard
        title="Kondisi tampilan inventaris"
        description="Stok, pergerakan barang, transfer, dan status stok rendah disiapkan tanpa kalkulasi stok di browser."
        states={previewStateMatrix.inventory}
      />
      <DetailLinks title="Preview detail inventaris" links={inventoryDetailLinks} />
      <div className="panel-grid">
        {inventoryResources.map((resource) => <ReadonlyResource key={resource.path} {...resource} />)}
      </div>
    </AdminShell>
  );
}
