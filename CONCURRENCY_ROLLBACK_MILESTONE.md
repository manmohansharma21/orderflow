# OrderFlow Concurrency and Rollback Test Milestone

This bundle adds PostgreSQL integration tests to the existing `internal/integration` package. It relies on the existing fixture helpers from `orderflow_postgres_test.go` and must be applied to the same branch that already contains commit `34c04f0`.

## Safety requirements

- Run only with `ORDERFLOW_TEST_DATABASE_URL` pointing to the dedicated database named exactly `orderflow_test`.
- Do not point the variable at `.env`'s normal `orderflow` database or at production.
- Do not rerun `scripts/apply-integration-migrations.sh`; migrations were already applied to `orderflow_test` for the previous milestone.
- The tests create uniquely named fixture rows and clean up only their own data.

## Install

From the repository root, extract this ZIP and preserve paths. In Finder, double-click the ZIP and move the folder contents into the repository, or run:

```bash
unzip -o ~/Downloads/ORDERFLOW_CONCURRENCY_ROLLBACK_TESTS.zip -d .
```

Then run:

```bash
gofmt -w internal/integration/orderflow_concurrency_test.go
go test -count=1 ./internal/integration/...
go test -race -count=1 ./internal/integration/...
go test ./...
go vet ./...
git diff --check
git status --short
```

Keep the terminal in which `ORDERFLOW_TEST_DATABASE_URL` was exported. `go test ./...` skips database integration tests if the variable is absent, so the explicit integration command is important.

## New coverage

- Concurrently tries to place ten one-unit orders against five units of stock; expects exactly five successful orders, five insufficient-stock responses, and matching reservation movements.
- Triggers an error after stock reservation would have occurred (the requested user does not exist); expects the transaction to roll back the reservation and leave no order or audit movement.
- Initiates payment concurrently ten times with the same idempotency key; expects one payment record and one returned payment ID.
- Replays a failed payment callback concurrently ten times; expects stock to be released exactly once and only one `ORDER_RELEASED` movement.

## If a test fails

Do not drop databases or rerun migrations automatically. Preserve the failing output, confirm `ORDERFLOW_TEST_DATABASE_URL` is set without printing its value, and inspect only the dedicated `orderflow_test` database.
