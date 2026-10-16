import { useRef, useState } from 'react'
import { Eye } from 'lucide-react'
import { apiCancelStockOpname, apiCreateStockOpname, apiGetStockOpname, apiListStockOpnames, apiPostStockOpname, apiUpdateStockOpnameItems, type StockOpnameCountsInput, type StockOpnameDetail, type StockOpnameItem, type StockOpnameSession, type StockOpnameStatus } from '../lib/api'
import { PageHeader } from '../lib/PageHeader'
import { useCache } from '../lib/cache'
import { countStockOpnameDifferences, stockOpnameChangedItems, stockOpnameCount, stockOpnameDifference } from '../lib/stock-opname'
import { fmtDate, useDB } from '../lib/store'
import { ModalActions, AlertStrip, Button, CardGrid, DataCard, Empty, InfoStrip, Modal, NumInput, Pager, Pill, RowAction, SearchInput, Select, SkeletonRows, TableWrap, Td, Th } from '../lib/ui'
import { Skeleton } from '@/components/ui/skeleton'

type ConfirmationAction = 'post' | 'cancel' | 'discard'

const STATUS_LABEL: Record<StockOpnameStatus, string> = {
  counting: 'Berlangsung',
  posted: 'Selesai',
  cancelled: 'Dibatalkan',
}

function statusPill(status: StockOpnameStatus) {
  const tone = status === 'posted' ? 'ok' : status === 'counting' ? 'warn' : 'muted'
  return <Pill tone={tone}>{STATUS_LABEL[status]}</Pill>
}

function itemName(item: StockOpnameItem) {
  return item.product_name ?? `Produk #${item.product_id}`
}

function variantName(item: StockOpnameItem) {
  if (item.variant_id == null) return '—'
  return item.variant_name ?? `Varian #${item.variant_id}`
}

function formatDifference(value: number | null) {
  if (value === null) return '—'
  return `${value > 0 ? '+' : ''}${value.toLocaleString('id-ID')}`
}


