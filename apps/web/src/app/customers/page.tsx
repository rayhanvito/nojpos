import { TenantPreviewPage } from '@/components/tenant-preview-page';
import { customerDetailLinks, customerLanes, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';
import { getCustomersPageModel } from '@/lib/server/customers';

export default async function CustomersPage() {
  const model = await getCustomersPageModel();

  return (
    <TenantPreviewPage
      title="Pelanggan"
      kicker="Direktori pelanggan toko"
      description={`${model.description} Data kontak pelanggan tetap masked dan semua aksi edit/export tetap disabled.`}
      action="Tambah pelanggan preview"
      lanes={customerLanes}
      metrics={model.metrics}
      actionGroup={previewActionGroups.customers}
      states={previewStateMatrix.customers}
      table={model.table}
      detailLinks={customerDetailLinks}
    />
  );
}
