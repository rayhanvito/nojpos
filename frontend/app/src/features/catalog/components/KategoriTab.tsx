import { useState, type FormEvent } from 'react'
import { Pencil, Plus, Trash2 } from 'lucide-react'
import type { Category } from '../../../shared/api'
import { AlertStrip, Button, CardGrid, DataCard, Empty, Input, Modal, ModalActions, RowAction, TableWrap, Td, Th } from '../../../shared/ui'

export function KategoriTab({ categories, loading, busy, onAdd, onRename, onRemove }: {
  categories: Category[]
  loading: boolean
  busy: boolean
  onAdd: (name: string) => Promise<void>
  onRename: (category: Category, name: string) => Promise<void>
  onRemove: (category: Category) => void
}) {
  const [formOpen, setFormOpen] = useState(false)
  const [editing, setEditing] = useState<Category | null>(null)
  const [name, setName] = useState('')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  function openAdd() {
    setEditing(null)
    setName('')
    setError('')
    setFormOpen(true)
  }

  function openRename(category: Category) {
    setEditing(category)
    setName(category.name)
    setError('')
    setFormOpen(true)
  }

  function closeForm() {
    if (saving) return
    setFormOpen(false)
    setEditing(null)
    setError('')
  }

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const nextName = name.trim()
    if (!nextName) return setError('Nama kategori wajib diisi.')
    setSaving(true)
    setError('')
    try {
      if (editing) await onRename(editing, nextName)
      else await onAdd(nextName)
      setFormOpen(false)
      setEditing(null)
    } catch (e) {
      setError(e instanceof Error ? e.message : `Gagal ${editing ? 'mengubah' : 'menambah'} kategori.`)
    } finally {
      setSaving(false)
    }
  }

  function rowActions(category: Category) {
    return (
      <div className="flex shrink-0 items-center gap-1">
        <RowAction title="Ubah nama" label={`Ubah nama kategori ${category.name}`} onClick={() => openRename(category)} disabled={busy || saving}>
          <Pencil className="size-4" />
        </RowAction>
        <RowAction title="Hapus" label={`Hapus kategori ${category.name}`} onClick={() => onRemove(category)} disabled={busy || saving} danger>
          <Trash2 className="size-4" />
        </RowAction>
      </div>
    )
  }

  return (
    <>
      <div className="mb-4 flex items-center justify-between gap-3">
        <h2 className="text-sm font-semibold">Daftar kategori</h2>
        <Button onClick={openAdd} disabled={busy || saving}><Plus className="size-4" />Tambah Kategori</Button>
      </div>

      {loading ? (
        <Empty title="Memuat kategori…" sub="Memuat daftar kategori." />
      ) : categories.length === 0 ? (
        <Empty title="Belum ada kategori" sub="Tambah kategori untuk mengelompokkan produk." />
      ) : (
        <>
          <TableWrap className="hidden min-[1280px]:block">
            <table className="w-full border-collapse">
              <thead>
                <tr><Th>Kategori</Th><Th>Jumlah Produk</Th><Th>Aksi</Th></tr>
              </thead>
              <tbody>
                {categories.map((category) => (
                  <tr key={category.id}>
                    <Td><span className="font-medium text-fg">{category.name}</span></Td>
                    <Td>{category.product_count} produk</Td>
                    <Td>{rowActions(category)}</Td>
                  </tr>
                ))}
              </tbody>
            </table>
          </TableWrap>

          <CardGrid>
            {categories.map((category) => (
              <DataCard key={category.id}>
                <div className="flex min-w-0 items-start justify-between gap-3">
                  <p className="min-w-0 flex-1 truncate text-[15px] font-medium text-fg" title={category.name}>{category.name}</p>
                  <span className="shrink-0 text-sm text-muted">{category.product_count} produk</span>
                </div>
                <div className="mt-3 flex justify-end">{rowActions(category)}</div>
              </DataCard>
            ))}
          </CardGrid>
        </>
      )}

      <Modal open={formOpen} title={editing ? 'Ubah Nama Kategori' : 'Tambah Kategori'} onClose={closeForm}>
        <form className="space-y-3" onSubmit={(event) => void save(event)}>
          <Input label={editing ? 'Nama kategori' : 'Nama kategori baru'} value={name} onChange={setName} placeholder={editing ? undefined : 'cth: Minuman'} />
          {error && <AlertStrip>{error}</AlertStrip>}
          <ModalActions>
            <Button type="button" variant="ghost" disabled={busy || saving} onClick={closeForm}>Batal</Button>
            <Button type="submit" disabled={busy || saving}>{saving ? 'Menyimpan…' : editing ? 'Simpan Nama' : 'Simpan'}</Button>
          </ModalActions>
        </form>
      </Modal>
    </>
  )
}
