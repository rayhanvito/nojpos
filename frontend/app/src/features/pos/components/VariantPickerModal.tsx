import type { Product, Variant } from '../../../shared/api'
import { fmtRp } from '../../../shared/session/store'
import { getUnitPrice } from '../../../shared/price-tier'
import { Modal } from '../../../shared/ui'
import { availableStock as cartAvailableStock, type CartLine } from '../logic'

// Modal pemilih varian produk. Markup dipindah apa adanya dari
// halaman Kasir lama.
export function VariantPickerModal(props: {
  product: Product | null
  cart: CartLine[]
  onAdd: (product: Product, variant?: Variant) => void
  onClose: () => void
}) {
  const { product, cart, onAdd, onClose } = props
  return (
    <Modal open={!!product} title={`Pilih Varian · ${product?.name ?? ''}`} onClose={onClose}>
      <div className="space-y-2">
        {product?.variants?.map((variant) => {
          const remaining = cartAvailableStock(cart, product, variant)
          const price = getUnitPrice(1, variant.sell_price, variant.price_tiers ?? [])
          const stockBlocked = navigator.onLine && remaining <= 0
          return (
            <button
              key={variant.id}
              type="button"
              disabled={stockBlocked}
              onClick={() => { onAdd(product, variant); onClose() }}
              className="flex w-full items-center justify-between gap-3 rounded-lg border border-dove bg-paper px-3.5 py-3 text-left transition hover:border-jet disabled:cursor-not-allowed disabled:opacity-40"
            >
              <span className="min-w-0">
                <span className="block truncate text-sm font-medium text-fg">{variant.name}</span>
                <span className="block truncate font-mono text-xs text-muted">{variant.sku} · stok {remaining}</span>
              </span>
              <span className="shrink-0 font-mono text-sm tabular-nums">{fmtRp(price)}</span>
            </button>
          )
        })}
        {product?.variants?.length === 0 && <p className="text-sm text-muted">Belum ada varian untuk produk ini.</p>}
      </div>
    </Modal>
  )
}
