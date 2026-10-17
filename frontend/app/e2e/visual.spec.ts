import { expect, test, type Page } from '@playwright/test'

// Snapshot visual baseline halaman kunci melawan situs dev (FE-C8).
// Bukan uji perilaku: tujuannya menangkap regresi tampilan kasar pada
// restrukturisasi FE-C berikutnya. Baseline tersimpan di
// e2e/visual.spec.ts-snapshots/ dan ikut ter-commit; perbarui hanya
// dengan sengaja lewat `npx playwright test e2e/visual.spec.ts
// --update-snapshots` lalu periksa diff-nya.
//
// Pendekatan viewport (pilihan yang lebih sederhana & konsisten untuk
// config yang ada): SATU proyek bawaan playwright.config.ts, viewport
// diganti per fase di dalam test (1440x900 lalu 390x844) — makanya nama
// snapshot memuat label viewport-nya sendiri.
//
// Stabilitas: sebelum potret, halaman ditunggu sampai (1) URL cocok,
// (2) judul h1 halaman tampil, (3) tidak ada lagi indikator loading
// ([data-slot="skeleton"] / [aria-busy="true"]), (4) networkidle, dan
// (5) jeda singkat agar animasi grafik (Recharts, JS) selesai — opsi
// animations:'disabled' menangani animasi CSS. Toleransi
// maxDiffPixelRatio 0.02 menyerap perbedaan render sub-piksel/font.
//
// Mask: TIDAK ada pada baseline awal — tidak ada elemen yang terbukti
// volatil antar-run saat baseline dibuat (tanggal yang tampil adalah
// tanggal hari berjalan yang sama untuk generate + verifikasi; jam
// transaksi pada daftar adalah data historis). Bila kelak ada elemen
// yang terbukti volatil, tambahkan mask seperlunya di PAGES/entry-nya
// dan catat alasannya di sini.
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

const PAGES: { slug: string; path: string; title: string }[] = [
  { slug: 'beranda', path: '/app', title: 'Beranda' },
  { slug: 'kasir', path: '/app/pos', title: 'Kasir' },
  { slug: 'produk', path: '/app/products', title: 'Produk & Kategori' },
  { slug: 'transaksi', path: '/app/transactions', title: 'Transaksi' },
  { slug: 'pelanggan', path: '/app/customers', title: 'Pelanggan' },
  { slug: 'promo', path: '/app/promo', title: 'Promo & Voucher' },
  { slug: 'laporan', path: '/app/reports', title: 'Laporan' },
  { slug: 'pengaturan', path: '/app/settings', title: 'Pengaturan' },
]

const VIEWPORTS: { label: string; width: number; height: number }[] = [
  { label: 'desktop', width: 1440, height: 900 },
  { label: 'mobile', width: 390, height: 844 },
]

async function tungguStabil(page: Page, target: (typeof PAGES)[number]) {
  await page.waitForURL((url) => url.pathname === target.path, { timeout: 30_000 })
  await expect(page.getByRole('heading', { name: target.title, exact: true }).first()).toBeVisible({
    timeout: 30_000,
  })
  // Indikator loading hilang: skeleton shadcn + penanda aria-busy halaman.
  await expect
    .poll(
      async () =>
        (await page.locator('[data-slot="skeleton"]').count()) +
        (await page.locator('[aria-busy="true"]').count()),
      { timeout: 30_000 },
    )
    .toBe(0)
  await page.waitForLoadState('networkidle', { timeout: 30_000 }).catch(() => {})
  // Jeda tetap agar animasi grafik JS (Recharts) tuntas sebelum potret.
  await page.waitForTimeout(1_200)
}

test('snapshot visual halaman kunci desktop + 390px', async ({ page }) => {
  await login(page)

  for (const vp of VIEWPORTS) {
    await page.setViewportSize({ width: vp.width, height: vp.height })
    for (const target of PAGES) {
      await page.goto(target.path)
      await tungguStabil(page, target)
      await expect(page).toHaveScreenshot(`${target.slug}-${vp.label}.png`, {
        maxDiffPixelRatio: 0.02,
        animations: 'disabled',
      })
    }
  }
})
