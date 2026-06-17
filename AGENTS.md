# AGENTS.md - NojPOS Monorepo

## Source of Truth

- Product and architecture source of truth: `docs/NOJPOS_PRD.md` Revisi 8.
- ADR source files live in `docs/adr/`.

## Project

- NojPOS adalah POS SaaS Indonesia.
- Monorepo ini berisi Flutter Cashier App dan Laravel API.
- Future Next.js Admin dapat ditambahkan di `apps/admin/`.

## Shared Non-Negotiables

- MVP online-first.
- Dilarang membuat `/sync/pull`, `/sync/push`, Drift production DB, sync queue umum, atau conflict resolver di MVP.
- Semua nilai uang adalah integer rupiah. Tidak boleh memakai float untuk uang.
- Semua endpoint API berada di prefix `/api/v1` dan memakai header `Accept: application/json`.
- Multi-tenancy memakai `business_id` dari token/session, bukan input client.
- Idempotency wajib untuk write transaksional dengan header `Idempotency-Key`.
- Error envelope API wajib konsisten: sukses `{ "data": {}, "meta": {} }`; error `{ "error": { "code", "message", "details" } }`.
- Jangan mengubah logika, endpoint, skema, atau behavior di luar scope task.

## App Routing

- Flutter Cashier detail rules: `apps/cashier/AGENTS.md`.
- Laravel API detail rules: `apps/backend/AGENTS.md`.
- Root docs: `docs/NOJPOS_PRD.md` dan `docs/adr/`.

## Verification Gates

Sebelum task dianggap selesai:

- Backend: `cd apps/backend && php artisan test`.
- Backend route check: `cd apps/backend && php artisan route:list --path=api/v1`.
- Cashier: `cd apps/cashier && flutter analyze`.
- Cashier: `cd apps/cashier && flutter test`.
- Cashier: `cd apps/cashier && flutter build apk --debug`.

Format laporan:

`Built X. Verified Y. Blocker Z + alasan persis.`
