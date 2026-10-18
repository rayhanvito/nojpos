import assert from 'node:assert/strict'
import { test } from 'node:test'

const {
  normalizeTabParam, normalizeReportTab, normalizeStockTab, normalizePurchaseTab,
  REPORT_TABS, STOCK_TABS, PURCHASE_TABS,
} = await import('../src/shared/tab-params.ts')

test('canonical tab lists are English slugs', () => {
  assert.deepEqual(REPORT_TABS, ['sales', 'products', 'profit', 'stock', 'shift', 'cashier', 'tax'])
  assert.deepEqual(STOCK_TABS, ['stock', 'stocktake', 'movement', 'transfers'])
  assert.deepEqual(PURCHASE_TABS, ['purchases', 'expenses'])
})

test('report aliases: Indonesian values map to canonical and flag for URL replace', () => {
  assert.deepEqual(normalizeReportTab('penjualan'), { tab: 'sales', aliased: true })
  assert.deepEqual(normalizeReportTab('produk'), { tab: 'products', aliased: true })
  assert.deepEqual(normalizeReportTab('stok'), { tab: 'stock', aliased: true })
  assert.deepEqual(normalizeReportTab('kasir'), { tab: 'cashier', aliased: true })
  assert.deepEqual(normalizeReportTab('pajak'), { tab: 'tax', aliased: true })
})

test('stock aliases: riwayat -> movement, opname -> stocktake', () => {
  assert.deepEqual(normalizeStockTab('riwayat'), { tab: 'movement', aliased: true })
  assert.deepEqual(normalizeStockTab('opname'), { tab: 'stocktake', aliased: true })
  assert.deepEqual(normalizeStockTab('stok'), { tab: 'stock', aliased: true })
})

test('transfers is a canonical stock tab value', () => {
  assert.deepEqual(normalizeStockTab('transfers'), { tab: 'transfers', aliased: false })
})

test('purchase aliases: po/pembelian -> purchases, pengeluaran -> expenses', () => {
  assert.deepEqual(normalizePurchaseTab('po'), { tab: 'purchases', aliased: true })
  assert.deepEqual(normalizePurchaseTab('pembelian'), { tab: 'purchases', aliased: true })
  assert.deepEqual(normalizePurchaseTab('pengeluaran'), { tab: 'expenses', aliased: true })
})

test('canonical values pass through without alias flag', () => {
  for (const t of REPORT_TABS) assert.deepEqual(normalizeReportTab(t), { tab: t, aliased: false })
  for (const t of STOCK_TABS) assert.deepEqual(normalizeStockTab(t), { tab: t, aliased: false })
  for (const t of PURCHASE_TABS) assert.deepEqual(normalizePurchaseTab(t), { tab: t, aliased: false })
})

test('missing or unknown values fall back without alias flag', () => {
  assert.deepEqual(normalizeReportTab(null), { tab: 'sales', aliased: false })
  assert.deepEqual(normalizeReportTab(undefined), { tab: 'sales', aliased: false })
  assert.deepEqual(normalizeReportTab('tidak-ada'), { tab: 'sales', aliased: false })
  assert.deepEqual(normalizeStockTab(''), { tab: 'stock', aliased: false })
  assert.deepEqual(normalizePurchaseTab('xyz'), { tab: 'purchases', aliased: false })
})

test('generic normalizeTabParam honours caller canonical/alias/fallback', () => {
  assert.deepEqual(normalizeTabParam('lama', ['baru'], { lama: 'baru' }, 'baru'), { tab: 'baru', aliased: true })
  assert.deepEqual(normalizeTabParam('baru', ['baru'], { lama: 'baru' }, 'baru'), { tab: 'baru', aliased: false })
})
