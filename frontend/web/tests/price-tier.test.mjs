import assert from 'node:assert/strict'
import { test } from 'node:test'
import { getUnitPrice } from '../src/lib/price-tier.ts'

test('uses the highest quantity tier that applies, independent of tier order', () => {
  const tiers = [
    { id: 2, min_qty: 10, sell_price: 7000 },
    { id: 1, min_qty: 5, sell_price: 8500 },
  ]

  assert.equal(getUnitPrice(10, 10000, tiers), 7000)
  assert.equal(getUnitPrice(8, 10000, tiers), 8500)
})

test('uses the regular price when no quantity tier applies', () => {
  assert.equal(getUnitPrice(4, 10000, [{ id: 1, min_qty: 5, sell_price: 8500 }]), 10000)
})
