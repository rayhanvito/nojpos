# NojPOS — POS Flutter & Laravel API PRD

**Version:** 2.0 (revised, implementation-ready)

**Status:** Locked for implementation — product-owner decisions confirmed 2026-06-20

**Scope:** Flutter Cashier App + Laravel API

**Audience:** Product owner, Flutter engineer, Laravel engineer, QA

**Authoring lens:** Product management · System analysis · IT architecture · Senior engineering

> This revision keeps v1.0's intent and structure but resolves every contradiction, fills the undefined domains, and converts deferred "policy" placeholders into concrete defaults marked as **[DECISION]** so engineering can build without guessing. Open items that genuinely need a business owner are consolidated in Section 18.
> 

---

## 0. What changed from v1.0 (resolution log)

| # | v1.0 problem | Resolution in v2.0 |
| --- | --- | --- |
| A1 | "No offline queue" vs "checkout outbox exists" | §2.1 defines a **single-attempt checkout retry buffer** (not a queue); offline behavior bounded explicitly |
| A2 | Auto shift-close vs mandatory physical cash | §5.6 auto-close is now a **`pending_close` expiry**, never auto-finalizes cash |
| A3 | "Close store" exists, no "open store" | §5.7 adds **Open Store** workflow |
| A4 | Invoice "deferred" yet "required" | Invoices removed from v1 required endpoints; **deferred to Phase D** consistently (§3, §12, §13) |
| A5 | One idempotency key vs re-quote at checkout | §5.4.7 binds idempotency key to a **frozen quote snapshot**; re-quote ⇒ new key |
| A6 | Unlock by other staff, no audit | §10 makes **unlock always audited**; handover made explicit |
| B1 | Refund referenced but never specified | §7.6 adds full **Refund** workflow |
| B2 | "device identity" undefined | §5.8 adds **device enrollment** model |
| B3 | `payment_pending` async undefined | §5.4.8 adds **async payment confirmation** (poll + webhook + timeout) |
| B4 | No document numbering scheme | §13.2 adds **numbering rules** |
| B5 | Concurrency "or equivalent" | §5.4.5 decides **optimistic revision + server lease lock** |
| B6 | Customer data model missing | §13.1 adds **Customer** record |
| B7 | Transfer in-transit loss | §8.6 adds **in-transit & discrepancy** reconciliation |
| B8 | "genuine connection state" undefined | §5.9 defines **connectivity heartbeat** |
| C/D | Timezone, lockout scope, calc order, handover, negative stock | Resolved in §5.3, §9.4, §6.4, §5.6, §8.2 |

---

## 1. Purpose

This document defines the product contract for the NojPOS cashier application and the Laravel API that powers it. It consolidates the agreed direction for POS terminal, Shift & Cash, Settings, Reports, Inventory, Attendance, and Screen Lock into one implementation-ready source of truth.

The goal is a dependable **online-first** POS for Indonesian small and medium businesses (UMKM). Cashiers complete work quickly on a shared tablet; owners and admins retain control over money, stock, staff activity, and configuration.

This PRD does not create implementation tasks. The product owner approves it first; tasks and subtasks are created only after approval and after Section 18 decisions are confirmed.

---

## 2. Product Principles

- **Cashier-first:** checkout, shift control, stock visibility, and day-to-day operations take priority over cosmetic features.
- **Indonesia-first:** Indonesian language, rupiah conventions, and familiar retail terminology are the default.
- **Online-first (bounded):** all operational writes go to the API. The only client-side resiliency is the **single-attempt checkout retry buffer** in §2.1. No general sync queue, offline database, `/sync/pull`, or `/sync/push` exists.
- **Trust through traceability:** money, stock, attendance, and privileged actions produce immutable audit entries.
- **Tenant safety:** the backend resolves `business_id` from the authenticated token. Clients never supply tenant ownership.
- **Explicit scope:** every outlet-bound screen visibly identifies the selected outlet.
- **Server is the calculator:** Flutter and web never compute financial totals, promotions, tax, or stock balances. They render server responses.
- **Stable over clever:** a small number of reliable workflows beats an expansive but inconsistent feature set.

### 2.1 The checkout retry buffer (the only resiliency mechanism)

The retry buffer is **not** an offline mode and **not** a queue of pending writes. It is a strict, bounded mechanism with these rules:

1. It holds **at most one** in-flight checkout attempt at a time, identified by one idempotency key (§5.4.7).
2. It exists only to survive a **transient network failure or delayed response** during the final checkout submit — i.e., the request may have reached the server.
3. On reconnect it **re-sends the same idempotency key** and adopts the server's authoritative result (success, or `409`/`422`).
4. While a checkout attempt is buffered, the terminal is **blocked from starting a new sale** and shows "Menyelesaikan transaksi…".
5. A new sale **cannot begin while offline**, because catalog, stock, customer, and promotion data cannot be refreshed from the server.
6. The buffer is **memory-resident**; it survives backgrounding but a process kill is treated as an indeterminate transaction (§14, "indeterminate checkout").
7. No other domain (shift, cash movement, inventory, attendance, settings) uses any buffer. They require live connectivity and fail with a recoverable error otherwise.

---

## 3. Locked Scope

### 3.1 Included domains (v1 release)

1. POS terminal + checkout/shift workflows (catalog, cart, parked orders, payment, receipt actions).
2. Customers — **search + quick-create** (lightweight; full CRM profile is admin-web only and partially deferred, §5.4.4).
3. Settings for outlet operations and cashier behavior.
4. Reports: sales summary, sold products, payment methods, cashier shifts, void audit, top 10.
5. Inventory: stock on hand, movement history, purchase receipt, stock count, waste, inter-outlet transfer.
6. Shift & cash: open shift, cash movement, close shift, store open/close.
7. **Refund** (§7.6) — included because it is referenced by stock/cash/audit; specified, not assumed.
8. Attendance by staff selection + PIN.
9. Screen lock for a shared terminal.
10. Device enrollment (§5.8) — required infrastructure for shift uniqueness.

### 3.2 Explicit exclusions (v1 release)

- Local server or peer-to-peer cashier networks.
- General offline database / manual full-data sync / write queues beyond §2.1.
- Facial recognition, camera attendance evidence, or biometric matching.
- Supplier master as a separate CRM-like module.
- Payroll, schedules, leave management, commission calculation.
- Dynamic report builder or custom formula engine.
- Editing finalized inventory documents in place.
- **Sales invoices / accounts receivable** — deferred to Phase D (consistently removed from v1 required endpoints and the active data model; see §13.3).
- Promotions/coupons/loyalty as **active terminal features** — the quote contract and UI are delivered, but enabled only after the promotion engine ships (Phase B, §16).
- Direct receipt delivery via WhatsApp/email/SMS (only OS share sheet, print, copy text in v1).

---

## 4. Users and Permissions

