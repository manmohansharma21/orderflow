# Remote review and residual work tracker

## Verified remote baseline

The previous documentation/CI commit was `60c255b`. GitHub Actions run `39126595788` completed successfully with conclusion `success`: the disposable PostgreSQL migration chain, unit/integration tests, race-enabled integration tests, `go vet`, whitespace check, Docker image build, and runtime `/health` and `/ready` probes all passed at that revision. Do not extend that verification claim to later uncommitted changes without a new successful run.

## Implemented in the residual bundle (verify after pushing)

- Redis product list/detail cache with TTL and generation-based invalidation on product mutations.
- Cache adapter and PostgreSQL+Redis invalidation tests (Redis tests skip locally when `ORDERFLOW_REDIS_TEST_ADDR` is not set; CI sets it).
- Configurable read-only periodic payment reconciliation worker and worker cancellation/logging test.
- Compose Redis service, configuration example, README/report updates and HLD refresh.
- CI adds Redis service, Redis tests, Docker build and runtime health/readiness smoke checks.

## Cumulative tracker

| Area | Status | Next evidence/action |
|---|---|---|
| Go/Gin APIs, request DTOs, domain models, services and repositories | Implemented | Existing code; no duplicate packages needed |
| PostgreSQL schema/migrations and transactions | Implemented; prior CI passed | CI after new bundle |
| JWT, product CRUD, orders, inventory, mock payments | Implemented and previously runtime-tested | Postman run and screenshots |
| Integration/concurrency/rollback tests | Committed; prior CI passed | Retain in green CI |
| README, design-pattern guide, OpenAPI, Postman collection | Committed | Keep synchronized with route changes |
| HLD, LLD and order/payment state diagrams | Committed; HLD updated in bundle | Confirm rendering and capture if rubric asks |
| Redis product cache/TTL/invalidation | Added in residual bundle | Green CI proves compile/functional integration tests |
| Scheduled reconciliation | Read-only worker added in residual bundle | Confirm rubric scope; inspect logs after run |
| Docker build/runtime | Workflow smoke test added | Green Actions run after push; optionally local screenshot |
| Live Stripe/provider integration | Mock only | Add only if rubric requires it; needs provider configuration |
| AWS deployment | Plan only, no deployed resources verified | Deploy with credentials/budget and capture health evidence |
| Postman screenshots/API evidence | Pending | Follow `docs/evidence/README.md`; capture actual redacted responses and paste into named report chapters |
| Final project report PDF and submission ZIP | Pending | Finalize using actual run evidence |
| Earlier Lambda submission | Pending review | Add old deployment artifact/evidence to report |
| Learn CI configuration | Pending | Walk through `.github/workflows/ci.yml` step by step: triggers, service containers, migrations, tests, race, vet, Docker and smoke probes |
| Core capstone-spec gaps | Implemented in this reliability overlay; locally unverified | Run migration 000007, full tests and new CI; benchmark measurements still pending |
| Report template requirements | Draft generated (52-page DOCX/PDF working draft) | Replace highlighted personal/date placeholders, insert actual screenshots, refresh List of Figures/contents, finalize citations and export submission copy |
| Learning walkthrough | Pending | Explain the completed Go layers after submission |

## Evidence capture

Exact image filenames, capture timing and the chapter where each image belongs are listed in [`docs/evidence/README.md`](evidence/README.md). A local DOCX/PDF working draft may exist under `submission/`, but draft binaries are intentionally excluded from Git. The Markdown report source is in `docs/PROJECT_REPORT.md`; produce the final Word/PDF after actual measurements and the final evidence session.

## Scope guardrails

Never commit `.env`, tokens, database passwords, real webhook secrets, local dumps, or production data. CI secrets in its current smoke test are dummy values. Keep the worker read-only until a provider-specific remediation policy is specified. AWS steps can incur charges and must not be treated as completed by writing Terraform or a plan.


## Completion-bundle additions (verify locally/CI before marking complete)

- Multi-product order payloads and per-user order idempotency, with integration cases for identical replay and changed-payload conflicts.
- Amount/currency verification for HTTP callback payloads and durable payment failure reason.
- Successful-payment stock consumption plus `ORDER_FULFILLED` audit movement; repeated callback must not duplicate consumption.
- Periodic expiry for old unpaid orders with no pending/success payment; ambiguous/pending provider outcomes are intentionally retained for reconciliation.
- Admin order, sales and inventory summary endpoints.
- A repeatable `scripts/benchmark-product-cache.sh` benchmark; real measurements still need to be run and recorded.
- CI configuration walkthrough and earlier Lambda submission remain learning/review tasks. AWS deployment and screenshot evidence are deliberately pending until actual account access/deployment and the final verification session.
