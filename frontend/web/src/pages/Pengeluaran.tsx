import { useEffect, useMemo, useRef, useState } from 'react'
import { Pencil, ReceiptText, Tag, Trash2, Wallet } from 'lucide-react'
import {
  apiCreateExpense, apiDeleteExpense, apiListExpenses, apiUpdateExpense, fetchAll,
  type Expense,
} from '../lib/api'
import { useCache } from '../lib/cache'
import { fmtDate, fmtRp, useDB } from '../lib/store'
import { DatePicker, ModalActions, AlertStrip, Button, CardGrid, DataCard, Empty, FilterChips, FilterField, FilterPanel, Input, Modal, NumInput, Pager, RowAction, SearchInput, Select, SkeletonRows, StatCard, TableWrap, Td, Th } from '../lib/ui'
import { Skeleton } from '@/components/ui/skeleton'

const PRESET_CATEGORIES = ['Listrik & Air', 'Internet', 'Sewa', 'Kemasan & Perlengkapan', 'Transportasi', 'Gaji & Komisi', 'Lainnya']
const METHODS = ['Cash', 'QRIS', 'Bank Transfer']

type Draft = {
  id?: number
  date: string
  categorySel: string
  categoryCustom: string
  note: string
  amount: string
  method: string
  source: string
}

