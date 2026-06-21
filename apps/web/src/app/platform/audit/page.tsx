import { PlatformPreviewPage } from '@/components/platform-preview-page';
import { platformActivityLogs, platformHealthChecks, platformLanes, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function PlatformAuditPage() {
  return (
    <PlatformPreviewPage
      title="Aktivitas"
      kicker="Riwayat aktivitas"
      description="Lihat riwayat aksi penting untuk menjaga keamanan data toko dan platform."
      lanes={platformLanes}
      metrics={[]}
      actionGroup={previewActionGroups.platform}
      states={previewStateMatrix.platform}
      table={platformActivityLogs}
      healthChecks={platformHealthChecks}
      infoCards={[
        { title: 'Filter', value: 'Semua · Keamanan · Tagihan · Toko · Akses bantuan · Sistem', description: 'Filter masih chip preview dan belum menjalankan query backend.' },
        { title: 'Akses bantuan wajib tercatat', value: 'Wajib', description: 'Sistem harus menyimpan alasan, durasi akses, pelaku, toko target, dan aktivitas yang dilakukan.' },
        { title: 'Aksi sensitif', value: 'Audit log required', description: 'Perubahan status toko, paket, pembayaran, dan user security harus tercatat saat backend aktif.' },
      ]}
      tableKicker="Aktivitas"
      tableTitle="Riwayat aksi penting"
      tableDescription="Tampilan dibuat ringkas tanpa before/after panjang. Detail sensitif tetap membutuhkan alasan dan audit log saat backend aktif."
      noticeChildren="Akses bantuan wajib tercatat sebelum fitur digunakan. Semua aktivitas di halaman ini memakai fixture lokal."
    />
  );
}
