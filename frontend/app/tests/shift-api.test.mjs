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

test('shift endpoints use the F2.7 paths, filter query, camelCase bodies, and one CSV download route', async () => {
  sessionStorage.setItem('op_access', 'access-token')
  const requests = []
  const shift = { id: 8, cashier_id: 3, cashier_name: 'Ani', started_at: '2026-10-15T08:00:00Z', closed_at: null, opening_cash: 10000, sales: 90000, trx_count: 2, sales_by_method: { Cash: 90000 }, refund_total: 0, expected_cash: 100000, closing_cash_actual: null, cash_variance: null, close_note: null }
  const cashierShift = { shift: { started_at: shift.started_at, opening_cash: shift.opening_cash, sales: shift.sales, trx_count: shift.trx_count, sales_by_method: shift.sales_by_method, refund_total: shift.refund_total, expected_cash: shift.expected_cash } }
  const detail = { shift, transactions: { items: [], total: 0, page: 2, limit: 20 } }
  const closeResponse = { message: 'Shift ditutup.', summary: { ...shift, closed_at: '2026-10-15T16:00:00Z', closing_cash_actual: 101000, cash_variance: 1000 } }
  const forceCloseResponse = { message: 'Shift ditutup paksa.', summary: { ...shift, closed_at: '2026-10-15T16:00:00Z', close_note: 'Ditutup paksa oleh admin' } }
  globalThis.fetch = async (url, init = {}) => {
    requests.push({ url, init, body: init.body ? JSON.parse(init.body) : undefined })
    if (url === '/api/v1/shifts/export?outletId=outlet-2') {
      return new Response('id,kasir\n8,Ani', {
        headers: { 'Content-Disposition': 'attachment; filename="shifts.csv"' },
      })
    }
    const data = url === '/api/v1/cashier/shift' ? cashierShift
      : url.endsWith('/cashier/shift/close') ? closeResponse
        : url.endsWith('/force-close') ? forceCloseResponse
          : url === '/api/v1/shifts/8?page=2&limit=20' ? detail
            : url === '/api/v1/shifts?outletId=outlet-2' ? { items: [shift] }
              : { items: [] }
    return new Response(JSON.stringify(data), {
      headers: { 'Content-Type': 'application/json' },
    })
  }

  assert.deepEqual(await api.apiGetCashierShift(), cashierShift)
  assert.deepEqual(await api.apiCloseCashierShift({ closingCash: 0, note: 'Cocok' }), closeResponse)
  assert.deepEqual(await api.apiListShifts({ outletId: 'outlet-2' }), [shift])
  assert.deepEqual(await api.apiGetShift(8, { page: 2, limit: 20 }), detail)
  assert.deepEqual(await api.apiForceCloseShift(8, {}), forceCloseResponse)
  await api.apiForceCloseShift(8, { closingCash: 90000 })
  const file = await api.apiExportShifts({ outletId: 'outlet-2' })

  assert.deepEqual(requests.map(({ url, init, body }) => [url, init.method, body]), [
    ['/api/v1/cashier/shift', 'GET', undefined],
    ['/api/v1/cashier/shift/close', 'POST', { closingCash: 0, note: 'Cocok' }],
    ['/api/v1/shifts?outletId=outlet-2', 'GET', undefined],
    ['/api/v1/shifts/8?page=2&limit=20', 'GET', undefined],
    ['/api/v1/shifts/8/force-close', 'POST', {}],
    ['/api/v1/shifts/8/force-close', 'POST', { closingCash: 90000 }],
    ['/api/v1/shifts/export?outletId=outlet-2', undefined, undefined],
  ])
  assert.ok(requests.every(({ init }) => init.headers.Authorization === 'Bearer access-token'))
  assert.equal(await file.blob.text(), 'id,kasir\n8,Ani')
  assert.equal(file.filename, 'shifts.csv')
})
