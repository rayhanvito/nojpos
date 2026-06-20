# NojPOS DESIGN.md

Canonical design system and UX foundation for NojPOS.

**Product:** NojPOS - online-first POS SaaS for Indonesian UMKM  
**Consumers:** Flutter cashier app (tablet-first + mobile) and future admin web design language. There is no active `apps/web`/Next.js implementation in the current repo.  
**Status:** v1 canonical design foundation  
**Primary evidence:** `Flow dan Contoh UI/` PDF examples + `NOJPOS_POS_FLUTTER_BACKEND_PRD.md` v2.0  
**Architecture note:** `docs/ARCHITECTURE.md` is present in the workspace and confirms the current buildable apps are Flutter cashier and Laravel API; there is no current `apps/web`/Next.js application. This design system still defines the future admin web language requested by product, while implementation mapping for the current Flutter app follows `UI -> Riverpod provider/notifier -> repository interface -> ApiClient`.

---

## 1. Design principles

1. **Kasir-first under queue pressure.** The cashier must always know outlet, cashier, shift, connection, current order, and next safe action. Avoid decorative UI that competes with checkout.
2. **Indonesia-first.** Labels use Bahasa Indonesia, familiar retail terms, integer rupiah display (`Rp12.500`), and outlet-local time.
3. **Stable work surfaces.** Tablet POS uses four persistent regions. Admin web uses dense tables and filters. Mobile uses one primary task per screen.
4. **Traceable money and stock.** Payment, shift, cash movement, refund, void, and inventory document screens must show document number/status, actor, outlet, and irreversible impact before final submit.
5. **Text + color, never color alone.** Every status has a readable label and, where helpful, helper text.
6. **Success lands somewhere durable.** A write success returns to a detail/list state with new status visible. Snackbar is supporting feedback, not the only confirmation.
7. **PRD behavior wins over examples when they conflict.** The example PDFs include local-server, full manual sync, WhatsApp/email/SMS direct receipt, camera attendance, sales invoice, and active promotions. Those are treated as visual/interaction references only when the PRD excludes or defers their behavior.

---

## 2. Source catalog: `Flow dan Contoh UI/`

The examples are tutorial-style PDFs with embedded POS screens. They establish the intended Indonesian POS language: teal primary actions, white/light-gray surfaces, dense lists, popup dialogs for edit/confirmation, side/menu navigation, filter controls, print actions, and form flows that end in `Simpan`, `Cetak`, `Tutup`, or a named destructive action.

### 2.1 Absensi

| File | Shows | Visible design decisions |
| --- | --- | --- |
| `Absensi/Absensi.pdf` | Attendance list, clock in, clock out | Attendance is a dedicated module. Uses staff selector + PIN continuation, period filter, recent list, and success confirmation. Visual palette uses white surface, teal primary action, and simple filter affordance. |

### 2.2 Buka dan Tutup

| File | Shows | Visible design decisions |
| --- | --- | --- |
| `Buka dan Tutup/Login Logout POS.pdf` | Login, app entry, staff PIN, logout cashier, app logout | Entry flow is staged: app open -> account login -> staff/PIN. Keypad/PIN is a recurring authentication surface. Logout cashier is distinct from app logout. |
| `Buka dan Tutup/Tutup Kasir.pdf` | Close cashier/shift | Close shift uses popup confirmation, PIN, close report, optional print checkboxes, then final `Tutup Kasir`. Must not be one tap. |
| `Buka dan Tutup/Tutup Toko.pdf` | Store close | Store close is privileged, requires PIN/code, and uses explicit `Tutup Toko` final action. PRD adds blocking checks for open/pending shifts. |

### 2.3 Inventori

| File | Shows | Visible design decisions |
| --- | --- | --- |
| `Inventori/Faktur Pembelian POS.pdf` | Purchase document create/edit/print | Inventory write uses form -> product picker -> review -> `Simpan` -> confirm (`Ya, Lanjutkan`) -> print/detail. Document screens are list/detail-first. |
| `Inventori/Permintaan Stok.pdf` | Stock request, dispatch, partial process, void, print | Transfer-like flows are multi-state. Uses source/destination outlets, notes, product lines, status, action menus, and print. No silent stock movement. |
| `Inventori/Stok Opname.pdf` | Stock count create/detail/print | Count flow uses document list, `Tambah`, product picker, actual count, `Simpan`, confirmation, detail, print. |
| `Inventori/Stok Terbuang.pdf` | Waste record create/status/void | Waste is a document with outlet/date/note/product lines/status and void action. Reason and immutable status are important. |
| `Inventori/Terima Mutasi Stok.pdf` | Transfer receipt | Receipt flow uses add screen, accepted item picker, status, and filters. Supports PRD transfer receive/partial receive pattern. |

### 2.4 Kasir

| File | Shows | Visible design decisions |
| --- | --- | --- |
| `Kasir/Daftar Order.pdf` | Parked/open orders list | `Daftar Order` is accessible from POS and lists unfinished orders. Implies quick return/resume pattern. |
| `Kasir/Faktur Penjualan.pdf` | Sales invoice flow | Shows form, customer/order details, payment term, save/print/pay. PRD defers sales invoices; use visual pattern only for future document screens. |
| `Kasir/Kupon.pdf` | Coupon application | Promotions are reached from a `Promosi` action and use modal/list selection + `Simpan`. PRD keeps quote contract but active promo behavior is Phase B. |
| `Kasir/Menambah Pelanggan.pdf` | Quick create/select customer | Customer picker supports add customer, sorting/grouping, then attaching customer to transaction. For v1 use compact quick-create. |
| `Kasir/Poin.pdf` | Loyalty points application | Points require customer first, then `Promosi`/`Poin`. PRD defers loyalty but preserves ordering rule. |
| `Kasir/promo .pdf` | Promo list/apply/edit discount | Promo selection/edit is modal/list-based with clear active promo detail. PRD requires server quote and rejected reasons. |
| `Kasir/search produk.pdf` | Product search | Product search is a top-level POS pattern. It should search name/SKU/barcode while preserving selected category. |
| `Kasir/Struk Digital.pdf` | Share receipt via WhatsApp/email/SMS | Shows `Bagikan Struk` popup with tabs and `Kirim`. PRD v1 allows OS share sheet/copy/print only, not direct WhatsApp/email/SMS. Use modal/tab visual pattern only. |

### 2.5 Kunci Layar

| File | Shows | Visible design decisions |
| --- | --- | --- |
| `Kunci Layar/Kunci Layar.pdf` | Lock screen and unlock by PIN | Lock is a full interruption and unlock uses PIN. Example uses confirmation with `Lanjutkan`; NojPOS must use named outcome and audit unlock per PRD. |

