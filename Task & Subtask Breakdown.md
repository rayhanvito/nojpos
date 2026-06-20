# NojPOS — Task & Subtask Breakdown for Codex

> ⚠️ **LEGACY / RECONCILED REFERENCE ONLY — DO NOT EXECUTE BLINDLY**
>
> This file is a legacy planning reference. It contains future/outdated assumptions such as **Drift/general offline DB** and **Next.js admin web**. The current repo has `apps/cashier` Flutter and `apps/backend` Laravel API only; there is no active `apps/web`, no active Next.js admin web, and no Drift/offline database.
>
> Codex must follow this hierarchy instead: `NOJPOS_POS_FLUTTER_BACKEND_PRD.md` → `docs/ARCHITECTURE.md` → `docs/ANALYSIS_FINDINGS.md` → `docs/IMPROVEMENT_PLAN.md` → `docs/DESIGN.md` / `docs/design-tokens.json` → this file only after reconciliation.
>
> For the current repo, execution order is mandatory: **Wave 0 → Wave 1 → Wave 2**. Admin web is paused until P0 backend contracts are green and explicitly approved. Drift/general offline DB is forbidden unless explicitly approved by a future architecture decision.

## Current correction note

- Flutter cashier currently uses Riverpod, GoRouter, and Dio; it does **not** use Drift.
- Backend is a Laravel API under `apps/backend`.
- Admin web does not exist yet in this workspace.
- Immediate priority after this documentation guard is backend P0 hardening: atomic idempotency, transactional audit, and terminal actor/device/outlet authority. Do not prioritize admin UI or broad feature expansion before P0 is green.

**Sumber:** PRD v2.0 (di atas) · **Target eksekutor:** Codex (agen coding) · **Legacy stack note:** this file still mentions Flutter + Drift and Next.js Admin below as historical/future assumptions; reconcile before use.

> Cara pakai: tiap task punya **Prompt Codex** yang sudah lengkap kontrak + acceptance-nya, tapi sengaja tidak menyebut nama file/path spesifik — Codex menemukan struktur, lokasi, dan isi file sendiri lalu berimprovisasi. Tempel **Prompt Global** (bagian 0) sekali di awal sesi Codex, lalu tempel prompt per-task sesuai urutan dependensi.
> 

---

## 0. Prompt Global (tempel sekali di awal, berlaku untuk semua task)

```
You are working on NojPOS, an online-first POS SaaS for Indonesian UMKM.
Stack correction for the current repo: Flutter cashier app (Riverpod + GoRouter + Dio, feature-first: ui → provider → repository interface → network client; widgets never call HTTP directly) and Laravel API (Sanctum auth, business_id scoping, idempotency). No active Drift/offline DB. No active Next.js admin web.

Before writing code for any task:
1. Explore the repo and discover the actual folder structure, naming conventions, existing models, routes, providers, and tests. Match the existing style; do not invent a new architecture.
2. Reuse existing helpers (auth, scoping, money formatting, idempotency, audit) instead of duplicating them. If a helper is missing, create it following existing patterns.
3. Confirm where similar features live and mirror their layout.

Global engineering rules (must hold for every task):
- All API routes under /api/v1, require Accept: application/json. Success: { "data", "meta" }. Error: { "error": { "code", "message", "details" } }.
- Every domain query is scoped by business_id resolved from the Sanctum token. Clients NEVER supply business_id, cashier_id-as-actor, or tenant ownership.
- Money is integer rupiah (bigint). Server is the only calculator for totals, tax, service, rounding, promotions, and stock balances. UI renders server responses; it never computes financial totals.
- Authorization is enforced server-side via policies/gates. Flutter may hide controls but is never the auth layer.
- Money/stock/privileged writes are transactional AND idempotent (UUID v4 idempotency key scoped to (business_id, request_hash)). Same key + same body → stored result; same key + different body → 409 idempotency_mismatch. Each such write produces an immutable AuditLog (actor, approver?, action, entity, before, after, device, timestamp, request_id).
- Terminal-bound writes (shift, cash, lock) verify device_id enrollment; unknown/revoked device → 403 device_not_enrolled.
- Timestamps stored UTC, rendered in the OUTLET timezone (default Asia/Jakarta). Day boundaries (shift, auto-expiry, daily reports) use outlet timezone.
- Standard error codes: 401 unauthenticated; 403 forbidden/device_not_enrolled; 404 not_found; 409 conflict_revision/order_locked/quote_stale/idempotency_mismatch/state_conflict; 422 validation_error.

Quality bar per task:
- Write/extend automated tests (Laravel feature + unit; Flutter widget + provider; admin where relevant), including tenant-isolation, idempotency, and concurrency cases where applicable.
- Provide DB migrations (additive, backward-compatible), API resources, request validation, and policy classes.
- Keep changes scoped to the task. Output a short summary of files changed and any follow-up needed.
- If a needed business decision is ambiguous, use the documented [DECISION] default from the PRD and note it.
```

