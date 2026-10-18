# OrderFlow evidence and screenshot guide

Use genuine screenshots only. Do not include bearer tokens, passwords, webhook signatures, database URLs, `.env`, secrets, or unrelated account details.

## Captured already

| Capture | File | Status and use |
|---|---|---|
| Local health response | `browser/01-health-endpoint.png` | Captured; response body shows `{"status":"ok"}` |
| Local readiness response | `browser/02-readiness-endpoint.png` | Captured; response body shows `{"status":"ready"}` |
| HTTP metrics exposition | `observability/01-request-metrics.png` | Captured; displays request counters and histogram buckets |
| Initial Go server log | `docker/02-local-server-log-order-201-initial.png` | Supplementary only; shows local development startup and one `POST /orders` with HTTP 201 |

The terminal output supplied on 19 October separately showed `curl -i` returning HTTP 200 for `/health` and `/ready`. The browser screenshots show the response bodies but not the status line. Raw image originals are preserved under `initial-captures/`. See [`INITIAL_SCREENSHOTS_CAPTURED.md`](INITIAL_SCREENSHOTS_CAPTURED.md).

## Still to capture after final verification

| Capture after | Exact filename | Report placement |
|---|---|---|
| Latest pushed HEAD has a green Actions run | `ci/01-github-actions-green.png` | Chapter 14; show final submitted SHA |
| Local Docker build/smoke script succeeds | `docker/01-health-readiness.png` | Chapter 14; include build and `/health`/`/ready` |
| Customer login/profile succeeds | `postman/01-auth-profile.png` | Chapter 7; show profile result, not token |
| Product create/update and duplicate SKU test run | `postman/02-product-admin.png` | Chapter 7 |
| Inventory availability, receipt/adjustment and movements run | `postman/03-inventory-audit.png` | Chapters 8–9 |
| Multi-product order, idempotent replay/conflict and cancellation run | `postman/04-order-lifecycle.png` | Chapters 8–9 |
| Signed payment success/failure, invalid signature and replay run | `postman/05-payment-callbacks.png` | Chapter 10 |
| Missed callback recovered and admin reports run | `postman/06-reconciliation-and-reports.png` | Chapter 10/reporting |
| API cache benchmark run | `performance/01-api-cache-benchmark.png` | Chapter 13; include real workload and negative cache result |
| Bounded service benchmarks | `performance/02-order-benchmarks.png` | Chapter 13; distinguish service ns/op from HTTP P95 |
| `/metrics` and `X-Request-ID` verified with a representative request | `observability/02-request-id-postman.png` | Chapters 11–12 |
| Swagger UI loads `docs/openapi.yaml` (optional) | `api-docs/01-swagger-ui.png` | Chapter 7/appendix |
| Real approved AWS deployment, only if required | `aws/01-live-health.png` | Chapter 15; otherwise keep absent |

## macOS screenshot steps

1. Arrange the relevant result in a readable window.
2. Press **Command + Shift + 4** and drag to crop.
3. Save the PNG with the exact path listed above.
4. Inspect it before committing; redact sensitive data if needed.
5. Insert that same PNG into the matching Word report chapter.

The API-level product-list benchmark is committed under `performance/benchmark-results-20261018T191309Z.txt`. The warm Redis path was slower for this local workload. The completed bounded service benchmark output is `performance/order-flow-benchmarks-20261018T202650Z.txt`: order creation 1,848,705 ns/op (5,892 B/op, 112 allocs/op) and order history 551,748 ns/op (19,837 B/op, 312 allocs/op). These are single-sample local service/repository-path measurements, not HTTP P95 values.

## No microservices screenshots

OrderFlow is a modular monolith. Evidence should show its actual API routes, Go package layers, PostgreSQL/Redis dependencies, workers/reconciliation, tests and CI. Do not imply the application is split into separately deployed microservices.
