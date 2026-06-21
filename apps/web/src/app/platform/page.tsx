import { PlatformPreviewPage } from '@/components/platform-preview-page';
import {
  platformActivityMetrics,
  platformAlerts,
  platformBillingStatus,
  platformHealthChecks,
  platformLanes,
  platformOperationalLinks,
  platformOverviewMetrics,
  platformOverviewTable,
  platformPrivacyNotes,
  platformTenantStatus,
  previewActionGroups,
  previewStateMatrix,
} from '@/fixtures/preview';

export default function PlatformPage() {
  return (
    <PlatformPreviewPage
      title="Ringkasan Platform"
      kicker="Internal NojPOS · Preview UI"
      description="Pantau toko, paket, tagihan, bantuan, dan kondisi sistem NojPOS dari satu tempat."
      lanes={platformLanes}
      metrics={platformOverviewMetrics}
      actionGroup={previewActionGroups.platform}
      states={previewStateMatrix.platform}
      table={platformOverviewTable}
      riskQueue={platformAlerts}
      operationalLinks={platformOperationalLinks}
      healthChecks={platformHealthChecks}
      charts={[
        { title: 'Status toko', description: 'Aktif, trial, expired, dan suspended sebagai agregat contoh.', data: platformTenantStatus },
        { title: 'Tagihan', description: 'Ringkasan status tagihan contoh. Payment gateway belum terhubung.', data: platformBillingStatus },
      ]}
      infoCards={[
        { title: 'Data tenant tetap terlindungi', value: 'Privacy first', description: 'Ringkasan transaksi dan GMV ditampilkan secara agregat. Akses detail toko membutuhkan alasan operasional dan tercatat di aktivitas.' },
        ...platformActivityMetrics.map((metric) => ({ title: metric.label, value: metric.value, description: metric.description })),
        ...platformPrivacyNotes.slice(0, 2).map((note) => ({ title: 'Catatan privasi', value: 'Aman', description: note })),
      ]}
      tableKicker="Yang perlu dicek"
      tableTitle="Prioritas harian tim NojPOS"
      tableDescription="Daftar ini membantu tim melihat toko, tagihan, bantuan, dan sistem yang perlu perhatian tanpa membuka detail transaksi toko."
      noticeChildren="UI preview only — belum ada Laravel API, auth real, database, payment gateway, akses bantuan, atau pengumuman broadcast real."
    />
  );
}
