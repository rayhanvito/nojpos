# NojPOS Admin Web Guide

Read [root guide](../../AGENTS.md), [PRDPOSJA](../../PRDPOSJA.md), [docs index](../../docs/README.md), [story progress](../../docs/STORY_PROGRESS.md), [web admin story](../../docs/STORY_WEB_ADMIN.md), [integration story](../../docs/STORY_INTEGRATION.md), and [bug tracker](../../docs/BUG_TRACKER.md) before editing.

This package owns the Next.js Admin Web app for NojPOS inside the main monorepo.

Surfaces:

- Tenant Admin: `/dashboard` and tenant routes.
- Platform Super Admin: `/platform` and platform routes.

Current mode is UI-first preview until `INT-01` web admin session strategy is approved. Do not add backend/API integration from page/component before the integration boundary is defined.

## Run And Verify

```bash
cd apps/web
npm test
npm run typecheck
npm run lint
npm run build
npx playwright test --list
```

## MUST Rules

- Tenant Admin and Platform Super Admin stay separate.
- Landing remains `/`.
- Tenant Admin remains `/dashboard`.
- Platform Super Admin remains `/platform`.
- Do not store tokens in `localStorage` or `sessionStorage`.
- Do not add random `fetch`/API calls in UI components.
- Keep sensitive actions disabled/gated until integration stories allow them.
- Do not activate refund, void, reprint, export, shift close/open, stock adjustment, payment retry, cash reconciliation, access bantuan, suspend/delete toko, reset password/ban user, or announcement broadcast prematurely.
- Keep mock/fixture data honest while still in preview.
- Use existing dependencies only unless explicitly approved.

## Integration Gate

Before real API integration:

1. Complete `INT-01` session strategy.
2. Define API adapter boundary.
3. Define loading/empty/error/forbidden states.
4. Confirm endpoint DTO and pagination/filter contract.
5. Update tests and boundary scan rules.

## Definition Of Done

- `npm run typecheck` passes.
- `npm run lint` passes.
- `npm test` passes.
- `npm run build` passes.
- `npx playwright test --list` passes.
- Responsive routes have no horizontal overflow when relevant.
- Story docs and bug tracker are updated.

Do not modify backend or cashier from this package unless the user explicitly expands scope.

Final report format:

```text
Built: <changed>
Verified: <commands and results>
Blocker: <precise blocker or none>
```