### 2.6 Laporan

| File | Shows | Visible design decisions |
| --- | --- | --- |
| `Laporan/10 Laporan Teratas.pdf` | Top reports | Reports use menu navigation, filters, ranked tables, and vertical scanning. |
| `Laporan/Komisi.pdf` | Commission report | Report layout favors table/list with filters. Commission calculation is out of PRD v1; visual density pattern remains useful. |
| `Laporan/Laporan Kas Kasir.pdf` | Cash drawer movement report | Cash movement uses list plus `+Transaksi`, form popup, type, amount, note, optional proof, and `Simpan`. PRD defers proof attachment. |
| `Laporan/Laporan Kasir.pdf` | Cashier shift report and close cashier from report | Shift report has activity list, action menu, ringkasan/detail, close shift action, PIN popup. |
| `Laporan/Laporan Produk Terjual .pdf` | Sold products report | Product report uses period filters, outlet filter, table/list, and `Cetak`. |
| `Laporan/Ringkasan Penjualan.pdf` | Sales summary | Summary has left metrics and right chart; filters include preset period, calendar range, and outlet. |
| `Laporan/Void.pdf` | Void report/detail/filter | Void report displays summary totals, table rows, date/order/cashier/authorizer, detail page, and filter by period/order type. |

### 2.7 Pengaturan

| File | Shows | Visible design decisions |
| --- | --- | --- |
| `Pengaturan/Biaya Pos.pdf` | Service charge, rounding, receipt-related toggles | Settings use grouped list navigation, sliders, checkboxes, percentage/money inputs, and `Simpan`. |
| `Pengaturan/Lainnya.pdf` | Update app, tour, delete account, orientation | Settings root includes utility rows, sometimes external route. Orientation shows tablet vs smartphone modes. |
| `Pengaturan/Langganan & Support.pdf` | Subscription status, history, feedback, backup code | Detail screens use tabs (`Riwayat`), list rows, form fields, and `Kirim`. Subscription is not core cashier workflow. |
| `Pengaturan/Local Server POS.pdf` | Local server master/client/test/sync | Use only as visual reference for device/network status. PRD explicitly excludes local server/offline sync. |
| `Pengaturan/Menambah Opsi Pembayaran Non Tunai.pdf` | EDC, bank transfer, QRIS setup/edit/delete | Payment methods use add screen, type cards/list, metadata fields, edit popup, and deactivate/delete confirmation. PRD says deactivation rather than history-destroying delete. |
| `Pengaturan/Pajak.pdf` | Tax mode settings | Uses switch, radio/option list, percent input, order-type checkboxes, and confirmation for risky setting. PRD provides calculation order. |
| `Pengaturan/Pengaturan Daftar Pesan POS.pdf` | Messages list/read/bulk actions | Message/notif lists support checkboxes, overflow menu, read/unread/delete, detail popup. |
| `Pengaturan/Pengaturan Info Outlet POS.pdf` | Outlet logo, receipt logo, profile edit | Profile uses upload actions, image constraints, popup edit for outlet fields, and `Simpan`. |
| `Pengaturan/Pengaturan Karyawan POS.pdf` | Staff create/edit/delete, camera attendance toggle | Staff form includes role/access/outlet/PIN/status active. PRD v1 excludes camera attendance evidence. |
| `Pengaturan/Pengaturan Kasir POS.pdf` | Default POS display, order type, opening cash, auto close, cash-out limits, custom amount, batch order, authorization, hotkeys, merge order | Cashier settings are grouped by operational concern. Uses switches, checklists, money/time fields, and save. PRD renames auto-close to shift auto-expiry. |
| `Pengaturan/Pengaturan Kata Sandi dan PIN POS.pdf` | Password and PIN change | Account settings use simple forms and `Simpan`. |
| `Pengaturan/Pengaturan Notifikasi POS.pdf` | Notifications and pending transactions | Notif list uses tab, bulk selection, overflow actions, detail popup. Pending transaction/manual sync behavior conflicts with PRD; do not implement queue UI except checkout retry state. |
| `Pengaturan/Pengaturan Promo POS.pdf` | Basic promo and promo list | Toggle + percent setting + promo detail popup. PRD keeps active promo UI behind quote contract. |
| `Pengaturan/Produk & Kategori,.pdf` | Product/category create/edit/delete | Product CRUD uses list -> popup, image upload, fields, switches, detail sections, checkboxes, `Simpan`, named delete confirmation. |
| `Pengaturan/Sinkronisasi POS.pdf` | Manual standard/total sync | Visual pattern: action sheet/modal, progress, PIN for risky action. PRD excludes manual sync/full offline. |
| `Pengaturan/Struk Kasir POS.pdf` | Receipt header/footer/note/tax visibility/print limit/queue number | Receipt settings use grouped detail pages, text inputs, switches, number inputs, and save. |
| `Pengaturan/Tampilan Struk.pdf` | Receipt, delivery note, label display and preview | Uses checkbox groups, preview action, modal preview, and save. Good source for preview pattern. |

---

## 3. Recurring design language extracted from examples

### 3.1 Color language

Rendered PDFs repeatedly show:

- **Primary teal/green** for main actions, active controls, selected tabs, success/online states. Dominant rendered samples cluster around teal values close to `#00C0A0`, `#10D0A0`, and darker teal `#006050`/`#106050`.
- **White and very light gray** backgrounds (`#FFFFFF`, `#F7F8FA`, `#F0F0F0`) with subtle separation.
- **Dark neutral text** with low-decoration utilitarian layouts.
- **Soft tinted backgrounds** for selected/active states, usually pale teal.
- **Warning/destructive states are under-specified in examples** except delete/void flows. Proposal: standardize warning amber and danger red tokens, always with text labels.

### 3.2 Layout language

- Most flows begin from a **hamburger/menu in the upper-left**, then a module list, then detail/list screens.
- Repeated screen structures: **header + action button**, **list/table**, **filter control**, **detail popup**, **confirm popup**, **print action**.
- POS examples imply a right-side action area for cart/payment/promotions and a visible search/catalog area. PRD fixes exact tablet regions.
- Settings examples use **directory -> grouped setting detail -> save**, not one huge form.
- Reports examples use **preset period + calendar + outlet/order-type filters**, summary metrics, charts, and dense table rows.

### 3.3 Component language

- Primary actions are short verbs: `Simpan`, `Bayar`, `Cetak`, `Kirim`, `Tambah`, `Tutup Kasir`.
- Secondary navigation uses text rows with chevrons or overflow menus.
- Confirmation often uses `Ya, Lanjutkan`; NojPOS must replace this with explicit affected-object wording.
- Popups/dialogs are common for edits and confirmations. Bottom/full-screen sheets should be used on mobile.
- Switches/sliders are used for on/off settings; checkboxes for order types and receipt elements.
- Forms prefer direct labels, short helper text, and a single final submit.

