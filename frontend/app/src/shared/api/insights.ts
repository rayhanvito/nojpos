// Endpoint wawasan: notifikasi, aktivitas, shift, pengaturan, laporan.

import { request, requestFile, qs } from './core.ts'
import type { StoreSettings, StoreSettingsResponse, Page, CashierShift, Shift, ShiftDetail, ShiftCloseResponse, DashboardAdmin, DashboardCashier, ReportBundle, TaxReport, ActivityItem, NotifCategory, Notification } from './types.ts'
import { normalizeStoreSettings } from './types.ts'

export function apiListNotifications(f: {
  unread?: boolean
  category?: NotifCategory
  status?: 'all' | 'unread' | 'read'
  page?: number
  limit?: number
} = {}) {
  const { status, ...rest } = f
  return request<Page<Notification>>(
    'GET',
    '/notifications' + qs({ ...rest, unread: status === 'unread' ? true : status === 'read' ? false : rest.unread }),
  )
}

// Badge unread; fallback ke list bila endpoint count belum live.
// ponytail: satu endpoint kecil, fallback hapus bila count stabil live.
export async function apiGetNotifUnreadCount(): Promise<number> {
  try {
    const d = await request<{ total: number }>('GET', '/notifications/unread-count')
    return d.total
  } catch {
    const d = await apiListNotifications({ unread: true, limit: 1 })
    return d.total
  }
}

export function apiMarkNotifRead(id: number | string) {
  return request<{ status: string }>('PATCH', `/notifications/${id}/read`)
}

export function apiMarkAllNotifsRead() {
  return request<{ status: string }>('PATCH', '/notifications/read-all')
}

export function apiDeleteNotif(id: number | string) {
  return request<{ status: string }>('DELETE', `/notifications/${id}`)
}

// ── activity log (audit akun, admin-only) ─────────────────────────────
// Kontrak: docs/API-CONTRACT-USER-ACTIVITY.md (diajukan ke backend).
// Null bila endpoint belum live — pemanggil wajib fallback data turunan.

export async function apiListActivity(f: { page?: number; limit?: number } = {}): Promise<Page<ActivityItem> | null> {
  try {
    return await request<Page<ActivityItem>>('GET', '/activity' + qs({ ...f }))
  } catch {
    return null
  }
}

// ── transaksi ────────────────────────────────────────────────────────

export function apiGetCashierShift() {
  return request<{ shift: CashierShift }>('GET', '/cashier/shift')
}

export function apiStartCashierShift(openingCash: number) {
  // Kontrak backend F2.7: field start adalah opening_cash (snake_case), bukan camelCase.
  return request<{ shift: CashierShift }>('POST', '/cashier/shift/start', { opening_cash: openingCash })
}

export function apiCloseCashierShift(body: { closingCash: number; note?: string }) {
  return request<ShiftCloseResponse>('POST', '/cashier/shift/close', body)
}

export function apiListShifts(f: { outletId?: string; from?: string; to?: string } = {}) {
  return request<{ items: Shift[] }>('GET', '/shifts' + qs(f)).then((data) => data.items ?? [])
}

export function apiGetShift(id: number, f: { page?: number; limit?: number } = {}) {
  return request<ShiftDetail>('GET', `/shifts/${id}` + qs(f))
}

export function apiExportShifts(f: { outletId?: string; from?: string; to?: string } = {}) {
  return requestFile('/shifts/export' + qs(f))
}

export function apiForceCloseShift(id: number, body: { closingCash?: number } = {}) {
  return request<ShiftCloseResponse>('POST', `/shifts/${id}/force-close`, body)
}

// ── settings / dashboard / laporan ───────────────────────────────────

export function apiGetSettings() {
  return request<StoreSettingsResponse>('GET', '/settings').then(normalizeStoreSettings)
}

export function apiUpdateSettings(s: StoreSettings) {
  return request<StoreSettingsResponse>('PUT', '/settings', s).then(normalizeStoreSettings)
}

// Ganti kata sandi via OTP email (kontrak README backend §forgot-password).
// Alur: send → verify {email, code} → reset {email, code, new_password}.
// verify opsional (pemanggil lama langsung reset tetap jalan), tapi dipakai
// untuk pecah alur 3 langkah agar form sandi baru tidak muncul sebelum OTP
// terverifikasi. Sukses reset mencabut semua refresh token.

export function apiGetDashboard(date?: string, outletId?: string, range?: { from: string; to: string }) {
  return request<DashboardAdmin | DashboardCashier>('GET', '/dashboard' + qs({ date, outletId, from: range?.from, to: range?.to }))
}

export function apiGetReport(period: string, date?: string, outletId?: string, range?: { from: string; to: string }) {
  return request<ReportBundle>('GET', '/reports' + qs({ period, date, outletId, from: range?.from, to: range?.to }))
}

export function apiGetTaxReport(month: string, outletId?: string) {
  return request<TaxReport>('GET', '/reports/tax' + qs({ month, outletId }))
}

export function apiExportTaxReport(month: string, outletId?: string) {
  return requestFile('/reports/tax/export' + qs({ month, outletId }))
}

// ── util fetch ───────────────────────────────────────────────────────
