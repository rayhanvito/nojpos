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

test('supplier and purchase order API use contracted paths and camelCase request bodies', async () => {
  const requests = []
  const supplier = { id: 4, name: 'PT Maju', contact: '0812', address: 'Jl. Melati' }
  const order = { id: 9, number: 'PO-20261014-0001', supplier_id: 4, status: 'draft', created_at: '2026-10-14T00:00:00Z' }
  const detail = { ...order, items: [{ id: 31, product_id: 7, variant_id: 12, qty_ordered: 5, qty_received: 0, buy_price: 12000 }] }
  globalThis.fetch = async (url, init) => {
    requests.push({ url, method: init.method, body: init.body ? JSON.parse(init.body) : undefined })
    const data = url.endsWith('/suppliers') && init.method === 'GET' ? { items: [supplier] }
      : url.endsWith('/purchase-orders') && init.method === 'GET' ? { items: [order] }
        : url.endsWith('/purchase-orders/9') && init.method === 'GET' ? detail
          : supplier
    return new Response(JSON.stringify(data), { headers: { 'Content-Type': 'application/json' } })
  }

  const supplierInput = { name: 'PT Maju', contact: '0812', address: 'Jl. Melati' }
  const orderInput = { supplierId: 4, items: [{ productId: 7, variantId: 12, qty: 5, buyPrice: 12000 }] }
  assert.deepEqual(await api.apiListSuppliers(), [supplier])
  await api.apiCreateSupplier(supplierInput)
  await api.apiUpdateSupplier(4, supplierInput)
  await api.apiDeleteSupplier(4)
  assert.deepEqual(await api.apiListPurchaseOrders(), [order])
  assert.deepEqual(await api.apiGetPurchaseOrder(9), detail)
  await api.apiCreatePurchaseOrder(orderInput)
  await api.apiUpdatePurchaseOrder(9, orderInput)
  await api.apiDeletePurchaseOrder(9)
  await api.apiOrderPurchaseOrder(9)
  await api.apiReceivePurchaseOrder(9, { items: [{ itemId: 31, qty: 2 }] })
  await api.apiCancelPurchaseOrder(9)

  assert.deepEqual(requests.map(({ url, method, body }) => [url, method, body]), [
    ['/api/v1/suppliers', 'GET', undefined],
    ['/api/v1/suppliers', 'POST', supplierInput],
    ['/api/v1/suppliers/4', 'PUT', supplierInput],
    ['/api/v1/suppliers/4', 'DELETE', undefined],
    ['/api/v1/purchase-orders', 'GET', undefined],
    ['/api/v1/purchase-orders/9', 'GET', undefined],
    ['/api/v1/purchase-orders', 'POST', orderInput],
    ['/api/v1/purchase-orders/9', 'PUT', orderInput],
    ['/api/v1/purchase-orders/9', 'DELETE', undefined],
    ['/api/v1/purchase-orders/9/order', 'POST', undefined],
    ['/api/v1/purchase-orders/9/receive', 'POST', { items: [{ itemId: 31, qty: 2 }] }],
    ['/api/v1/purchase-orders/9/cancel', 'POST', undefined],
  ])
})
