import type { Product } from '../../../shared/api'
import { fmtRp } from '../../../shared/session/store'
import { AlertStrip } from '../../../shared/ui'
import { Skeleton } from '@/shared/ui/primitives/skeleton'
import { effectiveStock, productDisplayPrice, type CartLine } from '../logic'

// Grid katalog kasir: pencarian, chip kategori, dan kartu produk.
// Markup dipindah apa adanya dari halaman Kasir lama.
export function ProductGrid(props: {
  q: string
  onQueryChange: (value: string) => void
  cats: string[]
  cat: string
  onCategoryChange: (value: string) => void
  categoryName: (id: string | null) => string
  err: string
  products: Product[] | null
  filtered: Product[]
  cart: CartLine[]
  onChoose: (product: Product) => void
}) {
  const { q, onQueryChange, cats, cat, onCategoryChange, categoryName, err, products, filtered, cart, onChoose } = props
  return (
    <section>
      <input
        value={q}
        onChange={(e) => onQueryChange(e.target.value)}
        placeholder="Cari produk, SKU, atau barcode…"
        aria-label="Cari produk"
        className="mb-3 w-full rounded-md border border-border bg-paper px-3.5 py-3 text-[15px] focus:border-jet focus:outline-none"
      />
      <div className="mb-4 flex flex-wrap gap-2">
        <button
          onClick={() => onCategoryChange('Semua')}
          className={`rounded-full border px-3 py-1 text-xs ${cat === 'Semua' ? 'border-jet bg-jet text-paper' : 'border-dove text-muted hover:border-jet'}`}
        >
          Semua
        </button>
        {cats.map((id) => (
          <button
            key={id}
            onClick={() => onCategoryChange(id)}
            className={`rounded-full border px-3 py-1 text-xs ${cat === id ? 'border-jet bg-jet text-paper' : 'border-dove text-muted hover:border-jet'}`}
          >
            {categoryName(id)}
          </button>
        ))}
      </div>
      {err && <AlertStrip className="mb-3">{err}</AlertStrip>}
      {!products ? (
        <div aria-busy="true" aria-label="Memuat katalog">
          <div className="mb-4 flex flex-wrap gap-2" aria-hidden="true">
            {[0, 1, 2, 3].map((i) => (
              <Skeleton key={i} className="h-6 w-16 rounded-full" />
            ))}
          </div>
          <div className="grid max-h-[62vh] grid-cols-2 gap-2.5 overflow-y-auto pr-1 md:grid-cols-3 xl:grid-cols-4" aria-hidden="true">
            {Array.from({ length: 8 }).map((_, i) => (
              <div key={i} className="space-y-2 rounded-xl border border-dove p-3.5">
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-3 w-1/3" />
                <Skeleton className="h-4 w-1/2" />
              </div>
            ))}
          </div>
        </div>
      ) : (
        <div className="grid max-h-[62vh] grid-cols-2 gap-2.5 overflow-y-auto pr-1 md:grid-cols-3 xl:grid-cols-4">
          {filtered.map((p) => {
            const price = productDisplayPrice(p)
            const remaining = effectiveStock(cart, p)
            const stockBlocked = navigator.onLine && remaining <= 0
            return (
              <button
                key={p.id}
                onClick={() => onChoose(p)}
                disabled={stockBlocked}
                className={`rounded-xl border p-3.5 text-left transition ${stockBlocked ? 'cursor-not-allowed border-dove opacity-40' : 'border-dove hover:border-jet'}`}
              >
                <p className="text-[13px] font-medium leading-snug">{p.name}</p>
                <p className="mt-1 font-mono text-[11px] text-fog">{remaining} stok{p.variants?.length ? ` · ${p.variants.length} varian` : ''}</p>
                <p className="mt-1 font-mono text-sm font-medium tabular-nums">{p.variants?.length ? 'Mulai ' : ''}{fmtRp(price)}</p>
              </button>
            )
          })}
          {filtered.length === 0 && <p className="col-span-full py-10 text-center text-sm text-fog">Tidak ada produk ditemukan.</p>}
        </div>
      )}
    </section>
  )
}
