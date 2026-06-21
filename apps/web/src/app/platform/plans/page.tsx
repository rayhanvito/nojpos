import { PlatformPreviewPage } from '@/components/platform-preview-page';
import { platformHealthChecks, platformLanes, platformPlansTable, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function PlatformPlansPage() {
  return (
    <PlatformPreviewPage
      title="Paket"
      kicker="Konsep paket"
      description="Atur konsep paket NojPOS untuk toko yang baru mulai sampai multi-cabang."
      lanes={platformLanes}
      metrics={[]}
      actionGroup={previewActionGroups.platform}
      states={previewStateMatrix.platform}
      table={platformPlansTable}
      healthChecks={platformHealthChecks}
      infoCards={[
        { title: 'Fitur utama', value: 'Kasir · Produk & inventori · Laporan', description: 'Fitur ditampilkan untuk diskusi paket, belum menjadi entitlement backend.' },
        { title: 'Paket berkembang', value: 'Multi-cabang · Tim & absensi', description: 'Limit cabang, user, dan produk masih data contoh.' },
        { title: 'Bantuan prioritas', value: 'Preview', description: 'Add-on dan bantuan prioritas belum tersambung billing.' },
        { title: 'Action gate', value: 'Disabled', description: 'Edit paket, duplikat, dan publish tidak bekerja pada UI preview.' },
      ]}
      tableKicker="Paket"
      tableTitle="Free, Starter, Pro, Business"
      tableDescription="Harga dan limit final belum aktif. Perubahan paket membutuhkan backend billing."
      noticeChildren="Ini hanya UI preview. Jangan gunakan data paket ini sebagai keputusan harga atau limit produksi."
    />
  );
}
