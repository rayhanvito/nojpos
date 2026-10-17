import assert from 'node:assert/strict'
import { test } from 'node:test'
import path from 'node:path'
import ts from 'typescript'

class MemoryStorage {
  values = new Map()
  getItem(key) { return this.values.get(key) ?? null }
  setItem(key, value) { this.values.set(key, String(value)) }
  removeItem(key) { this.values.delete(key) }
}

globalThis.sessionStorage = new MemoryStorage()
globalThis.localStorage = new MemoryStorage()

test('customer API input email is optional and response email is present', () => {
  const fixturePath = path.resolve('tests/customer-email-contract.ts')
  const fixture = `
    import type { Customer, CustomerInput } from '../src/shared/api'
    type Assert<T extends true> = T
    type InputEmailIsOptional = Assert<{} extends Pick<CustomerInput, 'email'> ? true : false>
    type ResponseEmailIsRequired = Assert<{} extends Pick<Customer, 'email'> ? false : true>
    const createWithoutEmail: CustomerInput = { name: 'Sari' }
    const createWithEmail: CustomerInput = { name: 'Sari', email: 'sari@example.com' }
    const customer: Customer = { id: 7, name: 'Sari', phone: '', email: '', address: '', points: 0, created_at: '', updated_at: '' }
    const customerWithEmail: Customer = { ...customer, email: 'sari@example.com' }
    void [createWithoutEmail, createWithEmail, customer, customerWithEmail]
    void (null as unknown as [InputEmailIsOptional, ResponseEmailIsRequired])
  `
  const options = {
    allowImportingTsExtensions: true,
    jsx: ts.JsxEmit.ReactJSX,
    module: ts.ModuleKind.ESNext,
    moduleResolution: ts.ModuleResolutionKind.Bundler,
    noEmit: true,
    skipLibCheck: true,
    strict: true,
    target: ts.ScriptTarget.ES2022,
    types: ['vite/client', 'node'],
  }
  const host = ts.createCompilerHost(options)
  const originalGetSourceFile = host.getSourceFile.bind(host)
  const originalFileExists = host.fileExists.bind(host)
  const originalReadFile = host.readFile.bind(host)
  host.getSourceFile = (fileName, languageVersion, onError, shouldCreateNewSourceFile) => path.resolve(fileName) === fixturePath
    ? ts.createSourceFile(fileName, fixture, languageVersion, true)
    : originalGetSourceFile(fileName, languageVersion, onError, shouldCreateNewSourceFile)
  host.fileExists = (fileName) => path.resolve(fileName) === fixturePath || originalFileExists(fileName)
  host.readFile = (fileName) => path.resolve(fileName) === fixturePath ? fixture : originalReadFile(fileName)

  const program = ts.createProgram([fixturePath], options, host)
  const diagnostics = ts.getPreEmitDiagnostics(program).map((diagnostic) => ts.flattenDiagnosticMessageText(diagnostic.messageText, '\n'))
  assert.deepEqual(diagnostics, [])
})

const api = await import('../src/shared/api.ts')

