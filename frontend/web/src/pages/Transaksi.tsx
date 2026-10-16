// Transaksi — transaction management workspace: ringkasan → cari/filter → review → detail/refund.
// Token font/warna milik sistem (tidak ada token baru di file ini).
import { useMemo, useState } from 'react'
import { Banknote, CalendarDays, ReceiptText, Sigma, Eye, Undo2 } from 'lucide-react'
import { apiExportTransactions, apiGetDashboard, apiListTransactions, apiListUsers, apiRefundTransaction, fetchAll, type Trx, type User } from '../lib/api'
import { selectCsvExportHandler } from '../lib/export-role'
import { PageHeader } from '../lib/PageHeader'
import { useCache } from '../lib/cache'
import { downloadBlob, exportCSV, exportExcel, fmtDate, fmtRp, fmtTime, useDB } from '../lib/store'
import { MetaList, ModalActions, AlertStrip, Button, CardGrid, DataCard, Empty, ExportMenu, FilterChips, FilterField, FilterPanel, Modal, NumInput, Pager, PeriodPicker, RowAction, SearchInput, Select, SkeletonRows, StatCard, StatusPill, TableWrap, Td, Th, TrxItems } from '../lib/ui'
import { Skeleton } from '@/components/ui/skeleton'

const METHODS = ['Semua', 'Cash', 'Bank Transfer', 'QRIS', 'E-Wallet', 'Card']

const fmtInv = (id: string) => (isNaN(Number(id)) ? `#${id}` : `#TRX-${String(Number(id)).padStart(5, '0')}`)

