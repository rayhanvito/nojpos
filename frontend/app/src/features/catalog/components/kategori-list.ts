import type { Category } from '../../../shared/api'

export type KategoriSort = 'name-asc' | 'name-desc' | 'count-desc' | 'count-asc'

export function filterAndSortCategories(
  categories: Category[],
  { q, isi, sort }: { q: string; isi: string; sort: KategoriSort },
): Category[] {
  const query = q.trim().toLowerCase()
  const filtered = categories.filter((category) => {
    if (query && !category.name.toLowerCase().includes(query)) return false
    if (isi === 'with' && category.product_count <= 0) return false
    if (isi === 'empty' && category.product_count !== 0) return false
    return true
  })

  return filtered.sort((a, b) => {
    if (sort === 'name-desc') return b.name.localeCompare(a.name)
    if (sort === 'count-desc') return b.product_count - a.product_count || a.name.localeCompare(b.name)
    if (sort === 'count-asc') return a.product_count - b.product_count || a.name.localeCompare(b.name)
    return a.name.localeCompare(b.name)
  })
}
