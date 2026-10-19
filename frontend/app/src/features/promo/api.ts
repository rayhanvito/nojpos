// API fitur promo: endpoint /promos memakai klien fetch bersama
// (shared/api/core), mengikuti pola modul api per-domain yang ada.
// Tipe payload/respons memakai tipe ter-generate dari OpenAPI backend
// (shared/api/openapi.d.ts, jangan diedit tangan) untuk domain promo:
//   Promo <- components["schemas"]["Promo"], PromoInput <- ["PromoInput"],
//   PromoPreview <- ["PromoPreview"], Message <- ["Message"],
//   PromoId <- components["parameters"]["PromoId"].
// Pemetaan spec <-> fitur yang disengaja:
//   - Field opsional di spec (category_id/product_id/happy_start/happy_end)
//     dinormalkan menjadi selalu terisi | null, seperti tipe lokal lama,
//     agar halaman tidak perlu menangani undefined.
//   - PromoInput spec belum memuat `active` (backend menerimanya dan tipe
//     lokal lama memakainya) -> ditambah sebagai ekstensi lokal fitur.
//   - Body preview: PreviewItem spec memakai productId integer sedangkan
//     katalog FE memakai id string -> memakai CheckoutItem lokal lama;
//     respons preview tetap tipe spec PromoPreview.
import { request } from '../../shared/api/core.ts'
import type { components } from '../../shared/api/openapi'
import type { CheckoutItem, Page } from '../../shared/api/types'

type PromoSchema = components['schemas']['Promo']
type PromoInputSchema = components['schemas']['PromoInput']

export type Promo = Omit<PromoSchema, 'category_id' | 'product_id' | 'happy_start' | 'happy_end'> & {
  category_id: number | null
  product_id: number | null
  happy_start: string | null
  happy_end: string | null
}

export type PromoInput = PromoInputSchema & { active?: boolean }

export type PromoPreview = components['schemas']['PromoPreview']
export type PromoMessage = components['schemas']['Message']
export type PromoId = components['parameters']['PromoId']

function normalizePromo(promo: PromoSchema): Promo {
  return {
    ...promo,
    category_id: promo.category_id ?? null,
    product_id: promo.product_id ?? null,
    happy_start: promo.happy_start ?? null,
    happy_end: promo.happy_end ?? null,
  }
}

export function apiListPromos() {
  return request<Page<PromoSchema>>('GET', '/promos').then((d) => (d.items ?? []).map(normalizePromo))
}


export function apiCreatePromo(body: PromoInput) {
  return request<PromoSchema>('POST', '/promos', body).then(normalizePromo)
}

export function apiUpdatePromo(id: PromoId, body: PromoInput) {
  return request<PromoSchema>('PUT', `/promos/${id}`, body).then(normalizePromo)
}

export function apiSetPromoActive(id: PromoId, active: boolean) {
  return request<PromoMessage>('PATCH', `/promos/${id}/active`, { active })
}

export function apiDeletePromo(id: PromoId) {
  return request<PromoMessage>('DELETE', `/promos/${id}`)
}

export function apiPreviewPromo(body: { items: CheckoutItem[]; voucherCode?: string }) {
  return request<PromoPreview | null>('POST', '/promos/preview', body)
}