function isoLocal(d: Date): string {
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function presetToRange(id: string): { from: string; to: string } | null {
  const now = new Date()
  const today = isoLocal(now)
  if (id === 'today') return { from: today, to: today }
  if (id === 'yesterday') {
    const y = new Date(now)
    y.setDate(y.getDate() - 1)
    const ys = isoLocal(y)
    return { from: ys, to: ys }
  }
  if (id === 'week') {
    const m = new Date(now)
    m.setDate(m.getDate() - ((m.getDay() + 6) % 7))
    return { from: isoLocal(m), to: today }
  }
  if (id === 'month') return { from: `${today.slice(0, 8)}01`, to: today }
  if (id === 'year') return { from: `${today.slice(0, 4)}-01-01`, to: `${today.slice(0, 4)}-12-31` }
  return null
}

const TRX_STATUS_OPTIONS = [
  { value: '', label: 'Semua status' },
  { value: 'completed', label: 'Selesai' },
  { value: 'refunded', label: 'Refund' },
  { value: 'completed_offline_shortage', label: 'Selesai · Stok Kurang' },
  { value: 'pending', label: 'Proses' },
  { value: 'cancelled', label: 'Dibatalkan' },
]

const TRX_PERIODS = [
  { id: 'all', label: 'Semua' },
  { id: 'today', label: 'Hari ini' },
  { id: 'yesterday', label: 'Kemarin' },
  { id: 'week', label: 'Minggu ini' },
  { id: 'month', label: 'Bulan ini' },
  { id: 'year', label: 'Tahun ini' },
]

export default function Transaksi() {
  const db = useDB()
  const s = db.session!
  const [q, setQ] = useState('')
  const [method, setMethod] = useState('Semua')
  const [statusF, setStatusF] = useState('')
  const [cashierF, setCashierF] = useState('')
  const usersQ = useCache(`trxusers:${s.id}:${s.role}`, () => apiListUsers(), 'Gagal memuat kasir.')
  const kasirOptions = useMemo(() => {
    const opts: { value: string; label: string }[] = [{ value: '', label: 'Semua kasir' }]
    for (const u of (usersQ.data ?? []) as User[]) {
      if (!u.active) continue
      // cashier_id transaksi: owner = 0; kasir = id record kasir (id di daftar pengguna).
      if (u.role === 'admin') opts.push({ value: '0', label: u.name })
      else if (u.role === 'cashier') opts.push({ value: String(u.id), label: u.name })
    }
    return opts
  }, [usersQ.data])
  const [range, setRange] = useState<{ from: string; to: string } | null>(null)
  const [preset, setPreset] = useState('all')
  const effRange = range ?? presetToRange(preset)
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const f = {
    q: q.trim() || undefined,
    method: method === 'Semua' ? undefined : method,
    status: statusF || undefined,
    cashierId: cashierF || undefined,
    from: effRange?.from,
    to: effRange?.to,
  }
  const list = useCache(
    `trx:${s.id}:${s.role}:${q.trim()}:${method}:${statusF}:${cashierF}:${preset}:${range?.from ?? ''}:${range?.to ?? ''}:${page}:${pageSize}`,
    () => apiListTransactions({ ...f, page: page + 1, limit: pageSize }),
    'Gagal memuat transaksi.',
  )
  const trx = list.data?.items ?? []
  const total = list.data?.total ?? 0
  const loading = list.loading
  // Ringkasan mengikuti filter aktif (fetch-all, pola sama seperti Export).
  const agg = useCache(
    `trxagg:${s.id}:${s.role}:${q.trim()}:${method}:${statusF}:${cashierF}:${preset}:${range?.from ?? ''}:${range?.to ?? ''}`,
    () => fetchAll<Trx>((pg) => apiListTransactions({ ...f, page: pg, limit: 200 })),
    'Gagal memuat ringkasan.',
  )
  const dash = useCache(`dashtrx:${s.id}:${s.role}`, apiGetDashboard)
  const [err, setErr] = useState('')
  const [detail, setDetail] = useState<Trx | null>(null)
  const [refundFor, setRefundFor] = useState<Trx | null>(null)
  const [refundItems, setRefundItems] = useState<{ productId: string; qty: number }[]>([])
  const [refundReason, setRefundReason] = useState('')
  const [busy, setBusy] = useState(false)

  const sum = useMemo(() => {
    if (!agg.data) return null
    const omzet = agg.data.reduce((n, t) => n + t.total, 0)
    const n = agg.data.length
    return { omzet, n, avg: n > 0 ? Math.round(omzet / n) : 0 }
  }, [agg.data])

  function load() {
    list.reload()
    agg.reload()
  }

  const pages = Math.max(1, Math.ceil(total / pageSize))
  const safePage = Math.min(page, pages - 1)

  async function fetchFiltered(): Promise<Trx[]> {
    const all: Trx[] = []
    for (let p = 1; ; p++) {
      const r = await apiListTransactions({ q: q.trim() || undefined, method: method === 'Semua' ? undefined : method, status: statusF || undefined, cashierId: cashierF || undefined, from: effRange?.from, to: effRange?.to, page: p, limit: 200 })
      all.push(...r.items)
      if (all.length >= r.total) break
    }
    return all
  }

  function trxRows(all: Trx[]): string[][] {
    return [
      ['id', 'waktu', 'kasir', 'metode', 'subtotal', 'diskon', 'pajak', 'total', 'dibayar', 'kembalian', 'status'],
      ...all.map((t) => [t.id, t.created_at, t.cashier_name, t.method, String(t.subtotal), String(t.discount), String(t.tax), String(t.total), String(t.paid), String(t.change), t.status]),
    ]
  }

  async function exportListCSV() {
    exportCSV('transaksi.csv', trxRows(await fetchFiltered()))
  }

  async function exportServerCSV() {
    setErr('')
    try {
      const file = await apiExportTransactions({ period: 'all', from: effRange?.from, to: effRange?.to })
      const today = new Date().toISOString().slice(0, 10)
      await downloadBlob(file.filename ?? `transaksi-${today}.csv`, file.blob)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Gagal mengekspor transaksi.')
    }
  }

  const TRX_STATUS: Record<string, string> = {
    completed: 'Selesai', pending: 'Proses', cancelled: 'Dibatalkan', refunded: 'Refund',
  }

  async function exportListExcel() {
    const all = await fetchFiltered()
    const filterInfo = [
      method === 'Semua' ? null : `Metode: ${method}`,
      effRange ? `Tanggal: ${effRange.from === effRange.to ? fmtDate(effRange.from) : `${fmtDate(effRange.from)} – ${fmtDate(effRange.to)}`}` : null,
      q.trim() ? `Cari: ${q.trim()}` : null,
    ].filter(Boolean).join(' · ')
    await exportExcel('transaksi.xlsx', [{
      name: 'Transaksi',
      title: 'Daftar Transaksi',
      subtitle: [`${all.length} transaksi`, filterInfo].filter(Boolean).join(' · '),
      columns: [
        { header: 'ID' },
        { header: 'Waktu' },
        { header: 'Kasir' },
        { header: 'Metode' },
        { header: 'Subtotal (Rp)', money: true },
        { header: 'Diskon (Rp)', money: true },
        { header: 'Pajak (Rp)', money: true },
        { header: 'Total (Rp)', money: true },
        { header: 'Dibayar (Rp)', money: true },
        { header: 'Kembalian (Rp)', money: true },
        { header: 'Status', align: 'center' },
      ],
      rows: all.map((t) => [
        t.id, t.created_at, t.cashier_name, t.method, t.subtotal, t.discount,
        t.tax, t.total, t.paid, t.change, TRX_STATUS[t.status] ?? t.status,
      ]),
      sumCols: [7],
    }])
  }

  function openRefund(t: Trx) {
    setRefundFor(t)
    setRefundItems(t.items.map((i) => ({ productId: i.product_id, qty: i.qty })))
    setRefundReason('')
  }

  async function submitRefund() {
    if (!refundFor || !refundReason.trim()) return
    const items = refundItems.filter((i) => i.qty > 0)
    if (items.length === 0) return
    setBusy(true); setErr('')
    try {
      await apiRefundTransaction(refundFor.id, items, refundReason.trim())
      setRefundFor(null)
      load()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Gagal memproses refund.')
    } finally {
      setBusy(false)
    }
  }

  const isAdmin = s.role === 'admin'
  const cards = [
    { label: 'Total Transaksi', value: sum ? String(sum.n) : null, icon: ReceiptText },
    { label: 'Total Penjualan', value: sum ? fmtRp(sum.omzet) : null, icon: Banknote },
    { label: 'Transaksi Hari Ini', value: dash.data ? String(dash.data.today.trx_count) : null, icon: CalendarDays },
    { label: 'Rata-rata Transaksi', value: sum ? fmtRp(sum.avg) : null, icon: Sigma },
  ]

  return (
    <>
      <PageHeader
        title="Transaksi"
       
        crumb="Transaksi"
        actions={(
          <div className="opc-head-actions-stack">
            <div className="w-full sm:w-72">
              <PeriodPicker options={TRX_PERIODS} period={preset} range={range} onPreset={(id) => { setPreset(id); setRange(null); setPage(0) }} onRange={(from, to) => { setRange({ from, to }); setPage(0) }} label="Filter tanggal" />
            </div>
            <ExportMenu onCSV={selectCsvExportHandler(s.role, exportServerCSV, exportListCSV)} onExcel={exportListExcel} />
          </div>
        )}
      />

      {(err || list.err) && <AlertStrip className="mb-4">{err || list.err}</AlertStrip>}

      <div className="grid grid-cols-2 gap-3 sm:gap-4 xl:grid-cols-4">
        {cards.map((c) => (
          <StatCard key={c.label} density="stat" label={c.label} icon={c.icon} value={c.value === null ? <Skeleton className="h-6 w-24" /> : c.value} />
        ))}
      </div>

      <div className="mt-5 flex flex-col gap-3 sm:flex-row">
        <SearchInput value={q} onChange={(e) => { setQ(e.target.value); setPage(0) }} placeholder="Cari invoice, pelanggan, atau kasir…" aria-label="Cari transaksi" />
        <Select
          label="Filter metode pembayaran"
          value={method}
          onChange={(v) => { setMethod(v); setPage(0) }}
          className="shrink-0 sm:w-56"
          options={[...METHODS, 'Credit'].map((m) => ({ value: m, label: m === 'Semua' ? 'Semua metode' : m }))}
        />
        <FilterPanel
          className="shrink-0 sm:w-44"
          activeCount={(statusF ? 1 : 0) + (cashierF !== '' ? 1 : 0)}
          onReset={() => { setStatusF(''); setCashierF(''); setPage(0) }}
        >
          <FilterField label="Status">
            <Select label="Filter status transaksi" value={statusF} onChange={(v) => { setStatusF(v); setPage(0) }} options={TRX_STATUS_OPTIONS} />
          </FilterField>
          {s.role === 'admin' && (
            <FilterField label="Kasir">
              <Select label="Filter kasir" value={cashierF} onChange={(v) => { setCashierF(v); setPage(0) }} options={kasirOptions} />
            </FilterField>
          )}
        </FilterPanel>
      </div>
      <FilterChips items={[
        ...(statusF ? [{ key: 'status', label: `Status: ${TRX_STATUS_OPTIONS.find((o) => o.value === statusF)?.label ?? statusF}`, onClear: () => { setStatusF(''); setPage(0) } }] : []),
        ...(cashierF !== '' ? [{ key: 'kasir', label: `Kasir: ${kasirOptions.find((o) => o.value === cashierF)?.label ?? cashierF}`, onClear: () => { setCashierF(''); setPage(0) } }] : []),
      ]} />

      {/* Tabel 8-9 kolom butuh ~847px; di bawah 1280 (iPad/tablet) tabel dipindah
          ke kartu ringkas di bawah supaya total, status, dan tombol Detail/Refund
          tak terpotong. ≥1280 tetap tabel apa adanya. */}
      <TableWrap className="mt-4 hidden min-[1280px]:block">
        {loading || trx.length > 0 ? (
          <table className="w-full border-collapse">
            <thead>
              <tr>
                <Th>Invoice</Th><Th>Tanggal</Th><Th>Pelanggan</Th>{isAdmin && <Th>Kasir</Th>}<Th>Produk</Th><Th right>Total</Th><Th>Pembayaran</Th><Th>Status</Th><Th>Aksi</Th>
              </tr>
            </thead>
            {loading ? (
              <SkeletonRows cols={isAdmin ? 9 : 8} rows={10} />
            ) : (
            <tbody>
              {trx.map((t) => (
                <tr key={t.id}>
                  <Td mono><span className="font-medium text-fg">{fmtInv(t.id)}</span></Td>
                  <Td>
                    <span className="block text-[13px] text-fg">{fmtDate(t.created_at)}</span>
                    <span className="block font-mono text-xs text-fog">{fmtTime(t.created_at)}</span>
                  </Td>
                  <Td>{t.customer || '—'}</Td>
                  {isAdmin && <Td>{t.cashier_name}</Td>}
                  <Td>
                    <TrxItems items={t.items} className="max-w-full sm:max-w-64" />
                  </Td>
                  <Td right><span className="text-[15px] font-semibold text-fg">{fmtRp(t.total)}</span></Td>
                  <Td>{t.method}</Td>
                  <Td><StatusPill status={t.status} /></Td>
                  <Td>
                    <div className="flex items-center gap-1">
                      <RowAction title="Detail" label={`Detail transaksi ${t.id}`} onClick={() => setDetail(t)}><Eye className="size-4" /></RowAction>
                      {s.role === 'admin' && t.status === 'completed' && (
                        <RowAction title="Refund" label={`Refund transaksi ${t.id}`} onClick={() => openRefund(t)}><Undo2 className="size-4" /></RowAction>
                      )}
                    </div>
                  </Td>
                </tr>
              ))}
            </tbody>
            )}
          </table>
        ) : (
          <Empty
            title="Tidak ada transaksi ditemukan"
            sub="Coba kata kunci, tanggal, atau metode lain."
            action={<Button onClick={() => { setQ(''); setRange(null); setPreset('all'); setMethod('Semua'); setPage(0) }}>Tampilkan semua</Button>}
          />
        )}
      </TableWrap>

      {/* Kartu 1 kolom full-width menyisakan ruang kosong di kanan tiap kartu
          dan hanya ~1-4 transaksi muat per layar. Grid ini memangkas lebar
          kartu ke ~243-326px (chips item + total + tombol tetap terbaca) lalu
          2 kolom mulai 768px, 3 kolom mulai 1120px (konten 840px -> kartu
          ~273px, di dalam pita 264-300px). 3 kolom belum boleh aktif di
          1024px: konten 744px hanya memberi ~241px, di bawah pita. <768 tetap
          1 kolom; >=1280 kembali ke tabel apa adanya. */}
      <CardGrid className="mt-4">
        {loading ? (
          [0, 1, 2, 3, 4, 5].map((i) => <Skeleton key={i} className="h-[170px] w-full rounded-xl" />)
        ) : trx.length > 0 ? (
          trx.map((t) => (
            <DataCard key={t.id}>
              <div className="flex items-center justify-between gap-2">
                <p className="min-w-0 truncate font-mono text-[13px] font-medium text-fg">{fmtInv(t.id)}</p>
                <StatusPill status={t.status} />
              </div>
              <p className="mt-1.5 text-right text-[17px] font-semibold tabular-nums text-fg">{fmtRp(t.total)}</p>
              {/* Di kolom sempit meta mendapat lebar penuh kartu dan boleh
                  membungkus (bukan truncate) supaya tanggal/pelanggan/kasir/
                  metode tidak pernah terpotong. */}
              <p className="mt-1 text-xs leading-snug text-fog">
                {fmtDate(t.created_at)} {fmtTime(t.created_at)}{t.customer ? ` · ${t.customer}` : ''}{isAdmin ? ` · ${t.cashier_name}` : ''} · {t.method}
              </p>
              <div className="mt-2">
                <TrxItems items={t.items} />
              </div>
              {/* min-h-9 (36px) di HP, min-h-10 (40px) mulai 768px supaya
                  target sentuh aman di tablet; mt-auto menyamakan baris
                  footer antar kartu pada baris grid yang sama. */}
              <div className="mt-auto flex items-center gap-1 pt-1.5">
                <RowAction title="Detail" label={`Detail transaksi ${t.id}`} onClick={() => setDetail(t)}><Eye className="size-4" /></RowAction>
                {s.role === 'admin' && t.status === 'completed' && (
                  <RowAction title="Refund" label={`Refund transaksi ${t.id}`} onClick={() => openRefund(t)}><Undo2 className="size-4" /></RowAction>
                )}
              </div>
            </DataCard>
          ))
        ) : (
          <p className="col-span-full py-10 text-center text-sm text-fog">Tidak ada transaksi ditemukan.</p>
        )}
      </CardGrid>

      <Pager page={safePage} total={pages} onChange={setPage} className="mt-4" totalItems={total} pageSize={pageSize} onPageSizeChange={(n) => { setPageSize(n); setPage(0) }} />

      <Modal open={!!detail} title={`Detail ${detail ? fmtInv(detail.id) : ''}`} onClose={() => setDetail(null)} wide>
        {detail && <TrxDetail t={detail} showCashier={isAdmin} />}
      </Modal>

      <Modal open={!!refundFor} title={`Refund ${refundFor ? fmtInv(refundFor.id) : ''}`} onClose={() => setRefundFor(null)} wide>
        {refundFor && (
          <div className="space-y-4">
            <p className="text-sm text-muted">Pilih jumlah item yang direfund. Stok akan dikembalikan otomatis.</p>
            <div className="max-h-60 divide-y divide-dove overflow-y-auto rounded-lg border border-dove px-3">
              {refundFor.items.map((i) => (
                <div key={i.product_id} className="flex items-center justify-between gap-3 py-2 text-[13px]">
                  <p className="min-w-0"><span className="text-fg">{i.name}</span><br /><span className="text-xs text-fog">Terjual {i.qty}</span></p>
                  <NumInput
                    aria-label={`Qty refund ${i.name}`}
                    value={refundItems.find((r) => r.productId === i.product_id)?.qty ?? 0}
                    onValue={(r) => setRefundItems((rs) => rs.map((x) => x.productId === i.product_id ? { ...x, qty: Math.min(i.qty, Math.max(0, Number(r || 0))) } : x))}
                    className="w-20 shrink-0 rounded border border-dove bg-paper px-2 py-1.5 text-right font-mono tabular-nums focus:border-jet focus:outline-none"
                  />
                </div>
              ))}
            </div>
            <input
              value={refundReason} onChange={(e) => setRefundReason(e.target.value)} placeholder="Alasan refund (wajib)"
              className="w-full rounded-md border border-border bg-paper px-3.5 py-2.5 text-sm focus:border-jet focus:outline-none"
            />
            <ModalActions>
              <Button variant="ghost" onClick={() => setRefundFor(null)}>Batal</Button>
              <Button onClick={submitRefund} disabled={!refundReason.trim() || refundItems.every((i) => i.qty === 0) || busy}>{busy ? 'Memproses…' : 'Proses Refund'}</Button>
            </ModalActions>
          </div>
        )}
      </Modal>
    </>
  )
}

function TrxDetail({ t, showCashier }: { t: Trx; showCashier: boolean }) {
  return (
    <div className="space-y-3 text-sm">
      <div className="rounded-lg bg-surface p-4">
        <MetaList items={[
          { label: 'Invoice', value: fmtInv(t.id), mono: true },
          { label: 'Waktu', value: `${fmtDate(t.created_at)} ${fmtTime(t.created_at)}` },
          ...(showCashier ? [{ label: 'Kasir', value: t.cashier_name }] : []),
          ...(t.customer ? [{ label: 'Pelanggan', value: t.customer }] : []),
          { label: 'Metode', value: t.method },
          { label: 'Status', value: <StatusPill status={t.status} /> },
        ]} />
      </div>
      <div className="divide-y divide-dove rounded-lg border border-dove px-3">
        {t.items.map((i) => (
          <div key={i.product_id} className="flex items-center justify-between gap-3 py-2 text-[13px]">
            <p className="min-w-0"><span className="text-fg">{i.name}</span><br /><span className="text-xs text-fog">{fmtRp(i.price)} × {i.qty}</span></p>
            <p className="shrink-0 font-medium tabular-nums text-fg">{fmtRp(i.price * i.qty)}</p>
          </div>
        ))}
      </div>
      <div className="space-y-1 font-mono text-[13px]">
        <div className="flex justify-between"><span className="text-fog">Subtotal</span><span>{fmtRp(t.subtotal)}</span></div>
        {t.discount > 0 && <div className="flex justify-between"><span className="text-fog">Diskon</span><span>-{fmtRp(t.discount)}</span></div>}
        {(t.promo_discount ?? 0) > 0 && <div className="flex justify-between"><span className="text-fog">Promo · {t.promo_name}</span><span>-{fmtRp(t.promo_discount ?? 0)}</span></div>}
        {(t.voucher_discount ?? 0) > 0 && <div className="flex justify-between"><span className="text-fog">Voucher · {t.voucher_code}</span><span>-{fmtRp(t.voucher_discount ?? 0)}</span></div>}
        {t.tax > 0 && <div className="flex justify-between"><span className="text-fog">Pajak</span><span>{fmtRp(t.tax)}</span></div>}
        <div className="flex justify-between border-t border-dove pt-1.5 font-medium"><span>Total</span><span>{fmtRp(t.total)}</span></div>
        <div className="flex justify-between"><span className="text-fog">Dibayar</span><span>{fmtRp(t.paid)}</span></div>
        <div className="flex justify-between"><span className="text-fog">Kembalian</span><span>{fmtRp(t.change)}</span></div>
      </div>
    </div>
  )
}