| Role | Primary responsibility | Access summary |
| --- | --- | --- |
| Superadmin | SaaS operation | Outside this PRD except normal tenant enforcement |
| Owner | Business control | All outlets, all reports, configuration, stock documents, staff oversight, all authorizations |
| Admin | Daily management | Assigned outlet(s) configuration, reports, inventory, staff actions per the role matrix (§4.1) |
| Cashier | Point of sale operation | Own outlet, current shift, limited reports, attendance, screen lock, permitted cash actions |

Permission checks are enforced **server-side** via Laravel policies/gates. Flutter may hide controls for clarity but is **never** an authorization layer.

### 4.1 Authorization matrix [DECISION — defaults; confirm in §18]

| Action | Cashier | Admin | Owner |
| --- | --- | --- | --- |
| Checkout / park / resume order | ✅ | ✅ | ✅ |
| Void (within open shift) | ⚠️ with authorization PIN | ✅ | ✅ |
| Refund | ❌ | ⚠️ with PIN | ✅ |
| Discount/promotion override above auto | ⚠️ with authorization PIN | ✅ | ✅ |
| Cash in | ✅ | ✅ | ✅ |
| Cash out ≤ limit | ✅ | ✅ | ✅ |
| Cash out > limit | ❌ (needs approver) | ✅ with PIN | ✅ |
| Open shift | ✅ (self) | ✅ | ✅ |
| Close shift (own) | ✅ with fresh PIN | ✅ | ✅ |
| Close shift with variance > threshold | ⚠️ needs approver | ✅ with PIN | ✅ |
| Open/Close store | ❌ | ✅ with PIN | ✅ |
| Purchase / stock count / waste / transfer dispatch & receipt | ❌ | ✅ | ✅ |
| Settings changes | ❌ | ✅ (assigned outlet) | ✅ |

"Authorization PIN" = a privileged user (admin/owner) physically entering their PIN on the terminal to authorize a single action by a cashier. Every authorization is audited with both the requesting actor and the approver.

---

## 5. POS Terminal, Shift, Cash & Terminal Infrastructure

### 5.1 Flutter Cashier App (shared UX rules)

- Optimized for tablet landscape; usable on phone.
- Large touch targets, plain Indonesian labels.
- High-risk actions: full confirmation, explicit financial/stock impact, single final submit.
- Operational forms use a short stepper: information → line items → review → confirm.
- A successful write returns the user to a detail/list state with the new status visible — not only a transient snackbar.
- Empty, loading, error, unauthorized, conflict, and retry states are designed for every API-backed surface (state machine in §11).

### 5.2 Web Admin (shared UX rules)

- Dense, utilitarian: filters, tables, persistent outlet context, detail panels.
- Tables are default for audit/list surfaces; cards reserved for small summary metrics and dialogs.
- Subtle borders, radius ≤ 8px, no nested floating cards.
- Selected outlet appears in every outlet-scoped page header and filter state.

### 5.3 Common patterns

- Amounts displayed as `Rp12.500`; stored/transmitted as **integer rupiah** (bigint).
- **Timezone [DECISION]:** each **outlet** has its own IANA timezone (default `Asia/Jakarta`). All operational timestamps are stored in UTC and rendered in the **outlet's** timezone. Day-boundary logic for shifts, auto-close, and daily reports uses the **outlet timezone**. Cross-outlet reports render each row in its outlet timezone and aggregate by a user-selected reference timezone shown in the report header.
- Statuses have text labels in addition to color.
- Destructive actions name the affected document and result; never a bare `Lanjutkan`.
- All final-submit controls disable while in flight, then restore on a recoverable error.

### 5.4 POS terminal & cashier workspace

The terminal retains active outlet, cashier, shift, and device context throughout the transaction. It is a work surface, not a dashboard, and must stay fast under queue pressure.

#### 5.4.1 Tablet layout

Four stable regions:

1. 64px top bar: menu, outlet name, active cashier, **genuine connection state** (§5.9), shift state, operational notifications, order mode, screen lock, `Daftar Order`.
2. Left category rail: 72px collapsed / 220px expanded — only working features (categories, favorites, packages when supported, shortcuts). Promotions/menu-book/wallet/support are **not** shown as active actions until their contract exists.
3. Center product catalog.
4. 400–440px right order panel.

The connection label must reflect runtime state (§5.9); it must never always claim "online".

Catalog: search by product name, SKU, barcode; results from the catalog provider; retains selected category. Product cards have stable dimensions and show name, price, stock warning (when enabled), and unavailable state. Tap adds one unit; long-press/detail opens product options without interrupting the order.

#### 5.4.2 Mobile layout

Catalog uses full screen; compact bottom bar shows item count + total; opening it shows the order as a full-screen sheet. Checkout never crams catalog + cart + keypad into one narrow viewport. Screen lock, outlet/shift context, and order recovery remain accessible.

#### 5.4.3 Order lifecycle

| State | Meaning | Permitted action |
| --- | --- | --- |
| `draft_cart` | Local unsaved cart | edit, clear, save order, start payment |
| `parked_order` | Server-persisted unfinished order | resume, cancel under policy, start payment |
| `payment_pending` | Payment created, not yet confirmed (async methods) | poll/confirm, cancel per payment policy, expire |
| `paid` | Financially finalized | view, print/share receipt, void/refund under policy |
| `voided` | Reversed before settlement scope | view audit detail only |
| `refunded` / `partially_refunded` | Money returned after `paid` | view audit + refund detail |

Saving an order creates a server order document and clears the visible cart **only after** the server responds OK. Resuming loads it into the cart while preventing concurrent double-checkout (§5.4.5). A cart cannot enter payment without ≥1 valid line. Clearing a non-empty cart asks for confirmation. Removing a single line is immediate with undo, or a short confirmation, per cashier policy [DECISION default: immediate + undo].

#### 5.4.4 Order types & customer context

- Cashier selects an allowed order type before payment; default from outlet settings; backend owns the allowed set per outlet (dine-in, pickup, delivery, online, quick service, etc.).
- Customer selection optional unless an order/promotion/payment method requires it. Picker searches name + phone, offers `Tanpa Pelanggan`, and supports a compact create form (**name required**).
- Chosen customer is shown in the order header and included in the persisted order and final transaction. Changing the customer after promotion/points application triggers server revalidation (and re-quote, §5.4.7).
- Full customer profile enrichment is an admin-web workflow; v1 stores the minimal Customer record (§13.1).

#### 5.4.5 Concurrency control for parked orders **[DECISION]**

- Each `parked_order` carries an integer `revision`. Every mutation must send the last-seen `revision`; server increments on success. A stale `revision` returns `409 conflict_revision`.
- On **resume**, the server grants a **lease lock**: `locked_by` (user+device) with a TTL of **90 seconds**, auto-renewed by client heartbeat while the order is open. Another terminal attempting resume during an active lease gets `409 order_locked` with the holder's identity and remaining TTL.
- A lease expires automatically if the holding terminal stops heartbeating (crash/disconnect), so an order can never be permanently stuck.
- Checkout requires both a valid lease and the current `revision`.

