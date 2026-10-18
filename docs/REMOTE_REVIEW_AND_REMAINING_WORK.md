# Remote repository review and residual-work register

**Reviewed branch:** `main`  
**Remote HEAD at review:** `8ffd04e` — `test: cover concurrent orders and payment callbacks`  
**Review is of the GitHub tree, not a clone of the user's current local filesystem.**

## Present already at the review baseline

- `cmd/api/main.go`, config loader and PostgreSQL connection pool.
- Gin router, auth/JWT middleware, auth, product, order, payment and inventory handlers.
- Domain models for product, order, payment and inventory/movements.
- Four repositories and four services for product/order/payment/inventory.
- Six pairs of up/down SQL migration files.
- HTTP/service unit tests and two PostgreSQL integration test files, including concurrent order/payment behavior.
- Docker Compose PostgreSQL service, `.env.example`, test DB migration/runner scripts.

A `README.md` was not present at the reviewed HEAD. There was no OpenAPI/Swagger spec, Postman collection, GitHub Actions CI workflow, app Dockerfile, or committed HLD/LLD/report document in the tree at that time.

## Items supplied by this documentation/tooling bundle

- Root `README.md` with setup, API index, test instructions and embedded HLD/LLD.
- `docs/openapi.yaml` (OpenAPI 3.0), a Postman collection, architecture diagrams, DTO/repository patterns guide, project report, AWS deployment plan and screenshot checklist.
- Dockerfile and `.dockerignore`.
- GitHub Actions workflow that starts a dedicated PostgreSQL service, applies migrations, and runs tests/race/vet.
- Corrected local `.env.example` template containing the documented `JWT_SECRET` field as well as the mock payment secret.

These are proposed overlays. They become part of the repository only after the owner unzips, reviews, validates, commits and pushes them.

## Explicit residuals — do not mark completed from this bundle alone

| Priority | Residual | Why it remains |
|---|---|---|
| P1 | Validate this docs/tooling bundle locally and push it | Files have to be integrated into local tree and tests/CI reviewed |
| P1 | Collect screenshots/evidence | Real Postman execution and GitHub Actions run must be performed |
| P1/P2 | Confirm assessment rubric for Redis, cloud deployment, provider integration, cron | These are separate runtime features, not doc-only checkboxes |
| P2 | Redis product-cache integration and invalidation | Not present in the Go baseline; should be an isolated code milestone with tests |
| P2 | AWS deployment and infrastructure configuration | No AWS resources were provisioned; deployment guide is design only |
| P2 | Real payment provider (Stripe etc.) | Current callback/provider is mock only; requirements and credentials determine scope |
| P2 | Scheduled reconciliation worker | Current reconciliation runs only when admin calls its endpoint |
| P3 | DTO package refactor | Optional; current private request structs and domain models already serve the layers |
| P3 | Increase direct repository/config/database unit coverage | Integration tests cover important flows but are not a substitute for every unit test |
| P3 | Report PDF/final submission ZIP after screenshots | Markdown report can be reviewed now; final evidence should be included before freezing the submission artifact |
| P3 | Revisit prior Lambda deployment submission | Separate earlier submission remains on tracker |
| P3 | Post-submission code walkthrough | User asked to understand the implementation after finishing the submission |

## Remote review observations to check while finalizing

- Payment initiation currently returns HTTP 201 when the service returns an existing idempotent payment. This can be defensible as current behavior, but conventional semantics could return 200 for an existing resource; do not change without adding an explicit creation/existing result and tests.
- The mock callback requires HMAC on the raw body. Postman pre-request script in the collection calculates this for the final raw body; confirm the script on your Postman version and ensure the same secret is used locally.
- The CI workflow runs migrations into a separate `orderflow_test` service database, never the dev `orderflow` database.
- Current API doesn't mount Swagger UI at `/docs`; the OpenAPI YAML can be loaded in a separate Swagger UI container described in README.
