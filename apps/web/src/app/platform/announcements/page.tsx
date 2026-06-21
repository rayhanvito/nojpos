import { PlatformPreviewPage } from '@/components/platform-preview-page';
import { platformAnnouncements, platformHealthChecks, platformLanes, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function PlatformAnnouncementsPage() {
  return (
    <PlatformPreviewPage
      title="Pengumuman"
      kicker="Komunikasi toko"
      description="Kirim informasi maintenance, promo paket, dan update fitur untuk toko."
      lanes={platformLanes}
      metrics={[]}
      actionGroup={previewActionGroups.platform}
      states={previewStateMatrix.platform}
      table={platformAnnouncements}
      healthChecks={platformHealthChecks}
      infoCards={[
        { title: 'Draft', value: '2 sample', description: 'Draft masih fixture dan belum tersimpan.' },
        { title: 'Terjadwal', value: '1 sample', description: 'Jadwal belum membuat job pengiriman.' },
        { title: 'Terkirim', value: '1 sample', description: 'Terkirim sample tidak berarti notifikasi pernah dikirim.' },
        { title: 'Maintenance aktif', value: 'Preview', description: 'Maintenance banner belum memengaruhi toko.' },
        { title: 'Buat pengumuman', value: 'Input disabled', description: 'Title, message, dan target selector masih preview.' },
        { title: 'CTA', value: 'Simpan draft belum aktif', description: 'Tidak ada email, WhatsApp, atau in-app notification real.' },
      ]}
      tableKicker="Pengumuman"
      tableTitle="Daftar pengumuman toko"
      tableDescription="Target, jenis, channel, status, dan jadwal memakai data fixture. Composer tidak menyimpan draft atau mengirim notifikasi."
      noticeChildren="Simpan draft belum aktif. Jangan kirim email, WhatsApp, atau notifikasi in-app dari UI preview ini."
    />
  );
}
