import type { ProductImportResult } from '../../../shared/api'
import { ModalActions, AlertStrip, Button, Modal } from '../../../shared/ui'

export function ImportProdukModal({ open, onClose, busy, result, error, onDownloadTemplate, onPickFile, onImportAgain, onDone }: {
  open: boolean
  onClose: () => void
  busy: boolean
  result: ProductImportResult | null
  error: string
  onDownloadTemplate: () => void
  onPickFile: () => void
  onImportAgain: () => void
  onDone: () => void
}) {
  return (
<Modal open={open} title="Import Produk" onClose={onClose} wide>
  <div className="space-y-4">
    {!result ? (
      <>
        <p className="text-sm text-muted">
          Format file harus sama persis dengan hasil Ekspor Produk:
          {' '}<span className="font-mono text-[13px]">id, nama, sku, barcode, kategori, harga beli, harga jual, stok, satuan, aktif</span>{' '}
          (kolom id boleh kosong). Import hanya membuat produk baru — SKU yang sudah terdaftar ditolak dan dilaporkan, tidak pernah menimpa data lama. Kategori yang belum ada otomatis dibuat. Maksimal 1.000 baris per file.
        </p>
        {error && <AlertStrip>{error}</AlertStrip>}
        <ModalActions>
          <Button variant="ghost" onClick={onDownloadTemplate}>Unduh Template CSV</Button>
          <Button disabled={busy} onClick={onPickFile}>{busy ? 'Mengimpor…' : 'Pilih File CSV'}</Button>
        </ModalActions>
      </>
    ) : (
      <>
        <p className="text-sm">
          Total {result.total} baris: <strong className="text-sprout">{result.berhasil} berhasil</strong>,{' '}
          <strong className="text-ember">{result.gagal} gagal</strong>.
        </p>
        {result.kesalahan.length > 0 && (
          <div className="max-h-72 divide-y divide-dove overflow-y-auto rounded-lg border border-dove px-3">
            {result.kesalahan.map((k) => (
              <div key={k.baris} className="flex items-start gap-2.5 py-2 text-[13px]">
                <span className="shrink-0 rounded bg-surface px-1.5 py-0.5 font-mono text-xs text-fg">Baris {k.baris}</span>
                <span className="min-w-0 text-muted">{k.alasan}</span>
              </div>
            ))}
          </div>
        )}
        <ModalActions>
          <Button variant="ghost" onClick={onImportAgain}>Impor Lagi</Button>
          <Button onClick={onDone}>Selesai</Button>
        </ModalActions>
      </>
    )}
  </div>
</Modal>
  )
}
