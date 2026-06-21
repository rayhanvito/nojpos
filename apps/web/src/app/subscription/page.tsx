import { TenantPreviewPage } from '@/components/tenant-preview-page';
import { previewActionGroups, previewStateMatrix, subscriptionLanes, tenantSubscriptionStatus } from '@/fixtures/preview';
import { getSubscriptionPageModel } from '@/lib/server/settings-readonly';

export default async function SubscriptionPage() {
  const model = await getSubscriptionPageModel();

  return (
    <TenantPreviewPage
      title="Langganan"
      kicker="Status paket toko"
      description="Halaman langganan menampilkan paket, jadwal perpanjangan, dan posisi invoice tanpa aksi ubah paket, pembatalan, atau penagihan aktif."
      action="Ubah paket preview"
      lanes={subscriptionLanes}
      metrics={model.metrics}
      actionGroup={previewActionGroups.subscription}
      states={previewStateMatrix.subscription}
      table={model.table}
      subscription={tenantSubscriptionStatus}
    />
  );
}
