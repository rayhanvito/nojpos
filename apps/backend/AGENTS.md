# AGENTS.md - NojPOS Laravel API

## Project

- NojPOS adalah POS SaaS Indonesia.
- Backend ini adalah single Laravel API multi-tenant untuk Flutter Cashier App dan Next.js Admin/Superadmin.
- Greenfield: tidak ada legacy/data migration. Jika sebuah path lama tidak perlu, hapus; jangan buat jembatan kompatibilitas.
- Source of truth produk dan arsitektur: `../../docs/NOJPOS_PRD.md` Revisi 8.

## Aturan Non-Negotiable

- Semua endpoint berada di prefix `/api/v1`.
- Client wajib mengirim header `Accept: application/json`.
- MVP online-first. Dilarang membuat `/sync/pull`, `/sync/push`, atau machinery offline-first apa pun.
- Semua nilai uang adalah integer rupiah dengan storage bigint. Tidak boleh memakai float untuk uang.
- Multi-tenancy memakai single database dan kolom `business_id` di semua tabel domain.
- Resolusi `business_id` berasal dari token, bukan input client.
- Tenant global scope wajib dipakai untuk tabel domain.
- Semua tabel domain memakai PK UUID v4, `created_at`, `updated_at`, dan SoftDeletes `deleted_at`.
- Auth memakai Laravel Sanctum token-based.
- Roles: `superadmin`, `owner`, `admin`, `cashier`.
- RBAC wajib lewat Policy/Gate per endpoint.
- Idempotency wajib untuk POST transaksional: `/transactions`, `/payments`, `/refunds`, `/voids`.
- Header idempotency wajib `Idempotency-Key` UUID v4.
- Dedup idempotency scoped per `business_id`: key + hash sama mengembalikan response tersimpan; key sama hash berbeda mengembalikan `409`.
- Rate-limit/lockout login, PIN, dan pin-switch: 5x gagal mengunci 15 menit.
- Audit log wajib untuk aksi uang dan auth-sensitive: login, pin-switch, shift open/close, cash in/out, void/refund.
- Audit log menyimpan actor, action, entity, before, after, timestamp, dan `business_id`; audit log immutable.
- Response sukses wajib `{ "data": {}, "meta": {} }`.
- Response error wajib `{ "error": { "code": "STRING", "message": "STRING", "details": {} } }`.
- Status baku: `200`, `201`, `401`, `403`, `404`, `409`, `422`, `500`.

## Gate Verifikasi

Sebelum task dianggap selesai atau sebelum commit, wajib:

- `php artisan test` hijau.
- Sertakan test isolasi tenant: business A tidak bisa baca/tulis data business B.

Format laporan wajib:

`Built X. Verified Y (test). Blocker Z (kalau ada) + alasan persis.`

Jangan klaim selesai tanpa bukti test.

## Disiplin Kerja

- Satu task = satu area repo = satu diff bersih.
- Jangan menjelajah atau refactor area lain tanpa diminta.
- Pakai Laravel Boost untuk verifikasi state nyata sebelum menebak: application info, database schema, DB query, route inspector, artisan, tinker, search docs, last errors, dan log reader.
- Jangan beri tool akses tulis ke database produksi; gunakan hanya environment lokal/dev.

## Tooling AI

- Laravel Boost terpasang sebagai dev dependency.
- MCP server Boost dijalankan lewat `php artisan boost:mcp`.
- Konfigurasi Codex lokal tersedia di `.codex/config.toml`.
- Prompt boleh memakai `use context7` untuk dokumentasi version-specific Laravel dan Sanctum versi terpasang.
- Jangan hardcode API key Context7 di repo.
