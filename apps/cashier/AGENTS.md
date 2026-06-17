# AGENTS.md - NojPOS Flutter Cashier

## Project

- NojPOS Flutter Cashier App adalah aplikasi kasir untuk POS SaaS Indonesia.
- Aplikasi ini online-first, cashier-first, dan memakai Laravel API dengan prefix `/api/v1`.
- Greenfield: tidak ada legacy. Saat ini fase repository seam sudah ada (Wave 0 selesai); implementasi API-backed repository adalah Wave 1.
- Source of truth produk dan arsitektur: `../../docs/NOJPOS_PRD.md` Revisi 8.

## Arsitektur & Aturan Non-Negotiable

- Struktur wajib feature-first.
- Golden rule aliran data: UI -> provider -> repository -> network.
- UI tidak boleh import `core/mock/mock_app_state.dart` langsung; akses state/mock harus lewat provider.
- MVP online-first. Dilarang membuat folder offline-first: `core/database/`, `core/sync/`, atau `features/sync_status/`.
- Tidak ada Drift, sync queue, atau conflict resolver di MVP.
- Checkout outbox hanya untuk checkout, satu arah device -> server.
- Checkout outbox tidak boleh menyentuh master data, tidak boleh pull sync, dan tidak boleh punya conflict resolver.
- Idempotency: client generate UUID v4 per aksi transaksional dan kirim header `Idempotency-Key`.
- Retry checkout harus memakai key yang sama. Jangan generate key baru dan jangan membuat transaksi kedua saat retry.
- Semua nilai uang adalah integer rupiah. Tidak boleh memakai float untuk uang.
- Semua akses data fitur wajib lewat repository interface: Auth, Product, Customer, Order, Transaction, Inventory, Attendance, Shift.
- Swap Mock -> API dilakukan per fitur via provider; jangan menyentuh widget lebih dari perlu.
- Dependency baru harus disetujui eksplisit. Jangan menambah package tanpa izin.

## Gate Verifikasi

Sebelum task dianggap selesai atau sebelum commit, wajib jalankan:

- `flutter analyze` dan hasilnya `No issues found!`
- `flutter test` dan semua test pass.
- `flutter build apk --debug` dan build sukses.

Format laporan wajib:

`Built X. Verified Y (analyze/test/build). Blocker Z + alasan persis.`

Jangan klaim selesai tanpa bukti test.

## Disiplin Kerja

- Satu task = satu area = satu diff bersih.
- `operations_screen.dart` sengaja belum dipecah sampai Wave 3; jangan utak-atik kecuali diminta.
- Pakai Dart MCP untuk loop analyze -> fix -> test, bukan menebak.

## Tooling AI

- Dart & Flutter MCP server dijalankan dengan `dart mcp-server`.
- Konfigurasi Codex lokal tersedia di `.codex/config.toml`.
- SDK lokal saat setup: Flutter `C:\flutter`, Dart `C:\flutter\bin\dart.bat`.
- Prompt boleh memakai `use context7` untuk dokumentasi version-specific Riverpod, go_router, intl, dan package Flutter lain yang terpasang.
- Jangan hardcode API key Context7 di repo.
- `flutter-mcp-toolkit` runtime ditunda sampai Wave 1 sebagai opsi verifikasi UI runtime; jangan install sekarang.
