import { expect, test, type Page } from '@playwright/test'

// Redirect URL lama (Bahasa Indonesia) -> URL kanonis Inggris.
// Setiap URL lama harus mendarat di padanan barunya dengan query string
// yang dipertahankan persis (LegacyRedirect di src/App.tsx meneruskan
// location.search). Redirect tab warisan (/app/opname, /app/shift,
// /app/karyawan) juga diverifikasi target barunya.
//
// Kredensial hanya dari env, tidak pernah ditulis ke file/log:
// - E2E_EMAIL / E2E_PASSWORD: <redacted>
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

async function expectRedirect(page: Page, dari: string, ke: string) {
  await page.goto(dari)
  await page.waitForURL((url) => url.pathname + url.search === ke, { timeout: 30_000 })
  const now = new URL(page.url())
  expect(now.pathname + now.search).toBe(ke)
}

test('redirect publik tanpa sesi: /masuk, /daftar, /unduh', async ({ page }) => {
  await expectRedirect(page, '/masuk', '/login')
  await expectRedirect(page, '/daftar', '/register')
  // Query string diteruskan persis.
  await expectRedirect(page, '/daftar?sumber=uji', '/register?sumber=uji')
  await expectRedirect(page, '/unduh', '/download')
})

test('redirect aplikasi sesudah login: semua URL lama + query + redirect tab warisan', async ({ page }) => {
  await login(page)

  // Publik yang hanya stabil dengan sesi aktif: pemilih akun tetap di
  // /select-account karena sesi admin + kasir aktif memang menampilkannya.
  await expectRedirect(page, '/pilih-akun', '/select-account')

  await expectRedirect(page, '/app/produk', '/app/products')
  await expectRedirect(page, '/app/produk?edit=1', '/app/products?edit=1')
  await expectRedirect(page, '/app/stok', '/app/stock')
  await expectRedirect(page, '/app/stok?tab=riwayat', '/app/stock?tab=movement')
  await expectRedirect(page, '/app/transaksi', '/app/transactions')
  await expectRedirect(page, '/app/pelanggan', '/app/customers')
  await expectRedirect(page, '/app/pembelian', '/app/purchases')
  await expectRedirect(page, '/app/pembelian?tab=pengeluaran', '/app/purchases?tab=expenses')
  await expectRedirect(page, '/app/hutang-piutang', '/app/debts')
  await expectRedirect(page, '/app/laporan', '/app/reports')
  await expectRedirect(page, '/app/laporan?tab=pajak', '/app/reports?tab=tax')
  await expectRedirect(page, '/app/pengaturan', '/app/settings')

  // Redirect tab warisan yang sudah ada sebelum rename (target diperbarui).
  await expectRedirect(page, '/app/opname', '/app/stock?tab=stocktake')
  await expectRedirect(page, '/app/shift', '/app/reports?tab=shift')
  await expectRedirect(page, '/app/karyawan', '/app/reports?tab=cashier')

  // Alias nilai ?tab= lama (Bahasa Indonesia): diterima, tab benar aktif,
  // dan URL dinormalkan ke nilai kanonis Inggris (replace, parameter lain utuh).
  await expectRedirect(page, '/app/reports?tab=penjualan', '/app/reports?tab=sales')
  await expectRedirect(page, '/app/reports?tab=produk', '/app/reports?tab=products')
  await expectRedirect(page, '/app/reports?tab=stok', '/app/reports?tab=stock')
  await expectRedirect(page, '/app/reports?tab=kasir', '/app/reports?tab=cashier')
  await expectRedirect(page, '/app/reports?tab=pajak&sumber=uji', '/app/reports?tab=tax&sumber=uji')
  await expectRedirect(page, '/app/reports?tab=cashier', '/app/reports?tab=cashier')
  await expectRedirect(page, '/app/stock?tab=opname', '/app/stock?tab=stocktake')
  await expectRedirect(page, '/app/stock?tab=riwayat', '/app/stock?tab=movement')
  await expectRedirect(page, '/app/stock?tab=movement', '/app/stock?tab=movement')
  await expectRedirect(page, '/app/purchases?tab=po', '/app/purchases?tab=purchases')
  await expectRedirect(page, '/app/purchases?tab=pembelian', '/app/purchases?tab=purchases')
  await expectRedirect(page, '/app/purchases?tab=pengeluaran', '/app/purchases?tab=expenses')
  await expectRedirect(page, '/app/purchases?tab=expenses', '/app/purchases?tab=expenses')
})
