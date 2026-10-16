import { useEffect, useRef, useState } from 'react'
import { Eye } from 'lucide-react'
import {
  apiAddPayablePayment, apiAddReceivablePayment, apiGetPayable, apiGetReceivable,
  apiListPayables, apiListReceivables,
  type DebtPaymentInput, type DebtStatus, type Payable, type PayableDetail,
  type Receivable, type ReceivableDetail, type StandardPayMethod,
} from '../lib/api'
import { PageHeader } from '../lib/PageHeader'
import { useCache } from '../lib/cache'
import { canSubmitDebtPayment, debtStatusLabel, filterDebtItems, isDebtPaymentAmountValid, type DebtFilter } from '../lib/debt'
import { fmtDate, fmtRp, useDB } from '../lib/store'
import { MetaList, ModalActions, AlertStrip, Button, CardGrid, DataCard, Empty, Modal, NumInput, Pager, Pill, RowAction, SearchInput, Select, SkeletonRows, TableWrap, Td, Th } from '../lib/ui'
import { Skeleton } from '@/components/ui/skeleton'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'

type LedgerTab = 'receivables' | 'payables'
type DetailSelection = { kind: LedgerTab; id: number }
type LoadedDetail = { kind: 'receivables'; value: ReceivableDetail } | { kind: 'payables'; value: PayableDetail }

const TABS: { id: LedgerTab; label: string }[] = [
  { id: 'receivables', label: 'Piutang' },
  { id: 'payables', label: 'Hutang' },
]

const STANDARD_METHODS: StandardPayMethod[] = ['Cash', 'Bank Transfer', 'QRIS', 'E-Wallet', 'Card']

const FILTERS: { id: DebtFilter; label: string }[] = [
  { id: 'all', label: 'Semua' },
  { id: 'unpaid', label: 'Belum Lunas' },
  { id: 'partially_paid', label: 'Sebagian' },
  { id: 'paid', label: 'Lunas' },
]

function statusPill(status: DebtStatus, dueDate?: string) {
  const label = debtStatusLabel(status, dueDate)
  const tone = label === 'Lunas' ? 'ok' : label === 'Belum Lunas' ? 'muted' : 'warn'
  return <Pill tone={tone}>{label}</Pill>
}


