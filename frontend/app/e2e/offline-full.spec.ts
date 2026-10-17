import { expect, test, type Page } from '@playwright/test'

// Alur E2E offline-PENUH sejak awal (perbaikan Unit A) melawan situs dev.
//
// Berbeda dari offline-sync.spec.ts (jaringan putus DI TENGAH checkout
// lewat route abort), spec ini mematikan jaringan sepenuhnya lewat
// context.setOffline(true) SEBELUM klik Bayar / Selesaikan Transaksi —
// skenario bug aslinya: checkout() lama selalu `await refreshShift()`
// lebih dulu, refetch itu tertahan selama offline, dan transaksi tidak
// pernah masuk antrean (menggantung, lalu terkirim online saat pulih).
//
// Bukti yang diasertkan:
// - keterangan Kasir khusus cabang offline-dari-awal
//   ("Transaksi tersimpan di antrean offline. ...") tampil CEPAT selagi
//   perangkat masih offline — bukan keterangan jalur gagal-di-tengah,
// - rekaman IndexedDB nojpos-main-offline-queue berisi tepat 1 item
//   berstatus "menunggu" untuk produk yang sama, dengan clientTxId,
// - setelah online dipulihkan + aplikasi dimuat ulang, item yang sama
//   (clientTxId sama) berstatus "tersinkron" dengan serverTransactionId,
// - transaksi itu BENAR-BENAR ada di server: diverifikasi lewat API
//   GET /transactions memakai token sesi tab yang sama (dicari
//   berdasarkan serverTransactionId dari antrean), bukan cuma UI.
//
// Satu transaksi nyata memang tercipta di server dev oleh uji ini
// setelah sinkron — itu tujuannya.
//
// Kredensial hanya dari env, tidak pernah ditulis ke file/log:
// - E2E_EMAIL / E2E_PASSWORD: akun bootstrap dev.
// - E2E_PASSCODE: passcode pemilik untuk langkah Pilih Akun (wajib hanya
//   bila langkah itu muncul).

const email = process.env.E2E_EMAIL ?? ''
const password = process.env.E2E_PASSWORD ?? ''
const passcode = process.env.E2E_PASSCODE ?? ''

if (!email || !password) {
  throw new Error('E2E_EMAIL dan E2E_PASSWORD wajib diisi (lihat playwright.config.ts).')
}

async function login(page: Page) {
  await page.goto('/login')
  await page.getByPlaceholder('nama@tokosaya.com').fill(email)
  await page.locator('input[type="password"]').first().fill(password)
  await page.getByRole('button', { name: /Masuk ke toko/ }).click()
  await page.waitForURL(
    (url) => url.pathname === '/select-account' || url.pathname.startsWith('/app'),
    { timeout: 30_000 },
  )

  if (new URL(page.url()).pathname === '/select-account') {
    await page.locator('button').filter({ hasText: 'Admin' }).first().click()
    const pin = page.getByLabel('Passcode 5 angka')
    await pin.waitFor({ state: 'visible', timeout: 4_000 }).catch(() => {})
    if (await pin.isVisible().catch(() => false)) {
      if (!passcode) {
        throw new Error(
          'Langkah Pilih Akun meminta passcode pemilik, tetapi E2E_PASSCODE kosong. ' +
            'Isi dari baris passcode_owner= di file bootstrap yang sama dengan E2E_EMAIL/E2E_PASSWORD.',
        )
      }
      await pin.fill(passcode)
      await page.getByRole('button', { name: 'Masuk', exact: true }).click()
    }
    await page.waitForURL((url) => url.pathname.startsWith('/app'), { timeout: 30_000 })
  }
}

async function ensureShiftAktif(page: Page) {
  const statusBtn = page.getByRole('button', { name: /Status:/ })
  await expect(statusBtn).toBeVisible()
  await expect
    .poll(async () => (await statusBtn.innerText()).trim(), { timeout: 30_000 })
    .toMatch(/Shift aktif|Buka Shift/)

  if (!(await statusBtn.innerText()).includes('Buka Shift')) return

  await statusBtn.click()
  const dialog = page.getByRole('dialog', { name: 'Status' })
  await expect(dialog).toBeVisible()
  await dialog.getByLabel(/Modal awal di laci/).fill('100000')
  await dialog.getByRole('button', { name: 'Buka Shift', exact: true }).click()
  await expect(dialog.getByRole('button', { name: 'Tutup Shift' })).toBeVisible({ timeout: 30_000 })
  await dialog.getByRole('button', { name: 'Tutup', exact: true }).click()
  await expect(dialog).toBeHidden()
  await expect
    .poll(async () => (await statusBtn.innerText()).trim(), { timeout: 30_000 })
    .toContain('Shift aktif')
}

