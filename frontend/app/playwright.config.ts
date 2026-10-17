import { defineConfig } from '@playwright/test'

// Kredensial E2E TIDAK PERNAH ditulis ke repo: wajib lewat env.
// Gagal tegas bila kosong agar tidak ada run yang "hijau" tanpa login nyata.
if (!process.env.E2E_EMAIL || !process.env.E2E_PASSWORD) {
  throw new Error(
    'E2E_EMAIL dan E2E_PASSWORD wajib diisi sebelum menjalankan e2e. ' +
      'Sumber lokal: baris email=/password= di /opt/pos-pakesigap/bootstrap-account.txt (jangan commit nilainya).',
  )
}

export default defineConfig({
  testDir: './e2e',
  timeout: 120_000,
  expect: { timeout: 15_000 },
  workers: 1,
  retries: 0,
  reporter: 'list',
  use: {
    baseURL: 'https://pos.pakesigap.com',
    trace: 'retain-on-failure',
    // Layout desktop: panel keranjang Kasir tampil sebagai aside (di mobile
    // ia berupa sheet) — viewport tetap agar alur yang diuji deterministik.
    viewport: { width: 1440, height: 900 },
    actionTimeout: 20_000,
    navigationTimeout: 60_000,
    // SW aplikasi bisa menyajikan aset basi antar-deploy; uji harus selalu
    // melawan build yang benar-benar live.
    serviceWorkers: 'block',
  },
})