---

## PHASE A — Operational Foundation

### EPIC A0 — Platform & cross-cutting infrastructure

**Task A0.1 — Idempotency + standard API envelope + error codes**

Subtasks: BE middleware/envelope · BE idempotency store + replay · BE error-code mapping · tests.

```
Implement (or harden) the shared API foundation: a success/error response envelope, a reusable idempotency mechanism, and the standard error-code mapping defined in the global rules.
- Idempotency: a server-side store keyed by (business_id, idempotency_key, request_hash). Replaying same key+body returns the stored result; same key+different body → 409 idempotency_mismatch. Provide a trait/middleware that any write endpoint can opt into.
- Error mapping: ensure validation → 422 validation_error with field details; auth → 401; permission → 403; missing → 404; conflicts → 409 with the specific code.
Add feature tests proving replay returns identical body/status and mismatch returns 409.
Discover existing middleware/response patterns and integrate rather than replace.
```

**Task A0.2 — AuditLog model + writer**

Subtasks: migration · model · writer service · helper for before/after diff · tests.

```
Create an immutable AuditLog with fields: business, actor, approver (nullable), action, entity_type, entity_id, before (json), after (json), device_id (nullable), request_id, created_at. No update/delete.
Provide a single service/helper that domain code calls inside its DB transaction to record an audit entry, capturing before/after snapshots. Ensure soft-deletes elsewhere never cascade to audit. Add unit tests for diff capture and immutability.
```

**Task A0.3 — Device enrollment**

Subtasks: BE Device model+migration+endpoints · BE policy · BE middleware verifying device on terminal writes · Flutter device_uuid generation+persistence+enroll flow · Admin device list/label/revoke · tests.

```
Implement device enrollment (PRD §5.8).
Backend: Device { business, device_uuid, label, platform, assigned_outlets, status active|revoked, last_seen }. Endpoints: POST /devices (enroll, returns device_id), list/detail, PATCH label, POST revoke. A middleware/guard verifies device_id on all terminal-bound writes (shift, cash, lock) → 403 device_not_enrolled for unknown/revoked. Revoking a device that has an open shift is blocked (state_conflict) until the shift closes.
Flutter: generate a persistent device_uuid on first business login (survives restarts; cleared on app logout/uninstall), enroll, and attach device_id to terminal requests; update last_seen via heartbeat.
Admin (Next.js): device list with label edit and revoke, scoped to assigned outlets.
Web sessions register as platform=web and may NOT open shifts (per [DECISION]).
Tests: enrollment, revoke-blocked-with-open-shift, terminal write rejected for revoked device, tenant isolation.
```

**Task A0.4 — Connectivity heartbeat + genuine connection state (Flutter)**

Subtasks: health endpoint · connectivity controller (online/degraded/offline) · top-bar indicator · block-new-sale-when-offline.

```
Backend: add GET /health (lightweight, authenticated-ok).
Flutter: implement a connectivity controller (PRD §5.9) computing online (≤30s since success), degraded (30–90s, retrying), offline (>90s) from heartbeat (every 20s) + real API call results. Surface a genuine top-bar indicator (label + color) that reflects true state and never hard-codes "online". Block starting a NEW sale while offline; do not block an in-flight checkout (handled separately). Add provider tests for state transitions using a fake clock.
```

