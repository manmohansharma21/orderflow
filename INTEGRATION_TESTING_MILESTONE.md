# OrderFlow PostgreSQL Integration Test Milestone

This bundle adds database-backed integration tests for order creation, stock reservation, insufficient-stock rollback behavior, payment success/failure handling, callback replay, and payment idempotency.

## Safety model

- The tests use `ORDERFLOW_TEST_DATABASE_URL`, never `.env`'s `DATABASE_URL` directly.
- The Go test helper refuses to run against any database whose name is not exactly `orderflow_test`.
- Each test creates uniquely named user/product fixtures and cleans up only its own rows.
- The migration script applies migrations only to `orderflow_test` and refuses to rerun initial migrations when common application tables already exist.
- Never set `ORDERFLOW_TEST_DATABASE_URL` to your ordinary `orderflow` database or any production database.

## 1. Create the dedicated test database once

Run from the repository root. This connects to the existing `orderflow-postgres` container and creates only a separate database named `orderflow_test`:

```bash
docker exec orderflow-postgres psql -X -U orderflow -d postgres \
  -v ON_ERROR_STOP=1 \
  -c "CREATE DATABASE orderflow_test OWNER orderflow;"
```

If the database already exists, don't blindly create or drop it. Confirm it is the disposable integration-test database before proceeding.

If your container name differs, set `ORDERFLOW_POSTGRES_CONTAINER` for the migration script. The default is `orderflow-postgres`.

## 2. Apply the committed migrations to the test database

```bash
chmod +x scripts/apply-integration-migrations.sh scripts/run-integration-tests.sh
./scripts/apply-integration-migrations.sh
```

The script applies `migrations/*.up.sql` in lexical order, matching this repository’s filenames such as `000001_init_schema.up.sql`. If a migration fails partway through, stop and inspect the test database; don't rerun the initial migrations over a partially initialized schema. For a completely disposable test database, only recreate `orderflow_test` after verifying its name and contents. Do not drop or alter `orderflow`.

## 3. Build a test connection URL without printing the database password

The following reads `DATABASE_URL` from your local `.env`, verifies its current database path is `/orderflow`, and changes only the database name to `/orderflow_test`. The resulting URL is assigned to a shell variable and is not printed:

```bash
export ORDERFLOW_TEST_DATABASE_URL="$(python3 - <<'PY'
from pathlib import Path
from urllib.parse import urlsplit, urlunsplit

value = None
for line in Path('.env').read_text().splitlines():
    stripped = line.strip()
    if stripped.startswith('DATABASE_URL='):
        value = stripped.split('=', 1)[1].strip().strip('"').strip("'")
        break
if not value:
    raise SystemExit('DATABASE_URL was not found in .env')
parts = urlsplit(value)
if parts.path.rstrip('/') != '/orderflow':
    raise SystemExit('Expected .env DATABASE_URL to use database /orderflow; inspect it and adapt this command manually without displaying credentials')
print(urlunsplit((parts.scheme, parts.netloc, '/orderflow_test', parts.query, parts.fragment)))
PY
)"
```

If your development database is not named `orderflow`, inspect the connection settings locally and adapt the command carefully. Do not paste credentials or the generated URL into chat.

## 4. Run the tests

Confirm that the variable is set without printing its value:

```bash
printf 'ORDERFLOW_TEST_DATABASE_URL: %s\n' "${ORDERFLOW_TEST_DATABASE_URL:+set}"
```

If it prints blank, run the export command in Step 3 again in the same terminal. Then run:

```bash
./scripts/run-integration-tests.sh
```

To run the suite directly, keep `ORDERFLOW_TEST_DATABASE_URL` set and execute:

```bash
go test -count=1 ./internal/integration/...
```

Without the URL, the tests in `internal/integration` skip safely under `go test ./...`; the dedicated runner fails fast rather than reporting a misleading successful integration run.

## Coverage in this batch

- Creating an order reserves the expected stock and inserts its order item and `ORDER_RESERVED` audit movement.
- Insufficient stock leaves no partial order or reservation.
- A signed-callback service success transitions the order to `CONFIRMED` while retaining its reservation.
- A failure callback marks the payment/order failed and releases stock exactly once when replayed.
- Starting payment twice with the same idempotency key returns the same payment ID and leaves one payment row.

## Deliberate limits

This batch does not yet test simultaneous concurrent requests under load, real Stripe/webhook network behavior, or migration idempotency on a partially initialized database. Those are separate follow-up tasks. It also does not apply tests to the live development database.
