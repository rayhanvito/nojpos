import { expect, test, type Page } from '@playwright/test'
import { readFileSync } from 'node:fs'
import { tmpdir } from 'node:os'
import { join } from 'node:path'

// Uji tab Kategori di halaman Produk & Kategori.
// - Unit awal (2026-10-19): ?tab=kategori membuka tab Kategori; CRUD
//   kategori uji lewat UI (buat -> "0 produk" -> ubah nama -> hapus).
// - Amendment 1: tombol "Tambah Kategori" kini di header halaman, slot
//   yang sama dengan "Tambah Produk" — tingginya diukur dan dibandingkan
//   pada viewport 390px.
// - Amendment 2: toolbar tab Kategori (cari + Select urutan + FilterPanel
//   isi), ekspor CSV kategori (isi unduhan diperiksa), impor CSV kategori
//   (berkas uji kecil: 2 valid, 1 duplikat lama, 1 duplikat dalam berkas
//   -> 2 berhasil / 2 gagal), data uji dibersihkan di akhir.
// - Luapan horizontal 390px = 0 di kedua tab; snapshot tab Kategori
//   mencakup toolbar lengkap.
// Kredensial hanya dari env (E2E_EMAIL/E2E_PASSWORD/E2E_PASSCODE), sama
// seperti visual.spec.ts; nilai tidak pernah ditulis ke file/log.
// Nama data uji sengaja TIDAK berawalan huruf awal alfabet.

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

async function rowCounts(page: Page): Promise<number[]> {
  const texts = await page.locator('tbody tr').allTextContents()
  return texts.map((t) => {
    const m = t.match(/(\d+)\s*produk/)
    return m ? parseInt(m[1], 10) : -1
  })
}

async function hapusKategori(page: Page, nama: string) {
  const row = page.locator('tbody tr', { hasText: nama })
  await row.getByLabel(`Hapus kategori ${nama}`).click()
  await expect(page.locator('tbody tr', { hasText: nama })).toHaveCount(0, { timeout: 30_000 })
}

const NAMA_AWAL = 'Zeta Uji Kategori'
const NAMA_BARU = 'Zeta Uji Rubah Nama'

