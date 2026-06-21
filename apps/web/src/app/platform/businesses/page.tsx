import { PlatformPreviewPage } from '@/components/platform-preview-page';
import { platformBusinessDetailLinks, platformHealthChecks, platformLanes, platformPrivacyNotes, platformStoresTable, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function PlatformBusinessesPage() {
  return (
    <PlatformPreviewPage
      title="Toko"
      kicker="Daftar toko"
      description="Kelola daftar toko yang memakai NojPOS dan pantau status operasionalnya."
      lanes={platformLanes}
      metrics={[]}
      actionGroup={previewActionGroups.platform}
      states={previewStateMatrix.platform}
      table={platformStoresTable}
      detailLinks={platformBusinessDetailLinks}
      healthChecks={platformHealthChecks}
      infoCards={[
        { title: 'Filter', value: 'Semua · Aktif · Trial · Expired · Suspended · Perlu dicek', description: 'Filter chip masih statis dan belum mengubah query server.' },
        { title: 'Akses bantuan', value: 'Disabled', description: 'Akses bantuan membutuhkan alasan, durasi akses, dan akan tercatat di aktivitas saat backend aktif.' },
        { title: 'Aksi sensitif', value: 'Gated', description: 'Nonaktifkan sementara dan arsipkan hanya preview. Tidak ada aksi real.' },
        ...platformPrivacyNotes.slice(0, 2).map((note) => ({ title: 'Privasi toko', value: 'Dijaga', description: note })),
      ]}
      tableKicker="Toko"
      tableTitle="Daftar toko NojPOS"
      tableDescription="Kolom dibuat ringkas: nama toko, owner, kontak masked, status, paket, cabang, user, last active, dan tagihan."
      noticeChildren="Akses detail toko membutuhkan alasan dan tercatat di aktivitas. Akses bantuan belum aktif pada preview UI."
    />
  );
}
