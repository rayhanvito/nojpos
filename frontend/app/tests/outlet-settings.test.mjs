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

const outletSettings = {
  storeName: 'Toko', address: '', phone: '', taxEnabled: true, taxPct: 0,
  receiptHeader: '', receiptFooter: 'Footer Cabang', paper: '58mm', timezone: 'Asia/Jakarta',
  hours: [{ days: 'Senin – Jumat', open: '09:00', close: '18:00' }],
  overrides: { taxPct: 0, receiptFooter: 'Footer Cabang' },
}

test('apiGetOutletSettings returns effective settings and explicit overrides', async () => {
  let request
  globalThis.fetch = async (url, init) => {
    request = { url: new URL(String(url), 'http://localhost').pathname, method: init.method }
    return new Response(JSON.stringify(outletSettings), { headers: { 'Content-Type': 'application/json' } })
  }

  const result = await api.apiGetOutletSettings('out-1')

  assert.deepEqual(request, { url: '/api/v1/outlets/out-1/settings', method: 'GET' })
  assert.equal(result.taxPct, 0)
  assert.deepEqual(result.overrides, { taxPct: 0, receiptFooter: 'Footer Cabang' })
})

test('apiUpdateOutletSettings sends only the supplied fields, including null resets', async () => {
  const bodies = []
  globalThis.fetch = async (_url, init) => {
    if (init.method === 'PUT') bodies.push(JSON.parse(init.body))
    return new Response(JSON.stringify(outletSettings), { headers: { 'Content-Type': 'application/json' } })
  }

  await api.apiUpdateOutletSettings('out-1', { taxPct: 0 })
  await api.apiUpdateOutletSettings('out-1', { receiptFooter: null, hours: null })

  assert.deepEqual(bodies, [{ taxPct: 0 }, { receiptFooter: null, hours: null }])
})
