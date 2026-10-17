import assert from 'node:assert/strict'
import { test } from 'node:test'

const { isCurrentPurchaseOrderRequest, purchaseOrderItemName, purchaseOrderProductId, purchaseOrderTotal, remainingToReceive } = await import('../src/shared/purchase-order.ts')

test('purchase order product IDs are positive safe integers for the API', () => {
  assert.equal(purchaseOrderProductId('42'), 42)
  assert.equal(purchaseOrderProductId('0'), null)
  assert.equal(purchaseOrderProductId('p42'), null)
})

test('purchase order total uses quantity times integer Rupiah buy price', () => {
  assert.equal(purchaseOrderTotal([
    { qty: '3', buyPrice: '12500' },
    { qty: '2', buyPrice: '4000' },
  ]), 45500)
})

test('remaining receipt quantity stops at zero once fully received', () => {
  assert.equal(remainingToReceive({ qty_ordered: 5, qty_received: 2 }), 3)
  assert.equal(remainingToReceive({ qty_ordered: 5, qty_received: 5 }), 0)
})

test('purchase order item names distinguish variants when the server omits their names', () => {
  const first = purchaseOrderItemName({ product_id: 42, product_name: 'Kaos', variant_id: 21 })
  const second = purchaseOrderItemName({ product_id: 42, product_name: 'Kaos', variant_id: 22 })

  assert.equal(first, 'Kaos · Varian #21')
  assert.equal(second, 'Kaos · Varian #22')
})

test('purchase order item names resolve product and variant names from the catalog', () => {
  const name = purchaseOrderItemName(
    { product_id: 42, variant_id: 21 },
    [{ id: '42', name: 'Kaos', variants: [{ id: 21, name: 'Merah' }] }],
  )

  assert.equal(name, 'Kaos · Merah')
})

test('only the newest detail request remains current when the same PO is reopened', () => {
  const firstRequest = 8
  const latestRequest = 10

  assert.equal(isCurrentPurchaseOrderRequest(firstRequest, latestRequest), false)
  assert.equal(isCurrentPurchaseOrderRequest(latestRequest, latestRequest), true)
})
