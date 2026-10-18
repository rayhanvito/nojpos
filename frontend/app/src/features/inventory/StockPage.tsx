// Stok — inventory monitoring workspace: ringkasan → cari/filter → tab stok/riwayat → tindakan.
// Token font/warna milik sistem (tidak ada token baru di file ini).
import { useEffect, useMemo, useState } from 'react'
import { useNavigate } from 'react-router'
import { CircleCheck, OctagonX, Package, TriangleAlert, SlidersHorizontal, ExternalLink } from 'lucide-react'
import { type Movement, type Product } from './api'
import { useAdjustStock, useMovements, useStockCategories, useStockProducts } from './hooks'
import { PageHeader } from '../../shared/PageHeader'
import { fmtDate, fmtTime } from '../../shared/session/store'
import { AlertStrip, Button, CardGrid, DataCard, Empty, FilterChips, FilterField, FilterPanel, Input, Modal, NumInput, Pager, Pill, RowAction, SearchInput, Select, SkeletonRows, StatCard, TableWrap, Td, Th } from '../../shared/ui'
import { Tabs, TabsList, TabsTrigger } from '@/shared/ui/primitives/tabs'
import { STOCK_TABS, STOCK_TAB_ALIASES, type StockTab } from '../../shared/tab-params'
import { useTabParam } from '../../shared/use-tab-param'
import StockOpname from './StockOpnamePage'
import TransferTab from './TransferTab'
import { Card, CardContent } from '@/shared/ui/primitives/card'
import { Skeleton } from '@/shared/ui/primitives/skeleton'

const TYPE_LABEL: Record<Movement['type'], string> = {
  sale: 'Penjualan', refund: 'Refund', adjust: 'Penyesuaian', initial: 'Stok awal',
}

const NONE = '__none__'

const KONDISI_OPTS = [{ value: '', label: 'Semua kondisi' }, { value: 'aman', label: 'Aman' }, { value: 'menipis', label: 'Menipis' }, { value: 'habis', label: 'Habis' }]

const stockFor = (product: Product) => product.variants?.length
  ? product.variants.reduce((sum, variant) => sum + variant.stock, 0)
  : product.stock
const isLow = (product: Product) => product.variants?.length
  ? product.variants.some((variant) => variant.min_stock > 0 && variant.stock <= variant.min_stock)
  : product.stock > 0 && (product.min_stock ?? 0) > 0 && product.stock <= (product.min_stock ?? 0)

