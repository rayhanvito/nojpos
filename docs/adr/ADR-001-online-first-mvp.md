# ADR-001: Online-first MVP supersedes offline-first mandate

Tanggal: 2026-06-17

Status: Accepted

## Context

NojPOS sebelumnya diarahkan sebagai Flutter cashier app offline-first dengan Drift, sync queue, conflict resolver, dan endpoint `/sync/pull` serta `/sync/push` sejak fase awal.

Hasil audit aplikasi saat ini menunjukkan:

- Aplikasi masih 100% UI prototype di atas local Riverpod + seed mock.
- REAL API: 0; tidak ada `dio`/`http`.
- Repository interfaces ada tetapi belum diimplementasikan: `lib/core/repositories/nojpos_repositories.dart:4`.
- Jalur inti POS sudah hidup secara lokal: login -> PIN -> produk -> cart -> simpan order -> bayar -> struk.
- Tidak ada Drift, sync queue, atau conflict resolver yang benar-benar dibangun.

Untuk MVP SaaS Indonesia, risiko terbesar adalah membangun offline-first terlalu awal sebelum kontrak API, tenancy, auth, transaction integrity, dan cashier workflow stabil.

NojPOS juga membutuhkan fondasi SaaS multi-tenant:

- Satu Laravel API melayani Flutter Cashier App dan Next.js Admin/Superadmin.
- Satu database dengan `business_id` scoping.
- Sanctum token auth.
- Role/gate untuk superadmin, owner/admin, dan cashier.

## Decision

NojPOS MVP resmi menjadi online-first POS SaaS.

Keputusan:

- Flutter Cashier App memakai REST API + Sanctum token.
- Next.js Admin/Superadmin memakai REST API + Sanctum token.
- Laravel API menjadi satu backend untuk semua client.
- Database MVP adalah single database multi-tenant dengan `business_id` di semua tabel domain.
- Laravel wajib menerapkan tenant scoping otomatis dan authorization via Policy/Gate.
- Idempotency key wajib sejak hari pertama untuk write transaksional:
  - `POST /transactions`
  - `POST /payments`
  - `POST /refunds`
  - `POST /voids`
- MVP menyertakan checkout outbox ringan:
  - hanya untuk checkout transaction submit,
  - satu arah device -> server,
  - retry saat reconnect dan saat app foreground,
  - exponential backoff 5s -> 30s -> 2m dengan cap 2m,
  - needs-action setelah 5 kegagalan beruntun atau 15 menit belum terkirim,
  - tidak pernah auto-discard; resolusi manual wajib jika tidak bisa terkirim,
  - close shift diblokir selama masih ada outbox pending/sending/needs-action untuk shift tersebut,
  - tidak melakukan pull sync,
  - tidak menangani conflict resolver.
- Full offline-first menjadi Phase 2/future:
  - Drift production DB,
  - sync queue dua arah,
  - conflict resolver,
  - `/sync/pull`,
  - `/sync/push`,
  - `sync_status`,
  - `SyncQueueItem`,
  - server reconciliation.
- DB-per-tenant juga Phase 2/future, bukan MVP.

## Consequences

### Positive

- MVP lebih realistis dan lebih cepat menuju transaksi end-to-end real.
- API contract, tenant scoping, Sanctum auth, role gate, dan idempotency bisa distabilkan lebih awal.
- Flutter tetap cashier-first dan tidak perlu langsung membawa beban Drift/sync engine.
- Repository boundary tetap menjaga jalur migrasi ke offline-first penuh.
- Checkout outbox mengurangi risiko transaksi hilang saat network gagal tanpa membangun sync dua arah.

### Negative

- Kasir tetap membutuhkan koneksi untuk master data dan operasi utama.
- Offline support di MVP terbatas hanya pada retry checkout yang sudah selesai disusun.
- Beberapa kebutuhan lapangan dengan koneksi buruk belum terjawab sepenuhnya sampai Phase 2.
- Semua elemen DEAD/MOCK/PARTIAL tetap harus di-real-kan dalam MVP wave, sehingga scope masih perlu manajemen ketat.

### Superseded Decisions

Bagian dokumen lama yang disupersede:

- Pernyataan bahwa Flutter cashier app wajib offline-first sejak MVP.
- Keputusan memasang Drift production DB sebagai fase awal.
- Keputusan membangun sync queue dua arah dan conflict resolver sebelum API online-first stabil.
- Keputusan memakai `/sync/pull` dan `/sync/push` sebagai endpoint MVP.
- Keputusan menjadikan `sync_status`, `server_id`, dan `SyncQueueItem` sebagai field produksi MVP.

Bagian yang tetap dipertahankan:

- Flutter tetap feature-first.
- Repository boundary tetap wajib.
- Laravel tetap API-only.
- Next.js tetap Admin/Superadmin.
- Payment tetap dipisah dari Transaction.
- Stock changes tetap melalui StockMovement.
- Paid transaction tetap immutable dan dikoreksi lewat void/refund.

## Links

- `docs/NOJPOS_PRD.md`
- `docs/ARCHITECTURE_DECISION.md`
- `docs/FLUTTER_REFACTOR_MIGRATION_PLAN.md`
