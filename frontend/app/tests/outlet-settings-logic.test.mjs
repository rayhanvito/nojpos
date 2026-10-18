import assert from 'node:assert/strict'
import { test } from 'node:test'
import { hasOutletOverride } from '../src/features/settings/outlet-settings-logic.ts'

test('an explicit zero is an outlet override and an omitted field inherits', () => {
  assert.equal(hasOutletOverride({ taxPct: 0 }, 'taxPct'), true)
  assert.equal(hasOutletOverride({ receiptFooter: 'Khusus cabang' }, 'taxPct'), false)
})
