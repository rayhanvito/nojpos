import { useState } from 'react'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import { apiCreateSupplier, apiDeleteSupplier, apiListPayables, apiListSuppliers, apiUpdateSupplier, useFetch, type Supplier, type SupplierInput } from '../lib/api'
import { PageHeader } from '../lib/PageHeader'
import { useDB } from '../lib/store'
import { TextArea, ModalActions, AlertStrip, Button, CardGrid, DataCard, Empty, Input, Modal, Pager, Pill, RowAction, SearchInput, Select, SkeletonRows, TableWrap, Td, Th } from '../lib/ui'
import { Skeleton } from '@/components/ui/skeleton'

type Draft = { id?: number; name: string; contact: string; address: string }


export default function SupplierPage() {
  const { session } = useDB()
  const suppliers = useFetch(apiListSuppliers, [session?.id])
  const payables = useFetch(apiListPayables, [session?.id])
  const [supFilter, setSupFilter] = useState('')
  const [editing, setEditing] = useState<Draft | null>(null)
  const [deleting, setDeleting] = useState<Supplier | null>(null)
  const [formError, setFormError] = useState('')
  const [deleteError, setDeleteError] = useState('')
  const [busy, setBusy] = useState(false)

  function openCreate() {
    setFormError('')
    setEditing({ name: '', contact: '', address: '' })
  }

  function openEdit(supplier: Supplier) {
    setFormError('')
    setEditing({ id: supplier.id, name: supplier.name, contact: supplier.contact, address: supplier.address })
  }

  async function saveSupplier() {
    if (!editing) return
    const name = editing.name.trim()
    if (!name) return setFormError('Nama supplier wajib diisi.')
    const body: SupplierInput = { name, contact: editing.contact.trim(), address: editing.address.trim() }
    setFormError('')
    setBusy(true)
    try {
      if (editing.id) await apiUpdateSupplier(editing.id, body)
      else await apiCreateSupplier(body)
      setEditing(null)
      suppliers.reload()
    } catch (e) {
      setFormError(e instanceof Error ? e.message : 'Gagal menyimpan supplier.')
    } finally {
      setBusy(false)
    }
  }

  async function removeSupplier() {
    if (!deleting) return
    setDeleteError('')
    setBusy(true)
    try {
      await apiDeleteSupplier(deleting.id)
      setDeleting(null)
      suppliers.reload()
    } catch (e) {
      setDeleteError(e instanceof Error ? e.message : 'Gagal menghapus supplier.')
    } finally {
      setBusy(false)
    }
  }

  const items = suppliers.data ?? []
  const [q, setQ] = useState('')
  const [page, setPage] = useState(0)
  const [pageSize, setPageSize] = useState(10)
  const owedSupplierIds = new Set((payables.data ?? []).filter((p) => p.remaining > 0).map((p) => p.supplier_id))
  const filteredItems = items.filter((supplier) => {
    if (supFilter === 'utang' && !owedSupplierIds.has(supplier.id)) return false
    const term = q.trim().toLocaleLowerCase('id-ID')
    return !term || `${supplier.name} ${supplier.contact ?? ''}`.toLocaleLowerCase('id-ID').includes(term)
  })
  const pages = Math.max(1, Math.ceil(filteredItems.length / pageSize))
  const safePage = Math.min(page, pages - 1)
  const pageItems = filteredItems.slice(safePage * pageSize, safePage * pageSize + pageSize)

  return (
    <>
      <PageHeader
        title="Supplier"
       
        crumb="Supplier"
        actions={<Button onClick={openCreate}><Plus className="size-4" />Tambah Supplier</Button>}
      />

      {suppliers.err && <AlertStrip className="mb-4">{suppliers.err}</AlertStrip>}

      {items.length === 0 && suppliers.err ? null : items.length === 0 && !suppliers.loading ? (
        <Empty title="Belum ada supplier" sub="Supplier akan muncul setelah ditambahkan." />
      ) : (
        <>
          <div className="mb-4 flex flex-col gap-3 sm:flex-row">
            <SearchInput value={q} onChange={(e) => { setQ(e.target.value); setPage(0) }} placeholder="Cari nama atau kontak supplier…" aria-label="Cari supplier" />
            <Select
              label="Filter supplier"
              value={supFilter}
              onChange={(v) => { setSupFilter(v); setPage(0) }}
              className="shrink-0 sm:w-56"
              options={[{ value: '', label: 'Semua supplier' }, { value: 'utang', label: 'Punya hutang' }]}
            />
          </div>
          {filteredItems.length === 0 && !suppliers.loading ? (
            <Empty title="Tidak ada supplier yang cocok" sub="Coba kata kunci lain." />
          ) : (
          <>
          <TableWrap className="hidden min-[1280px]:block">
          <table className="w-full border-collapse">
            <thead><tr><Th>Nama</Th><Th>Kontak</Th><Th>Alamat</Th><Th right>Aksi</Th></tr></thead>
            {suppliers.loading && items.length === 0 ? (
              <SkeletonRows cols={4} />
            ) : (
              <tbody>
                {pageItems.map((supplier) => (
                  <tr key={supplier.id}>
                    <Td><span className="font-medium text-fg">{supplier.name}</span></Td>
                    <Td>{supplier.contact || '—'}</Td>
                    <Td><span className="block max-w-xs truncate" title={supplier.address}>{supplier.address || '—'}</span></Td>
                    <Td>
                      <div className="flex gap-1">
                        <RowAction title="Ubah" label={`Ubah ${supplier.name}`} onClick={() => openEdit(supplier)}><Pencil className="size-4" /></RowAction>
                        <RowAction title="Hapus" label={`Hapus ${supplier.name}`} onClick={() => { setDeleteError(''); setDeleting(supplier) }} danger><Trash2 className="size-4" /></RowAction>
                      </div>
                    </Td>
                  </tr>
                ))}
              </tbody>
            )}
          </table>
        </TableWrap>
        <CardGrid>
          {suppliers.loading && items.length === 0 ? (
            [0, 1, 2].map((i) => <Skeleton key={i} className="h-[110px] w-full rounded-xl" />)
          ) : pageItems.map((supplier) => (
            <DataCard key={supplier.id}>
              <div className="flex items-center justify-between gap-2">
                <p className="min-w-0 truncate font-medium text-fg">{supplier.name}</p>
                {owedSupplierIds.has(supplier.id) && <Pill tone="warn">Punya hutang</Pill>}
              </div>
              <p className="mt-1 text-xs leading-snug text-fog">{supplier.contact || 'Tanpa kontak'}{supplier.address ? ` · ${supplier.address}` : ''}</p>
              <div className="mt-auto flex items-center gap-1 pt-1.5">
                <RowAction title="Ubah" label={`Ubah ${supplier.name}`} onClick={() => openEdit(supplier)}><Pencil className="size-4" /></RowAction>
                <RowAction title="Hapus" label={`Hapus ${supplier.name}`} onClick={() => { setDeleteError(''); setDeleting(supplier) }} danger><Trash2 className="size-4" /></RowAction>
              </div>
            </DataCard>
          ))}
        </CardGrid>
          </>
          )}
        <Pager page={safePage} total={pages} onChange={setPage} totalItems={filteredItems.length} pageSize={pageSize} onPageSizeChange={(n) => { setPageSize(n); setPage(0) }} />
        </>
      )}

      <Modal open={!!editing} title={editing?.id ? 'Ubah Supplier' : 'Tambah Supplier'} onClose={() => { if (!busy) setEditing(null) }}>
        {editing && (
          <div className="space-y-4">
            {formError && <AlertStrip>{formError}</AlertStrip>}
            <Input label="Nama supplier *" value={editing.name} onChange={(name) => setEditing((current) => current ? { ...current, name } : current)} required placeholder="Nama supplier" />
            <Input label="Kontak" value={editing.contact} onChange={(contact) => setEditing((current) => current ? { ...current, contact } : current)} placeholder="Telepon atau kontak" />
            <TextArea label="Alamat" value={editing.address} onChange={(v) => setEditing((current) => current ? { ...current, address: v } : current)} rows={3} placeholder="Alamat supplier" />
            <ModalActions>
              <Button variant="ghost" onClick={() => setEditing(null)} disabled={busy}>Batal</Button>
              <Button onClick={saveSupplier} disabled={busy}>{busy ? 'Menyimpan…' : 'Simpan'}</Button>
            </ModalActions>
          </div>
        )}
      </Modal>

      <Modal open={!!deleting} title="Hapus Supplier" onClose={() => { if (!busy) setDeleting(null) }}>
        {deleting && (
          <div className="space-y-4">
            <p className="text-sm text-muted">Hapus supplier <span className="font-medium text-fg">{deleting.name}</span> dari daftar?</p>
            {deleteError && <AlertStrip>{deleteError}</AlertStrip>}
            <ModalActions>
              <Button variant="ghost" onClick={() => setDeleting(null)} disabled={busy}>Batal</Button>
              <Button variant="danger" onClick={removeSupplier} disabled={busy}>{busy ? 'Menghapus…' : 'Hapus'}</Button>
            </ModalActions>
          </div>
        )}
      </Modal>
    </>
  )
}