test('tab Kategori: toolbar + CRUD + ekspor + impor', async ({ page }) => {
  page.on('dialog', (d) => void d.accept())
  await page.setViewportSize({ width: 1440, height: 900 })
  await login(page)

  await page.goto('/app/products?tab=kategori')
  await tungguStabilKategori(page)
  expect(new URL(page.url()).searchParams.get('tab')).toBe('categories')
  await expect(page).toHaveScreenshot('kategori-tab-desktop.png', { animations: 'disabled', maxDiffPixelRatio: 0.02 })

  // --- Toolbar lengkap terlihat ---
  await expect(page.getByLabel('Cari kategori')).toBeVisible()
  await expect(page.getByLabel('Urutkan kategori')).toBeVisible()
  await expect(page.getByLabel('Filter lanjutan')).toBeVisible()

  const totalRows = await page.locator('tbody tr').count()
  expect(totalRows).toBeGreaterThan(2)

  // --- Search client-side ---
  await page.getByLabel('Cari kategori').fill('Minum')
  await expect(page.locator('tbody tr', { hasText: 'Minuman' })).toBeVisible({ timeout: 10_000 })
  await expect(page.locator('tbody tr', { hasText: 'Sembako' })).toHaveCount(0)
  await page.getByLabel('Cari kategori').fill('')
  await expect(page.locator('tbody tr')).toHaveCount(totalRows, { timeout: 10_000 })

  // --- Urutan: Produk terbanyak -> hitungan menurun, Minuman di atas ---
  await page.getByLabel('Urutkan kategori').click()
  await page.getByRole('option', { name: 'Produk terbanyak' }).click()
  await page.waitForTimeout(300)
  const desc = await rowCounts(page)
  expect(desc.length).toBe(totalRows)
  expect(desc).toEqual([...desc].sort((a, b) => b - a))
  await expect(page.locator('tbody tr').first()).toContainText('Minuman')
  await page.getByLabel('Urutkan kategori').click()
  await page.getByRole('option', { name: 'Nama A–Z' }).click()

  // --- Filter isi: Kosong -> semua baris 0 produk + chip; chip dihapus -> pulih ---
  await page.getByLabel('Filter lanjutan').click()
  await page.getByLabel('Filter isi kategori').click()
  await page.getByRole('option', { name: 'Kosong' }).click()
  await page.keyboard.press('Escape')
  await page.waitForTimeout(300)
  const kosong = await rowCounts(page)
  expect(kosong.length).toBeGreaterThanOrEqual(1)
  expect(kosong.length).toBeLessThan(totalRows)
  expect(kosong.every((n) => n === 0)).toBe(true)
  await expect(page.getByText('Isi: Kosong')).toBeVisible()
  await page.getByLabel('Hapus filter Isi: Kosong').click()
  await expect(page.locator('tbody tr')).toHaveCount(totalRows, { timeout: 10_000 })

  // --- Ekspor CSV: isi unduhan cocok dengan daftar ---
  const downloadPromise = page.waitForEvent('download', { timeout: 30_000 })
  await page.getByRole('button', { name: 'Export' }).click()
  await page.getByRole('menuitem', { name: 'CSV (.csv)' }).click()
  const download = await downloadPromise
  const csvPath = join(tmpdir(), 'ekspor-kategori-uji.csv')
  await download.saveAs(csvPath)
  const csv = readFileSync(csvPath, 'utf-8').replace(/^﻿/, '')
  const lines = csv.trim().split(/\r?\n/).map((l) => l.replaceAll('"', ''))
  expect(lines[0]).toBe('nama,jumlah produk')
  expect(lines.length).toBe(totalRows + 1)
  expect(lines.some((l) => l.startsWith('Minuman,'))).toBe(true)

  // --- CRUD kategori uji (form modal halaman, pola sama seperti dulu) ---
  await page.getByRole('button', { name: /Tambah Kategori/ }).click()
  await page.getByLabel(/Nama kategori baru/).fill(NAMA_AWAL)
  await page.getByRole('button', { name: 'Simpan', exact: true }).click()
  const newRow = page.locator('tbody tr', { hasText: NAMA_AWAL })
  await expect(newRow).toBeVisible({ timeout: 30_000 })
  await expect(newRow).toContainText('0 produk')

  await newRow.getByLabel(`Ubah nama kategori ${NAMA_AWAL}`).click()
  await page.getByLabel(/Nama kategori/).fill(NAMA_BARU)
  await page.getByRole('button', { name: 'Simpan Nama' }).click()
  await expect(page.locator('tbody tr', { hasText: NAMA_BARU })).toBeVisible({ timeout: 30_000 })
  await hapusKategori(page, NAMA_BARU)

  // --- Impor CSV: 2 valid + 1 duplikat lama + 1 duplikat dalam berkas ---
  await page.getByRole('button', { name: 'Import' }).click()
  await page.getByRole('menuitem', { name: 'CSV (.csv)' }).click()
  await expect(page.getByRole('dialog', { name: 'Import Kategori' })).toBeVisible({ timeout: 10_000 })
  const chooserPromise = page.waitForEvent('filechooser', { timeout: 10_000 })
  await page.getByRole('button', { name: 'Pilih File CSV' }).click()
  const chooser = await chooserPromise
  await chooser.setFiles({
    name: 'impor-kategori-uji.csv',
    mimeType: 'text/csv',
    buffer: Buffer.from('nama\nZeta Impor Satu\nZeta Impor Dua\nMinuman\nZeta Impor Satu\n', 'utf-8'),
  })
  await expect(page.getByText(/2 berhasil/)).toBeVisible({ timeout: 30_000 })
  await expect(page.getByText(/2 gagal/)).toBeVisible()
  await page.getByRole('button', { name: 'Selesai' }).click()
  await expect(page.locator('tbody tr', { hasText: 'Zeta Impor Satu' })).toBeVisible({ timeout: 30_000 })
  await expect(page.locator('tbody tr', { hasText: 'Zeta Impor Dua' })).toContainText('0 produk')

  // --- Bersihkan data impor uji ---
  await hapusKategori(page, 'Zeta Impor Satu')
  await hapusKategori(page, 'Zeta Impor Dua')
  await expect(page.locator('tbody tr')).toHaveCount(totalRows, { timeout: 10_000 })
})

test('tab Kategori 390px: tombol tambah konsisten + snapshot + luapan 0', async ({ page }) => {
  page.on('dialog', (d) => void d.accept())
  await page.setViewportSize({ width: 390, height: 844 })
  await login(page)

  // Acuan Amendment 1: tombol tambah utama tab Produk.
  await page.goto('/app/products')
  await expect(page.getByRole('tab', { name: 'Produk' })).toHaveAttribute('aria-selected', 'true', { timeout: 30_000 })
  const btnProduk = page.getByRole('button', { name: /Tambah Produk/ })
  await expect(btnProduk).toBeVisible({ timeout: 30_000 })
  const boxProduk = await btnProduk.boundingBox()
  expect(boxProduk).not.toBeNull()
  expect(boxProduk!.height).toBeGreaterThanOrEqual(41)
  expect(boxProduk!.height).toBeLessThanOrEqual(43)
  await assertNoOverflow(page, 'tab Produk 390px')

  await page.goto('/app/products?tab=kategori')
  await tungguStabilKategori(page)
  const btnKategori = page.getByRole('button', { name: /Tambah Kategori/ })
  await expect(btnKategori).toBeVisible({ timeout: 30_000 })
  const boxKategori = await btnKategori.boundingBox()
  expect(boxKategori).not.toBeNull()
  expect(boxKategori!.height).toBeGreaterThanOrEqual(41)
  expect(boxKategori!.height).toBeLessThanOrEqual(43)
  expect(Math.abs(boxKategori!.height - boxProduk!.height)).toBeLessThanOrEqual(1)
  await expect(page).toHaveScreenshot('kategori-tab-mobile.png', { animations: 'disabled', maxDiffPixelRatio: 0.02 })
  await assertNoOverflow(page, 'tab Kategori 390px')
})
