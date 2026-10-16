import type { Product } from '../../lib/api'

export const stockTotalOf = (p: Product) => p.variants?.length ? p.variants.reduce((sum, v) => sum + v.stock, 0) : p.stock
export const isLowOf = (p: Product) => p.variants?.length
  ? p.variants.some((v) => v.min_stock > 0 && v.stock <= v.min_stock)
  : (p.min_stock ?? 0) > 0 && p.stock > 0 && p.stock <= (p.min_stock ?? 0)
export const kondisiOf = (p: Product) => stockTotalOf(p) === 0 ? 'habis' : isLowOf(p) ? 'menipis' : 'aman'
// ponytail: paging client-side di atas fetchAll; pindah ke paging server-side bila katalog puluhan ribu.

export function StockCell({ product }: { product: Product }) {
  const stock = product.variants?.length ? product.variants.reduce((sum, variant) => sum + variant.stock, 0) : product.stock
  const low = product.variants?.length
    ? product.variants.some((variant) => variant.min_stock > 0 && variant.stock <= variant.min_stock)
    : (product.min_stock ?? 0) > 0 && product.stock > 0 && product.stock <= (product.min_stock ?? 0)
  const dot = stock === 0 ? 'var(--t-ember)' : low ? 'var(--t-sunbeam)' : 'var(--t-sprout)'
  return (
    <span className={`inline-flex items-center gap-1.5 ${low ? 'font-medium text-fg' : ''}`}>
      <span className="size-2 shrink-0 rounded-full" style={{ background: dot }} aria-hidden="true" />
      {stock} {product.unit}
    </span>
  )
}
