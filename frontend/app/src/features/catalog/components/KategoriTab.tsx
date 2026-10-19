import { useMemo, useState } from 'react'
import { Pencil, Trash2 } from 'lucide-react'
import type { Category } from '../../../shared/api'
import { CardGrid, DataCard, Empty, FilterChips, FilterField, FilterPanel, RowAction, SearchInput, Select, TableWrap, Td, Th } from '../../../shared/ui'
import { filterAndSortCategories, type KategoriSort } from './kategori-list'

const ISI_OPTIONS = [
  { value: '', label: 'Semua' },
  { value: 'with', label: 'Ada produk' },
  { value: 'empty', label: 'Kosong' },
]

const SORT_OPTIONS: { value: KategoriSort; label: string }[] = [
  { value: 'name-asc', label: 'Nama A–Z' },
  { value: 'name-desc', label: 'Nama Z–A' },
  { value: 'count-desc', label: 'Produk terbanyak' },
  { value: 'count-asc', label: 'Produk tersedikit' },
]

export function KategoriTab({ categories, loading, busy, onRenameClick, onRemove }: {
  categories: Category[]
  loading: boolean
  busy: boolean
  onRenameClick: (category: Category) => void
  onRemove: (category: Category) => void
}) {
  const [q, setQ] = useState('')
  const [isi, setIsi] = useState('')
  const [sort, setSort] = useState<KategoriSort>('name-asc')
  const filtered = useMemo(
    () => filterAndSortCategories(categories, { q, isi, sort }),
    [categories, q, isi, sort],
  )

  function rowActions(category: Category) {
    return (
      <div className="flex shrink-0 items-center gap-1">
        <RowAction title="Ubah nama" label={`Ubah nama kategori ${category.name}`} onClick={() => onRenameClick(category)} disabled={busy}>
          <Pencil className="size-4" />
        </RowAction>
        <RowAction title="Hapus" label={`Hapus kategori ${category.name}`} onClick={() => onRemove(category)} disabled={busy} danger>
          <Trash2 className="size-4" />
        </RowAction>
      </div>
    )
  }

  return (
    <>
      <div className="mb-4 flex flex-col gap-3 sm:flex-row">
        <SearchInput value={q} onChange={(event) => setQ(event.target.value)} placeholder="Cari nama kategori…" aria-label="Cari kategori" />
        <Select label="Urutkan kategori" value={sort} onChange={(value) => setSort(value as KategoriSort)} className="shrink-0 sm:w-56" options={SORT_OPTIONS} />
        <FilterPanel className="shrink-0 sm:w-44" activeCount={isi ? 1 : 0} onReset={() => setIsi('')}>
          <FilterField label="Isi kategori">
            <Select label="Filter isi kategori" value={isi} onChange={setIsi} options={ISI_OPTIONS} />
          </FilterField>
        </FilterPanel>
      </div>
      <FilterChips items={isi ? [{ key: 'isi', label: `Isi: ${ISI_OPTIONS.find((option) => option.value === isi)?.label ?? isi}`, onClear: () => setIsi('') }] : []} />

      {loading ? (
        <Empty title="Memuat kategori…" sub="Memuat daftar kategori." />
      ) : categories.length === 0 ? (
        <Empty title="Belum ada kategori" sub="Tambah kategori untuk mengelompokkan produk." />
      ) : filtered.length === 0 ? (
        <Empty title="Tidak ada kategori yang cocok" sub="Coba ubah kata kunci atau filter." />
      ) : (
        <>
          <TableWrap className="hidden min-[1280px]:block">
            <table className="w-full border-collapse">
              <thead>
                <tr><Th>Kategori</Th><Th>Jumlah Produk</Th><Th>Aksi</Th></tr>
              </thead>
              <tbody>
                {filtered.map((category) => (
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
            {filtered.map((category) => (
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
    </>
  )
}
