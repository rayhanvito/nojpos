import { useRef, useState } from 'react'
import { useSearchParams } from 'react-router'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import Pengeluaran from './Pengeluaran'
import { Plus, Trash2, Eye } from 'lucide-react'
import {
  apiCancelPurchaseOrder, apiCreatePurchaseOrder, apiDeletePurchaseOrder, apiGetPurchaseOrder,
  apiListPurchaseOrders, apiListProducts, apiListSuppliers, apiOrderPurchaseOrder,
  apiReceivePurchaseOrder, apiUpdatePurchaseOrder, fetchAll,
  type Product, type PurchaseOrder, type PurchaseOrderDetail, type PurchaseOrderInput,
  type PurchaseOrderStatus, type Supplier,
} from '../lib/api'
import { PageHeader } from '../lib/PageHeader'
import { useCache } from '../lib/cache'
import { fmtDate, fmtRp, useDB } from '../lib/store'
import { isCurrentPurchaseOrderRequest, purchaseOrderItemName, purchaseOrderProductId, purchaseOrderTotal, remainingToReceive } from '../lib/purchase-order'
import { ModalActions, AlertStrip, Button, CardGrid, DataCard, Empty, Modal, NumInput, Pager, Pill, RowAction, SearchInput, Select, SkeletonRows, TableWrap, Td, Th } from '../lib/ui'
import { Skeleton } from '@/components/ui/skeleton'

type DraftItem = {
  key: number
  productId: string
  variantId: string
  variantName?: string
  qty: string
  buyPrice: string
}
type Draft = { id?: number; supplierId: string; items: DraftItem[] }
type Confirmation = { action: 'cancel' | 'delete'; order: PurchaseOrderDetail }

const STATUS_LABEL: Record<PurchaseOrderStatus, string> = {
  draft: 'Draf',
  ordered: 'Dipesan',
  partially_received: 'Diterima Sebagian',
  fully_received: 'Selesai',
  cancelled: 'Dibatalkan',
}

function statusTone(status: PurchaseOrderStatus): 'ok' | 'warn' | 'muted' {
  if (status === 'fully_received') return 'ok'
  if (status === 'ordered' || status === 'partially_received') return 'warn'
  return 'muted'
}

function orderStatus(status: PurchaseOrderStatus) {
  return <Pill tone={statusTone(status)}>{STATUS_LABEL[status]}</Pill>
}

function emptyItem(key: number): DraftItem {
  return { key, productId: '', variantId: '', qty: '', buyPrice: '' }
}

function draftFromOrder(order: PurchaseOrderDetail): Draft {
  return {
    id: order.id,
    supplierId: String(order.supplier_id),
    items: order.items.map((item, index) => ({
      key: index + 1,
      productId: String(item.product_id),
      variantId: item.variant_id == null ? '' : String(item.variant_id),
      variantName: item.variant_name,
      qty: String(item.qty_ordered),
      buyPrice: String(item.buy_price),
    })),
  }
}