#### 5.4.6 Promotions, coupons, points (quote contract; enabled in Phase B)

- One `Promosi` action surfaces eligible promotions + conditions, coupon code + optional reference, loyalty redemption (only after a customer is selected), and product-bonus selection when several options exist.
- The backend returns a **quote** containing eligible promotions, rejected reasons, discount allocation, bonus availability, final amounts, and a `quote_id` + `expires_at`.
- Flutter renders the quote and **never** independently calculates entitlement.
- Override/coupon/redemption are revalidated at checkout. A cashier may change an automatic/manual promotion only when policy permits; a higher override requires authorization PIN and creates an audit event.
- Unavailable bonus products are disabled with a stock reason. Promotions that drop out of the quote are explained, never silently removed.

#### 5.4.7 Payment, quote freezing & idempotency **[DECISION — resolves A5]**

- Payment is a dedicated screen/full-screen modal showing an immutable order summary, total due, configured payment methods, amount received, change, references, and server-calculated tax/service/rounding.
- At "review", the client requests a **final quote** and receives `{ quote_id, frozen_totals, checkout_idempotency_key }`. The idempotency key is **bound to that frozen quote snapshot**.
- The final submit reuses that single `checkout_idempotency_key`. A retry reuses it and returns the stored result.
- If the quote becomes stale (e.g., promotion expiry, customer change, stock change) the server returns `409 quote_stale`. The client must **re-quote**, which yields a **new** `quote_id` and a **new** `checkout_idempotency_key`. This prevents the "same key, different request" deadlock: a retry never changes the body, and a changed body always uses a new key.
- The client waits for the API result and does not create a second transaction merely because the response was delayed (§2.1).

#### 5.4.8 Async payment confirmation (`payment_pending`) **[DECISION — resolves B3]**

For QRIS/EDC/transfer methods that confirm asynchronously:

- Checkout creates a `payment_pending` transaction and returns `{ transaction_id, payment_ref, confirm_expires_at }` (default expiry **10 minutes** [DECISION]).
- Confirmation arrives via two independent paths, whichever first:
    1. **Server webhook** from the payment provider → server marks `paid` and writes the settlement.
    2. **Client polling** `GET /transactions/{id}` every 3s while the payment screen is open.
- On confirmation: terminal shows the success screen (§5.4.9) and finalizes stock/cash atomically server-side.
- On expiry/decline: transaction moves to `payment_failed`; **no** stock or cash effect is applied; the cashier may retry payment (new attempt) or cancel.
- Cash payments are synchronous and skip `payment_pending`.

#### 5.4.9 Completion & receipt

Success screen shows transaction number, customer, applied discounts, payment lines, change, and receipt actions. v1 receipt actions: **print** (when a compatible printer is configured), **OS share sheet**, **copy digital receipt text**. Direct WhatsApp/email/SMS is excluded until a consented, audited delivery service exists.

#### 5.4.10 Sales invoice (deferred)

Sales invoices are a receivable concern, **deferred to Phase D**. They are not in the v1 order lifecycle, data model, or endpoints (§13.3). When built, an invoice will require customer, invoice number, payment term, billing/delivery address when relevant, notes, and status; issuing an invoice will not create or duplicate a payment.

#### 5.4.11 Screen lock in terminal context

The lock action is always available in the top bar. It fully replaces the POS workspace with the lock UI (§10), preserving draft cart and active shift but hiding all product/customer/price/transaction data until a valid authorized PIN unlocks it.

### 5.5 Four distinct session actions

Labels, confirmations, transitions, and permissions must **not** be merged.

| Action | Purpose | Result |
| --- | --- | --- |
| Lock screen | Temporarily protect the terminal | Returns to PIN; cart + shift remain active |
| Logout cashier | Change the employee on a shared terminal | Returns to staff PIN; shift remains open only when handover policy permits (§5.6) |
| Close shift | Reconcile the active drawer | Finalizes one shift; blocks new sales under it |
| Logout application | Remove the business account from the device | Clears local session → email/password login (device enrollment persists, §5.8) |

### 5.6 Shift lifecycle (open, cash movement, close, auto-expiry, handover)

#### 5.6.1 Open shift

After business login, outlet selection, and cashier PIN verification, the cashier reaches a **shift gate**. The POS workspace is unavailable until an open shift is resolved for the active **outlet + enrolled device**.

The opening form shows outlet, **enrolled device identity** (§5.8), cashier identity, current time, and an **opening-cash** input (integer rupiah; blank by default unless an outlet setting defines an approved default). Opening creates exactly **one active shift per (outlet, device)**; the backend rejects a second. The shift's cashier is the **PIN-authenticated active cashier** — never a client-supplied id. Result shows shift number, opening time, opening cash, and "checkout may begin". Audited as `shift.open`.

#### 5.6.2 Cash movement during an open shift

Cash in / cash out are separate auditable events requiring: open shift, positive integer amount, mandatory reason, active actor, permitted role. Cash-out above the outlet limit requires owner/admin authorization PIN (§4.1).

The terminal shows a cash-drawer view: opening cash, confirmed cash sales, cash in, cash out, cash reversals from void/refund, expected cash, and a chronological movement list. Non-cash totals are visible for reconciliation but never increase expected drawer cash.

Each movement is idempotent (own key) and immutable. Retry reuses the same key. Proof attachment is deferred until a secure upload contract exists.

#### 5.6.3 Close shift (controlled reconciliation)

Never a one-tap state change. Required sequence:

1. Request a **fresh server summary** for the exact active shift.
2. Show opening date/time, cashier, outlet, opening cash, confirmed cash sales, cash in/out, void/refund cash reversals, payment totals, **server-calculated expected cash**.
3. Require a **fresh PIN** from the active cashier (or an authorized owner/admin per policy).
4. Enter physical drawer cash — input starts **blank**; never prefilled.
5. Show **cash variance** before confirmation. Non-zero variance requires a reason; variance beyond the configured threshold requires owner/admin approval.
6. Offer print choices only when a compatible printer is configured: `Cetak laporan tutup shift`, `Cetak produk terjual`.
7. Confirm once; finalize with one idempotency key.
8. Show the immutable close report and route to staff PIN or shift gate — **never** back to an active POS workspace.

The server performs final calculation and status change **atomically** under a row lock / conditional status update. A second request with a **new** key cannot create a second close or overwrite declared cash; a retry with the **original** key returns the original result.

Final report fields (minimum): shift number, outlet, cashier, open/close timestamps, opening cash, cash sales, non-cash payment totals, cash in, cash out, refund/void reversal, expected cash, actual cash, variance, variance reason, close authorizer, document status.

#### 5.6.4 Automatic shift expiry **[DECISION — resolves A2]**

The Settings "automatic shift-close time" is renamed **"shift auto-expiry time"** and does **not** auto-finalize cash:

