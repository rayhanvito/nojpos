import type { Customer, PayMethod, PromoPreview, StoreSettings } from '../../../shared/api'
import { isCreditPaymentValid } from '../../../shared/checkout'
import { fmtRp } from '../../../shared/session/store'
import { AlertStrip, Button, CheckRow, Modal, ModalActions, NumInput } from '../../../shared/ui'
import { addQuickCash } from '../logic'

// Nominal cepat tunai: ketuk untuk menambah ke Jumlah dibayar.
const CASH_QUICK: { v: number; label: string }[] = [
  { v: 500, label: '500' }, { v: 1000, label: '1rb' },
  { v: 2000, label: '2rb' }, { v: 5000, label: '5rb' },
  { v: 10000, label: '10rb' }, { v: 20000, label: '20rb' },
  { v: 50000, label: '50rb' }, { v: 100000, label: '100rb' },
]

const METHODS: PayMethod[] = ['Cash', 'QRIS', 'Bank Transfer', 'Credit']

// Modal pembayaran kasir. Markup dipindah apa adanya dari halaman
// Kasir lama; semua status tetap dimiliki halaman (props masuk).
export function PaymentModal(props: {
  open: boolean
  onClose: () => void
  err: string
  promoError: string
  promoNetworkFailure: boolean
  method: PayMethod
  onMethodChange: (method: PayMethod) => void
  selectedCustomer: Customer | null
  onSelectCustomer: (customer: Customer) => void
  onClearCustomer: () => void
  customerQuery: string
  onCustomerQueryChange: (value: string) => void
  customerResults: Customer[]
  customerLookupBusy: boolean
  customerLookupError: string
  loyaltyActive: boolean
  pointsRedeem: string
  onPointsRedeemChange: (value: string) => void
  settings: StoreSettings | null
  exactCash: boolean
  onExactCashChange: (value: boolean) => void
  redeemingPoints: boolean
  paid: string
  onPaidChange: (value: string) => void
  change: number
  activePromo: PromoPreview | null
  discount: number
  taxPct: number
  tax: number
  total: number
  busy: boolean
  onCheckout: () => void
}) {
  const {
    open, onClose, err, promoError, promoNetworkFailure, method, onMethodChange,
    selectedCustomer, onSelectCustomer, onClearCustomer, customerQuery, onCustomerQueryChange,
    customerResults, customerLookupBusy, customerLookupError, loyaltyActive, pointsRedeem,
    onPointsRedeemChange, settings, exactCash, onExactCashChange, redeemingPoints, paid,
    onPaidChange, change, activePromo, discount, taxPct, tax, total, busy, onCheckout,
  } = props
  return (
    <Modal open={open} title="Pembayaran" onClose={onClose}>
      <div className="space-y-4">
        {err && <AlertStrip>{err}</AlertStrip>}
        {promoError && !promoNetworkFailure && <AlertStrip>{promoError}</AlertStrip>}
        <div className="space-y-2">
          <label htmlFor="checkout-customer" className="block text-[13px] font-medium text-steel">Pelanggan {method === 'Credit' ? '(wajib)' : '(opsional)'}</label>
          {selectedCustomer ? (
            <div className="flex items-center justify-between gap-3 rounded-lg border border-dove bg-paper px-3.5 py-2.5">
              <div className="min-w-0">
                <p className="truncate text-sm font-medium text-fg">{selectedCustomer.name}</p>
                <p className="truncate text-xs text-muted">{selectedCustomer.phone || 'Tanpa nomor telepon'}</p>
              </div>
              <button
                type="button"
                onClick={onClearCustomer}
                className="shrink-0 rounded-md px-2 py-1 text-xs text-steel hover:bg-surface hover:text-fg"
              >
                Kosongkan
              </button>
            </div>
          ) : (
            <>
              <input
                id="checkout-customer"
                value={customerQuery}
                onChange={(e) => onCustomerQueryChange(e.target.value)}
                placeholder="Cari nama atau telepon…"
                autoComplete="off"
                className="w-full rounded-md border border-border bg-paper px-3.5 py-2.5 text-[15px] focus:border-jet focus:outline-none"
              />
              {customerLookupBusy && <p className="text-xs text-muted" role="status">Mencari pelanggan…</p>}
              {customerLookupError && <p className="text-xs text-ember" role="alert">{customerLookupError}</p>}
              {customerResults.length > 0 && (
                <div className="max-h-40 overflow-y-auto rounded-lg border border-dove bg-paper">
                  {customerResults.map((customer) => (
                    <button
                      key={customer.id}
                      type="button"
                      onClick={() => onSelectCustomer(customer)}
                      className="flex w-full items-center justify-between gap-3 border-b border-dove px-3 py-2 text-left last:border-b-0 hover:bg-surface"
                    >
                      <span className="min-w-0">
                        <span className="block truncate text-sm font-medium text-fg">{customer.name}</span>
                        <span className="block truncate text-xs text-muted">{customer.phone || 'Tanpa nomor telepon'}</span>
                      </span>
                      <span className="shrink-0 font-mono text-xs text-muted">{customer.points.toLocaleString('id-ID')} poin</span>
                    </button>
                  ))}
                </div>
              )}
              {customerQuery.trim() && !customerLookupBusy && !customerLookupError && customerResults.length === 0 && (
                <p className="text-xs text-muted">Pelanggan tidak ditemukan.</p>
              )}
            </>
          )}
          {method === 'Credit' && <p className="text-xs text-muted" role="status">Pilih pelanggan untuk mencatat piutang. Jumlah dibayar untuk Kredit adalah Rp0.</p>}
          {method !== 'Credit' && loyaltyActive && selectedCustomer && (
            <div className="space-y-2 rounded-lg bg-surface px-3.5 py-3">
              <p className="text-sm text-fg">Saldo poin: <span className="font-mono font-medium tabular-nums">{selectedCustomer.points.toLocaleString('id-ID')} poin</span></p>
              <label htmlFor="points-redeem" className="block text-[13px] font-medium text-steel">Tukar poin</label>
              <input
                id="points-redeem"
                type="number"
                min="0"
                max={selectedCustomer.points}
                step="1"
                value={pointsRedeem}
                disabled={selectedCustomer.points < 1}
                onChange={(e) => onPointsRedeemChange(e.target.value)}
                placeholder={selectedCustomer.points > 0 ? 'Jumlah poin' : 'Belum ada poin'}
                className="w-full rounded-md border border-border bg-paper px-3.5 py-2.5 text-[15px] focus:border-jet focus:outline-none disabled:opacity-50"
              />
              {Number(pointsRedeem) > 0 && (
                <p className="text-xs text-muted">Nilai tukar {Number(pointsRedeem).toLocaleString('id-ID')} poin × {fmtRp(settings?.loyaltyPointValue ?? 100)} = {fmtRp(Number(pointsRedeem) * (settings?.loyaltyPointValue ?? 100))}.</p>
              )}
              <p className="text-xs text-muted">Nilai akhir dan kembalian setelah tukar poin ditampilkan dari hasil transaksi.</p>
            </div>
          )}
        </div>
        <div>
          <p className="mb-2 text-[13px] font-medium text-steel">Metode pembayaran</p>
          <div className="flex flex-wrap gap-2">
            {METHODS.map((m) => (
              <button
                key={m}
                onClick={() => onMethodChange(m)}
                className={`rounded-full border px-3.5 py-1.5 text-xs ${method === m ? 'border-jet bg-jet text-paper' : 'border-dove text-muted hover:border-jet'}`}
              >
                {m === 'Credit' ? 'Kredit' : m}
              </button>
            ))}
          </div>
        </div>
        {method === 'Cash' ? (
          <div className="space-y-3">
            <CheckRow
              checked={exactCash}
              disabled={redeemingPoints}
              onChange={onExactCashChange}
              className="text-sm font-normal text-fg"
            >
              Uang Pas{redeemingPoints ? ' (nonaktif saat tukar poin)' : ''}
            </CheckRow>
            {!exactCash && (
              <>
                <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
                  Jumlah dibayar
                  <NumInput
                    value={paid}
                    onValue={onPaidChange}
                    placeholder="0"
                    autoFocus
                    hint={redeemingPoints ? 'Isi 0 bila poin menutup seluruh total; sisa tunai divalidasi server.' : undefined}
                    className="rounded-md border border-border bg-paper px-3.5 py-2.5 text-[15px] focus:border-jet focus:outline-none"
                  />
                </label>
                <div>
                  <p className="mb-2 text-[13px] font-medium text-steel">Tambah cepat</p>
                  <div className="grid grid-cols-4 gap-2">
                    {CASH_QUICK.map((q) => (
                      <button
                        key={q.v} type="button"
                        onClick={() => onPaidChange(addQuickCash(paid, q.v))}
                        className="rounded-lg border border-dove bg-paper px-2 py-2 font-mono text-xs tabular-nums text-fg transition outline-none hover:border-jet hover:text-jet focus-visible:ring-2 focus-visible:ring-ring active:translate-y-px"
                      >
                        +{q.label}
                      </button>
                    ))}
                  </div>
                </div>
              </>
            )}
          </div>
        ) : (
          <p className="text-sm text-muted">
            {method === 'Credit'
              ? `Total ${fmtRp(total)} akan dicatat sebagai piutang dengan jatuh tempo ${settings?.creditDueDays ?? 14} hari.`
              : redeemingPoints
              ? `Total sebelum tukar poin: ${fmtRp(total)}. Nominal akhir ditentukan setelah transaksi.`
              : `Total ${fmtRp(total)} akan dicatat sebagai pembayaran ${method}.`}
          </p>
        )}
        {method === 'Cash' && !redeemingPoints && (exactCash ? change === 0 : paid) && change >= 0 && (
          <div className="flex justify-between rounded-lg bg-surface px-3.5 py-3 text-sm">
            <span className="text-muted">Kembalian</span>
            <span className="font-mono font-medium tabular-nums">{fmtRp(change)}</span>
          </div>
        )}
        <div className="space-y-1.5 text-sm">
          {activePromo ? (
            <div className="flex justify-between"><span className="text-muted">Promo · {activePromo.promo_name}</span><span className="font-mono tabular-nums">−{fmtRp(activePromo.promo_discount)}</span></div>
          ) : (
            <div className="flex justify-between"><span className="text-muted">Diskon</span><span className="font-mono tabular-nums">&minus;{fmtRp(discount)}</span></div>
          )}
          <div className="flex justify-between">
            <span className="text-muted">Pajak{taxPct > 0 ? ` (${taxPct}%)` : ''}</span>
            <span className="font-mono tabular-nums">+{fmtRp(tax)}</span>
          </div>
        </div>
        <div className="flex justify-between border-t border-dove pt-3 text-[15px]">
          <span>Total</span>
          <span className="font-mono font-medium tabular-nums">{fmtRp(total)}</span>
        </div>
        <ModalActions>
          <Button className="w-full" onClick={onCheckout} disabled={busy || (method === 'Credit' && !isCreditPaymentValid(selectedCustomer?.id ?? null))}>{busy ? 'Memproses…' : 'Selesaikan Transaksi'}</Button>
        </ModalActions>
      </div>
    </Modal>
  )
}
