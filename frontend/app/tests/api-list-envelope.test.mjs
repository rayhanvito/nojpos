import assert from 'node:assert/strict'
import { test } from 'node:test'

class MemoryStorage {
  values = new Map()
  getItem(key) { return this.values.get(key) ?? null }
  setItem(key, value) { this.values.set(key, String(value)) }
  removeItem(key) { this.values.delete(key) }
}

globalThis.sessionStorage = new MemoryStorage()
globalThis.localStorage = new MemoryStorage()

const api = await import('../src/shared/api.ts')
const featurePromoApi = await import('../src/features/promo/api.ts')
const category = { id: 'c1', name: 'Minuman', active: true, created_at: '2026-10-01' }
const user = { id: 'u1', email: 'kasir@example.com', name: 'Kasir', role: 'cashier', active: true, store_id: 's1', store_name: 'Toko', has_passcode: false }

function respondWithPage(items) {
  globalThis.fetch = async () => new Response(JSON.stringify({ items, total: items.length, page: 1, limit: 50 }), {
    headers: { 'Content-Type': 'application/json' },
  })
}

test('apiListCategories reads categories from the canonical items envelope', async () => {
  respondWithPage([category])
  assert.deepEqual(await api.apiListCategories(), [category])
})

test('apiListUsers reads users from the canonical items envelope', async () => {
  respondWithPage([user])
  assert.deepEqual(await api.apiListUsers(), [{ ...user, outlet_id: undefined }])
})

test('promo, supplier, and shift list methods read canonical Page items', async () => {
  const values = {
    '/promos': [{ id: 1, name: 'Promo' }],
    '/suppliers': [{ id: 2, name: 'Pemasok' }],
    '/shifts': [{ id: 3, created_at: '2026-10-01' }],
  }
  globalThis.fetch = async (url) => {
    const pathname = new URL(String(url), 'http://localhost').pathname
    const items = Object.entries(values).find(([suffix]) => pathname.endsWith(suffix))?.[1] ?? []
    return new Response(JSON.stringify({ items, total: items.length, page: 1, limit: 50 }), {
      headers: { 'Content-Type': 'application/json' },
    })
  }

  assert.deepEqual(await api.apiListPromos(), values['/promos'])
  assert.deepEqual(await api.apiListSuppliers(), values['/suppliers'])
  assert.deepEqual(await api.apiListShifts(), values['/shifts'])
})

test('PromoPage list adapter reads canonical Page items and normalizes nullable fields', async () => {
  const promo = {
    id: 1, created_at: '2026-10-01T00:00:00Z', updated_at: '2026-10-01T00:00:00Z', store_id: 1,
    name: 'Promo', type: 'percent', value: 10, scope: 'all', voucher_code: null,
    starts_at: '2026-10-01T00:00:00Z', ends_at: '2026-10-31T23:59:59Z', min_purchase: 0,
    max_discount: 0, combinable: false, active: true,
  }
  respondWithPage([promo])

  assert.deepEqual(await featurePromoApi.apiListPromos(), [{
    ...promo,
    category_id: null,
    product_id: null,
    happy_start: null,
    happy_end: null,
  }])
})