- At the configured outlet-local time, an open shift is flagged `pending_close` and **blocks new sales**.
- The drawer is **not** reconciled automatically; `actual_cash` stays null.
- The next time the cashier (or an owner/admin) authenticates on that device, the terminal forces the **close-shift flow** (§5.6.3) before any new sale.
- `shift.auto_expiry` is audited; the close itself is still a human, PIN-authorized, physical-cash reconciliation.

#### 5.6.5 Cashier handover **[DECISION — resolves D5/#7]**

- Default: **handover is OFF**; "Logout cashier" requires the current shift to be closed first.
- If handover is enabled per outlet, "Logout cashier" keeps the shift open, but the shift's `responsible_cashier` remains the **opening** cashier for variance accountability, while each transaction records its own `acting_cashier`. The close report shows both the opening cashier and all acting cashiers. Handover events are audited as `shift.handover`.

### 5.7 Store open & close **[DECISION — resolves A3]**

Both are outlet-level operational actions, disabled by default, enabled via outlet settings, available to owner/admin with a fresh authorization PIN.

- **Close store:** checks the outlet for any open/`pending_close` shift. If any remain, the response lists each outstanding shift (number, cashier, device, opening time) and the store cannot close. On success, shows a store-close report and **blocks new shift opening / checkout** at that outlet until a store-open.
- **Open store:** records `store.open` (authorizer, timestamp, outlet) and re-enables shift opening. If outlet operating hours are configured, store-open outside hours requires owner override + reason. A store cannot be opened while already open.

### 5.8 Device enrollment **[DECISION — resolves B2]**

Shift uniqueness depends on a stable device identity, so devices are first-class:

- On first business login, the app generates a persistent `device_uuid` (survives app restarts, cleared only on app logout/uninstall) and registers it via `POST /devices` → server returns a `device_id` bound to `business_id` and assigned outlet(s).
- A device record has: `device_id`, `business_id`, `device_uuid`, label, platform, assigned outlet(s), status (`active`/`revoked`), last seen.
- Shift, cash, and lock operations reference `device_id`. The server rejects operations from a `revoked` or unknown device with `403 device_not_enrolled`.
- Owner/admin can view, label, and revoke devices in web admin. Revoking a device with an open shift requires that shift to be closed first.
- Web admin sessions use a logical device record flagged `web` and may not open shifts (shifts are terminal-bound) [DECISION].

### 5.9 Genuine connection state **[DECISION — resolves B8]**

- The client maintains a connectivity state from a lightweight heartbeat: `GET /health` every **20s** plus the success/failure of real API calls.
- States: `online` (last heartbeat or API call ≤ 30s ago), `degraded` (last success 30–90s ago, retrying), `offline` (no success > 90s).
- The top-bar label reflects this real state and color. New sales are blocked in `offline`; an in-flight checkout uses the §2.1 buffer.

---

## 6. Settings

### 6.1 Information architecture

Settings is an admin navigation destination. The root is a compact directory, not a long form.

| Group | Contents | Scope |
| --- | --- | --- |
| Profil Outlet | Name, address, contact, outlet logo, receipt logo | Outlet |
| Transaksi & POS | Tax, service charge, rounding, order type, opening-cash default, cash-out limit, shift auto-expiry time | Outlet |
| Pembayaran | Cash, EDC, bank transfer, QRIS, active status | Outlet |
| Struk & Dokumen | Receipt content, header, footer, queue number, print limit, preview | Outlet |
| Keamanan & Otorisasi | PIN requirements for sensitive actions, lockout policy, terminal rules | Business, with outlet overrides where noted (§6.7) |
| Perangkat | Enrolled devices, labels, revoke (§5.8) | Business / outlet |
| Notifikasi | Low-stock and operational alerts | User or outlet by type |
| Akun Saya | Name, email, password, personal PIN | User |

Product, category, staff, and subscription remain dedicated modules; Settings links to them but does not duplicate CRUD.

### 6.2 Root page

Header: `Pengaturan`, selected-outlet control, context line (e.g., `Berlaku untuk Outlet Sudirman`). Each group is a compact list item: icon, title, one-line description, completion/warning state, chevron. Incomplete critical config surfaces a warning row (e.g., no payment method, tax enabled without rate, no receipt footer). No fake operational numbers.

### 6.3 Detail-page behavior

Breadcrumb, title, outlet-scope label, and a sticky save bar **only when changes exist** (`Batal perubahan`, `Simpan perubahan`). Dependent controls disable with an explanation. Every persisted config records `updated_by` + `updated_at`; transaction-affecting settings also generate an audit event.

### 6.4 Transaction & POS settings (with deterministic calculation order) **[DECISION — resolves D4]**

**Tax:** enabled status; price-includes-tax vs tax-before-discount vs tax-after-discount; transaction-level rate; product-level tax only after explicit confirmation that product config becomes authoritative (requires the product module's tax field; until then product-level tax is disabled); applicable order types.

**Service charge:** enabled, integer percentage, taxable/non-taxable, applicable order types.

**Rounding:** separate cash vs non-cash config, smallest allowed denomination, deterministic rule. Server computes; client only previews.

**Canonical calculation order (server-authoritative):**

1. Line subtotal = Σ(unit price × qty).
2. Line discounts → item discounts.
3. Order-level discounts/promotions (allocated to lines).
4. Service charge on the post-discount base.
5. Tax — applied per the tax mode (before/after discount; service charge taxed only if marked taxable).
6. Rounding (method-specific) on the grand total.
7. Grand total (integer rupiah).

**Cashier behavior:** default order type, opening-cash default, **shift auto-expiry time** (§5.6.4), daily cash-out limit, custom-amount policy, batch-order display, table-action authorization, desktop hotkeys, merge-remaining-orders policy. Each has an explicit owner/admin policy default.

### 6.5 Payment settings

An outlet can activate cash and non-cash methods. A non-cash method has type (`edc`, `bank_transfer`, `qris`), display name, provider/bank metadata where applicable, async-confirmation capability (§5.4.8), and active status. Changes affect only **new** transactions; existing payments keep their historical method label and amount. Removing a method with history **deactivates** rather than deletes.

### 6.6 Receipt & document settings

Header, footer, transaction note, tax-percentage visibility, print-limit policy (and reprint-needs-authorization toggle [DECISION default: reprint allowed, audited]), queue numbering, logo mode, transaction fields, product fields, QR/barcode fields, optional deposit info. Desktop: config panel + persistent 80mm preview. Mobile: full-screen preview sheet. Previewing never saves.

### 6.7 Security & authorization settings **[DECISION — resolves C3]**

- Which actions require PIN (void, refund, override, cash-out > limit, close shift, store open/close).
- **PIN lockout policy:** scope is **per-staff-account, per-business** (not per-device), so a single bad actor cannot lock the whole terminal. Default: 5 failed attempts → 15-minute lockout for that staff account; an owner/admin can clear a lockout with their PIN. Screen-unlock and attendance share this same per-staff lockout.
- Terminal rules: auto-lock idle timeout (default 3 minutes [DECISION]).

---

## 7. Reports

### 7.1 Structure

Web Reports has one shared filter rail (period preset, custom range, outlet, report-specific filters) and tabs: 1) Sales Summary, 2) Sold Products, 3) Payment Methods, 4) Cashier Shifts, 5) Void/Refund Audit, 6) Top 10.

