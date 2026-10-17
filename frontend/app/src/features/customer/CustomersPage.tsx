import { useEffect, useMemo, useState } from 'react'
import { Pencil, Plus, ReceiptText, Trash2 } from 'lucide-react'
import {
  apiCustomerTransactions, apiGetCustomer,
  type Customer, type CustomerInput, type CustomerTransaction,
} from './api'
import { useCreateCustomer, useCustomersList, useDeleteCustomer, useReceivables, useUpdateCustomer } from './hooks'
import { PageHeader } from '../../shared/PageHeader'
import { fmtDate, fmtRp, useDB } from '../../shared/session/store'
import { TextArea, ModalActions, Pager, AlertStrip, Button, CardGrid, DataCard, Empty, Input, Modal, Pill, RowAction, SearchInput, Select, SkeletonRows, StatusPill, TableWrap, Td, Th } from '../../shared/ui'
import { Skeleton } from '@/shared/ui/primitives/skeleton'

type Draft = { id?: number; name: string; phone: string; email: string; address: string }
type History = {
  customer: Customer
  detail: Customer | null
  transactions: CustomerTransaction[]
  loading: boolean
  error: string
}

export default function Customers() {
  const { session } = useDB()
  const isAdmin = session?.role === 'admin'
  const canCreate = isAdmin || session?.role === 'cashier'
  const [query, setQuery] = useState('')
  const [custFilter, setCustFilter] = useState('')
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const receivablesQ = useReceivables()
  const receivables = receivablesQ.data ?? []
  const owedCustomerIds = useMemo(() => new Set(receivables.filter((r) => r.remaining > 0).map((r) => r.customer_id)), [receivables])
  const list = useCustomersList(query.trim(), page, pageSize)
  const customers = list.data?.items ?? []
  const total = list.data?.total ?? 0
  const loading = list.loading
  const error = list.err
  const createCustomer = useCreateCustomer()
  const updateCustomer = useUpdateCustomer()
  const deleteCustomer = useDeleteCustomer()
  const [editing, setEditing] = useState<Draft | null>(null)
  const [formError, setFormError] = useState('')
  const [saveBusy, setSaveBusy] = useState(false)
  const [deleteFor, setDeleteFor] = useState<Customer | null>(null)
  const [deleteError, setDeleteError] = useState('')
  const [deleteBusy, setDeleteBusy] = useState(false)
  const [history, setHistory] = useState<History | null>(null)

  useEffect(() => { setPage(0) }, [query])

  function reload() {
    setPage(0)
    list.reload()
  }

  function openCreate() {
    setFormError('')
    setEditing({ name: '', phone: '', email: '', address: '' })
  }

  function openEdit(customer: Customer) {
    setFormError('')
    setEditing({ id: customer.id, name: customer.name, phone: customer.phone ?? '', email: customer.email ?? '', address: customer.address ?? '' })
  }

  async function saveCustomer() {
    if (!editing) return
    const name = editing.name.trim()
    if (!name) return setFormError('Nama pelanggan wajib diisi.')
    const body: CustomerInput = {
      name,
      ...(editing.phone.trim() ? { phone: editing.phone.trim() } : {}),
      email: editing.email.trim(),
      address: editing.address.trim(),
    }
    setFormError('')
    setSaveBusy(true)
    try {
      if (editing.id) await updateCustomer.mutateAsync({ id: editing.id, body })
      else await createCustomer.mutateAsync(body)
      setEditing(null)
      reload()
    } catch (e) {
      setFormError(e instanceof Error ? e.message : 'Gagal menyimpan pelanggan.')
    } finally {
      setSaveBusy(false)
    }
  }

  async function removeCustomer() {
    if (!deleteFor) return
    setDeleteError('')
    setDeleteBusy(true)
    try {
      await deleteCustomer.mutateAsync(deleteFor.id)
      setDeleteFor(null)
      reload()
    } catch (e) {
      setDeleteError(e instanceof Error ? e.message : 'Gagal menghapus pelanggan.')
    } finally {
      setDeleteBusy(false)
    }
  }

  function openHistory(customer: Customer) {
    setHistory({ customer, detail: null, transactions: [], loading: true, error: '' })
    Promise.all([apiGetCustomer(customer.id), apiCustomerTransactions(customer.id)])
      .then(([detail, transactions]) => {
        setHistory((current) => current?.customer.id === customer.id
          ? { customer, detail, transactions, loading: false, error: '' }
          : current)
      })
      .catch((e) => {
        setHistory((current) => current?.customer.id === customer.id
          ? { ...current, loading: false, error: e instanceof Error ? e.message : 'Gagal memuat riwayat transaksi.' }
          : current)
      })
  }

  const displayedPoints = (customer: Customer) => (history?.detail?.id === customer.id ? history.detail.points : customer.points)
  const displayedEmail = history?.detail?.email ?? history?.customer.email ?? ''

  const shownCustomers = customers.filter((c) => custFilter === 'poin' ? c.points > 0 : custFilter === 'utang' ? owedCustomerIds.has(c.id) : true)

  return (
    <>
      <PageHeader
        title="Pelanggan"
       
        crumb="Pelanggan"
        actions={canCreate ? <Button onClick={openCreate}><Plus className="size-4" />Tambah Pelanggan</Button> : undefined}
      />

      {error && <AlertStrip className="mb-4">{error}</AlertStrip>}

      <div className="mb-4 flex flex-col gap-3 sm:flex-row">
        <SearchInput value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Ketik nama atau nomor telepon…" aria-label="Cari nama atau telepon" />
        <Select
          label="Filter pelanggan"
          value={custFilter}
          onChange={setCustFilter}
          className="shrink-0 sm:w-56"
          options={[{ value: '', label: 'Semua pelanggan' }, { value: 'poin', label: 'Punya poin' }, { value: 'utang', label: 'Punya piutang' }]}
        />
      </div>

      {customers.length === 0 && error ? null : customers.length === 0 && !loading ? (
        <Empty
          title={query.trim() ? 'Pelanggan tidak ditemukan' : 'Belum ada pelanggan'}
          sub={query.trim() ? 'Coba nama atau nomor telepon lain.' : 'Pelanggan akan muncul setelah ditambahkan.'}
        />
      ) : (
        <>
          {custFilter && shownCustomers.length === 0 ? (
            <Empty title="Tidak ada pelanggan yang cocok" sub="Coba filter atau kata kunci lain." />
          ) : (
          <>
          <TableWrap className="hidden min-[1280px]:block">
            <table className="w-full border-collapse">
              <thead><tr><Th>Nama</Th><Th>Telepon</Th><Th>Alamat</Th><Th right>Poin</Th><Th right>Aksi</Th></tr></thead>
              {loading && customers.length === 0 ? (
                <SkeletonRows cols={5} />
              ) : (
                <tbody>
                  {shownCustomers.map((customer) => (
                    <tr key={customer.id}>
                      <Td><span className="font-medium text-fg">{customer.name}</span></Td>
                      <Td>{customer.phone || '—'}</Td>
                      <Td><span className="block max-w-xs truncate" title={customer.address}>{customer.address || '—'}</span></Td>
                      <Td right><span className="font-mono tabular-nums text-fg">{customer.points.toLocaleString('id-ID')}</span></Td>
                      <Td>
                        <div className="flex gap-1">
                          <RowAction title="Riwayat transaksi" label={`Riwayat transaksi ${customer.name}`} onClick={() => openHistory(customer)}><ReceiptText className="size-4" /></RowAction>
                          {isAdmin && (
                            <>
                              <RowAction title="Ubah" label={`Ubah ${customer.name}`} onClick={() => openEdit(customer)}><Pencil className="size-4" /></RowAction>
                              <RowAction title="Hapus" label={`Hapus ${customer.name}`} onClick={() => { setDeleteError(''); setDeleteFor(customer) }} danger><Trash2 className="size-4" /></RowAction>
                            </>
                          )}
                        </div>
                      </Td>
                    </tr>
                  ))}
                </tbody>
              )}
            </table>
          </TableWrap>
          <CardGrid className="mt-4">
            {loading && customers.length === 0 ? (
              [0, 1, 2].map((i) => <Skeleton key={i} className="h-[128px] w-full rounded-xl" />)
            ) : shownCustomers.map((customer) => (
              <DataCard key={customer.id}>
                <div className="flex items-center justify-between gap-2">
                  <p className="min-w-0 truncate font-medium text-fg">{customer.name}</p>
                  <p className="shrink-0 font-mono text-[15px] font-semibold tabular-nums text-fg">{customer.points.toLocaleString('id-ID')} <span className="font-sans text-xs font-normal text-fog">poin</span></p>
                </div>
                <p className="mt-1 text-xs leading-snug text-fog">{customer.phone || 'Tanpa telepon'}{customer.address ? ` · ${customer.address}` : ''}</p>
                <div className="mt-auto flex items-center gap-1 pt-1.5">
                  <RowAction title="Riwayat transaksi" label={`Riwayat transaksi ${customer.name}`} onClick={() => openHistory(customer)}><ReceiptText className="size-4" /></RowAction>
                  {isAdmin && (
                    <>
                      <RowAction title="Ubah" label={`Ubah ${customer.name}`} onClick={() => openEdit(customer)}><Pencil className="size-4" /></RowAction>
                      <RowAction title="Hapus" label={`Hapus ${customer.name}`} onClick={() => { setDeleteError(''); setDeleteFor(customer) }} danger><Trash2 className="size-4" /></RowAction>
                    </>
                  )}
                </div>
              </DataCard>
            ))}
          </CardGrid>
          </>
          )}
          <Pager page={page} total={Math.max(1, Math.ceil(total / pageSize))} onChange={setPage} totalItems={total} pageSize={pageSize} onPageSizeChange={(n) => { setPageSize(n); setPage(0) }} />
        </>
      )}

      <Modal open={!!editing} title={editing?.id ? 'Ubah Pelanggan' : 'Tambah Pelanggan'} onClose={() => { if (!saveBusy) setEditing(null) }}>
        {editing && (
          <div className="space-y-4">
            {formError && <AlertStrip>{formError}</AlertStrip>}
            <Input label="Nama *" value={editing.name} required onChange={(name) => setEditing((current) => current ? { ...current, name } : current)} placeholder="Nama pelanggan" />
            <Input label="Telepon" value={editing.phone} onChange={(phone) => setEditing((current) => current ? { ...current, phone } : current)} placeholder="08…" />
            <Input label="Email" type="email" value={editing.email} onChange={(email) => setEditing((current) => current ? { ...current, email } : current)} placeholder="nama@contoh.com" />
            <TextArea label="Alamat" value={editing.address} onChange={(v) => setEditing((current) => current ? { ...current, address: v } : current)} rows={3} placeholder="Alamat pelanggan" />
            <ModalActions>
              <Button variant="ghost" onClick={() => setEditing(null)} disabled={saveBusy}>Batal</Button>
              <Button onClick={saveCustomer} disabled={saveBusy}>{saveBusy ? 'Menyimpan…' : 'Simpan'}</Button>
            </ModalActions>
          </div>
        )}
      </Modal>

      <Modal open={!!deleteFor} title="Hapus Pelanggan" onClose={() => { if (!deleteBusy) setDeleteFor(null) }}>
        {deleteFor && (
          <div className="space-y-4">
            <p className="text-sm text-muted">Hapus pelanggan <span className="font-medium text-fg">{deleteFor.name}</span> dari daftar?</p>
            {deleteError && <AlertStrip>{deleteError}</AlertStrip>}
            <ModalActions>
              <Button variant="ghost" onClick={() => setDeleteFor(null)} disabled={deleteBusy}>Batal</Button>
              <Button variant="danger" onClick={removeCustomer} disabled={deleteBusy}>{deleteBusy ? 'Menghapus…' : 'Hapus'}</Button>
            </ModalActions>
          </div>
        )}
      </Modal>

      <Modal open={!!history} title="Riwayat Transaksi Pelanggan" onClose={() => setHistory(null)} wide>
        {history && (
          <div className="space-y-4">
            <div className="flex flex-wrap items-center justify-between gap-3 rounded-xl bg-surface px-4 py-3">
              <div>
                <p className="font-medium text-fg">{history.customer.name}</p>
                <p className="text-[13px] text-muted">{history.customer.phone || 'Tanpa nomor telepon'}</p>
                {displayedEmail && <p className="text-[13px] text-muted">{displayedEmail}</p>}
              </div>
              <Pill tone="muted">{displayedPoints(history.customer).toLocaleString('id-ID')} poin</Pill>
            </div>
            {history.error && <AlertStrip>{history.error}</AlertStrip>}
            {history.loading ? (
              <p className="py-8 text-center text-sm text-muted" role="status">Memuat riwayat transaksi…</p>
            ) : history.transactions.length === 0 ? (
              history.error ? null : <Empty title="Belum ada transaksi" sub="Riwayat transaksi pelanggan ini masih kosong." />
            ) : (
              <div className="divide-y divide-dove rounded-lg border border-dove px-3">
                {history.transactions.map((transaction) => (
                  <div key={transaction.id} className="flex items-center justify-between gap-3 py-2.5">
                    <p className="min-w-0">
                      <span className="font-mono text-[13px] font-medium text-fg">#{transaction.id}</span>{' '}
                      <StatusPill status={transaction.status} /><br />
                      <span className="text-xs text-fog">{fmtDate(transaction.created_at)} · {transaction.method}</span>
                    </p>
                    <p className="shrink-0 font-mono text-sm font-medium tabular-nums text-fg">{fmtRp(transaction.total)}</p>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
      </Modal>
    </>
  )
}
