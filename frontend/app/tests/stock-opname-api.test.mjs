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

test('stock opname API uses the contracted paths and camelCase request bodies', async () => {
  const requests = []
  const session = { id: 9, status: 'counting', note: '', created_at: '2026-10-15T00:00:00Z' }
  const detail = {
    ...session,
    items: [{ id: 31, product_id: 7, variant_id: null, system_stock: 5, counted_stock: null, difference: null }],
  }
  globalThis.fetch = async (url, init) => {
    requests.push({ url, method: init.method, body: init.body ? JSON.parse(init.body) : undefined })
    const data = url.endsWith('/stock-opnames') && init.method === 'GET' ? { items: [session] }
      : url.endsWith('/stock-opnames/9') && init.method === 'GET' ? detail
        : session
    return new Response(JSON.stringify(data), { headers: { 'Content-Type': 'application/json' } })
  }

  assert.deepEqual(await api.apiListStockOpnames(), [session])
  await api.apiCreateStockOpname()
  await api.apiCreateStockOpname('Opname bulanan')
  assert.deepEqual(await api.apiGetStockOpname(9), detail)
  await api.apiUpdateStockOpnameItems(9, { items: [{ itemId: 31, countedStock: 0 }] })
  await api.apiPostStockOpname(9)
  await api.apiCancelStockOpname(9)

  assert.deepEqual(requests.map(({ url, method, body }) => [url, method, body]), [
    ['/api/v1/stock-opnames', 'GET', undefined],
    ['/api/v1/stock-opnames', 'POST', {}],
    ['/api/v1/stock-opnames', 'POST', { note: 'Opname bulanan' }],
    ['/api/v1/stock-opnames/9', 'GET', undefined],
    ['/api/v1/stock-opnames/9/items', 'PUT', { items: [{ itemId: 31, countedStock: 0 }] }],
    ['/api/v1/stock-opnames/9/post', 'POST', {}],
    ['/api/v1/stock-opnames/9/cancel', 'POST', {}],
  ])
})
