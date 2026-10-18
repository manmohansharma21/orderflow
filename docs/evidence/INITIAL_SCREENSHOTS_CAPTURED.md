# Initial OrderFlow screenshot evidence — captured 19 October 2026

These are genuine local captures provided during the verification session. They have been renamed into stable repository paths; raw originals are retained in `initial-captures/`. The editable Word evidence-review draft inserts these images into Chapter 14. They are **initial local evidence**, not a substitute for the final Postman and CI screenshots.

## Captured now

| Status | File | What it supports | Limitation |
|---|---|---|---|
| Captured | `browser/01-health-endpoint.png` | Browser response body for `GET /health`: `{"status":"ok"}` | Browser screenshot does not show HTTP status; supplied `curl -i` output separately showed HTTP 200. |
| Captured | `browser/02-readiness-endpoint.png` | Browser response body for `GET /ready`: `{"status":"ready"}` | Browser screenshot does not show HTTP status; supplied `curl -i` output separately showed HTTP 200. |
| Captured | `observability/01-request-metrics.png` | `/metrics` request counters and latency histogram buckets | In-memory local sample, not a production dashboard; it does not calculate P95 directly. |
| Captured as supplementary | `docker/02-local-server-log-order-201-initial.png` | Development server startup and a `POST /orders` request returning HTTP 201 | Earlier terminal log in Gin debug mode; supporting evidence only, not a full Postman order response. |

## Still to capture in the final session

1. `ci/01-github-actions-green.png` — green Actions run for the **latest commit SHA actually submitted**.
2. `docker/01-health-readiness.png` — Docker image build/smoke output with `/health` and `/ready`. Browser-only images do not prove the Docker build.
3. `postman/01-auth-profile.png` and `postman/02-product-admin.png`.
4. `postman/03-inventory-audit.png` and `postman/04-order-lifecycle.png` (include multi-item order, idempotent replay/conflict and cancellation).
5. `postman/05-payment-callbacks.png` and `postman/06-reconciliation-and-reports.png` (success/failure, bad signature, missed-callback recovery and reports).
6. `performance/01-api-cache-benchmark.png` — actual output for both cache modes; report the Redis-slower result honestly.
7. `performance/02-order-benchmarks.png` — completed bounded service-level benchmark output, if a screenshot is required.
8. Optional: `api-docs/01-swagger-ui.png` to demonstrate the OpenAPI contract loaded in Swagger UI.
9. AWS evidence is optional unless the official rubric explicitly requires a live deployment. The AWS diagram is target architecture, not proof of a deployment.

## Security and quality rules

- Never capture `.env`, database URLs, JWTs, bearer tokens, webhook secrets, passwords or unrelated account details.
- In Postman, show method, endpoint, HTTP status and response fields; redact `access_token`.
- Save original PNGs under `docs/evidence/` and use those same files in the Word report.
- Do not create screenshots for microservices: OrderFlow is a modular monolith, not a microservices deployment. Show its real Go API routes, PostgreSQL/Redis dependencies, workers, tests and CI.
