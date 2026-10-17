import { expect, test, type Page } from '@playwright/test'

// Alur E2E pemilih jam happy-hour Promo (pengganti <input type="time">
// natif) melawan situs dev: buat promo 22:00–02:00 (lintas tengah malam —
// SAH, tidak boleh ditukar/ditolak) lewat TimePicker bersama dari UI,
// simpan, buka lagi dan pastikan nilai tampil persis; kosongkan kembali
// lewat tombol X pemilih dan pastikan pasangan kosong tetap tersimpan;
// promo tanpa jendela jam sejak awal juga tetap bisa disimpan. Semua
// promo uji dibuat di sini dan WAJIB dihapus lagi di akhir test yang
// sama (nama tidak berawalan "AUDIT").
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

function fmtID(d: Date) {
  return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'short', year: 'numeric' })
}

// Isi tanggal lewat DatePicker bersama (klik hari pada kalender yang
// terbuka) — pola yang sama seperti pemakai nyata.
async function pickDate(page: Page, label: string, date: Date) {
  await page.getByRole('button', { name: label, exact: true }).click()
  const cal = page.getByRole('dialog', { name: 'Kalender' })
  await expect(cal).toBeVisible()
  // Klik lewat keyboard (fokus + Enter): panel kalender adalah popover
  // absolut di dalam kartu Modal yang overflow-y-auto — pada form panjang
  // bagian bawah panel terpotong tepi kartu sehingga klik mouse pada hari
  // di baris bawah terhalang pembungkus modal (terbukti lewat
  // elementFromPoint saat diagnosis). Tombol hari adalah <button> asli,
  // jadi jalur keyboard adalah jalur pemakai yang sah.
  const day = cal.getByRole('button', { name: fmtID(date), exact: true })
  await day.focus()
  await page.keyboard.press('Enter')
  await expect(cal).toBeHidden()
}

// Pilih jam lewat TimePicker bersama: buka panel dari pemicu berlabel
// persis, klik opsi pada listbox "— Jam" lalu "— Menit". Panel terpilih
// menutup sendiri begitu jam+menit lengkap.
async function pickTime(page: Page, label: string, hh: string, mm: string) {
  const trigger = page.getByRole('button', { name: label, exact: true })
  // Gulir pemicu ke tengah kartu modal dulu agar panel popover yang
  // membuka ke bawah muat di dalam kartu (tidak terpotong overflow).
  await trigger.evaluate((el) => el.scrollIntoView({ block: 'center' }))
  await trigger.click()
  const panel = page.getByRole('dialog', { name: label, exact: true })
  await expect(panel).toBeVisible()
  await panel.getByRole('listbox', { name: `${label} — Jam` }).getByRole('option', { name: hh, exact: true }).click()
  await panel.getByRole('listbox', { name: `${label} — Menit` }).getByRole('option', { name: mm, exact: true }).click()
  await expect(panel).toBeHidden()
  await expect(page.getByRole('button', { name: label, exact: true })).toContainText(`${hh}:${mm}`)
}

async function openCreateForm(page: Page) {
  await page.goto('/app/promo')
  await page.getByRole('button', { name: 'Tambah Promo', exact: true }).click()
  await expect(page.getByRole('dialog', { name: 'Tambah Promo' })).toBeVisible()
}

async function fillRequired(page: Page, name: string) {
  const now = new Date()
  const end = new Date(now)
  end.setDate(end.getDate() + 2)
  await page.getByLabel('Nama promo').fill(name)
  await page.getByLabel('Nilai persentase').fill('10')
  await pickDate(page, 'Mulai berlaku', now)
  await pickDate(page, 'Berakhir', end)
}

async function saveForm(page: Page) {
  await page.getByRole('button', { name: 'Simpan', exact: true }).click()
  await expect(page.getByRole('dialog', { name: /Tambah Promo|Ubah Promo/ })).toBeHidden({ timeout: 30_000 })
}

async function searchPromo(page: Page, name: string) {
  await page.goto('/app/promo')
  await page.getByPlaceholder('Ketik nama atau kode…').fill(name)
  await expect(page.getByRole('button', { name: `Ubah ${name}` })).toBeVisible({ timeout: 30_000 })
}

async function deletePromo(page: Page, name: string) {
  await searchPromo(page, name)
  await page.getByRole('button', { name: `Hapus ${name}` }).click()
  const confirm = page.getByRole('dialog', { name: new RegExp(`Hapus Promo`) })
  await expect(confirm).toBeVisible()
  await confirm.getByRole('button', { name: 'Hapus', exact: true }).click()
  await expect(confirm).toBeHidden({ timeout: 30_000 })
  await expect(page.getByRole('button', { name: `Ubah ${name}` })).toHaveCount(0, { timeout: 30_000 })
}

