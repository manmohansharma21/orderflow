# Reliability milestone — current status

The reliability code is already merged into `main` at `1515cb0`. GitHub Actions for that exact commit is green: https://github.com/manmohansharma21/orderflow/actions/runs/39160539244. Do not reapply earlier incremental source-code ZIPs.

## Included features

- Migration 000007: order idempotency, payment currency/failure reason and EXPIRED order state.
- Migration 000008: independent mock-provider ledger and recovery hooks.
- Multi-product transactional order creation with per-user idempotency and stable product lock order.
- Payment callback amount/currency validation, failure reason, success inventory consumption and replay-safe stock release/consumption.
- Safe expiry of old unpaid orders, admin order/sales/inventory reports, optional Redis cache, request IDs and `/metrics`.
- Integration/concurrency/recovery tests and CI verification.

## Verified commands from the latest supplied terminal output

```bash
./scripts/verify-residual-milestone.sh
./scripts/docker-smoke-test.sh
```

The first completed with tests, race integration checks and static checks passing. The Docker image built and returned successful `/health` and `/ready` responses. The exact pushed revision also has a green Actions run.

## Benchmark status

The local API benchmark is committed and shows Redis slower for this small workload. Preserve these actual results and do not call the cache an optimization win. The order-creation benchmark produced samples, but the order-history benchmark did not finish before Ctrl+C. Run it separately with a bounded duration only if the report/assignment needs the value:

```bash
export ORDERFLOW_TEST_DATABASE_URL='(set your private connection string to the database named orderflow_test)'
go test -run='^$' -bench='BenchmarkOrderHistoryList' -benchtime=100ms -benchmem -count=1 ./internal/integration/...
```

Never echo the connection string, and never use a development or production database for tests/benchmarks.

## Remaining delivery work

The source build itself is verified at `1515cb0`. Remaining tasks are documentation cleanup, a final genuine Postman/API screenshot session, optional bounded order-history measurement, Word report completion and PDF export, final source ZIP download from `main`, and confirming the final branch CI is green. AWS deployment remains conditional on rubric and cost approval; the mock provider is not a live gateway.
