# NojPOS Admin Web

Status: UI-first preview. Backend integration is not approved in this phase.

This app is the UI-first preview surface for NojPOS Admin Web. Product scope is defined in `../../docs/ADMIN_WEB_PRD.md` and now includes two separated surfaces: Tenant Admin for store owners/admins, and Platform Super Admin for the NojPOS SaaS owner/operator. It exists to settle information architecture, layout, responsive behavior, visual states, and reusable UI primitives before connecting to Laravel.

## Current scope

Active Tenant Admin preview routes:

- `/`
- `/login`
- `/dashboard`
- `/settings`
- `/settings/business`
- `/settings/outlet`
- `/settings/payments`
- `/settings/receipt`
- `/reports`
- `/inventory`
- `/inventory/[resourceId]`
- `/staff`
- `/staff/[staffId]`
- `/catalog`
- `/catalog/[productId]`
- `/promotions`
- `/promotions/[promotionId]`
- `/audit`
- `/audit/[eventId]`

Additional active Tenant Admin preview routes:

- `/customers`
- `/customers/[customerId]`
- `/attendance`
- `/attendance/[recordId]`
- `/outlets`
- `/outlets/[outletId]`
- `/transactions`
- `/transactions/[transactionId]`
- `/payments`
- `/payments/[paymentId]`
- `/subscription`

Active Platform Super Admin preview routes:

- `/platform`
- `/platform/businesses`
- `/platform/businesses/[businessId]`
- `/platform/plans`
- `/platform/subscriptions`
- `/platform/revenue`
- `/platform/users`
- `/platform/audit`
- `/platform/system-health`

Every route must load without Laravel, a browser session, or real credentials. Route copy, placeholder-safe rows, detail fixtures, action-gate reasons, and state coverage live under `src/fixtures/preview/**` so future endpoint adapters can replace fixtures one feature family at a time. Tenant navigation and platform navigation remain separate via `AdminShell` and `PlatformShell`.

## Preview-mode rules

- Use preview fixtures only and label them with `Preview` or `Belum terhubung ke backend`.
- Do not call Laravel endpoints from rendered routes, components, hooks, tests, or route handlers.
- Do not persist bearer tokens, implement real browser login, or use local/session storage for production auth.
- Keep API/auth boundary files inert, contract-only, and unimported by preview pages/components until `../../docs/API_CONTRACT_MATRIX.md` is approved for a feature family.
- Keep Tenant Admin and Platform Super Admin navigation separate; do not mix `/platform/*` into the tenant sidebar as ordinary tenant pages.
- Do not show tenant outlet context inside platform routes.
- Do not create checkout, payment, shift, device, terminal, refund, store open/close, cash-management, or direct stock-edit UI.
- Do not calculate price, discounts, promotions, tax, service, rounding, totals, stock, or reports in the browser.
- Keep create/edit/delete/import/export controls disabled and explain the missing backend contract before any API integration.
- Prefer `—` instead of invented production-looking numbers.

## Premium template design and assets

The root `nextjs/` folder is the local premium-template source for admin-web concept, layout style, and UI assets. Use it to match the intended premium admin look: light shell, compact sidebar/header, table-first pages, polished cards, modest borders, restrained green accent, icons, illustrations, and empty/unavailable-state visuals.

Allowed:

- Copy only the specific static assets needed from `../../nextjs/public/**` or approved template icon/style assets into this app.
- Prefer destination paths under `public/template/**` or another clearly named local asset folder.
- Keep usage local to the NojPOS project and preserve any vendor/license notice that already exists.
- Rebuild the UI as NojPOS-owned components, copy, routes, and preview data.

Not allowed:

- Do not copy template auth, Prisma, provider, middleware, API, dependency, Docker, deployment, or business-logic code.
- Do not bulk-copy the whole template or media library into this app.
- Do not show vendor demo data as NojPOS production data.

## Commands

Run from this directory:

```powershell
npm test
npm run typecheck
npm run lint
npm run build
```

Use `npm run dev` only for local preview.

## Future integration gate

Backend integration can start only after an explicit approval that defines:

1. Browser session/auth model.
2. CSRF/logout/token-storage decision.
3. `../../docs/API_CONTRACT_MATRIX.md` with envelopes, filters, pagination, role/outlet behavior, error codes, idempotency, and audit requirements.
4. Feature-by-feature integration order.

Until then, this app remains a static UI-first preview.
