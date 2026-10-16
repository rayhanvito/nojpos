import assert from 'node:assert/strict'
import { test } from 'node:test'

const shift = await import('../src/lib/shift.ts')

test('close payload requires a nonnegative integer and trims an optional note', () => {
  assert.deepEqual(shift.cashierShiftCloseBody('0', ''), { closingCash: 0 })
  assert.deepEqual(shift.cashierShiftCloseBody('125000', '  Selisih laci  '), {
    closingCash: 125000,
    note: 'Selisih laci',
  })
  for (const value of ['', '-1', '1.5', '1e2', '0x10', '9007199254740992', 'abc']) {
    assert.equal(shift.cashierShiftCloseBody(value, ''), null, `${value || 'kosong'} ditolak`)
  }
})

test('force-close payload leaves cash optional and validates a supplied amount', () => {
  assert.deepEqual(shift.forceCloseShiftBody(''), {})
  assert.deepEqual(shift.forceCloseShiftBody('0'), { closingCash: 0 })
  assert.equal(shift.forceCloseShiftBody('-1'), null)
  assert.equal(shift.forceCloseShiftBody('2.5'), null)
  assert.equal(shift.forceCloseShiftBody('1e2'), null)
})

test('sales_by_method summary shows every payment method and fills missing methods with zero', () => {
  assert.deepEqual(shift.shiftMethodRows({ QRIS: 18000, Cash: 32000 }), [
    { method: 'Cash', total: 32000 },
    { method: 'Bank Transfer', total: 0 },
    { method: 'QRIS', total: 18000 },
    { method: 'E-Wallet', total: 0 },
    { method: 'Card', total: 0 },
    { method: 'Credit', total: 0 },
  ])
})

test('cash variance keeps its sign and returns a magnitude for the shared Rupiah formatter', () => {
  assert.deepEqual(shift.shiftVarianceParts(1200), { sign: '+', amount: 1200 })
  assert.deepEqual(shift.shiftVarianceParts(-1200), { sign: '−', amount: 1200 })
  assert.deepEqual(shift.shiftVarianceParts(0), { sign: '', amount: 0 })
})

test('shift list cache key separates sessions and selected outlets', () => {
  assert.notEqual(shift.shiftCacheKey('admin-a:admin', 'all'), shift.shiftCacheKey('admin-b:admin', 'all'))
  assert.notEqual(shift.shiftCacheKey('admin-a:admin', 'all'), shift.shiftCacheKey('admin-a:admin', 'outlet-2'))
})

test('shift rows are hidden until the selected outlet response has loaded', () => {
  const outletA = [{ id: 1, cashier_name: 'Ani' }]
  assert.deepEqual(shift.visibleShiftRowsForKey(outletA, 'outlet-1', 'outlet-2'), [])
  assert.deepEqual(shift.visibleShiftRowsForKey(outletA, 'outlet-1', 'outlet-1'), outletA)
})

test('detail transactions are shown only for the requested page', () => {
  const pageOne = { items: [{ id: 'trx-1' }], page: 1, limit: 10, total: 11 }
  assert.equal(shift.shiftTransactionsForPage(pageOne, 2), null)
  assert.equal(shift.shiftTransactionsForPage(pageOne, 1), pageOne)
})

test('shift status labels distinguish running, closed, and force-closed', () => {
  const base = { cashier_name: 'Kasir Satu', closed_by_name: 'Kasir Satu', closed_at: '2026-10-15T10:00:00Z' }
  assert.equal(shift.shiftStatusLabel({ ...base, closed_at: null }), 'Berjalan')
  assert.equal(shift.shiftStatusLabel(base), 'Selesai')
  assert.equal(shift.shiftStatusLabel({ ...base, closed_by_name: 'Owner' }), 'Ditutup Paksa')
  assert.equal(shift.isForceClosedShift({ ...base, closed_at: null, closed_by_name: 'Owner' }), false)
})

test('shift duration labels use hours and minutes', () => {
  assert.equal(shift.durationLabelFromMs(45 * 60000), '45m')
  assert.equal(shift.durationLabelFromMs((2 * 60 + 5) * 60000), '2j 5m')
  assert.equal(shift.durationLabelFromMs(0), '0m')
  const start = '2026-10-15T01:00:00Z'
  assert.equal(shift.shiftDurationLabel(start, '2026-10-15T03:30:00Z', 0), '2j 30m')
  assert.equal(shift.shiftDurationLabel(start, null, new Date('2026-10-15T01:20:00Z').getTime()), '20m')
})
