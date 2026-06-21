# Story Mobile Kasir Flutter

Lokasi: `apps/cashier`  
Stack: Flutter, Riverpod, GoRouter, Dio.  
PRD: `PRDPOSJA.md`

Mobile Kasir adalah aplikasi operasional harian untuk kasir di shared terminal. Semua write penting harus online, server-verified, dan tidak menghitung uang final di client.

---

## MOB-01 — Auth, Outlet, Device, PIN, Terminal Session

Status: `[IN PROGRESS]`  
Priority: `P0`  
Dependency: BE-01.

### Tujuan

Kasir bisa login, memilih outlet, memakai device/terminal valid, PIN switch, lock/unlock terminal, dan melanjutkan shift dengan aman.

### Implemented Sekarang

- Struktur feature: `auth`, `outlet`, `screen_lock`, `store`, session-related flow.
- Backend route tersedia untuk login, logout, me, outlets, PIN switch, terminal lock/unlock/state.

### Task dan Subtask

- `[DONE]` Login flow dasar.
- `[DONE]` Outlet/session bootstrap dasar.
- `[IN PROGRESS]` Terminal/device context.
  - `[TODO]` QA device ID/enrollment behavior.
  - `[TODO]` QA terminal lock/unlock.
  - `[TODO]` QA PIN switch di shared terminal.
  - `[TODO]` Pastikan write sensitif tidak menerima actor/device palsu dari UI.
- `[READY]` Error state auth.
  - `[TODO]` 401 session expired.
  - `[TODO]` 403 role/outlet denied.
  - `[TODO]` server unavailable.
  - `[TODO]` offline/connection warning.
- `[READY]` Security logging.
  - `[TODO]` Pastikan token, PIN, password, payment reference tidak muncul di log/screenshot.

### Acceptance Criteria

- User tidak bisa masuk fitur outlet tanpa session valid.
- PIN switch mengganti actor terminal secara aman dan auditable.
- Lock/unlock terminal tidak bypass role/device.
- Error state jelas dan tidak membuat user stuck.

### Bug/Kendala

- `[OPEN]` BUG-MOB-01 — perlu QA device/emulator untuk lock/unlock dan PIN switch.
- `[OPEN]` Token/device credential harus dipastikan tidak muncul di log/screenshot.

---

## MOB-02 — POS Checkout

Status: `[IN PROGRESS]`  
Priority: `P0`  
Dependency: BE-01, BE-03.

### Tujuan

Kasir bisa memilih produk, membuat quote ke server, menerima pembayaran, menyelesaikan transaksi, dan melihat hasil transaksi dengan aman.

### Implemented Sekarang

- Feature folder: `pos`, `payment`, `transactions`, `customers`, `orders`.
- Backend route tersedia untuk quote, transaction store, transaction list/detail, payment, void, refund, recovery, parked orders.

### Task dan Subtask

- `[DONE]` Catalog/cart/order panel UI dasar.
- `[DONE]` Quote API flow tersedia.
- `[DONE]` Transaction create endpoint tersedia.
- `[IN PROGRESS]` Checkout happy path QA.
  - `[TODO]` Pilih produk.
  - `[TODO]` Tambah customer opsional.
  - `[TODO]` Quote server.
  - `[TODO]` Payment.
  - `[TODO]` Success receipt.
- `[IN PROGRESS]` Checkout retry/recovery.
  - `[TODO]` QA bounded retry sesuai PRD.
  - `[TODO]` Pastikan hanya satu in-flight checkout retry.
  - `[TODO]` Pastikan retry memakai idempotency key yang sama.
  - `[TODO]` Pastikan tidak menjadi general offline write queue.
- `[READY]` Payment UX hardening.
  - `[TODO]` Loading quote/store/payment.
  - `[TODO]` Quote stale/conflict.
  - `[TODO]` Payment failed/pending.
  - `[TODO]` Transaction recovery state.
- `[DEFERRED]` Sensitive actions.
  - `[TODO]` Void.
  - `[TODO]` Refund.
  - `[TODO]` Reprint gated by policy.

### Acceptance Criteria

- Total final yang ditampilkan berasal dari quote/server.
- Transaksi tidak bisa disubmit jika shift belum open.
- Retry checkout bounded dan tidak membuat duplicate transaction.
- Offline tidak boleh memulai transaksi baru.

### Bug/Kendala

- `[OPEN]` Perlu test end-to-end checkout di emulator/device.
- `[OPEN]` Pastikan semua total final dari server, bukan hitungan client.

---

## MOB-03 — Shift dan Kas

Status: `[IN PROGRESS]`  
Priority: `P0`  
Dependency: BE-01, BE-03.

### Tujuan

