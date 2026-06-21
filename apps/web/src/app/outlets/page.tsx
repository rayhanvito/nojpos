import { TenantPreviewPage } from '@/components/tenant-preview-page';
import { outletDetailLinks, outletLanes, outletMetrics, outletsTable, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function OutletsPage() {
  return (
    <TenantPreviewPage
      title="Outlet"
      kicker="Operasional outlet toko"
      description="Area outlet memisahkan daftar outlet, status operasional, dan kontrol terminal. Buka/tutup outlet serta kunci terminal tetap dikunci."
      action="Tambah outlet preview"
      lanes={outletLanes}
      metrics={outletMetrics}
      actionGroup={previewActionGroups.outlets}
      states={previewStateMatrix.outlets}
      table={outletsTable}
      detailLinks={outletDetailLinks}
    />
  );
}