export default function Pembelian() {
  const { session } = useDB()
  const who = `${session?.id}:${session?.role}`
  const orders = useCache<PurchaseOrder[]>(`purchase-orders:${who}`, apiListPurchaseOrders, 'Gagal memuat PO.')
  const suppliers = useCache<Supplier[]>(`suppliers:${who}`, apiListSuppliers, 'Gagal memuat supplier.')
  const products = useCache<Product[]>(`purchase-products:${who}`, () => fetchAll<Product>((page) => apiListProducts({ page, limit: 200 })), 'Gagal memuat produk.')
  const [selectedId, setSelectedId] = useState<number | null>(null)
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const [q, setQ] = useState('')
  const [statusF, setStatusF] = useState<PurchaseOrderStatus | ''>('')
  const [params] = useSearchParams()
  const [tab, setTab] = useState<'po' | 'pengeluaran'>(params.get('tab') === 'pengeluaran' ? 'pengeluaran' : 'po')
  const [expSignal, setExpSignal] = useState(0)
  const selectedIdRef = useRef<number | null>(null)
  const detailRequestRef = useRef(0)
  const [detail, setDetail] = useState<PurchaseOrderDetail | null>(null)
  const [detailLoading, setDetailLoading] = useState(false)
  const [detailError, setDetailError] = useState('')
  const [knownSupplierNames, setKnownSupplierNames] = useState<Record<string, string>>({})
  const [editing, setEditing] = useState<Draft | null>(null)
  const [formError, setFormError] = useState('')
  const [formBusy, setFormBusy] = useState(false)
  const [actionBusy, setActionBusy] = useState(false)
  const [actionError, setActionError] = useState('')
  const [receiveError, setReceiveError] = useState('')
  const [receiveQty, setReceiveQty] = useState<Record<number, string>>({})
  const [confirmation, setConfirmation] = useState<Confirmation | null>(null)
  const [confirmError, setConfirmError] = useState('')

  function rememberSupplier(order: PurchaseOrderDetail) {
    const name = order.supplier_name || order.supplier?.name
    if (name) setKnownSupplierNames((current) => ({ ...current, [`${who}:${order.id}`]: name }))
  }

  async function openDetail(order: PurchaseOrder) {
    const requestToken = ++detailRequestRef.current
    selectedIdRef.current = order.id
    setSelectedId(order.id)
    setDetail(null)
    setDetailError('')
    setActionError('')
    setReceiveError('')
    setDetailLoading(true)
    try {
      const value = await apiGetPurchaseOrder(order.id)
      if (selectedIdRef.current === order.id && isCurrentPurchaseOrderRequest(requestToken, detailRequestRef.current)) {
        rememberSupplier(value)
        setDetail(value)
        setReceiveQty(Object.fromEntries(value.items.map((item) => [item.id, '0'])))
      }
    } catch (e) {
      if (selectedIdRef.current === order.id && isCurrentPurchaseOrderRequest(requestToken, detailRequestRef.current)) setDetailError(e instanceof Error ? e.message : 'Gagal memuat detail PO.')
    } finally {
      if (selectedIdRef.current === order.id && isCurrentPurchaseOrderRequest(requestToken, detailRequestRef.current)) setDetailLoading(false)
    }
  }

  function closeDetail() {
    detailRequestRef.current++
    selectedIdRef.current = null
    setSelectedId(null)
    setDetail(null)
    setDetailError('')
    setActionError('')
    setReceiveError('')
  }

  async function refreshDetail(id: number, successAction: string) {
    const requestToken = ++detailRequestRef.current
    orders.reload()
    if (selectedIdRef.current === id) setDetailLoading(true)
    try {
      const value = await apiGetPurchaseOrder(id)
      if (selectedIdRef.current === id && isCurrentPurchaseOrderRequest(requestToken, detailRequestRef.current)) {
        rememberSupplier(value)
        setDetail(value)
        setReceiveQty(Object.fromEntries(value.items.map((item) => [item.id, '0'])))
        setDetailError('')
      }
    } catch {
      if (selectedIdRef.current === id && isCurrentPurchaseOrderRequest(requestToken, detailRequestRef.current)) setDetailError(`${successAction}, tetapi detail gagal dimuat ulang. Tutup lalu buka kembali PO ini.`)
    } finally {
      if (selectedIdRef.current === id && isCurrentPurchaseOrderRequest(requestToken, detailRequestRef.current)) setDetailLoading(false)
    }
  }

  function openCreate() {
    setFormError('')
    setEditing({ supplierId: '', items: [emptyItem(1)] })
  }

  function openEdit(order: PurchaseOrderDetail) {
    setFormError('')
    setEditing(draftFromOrder(order))
  }

  async function saveOrder(draft: Draft) {
    const supplierId = Number(draft.supplierId)
    if (!Number.isSafeInteger(supplierId) || supplierId <= 0) return setFormError('Pilih supplier terlebih dahulu.')
    if (draft.items.length === 0) return setFormError('Tambahkan minimal satu item produk.')

    const items: PurchaseOrderInput['items'] = []
    for (const item of draft.items) {
      const productId = purchaseOrderProductId(item.productId)
      const qty = Number(item.qty)
      const buyPrice = Number(item.buyPrice)
      const product = (products.data ?? []).find((entry) => purchaseOrderProductId(entry.id) === productId)
      if (productId === null || !Number.isSafeInteger(qty) || qty <= 0) return setFormError('Pilih produk dan isi qty pesan dengan bilangan bulat lebih dari 0.')
      if (!Number.isSafeInteger(buyPrice) || buyPrice < 0) return setFormError('Harga beli harus berupa Rupiah bulat dan tidak boleh negatif.')
      if (product?.variants?.length && !item.variantId) return setFormError(`Pilih varian untuk produk ${product.name}.`)
      items.push({
        productId,
        ...(item.variantId ? { variantId: Number(item.variantId) } : {}),
        qty,
        buyPrice,
      })
    }

    const body: PurchaseOrderInput = { supplierId, items }
    setFormError('')
    setFormBusy(true)
    try {
      if (draft.id) {
        await apiUpdatePurchaseOrder(draft.id, body)
        setEditing(null)
        await refreshDetail(draft.id, 'Perubahan PO tersimpan')
      } else {
        await apiCreatePurchaseOrder(body)
        setEditing(null)
        orders.reload()
      }
    } catch (e) {
      setFormError(e instanceof Error ? e.message : 'Gagal menyimpan PO.')
    } finally {
      setFormBusy(false)
    }
  }

  async function orderNow() {
    if (!detail) return
    setActionError('')
    setActionBusy(true)
    try {
      await apiOrderPurchaseOrder(detail.id)
      await refreshDetail(detail.id, 'Purchase order berhasil dipesan')
    } catch (e) {
      setActionError(e instanceof Error ? e.message : 'Gagal memesan PO.')
    } finally {
      setActionBusy(false)
    }
  }

  async function receiveItems() {
    if (!detail) return
    const items = detail.items.map((item) => ({ itemId: item.id, qty: Number(receiveQty[item.id] ?? '0') }))
    const selected = items.filter((item) => item.qty > 0)
    if (selected.length === 0) return setReceiveError('Isi jumlah yang diterima untuk minimal satu item.')
    for (const receipt of selected) {
        const item = detail.items.find((entry) => entry.id === receipt.itemId)!
      if (!Number.isSafeInteger(receipt.qty) || receipt.qty > remainingToReceive(item)) {
        return setReceiveError(`Qty terima untuk ${purchaseOrderItemName(item, products.data ?? [])} melebihi sisa yang dipesan (${remainingToReceive(item)}).`)
      }
    }

    setReceiveError('')
    setActionError('')
    setActionBusy(true)
    try {
      await apiReceivePurchaseOrder(detail.id, { items: selected })
      await refreshDetail(detail.id, 'Penerimaan berhasil disimpan')
    } catch (e) {
      setReceiveError(e instanceof Error ? e.message : 'Gagal menyimpan penerimaan.')
    } finally {
      setActionBusy(false)
    }
  }

  async function confirmOrderAction() {
    if (!confirmation) return
    setConfirmError('')
    setActionBusy(true)
    try {
      if (confirmation.action === 'delete') {
        await apiDeletePurchaseOrder(confirmation.order.id)
        setConfirmation(null)
        orders.reload()
        closeDetail()
      } else {
        await apiCancelPurchaseOrder(confirmation.order.id)
        setConfirmation(null)
        await refreshDetail(confirmation.order.id, 'Purchase order berhasil dibatalkan')
      }
    } catch (e) {
      setConfirmError(e instanceof Error ? e.message : 'Gagal memproses PO.')
    } finally {
      setActionBusy(false)
    }
  }

  const orderItems = orders.data ?? []
  const filteredOrders = orderItems.filter((order) => {
    if (statusF && order.status !== statusF) return false
    const term = q.trim().toLocaleLowerCase('id-ID')
    if (!term) return true
    const supplier = order.supplier_name || order.supplier?.name || knownSupplierNames[`${who}:${order.id}`] || ''
    return `${order.number} ${supplier}`.toLocaleLowerCase('id-ID').includes(term)
  })
  const pages = Math.max(1, Math.ceil(filteredOrders.length / pageSize))
  const safePage = Math.min(page, pages - 1)
  const pageItems = filteredOrders.slice(safePage * pageSize, safePage * pageSize + pageSize)
  const supplierItems = suppliers.data ?? []
  const productItems = products.data ?? []
  const supplierName = (id: number) => supplierItems.find((supplier) => supplier.id === id)?.name ?? `Supplier #${id}`
  const canReceive = detail?.status === 'ordered' || detail?.status === 'partially_received'
  const detailTotal = detail ? purchaseOrderTotal(detail.items.map((item) => ({ qty: item.qty_ordered, buyPrice: item.buy_price }))) : 0

  return (
    <>
      <PageHeader
        title="Pembelian"
       
        crumb="Pembelian"
        actions={tab === 'po'
          ? <Button onClick={openCreate}><Plus className="size-4" />Buat PO</Button>
          : <Button onClick={() => setExpSignal((n) => n + 1)}><Plus className="size-4" />Tambah Pengeluaran</Button>}
      />

      <Tabs value={tab} onValueChange={(v) => setTab(v as 'po' | 'pengeluaran')} className="mt-5">
        <TabsList variant="line" aria-label="Jenis pembelian">
          <TabsTrigger value="po">Pembelian</TabsTrigger>
          <TabsTrigger value="pengeluaran">Pengeluaran</TabsTrigger>
        </TabsList>
      </Tabs>

      {tab === 'po' && orders.err && <AlertStrip className="mt-4">{orders.err}</AlertStrip>}

      <div className="mt-5">
      {tab === 'pengeluaran' ? (
        <Pengeluaran createSignal={expSignal} />
      ) : orderItems.length === 0 && orders.err ? null : orderItems.length === 0 && !orders.loading ? (
        <Empty title="Belum ada PO" sub="PO akan muncul setelah dibuat." />
      ) : (
        <>
          <div className="mb-4 flex flex-col gap-3 sm:flex-row">
            <SearchInput value={q} onChange={(e) => { setQ(e.target.value); setPage(0) }} placeholder="Cari nomor PO atau supplier…" aria-label="Cari pembelian" />
            <Select
              label="Filter status PO"
              value={statusF}
              onChange={(v) => { setStatusF(v as PurchaseOrderStatus | ''); setPage(0) }}
              className="shrink-0 sm:w-56"
              options={[{ value: '', label: 'Semua status' }, ...Object.entries(STATUS_LABEL).map(([value, label]) => ({ value, label }))]}
            />
          </div>
          {filteredOrders.length === 0 && !orders.loading ? (
            <Empty title="Tidak ada PO yang cocok" sub="Coba kata kunci atau status lain." />
          ) : (
          <>
          <TableWrap className="hidden min-[1280px]:block">
          <table className="w-full border-collapse">
            <thead><tr><Th>Nomor PO</Th><Th>Supplier</Th><Th>Status</Th><Th>Tanggal</Th><Th right>Aksi</Th></tr></thead>
            {orders.loading && orderItems.length === 0 ? (
              <SkeletonRows cols={5} />
            ) : (
              <tbody>
                {pageItems.map((order) => (
                  <tr key={order.id}>
                    <Td mono>{order.number}</Td>
                    <Td>{order.supplier_name || order.supplier?.name || knownSupplierNames[`${who}:${order.id}`] || supplierName(order.supplier_id)}</Td>
                    <Td>{orderStatus(order.status)}</Td>
                    <Td>{fmtDate(order.created_at)}</Td>
                    <Td><RowAction title="Detail" onClick={() => openDetail(order)} ><Eye className="size-4" /></RowAction></Td>
                  </tr>
                ))}
              </tbody>
            )}
          </table>
        </TableWrap>
        <CardGrid>
          {orders.loading && orderItems.length === 0 ? (
            [0, 1, 2].map((i) => <Skeleton key={i} className="h-[120px] w-full rounded-xl" />)
          ) : pageItems.map((order) => (
            <DataCard key={order.id}>
              <div className="flex items-center justify-between gap-2">
                <p className="min-w-0 truncate font-mono text-[13px] font-medium text-fg">{order.number}</p>
                {orderStatus(order.status)}
              </div>
              {order.total != null && <p className="mt-1.5 text-right text-[17px] font-semibold tabular-nums text-fg">{fmtRp(order.total)}</p>}
              <p className="mt-1 text-xs leading-snug text-fog">
                {order.supplier_name || order.supplier?.name || knownSupplierNames[`${who}:${order.id}`] || supplierName(order.supplier_id)} · {fmtDate(order.created_at)}
              </p>
              <div className="mt-auto flex items-center gap-1 pt-1.5">
                <RowAction title="Detail" label={`Detail ${order.number}`} onClick={() => openDetail(order)}><Eye className="size-4" /></RowAction>
              </div>
            </DataCard>
          ))}
        </CardGrid>
          </>
          )}
        <Pager page={safePage} total={pages} onChange={setPage} totalItems={filteredOrders.length} pageSize={pageSize} onPageSizeChange={(n) => { setPageSize(n); setPage(0) }} />
        </>
      )}
      </div>

      <Modal open={!!editing} title={editing?.id ? 'Ubah PO' : 'Buat PO'} onClose={() => { if (!formBusy) setEditing(null) }} wide>
        {editing && (
          <PurchaseOrderForm
            key={editing.id ?? 'new'}
            draft={editing}
            suppliers={supplierItems}
            products={productItems}
            supplierLoading={suppliers.loading}
            productLoading={products.loading}
            dataError={suppliers.err || products.err}
            busy={formBusy}
            error={formError}
            onSave={saveOrder}
            onCancel={() => setEditing(null)}
          />
        )}
      </Modal>

      <Modal open={selectedId !== null && !editing} title={detail?.number ?? 'Detail PO'} onClose={closeDetail} wide>
        {detailLoading ? <p className="py-8 text-center text-sm text-muted" role="status">Memuat detail PO…</p> : detailError ? (
          <AlertStrip>{detailError}</AlertStrip>
        ) : detail && (
          <div className="space-y-4">
            <div className="flex flex-wrap items-start justify-between gap-3 rounded-xl bg-surface px-4 py-3">
              <div>
                <p className="font-medium text-fg">{detail.supplier_name || detail.supplier?.name || supplierName(detail.supplier_id)}</p>
                <p className="mt-1 text-[13px] text-muted">{fmtDate(detail.created_at)}</p>
              </div>
              {orderStatus(detail.status)}
            </div>

            {actionError && <AlertStrip>{actionError}</AlertStrip>}

            <div className="divide-y divide-dove rounded-lg border border-dove px-3">
              {detail.items.map((item) => (
                <div key={item.id} className="flex items-center justify-between gap-3 py-2.5">
                  <p className="min-w-0">
                    <span className="font-medium text-fg">{purchaseOrderItemName(item, productItems)}</span><br />
                    <span className="text-xs text-fog">Pesan {item.qty_ordered.toLocaleString('id-ID')} · Diterima {item.qty_received.toLocaleString('id-ID')} · {fmtRp(item.buy_price)}</span>
                  </p>
                  <p className="shrink-0 font-medium tabular-nums text-fg">{fmtRp(item.qty_ordered * item.buy_price)}</p>
                </div>
              ))}
              <div className="flex items-center justify-between gap-3 py-2.5">
                <p className="font-medium text-fg">Total</p>
                <p className="shrink-0 font-semibold tabular-nums text-fg">{fmtRp(detailTotal)}</p>
              </div>
            </div>

            {canReceive && (
              <form className="space-y-3 rounded-xl bg-surface p-4" onSubmit={(event) => { event.preventDefault(); void receiveItems() }}>
                <div>
                  <h4 className="text-sm font-medium text-fg">Penerimaan</h4>
                  <p className="mt-1 text-[13px] text-muted">Isi jumlah yang diterima sekarang untuk item yang datang.</p>
                </div>
                {detail.items.map((item) => {
                  const remaining = remainingToReceive(item)
                  return (
                    <div key={item.id} className="grid items-end gap-3 border-t border-dove pt-3 sm:grid-cols-[minmax(0,1fr)_auto_8rem]">
                      <p className="text-[13px] font-medium text-fg">{purchaseOrderItemName(item, productItems)}<span className="ml-2 font-normal text-muted">Sisa {remaining.toLocaleString('id-ID')}</span></p>
                      <span className="text-[12px] text-muted">Dipesan {item.qty_ordered.toLocaleString('id-ID')}</span>
                      <NumInput label="Qty terima" value={receiveQty[item.id] ?? '0'} onValue={(value) => setReceiveQty((current) => ({ ...current, [item.id]: value }))} />
                    </div>
                  )
                })}
                {receiveError && <AlertStrip>{receiveError}</AlertStrip>}
                <ModalActions>
                  <Button type="submit" disabled={actionBusy}>{actionBusy ? 'Menyimpan…' : 'Terima'}</Button>
                </ModalActions>
              </form>
            )}

            <ModalActions>
              {detail.status === 'draft' && <Button variant="ghost" onClick={() => openEdit(detail)} disabled={actionBusy}>Ubah</Button>}
              {detail.status === 'draft' && <Button variant="danger" onClick={() => { setConfirmError(''); setConfirmation({ action: 'delete', order: detail }) }} disabled={actionBusy}>Hapus</Button>}
              {detail.status === 'draft' && <Button onClick={() => void orderNow()} disabled={actionBusy}>{actionBusy ? 'Memproses…' : 'Pesan'}</Button>}
              {(detail.status === 'draft' || detail.status === 'ordered') && <Button variant="danger" onClick={() => { setConfirmError(''); setConfirmation({ action: 'cancel', order: detail }) }} disabled={actionBusy}>Batalkan</Button>}
            </ModalActions>
          </div>
        )}
      </Modal>

      <Modal open={!!confirmation} title={confirmation?.action === 'delete' ? 'Hapus PO' : 'Batalkan PO'} onClose={() => { if (!actionBusy) setConfirmation(null) }}>
        {confirmation && (
          <div className="space-y-4">
            <p className="text-sm text-muted">
              {confirmation.action === 'delete'
                ? <>Hapus PO <span className="font-medium text-fg">{confirmation.order.number}</span>?</>
                : <>Batalkan PO <span className="font-medium text-fg">{confirmation.order.number}</span>? Stok tidak berubah.</>}
            </p>
            {confirmError && <AlertStrip>{confirmError}</AlertStrip>}
            <ModalActions>
              <Button variant="ghost" onClick={() => setConfirmation(null)} disabled={actionBusy}>Kembali</Button>
              <Button variant="danger" onClick={() => void confirmOrderAction()} disabled={actionBusy}>{actionBusy ? 'Memproses…' : confirmation.action === 'delete' ? 'Hapus' : 'Batalkan PO'}</Button>
            </ModalActions>
          </div>
        )}
      </Modal>
    </>
  )
}