Kasir bisa buka shift, melihat shift berjalan, mencatat cash in/out, dan tutup shift dengan kas aktual.

### Implemented Sekarang

- Feature folder: `shift`, `store`.
- Backend route tersedia untuk shift open/current/cash movement/close dan store open/close state.

### Task dan Subtask

- `[DONE]` Shift endpoint tersedia.
- `[IN PROGRESS]` Mobile shift gate.
  - `[TODO]` QA tidak bisa checkout jika shift belum open.
  - `[TODO]` QA current shift render benar.
  - `[TODO]` QA close shift dengan kas aktual.
  - `[TODO]` QA cash in/out movement.
- `[READY]` Cash discrepancy UX.
  - `[TODO]` Tampilkan expected cash dan actual cash dari server.
  - `[TODO]` Tampilkan selisih dengan bahasa ramah.
  - `[TODO]` Jangan klaim akurat jika data belum final.
- `[READY]` Store open/close.
  - `[TODO]` QA owner/admin store open.
  - `[TODO]` QA store close.
  - `[TODO]` Pastikan close store tidak otomatis finalisasi cash shift.

### Acceptance Criteria

- Checkout gated oleh shift state.
- Shift close tidak race dengan transaksi aktif.
- Cash movement masuk audit backend.
- Kasir melihat error yang bisa ditindaklanjuti.

### Bug/Kendala

- `[OPEN]` Perlu QA race condition shift close saat transaksi masih berjalan.
- `[OPEN]` Web admin belum boleh menjalankan shift close/open real.

---

## MOB-04 — Inventory, Attendance, Staff, Reports

Status: `[IN PROGRESS]`  
Priority: `P1`  
Dependency: BE-01, BE-02, BE-03, BE-04.

### Tujuan

Kasir/owner/admin bisa melihat stok, absensi, staff, dan laporan sesuai role tanpa membuka data yang dilarang.

### Implemented Sekarang

- Feature folder: `inventory`, `attendance`, `reports`, `staff`.
- Backend route tersedia untuk inventory, attendance, staff, dan reports.

### Task dan Subtask

- `[DONE]` Inventory read endpoint tersedia.
- `[DONE]` Attendance endpoint tersedia.
- `[DONE]` Staff endpoint tersedia.
- `[DONE]` Reports endpoint tersedia.
- `[IN PROGRESS]` Mobile read screen QA.
  - `[TODO]` Inventory list dan low stock.
  - `[TODO]` Inventory movement visibility.
  - `[TODO]` Attendance submit/read.
  - `[TODO]` Staff list role-limited.
  - `[TODO]` Report summary.
- `[READY]` Role visibility.
  - `[TODO]` Kasir tidak melihat fitur owner/admin jika dilarang.
  - `[TODO]` 403 render dengan copy sederhana.
  - `[TODO]` Backend policy tetap sumber security.
- `[DEFERRED]` Inventory write from mobile owner/admin.
  - `[TODO]` Purchase/count/waste/transfer UX setelah backend hardening.

### Acceptance Criteria

- Screen membaca data sesuai role.
- Role violation menghasilkan 403, bukan crash.
- Reports tidak dihitung client-side.
- Attendance write idempotent dan auditable.

### Bug/Kendala

- `[OPEN]` Perlu review role visibility: hiding UI bukan security.

---

## MOB-05 — Receipt, Print/Share, Notification, Retry Center

Status: `[IN PROGRESS]`  
Priority: `P1`  
Dependency: MOB-02.

### Tujuan

Kasir bisa melihat struk transaksi, print/share jika adapter tersedia, dan memantau retry checkout bounded.

### Implemented Sekarang

- Feature folder: `notifications`.
- Core printing/share dan retry center disebut sudah tersedia di app.

### Task dan Subtask

- `[IN PROGRESS]` Receipt success screen.
  - `[TODO]` QA tampilan struk.
  - `[TODO]` QA print preview/adapter.
  - `[TODO]` QA share/copy receipt.
- `[IN PROGRESS]` Retry center.
  - `[TODO]` Tampilkan transaksi in-flight saja.
  - `[TODO]` Resolve recovery ke server result.
  - `[TODO]` Jangan simpan general offline queue.
- `[READY]` Hardware QA.
  - `[TODO]` Tentukan printer target.
  - `[TODO]` Test device nyata jika hardware sudah ada.

### Acceptance Criteria

- Receipt berasal dari server transaction result.
- Reprint/share tidak expose data sensitif berlebihan.
- Retry center tidak menyimpan banyak write pending.

### Bug/Kendala

- `[OPEN]` Perlu test printer adapter/device nyata jika target hardware sudah dipilih.
- `[OPEN]` Reprint/void/refund harus tetap policy-gated.
