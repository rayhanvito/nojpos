import { PlatformPreviewPage } from '@/components/platform-preview-page';
import { platformHealthChecks, platformHealthMetrics, platformLanes, platformSystemServices, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function PlatformSystemHealthPage() {
  return (
    <PlatformPreviewPage
      title="Sistem"
      kicker="Kondisi layanan"
      description="Pantau kondisi layanan penting yang mendukung operasional NojPOS."
      lanes={platformLanes}
      metrics={platformHealthMetrics}
      actionGroup={previewActionGroups.platform}
      states={previewStateMatrix.platform}
      table={platformSystemServices}
      healthChecks={platformHealthChecks}
      infoCards={[
        { title: 'Status sederhana', value: 'Normal · Perlu dicek · Gangguan', description: 'Bahasa dibuat sederhana agar tidak terlalu DevOps-heavy.' },
        { title: 'Tanpa polling', value: 'Static fixture', description: 'Semua status berasal dari data mock dan tidak melakukan fetch.' },
        { title: 'Raw logs', value: 'Tidak ditampilkan', description: 'Halaman hanya menampilkan incident singkat, bukan log panjang.' },
      ]}
      tableKicker="Sistem"
      tableTitle="Layanan yang perlu dipantau"
      tableDescription="Daftar ringkas layanan, status, dampak, update terakhir, dan penanggung jawab. Tidak ada polling atau monitoring produksi."
      noticeChildren="Sistem adalah UI preview statis. API, database, gateway, webhook, backup, dan notifikasi belum terhubung."
    />
  );
}
