import { TenantPreviewPage } from '@/components/tenant-preview-page';
import { outletDetailLinks, outletLanes, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';
import { getOutletsPageModel } from '@/lib/server/outlets';

export default async function OutletsPage() {
  const model = await getOutletsPageModel();

  return (
    <TenantPreviewPage
      title="Outlet"
      kicker="Operasional outlet toko"
      description={`${model.description} Status outlet ditampilkan read-only dan perubahan tetap disabled.`}
      action="Tambah outlet preview"
      lanes={outletLanes}
      metrics={model.metrics}
      actionGroup={previewActionGroups.outlets}
      states={previewStateMatrix.outlets}
      table={model.table}
      detailLinks={outletDetailLinks}
    />
  );
}
