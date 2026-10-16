import assert from 'node:assert/strict'
import { test } from 'node:test'

const { selectCsvExportHandler } = await import('../src/lib/export-role.ts')

test('selects server CSV for admins and client CSV for cashiers', () => {
  const calls = []
  const serverExport = () => calls.push('server')
  const clientExport = () => calls.push('client')

  selectCsvExportHandler('admin', serverExport, clientExport)()
  selectCsvExportHandler('cashier', serverExport, clientExport)()

  assert.deepEqual(calls, ['server', 'client'])
})