async function bersihkanAntreanSelesai(page: Page) {
  const statusBtn = page.getByRole('button', { name: /Status:/ })
  const dialog = page.getByRole('dialog', { name: 'Status' })
  await statusBtn.click()
  await expect(dialog).toBeVisible()
  await expect(dialog.getByText('Koneksi & Antrean')).toBeVisible({ timeout: 30_000 })
  for (let i = 0; i < 20; i++) {
    const buangBtn = dialog
      .locator('article')
      .locator('button')
      .filter({ hasText: /^(Tutup|Abaikan)$/ })
      .first()
    if ((await buangBtn.count()) === 0) break
    await buangBtn.click()
    await page.waitForTimeout(500)
  }
  await expect(dialog.getByText('Belum ada transaksi dalam antrean.')).toBeVisible({ timeout: 30_000 })
  await dialog.locator('button[aria-label="Tutup"]').click()
  await expect(dialog).toBeHidden()
}

interface QueueRecord {
  clientTxId: string
  status: string
  serverTransactionId?: string
  serverTotal?: number
  summary?: { items?: { name: string; qty: number }[]; estimatedTotal?: number }
}

async function readQueue(page: Page): Promise<QueueRecord[]> {
  return page.evaluate(async () => {
    const db = await new Promise<IDBDatabase>((resolve, reject) => {
      const req = indexedDB.open('nojpos-main-offline-queue', 1)
      req.onsuccess = () => resolve(req.result)
      req.onerror = () => reject(req.error ?? new Error('IndexedDB gagal dibuka'))
    })
    try {
      return await new Promise<QueueRecord[]>((resolve, reject) => {
        const tx = db.transaction('transactions', 'readonly')
        const req = tx.objectStore('transactions').getAll()
        req.onsuccess = () => resolve(req.result as QueueRecord[])
        req.onerror = () => reject(req.error ?? new Error('IndexedDB gagal dibaca'))
      })
    } finally {
      db.close()
    }
  })
}

