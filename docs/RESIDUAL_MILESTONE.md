# Residual implementation milestone: Redis, worker, Docker smoke test

This overlay adds optional Redis product caching, TTL and generation-based invalidation; a read-only scheduled payment reconciliation reporter; and CI checks for PostgreSQL, Redis, Docker build and container readiness.

## Install safely

Unzip this bundle into the repository root. It deliberately does not include `.env`, database dumps, or credentials. It replaces the product service, configuration, router, main composition root, Docker Compose, CI workflow, README/report text, and HLD source.

## Verify with one command

Keep `ORDERFLOW_TEST_DATABASE_URL` set to the existing dedicated database named exactly `orderflow_test`, as documented in `INTEGRATION_TESTING_MILESTONE.md`, then run:

```bash
chmod +x scripts/verify-residual-milestone.sh scripts/docker-smoke-test.sh
./scripts/verify-residual-milestone.sh
```

The verification script starts the local Redis service, runs `go mod tidy` (this updates `go.sum`), formats the added/updated Go files, runs all tests, race-enabled PostgreSQL integration tests, `go vet`, and `git diff --check`. Redis checks run against DB 14 and DB 15 of the project's disposable local Redis service. Do not use those DBs for production data.

Then run the container build and live API smoke test:

```bash
./scripts/docker-smoke-test.sh
```

This uses `.env` locally without printing its values, builds the image, starts an ephemeral container, and checks `/health` and `/ready`. It expects your development PostgreSQL schema/migrations to be applied. It will start Compose PostgreSQL/Redis services, but will not drop or recreate any database.

## Commit and push

After the commands above all pass, inspect staged/modified files and check that `.env`, tokens, database dumps or real secrets are not staged. Add the exact paths from your `git status --short`, including `go.mod` and `go.sum` if changed. Then commit and push. The post-push green GitHub Actions run is required before calling this milestone verified.

## Current capstone-spec gaps that still need a code milestone

Before final submission, verify/fix multi-product orders, order-creation idempotency, callback amount/currency and failure-reason validation, stock consumption after payment success, automated expiry release, required admin reporting, and benchmark evidence against the project spec. These are intentionally not declared complete by adding Redis or a reconciliation-reporting worker.

## Behavior and limitations

Redis cache failures fall back to PostgreSQL. Product writes increment a generation key after a successful database mutation; old generation keys expire by TTL. If Redis invalidation fails during an outage, old values may remain visible until their TTL expires, and the service logs a warning.

The reconciliation worker is read-only. It logs stale pending payments and inconsistent order/payment states from the existing report; it does not charge, refund, or mutate status. `RECONCILIATION_INTERVAL_SECONDS=0` disables the worker.

`go-redis/v9` is set to `v9.23.0`, which is compatible with this repository's Go 1.26 module directive. `go mod tidy` must refresh `go.sum` locally before commit.
