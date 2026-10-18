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

const outlet = {
  id: 'outlet-1',
  name: 'Cabang Utama',
  address: 'Jl. Melati',
  is_default: true,
}

const {
  outletPayload,
  outletQuery,
  outletRouteKey,
  shouldShowCashierOutletLabel,
  shouldShowOutletFilter,
  shouldShowOutletSwitcher,
  userOutletPayload,
} = await import('../src/shared/outlet.ts')
const api = await import('../src/shared/api.ts')
const { toSession } = await import('../src/shared/session/store.ts')

test('outlet controls stay hidden for one outlet and follow role rules', () => {
  assert.equal(shouldShowOutletSwitcher('admin', 1), false)
  assert.equal(shouldShowOutletFilter('admin', 1), false)
  assert.equal(shouldShowCashierOutletLabel('cashier', 1), false)
  assert.equal(shouldShowOutletSwitcher('admin', 2), true)
  assert.equal(shouldShowOutletFilter('admin', 2), true)
  assert.equal(shouldShowCashierOutletLabel('admin', 2), false)
  assert.equal(shouldShowOutletSwitcher('cashier', 2), false)
  assert.equal(shouldShowOutletFilter('cashier', 2), false)
  assert.equal(shouldShowCashierOutletLabel('cashier', 2), true)
})

test('all outlets and empty selections omit outletId from query params', () => {
  assert.deepEqual(outletQuery(), {})
  assert.deepEqual(outletQuery(null), {})
  assert.deepEqual(outletQuery(''), {})
  assert.deepEqual(outletQuery('outlet-1'), { outletId: 'outlet-1' })
})

test('outlet and user payload helpers trim text and omit empty outlet assignment', () => {
  assert.deepEqual(outletPayload(' Cabang Baru ', ' Jl. Kenanga '), { name: 'Cabang Baru', address: 'Jl. Kenanga' })
  assert.deepEqual(userOutletPayload(' Sari '), { name: 'Sari' })
  assert.deepEqual(userOutletPayload(' Sari ', '   '), { name: 'Sari' })
  assert.deepEqual(userOutletPayload(' Sari ', ' outlet-1 '), { name: 'Sari', outletId: 'outlet-1' })
})

test('route keys differ when the active outlet changes', () => {
  assert.notEqual(outletRouteKey('outlet-1'), outletRouteKey('outlet-2'))
  assert.notEqual(outletRouteKey(null), outletRouteKey('outlet-1'))
})

test('session maps active outlet identity from the auth user', () => {
  const session = toSession({
    id: 'admin-1', email: 'admin@example.com', name: 'Pemilik', role: 'admin', active: true,
    store_id: 'store-1', store_name: 'Toko', has_passcode: false, outlet_id: 'outlet-1', outlet_name: 'Cabang Utama',
  })
  assert.equal(session.outletId, 'outlet-1')
  assert.equal(session.outletName, 'Cabang Utama')
})

test('outlet and cashier APIs use contracted paths and camelCase bodies', async () => {
  const requests = []
  const wireOutlet = { ...outlet, id: 12 }
  const normalizedOutlet = { ...wireOutlet, id: '12' }
  const auth = {
    access_token: 'rotated-access',
    refresh_token: 'rotated-refresh',
    user: { id: 'admin-1', email: 'admin@example.com', name: 'Pemilik', role: 'admin', active: true, store_id: 'store-1', store_name: 'Toko', has_passcode: false, outlet_id: '12', outlet_name: 'Cabang Utama' },
  }
  const wireAuth = { ...auth, user: { ...auth.user, outlet_id: 12 } }
  globalThis.fetch = async (url, init) => {
    const body = init.body ? JSON.parse(init.body) : undefined
    requests.push({ url, method: init.method, body })
    const data = url.endsWith('/outlets') && init.method === 'GET'
      ? { items: [wireOutlet] }
      : url.endsWith('/outlets/switch')
        ? wireAuth
        : url.endsWith('/outlets') || url.endsWith('/outlets/12')
          ? { outlet: wireOutlet }
          : url.endsWith('/users') && init.method === 'GET'
            ? { items: [{ id: 'cashier-1', outlet_id: 12 }], total: 1, page: 1, limit: 50 }
            : url.endsWith('/users')
            ? { user: { id: 'cashier-1', outlet_id: 12 } }
            : { message: 'OK' }
    return new Response(JSON.stringify(data), { headers: { 'Content-Type': 'application/json' } })
  }

  assert.deepEqual(await api.apiListOutlets(), [normalizedOutlet])
  assert.deepEqual(await api.apiCreateOutlet(outletPayload(' Cabang Baru ', ' Jl. Kenanga ')), normalizedOutlet)
  assert.deepEqual(await api.apiUpdateOutlet('12', outletPayload(' Cabang Baru ', ' Jl. Kenanga ')), normalizedOutlet)
  assert.deepEqual(await api.apiSwitchOutlet('12'), auth)
  assert.deepEqual(await api.apiListUsers(), [{ id: 'cashier-1', outlet_id: '12' }])
  assert.deepEqual(await api.apiCreateUser(userOutletPayload(' Sari ', ' 12 ')), { id: 'cashier-1', outlet_id: '12' })
  await api.apiRenameUser('cashier-1', 'Sari', '12')
  await api.apiRenameUser('cashier-1', 'Sari')

  assert.deepEqual(requests.map(({ url, method, body }) => [url, method, body]), [
    ['/api/v1/outlets', 'GET', undefined],
    ['/api/v1/outlets', 'POST', { name: 'Cabang Baru', address: 'Jl. Kenanga' }],
    ['/api/v1/outlets/12', 'PUT', { name: 'Cabang Baru', address: 'Jl. Kenanga' }],
    ['/api/v1/outlets/switch', 'POST', { outletId: 12 }],
    ['/api/v1/users', 'GET', undefined],
    ['/api/v1/users', 'POST', { name: 'Sari', outletId: 12 }],
    ['/api/v1/users/cashier-1', 'PATCH', { name: 'Sari', outletId: 12 }],
    ['/api/v1/users/cashier-1', 'PATCH', { name: 'Sari' }],
  ])
  assert.equal(sessionStorage.getItem('op_access'), 'rotated-access')
  assert.equal(sessionStorage.getItem('op_refresh'), 'rotated-refresh')
})

test('dashboard and report APIs omit outletId for all outlets and include explicit selections', async () => {
  const requests = []
  globalThis.fetch = async (url) => {
    requests.push(String(url))
    return new Response(url.includes('/reports/tax/export') ? 'csv' : '{}', {
      headers: { 'Content-Type': 'application/json' },
    })
  }

  await api.apiGetDashboard('2026-10-15', 'outlet-1')
  await api.apiGetDashboard()
  await api.apiGetReport('today', '2026-10-15', 'outlet-1')
  await api.apiGetReport('all')
  await api.apiGetTaxReport('2026-10', 'outlet-1')
  await api.apiGetTaxReport('2026-10')
  await api.apiExportTaxReport('2026-10', 'outlet-1')
  await api.apiExportTaxReport('2026-10')

  assert.deepEqual(requests.map((url) => url.replace('/api/v1', '')), [
    '/dashboard?date=2026-10-15&outletId=outlet-1',
    '/dashboard',
    '/reports?period=today&date=2026-10-15&outletId=outlet-1',
    '/reports?period=all',
    '/reports/tax?month=2026-10&outletId=outlet-1',
    '/reports/tax?month=2026-10',
    '/reports/tax/export?month=2026-10&outletId=outlet-1',
    '/reports/tax/export?month=2026-10',
  ])
})