test('customer API uses the contracted endpoints and camelCase request bodies', async () => {
  sessionStorage.setItem('op_access', 'access-token')
  const requests = []
  const customer = { id: 7, name: 'Sari', phone: '081234567890', email: 'sari@example.com', address: 'Jl. Melati', points: 12, created_at: '2026-10-14T00:00:00Z', updated_at: '2026-10-14T00:00:00Z' }
  globalThis.fetch = async (url, init) => {
    requests.push({ url, init, body: init.body ? JSON.parse(init.body) : undefined })
    return new Response(JSON.stringify(url.endsWith('/transactions') ? { items: [] } : url.endsWith('/customers?q=Sari&page=2') ? { items: [customer], total: 1, page: 2, limit: 20 } : customer), {
      headers: { 'Content-Type': 'application/json' },
    })
  }

  assert.deepEqual(await api.apiListCustomers({ search: 'Sari', page: 2 }), { items: [customer], total: 1, page: 2, limit: 20 })
  assert.deepEqual(await api.apiGetCustomer(7), customer)
  await api.apiCreateCustomer({ name: 'Sari', phone: '081234567890', email: 'sari@example.com', address: 'Jl. Melati' })
  await api.apiUpdateCustomer(7, { name: 'Sari Baru', phone: '081234567890', email: 'sari.baru@example.com', address: 'Jl. Kenanga' })
  await api.apiDeleteCustomer(7)
  assert.deepEqual(await api.apiCustomerTransactions(7), [])

  assert.deepEqual(requests.map(({ url, init, body }) => [url, init.method, body]), [
    ['/api/v1/customers?q=Sari&page=2', 'GET', undefined],
    ['/api/v1/customers/7', 'GET', undefined],
    ['/api/v1/customers', 'POST', { name: 'Sari', phone: '081234567890', email: 'sari@example.com', address: 'Jl. Melati' }],
    ['/api/v1/customers/7', 'PUT', { name: 'Sari Baru', phone: '081234567890', email: 'sari.baru@example.com', address: 'Jl. Kenanga' }],
    ['/api/v1/customers/7', 'DELETE', undefined],
    ['/api/v1/customers/7/transactions', 'GET', undefined],
  ])
  assert.equal(requests[0].init.headers.Authorization, 'Bearer access-token')
})

test('duplicate customer phone error is passed through from the server', async () => {
  globalThis.fetch = async () => new Response(JSON.stringify({ error: 'Nomor telepon sudah digunakan.' }), {
    status: 409,
    headers: { 'Content-Type': 'application/json' },
  })

  await assert.rejects(
    api.apiCreateCustomer({ name: 'Sari', phone: '081234567890', address: '' }),
    (error) => error instanceof api.ApiError && error.status === 409 && error.message === 'Nomor telepon sudah digunakan.',
  )
})

test('checkout forwards customerId and pointsRedeem only when supplied', async () => {
  const requests = []
  globalThis.fetch = async (_url, init) => {
    requests.push(JSON.parse(init.body))
    return new Response(JSON.stringify({ id: 'trx-1' }), { headers: { 'Content-Type': 'application/json' } })
  }
  const items = [{ productId: 'product-1', qty: 1 }]

  await api.apiCheckout({ items, method: 'Cash', paid: 9000, customerId: 7, pointsRedeem: 10 })
  await api.apiCheckout({ items, method: 'Cash', paid: 10000 })

  assert.deepEqual(requests, [
    { items, method: 'Cash', paid: 9000, customerId: 7, pointsRedeem: 10 },
    { items, method: 'Cash', paid: 10000 },
  ])
})

test('settings map snake_case loyalty fields and keep camelCase update payloads', async () => {
  const requests = []
  const response = {
    storeName: 'Toko Sari', address: '', phone: '', taxEnabled: false, taxPct: 0,
    receiptHeader: '', receiptFooter: '', paper: '58mm', timezone: 'Asia/Jakarta',
    loyalty_enabled: true, loyalty_earn_per: 20000, loyalty_point_value: 250,
  }
  globalThis.fetch = async (_url, init) => {
    requests.push({ method: init.method, body: init.body ? JSON.parse(init.body) : undefined })
    return new Response(JSON.stringify(response), { headers: { 'Content-Type': 'application/json' } })
  }

  const loaded = await api.apiGetSettings()
  assert.deepEqual([loaded.loyaltyEnabled, loaded.loyaltyEarnPer, loaded.loyaltyPointValue], [true, 20000, 250])
  await api.apiUpdateSettings({ ...loaded, loyaltyEnabled: false, loyaltyEarnPer: 15000, loyaltyPointValue: 500 })
  assert.equal(requests[0].method, 'GET')
  assert.equal(requests[1].method, 'PUT')
  assert.deepEqual(
    [requests[1].body.loyaltyEnabled, requests[1].body.loyaltyEarnPer, requests[1].body.loyaltyPointValue],
    [false, 15000, 500],
  )
})
