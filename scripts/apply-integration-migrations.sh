#!/usr/bin/env bash
set -euo pipefail

CONTAINER_NAME="${ORDERFLOW_POSTGRES_CONTAINER:-orderflow-postgres}"
DB_NAME="orderflow_test"

if ! command -v docker >/dev/null 2>&1; then
  echo "ERROR: Docker CLI not found." >&2
  exit 1
fi
if ! docker exec "$CONTAINER_NAME" psql -X -U orderflow -d postgres -Atqc 'SELECT 1' >/dev/null 2>&1; then
  echo "ERROR: Could not connect to PostgreSQL in container '$CONTAINER_NAME'. Start the container and verify its name." >&2
  exit 1
fi

existing_tables="$(docker exec "$CONTAINER_NAME" psql -X -U orderflow -d "$DB_NAME" -Atqc \
  "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public' AND table_name IN ('users','products','orders','payments','inventory_movements')" 2>/dev/null || true)"
if [[ -z "$existing_tables" ]]; then
  echo "ERROR: Database '$DB_NAME' does not exist or is not accessible." >&2
  echo "Create only the dedicated test database first; do not point this script at your development database." >&2
  exit 1
fi
if [[ "$existing_tables" != "0" ]]; then
  echo "ERROR: '$DB_NAME' already has application tables; refusing to reapply the initial migrations." >&2
  echo "For a disposable test DB, review INTEGRATION_TESTING_MILESTONE.md before recreating this database." >&2
  exit 1
fi

shopt -s nullglob
migrations=(migrations/*.up.sql)
if [[ ${#migrations[@]} -eq 0 ]]; then
  echo "ERROR: no migrations/*_up.sql files found. Run this script from the repository root." >&2
  exit 1
fi

for migration in "${migrations[@]}"; do
  echo "Applying $migration to $DB_NAME ..."
  docker exec -i "$CONTAINER_NAME" psql -X -v ON_ERROR_STOP=1 -U orderflow -d "$DB_NAME" < "$migration"
done

echo "Applied ${#migrations[@]} migrations to $DB_NAME."
