import assert from 'node:assert/strict'
import { test } from 'node:test'
import * as React from 'react'

const cache = await import('../src/shared/cache.ts')

function renderCache(key, load) {
  const effects = []
  const internals = React.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE
  const previous = internals.H
  internals.H = {
    useState(initial) { return [typeof initial === 'function' ? initial() : initial, () => {}] },
    useRef(current) { return { current } },
    useEffect(effect) { effects.push(effect) },
  }
  try {
    return { hook: cache.useCache(key, load), effects }
  } finally {
    internals.H = previous
  }
}

test('a request started before cache clearing cannot repopulate outlet-scoped data', async () => {
  cache.clearCache()
  let resolveOld
  const oldRequest = new Promise((resolve) => { resolveOld = resolve })
  const oldPage = renderCache('products:user:all', () => oldRequest)
  const unmountOldPage = oldPage.effects[0]()

  cache.clearCache()
  unmountOldPage()
  resolveOld(['outlet A stock'])
  await oldRequest
  await Promise.resolve()

  const newPage = renderCache('products:user:all', async () => ['outlet B stock'])
  assert.equal(newPage.hook.data, null)
  cache.clearCache()
})
