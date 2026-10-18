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

test('apiRenameCategory patches the category name', async () => {
  let request
  globalThis.fetch = async (url, init) => {
    request = { url: new URL(String(url), 'http://localhost').pathname, method: init.method, body: JSON.parse(init.body) }
    return new Response(JSON.stringify({ message: 'Kategori diperbarui.' }), { headers: { 'Content-Type': 'application/json' } })
  }

  await api.apiRenameCategory('c1', 'Minuman')

  assert.deepEqual(request, { url: '/api/v1/categories/c1', method: 'PATCH', body: { name: 'Minuman' } })
})

test('apiRenameCategory preserves the server validation message', async () => {
  globalThis.fetch = async () => new Response(JSON.stringify({ error: 'Nama kategori sudah digunakan.', code: 'CATEGORY_NAME_DUPLICATE' }), {
    status: 409,
    headers: { 'Content-Type': 'application/json' },
  })

  await assert.rejects(api.apiRenameCategory('c1', 'Minuman'), (error) => error.status === 409 && error.message === 'Nama kategori sudah digunakan.')
})
