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

test('promo API follows the admin endpoints and camelCase request contract', async () => {
  sessionStorage.setItem('op_access', 'access-token')
  const requests = []
  globalThis.fetch = async (url, init) => {
    requests.push({ url, init, body: init.body ? JSON.parse(init.body) : undefined })
    return new Response(JSON.stringify({ items: [] }), { headers: { 'Content-Type': 'application/json' } })
  }

  const body = {
    name: 'Promo Jumat', type: 'percent', value: 10, scope: 'category',
    categoryId: 1, voucherCode: 'JUMAT', startsAt: '2026-10-14T00:00:00Z',
    endsAt: '2026-10-15T00:00:00Z', happyStart: '10:00', happyEnd: '12:00',
    minPurchase: 50000, maxDiscount: 20000,
  }

  await api.apiListPromos()
  await api.apiCreatePromo(body)
  await api.apiUpdatePromo(1, body)
  await api.apiSetPromoActive(1, false)
  await api.apiDeletePromo(1)

  assert.deepEqual(requests.map(({ url, init, body: requestBody }) => [url, init.method, requestBody]), [
    ['/api/v1/promos', 'GET', undefined],
    ['/api/v1/promos', 'POST', body],
    ['/api/v1/promos/1', 'PUT', body],
    ['/api/v1/promos/1/active', 'PATCH', { active: false }],
    ['/api/v1/promos/1', 'DELETE', undefined],
  ])
  assert.equal(requests[0].init.headers.Authorization, 'Bearer access-token')
})

test('promo preview sends cart items and an optional voucher code', async () => {
  let request
  globalThis.fetch = async (url, init) => {
    request = { url, init, body: JSON.parse(init.body) }
    return new Response(JSON.stringify({ promo_id: 1, promo_name: 'Promo Jumat', promo_discount: 5000 }), {
      headers: { 'Content-Type': 'application/json' },
    })
  }

  const items = [{ productId: 'product-1', qty: 2 }]
  const result = await api.apiPreviewPromo({ items, voucherCode: 'JUMAT' })

  assert.equal(request.url, '/api/v1/promos/preview')
  assert.equal(request.init.method, 'POST')
  assert.deepEqual(request.body, { items, voucherCode: 'JUMAT' })
  assert.deepEqual(result, { promo_id: 1, promo_name: 'Promo Jumat', promo_discount: 5000 })
})

test('checkout forwards voucherCode and preserves PROMO_INVALID from the server', async () => {
  let request
  globalThis.fetch = async (_url, init) => {
    request = { init, body: JSON.parse(init.body) }
    return new Response(JSON.stringify({ error: 'Kode voucher tidak berlaku.', code: 'PROMO_INVALID' }), {
      status: 400,
      headers: { 'Content-Type': 'application/json' },
    })
  }

  await assert.rejects(
    api.apiCheckout({ items: [{ productId: 'product-1', qty: 1 }], method: 'Cash', paid: 10000, voucherCode: 'SALAH' }),
    (error) => error instanceof api.ApiError && error.code === 'promo_invalid' && error.message === 'Kode voucher tidak berlaku.',
  )
  assert.deepEqual(request.body, {
    items: [{ productId: 'product-1', qty: 1 }], method: 'Cash', paid: 10000, voucherCode: 'SALAH',
  })
})
