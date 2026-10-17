import { expect, test, type Page } from '@playwright/test'

// Alur E2E checkout tunai melawan situs dev (data demo). Satu transaksi
// nyata memang tercipta oleh uji ini — itu tujuannya (gerbang FE-C1,
// sebelum halaman Kasir dipecah di FE-C3).
//
// Kredensial hanya dari env, tidak pernah ditulis ke file/log:
// - E2E_EMAIL / E2E_PASSWORD: akun pemilik (bootstrap).
// - E2E_PASSCODE: passcode pemilik untuk langkah Pilih Akun (akun pemilik
//   di dev dilindungi passcode; wajib hanya bila langkah itu muncul).

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

test('checkout tunai end-to-end', async ({ page }) => {
  await login(page)

  await page.getByRole('link', { name: 'Kasir', exact: true }).click()
  await page.waitForURL('**/app/pos')
  await expect(page.getByLabel('Cari produk')).toBeVisible()

  await ensureShiftAktif(page)

  // Produk pertama yang bisa dijual: tombol katalog non-disabled yang
  // memuat stok + harga (tombol kategori tidak memuat keduanya).
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

  // Item masuk keranjang (panel aside di layout desktop).
  const cartPanel = page.locator('aside').filter({ hasText: /keranjang/i })
  await expect(cartPanel.getByText(productName).first()).toBeVisible()

  // Tombol Bayar aktif setelah pratinjau promo selesai.
  const bayarBtn = cartPanel.getByRole('button', { name: /Bayar/ })
  await expect(bayarBtn).toBeEnabled({ timeout: 30_000 })
  await bayarBtn.click()

  const payDialog = page.getByRole('dialog', { name: 'Pembayaran' })
  await expect(payDialog).toBeVisible()
  await payDialog.getByRole('button', { name: 'Cash', exact: true }).click()
  // Uang Pas: jumlah dibayar = total persis (alur tunai tanpa kembalian).
  await payDialog.getByRole('checkbox', { name: /Uang Pas/ }).check()
  await payDialog.getByRole('button', { name: 'Selesaikan Transaksi' }).click()

  // Sukses menurut render aplikasi: modal struk "Transaksi berhasil"
  // dengan nomor transaksi server.
  const receiptDialog = page.getByRole('dialog', { name: 'Transaksi berhasil' })
  await expect(receiptDialog).toBeVisible({ timeout: 30_000 })
  const receiptText = await receiptDialog.innerText()
  const trxNo = receiptText.match(/#TRX-\d+/)?.[0]
  expect(trxNo, 'nomor transaksi pada struk').toBeTruthy()
  console.log(`Transaksi tercipta: ${trxNo}`)

  // Tutup struk → keranjang kembali kosong.
  await receiptDialog.locator('button').filter({ hasText: /^Tutup$/ }).click()
  await expect(receiptDialog).toBeHidden()
  await expect(cartPanel.getByText('Belum ada item. Klik produk untuk menambah.')).toBeVisible()
})
