import type { PayMethod } from './api'

const PAYMENT_METHODS: PayMethod[] = ['Cash', 'Bank Transfer', 'QRIS', 'E-Wallet', 'Card', 'Credit']

export function shiftCacheKey(sessionKey: string, outletKey: string, rangeKey = 'all'): string {
  return `shifts:${sessionKey}:${outletKey}:${rangeKey}`
}

function nonnegativeInteger(value: string): number | null {
  const digits = value.trim()
  if (!/^\d+$/.test(digits)) return null
  const number = Number(digits)
  return Number.isSafeInteger(number) && number >= 0 ? number : null
}

export function cashierShiftCloseBody(closingCash: string, note: string): { closingCash: number; note?: string } | null {
  const cash = nonnegativeInteger(closingCash)
  if (cash === null) return null
  const trimmedNote = note.trim()
  return { closingCash: cash, ...(trimmedNote ? { note: trimmedNote } : {}) }
}

export function forceCloseShiftBody(closingCash: string): { closingCash?: number } | null {
  if (closingCash.trim() === '') return {}
  const cash = nonnegativeInteger(closingCash)
  return cash === null ? null : { closingCash: cash }
}

export function shiftMethodRows(salesByMethod: Partial<Record<PayMethod, number>>) {
  return PAYMENT_METHODS.map((method) => ({ method, total: salesByMethod[method] ?? 0 }))
}

export function shiftVarianceParts(value: number): { sign: '+' | '−' | ''; amount: number } {
  const sign = value > 0 ? '+' : value < 0 ? '−' : ''
  return { sign, amount: Math.abs(value) }
}

export function visibleShiftRowsForKey<T>(rows: T[], loadedKey: string, selectedKey: string): T[] {
  return loadedKey === selectedKey ? rows : []
}

export function shiftTransactionsForPage<T extends { page: number }>(pageData: T, requestedPage: number): T | null {
  return pageData.page === requestedPage ? pageData : null
}

export interface ShiftStatusLike {
  closed_at: string | null
  closed_by_name?: string | null
  cashier_name: string
}

// Ditutup paksa = sesi ditutup oleh orang lain (admin), bukan kasirnya sendiri.
export function isForceClosedShift(shift: ShiftStatusLike): boolean {
  return shift.closed_at != null && !!shift.closed_by_name && shift.closed_by_name !== shift.cashier_name
}

export function shiftStatusLabel(shift: ShiftStatusLike): 'Berjalan' | 'Selesai' | 'Ditutup Paksa' {
  if (shift.closed_at == null) return 'Berjalan'
  return isForceClosedShift(shift) ? 'Ditutup Paksa' : 'Selesai'
}

export function durationLabelFromMs(ms: number): string {
  if (!(ms > 0)) return '0m'
  const mins = Math.floor(ms / 60000)
  const h = Math.floor(mins / 60)
  const m = mins % 60
  return h > 0 ? `${h}j ${m}m` : `${m}m`
}

export function shiftDurationLabel(startedAt: string, closedAt: string | null, nowMs: number): string {
  const end = closedAt ? new Date(closedAt).getTime() : nowMs
  return durationLabelFromMs(end - new Date(startedAt).getTime())
}
