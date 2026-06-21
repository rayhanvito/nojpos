import { TenantPreviewPage } from '@/components/tenant-preview-page';
import { previewActionGroups, previewStateMatrix, subscriptionLanes, subscriptionMetrics, subscriptionTable, tenantSubscriptionStatus } from '@/fixtures/preview';

export default function SubscriptionPage() {
  return (
    <TenantPreviewPage
      title="Langganan"
      kicker="Status paket toko"
      description="Halaman langganan menampilkan paket, jadwal perpanjangan, dan posisi invoice tanpa aksi ubah paket, pembatalan, atau penagihan aktif."
      action="Ubah paket preview"
      lanes={subscriptionLanes}
      metrics={subscriptionMetrics}
      actionGroup={previewActionGroups.subscription}
      states={previewStateMatrix.subscription}
      table={subscriptionTable}
      subscription={tenantSubscriptionStatus}
    />
  );
}