function PurchaseOrderForm({ draft, suppliers, products, supplierLoading, productLoading, dataError, busy, error, onSave, onCancel }: {
  draft: Draft
  suppliers: Supplier[]
  products: Product[]
  supplierLoading: boolean
  productLoading: boolean
  dataError: string
  busy: boolean
  error: string
  onSave: (draft: Draft) => void
  onCancel: () => void
}) {
  const [value, setValue] = useState(draft)
  const setItem = (key: number, changes: Partial<DraftItem>) => setValue((current) => ({
    ...current,
    items: current.items.map((item) => item.key === key ? { ...item, ...changes } : item),
  }))
  const total = purchaseOrderTotal(value.items)

  function addItem() {
    const key = Math.max(0, ...value.items.map((item) => item.key)) + 1
    setValue((current) => ({ ...current, items: [...current.items, emptyItem(key)] }))
  }

  function changeProduct(item: DraftItem, productId: string) {
    const selectedId = purchaseOrderProductId(productId)
    const product = products.find((entry) => purchaseOrderProductId(entry.id) === selectedId)
    setItem(item.key, {
      productId,
      variantId: '',
      variantName: undefined,
      buyPrice: product?.variants?.length ? '' : product ? String(product.buy_price) : '',
    })
  }

  return (
    <form className="space-y-4" onSubmit={(event) => { event.preventDefault(); onSave(value) }}>
      {error && <AlertStrip>{error}</AlertStrip>}
      {dataError && <AlertStrip>{dataError}</AlertStrip>}
      <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
        Supplier *
        <Select label="Supplier" value={value.supplierId} onChange={(v) => setValue((current) => ({ ...current, supplierId: v }))} options={[{ value: '', label: supplierLoading ? 'Memuat supplier…' : 'Pilih supplier' }, ...suppliers.map((supplier) => ({ value: String(supplier.id), label: supplier.name }))]} />
      </label>

      <div className="space-y-3">
        <div className="flex items-center justify-between gap-3">
          <h4 className="text-sm font-medium text-fg">Item produk</h4>
          <Button type="button" variant="ghost" className="px-4 py-2 text-[13px]" onClick={addItem}><Plus className="size-4" />Tambah item</Button>
        </div>
        {value.items.map((item, index) => {
          const selectedId = purchaseOrderProductId(item.productId)
          const product = products.find((entry) => purchaseOrderProductId(entry.id) === selectedId)
          const variants = product?.variants ?? []
          const hasVariants = variants.length > 0 || !!item.variantId
          return (
            <div key={item.key} className="space-y-3 rounded-xl bg-surface p-3 sm:p-4">
              <div className="flex items-center justify-between gap-3">
                <p className="text-[13px] font-medium text-steel">Item {index + 1}</p>
                <button type="button" onClick={() => setValue((current) => ({ ...current, items: current.items.filter((entry) => entry.key !== item.key) }))} aria-label={`Hapus item ${index + 1}`} className="inline-flex items-center gap-1 rounded-md px-2 py-1 text-[12px] text-ember hover:bg-sand">
                  <Trash2 className="size-3.5" />Hapus
                </button>
              </div>
              <div className="grid gap-3 sm:grid-cols-2">
                <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
                  Produk *
                  <Select label={`Produk item ${index + 1}`} value={item.productId} onChange={(v) => changeProduct(item, v)} options={[{ value: '', label: productLoading ? 'Memuat produk…' : 'Pilih produk' }, ...products.map((entry) => ({ value: String(entry.id), label: entry.name })), ...(item.productId && !products.some((entry) => purchaseOrderProductId(entry.id) === purchaseOrderProductId(item.productId)) ? [{ value: item.productId, label: item.productId }] : [])]} />
                </label>
                {hasVariants && (
                  <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
                    Varian *
                    <Select label={`Varian item ${index + 1}`} value={item.variantId} onChange={(v) => { const variant = variants.find((entry) => String(entry.id) === v); setItem(item.key, { variantId: v, variantName: variant?.name, buyPrice: variant ? String(variant.buy_price) : item.buyPrice }) }} options={[{ value: '', label: 'Pilih varian' }, ...variants.map((variant) => ({ value: String(variant.id), label: variant.name })), ...(item.variantId && !variants.some((variant) => String(variant.id) === item.variantId) ? [{ value: item.variantId, label: item.variantName || `Varian ${item.variantId}` }] : [])]} />
                  </label>
                )}
                <NumInput label="Qty dipesan *" value={item.qty} onValue={(qty) => setItem(item.key, { qty })} required />
                <NumInput label="Harga beli (Rp) *" value={item.buyPrice} onValue={(buyPrice) => setItem(item.key, { buyPrice })} required />
              </div>
            </div>
          )
        })}
      </div>

      <div className="flex items-center justify-between rounded-xl bg-surface px-4 py-3">
        <span className="text-sm font-medium text-steel">Total berjalan</span>
        <span className="font-mono text-sm font-medium tabular-nums text-fg">{fmtRp(total)}</span>
      </div>

      <ModalActions>
        <Button type="button" variant="ghost" onClick={onCancel} disabled={busy}>Batal</Button>
        <Button type="submit" disabled={busy}>{busy ? 'Menyimpan…' : 'Simpan Draf'}</Button>
      </ModalActions>
    </form>
  )
}
