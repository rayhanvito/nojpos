// API fitur inventory: permukaan endpoint yang dipakai halaman Stok
// dan Opname. Endpoint-nya sendiri tetap implementasi bersama di
// shared/api per domain (stock untuk opname/pergerakan/penyesuaian,
// catalog untuk produk/kategori) — file ini hanya mengekspor ulang
// yang dipakai fitur inventory, mengikuti pola fitur pos (satu pintu
// masuk API per fitur).
export {
  apiAdjustStock, apiCancelStockOpname, apiCreateStockOpname, apiGetStockOpname, apiListCategories, apiListMovements,
  apiListProducts, apiListStockOpnames, apiPostStockOpname, apiUpdateStockOpnameItems, fetchAll,
} from '../../shared/api'
export type {
  Category, Movement, Page, Product, StockOpnameCountsInput, StockOpnameDetail, StockOpnameItem, StockOpnameSession, StockOpnameStatus,
} from '../../shared/api'
