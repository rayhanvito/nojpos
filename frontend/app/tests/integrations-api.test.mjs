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

test('API key and webhook lists consume Page envelopes without exposing secrets', async () => {
  globalThis.fetch = async (url) => {
    const pathname = new URL(String(url), 'http://localhost').pathname
    const items = pathname.endsWith('/api-keys')
      ? [{ id: 'k1', name: 'Laporan', prefix: 'njp_abcd', created_at: '2026-10-01T00:00:00Z', last_used_at: null }]
      : [{ id: 'w1', url: 'https://example.com/hook', events: ['transaction.created'], created_at: '2026-10-01T00:00:00Z' }]
    return new Response(JSON.stringify({ items, total: 1, page: 1, limit: 50 }), { headers: { 'Content-Type': 'application/json' } })
  }

  const keys = await api.apiListApiKeys()
  const webhooks = await api.apiListWebhooks()

  assert.equal(keys.items[0].prefix, 'njp_abcd')
  assert.equal('apiKey' in keys.items[0], false)
  assert.equal(webhooks.items[0].events[0], 'transaction.created')
  assert.equal('secret' in webhooks.items[0], false)
  assert.deepEqual([keys.total, keys.page, keys.limit], [1, 1, 50])
})

test('API key and webhook creation use their endpoint contracts and return one-time secrets', async () => {
  const requests = []
  globalThis.fetch = async (url, init) => {
    const pathname = new URL(String(url), 'http://localhost').pathname
    requests.push({ pathname, method: init.method, body: JSON.parse(init.body) })
    const response = pathname.endsWith('/api-keys')
      ? { id: 'k2', name: 'Deploy', prefix: 'njp_xyz', created_at: '2026-10-17T00:00:00Z', last_used_at: null, apiKey: 'njp_full_key' }
      : { id: 'w2', url: 'https://example.com/events', events: ['stock.updated'], created_at: '2026-10-17T00:00:00Z', secret: 'whsec_full_secret' }
    return new Response(JSON.stringify(response), { status: 201, headers: { 'Content-Type': 'application/json' } })
  }

  const key = await api.apiCreateApiKey('Deploy')
  const webhook = await api.apiCreateWebhook({ url: 'https://example.com/events', events: ['stock.updated'] })

  assert.deepEqual(requests, [
    { pathname: '/api/v1/api-keys', method: 'POST', body: { name: 'Deploy' } },
    { pathname: '/api/v1/webhooks', method: 'POST', body: { url: 'https://example.com/events', events: ['stock.updated'] } },
  ])
  assert.equal(key.apiKey, 'njp_full_key')
  assert.equal(webhook.secret, 'whsec_full_secret')
})

test('deletion uses the API key and webhook resource paths', async () => {
  const requests = []
  globalThis.fetch = async (url, init) => {
    requests.push(`${init.method} ${new URL(String(url), 'http://localhost').pathname}`)
    return new Response(JSON.stringify({ message: 'Dihapus.' }), { headers: { 'Content-Type': 'application/json' } })
  }

  await api.apiDeleteApiKey('k1')
  await api.apiDeleteWebhook('w1')

  assert.deepEqual(requests, ['DELETE /api/v1/api-keys/k1', 'DELETE /api/v1/webhooks/w1'])
})

test('webhook validation messages pass through unchanged', async () => {
  globalThis.fetch = async () => new Response(JSON.stringify({ error: 'URL privat tidak diizinkan.', code: 'WEBHOOK_URL_PRIVATE' }), {
    status: 400,
    headers: { 'Content-Type': 'application/json' },
  })

  await assert.rejects(api.apiCreateWebhook({ url: 'http://127.0.0.1/hook', events: ['transaction.created'] }),
    (error) => error.status === 400 && error.message === 'URL privat tidak diizinkan.')
})
