import assert from 'node:assert/strict'
import { test } from 'node:test'

const { taxMonthsForYear, taxMonthLabel, taxRowsForYear } = await import('../src/lib/tax-report.ts')

test('tax months cover January through December of the requested year', () => {
  assert.deepEqual(taxMonthsForYear(2026), [
    '2026-01', '2026-02', '2026-03', '2026-04', '2026-05', '2026-06',
    '2026-07', '2026-08', '2026-09', '2026-10', '2026-11', '2026-12',
  ])
})

test('tax month labels use Indonesian month names', () => {
  assert.equal(taxMonthLabel('2026-01'), 'Januari')
  assert.equal(taxMonthLabel('2026-12'), 'Desember')
})

test('tax rows stay in month order and use zeroes for months without a report', () => {
  const rows = taxRowsForYear(['2026-01', '2026-02'], {
    '2026-02': {
      month: '2026-02',
      omzet_bruto: 120000,
      refund_total: 20000,
      omzet_neto: 100000,
      dpp: 90090,
      ppn_terkumpul: 9910,
      trx_count: 4,
      refund_count: 1,
      pph_final_estimate: 500,
    },
  })

  assert.deepEqual(rows, [
    { month: '2026-01', omzet_bruto: 0, refund_total: 0, omzet_neto: 0, dpp: 0, ppn_terkumpul: 0, trx_count: 0, refund_count: 0, pph_final_estimate: 0 },
    { month: '2026-02', omzet_bruto: 120000, refund_total: 20000, omzet_neto: 100000, dpp: 90090, ppn_terkumpul: 9910, trx_count: 4, refund_count: 1, pph_final_estimate: 500 },
  ])
})
