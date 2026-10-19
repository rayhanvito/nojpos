import { expect, test, type Page } from '@playwright/test'

// Uji tab Kategori di halaman Produk & Kategori (unit tab-kategori 2026-10-19):
// - ?tab=kategori (alias Indonesia) membuka tab Kategori; snapshot baru
//   desktop + mobile untuk tab ini.
// - Alur CRUD lewat UI: buat kategori uji -> terlihat "0 produk" ->
//   ubah nama -> hapus (dialog confirm native diterima). Data uji dibuat
//   dan dihapus oleh test ini sendiri; nama uji sengaja TIDAK berawalan
//   huruf awal alfabet.
// - Luapan horizontal 390px = 0 di tab Kategori maupun tab Produk.
// Kredensial hanya dari env (E2E_EMAIL/E2E_PASSWORD/E2E_PASSCODE), sama
// seperti visual.spec.ts; nilai tidak pernah ditulis ke file/log.

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
      if (!passcode) throw new Error('E2E_PASSCODE kosong tetapi langkah Pilih Akun meminta passcode.')
      await pin.fill(passcode)
      await page.getByRole('button', { name: 'Masuk', exact: true }).click()
    }
    await page.waitForURL((url) => url.pathname.startsWith('/app'), { timeout: 30_000 })
  }
}

async function tungguStabilKategori(page: Page) {
  await expect(page.getByRole('heading', { name: 'Produk & Kategori', exact: true }).first()).toBeVisible({ timeout: 30_000 })
  await expect(page.getByRole('tab', { name: 'Kategori' })).toHaveAttribute('aria-selected', 'true', { timeout: 30_000 })
  await expect
    .poll(
      async () =>
        (await page.locator('[data-slot="skeleton"]').count()) +
        (await page.locator('[aria-busy="true"]').count()),
      { timeout: 30_000 },
    )
    .toBe(0)
  await page.waitForLoadState('networkidle', { timeout: 30_000 }).catch(() => {})
  await page.waitForTimeout(800)
}

async function assertNoOverflow(page: Page, label: string) {
  const m = await page.evaluate(() => ({
    sw: document.documentElement.scrollWidth,
    cw: document.documentElement.clientWidth,
  }))
  expect(m.sw, `${label}: scrollWidth ${m.sw} > clientWidth ${m.cw}`).toBeLessThanOrEqual(m.cw)
}

const NAMA_AWAL = 'Zeta Uji Kategori'
const NAMA_BARU = 'Zeta Uji Rubah Nama'

test('tab Kategori: tampil + snapshot + CRUD kategori uji', async ({ page }) => {
  page.on('dialog', (d) => void d.accept())
  await page.setViewportSize({ width: 1440, height: 900 })
  await login(page)

  // Deep-link alias Indonesia harus mendarat di tab Kategori.
  await page.goto('/app/products?tab=kategori')
  await tungguStabilKategori(page)
  expect(new URL(page.url()).searchParams.get('tab')).toBe('categories')

  // Daftar kategori tampil dan setiap baris memuat teks "N produk".
  const firstRow = page.locator('tbody tr').first()
  await expect(firstRow).toBeVisible({ timeout: 30_000 })
  await expect(firstRow).toContainText(/produk/)
  await expect(page).toHaveScreenshot('kategori-tab-desktop.png', { animations: 'disabled', maxDiffPixelRatio: 0.02 })

  // Buat kategori uji.
  await page.getByRole('button', { name: /Tambah Kategori/ }).click()
  await page.getByLabel(/Nama kategori baru/).fill(NAMA_AWAL)
  await page.getByRole('button', { name: 'Simpan', exact: true }).click()
  const newRow = page.locator('tbody tr', { hasText: NAMA_AWAL })
  await expect(newRow).toBeVisible({ timeout: 30_000 })
  await expect(newRow).toContainText('0 produk')

  // Ubah nama lewat aksi baris.
  await newRow.getByLabel(`Ubah nama kategori ${NAMA_AWAL}`).click()
  await page.getByLabel(/Nama kategori/).fill(NAMA_BARU)
  await page.getByRole('button', { name: 'Simpan Nama' }).click()
  const renamedRow = page.locator('tbody tr', { hasText: NAMA_BARU })
  await expect(renamedRow).toBeVisible({ timeout: 30_000 })

  // Hapus (dialog confirm diterima oleh handler di atas) -> baris hilang.
  await renamedRow.getByLabel(`Hapus kategori ${NAMA_BARU}`).click()
  await expect(page.locator('tbody tr', { hasText: NAMA_BARU })).toHaveCount(0, { timeout: 30_000 })
})

test('tab Kategori + Produk pada 390px: snapshot mobile + luapan 0', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await login(page)

  await page.goto('/app/products?tab=kategori')
  await tungguStabilKategori(page)
  await expect(page).toHaveScreenshot('kategori-tab-mobile.png', { animations: 'disabled', maxDiffPixelRatio: 0.02 })
  await assertNoOverflow(page, 'tab Kategori 390px')

  await page.getByRole('tab', { name: 'Produk' }).click()
  await expect(page.getByRole('tab', { name: 'Produk' })).toHaveAttribute('aria-selected', 'true')
  await page.waitForTimeout(800)
  await assertNoOverflow(page, 'tab Produk 390px')
})
