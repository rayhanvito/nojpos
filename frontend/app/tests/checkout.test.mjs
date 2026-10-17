import assert from 'node:assert/strict'
import { test } from 'node:test'
import * as checkout from '../src/shared/checkout.ts'

const { isCashPaymentValid } = checkout

test('cash payment requires a positive amount unless points are being redeemed', () => {
  assert.equal(isCashPaymentValid('0', -10000, false, false), false)
  assert.equal(isCashPaymentValid('0', -10000, false, true), true)
  assert.equal(isCashPaymentValid('', -10000, false, true), false)
  assert.equal(isCashPaymentValid('5000', -10000, false, true), true)
  assert.equal(isCashPaymentValid('0', -10000, true, false), true)
})

test('credit payment is valid only after a customer has been selected', () => {
  assert.equal(typeof checkout.isCreditPaymentValid, 'function')
  assert.equal(checkout.isCreditPaymentValid(null), false)
  assert.equal(checkout.isCreditPaymentValid(42), true)
})
