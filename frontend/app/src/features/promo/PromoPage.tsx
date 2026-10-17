import { useMemo, useState } from 'react'
import { Pencil, Plus, Power, Trash2 } from 'lucide-react'
import {
  apiListCategories, apiListProducts, fetchAll, type Category, type Product,
} from '../../shared/api'
import { PageHeader } from '../../shared/PageHeader'
import { useCache } from '../../shared/cache'
import { fmtRp, useDB } from '../../shared/session/store'
import { useCreatePromo, useDeletePromo, usePromos, useSetPromoActive, useUpdatePromo } from './hooks'
import type { Promo, PromoInput } from './api'
import { CheckRow, DateTimeField, ModalActions, AlertStrip, Button, CardGrid, DataCard, Empty, FilterChips, FilterField, FilterPanel, Input, Modal, NumInput, TimePicker, Pager, Pill, RowAction, SearchInput, Select, SkeletonRows, TableWrap, Td, Th } from '../../shared/ui'
import { Skeleton } from '@/shared/ui/primitives/skeleton'
import { Tabs, TabsList, TabsTrigger } from '@/shared/ui/primitives/tabs'

interface Draft {
  id?: number
  name: string
  type: 'percent' | 'amount'
  value: string
  scope: 'all' | 'category' | 'product'
  categoryId: string
  productId: string
  code: string
  startsAt: string
  endsAt: string
  happyStart: string
  happyEnd: string
  minPurchase: string
  maxDiscount: string
  active: boolean
  combinable: boolean
}

const emptyDraft: Draft = {
  name: '', type: 'percent', value: '', scope: 'all', categoryId: '', productId: '', code: '',
  startsAt: '', endsAt: '', happyStart: '', happyEnd: '', minPurchase: '0', maxDiscount: '0', active: true, combinable: false,
}

function localDateTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return ''
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`
}

function toDateTime(value: string) {
  return value ? new Date(value).toISOString() : ''
}

function formatDateTime(value: string) {
  return new Date(value).toLocaleString('id-ID', { dateStyle: 'medium', timeStyle: 'short' })
}

function draftFromPromo(promo: Promo): Draft {
  return {
    id: promo.id,
    name: promo.name,
    type: promo.type,
    value: String(promo.value),
    scope: promo.scope,
    categoryId: promo.category_id === null ? '' : String(promo.category_id),
    productId: promo.product_id === null ? '' : String(promo.product_id),
    code: promo.voucher_code ?? '',
    startsAt: localDateTime(promo.starts_at),
    endsAt: localDateTime(promo.ends_at),
    happyStart: promo.happy_start ?? '',
    happyEnd: promo.happy_end ?? '',
    minPurchase: String(promo.min_purchase),
    maxDiscount: String(promo.max_discount),
    active: promo.active,
    combinable: promo.combinable,
  }
}

export default function Promo() {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const promos = usePromos()
  const createPromo = useCreatePromo()
  const updatePromo = useUpdatePromo()
  const deletePromo = useDeletePromo()
  const setPromoActive = useSetPromoActive()
  const categories = useCache<Category[]>(`promo-categories:${who}`, () => apiListCategories(), 'Gagal memuat kategori.')
  const products = useCache<Product[]>(`promo-products:${who}`, () => fetchAll<Product>((page) => apiListProducts({ page, limit: 200 })), 'Gagal memuat produk.')
  const [tab, setTab] = useState<'promo' | 'voucher'>('promo')
  const [statusF, setStatusF] = useState('')
  const [periodF, setPeriodF] = useState('')
  const [editKind, setEditKind] = useState<'promo' | 'voucher'>('promo')
  const [q, setQ] = useState('')
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const [err, setErr] = useState('')
  const [formErr, setFormErr] = useState('')
  const [deleteErr, setDeleteErr] = useState('')
  const [editing, setEditing] = useState<Draft | null>(null)
  const [deleteFor, setDeleteFor] = useState<Promo | null>(null)
  const [busy, setBusy] = useState(false)
  const [busyId, setBusyId] = useState<number | null>(null)

  const promoItems = promos.data ?? []
  const categoryItems = categories.data ?? []
  const productItems = products.data ?? []
  const kindItems = useMemo(() => promoItems.filter((promo) => (tab === 'voucher' ? !!promo.voucher_code : !promo.voucher_code)), [promoItems, tab])
  const PERIOD_OPTS = [{ value: '', label: 'Semua masa berlaku' }, { value: 'berjalan', label: 'Sedang berjalan' }, { value: 'belum', label: 'Belum mulai' }, { value: 'berakhir', label: 'Sudah berakhir' }]
  const filtered = useMemo(() => {
    const byStatus = statusF ? kindItems.filter((promo) => (statusF === 'aktif') === promo.active) : kindItems
    const byPeriod = periodF ? byStatus.filter((promo) => {
      const now = Date.now()
      const start = new Date(promo.starts_at).getTime()
      const end = new Date(promo.ends_at).getTime()
      return periodF === 'berjalan' ? start <= now && now <= end : periodF === 'belum' ? now < start : now > end
    }) : byStatus
    const term = q.trim().toLocaleLowerCase('id-ID')
    if (!term) return byPeriod
    return byPeriod.filter((promo) => [promo.name, promo.voucher_code ?? ''].some((value) => value.toLocaleLowerCase('id-ID').includes(term)))
  }, [kindItems, q, statusF, periodF])
  const totalPages = Math.max(1, Math.ceil(filtered.length / pageSize))
  const safePage = Math.min(page, totalPages - 1)
  const pageItems = filtered.slice(safePage * pageSize, safePage * pageSize + pageSize)

  function load() {
    promos.reload()
  }

  function openCreate() {
    setErr('')
    setFormErr('')
    setEditKind(tab)
    setEditing({ ...emptyDraft })
  }

  async function save(draft: Draft) {
    setErr('')
    setFormErr('')
    if (editKind === 'voucher' && !draft.code.trim()) {
      setFormErr('Kode voucher wajib diisi.')
      return
    }
    if (draft.happyStart && !draft.happyEnd || !draft.happyStart && draft.happyEnd) {
      setFormErr('Isi kedua jam promo harian atau kosongkan keduanya.')
      return
    }
    const body: PromoInput = {
      name: draft.name.trim(),
      type: draft.type,
      value: Number(draft.value),
      scope: draft.scope,
      categoryId: draft.scope === 'category' ? Number(draft.categoryId) : null,
      productId: draft.scope === 'product' ? Number(draft.productId) : null,
      voucherCode: editKind === 'voucher' ? draft.code.trim() : null,
      startsAt: toDateTime(draft.startsAt),
      endsAt: toDateTime(draft.endsAt),
      happyStart: draft.happyStart || null,
      happyEnd: draft.happyEnd || null,
      minPurchase: Number(draft.minPurchase) || 0,
      maxDiscount: draft.type === 'percent' ? Number(draft.maxDiscount) || 0 : 0,
      active: draft.active,
      combinable: editKind === 'voucher' ? draft.combinable : false,
    }
    setBusy(true)
    try {
      if (draft.id) await updatePromo.mutateAsync({ id: draft.id, body })
      else await createPromo.mutateAsync(body)
      setEditing(null)
      load()
    } catch (e) {
      setFormErr(e instanceof Error ? e.message : 'Gagal menyimpan promo.')
    } finally {
      setBusy(false)
    }
  }

  async function toggleActive(promo: Promo) {
    setErr('')
    setBusyId(promo.id)
    try {
      await setPromoActive.mutateAsync({ id: promo.id, active: !promo.active })
      load()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Gagal mengubah status promo.')
    } finally {
      setBusyId(null)
    }
  }

  async function removePromo() {
    if (!deleteFor) return
    setErr('')
    setDeleteErr('')
    setBusy(true)
    try {
      await deletePromo.mutateAsync(deleteFor.id)
      setDeleteFor(null)
      load()
    } catch (e) {
      setDeleteErr(e instanceof Error ? e.message : 'Gagal menghapus promo.')
    } finally {
      setBusy(false)
    }
  }

  function scopeLabel(promo: Promo) {
    if (promo.scope === 'category') return `Kategori · ${categoryItems.find((item) => String(item.id) === String(promo.category_id))?.name ?? 'Tidak tersedia'}`
    if (promo.scope === 'product') return `Produk · ${productItems.find((item) => String(item.id) === String(promo.product_id))?.name ?? 'Tidak tersedia'}`
    return 'Semua produk'
  }

  return (
    <>
      <PageHeader
        title="Promo & Voucher"
       
        crumb="Promo & Voucher"
        actions={<Button onClick={openCreate}><Plus className="size-4" />Tambah {tab === 'voucher' ? 'Voucher' : 'Promo'}</Button>}
      />

      <Tabs value={tab} onValueChange={(v) => { setTab(v as 'promo' | 'voucher'); setPage(0) }} className="mb-5">
        <TabsList variant="line" aria-label="Jenis promosi">
          <TabsTrigger value="promo">Promo</TabsTrigger>
          <TabsTrigger value="voucher">Voucher</TabsTrigger>
        </TabsList>
      </Tabs>

      {(err || promos.err || categories.err || products.err) && (
        <AlertStrip className="mb-4">
          {err || promos.err || categories.err || products.err}
        </AlertStrip>
      )}

      <div className="mb-4 flex flex-col gap-3 sm:flex-row">
        <SearchInput value={q} onChange={(e) => { setQ(e.target.value); setPage(0) }} placeholder="Ketik nama atau kode…" aria-label="Cari nama atau kode promo" />
        <Select
          label="Filter status"
          value={statusF}
          onChange={(v) => { setStatusF(v); setPage(0) }}
          className="shrink-0 sm:w-56"
          options={[{ value: '', label: 'Semua status' }, { value: 'aktif', label: 'Aktif' }, { value: 'nonaktif', label: 'Nonaktif' }]}
        />
        <FilterPanel
          className="shrink-0 sm:w-44"
          activeCount={periodF ? 1 : 0}
          onReset={() => { setPeriodF(''); setPage(0) }}
        >
          <FilterField label="Masa berlaku">
            <Select label="Filter masa berlaku" value={periodF} onChange={(v) => { setPeriodF(v); setPage(0) }} options={PERIOD_OPTS} />
          </FilterField>
        </FilterPanel>
      </div>
      <FilterChips items={periodF ? [{ key: 'periode', label: `Masa berlaku: ${PERIOD_OPTS.find((o) => o.value === periodF)?.label ?? periodF}`, onClear: () => { setPeriodF(''); setPage(0) } }] : []} />
      <TableWrap className="hidden min-[1280px]:block">
        <table className="w-full border-collapse">
          <thead>
            <tr><Th>Nama</Th>{tab === 'voucher' && <Th>Kode</Th>}<Th>Tipe</Th><Th>Cakupan</Th><Th>Periode</Th><Th>Status</Th><Th>Aksi</Th></tr>
          </thead>
          {!promos.data ? (
            <SkeletonRows cols={tab === 'voucher' ? 7 : 6} />
          ) : pageItems.length > 0 ? (
            <tbody>
              {pageItems.map((promo) => (
                <tr key={promo.id}>
                  <Td><span className="font-medium text-fg">{promo.name}</span></Td>
                  {tab === 'voucher' && (
                    <Td mono>
                      <span className="flex flex-col items-start gap-1">
                        <span>{promo.voucher_code}</span>
                        {promo.combinable && <Pill tone="muted">Bisa digabung</Pill>}
                      </span>
                    </Td>
                  )}
                  <Td>{promo.type === 'percent' ? `${promo.value}%` : fmtRp(promo.value)}</Td>
                  <Td>{scopeLabel(promo)}</Td>
                  <Td>
                    <span className="block whitespace-nowrap">{formatDateTime(promo.starts_at)}</span>
                    <span className="block whitespace-nowrap text-[11px] text-fog">s.d. {formatDateTime(promo.ends_at)}</span>
                  </Td>
                  <Td><Pill tone={promo.active ? 'ok' : 'muted'}>{promo.active ? 'Aktif' : 'Nonaktif'}</Pill></Td>
                  <Td>
                    <div className="flex items-center gap-1">
                      <RowAction title="Ubah" label={`Ubah ${promo.name}`} onClick={() => { setErr(''); setFormErr(''); setEditKind(promo.voucher_code ? 'voucher' : 'promo'); setEditing(draftFromPromo(promo)) }}><Pencil className="size-4" /></RowAction>
                      <RowAction title={promo.active ? 'Nonaktifkan' : 'Aktifkan'} label={`${promo.active ? 'Nonaktifkan' : 'Aktifkan'} ${promo.name}`} onClick={() => toggleActive(promo)} disabled={busyId === promo.id}><Power className="size-4" /></RowAction>
                      <RowAction title="Hapus" label={`Hapus ${promo.name}`} onClick={() => { setDeleteErr(''); setDeleteFor(promo) }} danger><Trash2 className="size-4" /></RowAction>
                    </div>
                  </Td>
                </tr>
              ))}
            </tbody>
          ) : (
            <tbody>
              <tr><td colSpan={tab === 'voucher' ? 7 : 6} className="p-2"><Empty title={kindItems.length ? `Tidak ada ${tab === 'voucher' ? 'voucher' : 'promo'} yang cocok` : `Belum ada ${tab === 'voucher' ? 'voucher' : 'promo'}`} sub={kindItems.length ? 'Coba kata kunci lain.' : `Tambahkan ${tab === 'voucher' ? 'voucher' : 'promo'} pertama untuk mulai menggunakannya di POS.`} action={kindItems.length ? <Button onClick={() => setQ('')}>Tampilkan semua</Button> : undefined} /></td></tr>
            </tbody>
          )}
        </table>
      </TableWrap>
      <CardGrid>
        {!promos.data ? (
          [0, 1, 2].map((i) => <Skeleton key={i} className="h-[140px] w-full rounded-xl" />)
        ) : pageItems.map((promo) => (
          <DataCard key={promo.id}>
            <div className="flex items-center justify-between gap-2">
              <p className="min-w-0 truncate font-medium text-fg">{promo.name}</p>
              <Pill tone={promo.active ? 'ok' : 'muted'}>{promo.active ? 'Aktif' : 'Nonaktif'}</Pill>
            </div>
            {tab === 'voucher' && (
              <p className="mt-1 flex flex-wrap items-center gap-1.5">
                <span className="rounded bg-surface px-1.5 py-0.5 font-mono text-xs text-fg">{promo.voucher_code}</span>
                {promo.combinable && <Pill tone="muted">Bisa digabung</Pill>}
              </p>
            )}
            <p className="mt-1.5 text-right text-[17px] font-semibold tabular-nums text-fg">{promo.type === 'percent' ? `${promo.value}%` : fmtRp(promo.value)}</p>
            <p className="mt-1 text-xs leading-snug text-fog">
              {scopeLabel(promo)}{promo.min_purchase > 0 ? ` · Min. belanja ${fmtRp(promo.min_purchase)}` : ''}
            </p>
            <p className="mt-0.5 text-xs leading-snug text-fog">{formatDateTime(promo.starts_at)} s.d. {formatDateTime(promo.ends_at)}</p>
            <div className="mt-auto flex items-center gap-1 pt-1.5">
              <RowAction title="Ubah" label={`Ubah ${promo.name}`} onClick={() => { setErr(''); setFormErr(''); setEditKind(promo.voucher_code ? 'voucher' : 'promo'); setEditing(draftFromPromo(promo)) }}><Pencil className="size-4" /></RowAction>
              <RowAction title={promo.active ? 'Nonaktifkan' : 'Aktifkan'} label={`${promo.active ? 'Nonaktifkan' : 'Aktifkan'} ${promo.name}`} onClick={() => toggleActive(promo)} disabled={busyId === promo.id}><Power className="size-4" /></RowAction>
              <RowAction title="Hapus" label={`Hapus ${promo.name}`} onClick={() => { setDeleteErr(''); setDeleteFor(promo) }} danger><Trash2 className="size-4" /></RowAction>
            </div>
          </DataCard>
        ))}
      </CardGrid>
      {promos.data && pageItems.length === 0 && (
        <div className="mt-4 min-[1280px]:hidden">
          <Empty title={kindItems.length ? `Tidak ada ${tab === 'voucher' ? 'voucher' : 'promo'} yang cocok` : `Belum ada ${tab === 'voucher' ? 'voucher' : 'promo'}`} sub={kindItems.length ? 'Coba kata kunci lain.' : `Tambahkan ${tab === 'voucher' ? 'voucher' : 'promo'} pertama untuk mulai menggunakannya di POS.`} action={kindItems.length ? <Button onClick={() => setQ('')}>Tampilkan semua</Button> : undefined} />
        </div>
      )}
      {promos.data && <Pager page={safePage} total={totalPages} onChange={setPage} totalItems={filtered.length} pageSize={pageSize} onPageSizeChange={(n) => { setPageSize(n); setPage(0) }} />}

      <Modal open={!!editing} title={editing?.id ? `Ubah ${editKind === 'voucher' ? 'Voucher' : 'Promo'}` : `Tambah ${editKind === 'voucher' ? 'Voucher' : 'Promo'}`} onClose={() => setEditing(null)} wide>
        {editing && <FormPromo kind={editKind} draft={editing} categories={categoryItems} products={productItems} busy={busy} error={formErr} onSave={save} onCancel={() => setEditing(null)} />}
      </Modal>

      <Modal open={!!deleteFor} title={`Hapus ${deleteFor?.voucher_code ? 'Voucher' : 'Promo'} · ${deleteFor?.name ?? ''}`} onClose={() => setDeleteFor(null)}>
        <div className="space-y-4">
          <p className="text-sm text-muted">Data ini akan dihapus dan tidak dapat dipulihkan.</p>
          {deleteErr && <AlertStrip>{deleteErr}</AlertStrip>}
          <ModalActions>
            <Button variant="ghost" onClick={() => setDeleteFor(null)}>Batal</Button>
            <Button variant="danger" disabled={busy} onClick={removePromo}>{busy ? 'Menghapus…' : 'Hapus'}</Button>
          </ModalActions>
        </div>
      </Modal>
    </>
  )
}

function FormPromo({ kind, draft, categories, products, busy, error, onSave, onCancel }: {
  kind: 'promo' | 'voucher'
  draft: Draft
  categories: Category[]
  products: Product[]
  busy: boolean
  error: string
  onSave: (draft: Draft) => void
  onCancel: () => void
}) {
  const [value, setValue] = useState(draft)
  const set = (key: keyof Draft) => (next: string | boolean) => setValue((current) => ({ ...current, [key]: next }))

  return (
    <form className="grid gap-4 sm:grid-cols-2" onSubmit={(event) => { event.preventDefault(); onSave(value) }}>
      {error && <AlertStrip className="sm:col-span-2">{error}</AlertStrip>}
      <div className="sm:col-span-2"><Input label={kind === 'voucher' ? 'Nama voucher' : 'Nama promo'} value={value.name} onChange={set('name')} required /></div>
      <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
        Tipe promo
        <Select label="Tipe promo" value={value.type} onChange={set('type')} options={[{ value: 'percent', label: 'Persentase (%)' }, { value: 'amount', label: 'Potongan nominal (Rp)' }]} />
      </label>
      <NumInput label={value.type === 'percent' ? 'Nilai persentase' : 'Nilai potongan (Rp)'} value={value.value} onValue={set('value')} allowDecimal={value.type === 'percent'} required />
      <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
        Cakupan promo
        <Select label="Cakupan promo" value={value.scope} onChange={set('scope')} options={[{ value: 'all', label: 'Semua produk' }, { value: 'category', label: 'Kategori' }, { value: 'product', label: 'Produk' }]} />
      </label>
      {value.scope === 'category' && (
        <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
          Kategori
          <Select label="Kategori" value={value.categoryId} onChange={set('categoryId')} options={[{ value: '', label: 'Pilih kategori' }, ...categories.filter((category) => category.active || String(category.id) === value.categoryId).map((category) => ({ value: String(category.id), label: category.name }))]} />
        </label>
      )}
      {value.scope === 'product' && (
        <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
          Produk
          <Select label="Produk" value={value.productId} onChange={set('productId')} options={[{ value: '', label: 'Pilih produk' }, ...products.filter((product) => product.active || String(product.id) === value.productId).map((product) => ({ value: String(product.id), label: product.name }))]} />
        </label>
      )}
      {kind === 'voucher' && <Input label="Kode voucher" value={value.code} onChange={set('code')} required />}
      <DateTimeField label="Mulai berlaku" value={value.startsAt} onChange={set('startsAt')} required />
      <DateTimeField label="Berakhir" value={value.endsAt} onChange={set('endsAt')} required />
      <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
        Jam mulai harian (opsional)
        <TimePicker label="Jam mulai harian (opsional)" value={value.happyStart} onChange={set('happyStart')} />
      </label>
      <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
        Jam berakhir harian (opsional)
        <TimePicker label="Jam berakhir harian (opsional)" value={value.happyEnd} onChange={set('happyEnd')} />
      </label>
      <NumInput label="Minimal pembelian (Rp)" value={value.minPurchase} onValue={set('minPurchase')} />
      {value.type === 'percent' && <NumInput label="Maksimal potongan (Rp)" value={value.maxDiscount} onValue={set('maxDiscount')} />}
      <CheckRow checked={value.active} onChange={set('active')}>
        {kind === 'voucher' ? 'Voucher aktif' : 'Promo aktif'}
      </CheckRow>
      {kind === 'voucher' && (
        <CheckRow checked={value.combinable} onChange={set('combinable')} align="start">
          <span>Dapat digabungkan dengan promo otomatis<span className="block font-normal text-fog">Voucher dihitung dari sisa setelah promo otomatis diterapkan.</span></span>
        </CheckRow>
      )}
      <ModalActions className="sm:col-span-2">
        <Button type="button" variant="ghost" onClick={onCancel}>Batal</Button>
        <Button type="submit" disabled={busy}>{busy ? 'Menyimpan…' : 'Simpan'}</Button>
      </ModalActions>
    </form>
  )
}
