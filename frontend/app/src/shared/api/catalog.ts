// Endpoint katalog: kategori, produk, varian, tingkat harga, impor.

import { BASE, ApiError, getToken, refreshTokens, request, requestFile, qs } from './core.ts'
import type { Category, Product, PriceTier, Variant, Page, ProductFilter } from './types.ts'

export interface ProductImportFailure { baris: number; alasan: string }
export interface ProductImportResult { total: number; berhasil: number; gagal: number; kesalahan: ProductImportFailure[] }

async function requestForm<T>(path: string, form: FormData): Promise<T> {
  let token = getToken()
  for (let attempt = 0; attempt < 2; attempt++) {
    const res = await fetch(BASE + path, {
      method: 'POST',
      headers: token ? { Authorization: `Bearer ${token}` } : {},
      body: form,
    })
    if (res.status === 401 && attempt === 0 && token && !path.startsWith('/auth/')) {
      token = await refreshTokens()
      continue
    }
    const data = await res.json().catch(() => ({}))
    if (!res.ok) throw new ApiError(res.status, data.error ?? 'Terjadi kesalahan.', data.code ?? data.error ?? '')
    return data as T
  }
  throw new ApiError(401, 'Sesi berakhir. Silakan masuk kembali.')
}

export function apiImportProducts(file: File) {
  const form = new FormData()
  form.append('file', file)
  return requestForm<ProductImportResult>('/products/import', form)
}

export function apiImportCategories(file: File) {
  const form = new FormData()
  form.append('file', file)
  return requestForm<ProductImportResult>('/categories/import', form)
}

export function apiListCategories() {
  return request<Page<Category>>('GET', '/categories').then((d) => d.items ?? [])
}

export function apiListProductUnits() {
  return request<Page<string>>('GET', '/products/units').then((d) => d.items ?? [])
}

export function apiCreateCategory(name: string) {
  return request<{ category: Category }>('POST', '/categories', { name })
}

export function apiRenameCategory(id: string, name: string) {
  return request<{ message: string }>('PATCH', `/categories/${id}`, { name })
}

export function apiDeleteCategory(id: string) {
  return request<{ soft_deleted: boolean }>('DELETE', `/categories/${id}`)
}

export function apiListProducts(f: ProductFilter = {}) {
  return request<Page<Product>>('GET', '/products' + qs({ ...f }))
}

export function apiGetProduct(id: string) {
  return request<Product>('GET', `/products/${id}`)
}

export function apiExportProducts() {
  return requestFile('/products/export')
}

export function apiCreateProduct(body: {
  name: string; sku: string; barcode?: string; categoryId?: string | null
  buyPrice?: number; sellPrice: number; stock?: number; minStock?: number; unit?: string
}) {
  return request<Product>('POST', '/products', body)
}

export function apiUpdateProduct(id: string, body: {
  name: string; sku: string; barcode?: string; categoryId?: string | null
  buyPrice?: number; sellPrice: number; minStock?: number; unit?: string
}) {
  return request<Product>('PUT', `/products/${id}`, body)
}

export type VariantInput = {
  name: string; sku: string; buyPrice: number; sellPrice: number; stock: number; minStock?: number
}

export type PriceTierInput = { minQty: number; sellPrice: number }

export function apiCreateVariant(productId: string, body: VariantInput) {
  return request<Variant>('POST', `/products/${productId}/variants`, body)
}

export function apiUpdateVariant(productId: string, variantId: number, body: VariantInput) {
  return request<Variant>('PUT', `/products/${productId}/variants/${variantId}`, body)
}

export function apiDeleteVariant(productId: string, variantId: number) {
  return request<{ message: string }>('DELETE', `/products/${productId}/variants/${variantId}`)
}

export function apiCreateProductPriceTier(productId: string, body: PriceTierInput) {
  return request<PriceTier>('POST', `/products/${productId}/price-tiers`, body)
}

export function apiUpdateProductPriceTier(productId: string, tierId: number, body: PriceTierInput) {
  return request<PriceTier>('PUT', `/products/${productId}/price-tiers/${tierId}`, body)
}

export function apiDeleteProductPriceTier(productId: string, tierId: number) {
  return request<{ message: string }>('DELETE', `/products/${productId}/price-tiers/${tierId}`)
}

export function apiCreateVariantPriceTier(productId: string, variantId: number, body: PriceTierInput) {
  return request<PriceTier>('POST', `/products/${productId}/variants/${variantId}/price-tiers`, body)
}

export function apiUpdateVariantPriceTier(productId: string, variantId: number, tierId: number, body: PriceTierInput) {
  return request<PriceTier>('PUT', `/products/${productId}/variants/${variantId}/price-tiers/${tierId}`, body)
}

export function apiDeleteVariantPriceTier(productId: string, variantId: number, tierId: number) {
  return request<{ message: string }>('DELETE', `/products/${productId}/variants/${variantId}/price-tiers/${tierId}`)
}

export function apiSetProductActive(id: string, active: boolean) {
  return request<{ message: string }>('PATCH', `/products/${id}/active`, { active })
}

export function apiDeleteProduct(id: string) {
  return request<{ message: string }>('DELETE', `/products/${id}`)
}

// ── stok ─────────────────────────────────────────────────────────────
