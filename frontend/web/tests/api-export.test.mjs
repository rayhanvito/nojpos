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

test('transaction export sends auth and the selected date period', async () => {
  sessionStorage.setItem('op_access', 'access-token')
  let request
  globalThis.fetch = async (url, init) => {
    request = { url, init }
    return new Response('id,tanggal\n1,2026-10-14', {
      headers: { 'Content-Disposition': 'attachment; filename="transactions-2026-10-14.csv"' },
    })
  }

  const file = await api.apiExportTransactions({ period: 'today', date: '2026-10-14' })

  assert.equal(request.url, '/api/v1/transactions/export?period=today&date=2026-10-14')
  assert.equal(request.init.headers.Authorization, 'Bearer access-token')
  assert.equal(await file.blob.text(), 'id,tanggal\n1,2026-10-14')
  assert.equal(file.filename, 'transactions-2026-10-14.csv')
})

test('product export uses the authenticated products endpoint', async () => {
  sessionStorage.setItem('op_access', 'access-token')
  let request
  globalThis.fetch = async (url, init) => {
    request = { url, init }
    return new Response('id,nama\n1,Kopi', {
      headers: { 'Content-Disposition': 'attachment; filename="products-2026-10-14.csv"' },
    })
  }

  const file = await api.apiExportProducts()

  assert.equal(request.url, '/api/v1/products/export')
  assert.equal(request.init.headers.Authorization, 'Bearer access-token')
  assert.equal(await file.blob.text(), 'id,nama\n1,Kopi')
  assert.equal(file.filename, 'products-2026-10-14.csv')
})

test('tax report requests one month with the authenticated reports endpoint', async () => {
  sessionStorage.setItem('op_access', 'access-token')
  const report = {
    month: '2026-10',
    omzet_bruto: 120000,
    refund_total: 20000,
    omzet_neto: 100000,
    dpp: 90090,
    ppn_terkumpul: 9910,
    trx_count: 4,
    refund_count: 1,
    pph_final_estimate: 500,
  }
  let request
  globalThis.fetch = async (url, init) => {
    request = { url, init }
    return new Response(JSON.stringify(report), { headers: { 'Content-Type': 'application/json' } })
  }

  const result = await api.apiGetTaxReport('2026-10')

  assert.equal(request.url, '/api/v1/reports/tax?month=2026-10')
  assert.equal(request.init.headers.Authorization, 'Bearer access-token')
  assert.deepEqual(result, report)
})

test('tax CSV export downloads the selected month from the reports endpoint', async () => {
  sessionStorage.setItem('op_access', 'access-token')
  let request
  globalThis.fetch = async (url, init) => {
    request = { url, init }
    return new Response('bulan,omzet bruto\n2026-10,120000', {
      headers: { 'Content-Disposition': 'attachment; filename="tax-2026-10.csv"' },
    })
  }

  const file = await api.apiExportTaxReport('2026-10')

  assert.equal(request.url, '/api/v1/reports/tax/export?month=2026-10')
  assert.equal(request.init.headers.Authorization, 'Bearer access-token')
  assert.equal(await file.blob.text(), 'bulan,omzet bruto\n2026-10,120000')
  assert.equal(file.filename, 'tax-2026-10.csv')
})
