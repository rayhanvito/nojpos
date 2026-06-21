# Validation Checklist NojPOS

Gunakan checklist ini sebelum agent melaporkan task sebagai selesai.

## Global Checklist

- [ ] Agent hanya mengubah file dalam allowed scope.
- [ ] Tidak ada revert perubahan agent lain.
- [ ] Story/task status diperbarui.
- [ ] Bug/blocker dicatat di `docs/BUG_TRACKER.md`.
- [ ] Handoff dibuat jika task masuk review.
- [ ] Tidak ada secret, token, password, PIN, full PII, atau payment reference di log/fixture.
- [ ] Tidak ada package baru kecuali user menyetujui eksplisit.

## Backend Validation

Run dari `apps/backend`:

```bash
composer validate
php artisan route:list --path=api/v1
php artisan test
```

Jika mengubah migration:

```bash
php artisan migrate:fresh --seed
php artisan test
```

Checklist backend:

- [ ] Endpoint berada di `/api/v1`.
- [ ] Response sukses memakai `{ data, meta }`.
- [ ] Response error memakai `{ error: { code, message, details } }`.
- [ ] Query tenant-scoped by authenticated `business_id`.
- [ ] Role/policy/gate dicek di server.
- [ ] Write sensitif memakai transaction.
- [ ] Write sensitif audited.
- [ ] Write sensitif idempotent jika bisa retry.
- [ ] Money integer rupiah.
- [ ] Test tenant isolation ada untuk endpoint baru.
- [ ] Test forbidden/validation/not found ada.

## Web Admin Validation

Run dari `apps/web`:

```bash
npm run typecheck
npm run lint
npm test
npm run build
npx playwright test --list
```

Boundary scan:

```bash
rg -n "fetch\(|XMLHttpRequest|sessionStorage|localStorage|Authorization|Bearer |NEXT_PUBLIC_API_BASE_URL|@/lib/api|@/lib/auth" src/app src/components src/fixtures --glob '!*.test.ts' --glob '!*.test.tsx' || true
```

Checklist web:

- [ ] Landing tetap `/`.
- [ ] Tenant Admin tetap `/dashboard`.
- [ ] Platform Super Admin tetap `/platform`.
- [ ] Tenant Admin dan Platform Super Admin tetap terpisah.
- [ ] Tidak ada direct API call dari UI component sebelum integration gate.
- [ ] Tidak ada localStorage/sessionStorage untuk auth token.
- [ ] Tidak ada action sensitif real sebelum backend siap.
- [ ] Loading, empty, error, forbidden state tersedia untuk route yang sudah integrasi.
- [ ] Tidak ada horizontal overflow mobile/desktop.
- [ ] Copy jujur jika masih preview.

## Mobile Kasir Validation

Run dari `apps/cashier`:

```bash
flutter analyze
flutter test
```

Jika build diperlukan:

```bash
flutter build apk --debug
```

Checklist mobile:

- [ ] Widget tidak membuat HTTP call langsung.
- [ ] UI tidak menghitung final money/stock.
- [ ] Repository/API layer yang memanggil backend.
- [ ] Error state jelas untuk unauthenticated, forbidden, validation, conflict, network.
- [ ] Terminal/outlet/actor context tidak dipercaya bebas dari UI.
- [ ] Shift state tidak bisa dipalsukan dari client.
- [ ] Test provider/widget ditambahkan untuk flow penting.

## Integration Validation

- [ ] API contract ada sebelum web/mobile konsumsi endpoint.
- [ ] Contract response cocok dengan backend test fixture.
- [ ] Web/mobile DTO tidak menebak field backend.
- [ ] Error codes dipetakan ke UI state.
- [ ] Pagination/filter/sort terdokumentasi.
- [ ] Role dan outlet scope terdokumentasi.
- [ ] Sensitive write punya audit dan idempotency plan.

## Merge Readiness

Sebuah branch boleh di-merge jika:

- [ ] Status task `[REVIEW]`.
- [ ] Handoff lengkap.
- [ ] Validasi area terkait lolos.
- [ ] Tidak ada blocker P0/P1 baru.
- [ ] QA Integrator menyetujui merge order.
- [ ] Story board diperbarui ke `[DONE]` setelah merge.
