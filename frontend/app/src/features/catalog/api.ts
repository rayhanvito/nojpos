// API fitur catalog: permukaan endpoint katalog yang dipakai halaman
// Produk & Kategori. Endpoint-nya sendiri tetap implementasi bersama di
// shared/api per domain (catalog.ts untuk kategori/produk/varian/harga
// bertingkat/impor; insights untuk pengaturan toko) — file ini hanya
// mengekspor ulang yang dipakai fitur catalog, mengikuti pola fitur pos
// (satu pintu masuk API per fitur).
export {
  apiCreateCategory, apiRenameCategory, apiCreateProduct, apiCreateProductPriceTier, apiCreateVariant, apiCreateVariantPriceTier,
  apiDeleteCategory, apiDeleteProduct, apiDeleteProductPriceTier, apiDeleteVariant, apiDeleteVariantPriceTier,
  apiExportProducts, apiGetProduct, apiGetSettings, apiImportCategories, apiImportProducts, apiListCategories, apiListProductUnits, apiListProducts,
  apiSetProductActive, apiUpdateProduct, apiUpdateProductPriceTier, apiUpdateVariant, apiUpdateVariantPriceTier,
  fetchAll,
} from '../../shared/api'
export type { Category, PriceTier, Product, ProductImportResult, StoreSettings, Variant } from '../../shared/api'
