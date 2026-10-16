import assert from 'node:assert/strict'
import { test } from 'node:test'

const { stockOpnameCount, stockOpnameDifference, countStockOpnameDifferences, stockOpnameChangedItems } = await import('../src/lib/stock-opname.ts')

test('stock opname count accepts only non-negative safe integers', () => {
  assert.equal(stockOpnameCount('0'), 0)
  assert.equal(stockOpnameCount('125'), 125)
  assert.equal(stockOpnameCount(''), null)
  assert.equal(stockOpnameCount('-1'), null)
  assert.equal(stockOpnameCount('1.5'), null)
})

test('stock opname difference stays empty until counted and can be negative', () => {
  assert.equal(stockOpnameDifference(8, null), null)
  assert.equal(stockOpnameDifference(8, '0'), -8)
  assert.equal(stockOpnameDifference(8, '11'), 3)
})

test('stock opname summary counts only valid non-zero differences from current drafts', () => {
  const items = [
    { id: 1, system_stock: 8, counted_stock: null },
    { id: 2, system_stock: 5, counted_stock: 7 },
    { id: 3, system_stock: 0, counted_stock: 0 },
  ]

  assert.equal(countStockOpnameDifferences(items), 1)
  assert.equal(countStockOpnameDifferences(items, { 1: '8', 2: '', 3: '0' }), 0)
  assert.equal(countStockOpnameDifferences(items, { 1: '10', 2: '9', 3: 'invalid' }), 2)
})

test('stock opname detects draft counts that would be lost when closing detail', () => {
  const items = [
    { id: 31, counted_stock: null },
    { id: 32, counted_stock: 5 },
  ]

  assert.deepEqual(stockOpnameChangedItems(items, { 31: '', 32: '5' }), [])
  assert.deepEqual(stockOpnameChangedItems(items, { 31: '2', 32: '8' }), items)
})
