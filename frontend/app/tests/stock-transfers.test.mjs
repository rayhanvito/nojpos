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
const wireTransfer = {
  id: 12,
  from_outlet_id: 1,
  from_outlet_name: 'Gudang',
  to_outlet_id: 2,
  to_outlet_name: 'Cabang',
  items: [{ product_id: 55, variant_id: 7, qty: 2 }],
  note: 'Restock',
  created_by_name: 'Admin',
  created_at: '2026-10-17T09:30:00Z',
}

test('apiListTransfers sends canonical filters and normalizes the Page items', async () => {
  let requestUrl
  globalThis.fetch = async (url) => {
    requestUrl = new URL(String(url), 'http://localhost')
    return new Response(JSON.stringify({ items: [wireTransfer], total: 1, page: 2, limit: 10 }), {
      headers: { 'Content-Type': 'application/json' },
    })
  }

  const page = await api.apiListTransfers({ fromOutletId: '1', toOutletId: '2', from: '2026-10-01', to: '2026-10-17', page: 2, limit: 10 })

  assert.equal(requestUrl.pathname, '/api/v1/stock/transfers')
  assert.deepEqual(Object.fromEntries(requestUrl.searchParams), {
    fromOutletId: '1', toOutletId: '2', from: '2026-10-01', to: '2026-10-17', page: '2', limit: '10',
  })
  assert.deepEqual(page, {
    items: [{
      id: '12', fromOutletId: '1', fromOutletName: 'Gudang', toOutletId: '2', toOutletName: 'Cabang',
      items: [{ productId: '55', variantId: 7, qty: 2 }], note: 'Restock', createdBy: 'Admin', createdAt: '2026-10-17T09:30:00Z',
    }],
    total: 1, page: 2, limit: 10,
  })
})

test('apiCreateTransfer posts the camelCase contract and normalizes the response', async () => {
  let request
  globalThis.fetch = async (url, init) => {
    request = { url: new URL(String(url), 'http://localhost').pathname, method: init.method, body: JSON.parse(init.body) }
    return new Response(JSON.stringify(wireTransfer), { headers: { 'Content-Type': 'application/json' } })
  }

  const result = await api.apiCreateTransfer({ fromOutletId: 1, toOutletId: 2, items: [{ productId: 55, variantId: 7, qty: 2 }], note: 'Restock' })

  assert.deepEqual(request, {
    url: '/api/v1/stock/transfers', method: 'POST',
    body: { fromOutletId: 1, toOutletId: 2, items: [{ productId: 55, variantId: 7, qty: 2 }], note: 'Restock' },
  })
  assert.equal(result.fromOutletName, 'Gudang')
  assert.equal(result.items[0].productId, '55')
})

test('apiCreateTransfer preserves stock validation errors from the server', async () => {
  globalThis.fetch = async () => new Response(JSON.stringify({ error: 'Stok tidak mencukupi.', code: 'STOCK_INSUFFICIENT' }), {
    status: 400,
    headers: { 'Content-Type': 'application/json' },
  })

  await assert.rejects(api.apiCreateTransfer({ fromOutletId: 1, toOutletId: 2, items: [{ productId: 55, variantId: null, qty: 2 }], note: '' }),
    (error) => error.status === 400 && error.code === 'stock_insufficient' && error.message === 'Stok tidak mencukupi.')
})
