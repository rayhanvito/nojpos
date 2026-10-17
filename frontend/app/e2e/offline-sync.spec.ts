import { expect, test, type Page } from '@playwright/test'

// Alur E2E antrean offline + sinkronisasi melawan situs dev, mengikuti
// perilaku AKTUAL aplikasi (dibaca dari kode + dibuktikan dengan probe
// Playwright terhadap dev live, 2026-10-17):
//
// - Antrean offline TIDAK terisi bila browser sudah offline penuh saat
//   tombol "Selesaikan Transaksi" diklik. checkout() di PosPage terlebih
//   dahulu `await refreshShift()`; refetch TanStack Query itu tertahan
//   (networkMode bawaan 'online' di shared/api/query-client.ts, tanpa
//   override) selama perangkat offline, dan begitu koneksi pulih nilai
//   `navigator.onLine` yang dibaca SETELAH await itu sudah true —
//   transaksi lalu selesai sebagai checkout online biasa (struk), tidak
//   pernah masuk antrean. Cabang `if (!online) enqueue` di PosPage dengan
//   demikian tidak terjangkau pada skenario offline-penuh-saat-klik.
//   Ini dicatat sebagai dugaan bug aplikasi untuk diputuskan terpisah;
//   spec ini TIDAK mengubah aplikasi.
// - Jalur antrean yang benar-benar bekerja: kegagalan jaringan DI TENGAH
//   checkout (browser tetap "online", request POST /transactions gagal
//   sebagai TypeError) → catch isNetworkFailure di PosPage memasukkan
//   transaksi ke antrean IndexedDB. Itulah yang disimulasikan spec ini
//   secara jujur: page.route meng-abort HANYA POST checkout; semua
//   request lain (termasuk GET shift) tetap lewat.
//
// Bukti yang diasertkan semuanya render aktual aplikasi:
// - keterangan Kasir "Transaksi tersimpan di antrean. Penyinkronan akan
//   dicoba otomatis saat jaringan tersedia." + keranjang kosong,
// - badge pil Status "1 transaksi menunggu sinkronisasi" dan item
//   berlabel "Menunggu" di popup Status (seksi "Koneksi & Antrean",
//   label dari src/pages/ShiftControl.tsx + src/shared/offline-queue.ts),
// - setelah jaringan dipulihkan (route dicabut) dan aplikasi dimuat ulang
//   (AppShell menyinkronkan antrean saat mount), badge hilang, item yang
//   sama berlabel "Tersinkron", lalu antrean dikosongkan dari popup
//   sampai teks "Belum ada transaksi dalam antrean." tampil lagi.
//
// Satu transaksi nyata memang tercipta di server dev oleh uji ini
// setelah sinkron — itu tujuannya.
//
// Kredensial hanya dari env, tidak pernah ditulis ke file/log:
// - E2E_EMAIL / E2E_PASSWORD: akun bootstrap pemilik.
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
  // Login sukses mendarat di /select-account (bila ada kasir aktif) atau /app.
  await page.waitForURL(
    (url) => url.pathname === '/select-account' || url.pathname.startsWith('/app'),
    { timeout: 30_000 },
  )

  if (new URL(page.url()).pathname === '/select-account') {
    // Pilih akun pemilik: kartu akun dengan lencana "Admin".
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
  // Pil Status tunggal di header AppShell: labelnya "Buka Shift" atau
  // "Shift aktif" (selama memuat labelnya hanya "Shift").
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
  // Berhasil buka → dialog beralih ke rincian shift (ada tombol Tutup Shift).
  await expect(dialog.getByRole('button', { name: 'Tutup Shift' })).toBeVisible({ timeout: 30_000 })
  await dialog.getByRole('button', { name: 'Tutup', exact: true }).click()
  await expect(dialog).toBeHidden()
  await expect
    .poll(async () => (await statusBtn.innerText()).trim(), { timeout: 30_000 })
    .toContain('Shift aktif')
}

// Buka popup Status, buang semua rekaman antrean yang SUDAH selesai
// (tersinkron/gagal — tombol per itemnya "Tutup"/"Abaikan"; item yang masih
// "Menunggu" memang tidak bisa dibuang oleh aplikasi), lalu tutup popup.
// Ini membuat garis dasar uji deterministik: antrean mulai dari kosong.
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
    // Beri waktu antrean termuat ulang setelah pembuangan satu item.
    await page.waitForTimeout(500)
  }
  await expect(dialog.getByText('Belum ada transaksi dalam antrean.')).toBeVisible({ timeout: 30_000 })
  await dialog.locator('button[aria-label="Tutup"]').click()
  await expect(dialog).toBeHidden()
}

