import { useEffect, useState, type FormEvent } from 'react'
import type { Category } from '../../../shared/api'
import { AlertStrip, Button, Input, Modal, ModalActions } from '../../../shared/ui'

export function KategoriFormModal({ open, editing, busy, onClose, onSubmit }: {
  open: boolean
  editing: Category | null
  busy: boolean
  onClose: () => void
  onSubmit: (name: string) => Promise<void>
}) {
  const [name, setName] = useState('')
  const [error, setError] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    if (!open) return
    setName(editing?.name ?? '')
    setError('')
  }, [open, editing])

  function closeForm() {
    if (saving) return
    onClose()
  }

  async function save(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    const nextName = name.trim()
    if (!nextName) return setError('Nama kategori wajib diisi.')
    setSaving(true)
    setError('')
    try {
      await onSubmit(nextName)
      onClose()
    } catch (e) {
      setError(e instanceof Error ? e.message : `Gagal ${editing ? 'mengubah' : 'menambah'} kategori.`)
    } finally {
      setSaving(false)
    }
  }

  return (
    <Modal open={open} title={editing ? 'Ubah Nama Kategori' : 'Tambah Kategori'} onClose={closeForm}>
      <form className="space-y-3" onSubmit={(event) => void save(event)}>
        <Input label={editing ? 'Nama kategori' : 'Nama kategori baru'} value={name} onChange={setName} placeholder={editing ? undefined : 'cth: Minuman'} />
        {error && <AlertStrip>{error}</AlertStrip>}
        <ModalActions>
          <Button type="button" variant="ghost" disabled={busy || saving} onClick={closeForm}>Batal</Button>
          <Button type="submit" disabled={busy || saving}>{saving ? 'Menyimpan…' : editing ? 'Simpan Nama' : 'Simpan'}</Button>
        </ModalActions>
      </form>
    </Modal>
  )
}