Cashier Flutter shows a smaller menu: current-shift Sales Summary, Cash Drawer, permitted transaction history. It cannot view another cashier's shift or cross-outlet data.

### 7.2 Sales Summary (with explicit accounting definitions) **[DECISION — resolves #5]**

Definitions used by the server (shown in an info panel):

- **Gross sales** = Σ line subtotals before discounts, excluding tax and service charge.
- **Net sales** = Gross − discounts − refunds, excluding tax/service charge.
- **Collected** = amount actually tendered (includes tax + service charge).
- Voided transactions never count toward paid sales and are shown separately; refunds reduce net sales and are shown separately.

Displays gross and net per the above, paid transaction count, discount total, void total, refund total, average transaction value, payment-method totals, total sold quantity, and a time-series chart (daily range → by hour; weekly/monthly → by day, in outlet timezone).

### 7.3 Sold Products & Top 10

Sold Products: sortable table — product, category, quantity sold, gross sales, discount, net sales, outlet; print/export when available. Top 10: ranked tables — best-selling, highest gross-profit (only when cost coverage is complete; otherwise a clear data-availability state), low-stock, highest order-type value.

### 7.4 Cashier shifts & cash drawer

Shift reports: cashier, outlet, open/close time, opening cash, expected cash, declared closing cash, variance, cash in/out, refunds, status. The cashier Cash Drawer screen is scoped to the active shift (sales, movements, refunds, expected drawer cash, recorded movements). Cash in/out requires type, integer amount, reason, current shift; proof attachment only after the secure upload contract exists.

### 7.5 Void/Refund Audit

Shows transaction number, original order time, void/refund time, cashier, authorizer, reason, order type, amount (and refund amount for partial refunds). Detail view shows immutable transaction lines + void/refund metadata. Only authorized users can void/refund; a void is allowed only while its shift is open and must use idempotency; the API reverses applicable stock + cash atomically.

### 7.6 Refund workflow **[DECISION — resolves B1]**

Refund applies to a **`paid`** transaction after settlement scope (distinct from void, which reverses within the same open shift before/at settlement).

- **Eligibility:** owner, or admin with PIN (§4.1). Cashiers cannot refund.
- **Types:** full or partial (per line/quantity).
- **Flow:** select transaction → choose lines/amounts → mandatory reason → choose refund method (cash from drawer, or original non-cash where supported) → authorization PIN → confirm once → server executes atomically.
- **Effects:** creates `refund` record + `refund_reversal` stock movements (returned items only, unless reason = non-restockable) + a cash movement when refunded in cash (reduces expected drawer cash) → audited as `refund.create`.
- **Constraints:** total refunded never exceeds the paid amount; a fully refunded transaction becomes `refunded`, a partial one `partially_refunded`; idempotency mandatory; finalized refunds are immutable.

---

## 8. Inventory

### 8.1 Information architecture

Three primary views:

- `Stok Saat Ini` — stock on hand by product + outlet.
- `Pergerakan Stok` — immutable history of every change.
- `Dokumen` — purchase invoices, stock counts, waste records, inter-outlet transfers.

Each view has outlet context, search, and an appropriate filter. Stock is never edited directly in a list cell.

### 8.2 Stock on hand, movement history & negative-stock policy **[DECISION — resolves D6/#3]**

On hand: product, category, outlet, current quantity, minimum threshold (when defined), status `Aman` / `Rendah` / `Negatif`. Negative stock is visible as a warning, never hidden.

Movement history: timestamp, product, outlet, type, signed quantity delta, source document number, actor, resulting balance when available. Types: `sale`, `purchase`, `adjustment`, `waste`, `transfer_out`, `transfer_in`, `void_reversal`, `refund_reversal`.

**Negative-stock policy [DECISION default]:** stock-on-hand may go negative for **sales** (online-first must not block a real cashier sale on possibly-stale counts), and the resulting negative is flagged for review. **Inventory documents** (purchase/count/waste/transfer) never *intentionally* drive stock negative without an explicit override reason. Per-outlet, an owner may switch sales to "block sale when stock ≤ 0" mode. This is the single negative-stock switch confirmed in §18.

### 8.3 Purchase invoice

Owner/admin only. Flow: select outlet, document number (auto, §13.2), supplier name (free text), notes, purchase date → add products (integer qty + integer unit cost) → review line subtotals + total → confirm once → create purchase + stock movements in one DB transaction. Finalized document is immutable; corrections via void or a dedicated adjustment document, never silent edits.

### 8.4 Stock count

Controlled adjustment document. Flow: create draft (outlet, number, date, note) → select products → enter actual count beside server-calculated system count → review signed differences + mandatory reasons for non-zero diffs → privileged PIN when variance exceeds threshold → finalize atomically creating `adjustment` movements. An approved count cannot be edited; a new count corrects it.

### 8.5 Waste record

Removes unsellable stock. Required: outlet, document number, date, ≥1 product, quantity, reason. Batch/serial selection mandatory when the product requires it. States `Draf` / `Selesai` / `Dibatalkan`; only a draft can be cancelled. A completed waste record is immutable and creates `waste` movements atomically.

### 8.6 Inter-outlet transfer (with in-transit reconciliation) **[DECISION — resolves B7]**

A two-sided document, not a direct edit.

States: `Draf` → `Diajukan` → `Dikirim` → `Diterima Sebagian` → `Selesai`; or `Dibatalkan`.

- Requester selects destination outlet, source outlet, requested time, note, products, quantities, units.
- **Dispatch (`Dikirim`):** source dispatches a quantity → creates `transfer_out` at source and moves goods into an **in-transit holding bucket** owned by the transfer document (not yet added to destination).
- **Receipt (`Diterima`):** destination confirms actual received quantity → creates `transfer_in` at destination for the received amount, drawing down the in-transit bucket.
- **In-transit discrepancy (loss/damage):** if received < dispatched and the destination closes the line, the **remaining in-transit quantity must be resolved**, not silently dropped: either returned to source (`transfer_return` movement) or written off via a linked `waste`/`adjustment` with a mandatory reason and owner/admin authorization. The transfer cannot reach `Selesai` while in-transit quantity is unresolved. This guarantees Σ(out) = Σ(in) + Σ(returned) + Σ(written-off).
- Requester may cancel only while `Diajukan`. The destination may explicitly close a partial request with a reason.
- Each send/receive/return/write-off has its own immutable movement + idempotency key.

---

## 9. Attendance

