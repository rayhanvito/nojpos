// Klien API dipecah per domain di folder ./api/ (core, types, auth, catalog,
// sales, finance, stock, integrations, insights). File ini barrel agar jalur impor lama
// ('../shared/api') tetap berlaku untuk semua halaman.
export * from './api/core.ts'
export * from './api/types.ts'
export * from './api/auth.ts'
export * from './api/catalog.ts'
export * from './api/sales.ts'
export * from './api/finance.ts'
export * from './api/stock.ts'
export * from './api/integrations.ts'
export * from './api/insights.ts'
