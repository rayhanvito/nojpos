// Endpoint penjualan: promo/voucher dan transaksi kasir.

import { request, requestFile, qs } from './core.ts'
import type { Promo, PromoPreview, CheckoutItem, Trx, Page, PromoInput, CheckoutBody } from './types.ts'

export function apiListPromos() {
  return request<Page<Promo>>('GET', '/promos').then((d) => d.items ?? [])
}

export function apiCreatePromo(body: PromoInput) {
  return request<Promo>('POST', '/promos', body)
}

export function apiUpdatePromo(id: number, body: PromoInput) {
  return request<Promo>('PUT', `/promos/${id}`, body)
}

export function apiSetPromoActive(id: number, active: boolean) {
  return request<{ message: string }>('PATCH', `/promos/${id}/active`, { active })
}

export function apiDeletePromo(id: number) {
  return request<{ message: string }>('DELETE', `/promos/${id}`)
}

export function apiPreviewPromo(body: { items: CheckoutItem[]; voucherCode?: string }) {
  return request<PromoPreview | null>('POST', '/promos/preview', body)
}

export function apiCheckout(body: CheckoutBody) {
  return request<Trx>('POST', '/transactions', body)
}

export function apiListTransactions(f: { q?: string; method?: string; status?: string; cashierId?: string; date?: string; from?: string; to?: string; outletId?: string; page?: number; limit?: number } = {}) {
  return request<Page<Trx>>('GET', '/transactions' + qs(f))
}

export function apiExportTransactions(f: { period: 'today' | 'yesterday' | 'week' | 'month' | 'all'; date?: string; from?: string; to?: string }) {
  return requestFile('/transactions/export' + qs(f))
}

export function apiRefundTransaction(id: string, items: { productId: string; qty: number }[], reason: string) {
  return request<Trx>('POST', `/transactions/${id}/refund`, { items, reason })
}

// ── shift ───────────────────────────────────────────────────────────