### 9.1 Flutter terminal experience

Shared-terminal screen: outlet name, current time, staff selector, six-digit PIN keypad, one contextual action (`Clock In` / `Clock Out`). After success: confirm staff name, action, timestamp; return to ready without exposing the previous PIN. Today's recent attendance list is visible but exposes no PIN or unnecessary personal data.

### 9.2 Web attendance review

Owner/admin table: staff, outlet, clock in, clock out, duration, status. Filters: date range, outlet, staff, open/closed. Search by staff name. v1 stores PIN-verified time records only; camera capture is not shown until storage, retention, consent, and authorization design is approved.

### 9.3 Rules

- A staff member may have only one open attendance record **globally** (not per outlet) [DECISION — resolves D2]; clock-in at a second outlet is rejected with a clear message until clock-out.
- Clock out closes the latest open record for that staff member.
- PIN verified by backend; raw PIN never logged or persisted by Flutter.
- API records business, outlet, staff, action actor, timestamps; audit entries on clock in/out.
- Invalid PIN responses are generic; protected by the per-staff lockout policy (§6.7).

### 9.4 Attendance ↔ shift relationship [DECISION — resolves D1]

Attendance and shift are **independent** by default (a manager may clock in without operating a register). However, the outlet may enable "require clock-in before opening a shift"; when enabled, the shift gate checks for an open attendance record for that staff and blocks shift open otherwise.

---

## 10. Screen Lock

Protects a shared terminal when the cashier steps away. Not a logout; does not close the shift.

### 10.1 Behavior

- Lock action available from the POS header while a cashier session is active; an idle timeout (§6.7) also triggers lock.
- Confirming lock replaces the entire POS surface with a minimal lock screen showing only outlet name, current time, and `Terminal terkunci`.
- Cart, prices, customer info, and reports are not visible while locked.
- A valid PIN from an authorized staff member at the active outlet unlocks the terminal.
- **Unlock is always audited** (`terminal.unlock`) with the unlocking staff + device [DECISION — resolves A6].
- **Same-staff unlock** preserves cart, shift, and active cashier identity. **Different-staff unlock** prompts: *resume as current cashier (covering, audited)* or *perform a cashier handover* (§5.6.5). The terminal never silently changes the responsible cashier.
- Five failed PIN attempts trigger the per-staff 15-minute lockout (§6.7).

### 10.2 Security boundary

Unlocking does not grant approval for voids, refunds, discount overrides, cash-out, or stock adjustments — those retain their own role + PIN authorization. The app clears in-memory PIN input immediately after every success, failure, lockout, background transition, or widget disposal.

---

## 11. Flutter Architecture Requirements

- Feature-first organization retained.
- UI → provider → repository interface → network client. Widgets never call HTTP directly.
- Features `pos`, `shift`, `settings`, `reports`, `inventory`, `attendance`, `screen_lock`, `device` each own UI, provider, repository contract, and API implementation.
- The **checkout retry buffer** (§2.1) is the only resiliency mechanism; no other domain adds offline state.
- Every transactional write creates and retains one UUID v4 idempotency key for retries where the backend declares idempotency mandatory.
- **Canonical UI state machine** for every API-backed surface: `idle → loading → (success | validationError | forbidden | notFound | conflict | recoverableNetworkError)`; `conflict` carries the server reason code (`conflict_revision`, `order_locked`, `quote_stale`, `idempotency_mismatch`) so the UI shows the correct recovery (refresh, re-quote, wait for lease, etc.).
- Lease heartbeats for resumed orders (§5.4.5) and connectivity heartbeat (§5.9) run in dedicated controllers, cancelled on dispose.

---

## 12. Laravel API Requirements

### 12.1 Global rules

- Routes under `/api/v1`; require `Accept: application/json`.
- Success: `{ "data": ..., "meta": ... }`. Error: `{ "error": { "code", "message", "details" } }`.
- All domain data scoped by `business_id` resolved from Sanctum auth.
- UUID v4 for domain identifiers.
- Money: integer rupiah, bigint storage.
- Policies/gates enforce role + outlet assignment for every action.
- Sensitive writes and money/stock actions create immutable audit records.
- All terminal-bound writes verify `device_id` enrollment (§5.8).

### 12.2 Required endpoint families (v1)

| Domain | Read | Write |
| --- | --- | --- |
| Devices | device list/detail | `POST /devices` enroll, label, revoke |
| POS orders | catalog search, order list/detail, promotion quote, final quote | create/update parked order (with `revision`), resume (lease), cancel, checkout |
| Customers | search/detail | quick-create, approved profile updates |
| Promotions (Phase B) | eligible promo/coupon/points quote | apply/remove via order or checkout contract |
| Settings | `GET /outlets/{outlet}/settings`, `/payment-methods`, `/receipt`, `/security` | `PUT` settings, payment-method CRUD, `PUT` receipt, `PUT` security |
| Reports | sales-summary, sold-products, payment-methods, shifts, voids/refunds, top-10 | none |
| Inventory | `GET /inventory`, `/inventory/movements`, document list/detail | purchases, counts, waste, transfer request/send/receive/return/writeoff/cancel/close |
| Attendance | `GET /attendance` | `POST /attendance` (`clock_in`/`clock_out`) |
| Shift & cash | current shift, immutable close report, shift history | open, cash movement, close (PIN), store open, store close |
| Refund | refund detail | `POST /transactions/{id}/refund` (PIN) |
| Payment confirm | `GET /transactions/{id}` (poll) | provider webhook receiver (server-internal) |
| Terminal lock | none (local display lock) | `POST /terminal/unlock-event` (always audited) |

Sales invoices are **not** in v1 (Phase D, §13.3).

### 12.3 Transactional guarantees

Purchases, stock counts, waste completion, transfer send/receive/return/writeoff, checkout, payment settlement, refund, void, cash movement, and shift close execute within DB transactions; external effect is visible only after all stock, money, status, and audit records succeed.

Idempotency is mandatory for: checkout, payment settlement, refund, void, shift open, cash movement, shift close, store open/close, and inventory document finalization. Keys are scoped to `(business_id, request_hash)`. Reusing a key with the same request returns the stored result; reusing with a different request returns `409 idempotency_mismatch` (and, for checkout specifically, the client must re-quote — §5.4.7).

### 12.4 Reporting rules

The server calculates all report figures using the §7.2 definitions; clients never aggregate raw transactions. Cashiers are restricted to their current shift / permitted personal scope; owner/admin scope is controlled by tenant + outlet policy.

### 12.5 Standard error codes

| HTTP | Code | When |
| --- | --- | --- |
| 401 | `unauthenticated` | Session expired/invalid token |
| 403 | `forbidden` / `device_not_enrolled` | No permission / unknown or revoked device |
| 404 | `not_found` | Missing document |
| 409 | `conflict_revision` | Stale order revision |
| 409 | `order_locked` | Active resume lease held elsewhere |
| 409 | `quote_stale` | Frozen quote invalidated → re-quote |
| 409 | `idempotency_mismatch` | Same key, different body |
| 409 | `state_conflict` | Invalid transition / concurrent close / store has open shifts |
| 422 | `validation_error` | Invalid inputs, unopened shift, missing variance reason, etc. |

