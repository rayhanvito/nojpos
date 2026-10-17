import type { StoreHours, StoreSettings } from '../../../shared/api'
import { ImagePlus, Info, Mail, MapPin, Phone } from 'lucide-react'
import { Input, NumInput, Pill, Select, SettingsCard, SettingsHead } from '../../../shared/ui'
import { BUSINESS_TYPES, FormActions, INPUT_CLS, LABEL_CLS, Soon, TIMEZONES, defaultHours, tzLabel } from './shared'

export function TokoTab({ form, set, setV, hasExt, hasLowStockDefault, storeComplete, dirty, busy, onSave, onReset }: {
  form: StoreSettings
  set: (k: keyof StoreSettings) => (v: string) => void
  setV: <K extends keyof StoreSettings>(k: K, v: StoreSettings[K]) => void
  hasExt: boolean
  hasLowStockDefault: boolean
  storeComplete: boolean
  dirty: boolean
  busy: boolean
  onSave: () => void
  onReset: () => void
}) {
  return (
    <>
    <div className="grid items-start gap-4 lg:grid-cols-3">
      <SettingsCard className="lg:col-span-2">
        <SettingsHead title="Informasi Toko" sub="Lengkapi informasi toko Anda." />
        <div className="grid gap-4 sm:grid-cols-2">
          <Input label="Nama Toko *" value={form.storeName} onChange={set('storeName')} placeholder="Toko Andika" />
          {hasExt ? (
            <label className={LABEL_CLS}>
              Jenis Usaha *
              <Select label="Jenis usaha" value={form.businessType ?? ''} onChange={(v) => setV('businessType', v)} options={BUSINESS_TYPES} />
            </label>
          ) : (
          <Soon>
            <label className={LABEL_CLS}>
              Jenis Usaha *
              <Select label="Belum tersedia" disabled value="" onChange={() => {}} options={[{ value: '', label: 'Menunggu backend' }]} />
            </label>
          </Soon>
          )}
          <Input label="No. Telepon" value={form.phone} onChange={set('phone')} placeholder="0812…" />
          {hasExt ? (
            <label className={LABEL_CLS}>
              Email
              <input type="email" value={form.email ?? ''} onChange={(e) => setV('email', e.target.value)} placeholder="toko@gmail.com" className={INPUT_CLS} />
            </label>
          ) : (
          <Soon>
            <label className={LABEL_CLS}>
              Email
              <input disabled className={INPUT_CLS} value="" placeholder="Menunggu backend" onChange={() => {}} />
            </label>
          </Soon>
          )}
          <label className={`${LABEL_CLS} sm:col-span-2`}>
            Alamat
            <textarea value={form.address} onChange={(e) => set('address')(e.target.value)} rows={3} placeholder="Jl. …" className={INPUT_CLS} />
          </label>
          {hasExt ? (
            <label className={LABEL_CLS}>
              Kota
              <input value={form.city ?? ''} onChange={(e) => setV('city', e.target.value)} placeholder="Makassar" className={INPUT_CLS} />
            </label>
          ) : (
          <Soon>
            <label className={LABEL_CLS}>
              Kota
              <input disabled className={INPUT_CLS} value="" placeholder="Menunggu backend" onChange={() => {}} />
            </label>
          </Soon>
          )}
          {hasExt ? (
            <label className={LABEL_CLS}>
              Provinsi
              <input value={form.province ?? ''} onChange={(e) => setV('province', e.target.value)} placeholder="Sulawesi Selatan" className={INPUT_CLS} />
            </label>
          ) : (
          <Soon>
            <label className={LABEL_CLS}>
              Provinsi
              <Select label="Belum tersedia" disabled value="" onChange={() => {}} options={[{ value: '', label: 'Menunggu backend' }]} />
            </label>
          </Soon>
          )}
          {hasExt ? (
            <label className={LABEL_CLS}>
              Mata Uang
              <input value={form.currency ?? 'IDR'} onChange={(e) => setV('currency', e.target.value.toUpperCase().slice(0, 3))} placeholder="IDR" maxLength={3} className={`${INPUT_CLS} font-mono uppercase`} />
            </label>
          ) : (
          <Soon>
            <label className={LABEL_CLS}>
              Mata Uang
              <Select label="Mata uang" disabled value="IDR" onChange={() => {}} options={[{ value: 'IDR', label: 'Rupiah (IDR)' }]} />
            </label>
          </Soon>
          )}
          <label className={LABEL_CLS}>
            Timezone
            <Select label="Zona waktu" value={form.timezone} onChange={(v) => set('timezone')(v)} options={[...(TIMEZONES.some((t) => t.value === form.timezone) ? [] : [{ value: form.timezone, label: form.timezone }]), ...TIMEZONES]} />
          </label>
        </div>
        <div className="mt-5">
          <div className="mb-1 flex items-center gap-2">
            <p className="text-[13px] font-medium text-steel">Jam Operasional</p>
            {!hasExt && <Pill tone="muted">Segera</Pill>}
          </div>
          {hasExt ? (
          <>
          <p className="mb-2 text-xs text-fog">Atur jam operasional toko Anda. Hari tanpa centang = tutup.</p>
          <div className="space-y-2">
            {((form.hours?.length === 3 ? form.hours : defaultHours())).map((h, i) => {
              const on = h.open != null && h.close != null
              const upd = (patch: Partial<StoreHours>) => {
                const cur = form.hours?.length === 3 ? form.hours : defaultHours()
                setV('hours', cur.map((r, ix) => (ix === i ? { ...r, ...patch } : r)))
              }
              return (
              <div key={h.days} className="flex flex-wrap items-center gap-2 text-sm">
                <label className="flex min-w-32 cursor-pointer items-center gap-2 text-fg">
                  <input
                    type="checkbox" checked={on}
                    onChange={(e) => upd(e.target.checked ? { open: '08:00', close: '21:00' } : { open: null, close: null })}
                    className="h-4 w-4 accent-jet"
                  /> {h.days}
                </label>
                {on ? (
                  <>
                    <input type="time" value={h.open ?? '08:00'} onChange={(e) => upd({ open: e.target.value })} aria-label="Jam buka" className="w-24 rounded-md border border-border bg-paper px-2 py-1.5 font-mono text-[13px] tabular-nums" />
                    <span className="text-fog">—</span>
                    <input type="time" value={h.close ?? '21:00'} onChange={(e) => upd({ close: e.target.value })} aria-label="Jam tutup" className="w-24 rounded-md border border-border bg-paper px-2 py-1.5 font-mono text-[13px] tabular-nums" />
                  </>
                ) : (
                  <span className="text-[13px] text-fog">Tutup</span>
                )}
              </div>
              )
            })}
          </div>
          </>
          ) : (
          <>
          <p className="mb-2 text-xs text-fog">Atur jam operasional toko Anda. Segera hadir — menunggu backend.</p>
          <div className="space-y-2 opacity-80">
            {[['Senin – Jumat', true], ['Sabtu', true], ['Minggu', false]].map(([day, on]) => (
              <div key={day as string} className="flex flex-wrap items-center gap-2 text-sm">
                <label className="flex min-w-32 cursor-not-allowed items-center gap-2 text-fg">
                  <input type="checkbox" checked={!!on} disabled className="h-4 w-4 accent-jet" /> {day}
                </label>
                {on ? (
                  <>
                    <input disabled value="08:00" onChange={() => {}} aria-label="Jam buka" className="w-24 rounded-md border border-border bg-paper px-2 py-1.5 font-mono text-[13px] tabular-nums" />
                    <span className="text-fog">—</span>
                    <input disabled value="21:00" onChange={() => {}} aria-label="Jam tutup" className="w-24 rounded-md border border-border bg-paper px-2 py-1.5 font-mono text-[13px] tabular-nums" />
                  </>
                ) : (
                  <span className="text-[13px] text-fog">Tutup</span>
                )}
              </div>
            ))}
          </div>
          </>
          )}
        </div>
        <div className="mt-5 border-t border-dove pt-4">
          <SettingsHead title="Kredit Pelanggan" sub="Atur jarak jatuh tempo piutang sejak tanggal transaksi." />
          <NumInput
            label="Jatuh tempo kredit (hari)"
            value={form.creditDueDays ?? 14}
            onValue={(value) => setV('creditDueDays', Number(value || 0))}
            hint="Default 14 hari."
          />
        </div>
        <div className="mt-5 border-t border-dove pt-4">
          <SettingsHead title="Loyalty Pelanggan" sub="Atur perolehan dan penukaran poin pelanggan." />
          <div className="space-y-4">
            <label className="flex cursor-pointer items-center justify-between gap-3 text-sm text-fg">
              Aktifkan loyalty
              <input
                type="checkbox"
                checked={!!form.loyaltyEnabled}
                onChange={(e) => setV('loyaltyEnabled', e.target.checked)}
                className="h-4 w-4 accent-jet"
              />
            </label>
            <NumInput
              label="Perolehan (Rp per 1 poin)"
              value={form.loyaltyEarnPer ?? 10000}
              onValue={(value) => setV('loyaltyEarnPer', Number(value || 0))}
              hint="Nilai transaksi setelah pajak untuk memperoleh 1 poin."
            />
            <NumInput
              label="Nilai tukar (Rp per poin)"
              value={form.loyaltyPointValue ?? 100}
              onValue={(value) => setV('loyaltyPointValue', Number(value || 0))}
              hint="Nilai rupiah untuk setiap poin yang ditukar saat checkout."
            />
          </div>
        </div>
        <div className="mt-5 border-t border-dove pt-4">
          <SettingsHead title="Inventaris" sub="Atur nilai stok minimum bawaan untuk produk dan varian baru." />
          {hasLowStockDefault ? (
            <NumInput
              label="Stok minimum bawaan"
              value={form.lowStockDefault ?? 0}
              onValue={(value) => setV('lowStockDefault', Number(value || 0))}
              hint="Stok 0 menonaktifkan ambang stok menipis."
            />
          ) : (
            <Soon>
              <NumInput label="Stok minimum bawaan" value="0" onValue={() => {}} disabled hint="Segera hadir — menunggu backend." />
            </Soon>
          )}
        </div>
        <FormActions dirty={dirty} busy={busy} onReset={onReset} onSave={onSave} />
      </SettingsCard>

      <div className="space-y-4">
        <SettingsCard>
          <SettingsHead title="Preview Profil Toko" sub="Ini adalah tampilan informasi toko Anda." />
          <div className="flex items-center gap-3">
            <span className="grid size-12 shrink-0 place-items-center rounded-xl bg-surface text-fog"><ImagePlus className="size-5" /></span>
            <div className="min-w-0">
              <p className="truncate text-[15px] font-medium text-fg">{form.storeName || '—'}</p>
            </div>
          </div>
          <div className="mt-4 space-y-1.5 text-[13px]">
            <p className="flex items-center gap-2 text-muted"><Phone className="size-3.5 shrink-0" />{form.phone || '—'}</p>
            <p className="flex items-center gap-2 text-muted"><Mail className="size-3.5 shrink-0" />{form.email || '—'}</p>
            <p className="flex items-start gap-2 text-muted"><MapPin className="mt-0.5 size-3.5 shrink-0" /><span>{[form.address, form.city, form.province].filter(Boolean).join(', ') || '—'}</span></p>
          </div>
          <div className="mt-3 flex flex-wrap gap-1.5">
            <Pill tone="muted">{tzLabel(form.timezone)}</Pill>
            {!!form.businessType && <Pill tone="muted">{BUSINESS_TYPES.find((t) => t.value === form.businessType)?.label ?? form.businessType}</Pill>}
            {!!form.currency && <Pill tone="muted">{form.currency}</Pill>}
          </div>
        </SettingsCard>
        <div className={`flex items-start gap-2 rounded-xl border px-4 py-3 text-[13px] ${storeComplete ? 'border-sprout/40 bg-success-bg text-sprout' : 'border-dove bg-surface text-muted'}`}>
          <Info className="mt-0.5 size-4 shrink-0" aria-hidden="true" />
          <div>
            <p className="font-medium text-fg">{storeComplete ? 'Informasi toko lengkap' : 'Informasi toko belum lengkap'}</p>
            <p className="mt-0.5">{storeComplete ? 'Semua informasi penting sudah diisi dengan benar.' : 'Lengkapi nama toko yang masih diperlukan.'}</p>
          </div>
        </div>
      </div>
    </div>
    </>
  )
}
