import { PlatformPreviewPage } from '@/components/platform-preview-page';
import { platformBillingRows, platformHealthChecks, platformLanes, platformRevenueCards, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function PlatformSubscriptionsPage() {
  return (
    <PlatformPreviewPage
      title="Langganan"
      kicker="Route lama"
      description="Pantau status paket aktif, masa trial, jatuh tempo, dan pembaruan langganan. Route ini tetap build, sementara menu utama memakai Langganan & Tagihan."
      lanes={platformLanes}
      metrics={[]}
      actionGroup={previewActionGroups.platform}
      states={previewStateMatrix.platform}
      table={platformBillingRows}
      healthChecks={platformHealthChecks}
      infoCards={[
        ...platformRevenueCards,
        { title: 'Filter', value: 'Aktif · Trial · Segera habis · Belum dibayar · Dibatalkan', description: 'Filter hanya preview dan belum memanggil backend.' },
        { title: 'Aksi', value: 'Disabled', description: 'Manual extend, change plan, dan cancel subscription tetap dikunci.' },
      ]}
      tableKicker="Langganan"
      tableTitle="Status paket aktif dan masa berlaku"
      tableDescription="Data langganan memakai fixture lokal. Perubahan paket, perpanjangan manual, dan cancel tidak bekerja."
      noticeChildren="Aksi langganan membutuhkan audit log dan backend billing saat aktif. Semua data di halaman ini sample UI preview."
    />
  );
}
