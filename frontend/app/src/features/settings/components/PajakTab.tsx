import type { Dispatch, SetStateAction } from 'react'
import type { StoreSettings } from '../api'
import { useSettingsTransactions } from '../hooks'
import { fmtRp } from '../../../shared/session/store'
import { Info } from 'lucide-react'
import { NumInput, Select, SettingsCard, SettingsHead, StatCard } from '../../../shared/ui'
import { FormActions, INPUT_CLS, LABEL_CLS, Soon, calcTax, todayStr } from './shared'

export function PajakTab({ form, setV, setForm, hasExt, dirty, busy, onSave, onReset }: {
  form: StoreSettings
  setForm: Dispatch<SetStateAction<StoreSettings | null>>
  setV: <K extends keyof StoreSettings>(k: K, v: StoreSettings[K]) => void
  hasExt: boolean
  dirty: boolean
  busy: boolean
  onSave: () => void
  onReset: () => void
}) {
  const today = todayStr()
  const todayTrx = useSettingsTransactions(today)
  const taxed = todayTrx.data?.items.filter((t) => t.tax > 0).length ?? 0
  const trxTotal = todayTrx.data?.total ?? 0
  return (
    <>
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3 sm:gap-3 xl:grid-cols-4">
        {[
          { label: 'Pajak Aktif', value: form.taxEnabled ? 'Ya' : 'Tidak', sub: 'Pajak sedang digunakan' },
          { label: 'Tarif Pajak', value: `${form.taxPct}%`, sub: form.taxName || 'Tarif pajak saat ini' },
          { label: 'Harga Sudah Termasuk Pajak', value: !hasExt ? '—' : form.taxInclusive ? 'Ya' : 'Tidak', sub: !hasExt ? 'Segera hadir' : form.taxInclusive ? 'Pajak di dalam harga' : 'Pajak di luar harga' },
          { label: 'Transaksi Kena Pajak Hari Ini', value: String(taxed), sub: `Dari ${trxTotal} transaksi` },
        ].map((k) => (
          <StatCard key={k.label} label={k.label} value={k.value} sub={k.sub} />
        ))}
      </div>

      <div className="grid items-start gap-4 lg:grid-cols-3">
        <SettingsCard className="lg:col-span-2">
          <SettingsHead title="Pengaturan Pajak" sub="Atur konfigurasi pajak untuk transaksi di toko Anda." />
          <div className="space-y-4">
            {hasExt ? (
            <label className={LABEL_CLS}>
              Nama Pajak
              <input value={form.taxName ?? ''} onChange={(e) => setV('taxName', e.target.value.slice(0, 20))} placeholder="cth: PPN" maxLength={20} className={INPUT_CLS} />
            </label>
            ) : (
            <Soon>
              <label className={LABEL_CLS}>
                Nama Pajak
                <input disabled className={INPUT_CLS} value="" placeholder="cth: PPN" onChange={() => {}} />
              </label>
            </Soon>
            )}
            <NumInput allowDecimal label="Tarif Pajak (%)" value={form.taxPct} onValue={(r) => setForm((f) => (f ? { ...f, taxPct: r === '' ? 0 : Number(r.replace(',', '.')) } : f))} />
            {hasExt ? (
            <label className="flex cursor-pointer items-center justify-between gap-3 text-sm text-fg">
              Harga Sudah Termasuk Pajak
              <input type="checkbox" checked={!!form.taxInclusive} onChange={(e) => setV('taxInclusive', e.target.checked)} className="h-4 w-4 accent-jet" />
            </label>
            ) : (
            <Soon>
              <label className="flex cursor-not-allowed items-center justify-between gap-3 text-sm text-fg">
                Harga Sudah Termasuk Pajak
                <input type="checkbox" disabled className="h-4 w-4 accent-jet" />
              </label>
            </Soon>
            )}
            {hasExt ? (
            <label className={LABEL_CLS}>
              Pembulatan Pajak
              <Select label="Pembulatan pajak" value={form.taxRounding ?? 'none'} onChange={(v) => setV('taxRounding', v)} options={[{ value: 'none', label: 'Normal (setengah ke atas)' }, { value: 'down', label: 'Ke bawah (floor)' }, { value: 'up', label: 'Ke atas (ceil)' }]} />
            </label>
            ) : (
            <Soon>
              <label className={LABEL_CLS}>
                Pembulatan Pajak
                <Select label="Belum tersedia" disabled value="" onChange={() => {}} options={[{ value: '', label: 'Menunggu backend' }]} />
              </label>
            </Soon>
            )}
            {hasExt ? (
            <label className={LABEL_CLS}>
              Terapkan Pajak Pada
              <Select label="Pajak berlaku untuk" value={form.taxApplyTo ?? 'all'} onChange={(v) => setV('taxApplyTo', v)} options={[{ value: 'all', label: 'Semua produk' }]} />
            </label>
            ) : (
            <Soon>
              <label className={LABEL_CLS}>
                Terapkan Pajak Pada
                <Select label="Belum tersedia" disabled value="" onChange={() => {}} options={[{ value: '', label: 'Menunggu backend' }]} />
              </label>
            </Soon>
            )}
            <label className="flex cursor-pointer items-center justify-between gap-3 text-sm text-fg">
              Status Pajak
              <input type="checkbox" checked={form.taxEnabled} onChange={(e) => setForm((f) => (f ? { ...f, taxEnabled: e.target.checked } : f))} className="h-4 w-4 accent-jet" />
            </label>
            <p className="-mt-2 text-xs text-fog">Aktifkan atau nonaktifkan perhitungan pajak.</p>
          </div>
          <FormActions dirty={dirty} busy={busy} onReset={onReset} onSave={onSave} />
        </SettingsCard>

        <div className="space-y-4">
          <SettingsCard>
            <SettingsHead title="Contoh Perhitungan Pajak" sub="Simulasi perhitungan pada transaksi." />
            {(() => {
              const ex = calcTax(100000, form.taxPct, !!form.taxInclusive, form.taxRounding ?? 'none')
              return (
              <>
              <div className="space-y-1 font-mono text-[13px] tabular-nums">
                <div className="flex justify-between"><span className="font-sans text-muted">Harga Produk</span><span>{fmtRp(100000)}</span></div>
                <div className="flex justify-between"><span className="font-sans text-muted">Pajak{form.taxName ? ` (${form.taxName} ${form.taxPct}%)` : ` (${form.taxPct}%)`}</span><span>{fmtRp(ex.tax)}</span></div>
                <div className="my-2 border-t border-dashed border-dove" />
                <div className="flex justify-between font-bold"><span className="font-sans">Total</span><span>{fmtRp(ex.total)}</span></div>
              </div>
              <div className="mt-4 flex items-start gap-2 rounded-lg bg-surface px-3.5 py-3 text-[13px] text-muted">
                <Info className="mt-0.5 size-4 shrink-0 text-steel" aria-hidden="true" />
                <p>{form.taxInclusive ? 'Harga sudah termasuk pajak: total tetap, pajak dihitung dari dalam harga.' : 'Karena harga belum termasuk pajak, maka pajak akan ditambahkan ke total transaksi.'}</p>
              </div>
              </>
              )
            })()}
          </SettingsCard>
          <SettingsCard>
            <SettingsHead title="Aktivitas Pajak Terkini" sub="Log perubahan pengaturan pajak." />
            <p className="py-4 text-center text-sm text-fog">Belum ada aktivitas pengaturan.</p>
          </SettingsCard>
        </div>
      </div>
    </div>
    </>
  )
}
