# F1.3 Backend Inventori Tahap 1 Implementation Plan

> **Scope:** Backend only, as requested. No commit.

**Goal:** Add product variants, quantity price tiers, and minimum-stock alerts without changing checkout behavior for existing products.

**Architecture:** Keep catalog operations in the existing CatalogHandler/CatalogService/ProductRepo layers. Keep checkout pricing, stock mutations, refund snapshots, and stock movements in TrxRepo's transaction. Store settings in the existing Store model and settings flow. Reuse notification records with a store-scoped daily deduplication key.

**Tech:** Go, Gin, GORM AutoMigrate, SQLite router integration tests.

**Source:** `PLAN.md`, Unit F1.3; task-specific backend requirements in the user request.

## Constraints

- All routes remain under `/api/v1`; mutations are admin-only, reads require login, and store scope comes from claims.
- Keep camelCase JSON request fields and existing snake_case model response tags.
- Product stock/prices remain authoritative only for products without variants; variant stock/prices are authoritative for variant sales.
- Checkout resolves applicable tier before promo, then tax; only the server-selected price is stored in transaction items.
- `minStock == 0` disables alerts. `lowStockDefault` is stored as a settings suggestion only.
- All schema changes use GORM AutoMigrate; no SQL migrations and no new dependencies.
- Preserve old product checkout and refund data.

## Final API shape

- Authenticated reads: `GET /products/:id/variants`, `GET /products/:id/price-tiers`, and `GET /products/:id/variants/:variantId/price-tiers`.
- Admin variant mutations: `POST /products/:id/variants`, `PUT|DELETE /products/:id/variants/:variantId`.
- Admin product-tier mutations: `POST /products/:id/price-tiers`, `PUT|DELETE /products/:id/price-tiers/:tierId`.
- Admin variant-tier mutations: `POST /products/:id/variants/:variantId/price-tiers`, `PUT|DELETE /products/:id/variants/:variantId/price-tiers/:tierId`.
- Product list/detail responses include `variants` and product `price_tiers`; each variant includes its own `price_tiers`.
- Checkout and refund items accept optional `variantId`.

## Review focus

- A variant from another product or store must never pass checkout or nested CRUD ownership checks.
- Duplicate product/variant lines merge only when both IDs match, so quantity tiers use the combined line quantity without mixing variants.
- Product-level tiers are ignored for variant checkout; a variant uses only its own tier list and base price fallback.
- Existing refund JSON remains readable while variant refunds are keyed to the sold variant.
- Daily stock alerts deduplicate per store and product/variant, including after the previous alert was read.
- Dashboard low stock excludes disabled thresholds and stale parent product stock when variants exist.

## Tasks

### 1. Catalog models, nested CRUD, and product responses

- [x] Write router integration coverage for variant and tier nested CRUD, login/admin access, store ownership, product list/detail embedding, and minimum-stock settings fields.
- [x] Run the focused test and confirm routes/fields are missing.
- [x] Add ProductVariant and PriceTier models, Product min_stock/relationships, Store low_stock_default, and AutoMigrate registration.
- [x] Implement scoped nested CRUD through ProductRepo/CatalogService/CatalogHandler and register authenticated reads/admin mutations.
- [x] Run the focused test and confirm it passes.

### 2. Variant and price-tier checkout

- [x] Add router tests for required variant selection, cross-owner variant rejection, variant stock movement/decrement, tier boundary/fallback, variant-tier precedence, and existing product regression.
- [x] Run the focused tests and confirm they fail on missing checkout behavior.
- [x] Resolve product/variant and applicable tier within the existing checkout transaction; persist variant ID and actual unit price to transaction snapshots and movement rows.
- [x] Apply the same tier quote to promo preview so promo calculations start from the tier-adjusted subtotal.
- [x] Run the focused tests and confirm they pass.

### 3. Variant refunds

- [x] Add a router test that sells and partially refunds a variant, then checks variant stock, product stock, refund detail, and movement variant ID.
- [x] Run the focused test and confirm variant refunds are not represented yet.
- [x] Persist variant refund quantities alongside the legacy product refund map and restore stock to the same variant.
- [x] Run the focused test and confirm it passes.

### 4. Minimum-stock alerts and dashboard threshold

- [x] Add router tests for one alert across consecutive sales, alert disabled at zero, variant-specific text/reference, adjustment/refund stock paths, and dashboard product/variant counts.
- [x] Run the focused tests and confirm current fixed threshold/dedup behavior differs.
- [x] Check alerts against each item's positive min_stock and use a unique store-scoped local-day key so read notifications still count for deduplication.
- [x] Replace dashboard threshold 5 with positive product/variant min_stock checks; save `lowStockDefault` through settings.
- [x] Run the focused tests and confirm they pass.

### 5. Final verification

- [x] Run gofmt on all changed Go files and verify `gofmt -l` is empty.
- [x] Run `CGO_ENABLED=1 go test ./...`.
- [ ] Review diff against F1.3, ensure no unrelated/deployment files changed, and leave all work uncommitted.
