# OrderFlow Finalization Status

## Verified before this reliability overlay

- Existing Go/Gin/pgx backend, PostgreSQL migrations 000001–000006, auth, product CRUD, order lifecycle, inventory movements, mock-payment callbacks, and concurrent transactional tests were present in remote `main`.
- Commit `60c255b` passed GitHub Actions: full tests, integration race tests, vet, migration chain, Docker image build, and `/health`/`/ready` container probes.
- The user then locally ran the previous Redis/Docker overlay: Redis/service tests passed and Docker image built; `/health` returned `ok` and `/ready` returned `ready`.

## This overlay adds (must be verified and pushed)

- Multi-product order requests with server-side price snapshots and transactional multi-item reservation.
- Order-creation idempotency scoped to authenticated user + key; same request replays the original result and changed payload returns conflict.
- Mock payment callback amount/currency validation and persisted failure reason.
- Successful payment consumes both on-hand and reserved stock in the same transaction, with one `ORDER_FULFILLED` inventory movement; failed payment releases stock once.
- Safe unpaid-order expiry with `EXPIRED` state, batched row locking (`SKIP LOCKED`), reservation release/audit, and a configurable worker. Orders with pending or successful payment are not expired.
- Admin order status/revenue, sales/payment and inventory/low-stock reports.
- Redis product list/detail caching with TTL and generation-based invalidation; cache failure falls back to PostgreSQL.
- Real API-level p50/p95/mean/throughput benchmark (Redis off vs warm cache) plus a repository-layer benchmark; report only actual results.
- Migration 000007, reliability integration tests, CI learning guide and report/documentation updates.

## Must be done before describing the new overlay as complete

1. From the repo root, run `chmod +x scripts/*.sh && ./scripts/verify-residual-milestone.sh`. This starts dependencies and applies migration 000007 only when not already applied. The script refuses databases whose names are not exactly `orderflow` and `orderflow_test`.
2. Run `./scripts/docker-smoke-test.sh`. It builds and starts the container and validates health/readiness.
3. Run `./scripts/benchmark-product-cache.sh` and keep its actual terminal output for the report. Never fabricate before/after numbers.
4. Review the staged file list. Do not commit `.env`, credentials, backups, or `submission/*DRAFT*` generated reports.
5. Commit and push. Verify the new `main` commit has a successful GitHub Actions run.
6. Only after code/CI and benchmark are final, run the screenshot session and update the Word/PDF report with actual evidence.

## Scope limits / external work

- The callback integration is the intentionally simulated provider described in the project scope; real Stripe/money movement is not claimed.
- AWS is documented but not deployed. A real deployment requires an AWS account/region, credentials, DNS/network decisions, and explicit cost approval. Do not mark it complete using architecture diagrams alone.
- CI-learning walkthrough and the user's earlier Lambda deployment review remain follow-up learning tasks. The earlier Lambda source/repository is not present in this project overlay, so it cannot be truthfully reviewed here.
- Screenshots remain deferred to the final evidence session by request.