**Task A0.5 — Document numbering service**

Subtasks: counter table/sequence · generator inside txn · format helper · tests.

```
Implement document numbering (PRD §13.2): format {TYPE}-{OUTLET_CODE}-{YYYYMM}-{SEQ}, sequences per (business, outlet, type, month), generated server-side inside the same DB transaction with a row-locked counter (or DB sequence) — concurrency-safe, gap-tolerant. Provide a generator used by orders, transactions, purchases, counts, waste, transfers, shifts, refunds. Add a concurrency test (parallel generation produces unique, monotonic sequences).
```

**Task A0.6 — Data migration / backfill for v2 fields**

```
Add additive, backward-compatible migrations and a backfill: orders get revision=1 and null lease; shifts get device + responsible_cashier; create a synthetic enrolled device per existing active terminal so legacy data satisfies v2 constraints. Provide an idempotent backfill command with logging and a dry-run mode. Test on a seeded legacy dataset.
```

---

### EPIC A1 — Auth, PIN, outlet selection, shift gate

**Task A1.1 — Business login + outlet selection + cashier PIN**

Subtasks: BE login/session · BE outlet list scoped to user · BE PIN verify (hashed, rate-limited) · Flutter login → outlet pick → PIN → shift gate routing · tests.

```
Implement the cashier entry flow (PRD §5.6.1): business login (email/password via Sanctum) → outlet selection (only outlets the user may access) → cashier six-digit PIN verification (server-side, hashed with bcrypt/argon2, never logged, rate-limited) → route to the shift gate.
PIN lockout is per-staff-account, per-business (PRD §6.7): 5 failed attempts → 15-minute lockout; owner/admin can clear with their PIN. Invalid PIN responses are generic.
Flutter: build the screens and provider state machine (idle/loading/success/validationError/forbidden/lockout). Tests: wrong PIN lockout, generic error, outlet scoping.
```

**Task A1.2 — Shift gate guard (Flutter + BE)**

```
Block the POS workspace until an open shift exists for the active (outlet, enrolled device) (PRD §5.6.1). Provide GET current-shift that returns the open/pending_close shift for (outlet, device) or none. Flutter routes to the open-shift form when none, and to POS when open. Tests for gate behavior including pending_close (must force close, not allow POS).
```

---

### EPIC A2 — Shift & cash lifecycle

**Task A2.1 — Open shift**

