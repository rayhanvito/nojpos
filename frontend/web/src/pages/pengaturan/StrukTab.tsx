import type { StoreSettings } from '../../lib/api'
import { useMemo, useState } from 'react'
import type { Trx } from '../../lib/api'
import { Receipt } from '../../lib/receipt'
import { fmtRp, useDB } from '../../lib/store'
import { Printer } from 'lucide-react'
import { Button, Input, Modal, Pill, Select, SettingsCard, SettingsHead } from '../../lib/ui'
import { FormActions, INPUT_CLS, LABEL_CLS, SAMPLE_ITEMS, calcTax, todayStr} from './shared'

export function StrukTab({ form, set, setV, hasExt, dirty, busy, onSave, onReset }: {
  form: StoreSettings
  set: (k: keyof StoreSettings) => (v: string) => void
  setV: <K extends keyof StoreSettings>(k: K, v: StoreSettings[K]) => void
  hasExt: boolean
  dirty: boolean
  busy: boolean
  onSave: () => void
  onReset: () => void
}) {
  const { session } = useDB()
  const today = todayStr()
  const [printTest, setPrintTest] = useState(false)

  const sample = useMemo(() => {
    const taxOn = form?.taxEnabled ?? false
    const pct = form?.taxPct ?? 0
    const subtotal = SAMPLE_ITEMS.reduce((n, i) => n + i.qty * i.price, 0)
    if (!taxOn) return { subtotal, tax: 0, total: subtotal }
    const { tax, total } = calcTax(subtotal, pct, form?.taxInclusive ?? false, form?.taxRounding ?? 'none')
    return { subtotal, tax, total }
  }, [form])

  const sampleTrx: Trx = useMemo(() => ({
    id: '1',
    cashier_name: session?.name ?? 'Kasir',
    items: SAMPLE_ITEMS.map((i, ix) => ({ product_id: `s${ix}`, name: i.name, buy_price: i.price, price: i.price, qty: i.qty })),
    subtotal: sample.subtotal,
    discount: 0,
    tax: sample.tax,
    total: sample.total,
    method: 'Cash',
    paid: sample.total,
    change: 0,
    status: 'completed',
    customer: '',
    created_at: new Date().toISOString(),
  }), [sample, session?.name])
  return (
    <>
    <div className="grid items-start gap-4 lg:grid-cols-5">
      <SettingsCard className="lg:col-span-3">
        <SettingsHead title="Pengaturan Struk" sub="Atur tampilan dan konten struk sesuai kebutuhan toko Anda." />
        <div className="space-y-4">
          <Input label="Header Toko" value={form.storeName} onChange={set('storeName')} hint="Nama toko yang akan ditampilkan di struk" />
          <label className={LABEL_CLS}>
            Footer Struk
            <textarea value={form.receiptHeader} onChange={(e) => set('receiptHeader')(e.target.value)} rows={2} placeholder="Terima kasih telah berbelanja di toko kami." className={INPUT_CLS} />
          </label>
          <div>
            <p className="mb-1 text-[13px] font-medium text-steel">Tampilan struk</p>
            {hasExt ? (
            <div className="divide-y divide-dove rounded-xl border border-dove px-4">
              {([
                ['Tampilkan Nama Kasir', 'receiptShowCashier'],
                ['Tampilkan Metode Pembayaran', 'receiptShowMethod'],
                ['Tampilkan Pajak', 'receiptShowTax'],
                ['Tampilkan Diskon', 'receiptShowDiscount'],
                ['Tampilkan QRIS / Catatan', 'receiptShowNote'],
              ] as const).map(([l, k]) => (
                <label key={l} className="flex cursor-pointer items-center justify-between gap-3 py-2.5 text-sm text-fg">
                  {l}
                  <input
                    type="checkbox" checked={form[k] ?? true}
                    onChange={(e) => setV(k, e.target.checked)}
                    className="h-4 w-4 accent-jet"
                  />
                </label>
              ))}
            </div>
            ) : (
            <div className="divide-y divide-dove rounded-xl border border-dove px-4 opacity-80">
              {['Tampilkan Nama Kasir', 'Tampilkan Metode Pembayaran', 'Tampilkan Pajak', 'Tampilkan Diskon', 'Tampilkan QRIS / Catatan'].map((l) => (
                <label key={l} className="flex cursor-not-allowed items-center justify-between gap-3 py-2.5 text-sm text-fg">
                  {l}
                  <span className="flex items-center gap-2">
                    <Pill tone="muted">Segera</Pill>
                    <input type="checkbox" checked disabled className="h-4 w-4 accent-jet" />
                  </span>
                </label>
              ))}
            </div>
            )}
          </div>
          <label className={LABEL_CLS}>
            Lebar Kertas
            <Select label="Ukuran kertas struk" value={form.paper} onChange={(v) => set('paper')(v)} options={[{ value: '58mm', label: '58 mm (Thermal)' }, { value: '80mm', label: '80 mm (Thermal)' }]} />
          </label>
          <label className={LABEL_CLS}>
            Pesan Footer
            <textarea
              value={form.receiptFooter}
              onChange={(e) => set('receiptFooter')(e.target.value.slice(0, 200))}
              rows={2}
              placeholder="Barang yang sudah dibeli tidak dapat ditukar atau dikembalikan."
              className={INPUT_CLS}
            />
            <span className="text-xs font-normal text-fog">{form.receiptFooter.length}/200</span>
          </label>
        </div>
        <FormActions dirty={dirty} busy={busy} onReset={onReset} onSave={onSave} />
      </SettingsCard>

      <div className="space-y-4 lg:col-span-2">
        <SettingsCard>
          <SettingsHead title="Preview Struk" sub="Ini adalah tampilan struk yang akan dicetak." />
          <div className="flex justify-center overflow-x-auto rounded-xl bg-surface p-4">
            <div className="bg-white px-3 py-4 font-mono text-[11px] leading-relaxed text-black" style={{ width: form.paper, maxWidth: '100%' }}>
              <p className="text-center text-[13px] font-bold uppercase">{form.storeName || '—'}</p>
              {form.address && <p className="mt-0.5 text-center text-[10px]">{form.address}</p>}
              {form.phone && <p className="text-center text-[10px]">{form.phone}</p>}
              <div className="receipt-hr my-2" aria-hidden="true" />
              <p>No. #TRX-00001</p>
              {(form.receiptShowCashier ?? true) && <p>Kasir: {session?.name ?? '—'}</p>}
              <p>Tanggal: {today}</p>
              <div className="receipt-hr my-2" aria-hidden="true" />
              {SAMPLE_ITEMS.map((i) => (
                <div key={i.name} className="mt-1">
                  <p className="break-words">{i.name}</p>
                  <div className="flex justify-between gap-2 tabular-nums">
                    <span>{i.qty} × {fmtRp(i.price)}</span>
                    <span>{fmtRp(i.qty * i.price)}</span>
                  </div>
                </div>
              ))}
              <div className="receipt-hr my-2" aria-hidden="true" />
              <div className="flex justify-between tabular-nums"><span>Subtotal</span><span>{fmtRp(sample.subtotal)}</span></div>
              {form.taxEnabled && (form.receiptShowTax ?? true) && <div className="flex justify-between tabular-nums"><span>Pajak{form.taxInclusive ? ' (inklusif)' : ''}</span><span>{fmtRp(sample.tax)}</span></div>}
              <div className="flex justify-between font-bold tabular-nums"><span>TOTAL</span><span>{fmtRp(sample.total)}</span></div>
              <div className="receipt-hr my-2" aria-hidden="true" />
              {(form.receiptShowNote ?? true) && form.receiptHeader && <p className="text-center text-[10px]">{form.receiptHeader}</p>}
              {(form.receiptShowNote ?? true) && form.receiptFooter && <p className="mt-1 text-center text-[10px]">{form.receiptFooter}</p>}
            </div>
          </div>
          <Button variant="ghost" className="mt-3 w-full" onClick={() => setPrintTest(true)}>
            <Printer className="size-4" /> Cetak Uji Coba
          </Button>
        </SettingsCard>
      </div>
    </div>
    <Modal open={printTest} title="Cetak Uji Coba" onClose={() => setPrintTest(false)}>
      <Receipt trx={sampleTrx} settings={form} onClose={() => setPrintTest(false)} />
    </Modal>
    </>
  )
}
