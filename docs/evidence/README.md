# OrderFlow evidence and screenshot guide

This folder uses real screenshots only. Do not include bearer tokens, passwords, database URLs, webhook secrets/signatures, or personal account details. The same PNGs should be inserted into the final Word report after the final Postman run.

## Captured and suitable already

| Capture | File | Meaning / limitation |
|---|---|---|
| Combined health/readiness terminal check | `browser/01-health-readiness-http-200.png` | Shows `curl -i` returning HTTP 200 for `/health` and `/ready`, with JSON bodies `ok` and `ready`. |
| Prometheus exposition | `observability/01-request-metrics.png` | Shows request counters and latency histogram buckets at `/metrics`; the page itself does not calculate p95. |
| Initial local Go API log | `docker/02-local-server-log-order-201-initial.png` | Shows an earlier local `POST /orders` returning 201; supplementary historical evidence, not the final end-to-end workflow. |
| Docker Compose container health (cropped) | `initial-captures/01-docker-compose-ps-redacted-crop.png` | Cropped from the earlier terminal capture to exclude visible credentials; historical setup context only. |
| Initial migration output | `initial-captures/02-early-migration-only-not-final.png` | Shows only migration 000001 and the first three tables; historical only, not final schema proof. |
| Readiness browser body only | `initial-captures/03-readiness-body-only-duplicate.png` | Duplicate, less useful than the combined HTTP status-line capture. |

The original screenshot with `grep DATABASE_URL .env` visibly exposed a local connection URL and password. It is **not included** in the repository patch. If that password was reused outside a disposable local environment, rotate it. GitHub device authorization, Downloads-folder, Scaler form, and ChatGPT-preview screenshots are not application evidence and are not included.

| Public product listing (HTTP 200) | `postman/initial/08-product-list-public-200.png` | `GET /products?page=1&limit=20` returned 200 with one active product; endpoint is public. |

## Remaining screenshots: take these after final local test and green CI

| Sequence | Test to perform | Save as | Report placement |
|---|---|---|---|
| 1 | GitHub Actions run for the latest pushed evidence commit is green; show the commit SHA | `ci/01-github-actions-green.png` | Chapter 14 |
| 2 | Docker image build/smoke succeeds and both health/readiness are 200 | `docker/01-health-readiness.png` | Chapter 14 |
| 3 | Login/profile plus deliberate missing-token 401 and customer-token admin-route 403 | `postman/01-auth-profile-and-authorization.png` | Chapter 7 |
| 4 | ADMIN product create 201, duplicate SKU 409, update 200, deactivate 200 / inactive detail 404 | `postman/02-product-admin.png` | Chapter 7 |
| 5 | Inventory availability, stock receipt, adjustment and movement audit | `postman/03-inventory-audit.png` | Chapters 8–9 |
| 6 | Multi-product order 201, same-key replay 200, changed-payload reuse 409, detail/list and cancel 200 / repeat cancel 409 | `postman/04-order-lifecycle.png` | Chapters 8–9 |
| 7 | Payment initiate, signed SUCCESS, signed FAILED, invalid signature 401, amount/currency mismatch 422, callback replay | `postman/05-payment-callbacks.png` | Chapter 10 |
| 8 | Missed callback: set terminal result in mock ledger, run reconciliation, show recovered order and admin order/sales/inventory reports | `postman/06-reconciliation-and-reports.png` | Chapter 10 and reporting |
| 9 | Real product API benchmark output, show both cache-disabled and warm-cache variants | `performance/01-api-cache-benchmark.png` | Chapter 13 |
| 10 | Completed bounded order-creation/history service benchmark output | `performance/02-order-benchmarks.png` | Chapter 13; do not label these HTTP percentiles |
| 11 | `/metrics` and a representative response header `X-Request-ID` | `observability/02-request-id-postman.png` | Chapters 11–12 |
| 12 | Optional: Swagger UI successfully loads `docs/openapi.yaml` | `api-docs/01-swagger-ui.png` | Chapter 7 / appendix |
| 13 | Only if a real, approved AWS deployment is required and done: live health/readiness and redacted service status | `aws/01-live-health.png` | Chapter 15 |

## macOS capture instructions

1. Arrange the relevant browser/Postman/terminal window so the request name, method/path, HTTP status, meaningful response, and (for CI) actual SHA are readable.
2. Press **Command + Shift + 4**, then drag a tight crop.
3. Save the PNG in the exact repository path in the table.
4. Inspect the image and redact secrets before staging. Do not alter result content; cropping unrelated areas is fine.
5. Insert that same image into the corresponding report chapter, then export the PDF after page numbers/figure lists are refreshed.

## Architecture clarification

OrderFlow is a **modular monolith**, not a microservices system. The HTTP APIs are routes in one Go/Gin process. `internal/http`, `service`, `repository`, `domain`, `cache`, `worker`, and `observability` are modules/packages, not separate deployed services. PostgreSQL and optional Redis are infrastructure dependencies. Evidence should show APIs, transactional behavior, workers, DB/cache, and CI; do not invent service-to-service screenshots.

## Initial Postman captures and authorization diagnosis

User-supplied initial Postman screenshots are preserved under `docs/evidence/postman/initial/`. The login screenshot is redacted; the original displayed an active JWT and is intentionally not included. The 401/403 product-create screenshots are troubleshooting history, not successful CRUD evidence. See `docs/postman/POSTMAN_ADMIN_TOKEN_TROUBLESHOOTING.md`.

Important distinction: `GET /products` is public and expected to return `200`. `POST/PATCH/DELETE /products` require a valid ADMIN token. An empty `adminToken` produces `401 bearer token required`; a valid CUSTOMER token produces `403 insufficient permissions`. Run the collection's existing-ADMIN login request and verify `user.role == ADMIN` before admin operations.
