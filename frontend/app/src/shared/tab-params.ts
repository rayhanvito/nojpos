// Nilai ?tab= kanonis (Bahasa Inggris) + alias nilai lama (Bahasa Indonesia).
// SATU-SATUNYA sumber kebenaran pemetaan tab URL: halaman membaca parameter
// lewat use-tab-param (yang memakai modul ini), tes menguji modul ini, dan
// tidak ada halaman yang memetakan alias sendiri-sendiri. Label tampilan tab
// tetap Bahasa Indonesia di masing-masing halaman — yang Inggris hanya
// nilai di URL.

export interface TabNormalization<T extends string> {
  /** Nilai kanonis yang harus aktif (fallback bila raw kosong/tidak dikenal). */
  tab: T
  /** true bila raw adalah alias lama dan URL perlu dinormalkan (replace). */
  aliased: boolean
}

export function normalizeTabParam<T extends string>(
  raw: string | null | undefined,
  canonical: readonly T[],
  aliases: Readonly<Record<string, T>>,
  fallback: T,
): TabNormalization<T> {
  if (raw && (canonical as readonly string[]).includes(raw)) return { tab: raw as T, aliased: false }
  if (raw && Object.prototype.hasOwnProperty.call(aliases, raw)) return { tab: aliases[raw], aliased: true }
  return { tab: fallback, aliased: false }
}

// --- Laporan (/app/reports) ---
export const REPORT_TABS = ['sales', 'products', 'profit', 'stock', 'shift', 'cashier', 'tax'] as const
export type ReportTab = (typeof REPORT_TABS)[number]
export const REPORT_TAB_ALIASES: Readonly<Record<string, ReportTab>> = {
  penjualan: 'sales',
  produk: 'products',
  stok: 'stock',
  kasir: 'cashier',
  pajak: 'tax',
}
export const normalizeReportTab = (raw: string | null | undefined): TabNormalization<ReportTab> =>
  normalizeTabParam(raw, REPORT_TABS, REPORT_TAB_ALIASES, 'sales')

// --- Stok (/app/stock) ---
// 'stocktake' adalah slug Inggris untuk Opname; 'movement' untuk Riwayat.
export const STOCK_TABS = ['stock', 'stocktake', 'movement'] as const
export type StockTab = (typeof STOCK_TABS)[number]
export const STOCK_TAB_ALIASES: Readonly<Record<string, StockTab>> = {
  riwayat: 'movement',
  opname: 'stocktake',
  stok: 'stock',
}
export const normalizeStockTab = (raw: string | null | undefined): TabNormalization<StockTab> =>
  normalizeTabParam(raw, STOCK_TABS, STOCK_TAB_ALIASES, 'stock')

// --- Pembelian (/app/purchases) ---
export const PURCHASE_TABS = ['purchases', 'expenses'] as const
export type PurchaseTab = (typeof PURCHASE_TABS)[number]
export const PURCHASE_TAB_ALIASES: Readonly<Record<string, PurchaseTab>> = {
  po: 'purchases',
  pembelian: 'purchases',
  pengeluaran: 'expenses',
}
export const normalizePurchaseTab = (raw: string | null | undefined): TabNormalization<PurchaseTab> =>
  normalizeTabParam(raw, PURCHASE_TABS, PURCHASE_TAB_ALIASES, 'purchases')
