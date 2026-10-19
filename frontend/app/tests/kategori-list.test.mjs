import assert from 'node:assert/strict'
import { test } from 'node:test'

const { filterAndSortCategories } = await import('../src/features/catalog/components/kategori-list.ts')

const categories = [
  { id: 'c1', name: 'Makanan', product_count: 3 },
  { id: 'c2', name: 'Minuman', product_count: 1 },
  { id: 'c3', name: 'Alat Tulis', product_count: 3 },
  { id: 'c4', name: 'Kosong', product_count: 0 },
]

test('kategori search matches a case-insensitive name substring', () => {
  assert.deepEqual(
    filterAndSortCategories(categories, { q: 'iNuM', isi: '', sort: 'name-asc' }).map((category) => category.name),
    ['Minuman'],
  )
})

test('kategori isi filter returns all, populated, or empty categories', () => {
  assert.deepEqual(filterAndSortCategories(categories, { q: '', isi: '', sort: 'name-asc' }).map((category) => category.name), [
    'Alat Tulis', 'Kosong', 'Makanan', 'Minuman',
  ])
  assert.deepEqual(filterAndSortCategories(categories, { q: '', isi: 'with', sort: 'name-asc' }).map((category) => category.name), [
    'Alat Tulis', 'Makanan', 'Minuman',
  ])
  assert.deepEqual(filterAndSortCategories(categories, { q: '', isi: 'empty', sort: 'name-asc' }).map((category) => category.name), ['Kosong'])
})

test('kategori sort supports name and product count in both directions with name tiebreaks', () => {
  assert.deepEqual(filterAndSortCategories(categories, { q: '', isi: '', sort: 'name-asc' }).map((category) => category.name), [
    'Alat Tulis', 'Kosong', 'Makanan', 'Minuman',
  ])
  assert.deepEqual(filterAndSortCategories(categories, { q: '', isi: '', sort: 'name-desc' }).map((category) => category.name), [
    'Minuman', 'Makanan', 'Kosong', 'Alat Tulis',
  ])
  assert.deepEqual(filterAndSortCategories(categories, { q: '', isi: '', sort: 'count-desc' }).map((category) => category.name), [
    'Alat Tulis', 'Makanan', 'Minuman', 'Kosong',
  ])
  assert.deepEqual(filterAndSortCategories(categories, { q: '', isi: '', sort: 'count-asc' }).map((category) => category.name), [
    'Kosong', 'Minuman', 'Alat Tulis', 'Makanan',
  ])
})