### 3.4 Typography and density

Examples are dense and utilitarian. For NojPOS:

- Use a modern sans family: **Inter** for web and **Roboto/System** for Flutter unless app font policy changes.
- Tablet text must be readable at arm's length. Minimum body text: 14sp; operational totals and primary amounts: 20-28sp.
- Admin web can be denser than Flutter, but never below 12px for table metadata or 14px for editable controls.

### 3.5 Motion

The examples do not define motion. Proposal:

- Use short, functional motion only: 120-180ms for button/sheet/dialog transitions, 200ms for cart sheet, no playful easing.
- Loading is explicit and blocks final submit only where needed.
- Avoid animated decorations in cashier checkout paths.

---

## 4. Canonical flow maps

### 4.1 Terminal entry and shift gate

```text
Open app
  -> Account login (email/password)
  -> Device enrollment check
  -> Outlet selection
  -> Staff selection / 6-digit PIN
  -> Shift gate
      -> no active shift: Buka Shift
      -> active shift: POS terminal
      -> pending_close: forced Tutup Shift
      -> store_closed: blocked until Buka Toko by owner/admin
```

Sources: `Login Logout POS.pdf`, `Tutup Kasir.pdf`, `Tutup Toko.pdf`, PRD §5.6, §5.7, §5.8.

### 4.2 POS sale and payment

```text
POS terminal
  -> search/category/product tap
  -> cart line edits/customer/order type
  -> Bayar
  -> payment screen/full-screen modal
  -> final quote/review
  -> final submit with idempotency key
      -> cash: paid success
      -> QRIS/EDC/transfer async: payment_pending
          -> webhook/poll confirms: paid success
          -> expiry/decline: payment_failed, retry payment or cancel
  -> receipt actions: Cetak / Bagikan via OS / Salin teks
  -> return to new empty cart or transaction detail
```

Sources: `search produk.pdf`, `Menambah Pelanggan.pdf`, `Kupon.pdf`, `Poin.pdf`, `promo .pdf`, `Struk Digital.pdf`, PRD §5.4.

### 4.3 Parked orders

```text
Draft cart
  -> Simpan Order
  -> server creates parked_order + revision
  -> Daftar Order
  -> Resume order
      -> lease lock granted: cart opens
      -> order_locked: show holder + remaining TTL + retry
      -> conflict_revision: refresh latest order
  -> Checkout or cancel under policy
```

Sources: `Daftar Order.pdf`, PRD §5.4.3, §5.4.5.

### 4.4 Close shift

```text
Open shift
  -> Tutup Shift / Tutup Kasir
  -> fresh server summary
  -> cashier fresh PIN
  -> enter actual cash (blank)
  -> variance preview
      -> variance zero: confirm
      -> variance non-zero: reason required
      -> variance beyond threshold: admin/owner PIN
  -> optional print choices
  -> final submit once
  -> immutable close report
  -> staff PIN or shift gate, never active POS
```

Sources: `Tutup Kasir.pdf`, `Laporan Kasir.pdf`, PRD §5.6.3.

### 4.5 Inventory documents

```text
Inventory list/document tab
  -> Tambah [document]
  -> outlet/date/number/note fields
  -> add products via product picker
  -> line qty/cost/reason fields
  -> review
  -> confirm once
  -> immutable detail with status
  -> optional Cetak
```

Document-specific branches:

```text
Purchase: draft input -> finalize purchase -> stock movement purchase
Stock count: system count -> actual count -> difference reason -> adjustment movements
Waste: reason + product qty -> waste movement -> completed/void rules
Transfer: request -> dispatch -> in-transit -> receive partial/full -> resolve discrepancy -> complete
```

Sources: `Faktur Pembelian POS.pdf`, `Stok Opname.pdf`, `Stok Terbuang.pdf`, `Permintaan Stok.pdf`, `Terima Mutasi Stok.pdf`, PRD §8.

### 4.6 Attendance

```text
Absensi
  -> select staff
  -> PIN keypad
  -> Clock In or Clock Out
  -> success state with staff/action/time
  -> return to attendance ready screen + today's recent list
```

Sources: `Absensi.pdf`, PRD §9.

### 4.7 Screen lock

```text
POS top bar: Kunci Layar
  -> named confirmation: Kunci terminal ini
  -> locked full-screen surface
  -> staff PIN
      -> same staff: restore cart + shift
      -> different staff: choose cover/audited unlock or handover
  -> unlock audit event
```

Sources: `Kunci Layar.pdf`, PRD §10.

### 4.8 Settings

```text
Pengaturan root
  -> selected outlet context visible
  -> grouped list item
  -> detail page
  -> edit controls
  -> dirty state appears: Batal perubahan + Simpan perubahan
  -> final submit disabled while saving
  -> success returns to detail/list with updated timestamp/status
```

Sources: all `Pengaturan/*.pdf`, PRD §6.

### 4.9 Reports

```text
Laporan root
  -> choose report tab/type
  -> filter rail: period preset, custom date, outlet, report-specific filters
  -> summary metrics/charts where relevant
  -> table/list rows
  -> row detail panel/page
  -> print/export only when supported and authorized
```

Sources: all `Laporan/*.pdf`, PRD §7.

---

## 5. Design tokens

Use these names in Flutter now and in a future admin web only after that surface is explicitly approved. The current repo has no active Next.js/admin web app. The values are NojPOS tokens derived from the extracted teal/neutral example language, with explicit proposals for missing semantic colors.

### 5.1 Color tokens

