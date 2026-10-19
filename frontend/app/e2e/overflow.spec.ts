import { expect, test, type Page } from '@playwright/test'

// Sweep overflow horizontal SEMUA halaman aplikasi (FINAL-A).
//
// Mandat: tidak boleh ada overflow horizontal di halaman mana pun, pada
// viewport desktop 1440x900 maupun mobile 390x844 (standar 390px Rayhan).
// Test ini SATU test yang menyapu semua rute dan mengumpulkan SEMUA
// pelanggaran ke satu laporan assertion di akhir — tidak berhenti di
// pelanggaran pertama. Bila ada temuan, itu dilaporkan ke koordinator;
// test ini TIDAK memperbaiki CSS/halaman apa pun.
//
// Daftar rute diekstrak dari src/App.tsx (rute di dalam <Route
// path="/app" element={<AppShell />}> + rute publik /download yang diminta
// mandat):
//   /app                 Beranda (index)
//   /app/pos             Kasir
//   /app/products          Produk & Kategori (AdminOnly)
//   /app/transactions       Transaksi
//   /app/customers       Pelanggan
//   /app/promo           Promo & Voucher (AdminOnly)
//   /app/purchases       Pembelian (AdminOnly)
//   /app/purchases?tab=expenses  Pengeluaran (tab di Pembelian —
//                        bukan rute sendiri di App.tsx)
//   /app/supplier        Supplier (AdminOnly — rute sendiri)
//   /app/stock            Stok (AdminOnly)
//   /app/opname          redirect App.tsx -> /app/stock?tab=stocktake
//   /app/debts  Hutang Piutang (AdminOnly)
//   /app/reports         Laporan (AdminOnly)
//   /app/shift           redirect App.tsx -> /app/reports?tab=shift
//   /app/karyawan        redirect App.tsx -> /app/reports?tab=cashier
//   /app/users           Pengguna (AdminOnly)
//   /app/settings      Pengaturan (AdminOnly)
//   /app/settings (tab Outlet)  Pengaturan Outlet — tab internal
//                        SettingsPage, bukan rute/query sendiri, jadi
//                        dicapai dengan klik tab "Outlet".
//   /download               Unduh (rute publik di luar AppShell)
//
// Kredensial hanya dari env, tidak pernah ditulis ke file/log:
// - E2E_EMAIL / E2E_PASSWORD: akun bootstrap.
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

type Rute = {
  slug: string
  path: string
  // Pathname akhir setelah redirect App.tsx (default: pathname dari path).
  expectPath?: string
  // Tab internal yang diklik setelah halaman stabil (Pengaturan Outlet).
  tab?: string
}

const RUTES: Rute[] = [
  { slug: 'beranda', path: '/app' },
  { slug: 'kasir', path: '/app/pos' },
  { slug: 'produk', path: '/app/products' },
  { slug: 'transaksi', path: '/app/transactions' },
  { slug: 'pelanggan', path: '/app/customers' },
  { slug: 'promo', path: '/app/promo' },
  { slug: 'pembelian', path: '/app/purchases' },
  { slug: 'pengeluaran', path: '/app/purchases?tab=expenses', expectPath: '/app/purchases' },
  { slug: 'supplier', path: '/app/supplier' },
  { slug: 'stok', path: '/app/stock' },
  { slug: 'opname', path: '/app/opname', expectPath: '/app/stock' },
  { slug: 'hutang-piutang', path: '/app/debts' },
  { slug: 'laporan', path: '/app/reports' },
  { slug: 'shift', path: '/app/shift', expectPath: '/app/reports' },
  { slug: 'karyawan', path: '/app/karyawan', expectPath: '/app/reports' },
  { slug: 'pengguna', path: '/app/users' },
  { slug: 'pengaturan', path: '/app/settings' },
  { slug: 'pengaturan-outlet', path: '/app/settings', tab: 'Outlet' },
]

const VIEWPORTS: { label: string; width: number; height: number }[] = [
  { label: 'desktop', width: 1440, height: 900 },
  { label: 'mobile', width: 390, height: 844 },
]

async function tungguStabil(page: Page, rute: Rute) {
  const target = rute.expectPath ?? new URL(rute.path, 'https://x.local').pathname
  await page.waitForURL((url) => url.pathname === target, { timeout: 30_000 })
  // Konten utama ter-render: <main> tampil dan tidak kosong.
  const main = page.locator('main').first()
  await expect(main).toBeVisible({ timeout: 30_000 })
  await expect
    .poll(async () => (await main.innerText().catch(() => '')).trim().length, { timeout: 30_000 })
    .toBeGreaterThan(0)
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
  await page.waitForTimeout(600)
}

