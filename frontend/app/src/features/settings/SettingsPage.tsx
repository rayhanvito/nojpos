import { useEffect, useRef, useState } from 'react'
import type { StoreSettings } from './api'
import { useSettings, useUpdateSettings } from './hooks'
import { PageHeader } from '../../shared/PageHeader'
import { getSession, setSession } from '../../shared/session/store'
import { AlertStrip, Button } from '../../shared/ui'
import { Tabs, TabsList, TabsTrigger } from '@/shared/ui/primitives/tabs'
import { Skeleton } from '@/shared/ui/primitives/skeleton'
import OutletSettingsPage from './OutletSettingsPage'
import { AkunTab } from './components/AkunTab'
import { PajakTab } from './components/PajakTab'
import { StrukTab } from './components/StrukTab'
import { TokoTab } from './components/TokoTab'

const TABS = [
  { id: 'akun', label: 'Akun' },
  { id: 'toko', label: 'Toko' },
  { id: 'struk', label: 'Struk' },
  { id: 'pajak', label: 'Pajak' },
  { id: 'outlet', label: 'Outlet' },
] as const

type TabId = (typeof TABS)[number]['id']

export default function SettingsPage() {
  const settings = useSettings()
  const updateSettings = useUpdateSettings()
  const [tab, setTab] = useState<TabId>('akun')
  const [form, setForm] = useState<StoreSettings | null>(null)
  const [saved, setSaved] = useState<StoreSettings | null>(null)
  const [msg, setMsg] = useState('')
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)

  // Form lokal diisi sekali dari hasil query pengaturan (setara fetch
  // lama saat mount); sesudah itu form milik halaman sampai disimpan,
  // agar refetch latar tidak menimpa suntingan yang belum disimpan.
  const hydrated = useRef(false)
  useEffect(() => {
    if (settings.data && !hydrated.current) {
      hydrated.current = true
      setForm(settings.data); setSaved(settings.data)
    }
  }, [settings.data])

  const loadErr = form ? '' : settings.err

  function retry() {
    setErr('')
    settings.reload()
  }

  const set = (k: keyof StoreSettings) => (v: string) => setForm((f) => (f ? { ...f, [k]: v } : f))
  const setV = <K extends keyof StoreSettings>(k: K, v: StoreSettings[K]) => setForm((f) => (f ? { ...f, [k]: v } : f))
  // Dukungan backend dideteksi dari keberadaan key (kontrak §9): backend lama
  // tanpa key → field tampil nonaktif "Segera" seperti semula.
  const hasExt = !!saved && typeof saved === 'object' && 'businessType' in saved
  const hasLowStockDefault = saved?.lowStockDefault !== undefined
  const dirty = !!form && !!saved && JSON.stringify(form) !== JSON.stringify(saved)

  async function save() {
    if (!form) return
    setMsg(''); setErr(''); setBusy(true)
    try {
      const s = await updateSettings.mutateAsync(form)
      setForm(s); setSaved(s)
      const cur = getSession()
      if (cur && cur.store !== s.storeName) setSession({ ...cur, store: s.storeName })
      setMsg('Pengaturan berhasil disimpan.')
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Gagal menyimpan perubahan. Coba lagi.')
    } finally {
      setBusy(false)
    }
  }

  function resetForm() {
    if (saved) { setForm(saved); setMsg(''); setErr('') }
  }

  const storeComplete = (form?.storeName.trim() ?? '') !== ''

  if (!form) {
    return (
      <>
        <PageHeader title="Pengaturan" crumb="Pengaturan" />
        {loadErr ? (
          <div className="max-w-2xl rounded-2xl bg-cream p-4 text-center sm:p-6">
            <p className="text-sm text-ember">Gagal memuat pengaturan.</p>
            <p className="mt-1 text-[13px] text-muted">Coba muat ulang halaman.</p>
            <Button className="mt-4" onClick={retry}>Coba Lagi</Button>
          </div>
        ) : (
          <>
            <div className="mb-5 flex gap-2 overflow-x-auto pb-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden" aria-hidden="true">
              {['w-[76px]', 'w-[72px]', 'w-[76px]', 'w-[72px]'].map((w, i) => (
                <Skeleton key={i} className={`h-9 shrink-0 rounded-full ${w}`} />
              ))}
            </div>
            <div className="grid items-start gap-4 lg:grid-cols-2" aria-busy="true" aria-label="Memuat pengaturan">
              <div className="rounded-2xl bg-cream p-4 sm:p-6">
                <div className="flex flex-wrap items-center gap-3">
                  <Skeleton className="size-11 shrink-0 rounded-full" />
                  <div className="min-w-0 flex-1 basis-32 space-y-2">
                    <Skeleton className="h-4 w-32 max-w-full" />
                    <Skeleton className="h-3 w-48 max-w-full" />
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    <Skeleton className="h-5 w-14 rounded-full" />
                    <Skeleton className="h-5 w-20 rounded-full" />
                  </div>
                </div>
                <div className="mt-4 divide-y divide-dove border-t border-dove">
                  {[0, 1, 2].map((i) => (
                    <div key={i} className="flex items-center justify-between gap-3 py-2.5">
                      <Skeleton className="h-3.5 w-24" />
                      <Skeleton className="h-3.5 w-28" />
                    </div>
                  ))}
                </div>
                <div className="mt-5 flex flex-col gap-2 sm:flex-row">
                  <Skeleton className="h-12 w-full rounded-full sm:w-40" />
                  <Skeleton className="h-12 w-full rounded-full sm:w-28" />
                </div>
              </div>
              <div className="rounded-2xl bg-cream p-4 sm:p-6">
                <div className="flex items-center gap-2">
                  <Skeleton className="size-4 shrink-0 rounded" />
                  <div className="min-w-0 flex-1 space-y-1.5">
                    <Skeleton className="h-4 w-44 max-w-full" />
                    <Skeleton className="h-3 w-56 max-w-full" />
                  </div>
                </div>
                <div className="mt-4 divide-y divide-dove rounded-xl border border-dove px-4">
                  {[0, 1].map((i) => (
                    <div key={i} className="flex items-center justify-between gap-3 py-3">
                      <Skeleton className="h-3.5 w-28" />
                      <Skeleton className="h-5 w-16 rounded-full" />
                    </div>
                  ))}
                </div>
                <div className="mt-4 flex flex-col gap-2 sm:flex-row sm:flex-wrap">
                  <Skeleton className="h-12 w-full rounded-full sm:w-40" />
                  <Skeleton className="h-12 w-full rounded-full sm:w-44" />
                </div>
              </div>
            </div>
          </>
        )}
      </>
    )
  }

  return (
    <>
      <PageHeader title="Pengaturan" crumb="Pengaturan" />

      {err && <AlertStrip className="mb-4">{err}</AlertStrip>}
      {msg && <p className="mb-4 rounded-lg bg-surface px-3.5 py-2.5 text-[13px] text-sprout">{msg}</p>}

      <Tabs value={tab} onValueChange={(v) => setTab(v as TabId)} className="mb-5">
        <TabsList variant="line" aria-label="Subhalaman pengaturan">
          {TABS.map((t) => (
            <TabsTrigger key={t.id} value={t.id}>{t.label}</TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      {tab === 'akun' && <AkunTab busy={busy} setBusy={setBusy} setMsg={setMsg} setErr={setErr} />}

      {tab === 'outlet' && <OutletSettingsPage />}

      {tab === 'toko' && <TokoTab form={form} set={set} setV={setV} hasExt={hasExt} hasLowStockDefault={hasLowStockDefault} dirty={dirty} busy={busy} storeComplete={storeComplete} onSave={save} onReset={resetForm} />}
      {tab === 'struk' && <StrukTab form={form} set={set} setV={setV} hasExt={hasExt} dirty={dirty} busy={busy} onSave={save} onReset={resetForm} />}
      {tab === 'pajak' && <PajakTab form={form} setV={setV} setForm={setForm} hasExt={hasExt} dirty={dirty} busy={busy} onSave={save} onReset={resetForm} />}
    </>
  )
}

