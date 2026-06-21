import { PlatformPreviewPage } from '@/components/platform-preview-page';
import { platformHealthChecks, platformLanes, platformSupportTickets, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function PlatformSupportPage() {
  return (
    <PlatformPreviewPage
      title="Bantuan"
      kicker="Support toko"
      description="Pantau kendala toko terkait printer, pembayaran, kasir, stok, akun, dan laporan."
      lanes={platformLanes}
      metrics={[]}
      actionGroup={previewActionGroups.platform}
      states={previewStateMatrix.platform}
      table={platformSupportTickets}
      healthChecks={platformHealthChecks}
      infoCards={[
        { title: 'Tiket terbuka', value: '12', description: 'Jumlah tiket contoh untuk board bantuan.' },
        { title: 'Urgent', value: '3', description: 'Prioritas urgent masih label mock.' },
        { title: 'Risiko SLA', value: '4', description: 'SLA belum dihitung backend.' },
        { title: 'Selesai hari ini', value: '5 sample', description: 'Angka selesai hanya data contoh.' },
        { title: 'Konteks toko', value: 'Paket · status · last active · kontak masked', description: 'Konteks dibatasi dan tidak membuka detail transaksi toko.' },
        { title: 'Catatan internal', value: 'Preview box', description: 'Catatan internal dan assignment belum tersimpan karena backend belum aktif.' },
      ]}
      tableKicker="Tiket bantuan"
      tableTitle="Kendala toko yang perlu ditangani"
      tableDescription="Aksi lihat detail, assign, minta akses bantuan, dan tandai selesai hanya preview/disabled."
      noticeChildren="Catatan internal dan assignment belum tersimpan karena backend belum aktif. Akses bantuan belum aktif pada preview UI."
    />
  );
}
