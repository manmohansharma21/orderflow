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
| Core capstone-spec gaps | Open | Multi-item orders, order-request idempotency, callback amount/currency verification and failure reason, consumed-stock success semantics, expiry, reports and benchmark measurements |
| Report template requirements | Draft generated (52-page DOCX/PDF working draft) | Replace highlighted personal/date placeholders, insert actual screenshots, refresh List of Figures/contents, finalize citations and export submission copy |
| Learning walkthrough | Pending | Explain the completed Go layers after submission |

## Evidence capture

Exact image filenames, capture timing and the chapter where each image belongs are listed in [`docs/evidence/README.md`](evidence/README.md). The formatted draft report is included separately under `submission/`; it is not submission-final until actual evidence and placeholders are resolved.

## Scope guardrails

Never commit `.env`, tokens, database passwords, real webhook secrets, local dumps, or production data. CI secrets in its current smoke test are dummy values. Keep the worker read-only until a provider-specific remediation policy is specified. AWS steps can incur charges and must not be treated as completed by writing Terraform or a plan.
