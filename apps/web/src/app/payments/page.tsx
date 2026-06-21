import { TenantPreviewPage } from '@/components/tenant-preview-page';
import { paymentDetailLinks, paymentLanes, paymentMetrics, paymentStatusPreview, paymentsTable, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function PaymentsPage() {
  return (
    <TenantPreviewPage
      title="Pembayaran"
      kicker="Riwayat pembayaran toko"
      description="Area pembayaran membantu melihat metadata pembayaran, posisi settlement, dan status aksi pembayaran tanpa integrasi gateway aktif."
      action="Filter pembayaran preview"
      lanes={paymentLanes}
      metrics={paymentMetrics}
      actionGroup={previewActionGroups.payments}
      states={previewStateMatrix.payments}
      table={paymentsTable}
      detailLinks={paymentDetailLinks}
      payment={paymentStatusPreview}
    />
  );
}
