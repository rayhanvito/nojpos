import { expect, test, type Page } from '@playwright/test'

// Alur E2E shift penuh melawan situs dev: shift aktif -> tutup shift lewat
// popup Status -> terbukti tertutup -> buka lagi (keadaan lingkungan
// dipulihkan seperti semula: shift aktif untuk spec/urutan berikutnya).
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

test('alur shift: tutup shift lalu buka lagi sampai aktif', async ({ page }) => {
  await login(page)
  await ensureShiftAktif(page)

  const statusBtn = page.getByRole('button', { name: /Status:/ })
  const dialog = page.getByRole('dialog', { name: 'Status' })

  // --- Tutup shift lewat popup Status ---
  await statusBtn.click()
  await expect(dialog).toBeVisible()
  await expect(dialog.getByRole('button', { name: 'Tutup Shift' })).toBeVisible({ timeout: 30_000 })

  // Uang fisik yang diisi = kas yang diharapkan menurut rincian shift yang
  // dirender dialog ("Kas yang diharapkan Rp…"); cadangan 100000 bila pola
  // teksnya tidak terbaca.
  const dialogText = await dialog.innerText()
  const expectedMatch = dialogText.match(/Kas yang diharapkan\s*Rp\s*([\d.]+)/)
  const closingCash = expectedMatch ? expectedMatch[1]!.replace(/\D/g, '') : '100000'
  expect(closingCash, 'uang fisik untuk tutup shift').toBeTruthy()

  await dialog.getByLabel(/Uang fisik di laci/).fill(closingCash)
  const tutupBtn = dialog.getByRole('button', { name: 'Tutup Shift' })
  await expect(tutupBtn).toBeEnabled({ timeout: 30_000 })
  await tutupBtn.click()

  // Berhasil tutup menurut render aplikasi: ringkasan "Shift ditutup."
  // dengan selisih kas, lalu tombol Selesai menutup dialog.
  await expect(dialog.getByText('Shift ditutup.')).toBeVisible({ timeout: 30_000 })
  await expect(dialog.getByText('Selisih kas')).toBeVisible()
  await dialog.getByRole('button', { name: 'Selesai', exact: true }).click()
  await expect(dialog).toBeHidden()

  // Keadaan tertutup yang dirender aplikasi: pil Status menawarkan buka lagi.
  await expect
    .poll(async () => (await statusBtn.innerText()).trim(), { timeout: 30_000 })
    .toContain('Buka Shift')

  // --- Buka lagi: pulihkan keadaan lingkungan (shift aktif) ---
  await statusBtn.click()
  await expect(dialog).toBeVisible()
  await dialog.getByLabel(/Modal awal di laci/).fill('100000')
  await dialog.getByRole('button', { name: 'Buka Shift', exact: true }).click()
  await expect(dialog.getByRole('button', { name: 'Tutup Shift' })).toBeVisible({ timeout: 30_000 })
  await dialog.locator('button[aria-label="Tutup"]').click()
  await expect(dialog).toBeHidden()
  await expect
    .poll(async () => (await statusBtn.innerText()).trim(), { timeout: 30_000 })
    .toContain('Shift aktif')
})
