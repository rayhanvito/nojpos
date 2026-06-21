import { TenantPreviewPage } from '@/components/tenant-preview-page';
import { previewActionGroups, previewStateMatrix, transactionDetailLinks, transactionLanes, transactionMetrics, transactionsTable } from '@/fixtures/preview';

export default function TransactionsPage() {
  return (
    <TenantPreviewPage
      title="Transaksi"
      kicker="Riwayat transaksi toko"
      description="Area transaksi bersifat hanya lihat untuk struk, metode bayar, dan status pembayaran. Void, refund, dan ekspor tetap dikunci, serta total tidak dihitung di browser."
      action="Filter transaksi preview"
      lanes={transactionLanes}
      metrics={transactionMetrics}
      actionGroup={previewActionGroups.transactions}
      states={previewStateMatrix.transactions}
      table={transactionsTable}
      detailLinks={transactionDetailLinks}
    />
  );
}