export default function Opname({ embedded = false }: { embedded?: boolean }) {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const sessions = useCache<StockOpnameSession[]>(`stock-opnames:${who}`, apiListStockOpnames, 'Gagal memuat sesi opname.')
  const [pageError, setPageError] = useState('')
  const [createBusy, setCreateBusy] = useState(false)
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const [statusF, setStatusF] = useState<StockOpnameStatus | ''>('')
  const [q, setQ] = useState('')
  const selectedIdRef = useRef<number | null>(null)
  const detailRequestRef = useRef(0)
  const [detail, setDetail] = useState<StockOpnameDetail | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailError, setDetailError] = useState('')
  const [counts, setCounts] = useState<Record<number, string>>({})
  const [saveBusy, setSaveBusy] = useState(false)
  const [actionBusy, setActionBusy] = useState(false)
  const [actionError, setActionError] = useState('')
  const [confirmation, setConfirmation] = useState<ConfirmationAction | null>(null)
  const [confirmError, setConfirmError] = useState('')

  const sessionItems = sessions.data ?? []
  const filteredSessions = sessionItems.filter((item) => {
    if (statusF && item.status !== statusF) return false
    const term = q.trim().toLocaleLowerCase('id-ID')
    if (!term) return true
    const haystack = `sesi #${item.id} #${item.id} ${item.id} ${STATUS_LABEL[item.status]} ${fmtDate(item.created_at)} ${item.created_by ?? ''} ${item.note ?? ''}`.toLocaleLowerCase('id-ID')
    return haystack.includes(term)
  })
  const pages = Math.max(1, Math.ceil(filteredSessions.length / pageSize))
  const safePage = Math.min(page, pages - 1)
  const pageItems = filteredSessions.slice(safePage * pageSize, safePage * pageSize + pageSize)
  const changedItems = detail ? stockOpnameChangedItems(detail.items, counts) : []
  const hasInvalidChange = changedItems.some((item) => stockOpnameCount(counts[item.id] ?? '') === null)
  const differentCount = detail
    ? detail.status === 'counting'
      ? countStockOpnameDifferences(detail.items, counts)
      : detail.items.filter((item) => item.difference !== null && item.difference !== 0).length
    : 0

  async function openDetail(id: number) {
    const requestToken = ++detailRequestRef.current
    selectedIdRef.current = id
    setSelectedId(id)
    setDetail(null)
    setCounts({})
    setDetailError('')
    setActionError('')
    setDetailLoading(true)
    try {
      const value = await apiGetStockOpname(id)
      if (selectedIdRef.current === id && requestToken === detailRequestRef.current) {
        setDetail(value)
        setCounts(Object.fromEntries(value.items.map((item) => [item.id, item.counted_stock === null ? '' : String(item.counted_stock)])))
      }
    } catch (e) {
      if (selectedIdRef.current === id && requestToken === detailRequestRef.current) {
        setDetailError(e instanceof Error ? e.message : 'Gagal memuat detail sesi opname.')
      }
    } finally {
      if (selectedIdRef.current === id && requestToken === detailRequestRef.current) setDetailLoading(false)
    }
  }

  function closeDetail() {
    if (saveBusy || actionBusy || confirmation) return
    if (changedItems.length > 0) {
      setConfirmError('')
      setConfirmation('discard')
      return
    }
    dismissDetail()
  }

  function dismissDetail() {
    detailRequestRef.current++
    selectedIdRef.current = null
    setSelectedId(null)
    setDetail(null)
    setCounts({})
    setDetailError('')
    setActionError('')
  }

  async function startSession() {
    setPageError('')
    setCreateBusy(true)
    try {
      const created = await apiCreateStockOpname()
      sessions.reload()
      await openDetail(created.id)
    } catch (e) {
      setPageError(e instanceof Error ? e.message : 'Gagal memulai sesi opname.')
    } finally {
      setCreateBusy(false)
    }
  }

  async function refreshDetail(id: number, successAction: string) {
    const requestToken = ++detailRequestRef.current
    sessions.reload()
    if (selectedIdRef.current === id) setDetailLoading(true)
    try {
      const value = await apiGetStockOpname(id)
      if (selectedIdRef.current === id && requestToken === detailRequestRef.current) {
        setDetail(value)
        setCounts(Object.fromEntries(value.items.map((item) => [item.id, item.counted_stock === null ? '' : String(item.counted_stock)])))
        setDetailError('')
      }
    } catch {
      if (selectedIdRef.current === id && requestToken === detailRequestRef.current) {
        setDetail(null)
        setDetailError(`${successAction}, tetapi detail gagal dimuat ulang. Tutup lalu buka kembali sesi ini.`)
      }
    } finally {
      if (selectedIdRef.current === id && requestToken === detailRequestRef.current) setDetailLoading(false)
    }
  }

  async function saveCounts() {
    if (!detail || changedItems.length === 0 || hasInvalidChange) return
    const items: StockOpnameCountsInput['items'] = []
    for (const item of changedItems) {
      const countedStock = stockOpnameCount(counts[item.id] ?? '')
      if (countedStock === null) return
      items.push({ itemId: item.id, countedStock })
    }

    setActionError('')
    setSaveBusy(true)
    try {
      await apiUpdateStockOpnameItems(detail.id, { items })
      await refreshDetail(detail.id, 'Perubahan stok hitung tersimpan')
    } catch (e) {
      setActionError(e instanceof Error ? e.message : 'Gagal menyimpan stok hitung.')
    } finally {
      setSaveBusy(false)
    }
  }

  async function confirmAction() {
    if (confirmation === 'discard') {
      setConfirmation(null)
      dismissDetail()
      return
    }
    if (!detail || !confirmation || detail.status !== 'counting' || (confirmation === 'post' && changedItems.length > 0)) return
    const action = confirmation
    setConfirmError('')
    setActionBusy(true)
    try {
      if (action === 'post') await apiPostStockOpname(detail.id)
      else await apiCancelStockOpname(detail.id)
      setConfirmation(null)
      await refreshDetail(detail.id, action === 'post' ? 'Sesi berhasil diposting' : 'Sesi berhasil dibatalkan')
    } catch (e) {
      setConfirmError(e instanceof Error ? e.message : 'Gagal memproses sesi opname.')
    } finally {
      setActionBusy(false)
    }
  }

  return (
    <>
{embedded ? (
        <div className="mb-4 flex justify-end">
          <Button onClick={() => void startSession()} disabled={createBusy}>{createBusy ? 'Memulai…' : 'Mulai Opname'}</Button>
        </div>
      ) : (
        <PageHeader
          title="Opname"
         
          crumb="Opname"
          actions={<Button onClick={() => void startSession()} disabled={createBusy}>{createBusy ? 'Memulai…' : 'Mulai Opname'}</Button>}
        />
      )}

      {(pageError || sessions.err) && <AlertStrip className="mb-4">{pageError || sessions.err}</AlertStrip>}

      {sessions.data === null && sessions.err ? null : sessionItems.length === 0 && !sessions.loading ? (
        <Empty title="Belum ada sesi opname" sub="Mulai sesi untuk mencatat stok fisik dan melihat selisihnya." />
      ) : (
        <>
          <div className="mb-4 flex flex-col gap-3 sm:flex-row">
            <SearchInput value={q} onChange={(e) => { setQ(e.target.value); setPage(0) }} placeholder="Cari nomor sesi atau pembuat…" aria-label="Cari sesi opname" />
            <Select
              label="Filter status sesi"
              value={statusF}
              onChange={(v) => { setStatusF(v as StockOpnameStatus | ''); setPage(0) }}
              className="shrink-0 sm:w-56"
              options={[{ value: '', label: 'Semua status' }, ...Object.entries(STATUS_LABEL).map(([value, label]) => ({ value, label }))]}
            />
          </div>
          {filteredSessions.length === 0 && !sessions.loading ? (
            <Empty title="Tidak ada sesi yang cocok" sub="Coba status atau kata kunci lain." />
          ) : (
          <>
          <TableWrap className="hidden min-[1280px]:block">
          <table className="w-full border-collapse">
            <thead>
              <tr><Th>Sesi</Th><Th>Status</Th><Th>Tanggal</Th><Th>Pembuat</Th><Th right>Aksi</Th></tr>
            </thead>
            {sessions.loading && sessionItems.length === 0 ? (
              <SkeletonRows cols={5} />
            ) : (
              <tbody>
                {pageItems.map((item) => (
                  <tr key={item.id}>
                    <Td mono>Sesi #{item.id}</Td>
                    <Td>{statusPill(item.status)}</Td>
                    <Td>{fmtDate(item.created_at)}</Td>
                    <Td>{item.created_by ?? '—'}</Td>
                    <Td><RowAction title="Detail" onClick={() => void openDetail(item.id)} ><Eye className="size-4" /></RowAction></Td>
                  </tr>
                ))}
              </tbody>
            )}
          </table>
        </TableWrap>
        <CardGrid>
          {sessions.loading && sessionItems.length === 0 ? (
            [0, 1, 2].map((i) => <Skeleton key={i} className="h-[110px] w-full rounded-xl" />)
          ) : pageItems.map((item) => (
            <DataCard key={item.id}>
              <div className="flex items-center justify-between gap-2">
                <p className="min-w-0 truncate font-mono text-[13px] font-medium text-fg">Sesi #{item.id}</p>
                {statusPill(item.status)}
              </div>
              <p className="mt-1 text-xs leading-snug text-fog">
                {fmtDate(item.created_at)} · {item.created_by ?? '—'}{item.note ? ` · ${item.note}` : ''}
              </p>
              <div className="mt-auto flex items-center gap-1 pt-1.5">
                <RowAction title="Detail" label={`Detail sesi ${item.id}`} onClick={() => void openDetail(item.id)}><Eye className="size-4" /></RowAction>
              </div>
            </DataCard>
          ))}
        </CardGrid>
          </>
          )}
        <Pager page={safePage} total={pages} onChange={setPage} totalItems={filteredSessions.length} pageSize={pageSize} onPageSizeChange={(n) => { setPageSize(n); setPage(0) }} />
        </>
      )}

      <Modal
        open={selectedId !== null}
        title={detail ? `Sesi Opname #${detail.id}` : 'Detail Sesi Opname'}
        onClose={closeDetail}
        wide
      >
        {detailLoading ? <p className="py-8 text-center text-sm text-muted" role="status">Memuat detail sesi…</p> : detailError ? (
          <AlertStrip>{detailError}</AlertStrip>
        ) : detail && (
          <div className="space-y-4">
            <div className="flex flex-wrap items-start justify-between gap-3 rounded-xl bg-surface px-4 py-3">
              <div>
                <p className="text-[13px] text-muted">Dibuat {fmtDate(detail.created_at)}{detail.created_by != null ? ` · ${detail.created_by}` : ''}</p>
                {detail.posted_at && <p className="mt-1 text-[13px] text-muted">Diposting {fmtDate(detail.posted_at)}</p>}
                {detail.note && <p className="mt-2 text-sm text-fg">{detail.note}</p>}
              </div>
              {statusPill(detail.status)}
            </div>

            {actionError && <AlertStrip>{actionError}</AlertStrip>}

            <InfoStrip className="flex flex-wrap items-center justify-between gap-2 px-4 py-3">
              <p className="text-sm text-muted">Item berselisih: <strong className="font-medium text-fg">{differentCount.toLocaleString('id-ID')}</strong></p>
              {detail.status === 'counting' && <p className="text-xs text-fog">Selisih adalah pratinjau; nilai resmi ditetapkan server.</p>}
            </InfoStrip>

            <div className="space-y-2.5">
              {detail.items.map((item) => {
                const difference = detail.status === 'counting'
                  ? stockOpnameDifference(item.system_stock, counts[item.id] ?? '')
                  : item.difference
                return (
                  <DataCard key={item.id}>
                    <div className="flex items-start justify-between gap-2">
                      <p className="min-w-0 font-medium text-fg">{itemName(item)}</p>
                      <p className="shrink-0 text-xs text-fog">
                        Selisih{' '}
                        <span className={`font-mono text-[13px] font-medium tabular-nums ${difference === null || difference === 0 ? 'text-muted' : difference > 0 ? 'text-sprout' : 'text-ember'}`}>
                          {formatDifference(difference)}
                        </span>
                      </p>
                    </div>
                    <p className="mt-0.5 text-xs text-fog">{variantName(item)} · Stok sistem <span className="font-mono tabular-nums text-fg">{item.system_stock.toLocaleString('id-ID')}</span></p>
                    {detail.status === 'counting' ? (
                      <div className="mt-2">
                        <NumInput
                          label="Stok hitung"
                          value={counts[item.id] ?? ''}
                          onValue={(value) => setCounts((current) => ({ ...current, [item.id]: value }))}
                          placeholder="0"
                          disabled={saveBusy || actionBusy}
                          className="w-full text-right"
                        />
                      </div>
                    ) : (
                      <p className="mt-1 text-xs text-fog">Stok hitung: <span className="font-mono tabular-nums text-fg">{item.counted_stock === null ? '—' : item.counted_stock.toLocaleString('id-ID')}</span></p>
                    )}
                  </DataCard>
                )
              })}
              {detail.items.length === 0 && <p className="py-8 text-center text-sm text-muted">Sesi ini tidak memiliki item stok.</p>}
            </div>

            {detail.status === 'counting' && hasInvalidChange && <p className="text-[13px] text-ember" role="alert">Isi stok hitung dengan bilangan bulat 0 atau lebih sebelum menyimpan.</p>}
            {detail.status === 'counting' && changedItems.length > 0 && <p className="text-[13px] text-muted">Simpan perubahan sebelum sesi dapat diposting.</p>}

            {detail.status === 'counting' && (
              <ModalActions>
                <Button onClick={() => void saveCounts()} disabled={saveBusy || actionBusy || changedItems.length === 0 || hasInvalidChange}>
                  {saveBusy ? 'Menyimpan…' : 'Simpan'}
                </Button>
                <Button onClick={() => { setConfirmError(''); setConfirmation('post') }} disabled={saveBusy || actionBusy || changedItems.length > 0}>Posting</Button>
                <Button variant="danger" onClick={() => { setConfirmError(''); setConfirmation('cancel') }} disabled={saveBusy || actionBusy}>Batalkan Sesi</Button>
              </ModalActions>
            )}
          </div>
        )}
      </Modal>

      <Modal
        open={confirmation !== null}
        title={confirmation === 'post' ? 'Posting Sesi Opname' : confirmation === 'cancel' ? 'Batalkan Sesi Opname' : 'Buang Perubahan'}
        onClose={() => { if (!actionBusy) setConfirmation(null) }}
      >
        {confirmation && detail && (
          <div className="space-y-4">
            <p className="text-sm text-muted">
              {confirmation === 'post'
                ? <>Posting Sesi Opname #{detail.id} akan menyesuaikan stok sesuai hasil hitung dan mengunci sesi. Lanjutkan?</>
                : confirmation === 'cancel'
                  ? <>Batalkan Sesi Opname #{detail.id}? Sesi yang dibatalkan tidak dapat dilanjutkan.</>
                  : <>Buang hitungan yang belum disimpan dan tutup detail sesi?</>}
            </p>
            {confirmError && <AlertStrip>{confirmError}</AlertStrip>}
            <ModalActions>
              <Button variant="ghost" onClick={() => setConfirmation(null)} disabled={actionBusy}>Kembali</Button>
              <Button variant={confirmation === 'post' ? 'primary' : 'danger'} onClick={() => void confirmAction()} disabled={actionBusy}>
                {actionBusy ? 'Memproses…' : confirmation === 'post' ? 'Posting' : confirmation === 'cancel' ? 'Batalkan Sesi' : 'Buang Perubahan'}
              </Button>
            </ModalActions>
          </div>
        )}
      </Modal>
    </>
  )
}