test('checkout gagal jaringan masuk antrean lalu tersinkron saat jaringan pulih', async ({ page }) => {
  await login(page)

  await page.getByRole('link', { name: 'Kasir', exact: true }).click()
  await page.waitForURL('**/app/pos')
  await expect(page.getByLabel('Cari produk')).toBeVisible()

  await ensureShiftAktif(page)
  await bersihkanAntreanSelesai(page)

  const statusBtn = page.getByRole('button', { name: /Status:/ })
  const dialog = page.getByRole('dialog', { name: 'Status' })

  // --- Putuskan jaringan DI TENGAH checkout: hanya POST /transactions
  // yang di-abort (kegagalan jaringan pada request itu sendiri); browser
  // tetap "online" dan seluruh request lain tetap berjalan. ---
  await page.route('**/api/v1/transactions', async (route) => {
    if (route.request().method() === 'POST') await route.abort('connectionfailed')
    else await route.continue()
  })

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

  // Produk bervarian membuka pemilih varian: ambil varian pertama yang aktif.
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
  await bayarBtn.click()

  const payDialog = page.getByRole('dialog', { name: 'Pembayaran' })
  await expect(payDialog).toBeVisible()
  await payDialog.getByRole('button', { name: 'Cash', exact: true }).click()
  await payDialog.getByRole('checkbox', { name: /Uang Pas/ }).check()
  await payDialog.getByRole('button', { name: 'Selesaikan Transaksi' }).click()

  // Bukti masuk antrean #1: keterangan jalur-antrean yang dirender Kasir
  // + keranjang kembali kosong (transaksi berpindah ke antrean).
  await expect(
    page.getByText('Transaksi tersimpan di antrean. Penyinkronan akan dicoba otomatis saat jaringan tersedia.'),
  ).toBeVisible({ timeout: 30_000 })
  await expect(cartPanel.getByText('Belum ada item. Klik produk untuk menambah.')).toBeVisible()

  // Bukti masuk antrean #2: badge pil Status = 1 menunggu, dan popup
  // Status menampilkan item berlabel "Menunggu" untuk produk yang sama.
  await expect
    .poll(async () => (await statusBtn.getAttribute('aria-label')) ?? '', { timeout: 30_000 })
    .toContain('1 transaksi menunggu sinkronisasi')
  await statusBtn.click()
  await expect(dialog).toBeVisible()
  await expect(dialog.getByText('1 transaksi menunggu sinkronisasi')).toBeVisible()
  const antreanItem = dialog.locator('article').filter({ hasText: productName })
  await expect(antreanItem).toHaveCount(1)
  await expect(antreanItem.getByText('Menunggu', { exact: true })).toBeVisible()
  await dialog.locator('button[aria-label="Tutup"]').click()
  await expect(dialog).toBeHidden()

  // --- Pulihkan jaringan, lalu muat ulang aplikasi: AppShell
  // menyinkronkan antrean saat mount selama browser online. ---
  await page.unrouteAll({ behavior: 'wait' })
  await page.reload()
  await page.waitForURL('**/app/pos')
  await expect(page.getByLabel('Cari produk')).toBeVisible()

  await expect
    .poll(async () => (await statusBtn.getAttribute('aria-label')) ?? '', {
      timeout: 60_000,
      intervals: [1_000],
    })
    .not.toContain('menunggu sinkronisasi')

  // Bukti tersinkron: item yang sama kini berlabel "Tersinkron" di popup.
  await statusBtn.click()
  await expect(dialog).toBeVisible()
  const tersinkronItem = dialog.locator('article').filter({ hasText: productName })
  await expect(tersinkronItem).toHaveCount(1, { timeout: 60_000 })
  await expect(tersinkronItem.getByText('Tersinkron', { exact: true })).toBeVisible({ timeout: 60_000 })

  // Kosongkan antrean tuntas dari UI (buang rekaman tersinkron) → teks
  // kosong resmi aplikasi tampil kembali: antrean benar-benar kosong.
  await tersinkronItem.getByRole('button', { name: 'Tutup', exact: true }).click()
  await expect(dialog.getByText('Belum ada transaksi dalam antrean.')).toBeVisible({ timeout: 30_000 })
  await dialog.locator('button[aria-label="Tutup"]').click()
  await expect(dialog).toBeHidden()
})
