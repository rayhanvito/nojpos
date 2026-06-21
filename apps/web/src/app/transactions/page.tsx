import { TenantPreviewPage } from '@/components/tenant-preview-page';
import { getTransactionsPageModel } from '@/lib/server/transactions';

// WEB-02B keeps previewStateMatrix coverage semantics via server-side mapped states below.
export default async function TransactionsPage() {
  const model = await getTransactionsPageModel();

  return (
    <TenantPreviewPage
      title="Transaksi"
      kicker="Riwayat transaksi toko"
      description={`${model.description} ${model.detail}`}
      action="Filter transaksi"
      lanes={model.data.lanes}
      metrics={model.data.metrics}
      actionGroup={{
        title: 'Aksi transaksi tetap dikunci',
        description: 'Void, refund, reprint, export, payment retry, dan cash reconciliation belum aktif dari Web Admin.',
        actions: [
          { label: 'Void transaction', category: 'danger', reason: 'Menunggu kontrak aksi sensitif, audit, dan idempotency.' },
          { label: 'Refund transaction', category: 'danger', reason: 'Menunggu kontrak refund aman.' },
          { label: 'Reprint receipt', category: 'export', reason: 'Reprint/export belum diaktifkan dari Web Admin.' },
          { label: 'Export transactions', category: 'export', reason: 'Export job belum masuk scope read-only.' },
        ],
      }}
      states={[
        { tone: model.state === 'real' ? 'empty' : model.state === 'forbidden' ? 'forbidden' : model.state === 'error' ? 'error' : 'unavailable', title: model.title, message: `${model.sourceLabel}: ${model.detail}` },
        { tone: 'loading', title: 'Boundary BFF', message: 'Halaman membaca data melalui server-side helper/BFF, bukan langsung ke Laravel.' },
        { tone: 'forbidden', title: 'Aksi sensitif', message: 'Refund, void, reprint, dan export tetap disabled.' },
      ]}
      table={model.data.table}
      detailLinks={model.data.detailLinks}
    />
  );
}
