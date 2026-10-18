# Redis product caching and simulated-provider reconciliation

## Product cache

The service caches active product list/search responses and product detail reads. PostgreSQL remains authoritative. `internal/cache/redis_product_cache.go` implements the Redis adapter; `internal/service/product_cache.go` defines the cache contract; product service methods use cache-aside with a generation key and TTL.

After product create/update/deactivation, the service increments `orderflow:products:cache-version`. Reads build keys in the current generation. Old-generation entries become inaccessible to new reads and expire by TTL. Cache errors fail open to PostgreSQL. Environment: `REDIS_ENABLED`, `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB` (0–15), `PRODUCT_CACHE_TTL_SECONDS` (default 60). Redis is disabled by default because the measured small local product-list workload was slower with warm cache.

Actual local API results (300 requests per variant, 10 concurrent workers, Apple M2 / Go 1.27.2 / Docker Desktop 29.8.2):

| Metric | Redis disabled | Warm Redis |
|---|---:|---:|
| Successful requests | 300/300 | 300/300 |
| P50 | 2.480 ms | 4.447 ms |
| P95 | 4.270 ms | 13.966 ms |
| Mean | 3.315 ms | 5.702 ms |
| Throughput | 2,821.81 req/s | 1,697.92 req/s |

This is a valid negative experiment for a small/local dataset, not a universal conclusion that Redis always loses. Do not claim a speedup. Use a larger, representative workload before enabling cache by default.

## Reconciliation and expiry

`internal/worker/payment_reconciliation.go` periodically asks the payment service to recover known terminal outcomes from the separate `mock_provider_transactions` ledger, then logs remaining stale/inconsistent payment/order states. Recovery reuses the same callback path for amount/currency validation, idempotency, order-state checks, stock mutation and audit movements. `RECONCILIATION_INTERVAL_SECONDS` defaults to 300; zero disables the worker. `RECONCILIATION_OLDER_THAN_MINUTES` defaults to 15.

The simulator lets an ADMIN account (outside `APP_ENV=production`) set SUCCESS/FAILED provider state without invoking the callback, so a missed callback can be demonstrated deterministically. Routes:

- `POST /admin/mock-provider/payments/{id}/outcome` — set the simulated provider-side terminal outcome.
- `POST /admin/payments/reconciliation/run?older_than_minutes=0&limit=100` — immediately recover known outcomes for a local demo.
- `GET /admin/payments/reconciliation?older_than_minutes=15` — report remaining inconsistent/stale states.

The scheduled order-expiry worker marks old PENDING orders EXPIRED and releases stock only if no PENDING or SUCCESS payment exists. It uses database locking and audit movements; unresolved payment attempts are deliberately left for reconciliation.

These mechanisms are for the mock provider. They do not query Stripe or any live provider, do not move real money, and do not claim production gateway recovery. A live provider would need an authoritative status-query/webhook API and an explicit remediation policy.

## Test configuration

- `ORDERFLOW_TEST_DATABASE_URL` must target a database whose name is exactly `orderflow_test`.
- `ORDERFLOW_REDIS_TEST_ADDR` points to disposable Redis; CI provides a fresh service. Redis tests skip locally when the variable is not present.
- Migrations 000007 and 000008 have already been applied to the user's local `orderflow` and `orderflow_test` databases. Re-run the project's guard scripts; do not replay the initial migration loop against populated databases.

## Local verification

```bash
./scripts/verify-residual-milestone.sh
./scripts/docker-smoke-test.sh
./scripts/benchmark-api-cache.sh
```

Run each command separately and wait for it to finish. The previous order-history benchmark attempt was interrupted, so its number is not reportable. A bounded repeat can be run with:

```bash
go test -run='^$' -bench='BenchmarkOrderHistoryList' -benchtime=100ms -benchmem -count=1 ./internal/integration/...
```