---

## 13. Data & Audit Model

### 13.1 Core records (v1)

| Record | Essential fields |
| --- | --- |
| Device | business, device_uuid, label, platform, assigned outlets, status, last_seen |
| Customer | business, name (req), phone, email?, notes?, created_by, timestamps |
| Order & OrderItem | business, outlet, number, customer?, type, status, **revision**, **lease (locked_by, expires_at)**, line pricing, discounts, acting/responsible actor |
| Transaction (Paid) | business, outlet, order ref, number, payment lines, totals (gross/net/tax/service/rounding), status (`paid`/`payment_pending`/`payment_failed`/`voided`/`refunded`/`partially_refunded`), idempotency ref |
| PromotionQuote | business, order/cart context, quote_id, eligible/rejected detail, discount allocation, bonus choices, frozen_totals, expires_at |
| Refund | business, outlet, transaction ref, lines/amounts, reason, method, authorizer, idempotency ref, status |
| ShiftSession | business, outlet, **device**, responsible_cashier, acting_cashiers[], opening cash/time, close data, expected/actual cash, variance + reason, authorizer, status (`open`/`pending_close`/`closed`) |
| CashMovement | business, outlet, shift, actor, type, amount, mandatory reason, idempotency ref, timestamp |
| StoreState | business, outlet, state (`open`/`closed`), opened_by/at, closed_by/at, report ref |
| OutletSetting | business, outlet, timezone, tax, service, rounding, cashier behavior, auto-expiry time, updated_by/at |
| SecuritySetting | business, PIN-required actions, lockout policy, idle timeout, outlet overrides |
| PaymentMethod | business, outlet, type, display name, metadata, async-capable flag, active flag |
| ReceiptSetting | business, outlet, content flags, text fields, queue policy, print-limit, updated_by/at |
| AttendanceRecord | business, outlet, staff, clock in/out, creator, timestamps |
| InventoryPurchase & Item | business, outlet, number, supplier text, lines, amount, status, actor |
| StockMovement | business, outlet, product, type, signed qty, source document, actor, timestamp |
| StockTransferInTransit | transfer ref, product, dispatched qty, received qty, resolved (returned/writeoff) qty |
| StockCount | business, outlet, number, lines, system/actual qty, reasons, status |
| WasteRecord | business, outlet, number, lines, reason, status |
| StockTransfer | business, source outlet, dest outlet, status, requested/dispatched/received/returned lines |
| AuditLog | business, actor, approver?, action, entity, before, after, device, timestamp, request context |

Audit logs are immutable. Soft deletion must never erase financial, stock, or security history.

### 13.2 Document numbering **[DECISION — resolves B4]**

- Format: `{TYPE}-{OUTLET_CODE}-{YYYYMM}-{SEQ}` (e.g., `INV-SDM-202606-000123`), where `TYPE` ∈ {TRX, ORD, PUR, CNT, WST, TRF, SHF, REF}.
- Sequences are **per (business, outlet, type, month)**, generated server-side inside the same DB transaction using a row-locked counter (or DB sequence) to be concurrency-safe and gap-tolerant.
- Voided/refunded documents keep their number; the reversal references the original.

### 13.3 Deferred records (Phase D)

`SalesInvoice` and its line/payment-term/receivable fields are **not** created in v1. They are introduced in Phase D with their own receivable model, payment linkage, and invoice audit rules.

---

## 14. Error Handling & States

| Condition | User treatment | API expectation |
| --- | --- | --- |
| Validation failure | Preserve input, field-level explanation | `422 validation_error` |
| Unauthorized | Return to login when session expires | `401` |
| Forbidden | Explain lack of access without exposing other tenant data | `403` |
| Missing document | Not-found state + safe return | `404` |
| Stale order | Prompt refresh, reload latest revision | `409 conflict_revision` |
| Order locked | Show holder + remaining lease, allow retry after expiry | `409 order_locked` |
| Stale checkout quote | Auto re-quote with new key | `409 quote_stale` |
| Duplicate idempotent submit | Show original successful result | stored response |
| Idempotency mismatch | Stop, ask to refresh/review | `409 idempotency_mismatch` |
| Async payment pending | Show pending state + poll; allow cancel/expire | `payment_pending` until webhook/poll resolves |
| Network failure (non-checkout) | Keep draft form in memory, offer retry; no queue | recoverable client state |
| **Indeterminate checkout** (app killed mid-submit) | On relaunch, re-send buffered key; if absent, query by idempotency key before any new sale; show result | server returns stored result or `404` for the key |

---

## 15. Acceptance Criteria

**Devices & connectivity**

- Shift/cash/lock writes from an unenrolled or revoked device are rejected.
- The top-bar connection label reflects real heartbeat state and blocks new sales when `offline`.

**Settings**

- An owner sees which outlet a setting applies to before changing it.
- Tax, service, rounding, payment, receipt changes apply only to new transactions.
- Disabled parent settings prevent dependent controls from being submitted.
- Every saved configuration is auditable.
- Server-computed totals follow the canonical calculation order (§6.4) and are reproducible.

**POS terminal**

- Terminal always shows active outlet, cashier, shift state, and genuine connection state.
- Search, category, customer, cart edits, and payment are usable without layout overlap on supported tablet/mobile breakpoints.
- A parked order is persisted, recoverable by authorized users, protected from concurrent double-checkout via revision + lease.
- Promotion/coupon/points results are server-quoted and revalidated at checkout; a stale quote triggers re-quote, not a duplicate charge.
- Checkout retry reuses one idempotency key bound to a frozen quote and cannot create duplicate paid transactions.
- Async payments resolve via webhook or polling and never apply stock/cash on failure/expiry.

**Shift & store**

- POS cannot be used before an open shift exists for the active outlet + device.
- Opening records the PIN-authenticated cashier; a client cannot nominate another staff member.
- Cash movements require open shift, positive integer amount, reason, authorization, idempotency, audit.
- Shift auto-expiry blocks sales and forces a human close; it never auto-finalizes cash.
- Closing requests a fresh server reconciliation, fresh PIN, blank physical-cash input; non-zero variance records a reason and follows the threshold.
- A second close cannot overwrite original actual cash/variance/report/audit.
- Store close is blocked while any open/`pending_close` shift exists and identifies each; store open re-enables shift opening and is audited.
- Handover (when enabled) preserves variance accountability and audits acting cashiers.

**Reports**

- Summary totals are API-calculated per §7.2 definitions, integer-safe, filterable by valid scope.
- Cashiers cannot access another cashier's shift data.
- Void/refund detail identifies reason, actor, approver, time, and original transaction.