type HasilUkur = {
  innerWidth: number
  docScrollWidth: number
  bodyScrollWidth: number
  pelanggar: string[]
}

async function ukurOverflow(page: Page): Promise<HasilUkur> {
  return page.evaluate(() => {
    const iw = window.innerWidth
    const pelanggar: string[] = []
    const terlihat = new Set<string>()

    // Elemen "terpotong rapi" oleh leluhur yang memang bisa digeser /
    // memotong sumbu-x (tablist overflow-x-auto, TableWrap, dsb.) bukan
    // overflow halaman — lewati bila leluhur terdekat yang memotong
    // sendiri berada di dalam viewport.
    function terpotongLeluhur(el: Element): boolean {
      let p = el.parentElement
      while (p && p !== document.body && p !== document.documentElement) {
        const cs = getComputedStyle(p)
        if (cs.overflowX === 'auto' || cs.overflowX === 'scroll' || cs.overflowX === 'hidden') {
          const pr = p.getBoundingClientRect()
          return pr.right <= iw + 1 && pr.left >= -1
        }
        p = p.parentElement
      }
      return false
    }

    document.querySelectorAll('*').forEach((el) => {
      const r = el.getBoundingClientRect()
      if (r.width === 0 || r.height === 0) return
      if (r.right <= iw + 1) return
      if (r.left < -1) return // elemen off-canvas ke kiri (drawer dsb.)
      if (terpotongLeluhur(el)) return
      const cs = getComputedStyle(el)
      if (cs.position === 'fixed' && cs.visibility === 'hidden') return
      const cls = (el.getAttribute('class') ?? '').slice(0, 80)
      const teks = (el.textContent ?? '').trim().replace(/\s+/g, ' ').slice(0, 40)
      const kunci = `${el.tagName}.${cls}`
      if (terlihat.has(kunci)) return
      terlihat.add(kunci)
      pelanggar.push(
        `<${el.tagName.toLowerCase()} class="${cls}"> right=${Math.round(r.right)} ` +
          `width=${Math.round(r.width)} teks="${teks}"`,
      )
    })

    return {
      innerWidth: iw,
      docScrollWidth: document.documentElement.scrollWidth,
      bodyScrollWidth: document.body?.scrollWidth ?? 0,
      pelanggar: pelanggar.slice(0, 5),
    }
  })
}

test('sweep overflow semua halaman desktop + 390px', async ({ page }) => {
  test.setTimeout(1_200_000)
  await login(page)

  const pelanggaran: string[] = []

  for (const vp of VIEWPORTS) {
    await page.setViewportSize({ width: vp.width, height: vp.height })
    for (const rute of RUTES) {
      const label = `${rute.slug} [${rute.path}] @ ${vp.label} ${vp.width}x${vp.height}`
      try {
        await page.goto(rute.path)
        await tungguStabil(page, rute)
        if (rute.tab) {
          await page.getByRole('tab', { name: rute.tab, exact: true }).click()
          await page.waitForLoadState('networkidle', { timeout: 30_000 }).catch(() => {})
          await expect
            .poll(
              async () =>
                (await page.locator('[data-slot="skeleton"]').count()) +
                (await page.locator('[aria-busy="true"]').count()),
              { timeout: 30_000 },
            )
            .toBe(0)
          await page.waitForTimeout(600)
        }
        const h = await ukurOverflow(page)
        const masalah: string[] = []
        if (h.docScrollWidth > h.innerWidth + 1) {
          masalah.push(`documentElement.scrollWidth=${h.docScrollWidth} > innerWidth=${h.innerWidth}`)
        }
        if (h.bodyScrollWidth > h.innerWidth + 1) {
          masalah.push(`body.scrollWidth=${h.bodyScrollWidth} > innerWidth=${h.innerWidth}`)
        }
        if (h.pelanggar.length > 0) {
          masalah.push(`elemen melewati tepi kanan: ${h.pelanggar.join(' | ')}`)
        }
        if (masalah.length > 0) pelanggaran.push(`${label}: ${masalah.join('; ')}`)
      } catch (e) {
        pelanggaran.push(`${label}: GAGAL DIUKUR — ${e instanceof Error ? e.message.split('\n')[0] : String(e)}`)
      }
    }
  }

  expect(
    pelanggaran,
    pelanggaran.length === 0
      ? 'tidak ada overflow'
      : `Ditemukan ${pelanggaran.length} pelanggaran overflow:\n${pelanggaran.join('\n')}`,
  ).toEqual([])
})
