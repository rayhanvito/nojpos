import { useEffect, useState } from 'react'
import { Clock3, Banknote, Eye, ShieldAlert, OctagonX } from 'lucide-react'
import { OutletFilter } from '../../components/OutletFilter'
import { useOutlets } from '../../hooks/use-outlets'
import { apiExportShifts, apiGetShift, type Shift, type ShiftDetail } from './api'
import { useForceCloseShift, useShifts } from './hooks'
import { PageHeader } from '../../shared/PageHeader'
import { outletQuery } from '../../shared/outlet'
import { forceCloseShiftBody, isForceClosedShift, shiftDurationLabel, shiftMethodRows, shiftStatusLabel, shiftTransactionsForPage, shiftVarianceParts } from '../../shared/shift'
import { downloadBlob, fmtDate, fmtRp, fmtTime } from '../../shared/session/store'
import { ModalActions, AlertStrip, Button, CardGrid, DataCard, Empty, ExportMenu, InfoStrip, Modal, NumInput, Pager, Pill, SkeletonRows, StatCard, StatusPill, Td, Th, TableWrap, RowAction } from '../../shared/ui'
import { Skeleton } from '@/shared/ui/primitives/skeleton'


function shiftTime(value: string | null) {
  return value ? `${fmtDate(value)} ${fmtTime(value)}` : '—'
}

function varianceLabel(value: number) {
  const variance = shiftVarianceParts(value)
  return `${variance.sign}${fmtRp(variance.amount)}`
}