test('promo happy-hour 22:00–02:00 lewat pemilih jam baru, tersimpan persis, bisa dikosongkan', async ({ page }) => {
  const suffix = String(Date.now() % 100_000_000)
  const name = `UjiJamHH${suffix}`
  await login(page)

  // Buat promo dengan jendela lintas tengah malam lewat pemilih baru.
  await openCreateForm(page)
  await fillRequired(page, name)
  await pickTime(page, 'Jam mulai harian (opsional)', '22', '00')
  await pickTime(page, 'Jam berakhir harian (opsional)', '02', '00')
  await saveForm(page)

  // Buka lagi: nilai tampil persis 22:00 / 02:00 (tidak ditukar).
  await searchPromo(page, name)
  await page.getByRole('button', { name: `Ubah ${name}` }).click()
  await expect(page.getByRole('dialog', { name: 'Ubah Promo' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Jam mulai harian (opsional)', exact: true })).toContainText('22:00')
  await expect(page.getByRole('button', { name: 'Jam berakhir harian (opsional)', exact: true })).toContainText('02:00')

  // Kosongkan keduanya lewat tombol X pemilih, simpan lagi: pasangan
  // kosong tetap sah (tanpa jendela happy-hour).
  const clearButtons = page.getByRole('button', { name: 'Hapus jam' })
  await expect(clearButtons).toHaveCount(2)
  await clearButtons.first().click()
  await page.getByRole('button', { name: 'Hapus jam' }).click()
  await expect(page.getByRole('button', { name: 'Jam mulai harian (opsional)', exact: true })).toContainText('Pilih jam')
  await expect(page.getByRole('button', { name: 'Jam berakhir harian (opsional)', exact: true })).toContainText('Pilih jam')
  await saveForm(page)

  // Buka lagi: tetap kosong setelah tersimpan.
  await searchPromo(page, name)
  await page.getByRole('button', { name: `Ubah ${name}` }).click()
  await expect(page.getByRole('dialog', { name: 'Ubah Promo' })).toBeVisible()
  await expect(page.getByRole('button', { name: 'Jam mulai harian (opsional)', exact: true })).toContainText('Pilih jam')
  await expect(page.getByRole('button', { name: 'Jam berakhir harian (opsional)', exact: true })).toContainText('Pilih jam')
  await page.getByRole('button', { name: 'Batal', exact: true }).click()

  // Bersihkan data uji.
  await deletePromo(page, name)
})

test('promo tanpa jendela jam tetap bisa disimpan kosong', async ({ page }) => {
  const suffix = String(Date.now() % 100_000_000)
  const name = `UjiJamKosong${suffix}`
  await login(page)

  await openCreateForm(page)
  await fillRequired(page, name)
  // Kedua pemilih jam dibiarkan pada placeholder "Pilih jam".
  await expect(page.getByRole('button', { name: 'Jam mulai harian (opsional)', exact: true })).toContainText('Pilih jam')
  await expect(page.getByRole('button', { name: 'Jam berakhir harian (opsional)', exact: true })).toContainText('Pilih jam')
  await saveForm(page)

  await searchPromo(page, name)
  await deletePromo(page, name)
})

test('pemilih jam tampil sebagai sheet bawah pada 390px dan bisa memilih', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 })
  await login(page)

  await openCreateForm(page)
  await page.getByRole('button', { name: 'Jam mulai harian (opsional)', exact: true }).click()
  const panel = page.getByRole('dialog', { name: 'Jam mulai harian (opsional)', exact: true })
  await expect(panel).toBeVisible()
  // Sheet bawah: panel menempel tepi bawah viewport.
  const box = await panel.boundingBox()
  expect(box, 'bounding box panel pemilih jam').not.toBeNull()
  expect(Math.abs((box!.y + box!.height) - 844)).toBeLessThanOrEqual(2)
  await panel.getByRole('listbox', { name: 'Jam mulai harian (opsional) — Jam' }).getByRole('option', { name: '22', exact: true }).click()
  await panel.getByRole('listbox', { name: 'Jam mulai harian (opsional) — Menit' }).getByRole('option', { name: '00', exact: true }).click()
  await expect(panel).toBeHidden()
  await expect(page.getByRole('button', { name: 'Jam mulai harian (opsional)', exact: true })).toContainText('22:00')
  // Tanpa menyimpan: tidak ada data uji yang dibuat pada test ini.
  await page.getByRole('button', { name: 'Batal', exact: true }).click()
})