export default function HutangPiutang() {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const receivables = useCache<Receivable[]>(`receivables:${who}`, apiListReceivables, 'Gagal memuat daftar piutang.')
  const payables = useCache<Payable[]>(`payables:${who}`, apiListPayables, 'Gagal memuat daftar hutang.')
  const [tab, setTab] = useState<LedgerTab>('receivables')
  const [filter, setFilter] = useState<DebtFilter>('all')
  const [q, setQ] = useState('')
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  useEffect(() => { setPage(0) }, [tab, filter, q])
  const [selected, setSelected] = useState<DetailSelection | null>(null)
  const selectedRef = useRef<DetailSelection | null>(null)
  const requestRef = useRef(0)
  const [detail, setDetail] = useState<LoadedDetail | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailError, setDetailError] = useState('')
  const [paymentAmount, setPaymentAmount] = useState('')
  const [paymentMethod, setPaymentMethod] = useState<StandardPayMethod>('Cash')
  const [paymentBusy, setPaymentBusy] = useState(false)
  const [paymentNeedsRefresh, setPaymentNeedsRefresh] = useState(false)
  const [paymentError, setPaymentError] = useState('')

  const list = tab === 'receivables' ? receivables : payables
  const sourceItems = tab === 'receivables' ? receivables.data ?? [] : payables.data ?? []
  const visibleFilters = tab === 'receivables' ? [...FILTERS, { id: 'overdue' as const, label: 'Jatuh Tempo' }] : FILTERS
  const items = (tab === 'receivables'
    ? filterDebtItems(receivables.data ?? [], filter)
    : filterDebtItems(payables.data ?? [], filter)).filter((item) => {
      const term = q.trim().toLocaleLowerCase('id-ID')
      if (!term) return true
      const row = item as { customer_name?: string; supplier_name?: string; po_number?: string; purchase_order_number?: string }
      const hay = tab === 'receivables' ? row.customer_name ?? '' : `${row.supplier_name ?? ''} ${row.po_number ?? row.purchase_order_number ?? ''}`
      return hay.toLocaleLowerCase('id-ID').includes(term)
    })
  const pages = Math.max(1, Math.ceil(items.length / pageSize))
  const safePage = Math.min(page, pages - 1)
  const pageItems = items.slice(safePage * pageSize, safePage * pageSize + pageSize)

  function changeTab(next: LedgerTab) {
    setTab(next)
    setFilter('all')
  }

  async function openDetail(selection: DetailSelection) {
    const token = ++requestRef.current
    selectedRef.current = selection
    setSelected(selection)
    setDetail(null)
    setDetailLoading(true)
    setDetailError('')
    setPaymentAmount('')
    setPaymentMethod('Cash')
    setPaymentNeedsRefresh(false)
    setPaymentError('')
    try {
      const value = selection.kind === 'receivables'
        ? await apiGetReceivable(selection.id)
        : await apiGetPayable(selection.id)
      if (token === requestRef.current) setDetail({ kind: selection.kind, value } as LoadedDetail)
    } catch (e) {
      if (token === requestRef.current) setDetailError(e instanceof Error ? e.message : 'Gagal memuat detail.')
    } finally {
      if (token === requestRef.current) setDetailLoading(false)
    }
  }

  function closeDetail() {
    requestRef.current++
    selectedRef.current = null
    setSelected(null)
    setDetail(null)
    setDetailError('')
    setPaymentNeedsRefresh(false)
    setPaymentError('')
  }

  async function savePayment() {
    if (!selected || !detail) return
    if (paymentNeedsRefresh) return
    if (detail.value.status === 'paid') return
    if (!isDebtPaymentAmountValid(paymentAmount, detail.value.remaining)) {
      setPaymentError(`Masukkan jumlah Rupiah bulat lebih dari 0 dan tidak melebihi sisa ${fmtRp(detail.value.remaining)}.`)
      return
    }

    const selection = selected
    const token = requestRef.current
    const body: DebtPaymentInput = { amount: Number(paymentAmount), method: paymentMethod }
    setPaymentBusy(true)
    setPaymentError('')
    try {
      if (selection.kind === 'receivables') await apiAddReceivablePayment(selection.id, body)
      else await apiAddPayablePayment(selection.id, body)
    } catch (e) {
      setPaymentError(e instanceof Error ? e.message : 'Gagal mencatat pembayaran.')
      setPaymentBusy(false)
      return
    }

    setPaymentAmount('')
    setPaymentNeedsRefresh(true)

    if (selection.kind === 'receivables') receivables.reload()
    else payables.reload()

    try {
      const value = selection.kind === 'receivables'
        ? await apiGetReceivable(selection.id)
        : await apiGetPayable(selection.id)
      if (token === requestRef.current && selectedRef.current?.id === selection.id && selectedRef.current.kind === selection.kind) {
        setDetail({ kind: selection.kind, value } as LoadedDetail)
        setPaymentNeedsRefresh(false)
      }
    } catch {
      if (token === requestRef.current) setPaymentError('Pembayaran tercatat, tetapi detail belum terbarui. Tutup lalu buka kembali.')
    } finally {
      setPaymentBusy(false)
    }
  }

  const currentDetail = detail && selected && detail.kind === selected.kind ? detail.value : null
  const canPay = !!currentDetail && currentDetail.status !== 'paid' && canSubmitDebtPayment(paymentAmount, currentDetail.remaining, paymentNeedsRefresh)

  return (
    <>
      <PageHeader title="Hutang Piutang" crumb="Hutang Piutang" />

      <Tabs value={tab} onValueChange={(v) => changeTab(v as LedgerTab)} className="mb-5">
        <TabsList variant="line" aria-label="Jenis catatan hutang piutang">
          {TABS.map((item) => (
            <TabsTrigger key={item.id} value={item.id}>{item.label}</TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      <div className="mb-4 flex flex-col gap-3 sm:flex-row">
        <SearchInput value={q} onChange={(e) => setQ(e.target.value)} placeholder={tab === 'receivables' ? 'Cari nama pelanggan…' : 'Cari supplier atau nomor PO…'} aria-label="Cari catatan" />
        <Select
          label="Filter status"
          value={filter}
          onChange={(v) => setFilter(v as DebtFilter)}
          className="shrink-0 sm:w-56"
          options={visibleFilters.map((item) => ({ value: item.id, label: item.id === 'all' ? 'Semua status' : item.label }))}
        />
      </div>

      {list.err && <AlertStrip className="mb-4">{list.err}</AlertStrip>}

      {sourceItems.length === 0 && list.err ? null : items.length === 0 && !list.loading ? (
        <Empty
          title={sourceItems.length === 0 ? `Belum ada ${tab === 'receivables' ? 'piutang' : 'hutang'}` : 'Tidak ada data untuk filter ini'}
          sub={sourceItems.length === 0 ? 'Catatan akan muncul setelah transaksi Kredit atau penerimaan PO.' : 'Pilih filter status lain untuk melihat data.'}
        />
      ) : (
        <>
          <TableWrap className="hidden min-[1280px]:block">
          <table className="w-full border-collapse">
            {tab === 'receivables' ? (
              <thead><tr><Th>Pelanggan</Th><Th>Transaksi</Th><Th>Jatuh tempo</Th><Th right>Jumlah</Th><Th right>Sisa</Th><Th>Status</Th><Th right>Aksi</Th></tr></thead>
            ) : (
              <thead><tr><Th>Supplier</Th><Th>Nomor PO</Th><Th>Tanggal</Th><Th right>Jumlah</Th><Th right>Sisa</Th><Th>Status</Th><Th right>Aksi</Th></tr></thead>
            )}
            {list.loading && sourceItems.length === 0 ? <SkeletonRows cols={7} /> : (
              <tbody>
                {tab === 'receivables' ? (pageItems as Receivable[]).map((item) => (
                  <tr key={item.id}>
                    <Td><span className="font-medium text-fg">{item.customer_name || `Pelanggan #${item.customer_id}`}</span></Td>
                    <Td mono>{item.transaction_id}</Td>
                    <Td>{fmtDate(item.due_date)}</Td>
                    <Td right>{fmtRp(item.amount)}</Td>
                    <Td right><span className="font-mono font-medium tabular-nums">{fmtRp(item.remaining)}</span></Td>
                    <Td>{statusPill(item.status, item.due_date)}</Td>
                    <Td><RowAction title="Detail" onClick={() => void openDetail({ kind: 'receivables', id: item.id })} ><Eye className="size-4" /></RowAction></Td>
                  </tr>
                )) : (pageItems as Payable[]).map((item) => (
                  <tr key={item.id}>
                    <Td><span className="font-medium text-fg">{item.supplier_name || `Supplier #${item.supplier_id}`}</span></Td>
                    <Td mono>{item.po_number || item.purchase_order_number || `PO #${item.purchase_order_id}`}</Td>
                    <Td>{fmtDate(item.created_at)}</Td>
                    <Td right>{fmtRp(item.amount)}</Td>
                    <Td right><span className="font-mono font-medium tabular-nums">{fmtRp(item.remaining)}</span></Td>
                    <Td>{statusPill(item.status)}</Td>
                    <Td><RowAction title="Detail" onClick={() => void openDetail({ kind: 'payables', id: item.id })} ><Eye className="size-4" /></RowAction></Td>
                  </tr>
                ))}
              </tbody>
            )}
          </table>
        </TableWrap>
        <CardGrid>
          {list.loading && sourceItems.length === 0 ? (
            [0, 1, 2].map((i) => <Skeleton key={i} className="h-[128px] w-full rounded-xl" />)
          ) : tab === 'receivables' ? (pageItems as Receivable[]).map((item) => (
            <DataCard key={item.id}>
              <div className="flex items-center justify-between gap-2">
                <p className="min-w-0 truncate font-medium text-fg">{item.customer_name || `Pelanggan #${item.customer_id}`}</p>
                {statusPill(item.status, item.due_date)}
              </div>
              <p className="mt-1.5 text-right text-[17px] font-semibold tabular-nums text-fg">{fmtRp(item.remaining)}</p>
              <p className="mt-1 text-xs leading-snug text-fog">
                Sisa dari {fmtRp(item.amount)} · Trx {item.transaction_id} · Jatuh tempo {fmtDate(item.due_date)}
              </p>
              <div className="mt-auto flex items-center gap-1 pt-1.5">
                <RowAction title="Detail" label={`Detail piutang ${item.customer_name || item.id}`} onClick={() => void openDetail({ kind: 'receivables', id: item.id })}><Eye className="size-4" /></RowAction>
              </div>
            </DataCard>
          )) : (pageItems as Payable[]).map((item) => (
            <DataCard key={item.id}>
              <div className="flex items-center justify-between gap-2">
                <p className="min-w-0 truncate font-medium text-fg">{item.supplier_name || `Supplier #${item.supplier_id}`}</p>
                {statusPill(item.status)}
              </div>
              <p className="mt-1.5 text-right text-[17px] font-semibold tabular-nums text-fg">{fmtRp(item.remaining)}</p>
              <p className="mt-1 text-xs leading-snug text-fog">
                Sisa dari {fmtRp(item.amount)} · {item.po_number || item.purchase_order_number || `PO #${item.purchase_order_id}`} · {fmtDate(item.created_at)}
              </p>
              <div className="mt-auto flex items-center gap-1 pt-1.5">
                <RowAction title="Detail" label={`Detail hutang ${item.supplier_name || item.id}`} onClick={() => void openDetail({ kind: 'payables', id: item.id })}><Eye className="size-4" /></RowAction>
              </div>
            </DataCard>
          ))}
        </CardGrid>
        <Pager page={safePage} total={pages} onChange={setPage} totalItems={items.length} pageSize={pageSize} onPageSizeChange={(n) => { setPageSize(n); setPage(0) }} />
        </>
      )}

      <Modal
        open={selected !== null}
        title={selected?.kind === 'receivables' ? 'Detail Piutang' : 'Detail Hutang'}
        onClose={() => { if (!paymentBusy) closeDetail() }}
        wide
      >
        {detailLoading ? <p className="py-8 text-center text-sm text-muted" role="status">Memuat detail…</p> : detailError ? (
          <AlertStrip>{detailError}</AlertStrip>
        ) : currentDetail && selected && (
          <div className="space-y-4">
            <div className="rounded-xl bg-surface px-4 py-3">
              <MetaList items={[
                ...(selected.kind === 'receivables' && detail?.kind === 'receivables' ? [
                  { label: 'Pelanggan', value: detail.value.customer_name || `Pelanggan #${detail.value.customer_id}` },
                  { label: 'Transaksi', value: detail.value.transaction_id, mono: true },
                  { label: 'Jatuh tempo', value: fmtDate(detail.value.due_date) },
                ] : selected.kind === 'payables' && detail?.kind === 'payables' ? [
                  { label: 'Supplier', value: detail.value.supplier_name || `Supplier #${detail.value.supplier_id}` },
                  { label: 'Purchase order', value: detail.value.po_number || detail.value.purchase_order_number || `PO #${detail.value.purchase_order_id}`, mono: true },
                  { label: 'Tanggal', value: fmtDate(detail.value.created_at) },
                ] : []),
                { label: 'Jumlah', value: fmtRp(currentDetail.amount) },
                { label: 'Sisa', value: fmtRp(currentDetail.remaining) },
                { label: 'Status', value: statusPill(currentDetail.status, detail?.kind === 'receivables' ? detail.value.due_date : undefined) },
              ]} />
            </div>

            <section className="space-y-2">
              <h3 className="text-sm font-medium text-fg">Riwayat pembayaran</h3>
              {currentDetail.payments.length === 0 ? <p className="rounded-lg bg-cream px-3.5 py-3 text-[13px] text-muted">Belum ada pembayaran.</p> : (
                <div className="divide-y divide-dove rounded-xl bg-cream px-4">
                  {currentDetail.payments.map((payment) => (
                    <div key={payment.id} className="flex items-center justify-between gap-3 py-3 text-sm">
                      <span className="min-w-0"><span className="font-medium text-fg">{payment.method}</span><span className="ml-2 text-xs text-muted">{fmtDate(payment.created_at)}</span></span>
                      <span className="shrink-0 font-mono tabular-nums text-fg">{fmtRp(payment.amount)}</span>
                    </div>
                  ))}
                </div>
              )}
            </section>

            {currentDetail.status === 'paid' ? (
              <p className="rounded-lg bg-surface px-3.5 py-3 text-[13px] text-muted">Catatan ini sudah lunas.</p>
            ) : (
              <form className="space-y-3 rounded-xl bg-surface p-4" onSubmit={(event) => { event.preventDefault(); void savePayment() }}>
                <h3 className="text-sm font-medium text-fg">Catat pembayaran</h3>
                <NumInput
                  label="Jumlah pembayaran"
                  value={paymentAmount}
                  onValue={setPaymentAmount}
                  placeholder="0"
                  required
                  disabled={paymentBusy || paymentNeedsRefresh}
                  hint={`Maksimal ${fmtRp(currentDetail.remaining)}.`}
                />
                <div>
                  <p className="mb-2 text-[13px] font-medium text-steel">Metode pembayaran</p>
                  <div className="flex flex-wrap gap-2">
                    {STANDARD_METHODS.map((method) => (
                      <button
                        key={method}
                        type="button"
                        aria-pressed={paymentMethod === method}
                        disabled={paymentBusy || paymentNeedsRefresh}
                        onClick={() => setPaymentMethod(method)}
                        className={`rounded-full border px-3 py-1.5 text-xs transition ${paymentMethod === method ? 'border-jet bg-jet text-paper' : 'border-dove bg-paper text-muted hover:border-jet'}`}
                      >
                        {method}
                      </button>
                    ))}
                  </div>
                </div>
                {paymentError && <AlertStrip>{paymentError}</AlertStrip>}
                <ModalActions>
                  <Button type="submit" disabled={paymentBusy || !canPay}>{paymentBusy ? 'Menyimpan…' : 'Catat Pembayaran'}</Button>
                </ModalActions>
              </form>
            )}
          </div>
        )}
      </Modal>
    </>
  )
}