function todayLocal(): string {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

// Tab Pengeluaran di halaman Pembelian: catatan biaya operasional toko
// (token listrik, wifi, sewa) — terpisah dari PO stok. Pola toolbar, tabel,
// dan modal mengikuti standar halaman daftar lain.
export default function Pengeluaran({ createSignal = 0 }: { createSignal?: number }) {
  const { session } = useDB()
  const isAdmin = session?.role === 'admin'
  const who = `${session?.id}:${session?.role}`
  const list = useCache<Expense[]>(`expenses:${who}`, () => fetchAll<Expense>((page) => apiListExpenses({ page, limit: 200 })), 'Gagal memuat pengeluaran.')
  const items = list.data ?? []

  const [q, setQ] = useState('')
  const [catF, setCatF] = useState('')
  const [methodF, setMethodF] = useState('')
  const [byF, setByF] = useState('')
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const [draft, setDraft] = useState<Draft | null>(null)
  const [formError, setFormError] = useState('')
  const [busy, setBusy] = useState(false)
  const [deleting, setDeleting] = useState<Expense | null>(null)
  const [deleteError, setDeleteError] = useState('')

  const categoryOptions = useMemo(() => {
    const names = new Set<string>(PRESET_CATEGORIES.filter((c) => c !== 'Lainnya'))
    for (const e of items) names.add(e.category)
    return [{ value: '', label: 'Semua kategori' }, ...[...names].sort().map((n) => ({ value: n, label: n }))]
  }, [items])

  const creatorOptions = useMemo(() => {
    const byId = new Map<number, string>()
    for (const e of items) byId.set(e.created_by_id, e.created_by_name)
    return [{ value: '', label: 'Semua pencatat' }, ...[...byId.entries()].map(([id, name]) => ({ value: String(id), label: name }))]
  }, [items])

  const filtered = useMemo(() => items.filter((e) => {
    if (catF && e.category !== catF) return false
    if (methodF && e.method !== methodF) return false
    if (byF && String(e.created_by_id) !== byF) return false
    const term = q.trim().toLocaleLowerCase('id-ID')
    if (term && !`${e.category} ${e.note} ${e.created_by_name}`.toLocaleLowerCase('id-ID').includes(term)) return false
    return true
  }), [items, q, catF, methodF, byF])

  const pages = Math.max(1, Math.ceil(filtered.length / pageSize))
  const safePage = Math.min(page, pages - 1)
  const pageItems = filtered.slice(safePage * pageSize, safePage * pageSize + pageSize)

  const monthKey = todayLocal().slice(0, 7)
  const monthItems = useMemo(() => items.filter((e) => e.date.slice(0, 7) === monthKey), [items, monthKey])
  const monthTotal = monthItems.reduce((n, e) => n + e.amount, 0)
  const topCategory = useMemo(() => {
    const byCat = new Map<string, number>()
    for (const e of monthItems) byCat.set(e.category, (byCat.get(e.category) ?? 0) + e.amount)
    let best: { name: string; total: number } | null = null
    for (const [name, total] of byCat) if (!best || total > best.total) best = { name, total }
    return best
  }, [monthItems])

  function openCreate() {
    setFormError('')
    setDraft({ date: todayLocal(), categorySel: 'Listrik & Air', categoryCustom: '', note: '', amount: '', method: 'Cash', source: 'outside' })
  }

  const signalRef = useRef(createSignal)
  useEffect(() => {
    if (createSignal !== signalRef.current) {
      signalRef.current = createSignal
      openCreate()
    }
  }, [createSignal])

  function openEdit(e: Expense) {
    setFormError('')
    const isPreset = PRESET_CATEGORIES.includes(e.category) && e.category !== 'Lainnya'
    setDraft({
      id: e.id,
      date: e.date.slice(0, 10),
      categorySel: isPreset ? e.category : 'Lainnya',
      categoryCustom: isPreset ? '' : e.category,
      note: e.note ?? '',
      amount: String(e.amount),
      method: e.method,
      source: e.source,
    })
  }

  async function save() {
    if (!draft) return
    const category = draft.categorySel === 'Lainnya' ? draft.categoryCustom.trim() : draft.categorySel
    const amount = Number(draft.amount)
    if (!draft.date) return setFormError('Tanggal wajib diisi.')
    if (!category) return setFormError('Kategori wajib diisi.')
    if (!Number.isFinite(amount) || amount <= 0) return setFormError('Nominal harus lebih dari 0.')
    setFormError('')
    setBusy(true)
    try {
      const body = {
        date: draft.date, category, note: draft.note.trim(), amount,
        method: draft.method, source: draft.method === 'Cash' ? draft.source : 'outside',
      }
      if (draft.id) await apiUpdateExpense(draft.id, body)
      else await apiCreateExpense(body)
      setDraft(null)
      list.reload()
    } catch (e) {
      setFormError(e instanceof Error ? e.message : 'Gagal menyimpan pengeluaran.')
    } finally {
      setBusy(false)
    }
  }

  async function removeExpense() {
    if (!deleting) return
    setDeleteError('')
    setBusy(true)
    try {
      await apiDeleteExpense(deleting.id)
      setDeleting(null)
      list.reload()
    } catch (e) {
      setDeleteError(e instanceof Error ? e.message : 'Gagal menghapus pengeluaran.')
    } finally {
      setBusy(false)
    }
  }

  return (
    <>
      {list.err && <AlertStrip className="mb-4">{list.err}</AlertStrip>}

      <div className="grid grid-cols-2 gap-3 sm:gap-4 xl:grid-cols-3">
        <StatCard density="stat" label="Pengeluaran Bulan Ini" value={fmtRp(monthTotal)} icon={Wallet} />
        <StatCard density="stat" label="Catatan Bulan Ini" value={String(monthItems.length)} icon={ReceiptText} />
        <StatCard density="stat" label="Kategori Terbesar" value={topCategory ? topCategory.name : '—'} sub={topCategory ? fmtRp(topCategory.total) : 'belum ada pengeluaran'} icon={Tag} />
      </div>

      {items.length === 0 && !list.loading ? (
        <div className="mt-4">
          <Empty title="Belum ada pengeluaran" sub="Catat pengeluaran operasional pertama — token listrik, wifi, sewa, dan lainnya." />
        </div>
      ) : (
        <>
          <div className="mt-5 flex flex-col gap-3 sm:flex-row">
            <SearchInput value={q} onChange={(e) => { setQ(e.target.value); setPage(0) }} placeholder="Cari kategori, keterangan, atau pencatat…" aria-label="Cari pengeluaran" />
            <Select
              label="Filter kategori pengeluaran"
              value={catF}
              onChange={(v) => { setCatF(v); setPage(0) }}
              className="shrink-0 sm:w-56"
              options={categoryOptions}
            />
            <FilterPanel
              className="shrink-0 sm:w-44"
              activeCount={(methodF ? 1 : 0) + (byF ? 1 : 0)}
              onReset={() => { setMethodF(''); setByF(''); setPage(0) }}
            >
              <FilterField label="Metode bayar">
                <Select label="Filter metode bayar" value={methodF} onChange={(v) => { setMethodF(v); setPage(0) }} options={[{ value: '', label: 'Semua metode' }, ...METHODS.map((m) => ({ value: m, label: m }))]} />
              </FilterField>
              <FilterField label="Pencatat">
                <Select label="Filter pencatat" value={byF} onChange={(v) => { setByF(v); setPage(0) }} options={creatorOptions} />
              </FilterField>
            </FilterPanel>
          </div>
          <FilterChips items={[
            ...(methodF ? [{ key: 'metode', label: `Metode: ${methodF}`, onClear: () => { setMethodF(''); setPage(0) } }] : []),
            ...(byF ? [{ key: 'pencatat', label: `Pencatat: ${creatorOptions.find((o) => o.value === byF)?.label ?? byF}`, onClear: () => { setByF(''); setPage(0) } }] : []),
          ]} />

          {filtered.length === 0 && !list.loading ? (
            <Empty title="Tidak ada pengeluaran yang cocok" sub="Coba kata kunci atau filter lain." />
          ) : (
            <>
            <TableWrap className="hidden min-[1280px]:block">
              <table className="w-full border-collapse">
                <thead>
                  <tr><Th>Tanggal</Th><Th>Kategori</Th><Th>Keterangan</Th><Th>Metode</Th><Th>Sumber</Th><Th>Dicatat Oleh</Th><Th right>Nominal</Th>{isAdmin && <Th>Aksi</Th>}</tr>
                </thead>
                {list.loading && items.length === 0 ? (
                  <SkeletonRows cols={isAdmin ? 8 : 7} />
                ) : (
                  <tbody>
                    {pageItems.map((e) => (
                      <tr key={e.id}>
                        <Td><span className="whitespace-nowrap">{fmtDate(e.date)}</span></Td>
                        <Td><span className="font-medium text-fg">{e.category}</span></Td>
                        <Td><span className="block max-w-xs truncate" title={e.note}>{e.note || '—'}</span></Td>
                        <Td>{e.method}</Td>
                        <Td>{e.source === 'drawer' ? `Laci kasir${e.shift_id ? ` · Shift #${e.shift_id}` : ''}` : 'Di luar laci'}</Td>
                        <Td>{e.created_by_name}</Td>
                        <Td right><span className="font-medium text-fg">{fmtRp(e.amount)}</span></Td>
                        {isAdmin && (
                          <Td>
                            <div className="flex gap-1">
                              <RowAction title="Ubah" label={`Ubah pengeluaran ${e.category}`} onClick={() => openEdit(e)}><Pencil className="size-4" /></RowAction>
                              <RowAction title="Hapus" label={`Hapus pengeluaran ${e.category}`} onClick={() => { setDeleteError(''); setDeleting(e) }} danger><Trash2 className="size-4" /></RowAction>
                            </div>
                          </Td>
                        )}
                      </tr>
                    ))}
                  </tbody>
                )}
              </table>
            </TableWrap>
            <CardGrid>
              {list.loading && items.length === 0 ? (
                [0, 1, 2].map((i) => <Skeleton key={i} className="h-[120px] w-full rounded-xl" />)
              ) : pageItems.map((e) => (
                <DataCard key={e.id}>
                  <div className="flex items-center justify-between gap-2">
                    <p className="min-w-0 truncate font-medium text-fg">{e.category}</p>
                    <p className="shrink-0 text-[15px] font-semibold tabular-nums text-fg">{fmtRp(e.amount)}</p>
                  </div>
                  {e.note && <p className="mt-1 text-xs leading-snug text-fog">{e.note}</p>}
                  <p className="mt-1 text-xs leading-snug text-fog">
                    {fmtDate(e.date)} · {e.method} · {e.source === 'drawer' ? `Laci kasir${e.shift_id ? ` · Shift #${e.shift_id}` : ''}` : 'Di luar laci'} · {e.created_by_name}
                  </p>
                  {isAdmin && (
                    <div className="mt-auto flex items-center gap-1 pt-1.5">
                      <RowAction title="Ubah" label={`Ubah pengeluaran ${e.category}`} onClick={() => openEdit(e)}><Pencil className="size-4" /></RowAction>
                      <RowAction title="Hapus" label={`Hapus pengeluaran ${e.category}`} onClick={() => { setDeleteError(''); setDeleting(e) }} danger><Trash2 className="size-4" /></RowAction>
                    </div>
                  )}
                </DataCard>
              ))}
            </CardGrid>
            </>
          )}
          <Pager page={safePage} total={pages} onChange={setPage} totalItems={filtered.length} pageSize={pageSize} onPageSizeChange={(n) => { setPageSize(n); setPage(0) }} />
        </>
      )}

      <Modal open={!!draft} title={draft?.id ? 'Ubah Pengeluaran' : 'Tambah Pengeluaran'} onClose={() => { if (!busy) setDraft(null) }}>
        {draft && (
          <div className="space-y-4">
            <div className="flex flex-col gap-1.5">
              <span className="text-[13px] font-medium text-steel">Tanggal</span>
              <DatePicker value={draft.date} onChange={(v) => setDraft({ ...draft, date: v })} label="Tanggal" />
            </div>
            <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
              Kategori
              <Select
                label="Kategori pengeluaran"
                value={draft.categorySel}
                onChange={(v) => setDraft({ ...draft, categorySel: v })}
                options={PRESET_CATEGORIES.map((c) => ({ value: c, label: c }))}
              />
            </label>
            {draft.categorySel === 'Lainnya' && (
              <Input label="Nama kategori" value={draft.categoryCustom} onChange={(v) => setDraft({ ...draft, categoryCustom: v })} placeholder="Mis. Parkir, Fotokopi" required />
            )}
            <NumInput label="Nominal (Rp)" value={draft.amount} onValue={(v) => setDraft({ ...draft, amount: v })} placeholder="0" required />
            <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
              Metode bayar
              <Select
                label="Metode bayar"
                value={draft.method}
                onChange={(v) => setDraft({ ...draft, method: v })}
                options={METHODS.map((m) => ({ value: m, label: m }))}
              />
            </label>
            {draft.method === 'Cash' && (
              <>
                <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
                  Sumber dana
                  <Select
                    label="Sumber dana"
                    value={draft.source}
                    onChange={(v) => setDraft({ ...draft, source: v })}
                    options={[{ value: 'outside', label: 'Di luar laci' }, { value: 'drawer', label: 'Laci kasir' }]}
                  />
                </label>
                {draft.source === 'drawer' && (
                  <p className="rounded-lg bg-surface px-3.5 py-2.5 text-[13px] text-muted">Pengeluaran dari laci mengurangi kas yang diharapkan pada shift yang sedang berjalan atas namamu.</p>
                )}
              </>
            )}
            <Input label="Keterangan (opsional)" value={draft.note} onChange={(v) => setDraft({ ...draft, note: v })} placeholder="Mis. Token listrik 100 kWh" />
            {formError && <AlertStrip>{formError}</AlertStrip>}
            <ModalActions>
              <Button variant="ghost" onClick={() => setDraft(null)} disabled={busy}>Batal</Button>
              <Button onClick={save} disabled={busy}>{busy ? 'Menyimpan…' : 'Simpan'}</Button>
            </ModalActions>
          </div>
        )}
      </Modal>

      <Modal open={!!deleting} title="Hapus Pengeluaran" onClose={() => { if (!busy) setDeleting(null) }}>
        {deleting && (
          <div className="space-y-4">
            <p className="text-sm text-muted">
              Hapus catatan <strong className="text-fg">{deleting.category}</strong> sebesar <strong className="text-fg">{fmtRp(deleting.amount)}</strong> ({fmtDate(deleting.date)})? Tindakan ini tidak bisa dibatalkan.
            </p>
            {deleteError && <AlertStrip>{deleteError}</AlertStrip>}
            <ModalActions>
              <Button variant="ghost" onClick={() => setDeleting(null)} disabled={busy}>Batal</Button>
              <Button variant="danger" onClick={removeExpense} disabled={busy}>{busy ? 'Menghapus…' : 'Hapus'}</Button>
            </ModalActions>
          </div>
        )}
      </Modal>
    </>
  )
}
