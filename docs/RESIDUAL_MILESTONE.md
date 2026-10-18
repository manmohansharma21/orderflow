# OrderFlow reliability completion milestone

This overlay extends the existing Redis/Docker bundle to close the remaining application-level items identified in the OrderFlow project specification. See `ORDERFLOW_RELIABILITY_COMPLETION_INSTRUCTIONS.md` for the single extraction/verification/commit workflow. Extract it into the repository root. It does not include credentials, database dumps, or the generated DOCX/PDF working draft.

## Implemented in this overlay (must pass local verification and new CI)

- Multi-product order requests (up to 50 distinct products), backend price snapshots, and atomic reservations.
- Required `Idempotency-Key` for HTTP order creation; same-key/same-payload replay returns the original result, while changed payload returns `409 Conflict`.
- Mock callback amount/currency verification and persisted failed-payment reason.
- Successful payment atomically consumes on-hand and reserved stock and records `ORDER_FULFILLED`; replay does not consume twice.
- Periodic unpaid-order expiry releases reservations only if there is no pending/successful payment. The worker checks status/payment under the transaction after locking candidate rows.
- Admin order, sales/payment, and low-stock inventory summary endpoints.
- Redis product list/detail TTL caching plus mutation invalidation.
- Actual-measurement benchmark command for direct PostgreSQL list reads versus warm Redis cache.
- Reliability integration tests, migration `000007`, CI configuration learning guide, and updated OpenAPI/README/report source.

## One verification path

Use Go 1.26+, Docker and the dedicated local `orderflow_test` database. Keep `ORDERFLOW_TEST_DATABASE_URL` set to that database as described in `INTEGRATION_TESTING_MILESTONE.md`. From the repository root:

```bash
chmod +x scripts/*.sh
./scripts/verify-residual-milestone.sh
./scripts/docker-smoke-test.sh
./scripts/benchmark-product-cache.sh
```

`verify-residual-milestone.sh` starts local PostgreSQL/Redis, applies only migration 000007 to database names exactly `orderflow` and `orderflow_test` if it is not already present, then runs `go mod tidy`, formatting, all tests, integration race tests, `go vet` and whitespace checks. It refuses to use any other DB name for the test connection. The smoke script builds and starts the container, then probes `/health` and `/ready`. The benchmark emits real measurements from your machine; copy only those measured values into the report.

If a command fails, do not commit. Share the first failing command and output (with all connection details/secrets removed) and fix it before proceeding.

## Correct commit hygiene

Before staging, unstage the unpersonalized local draft report files if they were staged by an earlier `git add`:

```bash
git restore --staged submission/OrderFlow_Applied_Software_Project_Report_DRAFT.docx submission/OrderFlow_Applied_Software_Project_Report_DRAFT.pdf 2>/dev/null || true
```

Review `git status --short`, `git diff --cached --check`, and `git diff --cached --stat`. Never commit `.env`, passwords, JWTs, mock webhook secrets, SQL dumps, or the draft Word/PDF report. The final PDF and final submission ZIP must be produced after code/CI verification, real benchmark results, and the end-of-project evidence session.

## What cannot be honestly completed by this code overlay alone

- A new green GitHub Actions run for the exact reliability commit (requires push and hosted run).
- Performance result values (must be measured by the benchmark).
- AWS deployment (requires account access, region/network choices, and approval of any billable resources).
- Real payment-provider reconciliation. The current payment provider is deliberately mock; the periodic reconciliation reporter is read-only because there is no real provider lookup contract.
- The final report screenshots/evidence, PDF export, previous Lambda deployment review, and learning walkthrough. Screenshots are intentionally deferred until the end.
