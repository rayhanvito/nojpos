import type { DetailPreviewFixture, PreviewLane, PreviewMetricFixture, PreviewTableFixture } from './types';

export const attendanceMetrics: readonly PreviewMetricFixture[] = [
  { label: 'Clock records', value: '—', description: 'Jumlah absensi dari backend.', badge: 'Preview' },
  { label: 'Exceptions', value: '—', description: 'Late/missing clock tidak dihitung di client.', badge: 'Gated', tone: 'warning' },
  { label: 'Payroll export', value: 'Read-only', description: 'Export payroll menunggu contract.', badge: 'Disabled' },
] as const;

export const attendanceLanes: readonly PreviewLane[] = [
  { label: 'Capture', items: ['Clock-in list', 'Outlet context', 'Device label'] },
  { label: 'Review', items: ['Exception placement', 'Manager note slot', 'Forbidden state'] },
  { label: 'Governance', items: ['Edit disabled', 'Export disabled', 'Audit required'] },
] as const;

export const attendanceTable: PreviewTableFixture = {
  columns: ['Record', 'Staf', 'Outlet', 'Status'],
  rows: [
    ['ATT-PREVIEW-001', 'Kasir preview', 'Outlet preview', 'Preview'],
    ['ATT-PREVIEW-002', 'Supervisor preview', 'Outlet scope backend', 'Read-only'],
    ['ATT-PREVIEW-003', 'Missing clock placeholder', '—', 'Backend gated'],
  ],
  caption: 'Absensi tidak dapat diubah, import, atau export pada fase UI-first.',
} as const;

export const attendanceDetailPreviews: readonly DetailPreviewFixture[] = [
  {
    id: 'att-preview-001',
    title: 'ATT-PREVIEW-001',
    kicker: 'Attendance detail',
    badge: 'Read-only',
    description: 'Detail record absensi untuk preview koreksi dan timeline tanpa membuat rule jam kerja di browser.',
    backHref: '/attendance',
    backLabel: 'Kembali ke attendance',
    actions: [
      { label: 'Edit record', status: 'backend gated' },
      { label: 'Export record', status: 'disabled' },
    ],
    sections: [
      {
        title: 'Clock identity',
        rows: [
          { label: 'Record ID', value: 'ATT-PREVIEW-001', helper: 'ID fixture untuk route static.' },
          { label: 'Staff', value: 'Kasir preview', helper: 'Staff real dari backend.' },
          { label: 'Outlet', value: 'Outlet preview', helper: 'Outlet scope tetap server-authoritative.' },
        ],
      },
      {
        title: 'Review gate',
        rows: [
          { label: 'Clock-in', value: '—', helper: 'Timestamp real dari server.' },
          { label: 'Correction', value: 'Disabled', helper: 'Butuh approval dan audit contract.' },
          { label: 'Export', value: 'Disabled', helper: 'Payroll export belum aktif.' },
        ],
      },
    ],
    timeline: ['Clock record tampil read-only.', 'Correction action disabled.', 'Payroll/report export gated.'],
  },
  {
    id: 'att-preview-002',
    title: 'ATT-PREVIEW-002',
    kicker: 'Attendance exception',
    badge: 'Backend gated',
    description: 'Preview exception absensi tanpa menghitung keterlambatan atau lembur di frontend.',
    backHref: '/attendance',
    backLabel: 'Kembali ke attendance',
    actions: [
      { label: 'Approve correction', status: 'disabled' },
      { label: 'Reject correction', status: 'disabled' },
    ],
    sections: [
      {
        title: 'Exception',
        rows: [
          { label: 'Type', value: 'Missing clock', helper: 'Tipe exception dari backend.' },
          { label: 'Reason', value: '—', helper: 'Tidak ada alasan palsu.' },
          { label: 'Manager note', value: 'Read-only', helper: 'Form note belum aktif.' },
        ],
      },
      {
        title: 'Policy',
        rows: [
          { label: 'Approval', value: 'Disabled', helper: 'Perlu role policy backend.' },
          { label: 'Audit', value: 'Required', helper: 'Semua koreksi wajib tercatat.' },
          { label: 'Payroll', value: 'Gated', helper: 'Tidak ada kalkulasi payroll di browser.' },
        ],
      },
    ],
    timeline: ['Exception card disiapkan.', 'Approval tetap disabled.', 'Server menentukan status final.'],
  },
] as const;

export const attendanceDetailLinks = attendanceDetailPreviews.map((detail) => ({ href: `/attendance/${detail.id}`, label: detail.title, description: detail.description })) as readonly { href: string; label: string; description: string }[];

export function getAttendanceDetailPreview(id: string) {
  return attendanceDetailPreviews.find((detail) => detail.id === id);
}