**Inventory**

- No stock-changing document can finalize twice due to retry or double tap.
- Every finalized document creates a traceable stock movement.
- A user cannot mutate stock for another business or unauthorized outlet.
- Completed counts/waste/purchases cannot be edited in place.
- A transfer cannot reach `Selesai` while in-transit quantity is unresolved; Σ(out) = Σ(in) + Σ(returned) + Σ(written-off).
- Negative stock from sales is flagged; "block sale at ≤0" mode prevents it when enabled.

**Attendance & lock**

- Clock in/out requires a valid staff PIN and creates no duplicate open record (one open record globally).
- Screen lock hides POS data and unlocks without closing the shift or losing the cart; every unlock is audited.
- Different-staff unlock requires an explicit cover/handover choice.
- PIN is never logged, stored in plaintext, or retained after an attempt; lockout is per-staff.

**Refund**

- Refund applies only to `paid` transactions, never exceeds paid amount, reverses stock/cash atomically, requires authorization, and is idempotent and immutable.

---

## 16. Product Phases (sequencing guidance, not task breakdown)

**Phase A — Operational foundation**

- Stabilize login, PIN, outlet selection, **device enrollment**, shift, checkout, payment (sync cash), void, attendance, inventory read, purchase, sales summary.
- Add Flutter screen lock with audited unlock.
- Deliver controlled open/close shift: current-cashier binding, mandatory cash-movement reasons, fresh-PIN close, blank physical-cash entry, immutable close report, idempotent concurrency-safe backend, shift auto-expiry (`pending_close`).
- Finalize POS layout, genuine connection state, catalog search, cart behavior, order-type + customer context, share-sheet/print/copy receipt actions, checkout retry buffer.
- Improve attendance terminal UX (no camera).
- Web admin shows real sales summary, inventory, purchase data (no mock figures).

**Phase B — Outlet configuration, async payments, reporting, promotions**

- Settings root + outlet context, profile, tax, service, rounding, payments, receipt, security/lockout.
- Async payment confirmation for QRIS/EDC (webhook + polling).
- Report filters, payment-method detail, sold-product, shift report, void/refund audit.
- Parked orders, server-side recovery (revision + lease), customer quick-create, **promotion quote contract** → then enable promotions in the terminal.
- Audit coverage + role/outlet authorization tests for config and reports.

**Phase C — Inventory control**

- Immutable stock movement history.
- Stock count + waste flows.
- Transfer request/dispatch/partial receipt/return/writeoff/final/cancel with in-transit reconciliation.
- Transaction, idempotency, tenant-isolation, concurrency tests for every stock write.

**Phase D — Hardening, receivables & rollout**

- Print-preview integration where devices support it.
- Opt-in store open/close + close-report printing after settings, printer capability, unresolved-shift checks.
- Exports after report queries are reliable and access-controlled.
- **Refund hardening** and full reporting integration.
- **Sales invoices / receivables** (deferred domain) after the model, payment linkage, and invoice audit rules are approved.
- Tablet/mobile usability validation; Flutter + Laravel regression suites before each release candidate.

---

## 17. Non-Functional Requirements (added)

- **Security:** Sanctum tokens, HTTPS only, PIN hashed (bcrypt/argon2), no PIN in logs; rate-limited auth/PIN endpoints; per-staff lockout; webhook signature verification for payment providers.
- **Performance targets [DECISION]:** catalog search p95 < 400ms; checkout submit→result p95 < 1.5s; report queries p95 < 3s on a single-outlet month.
- **Reliability:** all money/stock writes transactional + idempotent; no partial side effects.
- **Auditability:** every privileged/money/stock action writes an immutable AuditLog with actor, approver, device, before/after.
- **Observability:** structured logs with `request_id`, `idempotency_key`, `business_id`, `outlet_id`, `device_id`; alerting on webhook failures and shift-close anomalies.
- **Data retention:** financial/stock/security history retained indefinitely (soft delete never erases it).
- **Migration:** existing orders/shifts backfilled with `revision = 1`, null lease, and a synthetic enrolled device per active terminal before v2 features go live.
- **Localization:** Indonesian default; rupiah + Asia/Jakarta-per-outlet formatting.

---

## 18. Locked Product Decisions

The product owner confirmed the following defaults on 2026-06-20. They are normative implementation rules until a later approved PRD revision changes them.

1. **Settings scope:** use the scope table in §6.1. Transaction, payment, receipt, and operational settings are outlet-scoped; security rules are business-scoped with permitted outlet overrides.
2. **Inventory roles:** only owner/admin may create or finalize stock counts, waste, transfer dispatches, and transfer receipts.
3. **Negative stock:** sales may create negative stock with a visible warning. The per-outlet `block sale at ≤0` policy exists but defaults to OFF.
4. **Variance thresholds:** every non-zero shift/count variance requires a reason. An owner/admin approval PIN is required when the absolute variance exceeds the lower of Rp100.000 or 1% of expected cash/value.
5. **Accounting definitions:** use §7.2 exactly. Gross-profit reporting is shown only when cost coverage is complete; otherwise the UI shows a data-availability state.
6. **Receipt/print/export:** support compatible ESC/POS Bluetooth/network printers when available, with OS share/copy as fallback. CSV is the first supported report export; no cloud export is required for v1.
7. **Cashier handover:** default OFF. A cashier must close the shift before logout unless a future outlet policy explicitly enables audited handover.
8. **Cash-movement limits:** every movement requires a reason. Cash-out is limited to Rp500.000 per shift by default; amounts above the limit require owner/admin approval PIN.
9. **Store policy:** store open/close is disabled by default. When enabled, owner/admin plus fresh PIN may operate it. Operating hours are optional; any configured out-of-hours opening requires owner override and a reason.
10. **Device policy:** web sessions may not open shifts. Revoked devices cannot perform terminal writes; a device with an open shift must have that shift resolved before revocation.
11. **Async payments:** QRIS, EDC, and bank transfer are asynchronous when configured as such. `payment_pending` expires after 10 minutes.
12. **Concurrency:** parked orders use optimistic integer revision plus a 90-second server lease; the client renews the lease every 30 seconds while the order is open.
13. **Refunds:** owner, or admin with authorization PIN, may issue full or partial refunds. Returned items restock by default; non-restockable items require a reason.
14. **PIN/terminal lock:** lockout is per staff account per business: 5 failed attempts causes a 15-minute lock. Idle screen lock defaults to 3 minutes.
15. **Timezone:** each outlet owns an IANA timezone; default is `Asia/Jakarta` for business-day boundaries and rendering.
16. **Numbering:** use `{TYPE}-{OUTLET_CODE}-{YYYYMM}-{SEQ}` with a server-side, concurrency-safe sequence per `(business, outlet, type, month)`.
17. **Performance targets:** retain §17 targets: catalog search p95 <400ms, checkout p95 <1.5s, and single-outlet monthly reports p95 <3s.
