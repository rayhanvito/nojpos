import { PlatformPreviewPage } from '@/components/platform-preview-page';
import { platformBillingRows, platformHealthChecks, platformLanes, platformRevenueCards, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function PlatformRevenuePage() {
  return (
    <PlatformPreviewPage
      title="Langganan & Tagihan"
      kicker="Tagihan toko"
      description="Pantau paket aktif, jatuh tempo, tagihan, dan pembayaran bermasalah."
      lanes={platformLanes}
      metrics={[]}
      actionGroup={previewActionGroups.platform}
      states={previewStateMatrix.platform}
      table={platformBillingRows}
      healthChecks={platformHealthChecks}
      infoCards={[
        ...platformRevenueCards,
        { title: 'Tab statis', value: 'Langganan · Tagihan · Pembayaran gagal · Riwayat', description: 'Tab masih preview dan belum memanggil backend.' },
        { title: 'Gateway payment', value: 'Belum terhubung', description: 'Data tagihan hanya preview UI.' },
        { title: 'Action gate', value: 'Terkunci', description: 'Tandai lunas, perpanjang manual, dan kirim ulang tagihan tidak bekerja.' },
      ]}
      tableKicker="Langganan & Tagihan"
      tableTitle="Paket aktif dan tagihan toko"
      tableDescription="Kolom ringkas untuk memantau toko, paket, jatuh tempo, status pembayaran, nominal preview, dan update terakhir."
      noticeChildren="Gateway payment belum terhubung. Data tagihan hanya preview UI. Payment retry dan manual mark paid tidak aktif."
    />
  );
}
