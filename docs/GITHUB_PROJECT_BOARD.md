# GitHub Project Board

Project name: NOJ POS P0 Production Readiness

## Columns

1. Backlog
2. Ready for Codex
3. In Progress
4. Review
5. Testing
6. Done
7. Blocked

## Custom fields

| Field | Values |
| --- | --- |
| Priority | P0 / P1 / P2 |
| Area | Backend / Flutter / Fullstack / Docs / QA / DevOps |
| Wave | Parked Orders / Settings / Async Payment / Reports / Inventory / Promo / Refund / Store / Screen Lock / Admin Web / Release |
| Verification | Not Run / Targeted Pass / Full Pass |
| Codex Status | Ready / Running / Needs Fix / Verified |

## Board mapping

### Done

| Issue | Priority | Area | Wave | Verification | Codex Status |
| --- | --- | --- | --- | --- | --- |
| Parked orders revision + lease | P0 | Fullstack | Parked Orders | Full Pass | Verified |
| Settings aggregate + security settings | P0 | Fullstack | Settings | Full Pass | Verified |
| Async payment webhook/polling/expiry | P0 | Fullstack | Async Payment | Full Pass | Verified |
| Reports backend + Flutter | P0 | Fullstack | Reports | Full Pass | Verified |
| Inventory count/waste/transfer/in-transit | P0 | Fullstack | Inventory | Full Pass | Verified |
| Promotion quote contract + UI promo | P0 | Fullstack | Promo | Full Pass | Verified |

### Ready for Codex

| Issue | Priority | Area | Wave | Verification | Codex Status |
| --- | --- | --- | --- | --- | --- |
| Refund contract and POS flow | P0 | Fullstack | Refund | Not Run | Ready |

### Backlog

| Issue | Priority | Area | Wave | Verification | Codex Status |
| --- | --- | --- | --- | --- | --- |
| Store open/close lifecycle | P0 | Fullstack | Store | Not Run | Ready |
| Screen lock and terminal idle policy | P0 | Flutter | Screen Lock | Not Run | Ready |
| Admin web decision / apps-web | P1 | Docs | Admin Web | Not Run | Ready |
| Release candidate QA | P0 | QA | Release | Not Run | Ready |

### In Progress

| Issue | Priority | Area | Wave | Verification | Codex Status |
| --- | --- | --- | --- | --- | --- |
| GitHub checkpoint after Promotion | P0 | DevOps | Release | Targeted Pass | Running |

### Review

No issues currently assigned.

### Testing

No issues currently assigned.

### Blocked

No issues currently assigned.

## Operating rules

- Move an issue to Done only after targeted tests and full verification pass.
- Keep backend contract green before Flutter integration for fullstack waves.
- Do not start Refund until the Promotion checkpoint branch has been pushed.
- Do not start `apps/web` until the Admin Web decision issue is approved.
- Do not commit secrets, local databases, logs, cache directories, vendor dependencies, build outputs, or production mobile credentials.