Subtasks: BE ShiftSession model+migration · POST /shifts/open (idempotent, device-bound, PIN-authenticated cashier) · uniqueness (one active per outlet+device) · audit [shift.open](http://shift.open) · Flutter opening form · tests.

```
Implement open shift (PRD §5.6.1). ShiftSession { business, outlet, device, responsible_cashier, acting_cashiers[], opening_cash, open_time, close fields, expected/actual cash, variance+reason, authorizer, status open|pending_close|closed }.
POST /shifts/open requires the PIN-authenticated active cashier context, outlet, device, opening_cash (integer, blank-by-default unless outlet default), idempotency key. Reject a second active shift for the same (outlet, device) with 409 state_conflict. The shift cashier is the authenticated cashier — never a client-supplied id. Audit shift.open. Return shift number, open time, opening cash.
Flutter: opening form (outlet, device identity, cashier, time, opening-cash input defaulting blank). Tests: duplicate-open rejected, client cannot set another cashier, idempotent retry.
```

**Task A2.2 — Cash movements (in/out)**

Subtasks: BE CashMovement model · POST /shifts/{shift}/cash-movements (idempotent) · cash-out limit + authorization PIN above limit · audit · Flutter cash-drawer view · tests.

```
Implement cash in/out (PRD §5.6.2). CashMovement { business, outlet, shift, actor, type cash_in|cash_out, amount (positive integer), reason (mandatory), idempotency_ref, timestamp }. Requires open shift, permitted role; cash-out above the outlet limit requires owner/admin authorization PIN (record actor + approver). Each movement immutable + idempotent + audited.
Flutter cash-drawer view: opening cash, confirmed cash sales, cash in, cash out, void/refund cash reversals, expected cash, chronological movement list; non-cash totals visible but excluded from expected drawer cash. Tests: missing reason → 422, over-limit requires approver, idempotent retry.
```

**Task A2.3 — Close shift**

Subtasks: BE fresh summary endpoint · POST /shifts/{shift}/close (atomic, idempotent, PIN, variance rules) · immutable close report · audit shift.close · Flutter close sequence · tests.

```
Implement close shift as a controlled reconciliation (PRD §5.6.3). Sequence: fresh server summary for the exact shift → show all figures incl. server-calculated expected cash → fresh PIN (active cashier or authorized owner/admin) → physical cash input starting BLANK → show variance → non-zero variance requires reason; beyond threshold requires owner/admin approval → confirm once → finalize atomically under row lock/conditional status update with one idempotency key.
A second request with a NEW key cannot create a second close or overwrite declared cash; retry with the ORIGINAL key returns the original result. Produce an immutable close report (all fields in PRD §5.6.3). After close, route terminal to staff PIN/shift gate — never back to active POS.
Tests: double-close prevented, blank-cash enforced, variance-threshold approval, idempotent retry returns original.
```

**Task A2.4 — Shift auto-expiry (pending_close)**

```
Implement shift auto-expiry (PRD §5.6.4): at the outlet-local configured time, flag the open shift pending_close, block new sales, do NOT reconcile cash automatically (actual_cash stays null). On next authentication on that device, force the close-shift flow before any new sale. Audit shift.auto_expiry. Use outlet timezone for the boundary. Tests with a controllable clock.
```

---

### EPIC A3 — POS terminal core (catalog, cart, parked orders, sync checkout)

**Task A3.1 — Catalog search + product cards**

```
Implement catalog search by product name, SKU, barcode from the API, retaining selected category context (PRD §5.4.1). Product cards: stable dimensions, name, price, stock warning (when enabled), unavailable state. Tap adds one unit; long-press opens product options. Build provider + repository + network client; UI for tablet (center catalog) and mobile (full screen). Tests for search and category retention.
```

**Task A3.2 — Cart (draft_cart) + order panel**

```
Implement the local draft cart and order panel (PRD §5.4.1): lines (qty, unit price, line discount, line total), footer (item count, subtotal, discounts, tax, service, rounding, grand total, Bayar). All totals come from a server quote; the client never computes them. Clearing a non-empty cart confirms; single-line removal is immediate with undo [DECISION]. Cannot enter payment with zero valid lines. Tests for cart edit rules and server-quote rendering.
```

**Task A3.3 — Order types + customer context + minimal Customer**

```
Implement order-type selection (allowed set + default from outlet settings, backend-owned) and customer context (PRD §5.4.4). Add Customer model { business, name (required), phone, email?, notes?, created_by, timestamps } with search (name+phone), quick-create (name required), and "Tanpa Pelanggan". Selected customer shown in header and persisted on order/transaction; changing customer after promotion triggers revalidation. Tests: quick-create validation, tenant isolation.
```

**Task A3.4 — Parked orders with revision + lease lock**

Subtasks: BE Order/OrderItem model · create/update with revision · resume lease (90s TTL + heartbeat) · cancel · 409 codes · Flutter resume + lease heartbeat · tests.

```
Implement parked orders + concurrency control (PRD §5.4.3, §5.4.5). Order { business, outlet, number, customer?, type, status, revision, lease(locked_by, expires_at), lines, discounts, acting/responsible actor }.
- Create/update parked order: client sends last-seen revision; stale → 409 conflict_revision; server increments on success.
- Resume grants a lease lock (locked_by = user+device, TTL 90s, auto-renewed by client heartbeat). Another resume during an active lease → 409 order_locked with holder + remaining TTL. Lease auto-expires if heartbeat stops.
- Checkout requires valid lease + current revision.
Flutter: resume flow, lease heartbeat controller (cancel on dispose), and UI handling for 409 conflict_revision/order_locked. Tests: concurrent resume blocked, stale revision rejected, lease expiry frees the order.
```

**Task A3.5 — Synchronous (cash) checkout + frozen quote + idempotency**

Subtasks: BE final-quote endpoint (quote_id, frozen_totals, checkout_idempotency_key) · checkout endpoint (atomic stock+cash+audit) · quote_stale handling · Flutter payment screen (cash) + retry buffer · success screen · tests.

```
Implement cash checkout (PRD §5.4.7, §5.4.9). At review, client requests a FINAL quote → { quote_id, frozen_totals, checkout_idempotency_key } where the key is bound to that frozen snapshot. Final submit reuses that single key; retry returns stored result. If the quote is invalidated (stock/customer/promo/expiry) → 409 quote_stale → client re-quotes (new quote_id + new key), preventing same-key/different-body deadlock.
Checkout executes atomically: create Transaction (paid), decrement stock (sale movements), record cash payment, generate number, write audit. Implement the single-attempt checkout retry buffer on the client (PRD §2.1): one in-flight attempt, re-send same key on reconnect, block new sale while buffered, treat process-kill as indeterminate (query by key before any new sale).
Success screen: transaction number, customer, discounts, payment lines, change, receipt actions placeholder. Tests: duplicate submit returns one transaction, quote_stale → re-quote, atomic rollback on failure.
```

**Task A3.6 — Receipt actions (print / share sheet / copy text)**

```
Implement v1 receipt actions (PRD §5.4.9): print when a compatible printer is configured, OS share sheet, copy digital receipt text. No direct WA/email/SMS. Build a receipt payload from the server transaction; render an 80mm-style receipt for print. Tests for payload composition; gate print on printer availability.
```

**Task A3.7 — Void**

```
Implement void (PRD §7.5): allowed only while the transaction's shift is open, by authorized users (cashier needs authorization PIN; admin/owner allowed). Idempotent. Reverses applicable stock (void_reversal) and cash atomically. Audit void with actor + approver + reason. Tests: void after shift close blocked, stock/cash reversal correctness, idempotency.
```

---

### EPIC A4 — Attendance (terminal) + Screen lock

**Task A4.1 — Attendance clock in/out**

```
Implement attendance (PRD §9): shared-terminal screen (outlet, time, staff selector, six-digit PIN, contextual Clock In/Out). AttendanceRecord { business, outlet, staff, clock_in/out, creator, timestamps }. Rules: one OPEN record GLOBALLY per staff (second-outlet clock-in rejected until clock-out); clock-out closes latest open record; PIN verified server-side, never logged; audit clock in/out; per-staff lockout on bad PIN. Optional outlet setting "require clock-in before opening a shift" — wire the hook but default OFF. Admin web review table comes in Phase B. Tests: duplicate open prevented, generic PIN error.
```

**Task A4.2 — Screen lock (Flutter) + audited unlock**

```
Implement screen lock (PRD §10). Lock from header or idle timeout (default 3 min) fully replaces POS with a minimal lock screen (outlet name, time, "Terminal terkunci" only). A valid authorized-staff PIN unlocks. ALWAYS audit unlock (POST /terminal/unlock-event) with unlocking staff + device. Same-staff unlock preserves cart/shift/identity; different-staff unlock prompts cover-vs-handover (handover defaults OFF). Five failed attempts → per-staff 15-min lockout. Clear in-memory PIN immediately on every success/failure/lockout/background/dispose. Unlocking grants NO action approvals. Tests: data hidden while locked, cart/shift preserved, unlock audited, PIN cleared.
```

---

### EPIC A5 — Admin web foundation (real data)

**Task A5.1 — Admin shell + outlet context + sales summary (real)**

```
In the Next.js admin, build the dense utilitarian shell (filters, tables, persistent outlet context in every header/filter; radius ≤8px; no nested floating cards). Wire a real (non-mock) Sales Summary using the server-calculated report (gross/net/collected per PRD §7.2). Add inventory-on-hand and purchase list read views using real endpoints. No fabricated figures anywhere. Tests/snapshot for outlet-context persistence.
```

---

## PHASE B — Outlet Configuration, Async Payments, Reporting, Promotions

### EPIC B1 — Settings

**Task B1.1 — Settings root + outlet scope**

```
Build the Settings root directory (PRD §6.1–6.2): compact group list (Profil Outlet, Transaksi & POS, Pembayaran, Struk & Dokumen, Keamanan & Otorisasi, Perangkat, Notifikasi, Akun Saya), each with icon/title/one-line/completion-or-warning/chevron, selected-outlet control, context line. Surface warnings for incomplete critical config (no payment method, tax enabled without rate, no receipt footer). No fake numbers.
```

**Task B1.2 — Transaction & POS settings + deterministic calc**

```
Implement OutletSetting (tax, service, rounding, cashier behavior, shift auto-expiry time, timezone) with the canonical server-side calculation order (PRD §6.4): subtotal → line discounts → order discounts/promotions → service charge on post-discount base → tax per mode → rounding → grand total. Product-level tax stays DISABLED until the product module exposes the field. Detail page: breadcrumb, outlet-scope label, sticky save bar only when dirty, dependent-control disabling with explanation, updated_by/at, audit on transaction-affecting changes. Add server unit tests proving reproducible totals for representative configs.
```

**Task B1.3 — Payment method settings**

```
Implement PaymentMethod { business, outlet, type cash|edc|bank_transfer|qris, display_name, metadata, async_capable, active } (PRD §6.5). Changes affect only NEW transactions; existing payments keep historical labels; removing a used method deactivates (never deletes). Admin CRUD + Flutter consumption. Tests for deactivate-not-delete and historical integrity.
```

**Task B1.4 — Receipt & document settings**

```
Implement ReceiptSetting (PRD §6.6): header, footer, note, tax-% visibility, print-limit (+ reprint-needs-auth toggle, default reprint allowed+audited), queue numbering, logo mode, transaction/product/QR fields. Desktop config panel + persistent 80mm preview; mobile full-screen preview sheet; preview never saves. Tests for save-vs-preview separation.
```

**Task B1.5 — Security & authorization settings**

```
Implement SecuritySetting (PRD §6.7): which actions require PIN (void, refund, override, cash-out>limit, close shift, store open/close); per-staff lockout policy (5/15min, owner clear); idle auto-lock timeout (default 3 min); business-level with outlet overrides. Wire these settings into the relevant guards. Tests that toggling a PIN requirement changes enforcement.
```

---

### EPIC B2 — Async payments

**Task B2.1 — Async payment confirmation (QRIS/EDC/transfer)**

```
Implement async payment (PRD §5.4.8). Checkout for async methods creates a payment_pending transaction → { transaction_id, payment_ref, confirm_expires_at } (default 10 min [DECISION]). Resolution by whichever first: provider webhook (signature-verified) marks paid + settles stock/cash atomically; OR client polling GET /transactions/{id} every 3s while payment screen open. On expiry/decline → payment_failed with NO stock/cash effect; allow retry or cancel. Cash stays synchronous. Tests: webhook settle, poll settle, expiry no-effect, idempotent settlement (no double-apply if webhook + poll race).
```

---

### EPIC B3 — Reporting

**Task B3.1 — Report filter rail + Sales Summary (definitions)**

```
Build the web Reports shell with one shared filter rail (period preset, custom range, outlet, report-specific filters) and tabs. Implement Sales Summary with explicit server definitions (PRD §7.2): gross, net, collected, paid count, discount total, void total, refund total, avg txn value, payment-method totals, total qty, time-series (daily→hour, weekly/monthly→day in outlet timezone). Info panel shows definitions. Voids/refunds excluded from paid sales and shown separately. Tests for definition correctness and scope filtering.
```

**Task B3.2 — Sold Products + Top 10**

```
Implement Sold Products (sortable: product, category, qty, gross, discount, net, outlet; print/export when available) and Top 10 (best-selling, highest gross-profit only when cost coverage complete else data-availability state, low-stock, highest order-type value). Server computes all figures. Tests for gross-profit suppression when cost incomplete.
```

**Task B3.3 — Cashier shifts + cash drawer report**

```
Implement shift reports (cashier, outlet, open/close, opening cash, expected, declared, variance, cash in/out, refunds, status) and the cashier-scoped Cash Drawer screen (active shift only). Cashiers cannot view other cashiers' shifts or cross-outlet data — enforce server-side. Tests for scope enforcement.
```

**Task B3.4 — Void/Refund audit report**

```
Implement Void/Refund Audit (PRD §7.5): list (txn number, original time, void/refund time, cashier, authorizer, reason, order type, amount, refund amount) + detail with immutable lines and metadata. Server-side authorization for who can view. Tests.
```

---

### EPIC B4 — Promotions (quote contract → enable in terminal)

**Task B4.1 — Promotion quote contract**

```
Implement the promotion/coupon/points QUOTE contract (PRD §5.4.6): given order/cart context (+ optional coupon, customer for points), return { quote_id, eligible promotions, rejected reasons, discount allocation, bonus availability, frozen final amounts, expires_at }. Backend owns all entitlement; the client only renders. Revalidate at checkout (ties into quote_stale, A3.5/B2.1). PromotionQuote persisted per PRD §13.1. Tests: rejected reasons surfaced, points require customer, expiry behavior.
```

**Task B4.2 — Enable Promosi action in terminal**

```
After B4.1 ships, surface the single Promosi action in the order panel (eligible promos + conditions, coupon entry, loyalty redemption after customer selected, product-bonus selection). Overrides require authorization PIN + audit; unavailable bonuses disabled with stock reason; dropped promotions explained, never silently removed. Tests for override audit and quote re-render.
```

---

## PHASE C — Inventory Control

**Task C1 — Stock on hand + movement history + negative-stock policy**

```
Implement StockMovement { business, outlet, product, type sale|purchase|adjustment|waste|transfer_out|transfer_in|void_reversal|refund_reversal|transfer_return, signed_qty, source_document, actor, timestamp } and read views: Stok Saat Ini (qty, threshold, status Aman/Rendah/Negatif — negatives flagged, never hidden) and Pergerakan Stok (immutable history with resulting balance when available). Negative-stock policy (PRD §8.2): sales may go negative-with-flag; optional per-outlet "block sale at ≤0" mode. Tests for status computation and the block-mode toggle.
```

**Task C2 — Purchase invoice**

```
Implement purchase invoice (PRD §8.3), owner/admin only: outlet, document number (A0.5), supplier free text, notes, date, lines (integer qty + integer unit cost), review totals, confirm once → create purchase + stock movements in ONE transaction. Finalized = immutable (corrections via void/adjustment). Tests: atomicity, immutability, role enforcement, idempotency.
```

**Task C3 — Stock count**

```
Implement stock count (PRD §8.4): draft (outlet, number, date, note) → select products → enter actual vs server-calculated system count → review signed diffs + mandatory reasons for non-zero → privileged PIN when variance > threshold → finalize atomically creating adjustment movements. Approved count immutable (correct via new count). Tests: variance-threshold PIN, atomic finalize, immutability.
```

**Task C4 — Waste record**

```
Implement waste (PRD §8.5): outlet, document number, date, ≥1 product, qty, reason; batch/serial mandatory when product requires. States Draf/Selesai/Dibatalkan (only draft cancellable). Completed = immutable, creates waste movements atomically. Tests for state rules and atomicity.
```

**Task C5 — Inter-outlet transfer with in-transit reconciliation**

```
Implement inter-outlet transfer (PRD §8.6) as a two-sided document with states Draf→Diajukan→Dikirim→Diterima Sebagian→Selesai / Dibatalkan, plus a StockTransferInTransit holding bucket.
- Dispatch: transfer_out at source → goods into the document's in-transit bucket (NOT yet at destination).
- Receive: transfer_in at destination for received qty, drawing down the bucket.
- Discrepancy: remaining in-transit must be resolved (transfer_return to source OR linked waste/adjustment with mandatory reason + owner/admin auth) before Selesai. Enforce invariant Σ(out) = Σ(in) + Σ(returned) + Σ(written-off).
- Requester may cancel only while Diajukan; destination may close a partial with a reason.
Each send/receive/return/writeoff has its own immutable movement + idempotency key.
Tests: invariant enforced, cannot Selesai with unresolved in-transit, idempotency, concurrency on dispatch/receive.
```

**Task C6 — Inventory hardening test suite**

```
Add a cross-cutting test suite for every stock write: DB-transaction atomicity, idempotency (no double-finalize on retry/double-tap), tenant isolation (no cross-business/outlet mutation), and concurrency (parallel finalize/receive). Ensure each finalized document yields a traceable movement.
```

---

## PHASE D — Hardening, Receivables & Rollout

**Task D1 — Print-preview integration**

```
Integrate print preview where devices support it (shift close report, sold-products, receipts). Gate on detected printer capability; graceful fallback when unavailable. Tests for capability gating.
```

**Task D2 — Store open/close**

```
Implement store open/close (PRD §5.7), disabled by default, enabled per outlet, owner/admin + fresh PIN. Close store checks for any open/pending_close shift; if present, list each (number, cashier, device, open time) and block; on success show store-close report and block shift open/checkout. Open store records store.open, re-enables shift opening; out-of-hours open requires owner override + reason; cannot open while already open. StoreState model per §13.1. Tests: close blocked with open shift, open re-enables, audit.
```

**Task D3 — Refund (full + partial)**

```
Implement refund (PRD §7.6) on paid transactions: owner, or admin with PIN; cashiers cannot. Full or partial (per line/qty). Flow: select txn → lines/amounts → mandatory reason → method (cash from drawer or original non-cash where supported) → authorization PIN → confirm → atomic execution: Refund record + refund_reversal stock movements (returned items unless non-restockable) + cash movement when cash (reduces expected drawer cash) → audit refund.create. Never exceed paid amount; status refunded/partially_refunded; idempotent; immutable. Tests: over-refund blocked, atomic reversal, idempotency, role enforcement.
```

**Task D4 — Exports**

```
Add report exports (CSV/whatever format confirmed in §18 #6) only after report queries are reliable and access-controlled. Enforce the same scope rules as the on-screen reports. Tests for scope-correct exports.
```

**Task D5 — Sales invoices / receivables (deferred domain)**

```
Only after PRD §18 sign-off: implement SalesInvoice + receivable model, payment-term/address/status, invoice numbering, and the rule that issuing an invoice does NOT create or duplicate a payment. Add invoice audit rules. Keep entirely separate from the paid-transaction flow. Full tests for receivable lifecycle.
```

**Task D6 — Regression + usability gate**

```
Run and stabilize full Flutter (widget + provider + integration) and Laravel (feature + unit) regression suites, plus tablet/mobile usability validation, as a release-candidate gate. Add CI checks enforcing tenant-isolation, idempotency, and concurrency test presence for money/stock endpoints.
```

---

## Catatan eksekusi untuk Codex

- **Urutan wajib:** selesaikan **EPIC A0** dulu (infra lintas-fitur) sebelum fitur lain, karena hampir semua task bergantung pada idempotency, audit, device, numbering, dan envelope.
- **Dependency utama:** A1 → A2 → A3 (shift gate sebelum POS); A3.4 (lease/revision) sebelum A3.5 (checkout); B4.1 (quote) sebelum B4.2 dan sebelum mengaktifkan promo di checkout; C1 sebelum C2–C5; D3 (refund) bergantung pada A2 (cash drawer) + C1 (stock movement).
- **Setiap task:** Codex eksplorasi repo dulu, ikuti pola yang ada, tulis test (termasuk tenant-isolation, idempotency, concurrency bila relevan), dan laporkan file yang berubah.
- **Keputusan bisnis:** pakai default **[DECISION]** dari PRD bila ambigu, dan tandai di output supaya kamu bisa override saat review Section 18.
