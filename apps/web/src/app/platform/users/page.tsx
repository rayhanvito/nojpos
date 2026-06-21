import { PlatformPreviewPage } from '@/components/platform-preview-page';
import { platformHealthChecks, platformLanes, platformUsersTable, previewActionGroups, previewStateMatrix } from '@/fixtures/preview';

export default function PlatformUsersPage() {
  return (
    <PlatformPreviewPage
      title="User & Akses"
      kicker="Route pendukung"
      description="Pantau operator internal dan user toko secara terbatas. Route ini tetap ada, tetapi bukan menu utama Super Admin."
      lanes={platformLanes}
      metrics={[]}
      actionGroup={previewActionGroups.platform}
      states={previewStateMatrix.platform}
      table={platformUsersTable}
      healthChecks={platformHealthChecks}
      infoCards={[
        { title: 'Operator NojPOS', value: 'owner · admin · support · finance · ops', description: 'Scope akses dibuat terbatas sesuai peran pada UI preview.' },
        { title: 'User toko', value: 'owner · admin · cashier', description: 'Akses user toko dibatasi dan membutuhkan alasan operasional.' },
        { title: 'Keamanan', value: 'Last login · device · status', description: 'Security logs masih fixture lokal dan tidak mengambil data login real.' },
        { title: 'Aksi sensitif', value: 'Disabled', description: 'Reset password, ban user, dan security action lain disabled/gated.' },
      ]}
      tableKicker="User & Akses"
      tableTitle="Operator NojPOS dan user toko"
      tableDescription="Akses user toko dibatasi. Reset password, ban user, dan view security log sensitif membutuhkan backend dan aktivitas tercatat."
      noticeChildren="Akses user toko dibatasi dan membutuhkan alasan operasional. Halaman ini tidak mengaktifkan auth real atau session browser."
    />
  );
}
