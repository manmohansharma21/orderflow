# Final screenshot checklist — capture only after verification

This checklist is intentionally an end-stage task. The source commit `1515cb0` has a green CI run, and local tests/Docker smoke passed according to the terminal output supplied. Capture evidence after the final API/benchmark runs and before the final PDF export. If any further code is committed, use the newest successful CI run/SHA instead.

## Capture order, exact filename and report placement

| Step / when | File to save | What must be visible | Report location |
|---|---|---|---|
| 1. After confirming green Actions for the final HEAD | `docs/evidence/ci/01-github-actions-green.png` | Completed successful workflow and commit SHA `1515cb0` (or newer final SHA) | Chapter 14: Docker/CI |
| 2. After local Docker smoke passes | `docs/evidence/docker/01-health-readiness.png` | Docker build/smoke plus `/health` and `/ready` success | Chapter 14 |
| 3. After Postman auth/profile request | `docs/evidence/postman/01-auth-profile.png` | `GET /profile`, status and safe JSON; no Authorization token shown | Chapter 7: REST APIs |
| 4. After product admin create/update and duplicate SKU | `docs/evidence/postman/02-product-admin.png` | Admin response/status and duplicate SKU conflict | Chapter 7 |
| 5. After inventory receipt/adjustment/movement history | `docs/evidence/postman/03-inventory-audit.png` | Stock numbers, movement type, reason and status | Chapters 8–9 |
| 6. After multi-item order/replay/conflict/cancel | `docs/evidence/postman/04-order-lifecycle.png` | Order ID/status/item snapshots, replay result and cancellation evidence | Chapters 8–9 |
| 7. After successful and failed payment callback plus replay | `docs/evidence/postman/05-payment-callbacks.png` | SUCCESS consumes stock; FAILED releases stock once; bad signature receives 401 (separate crops if clearer) | Chapter 10 |
| 8. After missed-callback recovery and admin reports | `docs/evidence/postman/06-reconciliation-and-reports.png` | Simulated provider outcome set without callback, reconciliation recovery count, and order/sales/inventory reports | Chapter 10 and feature/report chapter |
| 9. After latest HTTP cache API benchmark | `docs/evidence/performance/01-api-cache-benchmark.png` | Actual no-cache/warm-cache metrics including negative result | Chapter 13 |
| 10. After completing bounded order-history benchmark | `docs/evidence/performance/02-order-benchmarks.png` | Command and completed actual benchmark output | Chapter 13 |
| 11. Only after an actual AWS deployment, if required and approved | `docs/evidence/aws/01-live-health.png` | Live endpoint/health without sensitive data | Chapter 15 |

## macOS capture instructions

1. Create the directory if needed: `mkdir -p docs/evidence/{ci,docker,postman,performance,observability,aws}`.
2. Arrange the relevant app result so the request/command and response are legible.
3. Press **Command + Shift + 4**, then drag to select only the relevant area. macOS saves a PNG to the configured screenshot location; move it to the exact path above.
4. For Postman, show the request name, method/path, HTTP status and response JSON. Use a dedicated demo account. Do not leave the Authorization tab, token, password, webhook secret or `.env` visible. If a token appears in the response, permanently redact it in Preview before committing the image.
5. For terminal output, enlarge the font and keep the full success lines visible, not a partial command interrupted with Control+C.
6. Inspect the saved PNG, then paste that same PNG into the corresponding yellow placeholder in the Word report. Add its figure caption and refresh the list of figures/table of contents after page layout stops changing.

## Postman run order

Import `docs/postman/OrderFlow.postman_collection.json` into Postman. Use local demo accounts only. Run requests in collection order and the separate **Missed callback reconciliation demo** folder. Register/login a customer; use an explicitly promoted local ADMIN account for protected requests. Use the collection's variable capture and same idempotency key for replay tests. Never export or commit a collection with populated access tokens/secrets.

## Automated evidence

Before screenshots, run and preserve the real results of:

```bash
./scripts/verify-residual-milestone.sh
./scripts/docker-smoke-test.sh
./scripts/benchmark-api-cache.sh
```

Run the long order-history benchmark separately with a bounded duration if required:

```bash
go test -run='^$' -bench='BenchmarkOrderHistoryList' -benchtime=100ms -benchmem -count=1 ./internal/integration/...
```

The previous history benchmark attempt was interrupted before completion; do not invent a result. Do not take AWS screenshots unless an actual approved deployment exists. The target AWS architecture diagram is not proof of a live deployment.
