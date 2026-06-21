import type { DetailPreviewFixture, PreviewLane, PreviewTableFixture } from './types';

export const staffLanes: readonly PreviewLane[] = [
  { label: 'Directory', items: ['Role column', 'Outlet scope', 'Status badge'] },
  { label: 'Review', items: ['Matrix slot later', 'Server policy later', 'Invite flow disabled'] },
  { label: 'History', items: ['Change list slot', 'Actor from server', 'Read-only preview'] },
] as const;

export const staffDirectoryTable: PreviewTableFixture = {
  columns: ['Nama', 'Peran', 'Outlet', 'Status'],
  rows: [
    ['Rina Putri', 'Admin', 'Sudirman', 'Aktif'],
    ['Andi Saputra', 'Kasir', 'Sudirman', 'Aktif'],
    ['Dewi Lestari', 'Kasir', 'Kemang', 'Aktif'],
    ['Role supervisor', 'Menunggu kontrak', '—', 'Backend gated'],
  ],
} as const;

export const staffDetailPreviews: readonly DetailPreviewFixture[] = [
  {
    id: 'rina-putri',
    title: 'Rina Putri',
    kicker: 'Staff detail preview',
    badge: 'Read-only',
    description: 'Profil staf menampilkan struktur role, outlet, dan status tanpa membuat auth browser atau invite flow aktif.',
    backHref: '/staff',
    backLabel: 'Kembali ke staf',
    actions: [
      { label: 'Ubah role', status: 'disabled' },
      { label: 'Kirim invite', status: 'backend gated' },
    ],
    sections: [
      {
        title: 'Identitas staf',
        rows: [
          { label: 'Nama', value: 'Rina Putri', helper: 'Nama fixture untuk preview.' },
          { label: 'Peran', value: 'Admin', helper: 'Role final mengikuti backend authorization.' },
          { label: 'Outlet', value: 'Sudirman', helper: 'Scope outlet real dari server.' },
        ],
      },
      {
        title: 'Access guard',
        rows: [
          { label: 'Login web', value: 'Not implemented', helper: 'Browser auth belum disetujui.' },
          { label: 'Status', value: 'Preview active', helper: 'Status tidak dipakai untuk izin real.' },
          { label: 'Audit', value: 'Backend-owned', helper: 'Perubahan role wajib diaudit server.' },
        ],
      },
    ],
    timeline: [
      'Staff detail tersedia untuk review layout.',
      'Invite dan role update tetap disabled.',
      'Access policy menunggu gate API/auth.',
    ],
  },
  {
    id: 'andi-saputra',
    title: 'Andi Saputra',
    kicker: 'Staff detail preview',
    badge: 'Preview',
    description: 'Contoh detail kasir untuk menguji state role/outlet tanpa mengaktifkan session browser.',
    backHref: '/staff',
    backLabel: 'Kembali ke staf',
    actions: [
      { label: 'Reset akses', status: 'disabled' },
      { label: 'Review izin', status: 'backend gated' },
    ],
    sections: [
      {
        title: 'Identitas staf',
        rows: [
          { label: 'Nama', value: 'Andi Saputra', helper: 'Fixture preview.' },
          { label: 'Peran', value: 'Kasir', helper: 'Role real harus dari server.' },
          { label: 'Outlet', value: 'Sudirman', helper: 'Outlet scope belum dibaca.' },
        ],
      },
      {
        title: 'Access guard',
        rows: [
          { label: 'Permission matrix', value: '—', helper: 'Matrix belum dikontrak.' },
          { label: 'Last access', value: '—', helper: 'Tidak ada session tracking browser.' },
          { label: 'Audit', value: 'Server-owned', helper: 'Audit detail menunggu API.' },
        ],
      },
    ],
    timeline: [
      'Detail kasir dibuat sebagai static route.',
      'Reset akses tidak tersedia di preview.',
      'Permission matrix akan mengikuti backend.',
    ],
  },
] as const;

export const staffDetailLinks = staffDetailPreviews.map((detail) => ({
  href: `/staff/${detail.id}`,
  label: detail.title,
  description: detail.description,
})) as readonly { href: string; label: string; description: string }[];

export function getStaffDetailPreview(id: string) {
  return staffDetailPreviews.find((detail) => detail.id === id);
}