| Token | Value | Source | Usage |
| --- | ---: | --- | --- |
| `color.brand.50` | `#E6F8F4` | Proposed from example pale teal | Selected row, soft active background |
| `color.brand.100` | `#C9F1EA` | Proposed from example pale teal | Active tab background, icon badge |
| `color.brand.500` | `#00B894` | Derived from rendered teal cluster | Primary button, active switch, selected category |
| `color.brand.600` | `#009E82` | Derived/proposed | Button hover/pressed, active border |
| `color.brand.700` | `#007C66` | Derived from dark teal cluster | Strong icon, active text, success heading |
| `color.brand.900` | `#064E3B` | Derived from dark teal cluster | High contrast on pale teal |
| `color.surface.canvas` | `#F7F8FA` | Derived from light gray examples | App/page background |
| `color.surface.base` | `#FFFFFF` | Examples | Cards, dialogs, table body |
| `color.surface.subtle` | `#F1F5F4` | Derived/proposed | Table header, disabled row background |
| `color.surface.raised` | `#FFFFFF` | Examples | Dialog and POS order panel |
| `color.text.primary` | `#111827` | Proposed neutral | Main labels and totals |
| `color.text.secondary` | `#4B5563` | Proposed neutral | Metadata and descriptions |
| `color.text.muted` | `#6B7280` | Proposed neutral | Placeholder and timestamps |
| `color.text.inverse` | `#FFFFFF` | Examples/proposed | Text on solid brand/danger buttons |
| `color.border.subtle` | `#E5E7EB` | Proposed | Table row lines, input border |
| `color.border.strong` | `#D1D5DB` | Proposed | Focusable/selected boundaries when not brand |
| `color.success.50` | `#ECFDF5` | Proposed semantic | Success state background |
| `color.success.600` | `#059669` | Proposed semantic | Success text/icon |
| `color.warning.50` | `#FFFBEB` | Proposed semantic | Warning banner background |
| `color.warning.600` | `#D97706` | Proposed semantic | Warning text/icon |
| `color.danger.50` | `#FEF2F2` | Proposed semantic | Destructive confirmation background |
| `color.danger.600` | `#DC2626` | Proposed semantic | Destructive action/icon |
| `color.info.50` | `#EFF6FF` | Proposed semantic | Informational banner background |
| `color.info.600` | `#2563EB` | Proposed semantic | Info icon/text |
| `color.connection.online` | `#059669` | PRD + proposed | Online connection status |
| `color.connection.degraded` | `#D97706` | PRD + proposed | Degraded status |
| `color.connection.offline` | `#DC2626` | PRD + proposed | Offline status and blocking state |

Contrast rules:

- Solid primary buttons use `brand.600` or darker with white text.
- Do not put white text on `brand.500` for small text below 14px; use `brand.600`/`brand.700`.
- Semantic badges use pale background + dark text for table/list contexts.
- Offline/danger states must include label and recovery action, not only red.

### 5.2 Typography tokens

| Token | Flutter | Web | Weight | Line height | Usage |
| --- | ---: | ---: | ---: | ---: | --- |
| `text.display.sm` | 28sp | 28px | 700 | 36 | Payment total, close-shift variance |
| `text.heading.lg` | 22sp | 22px | 700 | 30 | Screen titles on tablet/mobile |
| `text.heading.md` | 18sp | 18px | 700 | 26 | Section titles, dialog titles |
| `text.heading.sm` | 16sp | 16px | 600 | 24 | Card/list headings |
| `text.body.lg` | 16sp | 16px | 400 | 24 | Tablet readable content |
| `text.body.md` | 14sp | 14px | 400 | 20 | Default body/table text |
| `text.body.sm` | 13sp | 13px | 400 | 18 | Metadata, helper text |
| `text.caption` | 12sp | 12px | 500 | 16 | Badges, compact labels |
| `text.button` | 14sp | 14px | 600 | 20 | Buttons and tabs |
| `text.numeric.lg` | 24sp | 24px | 700 | 32 | Rupiah amount/total |
| `text.numeric.md` | 16sp | 16px | 600 | 24 | Line totals and report amounts |

Font families:

- Flutter: `Roboto`, system fallback. Use tabular figures for totals if available.
- Next.js: `Inter`, system fallback. Enable `font-variant-numeric: tabular-nums` on amounts, tables, reports, and receipt preview.

### 5.3 Spacing and sizing tokens

| Token | Value | Usage |
| --- | ---: | --- |
| `space.0` | 0 | Reset |
| `space.1` | 4 | Tight icon/text gap |
| `space.2` | 8 | Small internal padding |
| `space.3` | 12 | Form field gap, row gap |
| `space.4` | 16 | Default screen padding/mobile section gap |
| `space.5` | 20 | Tablet form section gap |
| `space.6` | 24 | Dialog padding, desktop page padding |
| `space.8` | 32 | Major section separation |
| `space.10` | 40 | Large hero/empty state gap |
| `space.12` | 48 | Modal vertical rhythm |

Sizing:

| Token | Value | Usage |
| --- | ---: | --- |
| `size.touch.min` | 44 | Absolute minimum touch target |
| `size.touch.cashier` | 48 | Default cashier control target |
| `size.touch.keypad` | 64 | Numeric/PIN keypad button minimum |
| `size.topbar.tablet` | 64 | PRD POS tablet top bar |
| `size.category.collapsed` | 72 | PRD collapsed category rail |
| `size.category.expanded` | 220 | PRD expanded category rail |
| `size.orderPanel.min` | 400 | PRD tablet order panel min |
| `size.orderPanel.max` | 440 | PRD tablet order panel max |
| `size.admin.row` | 44 | Dense admin table row |
| `size.admin.header` | 56 | Admin page header compact height |

### 5.4 Radius, border, elevation

| Token | Value | Usage |
| --- | ---: | --- |
| `radius.none` | 0 | Table grid merge/flush areas |
| `radius.xs` | 4 | Badges, compact chips |
| `radius.sm` | 6 | Inputs, table controls |
| `radius.md` | 8 | Buttons, dialogs, cards. Admin maximum radius. |
| `radius.lg` | 12 | Flutter POS product cards/bottom sheets only |
| `border.subtle` | 1px `border.subtle` | Admin/table/input separators |
| `shadow.none` | none | Admin nested surfaces |
| `shadow.sm` | subtle 0 1 2 | Dialog/product card on Flutter only |
| `shadow.md` | subtle 0 8 24 | Modal/sheet overlay only |

Admin rule: radius must be <= 8px and avoid nested floating cards. Use borders and table sections instead.

---

## 6. Platform token consumption

### 6.1 Flutter ThemeData mapping

- `ColorScheme.primary` = `brand.600`
- `ColorScheme.secondary` = `brand.500`
- `ColorScheme.surface` = `surface.base`
- `Scaffold.backgroundColor` = `surface.canvas`
- `ColorScheme.error` = `danger.600`
- `TextTheme` maps to typography tokens above.
- `ElevatedButtonTheme` uses primary solid button with 48px minimum height on cashier surfaces.
- `OutlinedButtonTheme` uses border `border.subtle`, radius 8, and no heavy shadow.
- POS-specific components should read from app design tokens, not inline colors.

### 6.2 Future admin web CSS variables / Tailwind mapping

This is a future design mapping only; do not create `apps/web` or a Next.js app from this document during P0 hardening. When admin web is approved, expose tokens as CSS variables under `:root`, for example:

```css
:root {
  --color-brand-600: #009E82;
  --color-surface-canvas: #F7F8FA;
  --color-text-primary: #111827;
  --radius-md: 8px;
  --space-4: 16px;
}
```