export default function Stock() {
  const navigate = useNavigate()
  const [mpage, setMpage] = useState(0)
  const [movPageSize, setMovPageSize] = useState(10)
  const [mtype, setMtype] = useState('')
  const [mq, setMq] = useState('')
  const prod = useStockProducts()
  const products = prod.data
  const mov = useMovements(mpage, movPageSize, mtype, mq)
  const movements = mov.data?.items
  const catq = useStockCategories()
  const cats = (catq.data ?? []).filter((c) => c.active)
  const adjustStock = useAdjustStock()

  const [err, setErr] = useState('')
  const [q, setQ] = useState('')
  const [catFilter, setCatFilter] = useState('')
  const [kondisiF, setKondisiF] = useState('')
  const [tab, setTab] = useTabParam<StockTab>(STOCK_TABS, STOCK_TAB_ALIASES, 'stock')
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const [adjustFor, setAdjustFor] = useState<Product | null>(null)
  const [qty, setQty] = useState('')
  const [reason, setReason] = useState('')
  const [type, setType] = useState<'plus' | 'minus'>('plus')
  const [busy, setBusy] = useState(false)

  const actives = useMemo(() => (products ?? []).filter((p) => p.active), [products])

  const counts = useMemo(() => {
    let aman = 0, menipis = 0, habis = 0
    for (const p of actives) {
      if (stockFor(p) === 0) habis++
      else if (isLow(p)) menipis++
      else aman++
    }
    return { total: actives.length, aman, menipis, habis }
  }, [actives])

  const filtered = useMemo(() => {
    const needle = q.trim().toLowerCase()
    return actives.filter((p) => {
      if (catFilter === NONE && p.category_id) return false
      if (catFilter && catFilter !== NONE && p.category_id !== catFilter) return false
      if (needle && !`${p.name} ${p.sku} ${(p.variants ?? []).map((variant) => variant.sku).join(' ')}`.toLowerCase().includes(needle)) return false
      if (kondisiF === 'habis' && stockFor(p) !== 0) return false
      if (kondisiF === 'menipis' && (stockFor(p) === 0 || !isLow(p))) return false
      if (kondisiF === 'aman' && (stockFor(p) === 0 || isLow(p))) return false
      return true
    })
  }, [actives, q, catFilter, kondisiF])

  useEffect(() => { setPage(0) }, [q, catFilter, tab])
  useEffect(() => { setMpage(0) }, [tab])

  // ponytail: paging client-side di atas fetchAll; server-side bila katalog puluhan ribu.
  const stockPages = Math.max(1, Math.ceil(filtered.length / pageSize))
  const stockPage = Math.min(page, stockPages - 1)
  const stockItems = filtered.slice(stockPage * pageSize, stockPage * pageSize + pageSize)

  // Riwayat kini server-side: total dari server, halaman aktif tidak lagi
  // terpotong di 50 baris pertama seperti paging client sebelumnya.
  const movTotal = mov.data?.total ?? 0
  const movPages = Math.max(1, Math.ceil(movTotal / movPageSize))
  const movPage = Math.min(mpage, movPages - 1)
  const movItems = movements ?? []
  useEffect(() => { if (mov.data && mpage > movPages - 1) setMpage(movPages - 1) }, [mov.data, mpage, movPages])

  function load() {
    prod.reload()
    mov.reload()
  }

  async function submit() {
    const n = Number(qty)
    const r = reason.trim()
    if (!adjustFor || !Number.isFinite(n) || n <= 0 || !r) return
    setBusy(true); setErr('')
    try {
      await adjustStock.mutateAsync({ productId: adjustFor.id, direction: type, qty: n, reason: r })
      setAdjustFor(null); setQty(''); setReason(''); setType('plus')
      load()
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'Gagal menyesuaikan stok.')
    } finally {
      setBusy(false)
    }
  }

  const stats = [
    { label: 'Total Produk', value: counts.total, icon: Package },
    { label: 'Stok Aman', value: counts.aman, icon: CircleCheck },
    { label: 'Stok Menipis', value: counts.menipis, icon: TriangleAlert },
    { label: 'Stok Habis', value: counts.habis, icon: OctagonX },
  ]


  return (
    <>
      <PageHeader title="Stok" crumb="Stok" />

      {(err || prod.err) && <AlertStrip className="mb-4">{err || prod.err}</AlertStrip>}

      <div className="grid grid-cols-2 gap-3 sm:gap-4 xl:grid-cols-4">
        {!products ? (
          [0, 1, 2, 3].map((i) => (
            <Card key={i} aria-hidden="true">
              <CardContent className="space-y-3 p-5">
                <div className="flex items-center justify-between">
                  <Skeleton className="h-4 w-20" />
                  <Skeleton className="size-9 rounded-lg" />
                </div>
                <Skeleton className="h-7 w-16" />
              </CardContent>
            </Card>
          ))
        ) : (
          stats.map((s) => (
            <StatCard key={s.label} density="stat" label={s.label} value={String(s.value)} icon={s.icon} />
          ))
        )}
      </div>

      <Tabs value={tab} onValueChange={(v) => setTab(v as StockTab)} className="mt-5">
        <TabsList variant="line" aria-label="Jenis tampilan stok">
          {([
            { id: 'stock', label: 'Stok Saat Ini' },
            { id: 'stocktake', label: 'Opname' },
            { id: 'movement', label: 'Riwayat Pergerakan' },
            { id: 'transfers', label: 'Transfer' },
          ] as const).map((t) => (
            <TabsTrigger key={t.id} value={t.id}>{t.label}</TabsTrigger>
          ))}
        </TabsList>
      </Tabs>

      {tab === 'stocktake' && <StockOpname embedded />}

      {tab === 'stock' && (
        <div className="mt-4 flex flex-col gap-3 sm:flex-row">
          <SearchInput value={q} onChange={(e) => setQ(e.target.value)} placeholder="Cari nama atau SKU…" aria-label="Cari produk" />
          <Select
            label="Filter kategori"
            className="sm:w-56"
            value={catFilter}
            onChange={setCatFilter}
            options={[
              { value: '', label: 'Semua kategori' },
              ...cats.map((c) => ({ value: c.id, label: c.name })),
              { value: NONE, label: 'Tanpa kategori' },
            ]}
          />
          <FilterPanel
            className="shrink-0 sm:w-44"
            activeCount={kondisiF ? 1 : 0}
            onReset={() => { setKondisiF(''); setPage(0) }}
          >
            <FilterField label="Kondisi stok">
              <Select label="Filter kondisi stok" value={kondisiF} onChange={(v) => { setKondisiF(v); setPage(0) }} options={KONDISI_OPTS} />
            </FilterField>
          </FilterPanel>
        </div>
      )}
      {tab === 'stock' && (
        <FilterChips items={kondisiF ? [{ key: 'kondisi', label: `Kondisi: ${KONDISI_OPTS.find((o) => o.value === kondisiF)?.label ?? kondisiF}`, onClear: () => { setKondisiF(''); setPage(0) } }] : []} />
      )}

      {tab === 'stock' ? (
        <div className="mt-4">
          <TableWrap className="hidden min-[1280px]:block">
            {!products ? (
              <table className="w-full border-collapse">
                <thead>
                  <tr>
                    <Th>Produk</Th><Th>SKU</Th><Th>Kategori</Th><Th>Stok</Th><Th>Status</Th><Th>Aksi</Th>
                  </tr>
                </thead>
                <SkeletonRows cols={6} />
              </table>
            ) : filtered.length > 0 ? (
              <table className="w-full border-collapse">
                <thead>
                  <tr>
                    <Th>Produk</Th><Th>SKU</Th><Th>Kategori</Th><Th>Stok</Th><Th>Status</Th><Th>Aksi</Th>
                  </tr>
                </thead>
                <tbody>
                  {stockItems.map((p) => (
                    <tr key={p.id}>
                      <Td><span className="text-[15px] font-medium text-fg">{p.name}</span></Td>
                      <Td mono>{p.sku}</Td>
                      <Td>{p.category_name ?? 'Tanpa kategori'}</Td>
                      <Td><StockQty product={p} /></Td>
                      <Td><StockPill product={p} /></Td>
                      <Td>
                        <div className="flex">
                          {p.variants?.length
                            ? <RowAction title="Ubah stok di Produk" label={`Ubah stok ${p.name} di Produk`} onClick={() => navigate(`/app/products?edit=${p.id}`)}><ExternalLink className="size-4" /></RowAction>
                            : <RowAction title="Penyesuaian stok" label={`Penyesuaian stok ${p.name}`} onClick={() => setAdjustFor(p)}><SlidersHorizontal className="size-4" /></RowAction>}
                        </div>
                      </Td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <Empty
                title={products.length > 0 ? 'Tidak ada produk yang cocok' : 'Belum ada produk'}
                sub={products.length > 0 ? 'Coba kata kunci lain atau ubah filter kategori.' : 'Tambah produk dari menu Produk untuk mulai memantau stok.'}
                action={products.length > 0 ? <Button onClick={() => { setCatFilter(''); setQ('') }}>Tampilkan semua</Button> : undefined}
              />
            )}
          </TableWrap>

          <CardGrid>
            {!products ? (
              [0, 1, 2].map((i) => <Skeleton key={i} className="h-24 w-full rounded-xl" />)
            ) : filtered.length > 0 ? (
              stockItems.map((p) => (
                <DataCard key={p.id}>
                  <div className="flex items-start justify-between gap-2">
                    <p className="min-w-0 flex-1 truncate text-[15px] font-medium text-fg">{p.name}</p>
                    <StockPill product={p} />
                  </div>
                  <p className="mt-1 font-mono text-xs text-fog">{p.sku} · {p.category_name ?? 'Tanpa kategori'}</p>
                  <div className="mt-2.5 flex items-center justify-between">
                    <StockQty product={p} />
                    {p.variants?.length
                      ? <RowAction title="Ubah stok di Produk" label={`Ubah stok ${p.name} di Produk`} onClick={() => navigate(`/app/products?edit=${p.id}`)}><ExternalLink className="size-4" /></RowAction>
                      : <RowAction title="Penyesuaian stok" label={`Penyesuaian stok ${p.name}`} onClick={() => setAdjustFor(p)}><SlidersHorizontal className="size-4" /></RowAction>}
                  </div>
                </DataCard>
              ))
            ) : (
              <p className="py-10 text-center text-sm text-fog">
                {products.length > 0 ? 'Tidak ada produk yang cocok.' : 'Belum ada produk.'}
              </p>
            )}
          </CardGrid>

          <Pager page={stockPage} total={stockPages} onChange={setPage} totalItems={filtered.length} pageSize={pageSize} onPageSizeChange={(n) => { setPageSize(n); setPage(0) }} />
        </div>
      ) : tab === 'movement' ? (        <div className="mt-4">
          <div className="mb-4 flex flex-col gap-3 sm:flex-row">
            <SearchInput value={mq} onChange={(e) => { setMq(e.target.value); setMpage(0) }} placeholder="Cari nama produk…" aria-label="Cari pergerakan berdasarkan produk" />
            <Select
              label="Filter jenis pergerakan"
              value={mtype}
              onChange={(v) => { setMtype(v); setMpage(0) }}
              className="shrink-0 sm:w-56"
              options={[{ value: '', label: 'Semua jenis' }, ...Object.entries(TYPE_LABEL).map(([value, label]) => ({ value, label }))]}
            />
          </div>
          <TableWrap className="hidden min-[1280px]:block">
            {!movements ? (
              <table className="w-full border-collapse">
                <thead>
                  <tr>
                    <Th>Waktu</Th><Th>Produk</Th><Th>Jenis</Th><Th right>Qty</Th><Th>Alasan</Th><Th>Aktor</Th>
                  </tr>
                </thead>
                <SkeletonRows cols={6} />
              </table>
            ) : movements.length > 0 ? (
              <table className="w-full border-collapse">
                <thead>
                  <tr>
                    <Th>Waktu</Th><Th>Produk</Th><Th>Jenis</Th><Th right>Qty</Th><Th>Alasan</Th><Th>Aktor</Th>
                  </tr>
                </thead>
                <tbody>
                  {movItems.map((m) => (
                    <tr key={m.id}>
                      <Td mono>{fmtDate(m.created_at)} {fmtTime(m.created_at)}</Td>
                      <Td>{m.product_name ?? '—'}</Td>
                      <Td><Pill tone={m.type === 'sale' ? 'ok' : m.type === 'refund' ? 'warn' : 'muted'}>{TYPE_LABEL[m.type]}</Pill></Td>
                      <Td right><span className={`font-mono tabular-nums ${m.qty > 0 ? 'text-sprout' : 'text-ember'}`}>{m.qty > 0 ? '+' : ''}{m.qty}</span></Td>
                      <Td>{m.reason}</Td>
                      <Td>{m.actor}</Td>
                    </tr>
                  ))}
                </tbody>
              </table>
            ) : (
              <p className="py-10 text-center text-sm text-fog">Belum ada pergerakan stok.</p>
            )}
          </TableWrap>

          <CardGrid>
            {!movements ? (
              [0, 1, 2].map((i) => <Skeleton key={i} className="h-20 w-full rounded-xl" />)
            ) : movements.length > 0 ? (
              movItems.map((m) => (
                <DataCard key={m.id}>
                  <div className="flex items-center justify-between gap-2">
                    <p className="min-w-0 flex-1 truncate text-sm font-medium text-fg">{m.product_name ?? '—'}</p>
                    <span className={`shrink-0 font-mono text-sm tabular-nums ${m.qty > 0 ? 'text-sprout' : 'text-ember'}`}>{m.qty > 0 ? '+' : ''}{m.qty}</span>
                  </div>
                  <p className="mt-1 font-mono text-xs text-fog">{fmtDate(m.created_at)} {fmtTime(m.created_at)} · {TYPE_LABEL[m.type]} · {m.actor}</p>
                  {m.reason && <p className="mt-1 truncate text-[13px] text-muted">{m.reason}</p>}
                </DataCard>
              ))
            ) : (
              <p className="py-10 text-center text-sm text-fog">Belum ada pergerakan stok.</p>
            )}
          </CardGrid>

          <Pager page={movPage} total={movPages} onChange={setMpage} totalItems={movTotal} pageSize={movPageSize} onPageSizeChange={(n) => { setMovPageSize(n); setMpage(0) }} />
        </div>
      ) : tab === 'transfers' ? (
        <TransferTab />
      ) : null}

      <Modal open={!!adjustFor} title={`Penyesuaian stok · ${adjustFor?.name ?? ''}`} onClose={() => setAdjustFor(null)}>
        {adjustFor && (
          <div className="space-y-4">
            <p className="text-sm text-muted">Stok saat ini: <strong className="text-fg">{adjustFor.stock} {adjustFor.unit}</strong></p>
            <div className="flex gap-2">
              <button
                onClick={() => setType('plus')}
                className={`flex-1 rounded-full border py-2 text-sm ${type === 'plus' ? 'border-jet bg-jet text-paper' : 'border-dove text-muted'}`}
              >
                + Tambah
              </button>
              <button
                onClick={() => setType('minus')}
                className={`flex-1 rounded-full border py-2 text-sm ${type === 'minus' ? 'border-jet bg-jet text-paper' : 'border-dove text-muted'}`}
              >
                − Kurangi
              </button>
            </div>
            <NumInput label="Jumlah" value={qty} onValue={setQty} placeholder="0" />
            <Input label="Alasan (wajib)" value={reason} onChange={setReason} placeholder="cth: barang rusak, stok fisik berbeda" />
            <Button className="w-full" onClick={submit} disabled={!Number(qty) || !reason.trim() || busy}>{busy ? 'Menyimpan…' : 'Simpan Penyesuaian'}</Button>
          </div>
        )}
      </Modal>
    </>
  )
}

function StockPill({ product }: { product: Product }) {
  const stock = stockFor(product)
  if (stock === 0) return <Pill tone="warn">Habis</Pill>
  if (isLow(product)) return <Pill tone="warn">Menipis</Pill>
  return <Pill>Aman</Pill>
}

function StockQty({ product }: { product: Product }) {
  const stock = stockFor(product)
  const low = isLow(product)
  const dot = stock === 0 ? 'var(--t-ember)' : low ? 'var(--t-sunbeam)' : 'var(--t-sprout)'
  return (
    <span className={`inline-flex items-center gap-1.5 text-[15px] tabular-nums ${stock === 0 || low ? 'font-semibold text-fg' : 'text-muted'}`}>
      <span className="size-2 shrink-0 rounded-full" style={{ background: dot }} aria-hidden="true" />
      {stock} {product.unit}
    </span>
  )
}
