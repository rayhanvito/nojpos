import { useEffect, useState } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { type Product, type StockTransferInput } from './api'
import { useCreateTransfer, useStockProducts, useTransfers } from './hooks'
import { useOutlets } from '../../hooks/use-outlets'
import { useDB, fmtDate, fmtTime } from '../../shared/session/store'
import { AlertStrip, Button, DataCard, Empty, Modal, ModalActions, NumInput, Pager, Select } from '../../shared/ui'

type DraftItem = { key: number; productId: string; variantId: string; qty: string }

function emptyItem(key: number): DraftItem {
  return { key, productId: '', variantId: '', qty: '' }
}

function wireId(id: string): string | number {
  const numeric = Number(id)
  return /^\d+$/.test(id) && Number.isSafeInteger(numeric) ? numeric : id
}

function selectedProduct(products: Product[], id: string) {
  return products.find((product) => product.id === id)
}

export default function TransferTab() {
  const { session } = useDB()
  const { outlets, loading: outletsLoading, error: outletsError } = useOutlets()
  const productsQuery = useStockProducts()
  const products = (productsQuery.data ?? []).filter((product) => product.active)
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const transfers = useTransfers(page, pageSize)
  const createTransfer = useCreateTransfer()
  const [open, setOpen] = useState(false)
  const [fromOutletId, setFromOutletId] = useState('')
  const [toOutletId, setToOutletId] = useState('')
  const [items, setItems] = useState<DraftItem[]>([emptyItem(1)])
  const [note, setNote] = useState('')
  const [formError, setFormError] = useState('')
  const [message, setMessage] = useState('')
  const [busy, setBusy] = useState(false)
  const total = transfers.data?.total ?? 0
  const pages = Math.max(1, Math.ceil(total / pageSize))
  const safePage = Math.min(page, pages - 1)
  const outletNames = new Map(outlets.map((outlet) => [outlet.id, outlet.name]))

  useEffect(() => {
    if (transfers.data && page > pages - 1) setPage(pages - 1)
  }, [transfers.data, page, pages])

  function openForm() {
    const source = outlets.find((outlet) => outlet.id === session?.outletId)?.id ?? outlets[0]?.id ?? ''
    setFromOutletId(source)
    setToOutletId(outlets.find((outlet) => outlet.id !== source)?.id ?? '')
    setItems([emptyItem(1)])
    setNote('')
    setFormError('')
    setOpen(true)
  }

  function changeItem(key: number, patch: Partial<DraftItem>) {
    setItems((current) => current.map((item) => item.key === key ? { ...item, ...patch } : item))
  }

  function addItem() {
    const key = Math.max(0, ...items.map((item) => item.key)) + 1
    setItems((current) => [...current, emptyItem(key)])
  }

  async function submit() {
    if (!fromOutletId || !toOutletId) return setFormError('Pilih outlet asal dan tujuan.')
    if (fromOutletId === toOutletId) return setFormError('Outlet asal dan tujuan harus berbeda.')
    if (items.length === 0) return setFormError('Tambahkan minimal satu item produk.')

    const transferItems: StockTransferInput['items'] = []
    for (const item of items) {
      const product = selectedProduct(products, item.productId)
      const qty = Number(item.qty)
      if (!product || !Number.isSafeInteger(qty) || qty <= 0) return setFormError('Pilih produk dan isi qty transfer dengan bilangan bulat lebih dari 0.')
      if (product.variants?.length && !item.variantId) return setFormError(`Pilih varian untuk produk ${product.name}.`)
      transferItems.push({ productId: wireId(item.productId), variantId: item.variantId ? Number(item.variantId) : null, qty })
    }

    const body: StockTransferInput = {
      fromOutletId: wireId(fromOutletId),
      toOutletId: wireId(toOutletId),
      items: transferItems,
      note: note.trim(),
    }
    setFormError('')
    setBusy(true)
    try {
      await createTransfer.mutateAsync(body)
      setOpen(false)
      setPage(0)
      setMessage('Transfer stok berhasil dibuat.')
    } catch (e) {
      setFormError(e instanceof Error ? e.message : 'Gagal membuat transfer stok.')
    } finally {
      setBusy(false)
    }
  }

  const transferList = transfers.data?.items ?? []

  return (
    <section className="mt-5">
      {(outletsError || productsQuery.err || transfers.err) && <AlertStrip className="mb-4">{outletsError || productsQuery.err || transfers.err}</AlertStrip>}
      <div className="mb-4 flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-[15px] font-medium text-fg">Transfer Stok</h2>
          <p className="mt-0.5 text-[13px] text-muted">Pindahkan stok antar-outlet.</p>
        </div>
        <Button onClick={openForm} disabled={busy || outlets.length < 2 || outletsLoading}>Transfer Baru</Button>
      </div>
      {message && <p className="mb-4 rounded-lg bg-surface px-3.5 py-2.5 text-[13px] text-sprout" role="status">{message}</p>}
      {outlets.length < 2 && !outletsLoading && <p className="mb-4 text-sm text-muted">Transfer membutuhkan sedikitnya dua outlet.</p>}

      {transfers.loading ? (
        <p className="rounded-2xl bg-cream p-5 text-sm text-muted">Memuat transfer…</p>
      ) : transferList.length === 0 ? (
        <div className="rounded-2xl bg-cream p-4 sm:p-6"><Empty title="Belum ada transfer stok" sub="Transfer antar-outlet yang dibuat akan muncul di sini." /></div>
      ) : (
        <div className="grid grid-cols-1 gap-2.5 min-[768px]:grid-cols-2">
          {transferList.map((transfer) => (
            <DataCard key={transfer.id}>
              <div className="flex flex-wrap items-start justify-between gap-2">
                <p className="min-w-0 flex-1 break-words text-[15px] font-medium text-fg">
                  {transfer.fromOutletName || outletNames.get(transfer.fromOutletId) || transfer.fromOutletId}
                  <span className="mx-2 text-fog" aria-hidden="true">→</span>
                  {transfer.toOutletName || outletNames.get(transfer.toOutletId) || transfer.toOutletId}
                </p>
                <span className="shrink-0 rounded-full bg-surface px-2.5 py-1 text-xs text-muted">{transfer.items.length} item</span>
              </div>
              <p className="mt-1 text-xs text-fog">{transfer.createdBy || '—'} · {fmtDate(transfer.createdAt)} {fmtTime(transfer.createdAt)}</p>
              {transfer.note && <p className="mt-2 break-words text-[13px] text-muted">{transfer.note}</p>}
            </DataCard>
          ))}
        </div>
      )}
      <Pager page={safePage} total={pages} onChange={setPage} totalItems={total} pageSize={pageSize} onPageSizeChange={(size) => { setPageSize(size); setPage(0) }} />

      <Modal open={open} title="Transfer Baru" onClose={() => { if (!busy) setOpen(false) }} wide>
        <div className="space-y-4">
          {formError && <AlertStrip>{formError}</AlertStrip>}
          <div className="grid gap-3 sm:grid-cols-2">
            <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
              Outlet asal *
              <Select label="Outlet asal" value={fromOutletId} onChange={(value) => {
                setFromOutletId(value)
                if (toOutletId === value) setToOutletId(outlets.find((outlet) => outlet.id !== value)?.id ?? '')
              }} options={[{ value: '', label: 'Pilih outlet asal' }, ...outlets.map((outlet) => ({ value: outlet.id, label: outlet.name }))]} />
            </label>
            <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
              Outlet tujuan *
              <Select label="Outlet tujuan" value={toOutletId} onChange={setToOutletId} options={[{ value: '', label: 'Pilih outlet tujuan' }, ...outlets.filter((outlet) => outlet.id !== fromOutletId).map((outlet) => ({ value: outlet.id, label: outlet.name }))]} />
            </label>
          </div>

          <div className="space-y-3">
            <div className="flex items-center justify-between gap-3">
              <h3 className="text-sm font-medium text-fg">Item produk</h3>
              <Button variant="ghost" className="h-[42px] px-3" onClick={addItem}><Plus className="size-4" />Tambah item</Button>
            </div>
            {items.map((item, index) => {
              const product = selectedProduct(products, item.productId)
              const variants = product?.variants ?? []
              return (
                <div key={item.key} className="space-y-3 rounded-xl bg-surface p-3 sm:p-4">
                  <div className="flex items-center justify-between gap-2">
                    <p className="text-[13px] font-medium text-steel">Item {index + 1}</p>
                    <Button type="button" variant="ghost" className="h-[42px] px-2 text-xs" disabled={items.length === 1} onClick={() => setItems((current) => current.filter((row) => row.key !== item.key))}>
                      <Trash2 className="size-3.5" />Hapus
                    </Button>
                  </div>
                  <div className="grid gap-3 sm:grid-cols-2">
                    <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
                      Produk *
                      <Select label={`Produk item ${index + 1}`} value={item.productId} onChange={(productId) => changeItem(item.key, { productId, variantId: '' })} options={[{ value: '', label: productsQuery.loading ? 'Memuat produk…' : 'Pilih produk' }, ...products.map((entry) => ({ value: entry.id, label: entry.name }))]} />
                    </label>
                    {variants.length > 0 && (
                      <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
                        Varian *
                        <Select label={`Varian item ${index + 1}`} value={item.variantId} onChange={(variantId) => changeItem(item.key, { variantId })} options={[{ value: '', label: 'Pilih varian' }, ...variants.map((variant) => ({ value: String(variant.id), label: variant.name }))]} />
                      </label>
                    )}
                    <NumInput label="Qty transfer *" value={item.qty} onValue={(qty) => changeItem(item.key, { qty })} />
                  </div>
                </div>
              )
            })}
          </div>

          <label className="flex flex-col gap-1.5 text-[13px] font-medium text-steel">
            Catatan (opsional)
            <textarea value={note} onChange={(event) => setNote(event.target.value)} rows={2} className="min-h-[42px] w-full rounded-md border border-border bg-paper px-3.5 py-2 text-[15px] text-fg focus:border-jet focus:outline-2 focus:outline-accent-soft" />
          </label>
          <ModalActions>
            <Button type="button" variant="ghost" disabled={busy} onClick={() => setOpen(false)}>Batal</Button>
            <Button onClick={() => void submit()} disabled={busy || productsQuery.loading}>{busy ? 'Mentransfer…' : 'Buat Transfer'}</Button>
          </ModalActions>
        </div>
      </Modal>
    </section>
  )
}
