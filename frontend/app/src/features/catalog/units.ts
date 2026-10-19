// Satuan baku Indonesia + penggabungan dengan satuan yang sedang
// dipakai toko. Daftar satuan mengelola dirinya sendiri: satuan baku
// selalu tersedia; satuan lain muncul karena dipakai produk (dari
// GET /products/units) dan hilang sendiri saat tidak lagi dipakai —
// tidak ada tabel master dan tidak ada halaman kelola.
export const BUILTIN_UNITS: string[] = [
  'pcs', 'pack', 'box', 'dus', 'karton',
  'botol', 'kaleng', 'bungkus', 'sachet', 'renteng',
  'slop', 'lusin', 'kodi', 'ikat', 'karung',
  'galon', 'jeriken', 'tube', 'roll', 'lembar',
  'set', 'pasang', 'kg', 'gram', 'ons', 'liter', 'ml',
]

// Gabung satuan baku (urutan kurasi) + satuan terpakai khas toko
// (abjad). Duplikat dilipat tanpa membedakan huruf besar/kecil;
// ejaan yang dipertahankan adalah yang pertama masuk (baku
// didahulukan). Nilai yang sedang diketik SENGAJA tidak dimasukkan:
// satuan produk yang ada selalu sudah termasuk lewat daftar
// terpakai, dan teks baru harus terbaca sebagai "belum ada" agar
// baris Tambah muncul.
export function mergeUnits(used: string[]): string[] {
  const seen = new Set<string>()
  const out: string[] = []
  const push = (u: string) => {
    const v = (u ?? '').trim()
    if (!v) return
    const key = v.toLowerCase()
    if (seen.has(key)) return
    seen.add(key)
    out.push(v)
  }
  BUILTIN_UNITS.forEach(push)
  ;[...used].sort((a, b) => a.localeCompare(b)).forEach(push)
  return out
}
