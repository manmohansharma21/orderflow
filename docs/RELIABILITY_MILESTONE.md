# Reliability milestone: migration, tests and commit

This is a repository overlay. Extract from the repository root. It adds migration 000007 and updates the existing domain/service/repository/HTTP worker code plus integration tests. It must be applied on top of the previously installed Redis/Docker milestone.

## Apply migration 000007 once to the existing local and integration databases

The script checks the exact database names (`orderflow` and `orderflow_test`) in the existing Docker container. It applies only migration 000007 and skips either database if the migration's identifying column already exists. Run it from the repository root:

```bash
chmod +x scripts/apply-reliability-migration.sh scripts/verify-residual-milestone.sh scripts/benchmark-product-cache.sh
./scripts/apply-reliability-migration.sh
```

Do not point `ORDERFLOW_TEST_DATABASE_URL` at a production database. This script does not rerun migrations 000001–000006. If the database/container names differ, stop and inspect them rather than editing the guard casually. GitHub Actions always applies the entire migration chain to a disposable PostgreSQL service database.

## Verification commands

One command applies migration 000007 only when needed, starts PostgreSQL/Redis, formats Go source, runs all tests, race tests, vet, and whitespace checks:

```bash
chmod +x scripts/*.sh
./scripts/verify-residual-milestone.sh
./scripts/docker-smoke-test.sh
```

For measured API latency and throughput evidence after tests and Docker smoke pass:

```bash
./scripts/benchmark-api-cache.sh
```

It records actual p50/p95/mean and throughput for `GET /products` with Redis disabled versus enabled, saving output to `docs/evidence/performance/`. It does not assume caching wins. The repository-layer comparison is also available via `./scripts/benchmark-product-cache.sh`, which uses `orderflow_test` and Redis DB 15.

## Review / commit

```bash
git status --short
git diff --stat
git diff --check
```

Stage the intended files after reviewing them. Do not stage `.env`, credentials, SQL backups or local report placeholders. Then commit/push and verify the newest Actions run corresponds to the new HEAD.
