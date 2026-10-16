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

test('receivables and payables use the F2.3 endpoints and standard payment methods', async () => {
  const names = [
    'apiListReceivables', 'apiGetReceivable', 'apiAddReceivablePayment',
    'apiListPayables', 'apiGetPayable', 'apiAddPayablePayment',
  ]
  for (const name of names) assert.equal(typeof api[name], 'function', `${name} tersedia`)

  const requests = []
  globalThis.fetch = async (url, init) => {
    requests.push({ url, init, body: init.body ? JSON.parse(init.body) : undefined })
    return new Response(JSON.stringify({ items: [] }), { headers: { 'Content-Type': 'application/json' } })
  }

  await api.apiListReceivables()
  await api.apiGetReceivable(7)
  await api.apiAddReceivablePayment(7, { amount: 25000, method: 'Cash' })
  await api.apiListPayables()
  await api.apiGetPayable(9)
  await api.apiAddPayablePayment(9, { amount: 15000, method: 'Bank Transfer' })

  assert.deepEqual(requests.map(({ url, init, body }) => [url, init.method, body]), [
    ['/api/v1/receivables', 'GET', undefined],
    ['/api/v1/receivables/7', 'GET', undefined],
    ['/api/v1/receivables/7/payments', 'POST', { amount: 25000, method: 'Cash' }],
    ['/api/v1/payables', 'GET', undefined],
    ['/api/v1/payables/9', 'GET', undefined],
    ['/api/v1/payables/9/payments', 'POST', { amount: 15000, method: 'Bank Transfer' }],
  ])
})

test('settings normalize creditDueDays and default to 14 when omitted', async () => {
  const urls = []
  globalThis.fetch = async (url) => {
    urls.push(url)
    const settings = {
      storeName: 'Toko', address: '', phone: '', taxEnabled: false, taxPct: 0,
      receiptHeader: '', receiptFooter: '', paper: '58mm', timezone: 'Asia/Jakarta',
    }
    return new Response(JSON.stringify(urls.length === 1 ? { ...settings, credit_due_days: 30 } : settings), {
      headers: { 'Content-Type': 'application/json' },
    })
  }
  assert.equal((await api.apiGetSettings()).creditDueDays, 30)
  assert.equal((await api.apiGetSettings()).creditDueDays, 14)
})

test('detail endpoints normalize omitted payment histories to empty lists', async () => {
  globalThis.fetch = async () => new Response(JSON.stringify({}), {
    headers: { 'Content-Type': 'application/json' },
  })

  assert.deepEqual((await api.apiGetReceivable(7)).payments, [])
  assert.deepEqual((await api.apiGetPayable(9)).payments, [])
})
