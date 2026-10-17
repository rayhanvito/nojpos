import type { PromoPreview } from '../../../shared/api'
import { fmtRp } from '../../../shared/session/store'
import { AlertStrip, Button, NumInput } from '../../../shared/ui'
import { lineKey, unitPrice as lineUnitPrice, type CartLine } from '../logic'

// Badan keranjang bersama: dipakai aside desktop & sheet mobile (satu-satunya
// perbedaan: kelas daftar item + sufiks id voucher agar tidak ganda di DOM).
// Markup dipindah apa adanya dari halaman Kasir lama.
export function CartPanel(props: {
  idSuffix: string
  listCls: string
  hasShift: boolean | null
  cart: CartLine[]
  onRemoveLine: (line: CartLine) => void
  onSetQty: (line: CartLine, qty: number) => void
  voucherCode: string
  onVoucherCodeChange: (value: string) => void
  onApplyVoucher: () => void
  previewLoading: boolean
  promoError: string
  promoNetworkFailure: boolean
  activePromo: PromoPreview | null
  subtotal: number
  discount: number
  onDiscountChange: (raw: string) => void
  taxPct: number
  tax: number
  total: number
}) {
  const {
    idSuffix, listCls, hasShift, cart, onRemoveLine, onSetQty, voucherCode, onVoucherCodeChange,
    onApplyVoucher, previewLoading, promoError, promoNetworkFailure, activePromo, subtotal,
    discount, onDiscountChange, taxPct, tax, total,
  } = props
  return (
    <>
    {hasShift === false && (
      <div className="mb-3 rounded-lg border border-sunbeam/60 bg-sunbeam/10 px-3 py-2 text-[13px] leading-snug text-fg">
        Belum ada shift aktif. Buka shift dulu lewat tombol <b>Status</b> di kanan atas untuk bisa transaksi.
      </div>
    )}
    <div className={listCls}>
      {cart.length === 0 && <p className="py-6 text-center text-sm text-fog">Belum ada item. Klik produk untuk menambah.</p>}
      {cart.map((l) => (
        <div key={lineKey(l.product, l.variant)} className="rounded-lg border border-dove bg-paper p-2.5">
          <div className="flex items-start justify-between gap-2">
            <div>
              <p className="text-[13px] font-medium leading-snug">{l.product.name}{l.variant ? ` · ${l.variant.name}` : ''}</p>
              {l.variant && <p className="mt-0.5 font-mono text-[11px] text-fog">SKU {l.variant.sku}</p>}
            </div>
            <button onClick={() => onRemoveLine(l)} aria-label={`Hapus ${l.product.name}${l.variant ? ` ${l.variant.name}` : ''}`} className="-mr-1 grid size-8 place-items-center rounded-full text-fog hover:text-ember">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round"><path d="M18 6 6 18M6 6l12 12" /></svg>
            </button>
          </div>
          <div className="mt-2 flex items-center justify-between">
            <div className="flex items-center gap-1.5">
              <button onClick={() => onSetQty(l, l.qty - 1)} aria-label={`Kurangi ${l.product.name}`} className="grid h-9 w-9 place-items-center rounded-full border border-dove text-base hover:border-jet">−</button>
              <span className="w-8 text-center font-mono text-sm tabular-nums">{l.qty}</span>
              <button onClick={() => onSetQty(l, l.qty + 1)} aria-label={`Tambah ${l.product.name}`} className="grid h-9 w-9 place-items-center rounded-full border border-dove text-base hover:border-jet">+</button>
            </div>
            <span className="text-right">
              <span className="block font-mono text-[11px] text-muted">{fmtRp(lineUnitPrice(l))} / satuan</span>
              <span className="font-mono text-[13px] tabular-nums">{fmtRp(lineUnitPrice(l) * l.qty)}</span>
            </span>
          </div>
        </div>
      ))}
    </div>

    <div className="mt-3 border-t border-dove pt-3">
      <label htmlFor={`voucher-code${idSuffix}`} className="mb-1.5 block text-[13px] font-medium text-steel">Kode voucher</label>
      <div className="flex gap-2">
        <input
          id={`voucher-code${idSuffix}`}
          value={voucherCode}
          onChange={(e) => onVoucherCodeChange(e.target.value)}
          onKeyDown={(e) => { if (e.key === 'Enter') { e.preventDefault(); onApplyVoucher() } }}
          placeholder="Masukkan kode"
          className="min-w-0 flex-1 rounded-md border border-border bg-paper px-3 py-2 text-sm focus:border-jet focus:outline-none"
        />
        <Button onClick={onApplyVoucher} disabled={navigator.onLine && previewLoading} className="px-4 py-2 text-sm">Terapkan</Button>
      </div>
      {previewLoading && <p className="mt-2 text-xs text-muted">Memeriksa promo…</p>}
      {promoError && !promoNetworkFailure && <AlertStrip className="mt-2">{promoError}</AlertStrip>}
    </div>

    <div className="mt-3 space-y-1.5 border-t border-dove pt-3 text-sm">
      <div className="flex justify-between"><span className="text-muted">Subtotal</span><span className="font-mono tabular-nums">{fmtRp(subtotal)}</span></div>
      {activePromo ? (
        <>
          {(activePromo.promo_discount > 0 || activePromo.voucher_discount === 0) && (
            <div className="flex justify-between"><span className="min-w-0 truncate text-muted">Promo · {activePromo.promo_name || activePromo.voucher_name}</span><span className="shrink-0 font-mono tabular-nums">−{fmtRp(activePromo.promo_discount)}</span></div>
          )}
          {activePromo.voucher_discount > 0 && (
            <div className="flex justify-between"><span className="min-w-0 truncate text-muted">Voucher · {activePromo.voucher_code}</span><span className="shrink-0 font-mono tabular-nums">−{fmtRp(activePromo.voucher_discount)}</span></div>
          )}
        </>
      ) : (
        <div className="flex items-center justify-between">
          <span className="text-muted">Diskon</span>
          <NumInput
            value={discount}
            onValue={onDiscountChange}
            className="w-24 rounded border border-dove bg-paper px-2 py-0.5 text-right font-mono text-[13px] tabular-nums focus:border-jet focus:outline-none"
          />
        </div>
      )}
      {taxPct > 0 && (
        <div className="flex justify-between"><span className="text-muted">Pajak ({taxPct}%)</span><span className="font-mono tabular-nums">{fmtRp(tax)}</span></div>
      )}
      <div className="flex justify-between text-base font-medium"><span>Total</span><span className="font-mono tabular-nums">{fmtRp(total)}</span></div>
    </div>
    </>
  )
}
