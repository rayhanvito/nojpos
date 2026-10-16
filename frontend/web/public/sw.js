const buildId = new URL(self.location.href).searchParams.get('build') || 'dev'
const cacheName = `openpos-main-shell-${buildId.replace(/[^a-z0-9._-]/gi, '_')}`
const scopePath = new URL(self.registration.scope).pathname
const basePath = scopePath.endsWith('/') ? scopePath : `${scopePath}/`
const rootUrl = new URL(basePath, self.location.origin).href
const standaloneOfflineUrl = `${basePath}offline.html`
const staticShellFiles = new Set([
  `${basePath}Logo-Hitam.png`,
  `${basePath}Logo-Putih.png`,
  `${basePath}favicon.png`,
  `${basePath}favicon.svg`,
])

function isApiRequest(url) {
  const apiPath = `${basePath}api`
  return url.pathname === apiPath || url.pathname.startsWith(`${apiPath}/`)
}

function isShellAsset(url) {
  return url.pathname.startsWith(`${basePath}assets/`) || staticShellFiles.has(url.pathname)
}

function shellAssetsFrom(html) {
  const urls = new Set()
  for (const match of html.matchAll(/\b(?:src|href)=["']([^"']+)["']/gi)) {
    try {
      const url = new URL(match[1], rootUrl)
      if (url.origin === self.location.origin && isShellAsset(url) && !isApiRequest(url)) urls.add(url.href)
    } catch {
      // Abaikan URL markup yang tidak valid.
    }
  }
  return [...urls]
}

async function isStandaloneOfflineClient(clientId) {
  if (!clientId) return false
  const client = await self.clients.get(clientId)
  return !!client && new URL(client.url).pathname === standaloneOfflineUrl
}

self.addEventListener('install', (event) => {
  event.waitUntil((async () => {
    const cache = await caches.open(cacheName)
    const response = await fetch(rootUrl, { cache: 'reload' })
    if (!response.ok) throw new Error('Gagal menyimpan shell aplikasi.')
    const html = await response.clone().text()
    await cache.put(rootUrl, response)
    await Promise.all(shellAssetsFrom(html).map(async (url) => {
      const asset = await fetch(url, { cache: 'reload' })
      if (asset.ok) await cache.put(url, asset)
    }))
    await self.skipWaiting()
  })())
})

self.addEventListener('activate', (event) => {
  event.waitUntil((async () => {
    const oldCaches = (await caches.keys()).filter((name) => name.startsWith('openpos-main-shell-') && name !== cacheName)
    await Promise.all(oldCaches.map((name) => caches.delete(name)))
    await self.clients.claim()
  })())
})

self.addEventListener('fetch', (event) => {
  const request = event.request
  const url = new URL(request.url)
  if (request.method !== 'GET' || url.origin !== self.location.origin || isApiRequest(url)) return
  if (url.pathname === standaloneOfflineUrl) return

  if (request.mode === 'navigate') {
    event.respondWith((async () => {
      if (await isStandaloneOfflineClient(event.clientId)) return fetch(request)
      try {
        return await fetch(request)
      } catch {
        const cache = await caches.open(cacheName)
        return await cache.match(rootUrl) || Response.error()
      }
    })())
    return
  }

  if (!isShellAsset(url)) return
  event.respondWith((async () => {
    if (await isStandaloneOfflineClient(event.clientId)) return fetch(request)
    try {
      const response = await fetch(request)
      if (response.ok) {
        const cache = await caches.open(cacheName)
        await cache.put(request, response.clone())
      }
      return response
    } catch {
      const cache = await caches.open(cacheName)
      return await cache.match(request) || Response.error()
    }
  })())
})
