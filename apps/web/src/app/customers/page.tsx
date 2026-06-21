import { TenantPreviewPage } from '@/components/tenant-preview-page';
import { customerDetailLinks, customerLanes, customerMetrics, customersTable, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function CustomersPage() {
  return (
    <TenantPreviewPage
      title="Pelanggan"
      kicker="Direktori pelanggan toko"
      description="Area pelanggan membantu pemilik usaha melihat daftar pelanggan, persetujuan, dan segmentasi tanpa impor, ekspor, atau edit aktif."
      action="Tambah pelanggan preview"
      lanes={customerLanes}
      metrics={customerMetrics}
      actionGroup={previewActionGroups.customers}
      states={previewStateMatrix.customers}
      table={customersTable}
      detailLinks={customerDetailLinks}
    />
  );
}