Tailwind config should map variables to semantic names (`brand`, `surface`, `text`, `border`, `success`, `warning`, `danger`, `info`) rather than hard-coded hex per component.

---

## 7. Component library specifications

### 7.1 Buttons

Anatomy: leading icon optional, label, trailing icon/loading optional. Label must be a verb or verb+noun.

Variants:

| Variant | Visual | Use |
| --- | --- | --- |
| Primary | Solid `brand.600`, white text | Main safe action: `Bayar`, `Simpan`, `Buka Shift` |
| Secondary | White/subtle surface, neutral border | Navigation or secondary action: `Batal`, `Lihat Tampilan` |
| Tertiary/ghost | Transparent, brand/neutral text | Low-emphasis row actions |
| Destructive | Solid or outline danger | `Hapus Produk`, `Void Transaksi`, `Batalkan Permintaan Stok` |
| Warning | Amber outline/solid | Risky but not destructive: variance approval, re-quote |

Sizes:

- Cashier tablet: min height 48, horizontal padding 16-20.
- Key checkout/payment CTA: min height 56.
- Admin dense: min height 36-40.
- Mobile bottom CTA: height 52-56, full width.

States:

- Default, hover (web), focus-visible, pressed, disabled, loading.
- Loading final submit shows spinner + label (`Menyimpan...`, `Memproses pembayaran...`) and disables repeat tap.
- Recoverable error restores enabled state and preserves form/cart.

Rules:

- Destructive buttons must name the object/result: `Hapus Produk Es Kopi`, `Void Transaksi TRX-SDM-202606-000123`, `Tutup Shift SHF-SDM-202606-000004`.
- Never use bare `Lanjutkan` for destructive actions.

Sources: repeated `Simpan`, `Tambah`, `Cetak`, `Tutup`, `Hapus`, `Bayar`, and `Kirim` patterns in Kasir/Pengaturan/Inventori PDFs; PRD §5.3.

### 7.2 Inputs

Anatomy: label, required mark, input area, optional prefix/suffix, helper/error text.

Variants:

- Text input
- Number input
- Money input
- Percent input
- Date/time picker
- Select/dropdown
- Multi-select checkbox list
- Switch/toggle
- Textarea/note
- Search field

States: default, focus, filled, disabled, readonly, validation error, loading/dependent disabled.

Rules:

- Field-level errors appear under the field and preserve input.
- Dependent disabled controls include explanation (`Aktifkan Pajak terlebih dahulu`).
- Money and quantity inputs reject invalid characters before submit when possible, but server validation remains source of truth.

### 7.3 Money input

Anatomy: label, `Rp` prefix, formatted integer display, raw integer value in state, helper text.

Rules:

- Display: `Rp12.500`, no decimals.
- Storage/transmission: integer rupiah only.
- Blank is distinct from zero where PRD requires blank (actual cash in close shift must start blank).
- For shift close actual cash, never prefill expected cash.

Sources: PRD §5.3, §5.6.3; settings/cash reports examples.

### 7.4 Numeric/PIN keypad

Anatomy: context title, masked PIN display, 3x4 keypad, delete/backspace, submit/auto-submit behavior, lockout helper text.

Sizes:

- Keypad button min 64px on tablet, 56px mobile.
- PIN dots 12-14px, spacing 8.

Rules:

- PIN is masked and cleared after success, failure, lockout, background transition, or dispose.
- Error message is generic (`PIN tidak valid`) and does not reveal whether staff exists.
- Lockout state shows remaining time and safe return action.
- PIN entry must be reachable from hardware keyboard on web/admin if used there.

Sources: `Login Logout POS.pdf`, `Tutup Kasir.pdf`, `Absensi.pdf`, `Kunci Layar.pdf`, PRD §6.7, §10.

### 7.5 Search field

Anatomy: leading search icon, placeholder, input, clear button, optional scan/barcode action.

POS behavior:

- Search product name, SKU, barcode.
- Preserve selected category while filtering.
- Empty result uses helpful state (`Produk tidak ditemukan`) and safe actions (`Bersihkan pencarian`, `Tambah produk` only for authorized admin contexts).

Admin behavior:

- Search pairs with filter rail/table header.
- Debounce 250-400ms.

Sources: `search produk.pdf`, PRD §5.4.1.

### 7.6 List and table rows

Anatomy: leading optional checkbox/icon/status, primary text, metadata, amount/status, trailing action/chevron/overflow.

Variants:

- POS list row (larger touch, 56-64 height)
- Admin dense table row (44 height)
- Document row (status + number + outlet + actor + date)
- Report row (numeric aligned right)
- Bulk-select row

States: default, hover (web), selected, active, disabled, loading placeholder, conflict/stale.

Rules:

- Numeric amounts align right and use tabular figures.
- Document rows always show status text and document number.
- Outlet-scoped rows show outlet unless the entire screen has a single outlet header and no cross-outlet rows.

Sources: reports, notifications/messages, inventory documents, staff/product settings examples; PRD §5.2, §7, §8.

### 7.7 Product card

Anatomy: image/placeholder, product name, price, stock/status, optional SKU/category, tap target.

Tablet POS:

- Stable dimensions in grid.
- Tap adds one unit.
- Long press or info action opens detail/options.
- Unavailable state dims card, disables add, and explains reason (`Stok habis`, `Tidak tersedia di outlet ini`).

Mobile:

- Grid or list depending width; never share screen with full cart and keypad.

Sources: `search produk.pdf`, product/category examples, PRD §5.4.1.

### 7.8 Cart line

Anatomy: product name, modifiers/options, qty stepper, unit price, line subtotal, remove action, warnings.

Rules:

- Quantity controls have 44+ touch targets.
- Remove single line is immediate with undo by default, per PRD decision.
- If price/stock/promo changes server-side, show stale quote/conflict state and recovery.
- Cart cannot start payment with zero valid lines.

Sources: cashier examples; PRD §5.4.3, §5.4.7.

### 7.9 Order panel

Anatomy: order header (order type, customer, status), cart lines, promo/discount summary, totals, primary CTA, secondary actions.

Tablet:

- Width 400-440px.
- Sticky total and `Bayar` CTA at bottom.
- Shows customer as `Tanpa Pelanggan` by default.

Mobile:

- Compact bottom bar shows item count + total.
- Full-screen cart sheet for order review.

Sources: POS examples, PRD §5.4.1-5.4.4.

### 7.10 Status chip/badge

Anatomy: optional dot/icon, label, optional secondary detail.

Variants:

- Connection: `Online`, `Gangguan`, `Offline`
- Order: `Draft`, `Tersimpan`, `Menunggu Pembayaran`, `Lunas`, `Void`, `Refund Sebagian`
- Shift: `Belum Buka`, `Aktif`, `Perlu Ditutup`, `Tutup`
- Inventory document: `Draf`, `Diajukan`, `Dikirim`, `Diterima Sebagian`, `Selesai`, `Dibatalkan`
- Stock: `Aman`, `Rendah`, `Negatif`

Rules:

- Always label with text.
- Use pale background + strong text in tables.
- Use solid/dot only as supplement.

Sources: document/report examples; PRD §5.3, §5.4.3, §8.2, §8.6.

### 7.11 Dialog and confirmation

Anatomy: title, body with object/result, impact summary, optional field(s), secondary cancel, named primary action.

Types:

- Info/detail popup (`Detail Notifikasi`, `Preview Struk`)
- Edit popup (`Ubah Karyawan`, `Ubah Produk`)
- Confirmation (`Tutup Shift`, `Hapus Produk`, `Void Transaksi`)
- Authorization PIN dialog
- Conflict dialog

Rules:

- Destructive confirmation must include object and result.
- For money/stock impact, include amount/qty and affected document.
- Never use `Ya, Lanjutkan` alone; use `Ya, hapus produk`, `Ya, void transaksi`, etc.
- On mobile, prefer full-screen dialog or bottom sheet for complex forms.

Sources: examples repeatedly use popups for read/edit/confirm; PRD §5.3.

### 7.12 Bottom sheet / full-screen sheet

Use for mobile cart, payment, product options, filter panels, and previews.

Rules:

- Mobile payment is full screen, not nested in cart.
- Sheet must have clear title, close/back, sticky final CTA if form is long.
- Drag-to-dismiss disabled during final submit.

Sources: mobile requirement from PRD §5.4.2; popup patterns from examples.

### 7.13 Stepper

Use for operational documents and high-risk flows.

Default steps:

1. Informasi
2. Item/Detail
3. Review
4. Konfirmasi

Rules:

- Allow back before final submit.
- Show completed step summary.
- Final submit only appears in review/confirm.
- Server response determines final document status.

Sources: inventory examples; PRD §5.1, §8.

### 7.14 Toast/snackbar

Use for short feedback only.

Rules:

- Success toast may confirm save but UI must also show durable status/detail/list.
- Error toast may supplement field/banner error but must not be the only recovery information.
- Use max one primary action in snackbar (`Lihat detail`, `Coba lagi`) and keep duration long enough for reading.

Sources: PRD §5.1.

---

## 8. Canonical UI state machine treatment

Every API-backed surface implements:

```text
idle -> loading -> success | validationError | forbidden | notFound | conflict | recoverableNetworkError
```

| State | UI treatment | Recovery action |
| --- | --- | --- |
| `idle` | Ready screen with current context and no blocking spinner. | User acts normally. |
| `loading` | Skeleton for lists; inline spinner for sections; full blocker only for final submit or payment. | Prevent duplicate submit. |
| `success` | Detail/list state with updated document number/status visible. Optional snackbar. | Continue workflow, print/share where relevant. |
| `validationError` | Preserve input; field-level errors; summary banner for form-level issue. | Edit fields and submit again. |
| `forbidden` | Neutral forbidden state; do not expose tenant/private data. | Request authorization PIN if PRD allows, or return. |
| `notFound` | Empty safe state: document not found or already inaccessible. | Back to list, refresh. |
| `conflict_revision` | Banner/dialog: order changed elsewhere, show latest revision required. | Refresh order; re-apply changes manually if safe. |
| `order_locked` | Show holder identity and remaining TTL. | Wait/retry after expiry; privileged release only if policy exists. |
| `quote_stale` | Explain price/promo/stock changed. | Auto re-quote, show changed totals, require review before submit. |
| `idempotency_mismatch` | Stop flow; warn that request changed under same key. | Refresh/review from server; do not retry blindly. |
| `state_conflict` | Show invalid transition or blocking entity list (e.g., open shifts). | Resolve listed blockers. |
| `recoverableNetworkError` | Preserve local draft; no queue except checkout retry buffer. | Retry when online; for checkout, resend same key if buffered. |
| `payment_pending` | Dedicated pending screen with reference, expiry, polling indicator. | Wait, cancel/expire per policy, or retry payment after failure. |
| `unauthenticated` | Session expired state. | Return to login. |
| `device_not_enrolled` | Blocking device state with device identity. | Enroll/ask owner/admin. |

---

## 9. Layout and screen patterns

### 9.1 POS tablet layout

Use PRD's four stable regions.

```text
┌──────────────────────────────────────────────────────────────────────────────┐
│ Top bar 64px: menu, outlet, cashier, shift, connection, mode, lock, orders  │
├──────────┬───────────────────────────────────────────────┬─────────────────┤
│ Category │ Catalog/search/product grid                   │ Order panel     │
│ rail     │                                               │ 400-440px       │
│ 72/220px │                                               │ cart + totals   │
└──────────┴───────────────────────────────────────────────┴─────────────────┘
```

Top bar contains:

- Menu
- Outlet name
- Active cashier
- Shift status
- Genuine connection state: `Online`, `Gangguan`, `Offline`
- Order mode/type
- Lock action
- `Daftar Order`

Responsive rules:

- >= 1024px landscape: full four-region layout.
- 900-1023px: keep order panel 400px, category rail collapsed 72px.
- < 900px: switch to mobile pattern.
- Never let catalog and order panel overlap.
- Offline state blocks new sale; in-flight checkout uses retry buffer.

### 9.2 POS mobile layout

```text
Catalog full screen
  -> top context bar (outlet/shift/connection)
  -> search/category/product list
  -> compact bottom bar: item count + total + Buka Keranjang
  -> full-screen cart sheet
  -> full-screen payment
  -> full-screen success/receipt
```

Rules:

- Never cram catalog + cart + keypad together.
- Cart sheet must preserve context and allow back to catalog.
- Payment screen is isolated with immutable summary and final CTA.

### 9.3 Admin web layout

Default admin web structure:

```text
Persistent app shell
  -> page header: title + outlet context + primary action
  -> filter rail / filter bar
  -> table/list
  -> right detail panel or route detail page
```

Rules:

- Dense tables by default.
- Cards only for small summary metrics and dialogs.
- Radius <= 8px.
- Subtle borders over heavy shadows.
- No nested floating cards.
- Selected outlet appears in every outlet-scoped header and filter state.

---

## 10. Key screen blueprints

### 10.1 Login, outlet, PIN

Blueprint:

- App/logo area
- Email/password login
- Device enrollment/identity state after first login
- Outlet selection list with assigned outlets
- Staff selection + PIN keypad
- Errors: unauthenticated, forbidden outlet, device revoked, PIN lockout

Sources: `Login Logout POS.pdf`, PRD §5.6.1, §5.8.

### 10.2 Shift gate and open shift

Blueprint:

- Header: outlet, device, current time, connection
- State card: `Belum ada shift aktif`, `Shift perlu ditutup`, or `Toko tutup`
- Open shift form: cashier, opening cash money input, note optional
- Primary CTA: `Buka Shift`
- Success detail: shift number, opening time, opening cash, cashier, device

Sources: `Tutup Kasir.pdf` as close counterpart; PRD §5.6.1.

### 10.3 Close shift

Blueprint:

- Fresh server summary panel
- Opening cash, cash sales, cash in/out, refund/void reversals, expected cash
- Fresh PIN input
- Actual cash money input blank by default
- Variance preview with warning/danger based on threshold
- Variance reason field when non-zero
- Approver PIN section when threshold exceeded
- Print options only if printer configured
- Named final CTA: `Tutup Shift SHF-...`
- Success: immutable close report

Sources: `Tutup Kasir.pdf`, `Laporan Kasir.pdf`, PRD §5.6.3.

### 10.4 POS terminal

Blueprint:

- 64px top bar with genuine connection state
- Category rail
- Search + product grid/list
- Order panel with customer/order type/cart/totals/CTA
- Conflict and offline banners in the top of work area
- `Daftar Order` access

Sources: `search produk.pdf`, `Daftar Order.pdf`, customer/promo examples, PRD §5.4.

### 10.5 Payment, pending, success

Payment blueprint:

- Immutable order summary
- Payment method list: Tunai, EDC, Transfer Bank, QRIS when configured
- Amount received/change for cash
- Reference field for non-cash if required
- Final quote section with totals
- Primary CTA: `Bayar Rp...`

Pending blueprint:

- Status `Menunggu pembayaran`
- Payment reference
- Expiry countdown/time
- Polling indicator
- Actions allowed by method policy

Success blueprint:

- Transaction number
- Customer/order type
- Payment lines, change, applied discounts
- Receipt actions: `Cetak`, `Bagikan`, `Salin teks`
- CTA: `Transaksi baru`

Sources: `Faktur Penjualan.pdf`, `Struk Digital.pdf` visual patterns; PRD §5.4.7-§5.4.9.

### 10.6 Settings root and detail

Root blueprint:

- Header: `Pengaturan`, selected outlet control, scope helper line
- Compact grouped list:
  - Profil Outlet
  - Transaksi & POS
  - Pembayaran
  - Struk & Dokumen
  - Keamanan & Otorisasi
  - Perangkat
  - Notifikasi
  - Akun Saya
- Each row: icon, title, one-line description, warning/completion state, chevron

Detail blueprint:

- Breadcrumb
- Title
- Outlet/business scope label
- Form sections
- Sticky save bar only when dirty: `Batal perubahan`, `Simpan perubahan`

Sources: all `Pengaturan/*.pdf`, PRD §6.

### 10.7 Reports

Blueprint:

- Header with selected outlet/scope
- Filter rail/bar: period preset, custom range, outlet, report-specific filters
- Metrics cards only at top for summary numbers
- Chart where useful
- Dense table
- Row detail panel/page
- Print/export only when supported

Sources: all `Laporan/*.pdf`, PRD §7.

### 10.8 Inventory documents

Blueprint:

- Document list with status chips and filters
- Add document button
- Stepper form
- Product picker
- Review table with line totals/differences/reasons
- Confirmation dialog naming document and result
- Immutable detail after finalize
- Print action if supported

Sources: all `Inventori/*.pdf`, PRD §8.

### 10.9 Attendance

Blueprint:

- Outlet + current time header
- Staff selector
- PIN keypad
- Contextual action: `Absen Masuk` or `Absen Pulang`
- Recent attendance list for today
- Success state with staff/action/time

Sources: `Absensi.pdf`, PRD §9.

### 10.10 Screen lock

Blueprint:

- Minimal full-screen lock surface
- Outlet name + current time
- `Terminal terkunci`
- PIN keypad
- Different-staff unlock branch: `Lanjut sebagai kasir saat ini` vs `Serah terima kasir`
- No prices/customer/cart/report details visible

Sources: `Kunci Layar.pdf`, PRD §10.

---

## 11. Cross-cutting UX rules

### 11.1 Bahasa Indonesia glossary

Use these standard labels consistently.

| Concept | Label |
| --- | --- |
| Pay | `Bayar` |
| Save | `Simpan` |
| Save changes | `Simpan perubahan` |
| Cancel changes | `Batal perubahan` |
| Add | `Tambah` |
| Print | `Cetak` |
| Close dialog | `Tutup` |
| Customer absent | `Tanpa Pelanggan` |
| Open shift | `Buka Shift` |
| Close shift | `Tutup Shift` |
| Cash in | `Kas Masuk` |
| Cash out | `Kas Keluar` |
| Close store | `Tutup Toko` |
| Open store | `Buka Toko` |
| Lock screen | `Kunci Layar` |
| Unlock | `Buka Kunci` |
| Parked orders | `Daftar Order` |
| Product search | `Cari produk, SKU, atau barcode` |
| Payment pending | `Menunggu pembayaran` |
| Retry | `Coba lagi` |
| Refresh | `Muat ulang` |
| Void transaction | `Void Transaksi` |
| Refund | `Refund` or `Pengembalian Dana` (choose one per product copy policy; proposal: show `Refund` in reports, `Pengembalian Dana` in helper text) |

### 11.2 Rupiah and dates

- Display rupiah as `Rp12.500`, no decimals.
- Align amounts right in tables and use tabular figures.
- Operational dates render in outlet timezone.
- Reports spanning outlets must show reference timezone in header.

### 11.3 Accessibility

- Touch target: minimum 44px, cashier default 48px, keypad 64px.
- Focus order follows visual order: header -> filters -> main list/form -> primary action -> secondary actions.
- All controls need visible focus ring on web and keyboard/hardware input support where relevant.
- Status must use text + color.
- Error messages identify what to fix and preserve user input.
- Destructive actions need confirmation and object name.
- Do not hide critical totals behind hover.

### 11.4 Destructive and high-risk actions

Always include:

1. Affected object/document number/name.
2. Result of action.
3. Money/stock impact when relevant.
4. Required PIN/approver if relevant.
5. Named final CTA.

Examples:

- Good: `Void Transaksi TRX-SDM-202606-000123`.
- Good: `Hapus Produk Es Kopi Susu`.
- Good: `Tutup Shift SHF-SDM-202606-000004`.
- Bad: `Lanjutkan`.

