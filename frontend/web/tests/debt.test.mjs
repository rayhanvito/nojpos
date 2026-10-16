import assert from 'node:assert/strict'
import { test } from 'node:test'

import * as debt from '../src/lib/debt.ts'

test('debt status labels overdue balances before unpaid or partially paid labels', () => {
  assert.equal(typeof debt.debtStatusLabel, 'function')
  assert.equal(debt.debtStatusLabel('unpaid', '2026-10-14', '2026-10-15'), 'Jatuh Tempo')
  assert.equal(debt.debtStatusLabel('partially_paid', '2026-10-14', '2026-10-15'), 'Jatuh Tempo')
  assert.equal(debt.debtStatusLabel('paid', '2026-10-14', '2026-10-15'), 'Lunas')
  assert.equal(debt.debtStatusLabel('unpaid', '2026-10-15', '2026-10-15'), 'Belum Lunas')
  assert.equal(debt.debtStatusLabel('partially_paid', '2026-10-16', '2026-10-15'), 'Sebagian')
})

test('debt filters distinguish due, unpaid, partial, and paid records', () => {
  assert.equal(typeof debt.filterDebtItems, 'function')
  const items = [
    { status: 'unpaid', due_date: '2026-10-14' },
    { status: 'unpaid', due_date: '2026-10-15' },
    { status: 'partially_paid', due_date: '2026-10-14' },
    { status: 'paid', due_date: '2026-10-14' },
  ]
  assert.deepEqual(debt.filterDebtItems(items, 'overdue', '2026-10-15'), [items[0], items[2]])
  assert.deepEqual(debt.filterDebtItems(items, 'unpaid', '2026-10-15'), [items[0], items[1]])
  assert.deepEqual(debt.filterDebtItems(items, 'partially_paid', '2026-10-15'), [items[2]])
  assert.deepEqual(debt.filterDebtItems(items, 'paid', '2026-10-15'), [items[3]])
})

test('payment amount must be a positive whole Rupiah amount no greater than remaining', () => {
  assert.equal(typeof debt.isDebtPaymentAmountValid, 'function')
  assert.equal(debt.isDebtPaymentAmountValid('', 10000), false)
  assert.equal(debt.isDebtPaymentAmountValid('0', 10000), false)
  assert.equal(debt.isDebtPaymentAmountValid('1.5', 10000), false)
  assert.equal(debt.isDebtPaymentAmountValid('10001', 10000), false)
  assert.equal(debt.isDebtPaymentAmountValid('10000', 10000), true)
})

test('payment submission stays blocked until detail refresh after a successful payment', () => {
  assert.equal(typeof debt.canSubmitDebtPayment, 'function')
  assert.equal(debt.canSubmitDebtPayment('25000', 100000, false), true)
  assert.equal(debt.canSubmitDebtPayment('25000', 100000, true), false)
})
