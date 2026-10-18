import { useEffect, useRef, useState } from 'react'
import type { Outlet, OutletSettingsPatch, StoreHours } from './api'
import { useDB } from '../../shared/session/store'
import { AlertStrip, Button, NumInput, Pill, Select, SettingsCard, SettingsHead } from '../../shared/ui'
import { defaultHours } from './components/shared'
import { useOutletSettings, useUpdateOutletSettings } from './hooks'
import { hasOutletOverride } from './outlet-settings-logic'

function hoursText(hours: StoreHours[]) {
  return hours.map((hour) => `${hour.days}: ${hour.open && hour.close ? `${hour.open}–${hour.close}` : 'Tutup'}`).join(' · ')
}

export default function OutletOverridesSection({ outlets, outletsLoading, outletsError }: { outlets: Outlet[]; outletsLoading: boolean; outletsError: string }) {
  const { session } = useDB()
  const [outletId, setOutletId] = useState(session?.outletId ?? '')
  const settings = useOutletSettings(outletId)
  const updateSettings = useUpdateOutletSettings()
  const [taxPct, setTaxPct] = useState('')
  const [receiptFooter, setReceiptFooter] = useState('')
  const [hours, setHours] = useState<StoreHours[]>(defaultHours())
  const [busy, setBusy] = useState(false)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const hydratedOutlet = useRef('')

  useEffect(() => {
    if (!outlets.length || outlets.some((outlet) => outlet.id === outletId)) return
    setOutletId(outlets.find((outlet) => outlet.id === session?.outletId)?.id ?? outlets[0].id)
  }, [outlets, outletId, session?.outletId])

  useEffect(() => {
    if (!settings.data || hydratedOutlet.current === outletId) return
    hydratedOutlet.current = outletId
    setTaxPct(String(settings.data.taxPct))
    setReceiptFooter(settings.data.receiptFooter)
    setHours(settings.data.hours?.length === 3 ? settings.data.hours : defaultHours())
  }, [outletId, settings.data])

  async function save(patch: OutletSettingsPatch) {
    if (!outletId || busy) return
    setBusy(true)
    setError('')
    setMessage('')
    try {
      await updateSettings.mutateAsync({ outletId, patch })
      const updated = await settings.reload()
      if (updated) {
        setTaxPct(String(updated.taxPct))
        setReceiptFooter(updated.receiptFooter)
        setHours(updated.hours?.length === 3 ? updated.hours : defaultHours())
      }
      setMessage('Pengaturan outlet berhasil disimpan.')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Gagal menyimpan pengaturan outlet.')
    } finally {
      setBusy(false)
    }
  }

  function updateHour(index: number, patch: Partial<StoreHours>) {
    setHours((current) => current.map((hour, i) => i === index ? { ...hour, ...patch } : hour))
  }

  const current = settings.data
  const taxOverride = current ? hasOutletOverride(current.overrides, 'taxPct') : false
  const footerOverride = current ? hasOutletOverride(current.overrides, 'receiptFooter') : false
  const hoursOverride = current ? hasOutletOverride(current.overrides, 'hours') : false

  return (
    <SettingsCard className="mt-4">
      <SettingsHead title="Pengaturan Outlet" sub="Pajak, footer struk, dan jam operasional dapat mengikuti toko atau diatur khusus per outlet." />
      {outletsError && <AlertStrip className="mb-4">{outletsError}</AlertStrip>}
      {error && <AlertStrip className="mb-4">{error}</AlertStrip>}
      {message && <p className="mb-4 rounded-lg bg-surface px-3.5 py-2.5 text-[13px] text-sprout" role="status">{message}</p>}

      <div className="max-w-sm">
        <Select
          label="Pilih outlet"
          value={outletId}
          onChange={(value) => { setOutletId(value); setMessage(''); setError('') }}
          disabled={outletsLoading || outlets.length === 0}
          options={outlets.map((outlet) => ({ value: outlet.id, label: outlet.name }))}
        />
      </div>

      {outletsLoading && <p className="mt-4 text-sm text-muted">Memuat outlet…</p>}
      {!outletsLoading && outlets.length === 0 && <p className="mt-4 text-sm text-muted">Tambahkan outlet untuk mengatur nilai khusus.</p>}
      {outletId && settings.loading && <p className="mt-4 text-sm text-muted">Memuat pengaturan outlet…</p>}
      {settings.err && <AlertStrip className="mt-4">{settings.err}</AlertStrip>}

      {current && (
        <div className="mt-4 divide-y divide-dove border-t border-dove">
          <div className="py-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h3 className="text-sm font-semibold text-fg">Tarif pajak (%)</h3>
              <Pill tone={taxOverride ? 'warn' : 'muted'}>{taxOverride ? 'Khusus outlet ini' : 'Mengikuti pengaturan toko'}</Pill>
            </div>
            <p className="mt-1 text-[13px] text-muted">Nilai efektif sekarang: {current.taxPct}%</p>
            <div className="mt-3 max-w-xs">
              <NumInput label="Tarif pajak outlet (%)" value={taxPct} onValue={setTaxPct} allowDecimal />
            </div>
            <div className="mt-3 flex flex-wrap gap-2">
              <Button disabled={busy} onClick={() => {
                if (!taxPct.trim() || !Number.isFinite(Number(taxPct))) return setError('Tarif pajak wajib diisi dengan angka.')
                void save({ taxPct: Number(taxPct) })
              }}>Simpan Pajak</Button>
              <Button variant="ghost" disabled={busy || !taxOverride} onClick={() => void save({ taxPct: null })}>Ikuti toko</Button>
            </div>
          </div>

          <div className="py-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h3 className="text-sm font-semibold text-fg">Footer struk</h3>
              <Pill tone={footerOverride ? 'warn' : 'muted'}>{footerOverride ? 'Khusus outlet ini' : 'Mengikuti pengaturan toko'}</Pill>
            </div>
            <p className="mt-1 break-words text-[13px] text-muted">Nilai efektif sekarang: {current.receiptFooter || 'Kosong'}</p>
            <label className="mt-3 flex flex-col gap-1.5 text-[13px] font-medium text-steel">
              Footer struk outlet
              <textarea value={receiptFooter} maxLength={200} rows={2} onChange={(event) => setReceiptFooter(event.target.value)} className="min-h-[42px] w-full rounded-md border border-border bg-paper px-3.5 py-2 text-[15px] text-fg focus:border-jet focus:outline-2 focus:outline-accent-soft" />
            </label>
            <div className="mt-3 flex flex-wrap gap-2">
              <Button disabled={busy} onClick={() => void save({ receiptFooter })}>Simpan Footer</Button>
              <Button variant="ghost" disabled={busy || !footerOverride} onClick={() => void save({ receiptFooter: null })}>Ikuti toko</Button>
            </div>
          </div>

          <div className="py-4">
            <div className="flex flex-wrap items-center justify-between gap-2">
              <h3 className="text-sm font-semibold text-fg">Jam operasional</h3>
              <Pill tone={hoursOverride ? 'warn' : 'muted'}>{hoursOverride ? 'Khusus outlet ini' : 'Mengikuti pengaturan toko'}</Pill>
            </div>
            <p className="mt-1 break-words text-[13px] text-muted">Nilai efektif sekarang: {hoursText(current.hours?.length === 3 ? current.hours : defaultHours())}</p>
            <p className="mt-1 text-xs text-fog">Hari tanpa centang berarti tutup.</p>
            <div className="mt-3 space-y-2">
              {hours.map((hour, index) => {
                const open = hour.open != null && hour.close != null
                return (
                  <div key={hour.days} className="flex flex-wrap items-center gap-2 text-sm">
                    <label className="flex min-w-32 cursor-pointer items-center gap-2 text-fg">
                      <input
                        type="checkbox"
                        checked={open}
                        onChange={(event) => updateHour(index, event.target.checked ? { open: '08:00', close: '21:00' } : { open: null, close: null })}
                        className="h-4 w-4 accent-jet"
                      />
                      {hour.days}
                    </label>
                    {open ? (
                      <>
                        <input type="time" value={hour.open ?? '08:00'} onChange={(event) => updateHour(index, { open: event.target.value })} aria-label={`Jam buka ${hour.days}`} className="h-[42px] w-28 rounded-md border border-border bg-paper px-2 font-mono text-[13px] tabular-nums" />
                        <span className="text-fog">—</span>
                        <input type="time" value={hour.close ?? '21:00'} onChange={(event) => updateHour(index, { close: event.target.value })} aria-label={`Jam tutup ${hour.days}`} className="h-[42px] w-28 rounded-md border border-border bg-paper px-2 font-mono text-[13px] tabular-nums" />
                      </>
                    ) : <span className="text-[13px] text-fog">Tutup</span>}
                  </div>
                )
              })}
            </div>
            <div className="mt-3 flex flex-wrap gap-2">
              <Button disabled={busy} onClick={() => void save({ hours })}>Simpan Jam</Button>
              <Button variant="ghost" disabled={busy || !hoursOverride} onClick={() => void save({ hours: null })}>Ikuti toko</Button>
            </div>
          </div>
        </div>
      )}
    </SettingsCard>
  )
}
