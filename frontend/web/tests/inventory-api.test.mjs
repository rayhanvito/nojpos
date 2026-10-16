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

const api = await import('../src/lib/api.ts')

test('inventory mutations use the contracted paths and camelCase inputs', async () => {
  const requests = []
  globalThis.fetch = async (url, init) => {
    requests.push({ url, method: init.method, body: init.body ? JSON.parse(init.body) : undefined })
    return new Response(JSON.stringify({ id: 12, name: 'Biru' }), { headers: { 'Content-Type': 'application/json' } })
  }
  const productInput = { name: 'Kaos', sku: 'KAOS', sellPrice: 10000, minStock: 4 }
  const variantInput = { name: 'Biru', sku: 'KAOS-B', buyPrice: 5000, sellPrice: 10000, stock: 8, minStock: 2 }
  const tierInput = { minQty: 10, sellPrice: 8000 }

  await api.apiCreateProduct(productInput)
  await api.apiUpdateProduct('p1', productInput)
  await api.apiCreateVariant('p1', variantInput)
  await api.apiUpdateVariant('p1', 12, variantInput)
  await api.apiDeleteVariant('p1', 12)
  await api.apiCreateProductPriceTier('p1', tierInput)
  await api.apiUpdateProductPriceTier('p1', 14, tierInput)
  await api.apiDeleteProductPriceTier('p1', 14)
  await api.apiCreateVariantPriceTier('p1', 12, tierInput)
  await api.apiUpdateVariantPriceTier('p1', 12, 15, tierInput)
  await api.apiDeleteVariantPriceTier('p1', 12, 15)
  await api.apiCheckout({ items: [{ productId: 'p1', variantId: 12, qty: 10 }], method: 'Cash', paid: 80000 })

  assert.deepEqual(requests, [
    { url: '/api/v1/products', method: 'POST', body: productInput },
    { url: '/api/v1/products/p1', method: 'PUT', body: productInput },
    { url: '/api/v1/products/p1/variants', method: 'POST', body: variantInput },
    { url: '/api/v1/products/p1/variants/12', method: 'PUT', body: variantInput },
    { url: '/api/v1/products/p1/variants/12', method: 'DELETE', body: undefined },
    { url: '/api/v1/products/p1/price-tiers', method: 'POST', body: tierInput },
    { url: '/api/v1/products/p1/price-tiers/14', method: 'PUT', body: tierInput },
    { url: '/api/v1/products/p1/price-tiers/14', method: 'DELETE', body: undefined },
    { url: '/api/v1/products/p1/variants/12/price-tiers', method: 'POST', body: tierInput },
    { url: '/api/v1/products/p1/variants/12/price-tiers/15', method: 'PUT', body: tierInput },
    { url: '/api/v1/products/p1/variants/12/price-tiers/15', method: 'DELETE', body: undefined },
    { url: '/api/v1/transactions', method: 'POST', body: { items: [{ productId: 'p1', variantId: 12, qty: 10 }], method: 'Cash', paid: 80000 } },
  ])
})

test('product detail can be refreshed before nested inventory changes', async () => {
  let requestedUrl
  globalThis.fetch = async (url, init) => {
    requestedUrl = `${init.method} ${url}`
    return new Response(JSON.stringify({ id: 'p1', variants: [] }), { headers: { 'Content-Type': 'application/json' } })
  }

  const product = await api.apiGetProduct('p1')
  assert.equal(requestedUrl, 'GET /api/v1/products/p1')
  assert.equal(product.id, 'p1')
})

test('settings map low_stock_default from the server and keep camelCase updates', async () => {
  const settingsResponse = {
    storeName: 'Toko Sari', address: '', phone: '', taxEnabled: false, taxPct: 0,
    receiptHeader: '', receiptFooter: '', paper: '58mm', timezone: 'Asia/Jakarta', lowStockDefault: 6,
  }
  let updateBody
  globalThis.fetch = async (_url, init) => {
    if (init.method === 'PUT') updateBody = JSON.parse(init.body)
    return new Response(JSON.stringify(settingsResponse), { headers: { 'Content-Type': 'application/json' } })
  }

  const loaded = await api.apiGetSettings()
  assert.equal(loaded.lowStockDefault, 6)
  await api.apiUpdateSettings({ ...loaded, lowStockDefault: 9 })
  assert.equal(updateBody.lowStockDefault, 9)
})