export default function ShiftPage({ embedded = false, outletId: controlledOutletId, from, to }: { embedded?: boolean; outletId?: string; from?: string; to?: string }) {
  const { outlets } = useOutlets()
  const [ownOutletId, setOwnOutletId] = useState('')
  const outletId = controlledOutletId ?? ownOutletId
  const setOutletId = setOwnOutletId
  const selectedOutlet = outletQuery(outletId).outletId
  const shifts = useShifts(selectedOutlet, from, to)
  const forceCloseShift = useForceCloseShift()
  const [selectedShift, setSelectedShift] = useState<Shift | null>(null)
  const [detailPage, setDetailPage] = useState(0)
  const [detailLimit, setDetailLimit] = useState(10)
  const [detail, setDetail] = useState<ShiftDetail | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailError, setDetailError] = useState('')
  const [forceTarget, setForceTarget] = useState<Shift | null>(null)
  const [forceCash, setForceCash] = useState('')
  const [forceBusy, setForceBusy] = useState(false)
  const [forceError, setForceError] = useState('')
  const [exportBusy, setExportBusy] = useState(false)
  const [exportError, setExportError] = useState('')
  const [notice, setNotice] = useState('')

  useEffect(() => {
    if (!selectedShift) {
      setDetail(null)
      setDetailLoading(false)
      setDetailError('')
      return
    }
    let dead = false
    setDetailLoading(true)
    setDetailError('')
    apiGetShift(selectedShift.id, { page: detailPage + 1, limit: detailLimit })
      .then((value) => { if (!dead) setDetail(value) })
      .catch((error) => { if (!dead) setDetailError(error instanceof Error ? error.message : 'Gagal memuat detail shift.') })
      .finally(() => { if (!dead) setDetailLoading(false) })
    return () => { dead = true }
  }, [selectedShift, detailPage, detailLimit])

  function openDetail(shift: Shift) {
    setDetailPage(0)
    setDetail(null)
    setSelectedShift(shift)
  }

  async function exportCsv() {
    setExportBusy(true)
    setExportError('')
    try {
      const file = await apiExportShifts(outletQuery(selectedOutlet))
      await downloadBlob(file.filename ?? 'shift.csv', file.blob)
    } catch (error) {
      setExportError(error instanceof Error ? error.message : 'Gagal mengekspor shift.')
    } finally {
      setExportBusy(false)
    }
  }

  function openForceClose(shift: Shift) {
    setForceTarget(shift)
    setForceCash('')
    setForceError('')
  }

  async function forceClose() {
    if (!forceTarget) return
    const body = forceCloseShiftBody(forceCash)
    if (body === null) {
      setForceError('Uang aktual harus berupa Rupiah bulat minimal 0 atau dikosongkan.')
      return
    }
    setForceBusy(true)
    setForceError('')
    try {
      await forceCloseShift.mutateAsync({ id: forceTarget.id, body })
      setNotice(`Shift ${forceTarget.cashier_name} berhasil ditutup paksa.`)
      setForceTarget(null)
      setSelectedShift(null)
      shifts.reload()
    } catch (error) {
      setForceError(error instanceof Error ? error.message : 'Gagal menutup shift.')
    } finally {
      setForceBusy(false)
    }
  }

  const rows = shifts.data ?? []
  const openRows = rows.filter((r) => r.closed_at == null)
  const closedRows = rows.filter((r) => r.closed_at != null)
  const totalVariance = closedRows.reduce((n, r) => n + (r.cash_variance ?? 0), 0)
  const forceClosedCount = rows.filter((r) => isForceClosedShift(r)).length
  const currentShift = detail?.shift
  const methodRows = shiftMethodRows(currentShift?.sales_by_method ?? {})
  const trxPage = detail ? shiftTransactionsForPage(detail.transactions, detailPage + 1) : undefined
  const error = shifts.err || exportError

  return (
    <>
      {!embedded && (
        <PageHeader
          title="Shift"
         
          crumb="Shift"
          actions={(
            <div className="flex flex-wrap items-center gap-2">
              <OutletFilter outlets={outlets} value={outletId} onChange={setOutletId} />
              <ExportMenu label="Ekspor CSV" onCSV={() => void exportCsv()} disabled={exportBusy} />
            </div>
          )}
        />
      )}

      {error && <AlertStrip className="mb-4">{error}</AlertStrip>}
      {notice && <AlertStrip tone="success" className="mb-4">{notice}</AlertStrip>}

      {rows.length > 0 && (
        <div className="mb-4 grid grid-cols-2 gap-3 sm:grid-cols-3">
          <StatCard icon={Clock3} label="Shift berjalan" value={String(openRows.length)} sub={openRows.length ? openRows.map((r) => r.cashier_name).join(', ') : 'Tidak ada shift terbuka'} />
          <StatCard icon={Banknote} label="Total selisih kas" value={closedRows.length ? varianceLabel(totalVariance) : '—'} sub={`dari ${closedRows.length} shift selesai pada tampilan ini`} />
          <StatCard icon={ShieldAlert} label="Ditutup paksa" value={String(forceClosedCount)} sub="ditutup admin, bukan kasirnya" />
        </div>
      )}

      {!shifts.loading && !shifts.err && rows.length === 0 ? (
        <Empty title="Belum ada data shift" sub="Shift yang sudah berjalan akan tampil di sini." />
      ) : (
        <>
        <TableWrap className="hidden min-[1280px]:block">
          <table className="w-full border-collapse">
            <thead>
              <tr>
                <Th>Kasir</Th><Th>Mulai</Th><Th>Selesai</Th><Th>Durasi</Th><Th right>Modal Awal</Th><Th right>Penjualan</Th><Th right>Transaksi</Th><Th right>Kas diharapkan</Th><Th right>Kas Aktual</Th><Th right>Selisih</Th><Th>Status</Th><Th right>Aksi</Th>
              </tr>
            </thead>
            {shifts.loading && rows.length === 0 ? <SkeletonRows cols={12} /> : (
              <tbody>
                {rows.map((shift) => {
                  const isOpen = shift.closed_at === null || shift.closed_at === undefined
                  return (
                    <tr
                      key={shift.id}
                      tabIndex={0}
                      aria-label={`Buka detail shift kasir ${shift.cashier_name}`}
                      onClick={() => openDetail(shift)}
                      onKeyDown={(event) => {
                        if (event.target !== event.currentTarget || (event.key !== 'Enter' && event.key !== ' ')) return
                        event.preventDefault()
                        openDetail(shift)
                      }}
                      className="cursor-pointer hover:bg-paper/50 focus-visible:outline-2 focus-visible:outline-accent-soft"
                    >
                      <Td>{shift.cashier_name}</Td>
                      <Td mono>{shiftTime(shift.started_at)}</Td>
                      <Td mono>{shiftTime(shift.closed_at)}</Td>
                      <Td mono>{shiftDurationLabel(shift.started_at, shift.closed_at, Date.now())}</Td>
                      <Td right>{fmtRp(shift.opening_cash)}</Td>
                      <Td right>{fmtRp(shift.sales)}</Td>
                      <Td right>{shift.trx_count}</Td>
                      <Td right>{shift.expected_cash === null ? '—' : fmtRp(shift.expected_cash)}</Td>
                      <Td right>{shift.closing_cash_actual === null || shift.closing_cash_actual === undefined ? '—' : fmtRp(shift.closing_cash_actual)}</Td>
                      <Td right>{isOpen || shift.cash_variance === null || shift.cash_variance === undefined ? '—' : shift.cash_variance === 0 ? <Pill tone="ok">Pas</Pill> : <Pill tone={shift.cash_variance < 0 ? 'warn' : 'muted'}>{varianceLabel(shift.cash_variance)}</Pill>}</Td>
                      <Td>{isOpen ? <Pill tone="warn">Berjalan</Pill> : isForceClosedShift(shift) ? <Pill tone="muted">Ditutup Paksa</Pill> : <Pill tone="ok">Selesai</Pill>}</Td>
                      <Td right>
                        {isOpen && (
                          <span onClick={(event) => event.stopPropagation()}>
                            <RowAction title="Tutup paksa" onClick={() => openForceClose(shift)} danger ><OctagonX className="size-4" /></RowAction>
                          </span>
                        )}
                      </Td>
                    </tr>
                  )
                })}
              </tbody>
            )}
          </table>
        </TableWrap>
        <CardGrid>
          {shifts.loading && rows.length === 0 ? (
            [0, 1, 2].map((i) => <Skeleton key={i} className="h-[150px] w-full rounded-xl" />)
          ) : rows.map((shift) => {
            const isOpen = shift.closed_at === null || shift.closed_at === undefined
            return (
              <DataCard key={shift.id}>
                <div className="flex items-center justify-between gap-2">
                  <p className="min-w-0 truncate font-medium text-fg">{shift.cashier_name}</p>
                  {isOpen ? <Pill tone="warn">Berjalan</Pill> : isForceClosedShift(shift) ? <Pill tone="muted">Ditutup Paksa</Pill> : <Pill tone="ok">Selesai</Pill>}
                </div>
                <p className="mt-1.5 text-right text-[17px] font-semibold tabular-nums text-fg">
                  {isOpen ? fmtRp(shift.sales) : shift.cash_variance === null || shift.cash_variance === undefined ? '—' : varianceLabel(shift.cash_variance)}
                </p>
                <p className="mt-0.5 text-right text-xs text-fog">{isOpen ? 'penjualan selama shift berjalan' : 'selisih kas'}</p>
                <div className="mt-2 grid grid-cols-2 gap-x-3 gap-y-1 text-xs leading-snug">
                  <p className="text-fog">Modal awal<br /><span className="font-medium tabular-nums text-fg">{fmtRp(shift.opening_cash)}</span></p>
                  <p className="text-fog">Penjualan · {shift.trx_count} trx<br /><span className="font-medium tabular-nums text-fg">{fmtRp(shift.sales)}</span></p>
                  <p className="text-fog">Kas diharapkan<br /><span className="font-medium tabular-nums text-fg">{shift.expected_cash === null ? '—' : fmtRp(shift.expected_cash)}</span></p>
                  <p className="text-fog">Kas aktual<br /><span className="font-medium tabular-nums text-fg">{shift.closing_cash_actual === null || shift.closing_cash_actual === undefined ? '—' : fmtRp(shift.closing_cash_actual)}</span></p>
                </div>
                <p className="mt-1.5 text-xs leading-snug text-fog">{shiftTime(shift.started_at)} – {shiftTime(shift.closed_at)} · {shiftDurationLabel(shift.started_at, shift.closed_at, Date.now())}</p>
                <div className="mt-auto flex items-center gap-1 pt-1.5">
                  <RowAction title="Detail" label={`Buka detail shift kasir ${shift.cashier_name}`} onClick={() => openDetail(shift)}><Eye className="size-4" /></RowAction>
                  {isOpen && <RowAction title="Tutup paksa" label={`Tutup paksa shift ${shift.cashier_name}`} onClick={() => openForceClose(shift)} danger><OctagonX className="size-4" /></RowAction>}
                </div>
              </DataCard>
            )
          })}
        </CardGrid>
        </>
      )}

      <Modal
        open={selectedShift !== null}
        title={`Detail Shift${currentShift?.cashier_name || selectedShift?.cashier_name ? ` · ${currentShift?.cashier_name ?? selectedShift?.cashier_name}` : ''}`}
        onClose={() => setSelectedShift(null)}
        wide
      >
        {detailLoading && <p className="py-6 text-center text-sm text-muted" role="status">Memuat detail shift…</p>}
        {detailError && <AlertStrip className="mb-3">{detailError}</AlertStrip>}
        {currentShift && !detailLoading && (
          <div className="space-y-5">
            <div className="grid gap-3 sm:grid-cols-2">
              <p><span className="block text-xs text-muted">Kasir</span><strong className="text-sm">{currentShift.cashier_name}</strong></p>
              <p><span className="block text-xs text-muted">Status</span><strong className="text-sm">{shiftStatusLabel(currentShift)}</strong></p>
              <p><span className="block text-xs text-muted">Mulai</span><strong className="text-sm">{shiftTime(currentShift.started_at)}</strong></p>
              <p><span className="block text-xs text-muted">Selesai</span><strong className="text-sm">{shiftTime(currentShift.closed_at)}</strong></p>
              <p><span className="block text-xs text-muted">Modal awal</span><strong className="text-sm tabular-nums">{fmtRp(currentShift.opening_cash)}</strong></p>
              <p><span className="block text-xs text-muted">Total penjualan</span><strong className="text-sm tabular-nums">{fmtRp(currentShift.sales)}</strong></p>
              <p><span className="block text-xs text-muted">Jumlah transaksi</span><strong className="text-sm tabular-nums">{currentShift.trx_count}</strong></p>
              <p><span className="block text-xs text-muted">Total refund</span><strong className="text-sm tabular-nums">{fmtRp(currentShift.refund_total)}</strong></p>
              <p><span className="block text-xs text-muted">Kas diharapkan</span><strong className="text-sm tabular-nums">{currentShift.expected_cash === null ? '—' : fmtRp(currentShift.expected_cash)}</strong></p>
              {(currentShift.cash_expenses ?? 0) > 0 && <p><span className="block text-xs text-muted">Pengeluaran laci</span><strong className="text-sm tabular-nums">−{fmtRp(currentShift.cash_expenses ?? 0)}</strong></p>}
              <p><span className="block text-xs text-muted">Uang aktual</span><strong className="text-sm tabular-nums">{currentShift.closing_cash_actual === null ? '—' : fmtRp(currentShift.closing_cash_actual)}</strong></p>
              <p><span className="block text-xs text-muted">Selisih kas</span><strong className="text-sm tabular-nums">{currentShift.cash_variance === null ? '—' : varianceLabel(currentShift.cash_variance)}</strong></p>
              {currentShift.close_note && <p className="sm:col-span-2"><span className="block text-xs text-muted">Catatan penutupan</span><strong className="text-sm">{currentShift.close_note}</strong></p>}
            </div>

            <section>
              <h4 className="mb-2 text-sm font-medium">Rincian per metode pembayaran</h4>
              <ul className="divide-y divide-dove/60 rounded-xl bg-paper px-3.5">
                {methodRows.map((row) => (
                  <li key={row.method} className="flex items-center justify-between py-2 text-sm">
                    <span>{row.method}</span>
                    <span className="tabular-nums">{fmtRp(row.total)}</span>
                  </li>
                ))}
              </ul>
            </section>

            {detail?.hourly && detail.hourly.length > 0 && (
              <section>
                <h4 className="mb-2 text-sm font-medium">Penjualan per jam</h4>
                <div className="flex items-end gap-1.5 rounded-xl bg-paper p-3" role="img" aria-label="Grafik penjualan per jam">
                  {detail.hourly.map((h) => {
                    const max = Math.max(...detail.hourly!.map((x) => x.omzet), 1)
                    return (
                      <div key={h.hour} className="flex flex-1 flex-col items-center gap-1">
                        <div className="w-full rounded-t" style={{ height: `${Math.max(6, Math.round((h.omzet / max) * 72))}px`, backgroundColor: 'var(--chart-1)' }} title={`${String(h.hour).padStart(2, '0')}.00 · ${fmtRp(h.omzet)}`} />
                        <span className="text-[10px] tabular-nums text-muted">{String(h.hour).padStart(2, '0')}</span>
                      </div>
                    )
                  })}
                </div>
              </section>
            )}

            {detail?.top_products && detail.top_products.length > 0 && (
              <section>
                <h4 className="mb-2 text-sm font-medium">Produk terlaris shift ini</h4>
                <ul className="divide-y divide-dove/60 rounded-xl bg-paper px-3.5">
                  {detail.top_products.map((prod) => (
                    <li key={prod.product_id} className="flex items-center justify-between py-2 text-sm">
                      <span>{prod.name}</span>
                      <span className="tabular-nums text-muted">×{prod.qty}</span>
                    </li>
                  ))}
                </ul>
              </section>
            )}

            <section>
              <h4 className="mb-2 text-sm font-medium">Transaksi shift</h4>
              {!trxPage || trxPage.items.length === 0 ? (
                <p className="rounded-xl border border-dashed border-dove px-4 py-6 text-center text-sm text-muted">Belum ada transaksi pada shift ini.</p>
              ) : (
                <div className="divide-y divide-dove rounded-lg border border-dove px-3">
                  {trxPage.items.map((trx) => (
                    <div key={trx.id} className="flex items-center justify-between gap-3 py-2.5">
                      <p className="min-w-0">
                        <span className="font-mono text-[13px] font-medium text-fg">#{trx.seq ?? trx.id}</span>{' '}
                        <StatusPill status={trx.status} /><br />
                        <span className="text-xs text-fog">{shiftTime(trx.created_at)} · {trx.method}</span>
                      </p>
                      <p className="shrink-0 font-medium tabular-nums text-fg">{fmtRp(trx.total)}</p>
                    </div>
                  ))}
                </div>
              )}
              <Pager page={detailPage} total={Math.ceil((trxPage?.total ?? 0) / detailLimit)} onChange={setDetailPage} totalItems={trxPage?.total ?? 0} pageSize={detailLimit} onPageSizeChange={(n) => { setDetailLimit(n); setDetailPage(0) }} />
            </section>

            {currentShift.closed_at == null && (
              <div className="flex justify-end">
                <Button variant="danger" onClick={() => openForceClose(currentShift)}>Tutup paksa shift</Button>
              </div>
            )}
          </div>
        )}
      </Modal>

      <Modal
        open={forceTarget !== null}
        title="Konfirmasi tutup paksa"
        onClose={() => { if (!forceBusy) setForceTarget(null) }}
      >
        {forceTarget && (
          <div className="space-y-4">
            <InfoStrip className="flex items-start gap-3 p-3.5">
              <Clock3 className="mt-0.5 size-4 shrink-0 text-muted" aria-hidden="true" />
              <p className="text-sm text-muted">Shift <strong className="text-fg">{forceTarget.cashier_name}</strong> masih terbuka. Tutup shift ini sekarang?</p>
            </InfoStrip>
            <NumInput
              label="Uang aktual (opsional)"
              value={forceCash}
              onValue={setForceCash}
              disabled={forceBusy}
              hint="Kosongkan jika uang aktual tidak dicatat."
            />
            {forceError && <AlertStrip>{forceError}</AlertStrip>}
            <ModalActions>
              <Button variant="ghost" disabled={forceBusy} onClick={() => setForceTarget(null)}>Batal</Button>
              <Button variant="danger" disabled={forceBusy} onClick={() => void forceClose()}>{forceBusy ? 'Menutup…' : 'Tutup paksa'}</Button>
            </ModalActions>
          </div>
        )}
      </Modal>
    </>
  )
}
