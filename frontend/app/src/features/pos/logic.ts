// Logika murni kasir (fitur pos): operasi keranjang + perhitungan total
// yang sebelumnya inline di halaman Kasir (pages/Pos). Semua fungsi di
// sini tanpa efek samping dan tanpa mengubah rumus, urutan, maupun
// pembulatan dari kode aslinya — ekstraksi murni, perilaku checkout
// tidak berubah sedikit pun.
//
// File ini MELENGKAPI logika bersama yang sudah ada, bukan menyalinnya:
// harga satuan bertingkat tetap dari shared/price-tier (getUnitPrice),
// validasi bayar tetap dari shared/checkout, antrean offline tetap dari
// shared/offline-queue.
import { getUnitPrice } from '../../shared/price-tier.ts'
import type { CheckoutBody, CheckoutItem, Customer, PayMethod, Product, PromoPreview, StoreSettings, Variant } from '../../shared/api/types.ts'

export interface CartLine { product: Product; variant?: Variant; qty: number }

export function lineKey(product: Product, variant?: Variant): string {
  return `${product.id}:${variant?.id ?? ''}`
}

export function lineQuantity(cart: CartLine[], product: Product, variant?: Variant): number {
  return cart.find((line) => lineKey(line.product, line.variant) === lineKey(product, variant))?.qty ?? 0
}

export function availableStock(cart: CartLine[], product: Product, variant?: Variant): number {
  return (variant?.stock ?? product.stock) - lineQuantity(cart, product, variant)
}

export function effectiveStock(cart: CartLine[], product: Product): number {
  return product.variants?.length
    ? product.variants.reduce((sum, variant) => sum + Math.max(0, availableStock(cart, product, variant)), 0)
    : availableStock(cart, product)
}

export function unitPrice(line: CartLine): number {
  return getUnitPrice(
    line.qty,
    (line.variant ?? line.product).sell_price,
    (line.variant ?? line.product).price_tiers ?? [],
  )
}

export function lineTotal(line: CartLine): number {
  return unitPrice(line) * line.qty
}

export function cartSubtotal(cart: CartLine[]): number {
  return cart.reduce((n, line) => n + unitPrice(line) * line.qty, 0)
}

export function cartItemCount(cart: CartLine[]): number {
  return cart.reduce((n, l) => n + l.qty, 0)
}

export function cartKey(cart: CartLine[]): string {
  return JSON.stringify(cart.map((line) => ({ productId: line.product.id, variantId: line.variant?.id, qty: line.qty })))
}

export function checkoutItems(cart: CartLine[]): CheckoutItem[] {
  return cart.map((line) => ({ productId: line.product.id, ...(line.variant ? { variantId: line.variant.id } : {}), qty: line.qty }))
}

export function cartSummaryItems(cart: CartLine[]): { name: string; qty: number }[] {
  return cart.map((line) => ({ name: `${line.product.name}${line.variant ? ` · ${line.variant.name}` : ''}`, qty: line.qty }))
}

export interface CartOpResult {
  cart: CartLine[]
  blocked: boolean
}

// Tambah satu item: baris produk+varian yang sama digabung (qty + 1).
// Saat online, penambahan yang melewati stok ditolak (blocked) dan
// keranjang dikembalikan apa adanya — sama seperti penjaga stok inline
// sebelumnya; saat offline tidak ada penjagaan stok sama sekali.
export function addToCart(cart: CartLine[], product: Product, variant: Variant | undefined, online: boolean): CartOpResult {
  const key = lineKey(product, variant)
  const stock = variant?.stock ?? product.stock
  const ex = cart.find((l) => lineKey(l.product, l.variant) === key)
  if (ex) {
    if (online && ex.qty >= stock) return { cart, blocked: true }
    return { cart: cart.map((l) => (lineKey(l.product, l.variant) === key ? { ...l, qty: l.qty + 1 } : l)), blocked: false }
  }
  if (online && stock < 1) return { cart, blocked: true }
  return { cart: [...cart, { product, ...(variant ? { variant } : {}), qty: 1 }], blocked: false }
}

// Ubah qty satu baris: qty < 1 menghapus baris; saat online qty di atas
// stok ditolak (baris dikembalikan apa adanya, blocked).
export function setLineQty(cart: CartLine[], line: CartLine, qty: number, online: boolean): CartOpResult {
  const key = lineKey(line.product, line.variant)
  if (qty < 1) return { cart: cart.filter((item) => lineKey(item.product, item.variant) !== key), blocked: false }
  let blocked = false
  const next = cart.map((l) => {
    if (lineKey(l.product, l.variant) !== key) return l
    if (online && qty > (l.variant?.stock ?? l.product.stock)) { blocked = true; return l }
    return { ...l, qty }
  })
  return { cart: next, blocked }
}

export function removeLine(cart: CartLine[], line: CartLine): CartLine[] {
  return cart.filter((x) => lineKey(x.product, x.variant) !== lineKey(line.product, line.variant))
}

export function productCategories(products: Product[]): string[] {
  return [...new Set(products.map((p) => p.category_id).filter((x): x is string => !!x))]
}

export function categoryName(products: Product[], id: string | null): string {
  return products.find((p) => p.category_id === id)?.category_name ?? '—'
}

export function filterProducts(products: Product[], cat: string, q: string): Product[] {
  return products.filter((p) => {
    if (cat !== 'Semua' && p.category_id !== cat) return false
    const s2 = q.toLowerCase()
    return !s2 || p.name.toLowerCase().includes(s2) || p.sku.toLowerCase().includes(s2) || p.barcode.toLowerCase().includes(s2) || (p.variants ?? []).some((variant) => variant.sku.toLowerCase().includes(s2))
  })
}

