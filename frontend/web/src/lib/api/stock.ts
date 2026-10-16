// Endpoint stok: opname, pergerakan, penyesuaian.

import { request, qs } from './core.ts'
import type { StockOpnameSession, StockOpnameItem, StockOpnameDetail, StockOpnameCountsInput, Product, Movement, Page } from './types.ts'

export function apiListStockOpnames() {
  return request<{ items: StockOpnameSession[] | null }>('GET', '/stock-opnames').then((d) => d.items ?? [])
}

export function apiCreateStockOpname(note?: string) {
  return request<StockOpnameSession & { items?: StockOpnameItem[] }>('POST', '/stock-opnames', note === undefined ? {} : { note })
}

export function apiGetStockOpname(id: number) {
  return request<StockOpnameDetail>('GET', `/stock-opnames/${id}`)
}

export function apiUpdateStockOpnameItems(id: number, body: StockOpnameCountsInput) {
  return request<StockOpnameDetail>('PUT', `/stock-opnames/${id}/items`, body)
}

export function apiPostStockOpname(id: number) {
  return request<StockOpnameSession>('POST', `/stock-opnames/${id}/post`, {})
}

export function apiCancelStockOpname(id: number) {
  return request<StockOpnameSession>('POST', `/stock-opnames/${id}/cancel`, {})
}

export function apiListMovements(f: { type?: string; productId?: string; q?: string; page?: number; limit?: number } = {}) {
  return request<Page<Movement>>('GET', '/movements' + qs(f))
}

export function apiAdjustStock(productId: string, direction: 'plus' | 'minus', qty: number, reason: string) {
  return request<{ product: Product }>('POST', '/stock/adjustments', { productId, direction, qty, reason })
}

// ── notifikasi (admin-only) ────────────────────────────────────────────
// Kontrak: docs/API-CONTRACT-ADMIN-NOTIFICATIONS.md (diajukan ke backend).
// Server wajib scope ke role admin; kasir tidak menerima panel ini.
