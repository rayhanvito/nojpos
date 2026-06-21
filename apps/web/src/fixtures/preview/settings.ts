import type { DisabledFieldFixture, SettingsSectionFixture, PreviewTableFixture, ReadonlyResourceFixture } from './types';

export const settingsSummaryTable: PreviewTableFixture = {
  columns: ['Area', 'Contract', 'State', 'Status'],
  rows: [
    ['Bisnis', '/settings/business', 'Read-only preview', 'Preview'],
    ['Outlet', '/settings/outlet', 'Read-only preview', 'Preview'],
    ['Pembayaran', '/settings/payments', 'Menunggu backend', 'Backend gated'],
    ['Struk', '/settings/receipt', 'Menunggu backend', 'Backend gated'],
  ],
} as const;

export const businessProfileFields: readonly DisabledFieldFixture[] = [
  { label: 'Nama bisnis', value: 'NojPOS Tenant Preview', helper: 'Nilai final dari backend tenant profile.' },
  { label: 'Outlet aktif', value: '—', helper: 'Scope outlet tidak disimpan di browser.' },
  { label: 'Zona waktu', value: 'Asia/Jakarta', helper: 'Akan mengikuti konfigurasi server.' },
  { label: 'Mata uang', value: 'IDR', helper: 'Tidak dipakai untuk kalkulasi browser.' },
] as const;

export const settingsResources: readonly ReadonlyResourceFixture[] = [
  { label: 'Konfigurasi akses', path: '/settings/access' },
  { label: 'Profil bisnis', path: '/settings/business' },
  { label: 'Outlet', path: '/settings/outlet' },
  { label: 'Pembayaran', path: '/settings/payments' },
  { label: 'Struk', path: '/settings/receipt' },
] as const;

export const settingsSections: readonly SettingsSectionFixture[] = [
  {
    title: 'Profil bisnis',
    kicker: 'Business profile',
    badge: 'Preview',
    description: 'Identitas tenant, zona waktu, dan preferensi display hanya tampil sebagai form read-only.',
    path: '/settings/business',
    fields: businessProfileFields,
  },
  {
    title: 'Outlet',
    kicker: 'Outlet scope',
    badge: 'Preview',
    description: 'Daftar outlet dan scope aktif akan mengikuti policy backend saat integrasi.',
    path: '/settings/outlet',
    fields: [
      { label: 'Outlet utama', value: '—', helper: 'Nama dan alamat outlet berasal dari backend.' },
      { label: 'Status operasional', value: 'Read-only', helper: 'Tidak ada open/close store di admin preview.' },
      { label: 'Scope user', value: 'Backend gated', helper: 'Hak akses outlet tidak disimpan di browser.' },
    ],
  },
  {
    title: 'Pembayaran',
    kicker: 'Payment settings',
    badge: 'Backend gated',
    description: 'Payment method, settlement, dan provider hanya ditampilkan setelah contract disetujui.',
    path: '/settings/payments',
    fields: [
      { label: 'Metode aktif', value: '—', helper: 'Sumber kebenaran dari backend payment configuration.' },
      { label: 'Provider', value: 'Menunggu kontrak', helper: 'Tidak ada credential atau key di browser preview.' },
      { label: 'Settlement', value: 'Backend gated', helper: 'Status settlement tidak dihitung atau disimpan di frontend.' },
    ],
  },
  {
    title: 'Struk',
    kicker: 'Receipt settings',
    badge: 'Backend gated',
    description: 'Template struk, footer, dan nomor pajak menunggu contract receipt settings.',
    path: '/settings/receipt',
    fields: [
      { label: 'Header struk', value: 'NojPOS Preview', helper: 'Copy final berasal dari konfigurasi backend.' },
      { label: 'Footer', value: '—', helper: 'Preview tidak menyimpan konfigurasi struk.' },
      { label: 'Printer profile', value: 'Belum tersedia', helper: 'Device dan terminal bukan scope admin preview.' },
    ],
  },
] as const;

export function findSettingsSection(path: string) {
  return settingsSections.find((section) => section.path === path);
}