test('checkout offline penuh sejak awal masuk antrean lalu tersinkron dan ada di server', async ({ page }) => {
  await login(page)

  await page.getByRole('link', { name: 'Kasir', exact: true }).click()
  await page.waitForURL('**/app/pos')
  await expect(page.getByLabel('Cari produk')).toBeVisible()

  await ensureShiftAktif(page)
  await bersihkanAntreanSelesai(page)

  const statusBtn = page.getByRole('button', { name: /Status:/ })
  const dialog = page.getByRole('dialog', { name: 'Status' })

  // Susun keranjang SELAGI online sampai tombol Bayar benar-benar aktif
  // (pratinjau promo selesai) — setelah ini tidak ada lagi kebutuhan
  // jaringan untuk menyelesaikan checkout secara lokal.
  const productButtons = page
    .locator('button:not([disabled])')
    .filter({ hasText: /stok/ })
    .filter({ hasText: 'Rp' })
  await expect(productButtons.first()).toBeVisible({ timeout: 30_000 })
  const firstProduct = productButtons.first()
  const productName = (await firstProduct.innerText())
    .split('\n')
    .map((s) => s.trim())
    .filter(Boolean)[0]
  expect(productName, 'nama produk pertama dari katalog').toBeTruthy()
  await firstProduct.click()

  const variantDialog = page.getByRole('dialog', { name: /Pilih Varian/ })
  await variantDialog.waitFor({ state: 'visible', timeout: 3_000 }).catch(() => {})
  if (await variantDialog.isVisible().catch(() => false)) {
    await variantDialog.locator('button:not([disabled])').first().click()
    await expect(variantDialog).toBeHidden()
  }

  const cartPanel = page.locator('aside').filter({ hasText: /keranjang/i })
  await expect(cartPanel.getByText(productName).first()).toBeVisible()

  const bayarBtn = cartPanel.getByRole('button', { name: /Bayar/ })
  await expect(bayarBtn).toBeEnabled({ timeout: 30_000 })

  // --- Offline PENUH sebelum klik Bayar. ---
  await page.context().setOffline(true)
  await expect.poll(() => page.evaluate(() => navigator.onLine)).toBe(false)

  await bayarBtn.click()
  const payDialog = page.getByRole('dialog', { name: 'Pembayaran' })
  await expect(payDialog).toBeVisible()
  await payDialog.getByRole('button', { name: 'Cash', exact: true }).click()
  await payDialog.getByRole('checkbox', { name: /Uang Pas/ }).check()
  await payDialog.getByRole('button', { name: 'Selesaikan Transaksi' }).click()

  // Bukti #1 (anti-gantung): keterangan cabang offline-dari-awal tampil
  // selagi perangkat MASIH offline. Pada kode lama langkah ini tidak
  // pernah tercapai (refetch status shift tertahan offline).
  await expect(
    page.getByText('Transaksi tersimpan di antrean offline. Keranjang siap untuk transaksi berikutnya.'),
  ).toBeVisible({ timeout: 15_000 })
  expect(await page.evaluate(() => navigator.onLine)).toBe(false)
  await expect(cartPanel.getByText('Belum ada item. Klik produk untuk menambah.')).toBeVisible()

  // Bukti #2 (IndexedDB): tepat 1 item menunggu untuk produk ini.
  const queued = await readQueue(page)
  const waiting = queued.filter((item) => item.status === 'menunggu')
  expect(waiting, `antrean IndexedDB: ${JSON.stringify(queued)}`).toHaveLength(1)
  const queuedItem = waiting[0]!
  expect(queuedItem.clientTxId).toBeTruthy()
  expect(queuedItem.summary?.items?.some((item) => item.name.includes(productName))).toBe(true)

  await expect
    .poll(async () => (await statusBtn.getAttribute('aria-label')) ?? '', { timeout: 30_000 })
    .toContain('1 transaksi menunggu sinkronisasi')

  // --- Pulihkan online, muat ulang: AppShell menyinkronkan antrean. ---
  await page.context().setOffline(false)
  await expect.poll(() => page.evaluate(() => navigator.onLine)).toBe(true)
  await page.reload()
  await page.waitForURL('**/app/pos')
  await expect(page.getByLabel('Cari produk')).toBeVisible()

  await expect
    .poll(async () => (await statusBtn.getAttribute('aria-label')) ?? '', {
      timeout: 60_000,
      intervals: [1_000],
    })
    .not.toContain('menunggu sinkronisasi')

  // Bukti #3: item yang sama (clientTxId sama) kini tersinkron.
  await expect
    .poll(async () => {
      const items = await readQueue(page)
      return items.find((item) => item.clientTxId === queuedItem.clientTxId)?.status ?? 'hilang'
    }, { timeout: 60_000, intervals: [1_000] })
    .toBe('tersinkron')
  const syncedItem = (await readQueue(page)).find((item) => item.clientTxId === queuedItem.clientTxId)!
  expect(syncedItem.serverTransactionId, 'serverTransactionId hasil sinkron').toBeTruthy()
  console.log(`Tersinkron: clientTxId=${syncedItem.clientTxId} serverId=${syncedItem.serverTransactionId}`)

  // Bukti #4 (server, bukan cuma UI): transaksi ada lewat API dengan
  // token sesi tab ini, id-nya sama dengan serverTransactionId antrean,
  // dan memuat produk yang dijual.
  const serverTrx = await page.evaluate(async (serverId) => {
    const token = sessionStorage.getItem('op_access')
    const res = await fetch('/api/v1/transactions?page=1&limit=50', {
      headers: token ? { Authorization: `Bearer ${token}` } : {},
    })
    if (!res.ok) throw new Error(`GET /transactions menjawab ${res.status}`)
    const data = (await res.json()) as {
      items?: { id: string; status: string; total: number; items?: { name: string }[] }[]
    }
    return (data.items ?? []).find((trx) => trx.id === serverId) ?? null
  }, syncedItem.serverTransactionId)
  expect(serverTrx, 'transaksi hasil sinkron harus ada di server via API').not.toBeNull()
  expect(serverTrx!.status).toMatch(/^completed/)
  expect(serverTrx!.items?.some((item) => item.name.includes(productName))).toBe(true)

  // Bersihkan rekaman antrean tersinkron dari UI seperti spec lain.
  await statusBtn.click()
  await expect(dialog).toBeVisible()
  const tersinkronItem = dialog.locator('article').filter({ hasText: productName })
  await expect(tersinkronItem).toHaveCount(1, { timeout: 30_000 })
  await expect(tersinkronItem.getByText('Tersinkron', { exact: true })).toBeVisible()
  await tersinkronItem.getByRole('button', { name: 'Tutup', exact: true }).click()
  await expect(dialog.getByText('Belum ada transaksi dalam antrean.')).toBeVisible({ timeout: 30_000 })
  await dialog.locator('button[aria-label="Tutup"]').click()
  await expect(dialog).toBeHidden()
})
