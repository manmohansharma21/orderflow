# Final Postman screenshot runbook — OrderFlow

Use a disposable local database and demo users/products. Import `docs/postman/OrderFlow.postman_collection.json` into Postman and set `baseUrl=http://localhost:8080`. Use only local test credentials. Store tokens and `MOCK_PAYMENT_WEBHOOK_SECRET` in local Postman variables; never commit an exported collection with populated secrets.

## Start once

Terminal A (keep open):

```bash
docker compose up -d postgres redis
go run ./cmd/api
```

Terminal B: `curl -i http://localhost:8080/health`, `curl -i http://localhost:8080/ready`, `curl -i http://localhost:8080/metrics`. Expected HTTP 200.

## Capture order (one screenshot per logical flow)

### A. Authentication and authorization

1. `GET /profile` without Authorization -> **401**. Capture response status/body (do not show token).
2. `POST /auth/register` with a new local email/password -> **201** (a reused email may return **409**).
3. `POST /auth/login` -> **200**; allow the collection script to save the token.
4. `GET /profile` with customer bearer token -> **200**.
5. `POST /products` without token -> **401**. Repeat with customer token -> **403**. Use an ADMIN token for actual product operations.

Save one clean screen as `docs/evidence/postman/01-auth-profile-and-authorization.png`.

### B. Product administration

1. ADMIN `POST /products` -> **201**. Use a unique SKU and non-negative `price_paise`/`initial_stock`.
2. Repeat with same SKU -> **409 Conflict**.
3. ADMIN `PATCH /products/{id}` -> **200**.
4. ADMIN `DELETE /products/{id}` -> **200** (deactivation).
5. `GET /products/{id}` after deactivation -> **404**. Use another fresh product for the order flow.

Save as `docs/evidence/postman/02-product-admin.png`.

### C. Inventory receipt and audit

1. `GET /inventory/{product_id}` -> **200**.
2. ADMIN `POST /admin/inventory/receipts` with product_id, quantity, reason -> **200**.
3. ADMIN `POST /admin/inventory/adjustments` with product_id, quantity_delta, reason -> **200**.
4. ADMIN `GET /admin/inventory/movements?product_id={id}&page=1&limit=20` -> **200**.

Save availability + a receipt/adjustment + movement rows as `docs/evidence/postman/03-inventory-audit.png`.

### D. Multi-product order and idempotency

1. `POST /orders` without bearer token -> **401**.
2. With customer token, send `items` with 2 distinct product IDs and quantities, plus `Idempotency-Key: orderflow-demo-unique-001` -> **201 Created**, status `PENDING`. Use the server-calculated total, not a client-provided total.
3. Send the exact same body and same key again -> **200 OK** and the same order ID.
4. Reuse the same key but alter a quantity -> **409 Conflict**.
5. Customer `GET /orders` and `GET /orders/{id}` -> **200**. Use another customer token for the same order detail to show owner-scope behavior -> **404**.
6. `POST /orders/{id}/cancel` while eligible -> **200**; repeat cancellation -> **409**. Confirm inventory movements show one release only.

For a successful-payment path, create a **new** order and key. Save as `docs/evidence/postman/04-order-lifecycle.png`.

### E. Payment callback success/failure

1. Customer `POST /orders/{id}/payments` with a unique `Idempotency-Key` -> **201**.
2. For signature failure, send deliberately invalid `X-Mock-Signature` -> **401**.
3. Send a correctly HMAC-signed callback over the exact raw JSON body with correct `payment_id`, terminal `outcome`, unique `event_id`, matching `amount_paise`, and matching currency -> **200**. For SUCCESS, order becomes `CONFIRMED` and stock is consumed. For FAILED, include a failure reason; order becomes `FAILED` and its reservation is released once.
4. Replay the same callback/event -> expected idempotent 200 response/no duplicate movement.
5. For a mismatch test, create a separate pending payment and sign a callback with a changed amount/currency -> **422**; verify state/stock did not change.

Save as `docs/evidence/postman/05-payment-callbacks.png`. Keep the signing secret off screen.

### F. Missed callback recovery and reports

Run the collection folder **Missed callback reconciliation demo** in order: create recovery order, initiate payment, set a terminal state in the separate simulated-provider ledger without a callback, run `/admin/payments/reconciliation/run?older_than_minutes=0&limit=100`, then get the order and confirm `CONFIRMED` after SUCCESS recovery. Then call admin reports: `GET /admin/reports/orders`, `GET /admin/reports/sales?from=2026-10-01&to=2026-10-31`, and `GET /admin/reports/inventory?low_stock_threshold=5`; each should return **200** with ADMIN bearer token.

Save reconciliation + report evidence as `docs/evidence/postman/06-reconciliation-and-reports.png`.

## Also capture

- Current CI green run showing commit SHA `3df0139` or the newest SHA: `docs/evidence/ci/01-github-actions-green.png`. If committing screenshots afterward, the screenshot is a record of the parent commit's green check; confirm the evidence-only commit also has green CI.
- Fresh `scripts/docker-smoke-test.sh` output with image build and `/health`/`/ready`: `docs/evidence/docker/01-health-readiness.png`.
- Run `./scripts/benchmark-api-cache.sh` separately; save output and screenshot under `docs/evidence/performance/01-api-cache-benchmark.png`. The current result was slower with Redis; preserve that honest finding.
- Run the bounded service benchmark with `./scripts/benchmark-order-flows.sh` by itself; capture its completed output as `docs/evidence/performance/02-order-benchmarks.png`.
- Optional Swagger UI: load `docs/openapi.yaml` in Swagger UI and capture the rendered contract.

Do not capture microservices views: this project is a modular monolith. Do not claim live AWS deployment without actual resources and explicit assessment requirement.