// Harga tampil di kartu katalog: harga varian termurah (qty 1, ikut
// tingkat harga) atau harga jual produk itu sendiri.
export function productDisplayPrice(product: Product): number {
  return product.variants?.length
    ? Math.min(...product.variants.map((variant) => getUnitPrice(1, variant.sell_price, variant.price_tiers ?? [])))
    : product.sell_price
}

export function taxPctOf(settings: Pick<StoreSettings, 'taxEnabled' | 'taxPct'> | null | undefined): number {
  return settings?.taxEnabled ? settings.taxPct : 0
}

// Diskon efektif: bila pratinjau promo aktif, nilainya promo + voucher
// dari server; selain itu diskon manual kasir.
export function effectiveDiscount(activePromo: PromoPreview | null, manualDiscount: number): number {
  return activePromo ? activePromo.promo_discount + activePromo.voucher_discount : manualDiscount
}

// Pajak dihitung dari (subtotal − diskon efektif) dan dibulatkan
// Math.round — persis rumus inline sebelumnya.
export function computeTax(subtotal: number, effDiscount: number, taxPct: number): number {
  return Math.round((subtotal - effDiscount) * (taxPct / 100))
}

export function computeTotal(subtotal: number, effDiscount: number, tax: number): number {
  return subtotal - effDiscount + tax
}

// Kembalian: uang pas selalu 0; selain itu jumlah dibayar − total
// (negatif berarti kurang bayar).
export function computeChange(paid: string, total: number, exactCash: boolean): number {
  return exactCash ? 0 : Number(paid) - total
}

// Diskon manual dibatasi 0..subtotal (dari onValue NumInput mentah).
export function clampManualDiscount(raw: string, subtotal: number): number {
  return Math.min(subtotal, Math.max(0, Number(raw || 0)))
}

export function isVoucherDirty(voucherCode: string, appliedVoucherCode: string | null): boolean {
  return voucherCode.trim() !== (appliedVoucherCode ?? '')
}

// Kode voucher yang dikirim saat checkout: saat offline kode yang belum
// diterapkan tetap dipakai apa adanya; saat online hanya kode yang
// sudah diterapkan.
export function resolveCheckoutVoucherCode(online: boolean, voucherCode: string, appliedVoucherCode: string | null): string | null {
  return !online && isVoucherDirty(voucherCode, appliedVoucherCode) ? voucherCode.trim() || null : appliedVoucherCode
}

export function isRedeemingPoints(method: PayMethod, loyaltyActive: boolean, selectedCustomer: Customer | null, pointsRedeem: string): boolean {
  return method !== 'Credit' && loyaltyActive && !!selectedCustomer && Number(pointsRedeem) > 0
}

// Nominal cepat tunai: jumlah dibayar ditambah nominal ketukan.
export function addQuickCash(paid: string, amount: number): string {
  return String((Number(paid) || 0) + amount)
}

// Keputusan gerbang shift di awal checkout (perbaikan bug offline-penuh).
//
// Bug lama: checkout() selalu `await refreshShift()` lebih dulu. Refetch
// TanStack Query dengan networkMode bawaan 'online' tertahan selama
// perangkat offline, dan navigator.onLine baru dibaca SETELAH await itu
// — begitu koneksi pulih nilainya sudah true, sehingga cabang antrean
// offline tidak pernah terjangkau bila perangkat sudah offline penuh
// sebelum klik Bayar.
//
// Aturan baru (fungsi murni, tanpa efek samping):
// - Offline: JANGAN refresh sama sekali. Keputusan hanya dari cache
//   query status shift terakhir yang diketahui — `false` berarti pasti
//   belum ada shift (blokir, sama seperti gerbang online), `true`/`null`
//   (belum pernah diketahui) berarti lanjut ke antrean offline; backend
//   tetap penegak terakhir saat antrean tersinkron.
// - Online: wajib refresh lebih dulu (`needs-refresh` selama hasil
//   refresh belum ada). Hasil refresh `false` memblokir; `true`/`null`
//   (status tak diketahui karena gagal jaringan) mengizinkan — persis
//   perilaku lama yang hanya memblokir pada nilai `false`.
export type CheckoutShiftGate = 'allow' | 'block-no-shift' | 'needs-refresh'

export function resolveCheckoutShiftGate(args: {
  online: boolean
  cachedHasShift: boolean | null
  refreshedHasShift?: boolean | null
}): CheckoutShiftGate {
  if (!args.online) {
    return args.cachedHasShift === false ? 'block-no-shift' : 'allow'
  }
  if (args.refreshedHasShift === undefined) return 'needs-refresh'
  return args.refreshedHasShift === false ? 'block-no-shift' : 'allow'
}

// Rakit body checkout persis seperti sebelumnya: field opsional hanya
// disertakan bila nilainya ada (voucher, pelanggan, tukar poin).
export function buildCheckoutBody(args: {
  cart: CartLine[]
  discount: number
  method: PayMethod
  paid: string
  exactCash: boolean
  total: number
  voucherCode: string | null
  customerId?: number
  redeemingPoints: boolean
  pointsRedeem: string
}): CheckoutBody {
  return {
    items: checkoutItems(args.cart),
    discount: args.discount,
    method: args.method,
    paid: args.method === 'Cash' ? (args.exactCash ? args.total : Number(args.paid)) : 0,
    ...(args.voucherCode ? { voucherCode: args.voucherCode } : {}),
    ...(args.customerId !== undefined ? { customerId: args.customerId } : {}),
    ...(args.redeemingPoints ? { pointsRedeem: Number(args.pointsRedeem) } : {}),
  }
}