### 11.5 Form and submit rules

- Final submit disabled while in flight.
- Recoverable network error restores submit and preserves form/cart.
- Successful write routes to detail/list with new status visible.
- Validation errors render next to fields and in a summary when needed.
- Dirty settings show sticky save bar only when changes exist.

### 11.6 Genuine connection state

Top-bar connection labels:

| State | Label | Color | Behavior |
| --- | --- | --- | --- |
| `online` | `Online` | success | Normal operations allowed. |
| `degraded` | `Gangguan` | warning | Reads/writes may still work; show subtle banner if API calls fail. |
| `offline` | `Offline` | danger | New sales blocked; final checkout retry buffer may resolve one in-flight checkout. |

Heartbeat rules come from PRD §5.9: `/health` every 20s; online <=30s since success; degraded 30-90s; offline >90s.

---

## 12. Source mapping appendix

### Decisions directly derived from examples

- Teal primary action and active states: all rendered PDFs, strongest in `Kasir/*`, `Pengaturan/*`, `Laporan/*`, `Inventori/*`.
- Menu/module navigation pattern: almost every PDF starts with menu upper-left -> module.
- Popup/dialog usage for read/edit/confirm: `Pengaturan Daftar Pesan POS.pdf`, `Pengaturan Karyawan POS.pdf`, `Produk & Kategori,.pdf`, `Tampilan Struk.pdf`, `Tutup Kasir.pdf`.
- Dense report/list/table pattern: all `Laporan/*.pdf`, inventory document PDFs, message/notification PDFs.
- Settings grouped directory/detail/save pattern: all `Pengaturan/*.pdf`.
- PIN keypad/auth surfaces: `Login Logout POS.pdf`, `Tutup Kasir.pdf`, `Absensi.pdf`, `Kunci Layar.pdf`.
- Product search/customer/promo entry points: `Kasir/search produk.pdf`, `Menambah Pelanggan.pdf`, `Kupon.pdf`, `Poin.pdf`, `promo .pdf`.
- Receipt preview/share/print modal patterns: `Struk Digital.pdf`, `Struk Kasir POS.pdf`, `Tampilan Struk.pdf`.

### Decisions from `docs/ARCHITECTURE.md`

- Current buildable apps are `apps/cashier` Flutter and `apps/backend` Laravel API; no current `apps/web`/Next.js app exists.
- Flutter implementation mapping is `UI -> Riverpod provider/notifier -> repository interface -> ApiClient`.
- Routes and features currently include splash, login, outlet, PIN, sync, shift, POS, operations, payment, and success.
- Drift/general offline database is absent; only checkout outbox/retry is present for write recovery.

### Decisions from `docs/ARCHITECTURE.md`

- Current buildable apps are `apps/cashier` Flutter and `apps/backend` Laravel API; no current `apps/web`/Next.js app exists.
- Flutter implementation mapping is `UI -> Riverpod provider/notifier -> repository interface -> ApiClient`.
- Routes and features currently include splash, login, outlet, PIN, sync, shift, POS, operations, payment, and success.
- Drift/general offline database is absent; only checkout outbox/retry is present for write recovery.

### Decisions from PRD v2.0

- POS tablet 4-region layout and exact region sizes: PRD §5.4.1.
- Mobile catalog/cart/payment separation: PRD §5.4.2.
- Online-first and only one checkout retry buffer: PRD §2.1.
- Common patterns: rupiah format, outlet timezone, text+color statuses, destructive naming, final submit behavior: PRD §5.3.
- Shift open/close, auto-expiry, handover: PRD §5.6.
- Store open/close: PRD §5.7.
- Device enrollment: PRD §5.8.
- Genuine connection state: PRD §5.9.
- Settings IA: PRD §6.1-§6.7.
- Reports definitions and scope: PRD §7.
- Inventory IA and document states: PRD §8.
- Attendance rules: PRD §9.
- Screen lock behavior: PRD §10.
- UI state machine: PRD §11 and §14.
- API/client layering: PRD §11.

### Proposals to fill gaps

- Exact NojPOS token values around extracted teal family, especially `brand.500 #00B894` and `brand.600 #009E82`.
- Semantic warning/danger/info palettes because examples under-specify them.
- Typography family and scale: Inter for web, Roboto/system for Flutter.
- Motion durations 120-200ms.
- Admin row/header sizing and CSS variable names.
- Explicit component state tables.
- Glossary standardization where examples use mixed labels (`Tutup Kasir`) and PRD prefers `Tutup Shift`. Proposal: product copy uses `Tutup Shift`; helper text may say `tutup kasir` when explaining cashier familiarity.

### Example behavior intentionally not adopted because PRD excludes/defers it

- Local server/master/client/offline sync from `Local Server POS.pdf`.
- Manual full sync and pending transaction batch sync from `Sinkronisasi POS.pdf` and `Pengaturan Notifikasi POS.pdf`.
- Direct WhatsApp/email/SMS receipt delivery from `Struk Digital.pdf`; v1 uses print, OS share sheet, copy text.
- Camera attendance from `Pengaturan Karyawan POS.pdf`.
- Active sales invoice flow from `Faktur Penjualan.pdf`; PRD defers invoices to Phase D.
- Active coupon/points/promo behavior from `Kupon.pdf`, `Poin.pdf`, `promo .pdf`; terminal UI is enabled only after server quote contract ships.

---

## 13. How to design a new screen checklist

Before creating a new screen:

1. Identify domain and source of truth: POS, shift, settings, reports, inventory, attendance, lock.
2. Confirm outlet scope. Put selected outlet in header/filter if outlet-bound.
3. Choose the right pattern:
   - POS tablet four-region surface
   - POS mobile single-task/full-screen sheet
   - Admin dense table/detail panel
   - Settings directory/detail/sticky save
   - Operational stepper for money/stock documents
4. Map the UI state machine: idle, loading, success, validationError, forbidden, notFound, conflict, recoverableNetworkError.
5. Define document/status labels with text + color.
6. Use tokens only; no inline hex, spacing, radius, or typography.
7. Use Indonesian labels from glossary.
8. Show rupiah as `Rp12.500` and send integer rupiah.
9. For final submit: disable while in-flight, preserve input on recoverable error, route to durable success state.
10. For destructive/high-risk action: name object + result + impact + approver/PIN + named CTA.
11. For Flutter: keep UI -> provider -> repository -> network; no direct HTTP in widgets.
12. For future admin web after explicit approval: prefer table/filter/detail; avoid nested floating cards and radius > 8px.
13. If examples and PRD conflict, keep example visual pattern but follow PRD behavior.
