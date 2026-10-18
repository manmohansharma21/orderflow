# Redis and scheduled reconciliation

## Product cache

The service caches active product list/search responses and product detail reads. PostgreSQL remains authoritative. Cache errors, misses, malformed JSON, or a Redis server unavailable at API startup do not prevent product endpoints from using PostgreSQL.

Cache keys include a generation obtained from `orderflow:products:cache-version`. On successful create/update/deactivation, the service increments that generation. This makes old cached list/detail keys inaccessible to subsequent reads without scanning/deleting every old key. Old generations expire by `PRODUCT_CACHE_TTL_SECONDS` (default 60 seconds). If invalidation fails, the service logs a warning; old values can persist only until TTL or Redis becomes available, so check warnings operationally.

Environment variables: `REDIS_ENABLED`, `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB` (0–15), `PRODUCT_CACHE_TTL_SECONDS`. Redis is optional; set `REDIS_ENABLED=false` to disable it.

## Scheduled reconciliation

`internal/worker/payment_reconciliation.go` periodically checks the simulated-provider ledger for known terminal outcomes on still-pending payments, applies them through the same callback transition path, then logs remaining stale/inconsistent records from `PaymentService.ReconciliationReport`. `RECONCILIATION_INTERVAL_SECONDS` defaults to 300; setting it to 0 disables the worker. `RECONCILIATION_OLDER_THAN_MINUTES` defaults to 15.

The recovery logic is for the simulated provider only. It does not query a live gateway, initiate charges, or refund money. Production provider reconciliation must rely on the authoritative provider API/webhooks and an explicit remediation workflow. The development-only ADMIN endpoint can simulate a provider result without delivering the callback; it returns 404 when `APP_ENV=production`.

## Test configuration

- PostgreSQL tests: `ORDERFLOW_TEST_DATABASE_URL` must point to database `orderflow_test`.
- Redis tests: `ORDERFLOW_REDIS_TEST_ADDR` must point to a disposable test Redis instance. Adapter tests use DB 14; PostgreSQL+Redis product invalidation tests use DB 15. CI starts a fresh Redis service and sets this address.
- Tests skip Redis-specific cases if the env variable is missing; the CI workflow sets it.

## Local commands

```bash
docker compose up -d postgres redis
# Ensure .env points to your PostgreSQL on localhost:5433, then start the Go API.
go mod tidy
go test -count=1 ./...
go test -race -count=1 ./internal/integration/...
go vet ./...
docker build -t orderflow:local .
```


## Reliability additions

The `000007_order_payment_reliability` migration introduces per-customer order idempotency columns/index, payment currency and failure reason, and the `EXPIRED` order status. The order worker expires aged unpaid orders only if they have no pending/successful payment. It does not release inventory while a payment status is unresolved. The scheduled payment reconciliation worker currently emits a read-only report of stale or inconsistent records; it does not claim provider-side recovery because the mock provider has no independent status-lookup interface.

Successful callbacks now consume both on-hand and reserved stock and write `ORDER_FULFILLED` movements. External callback bodies must include amount and currency, and failed callbacks must include a meaningful failure reason.
