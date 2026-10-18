#!/usr/bin/env bash
set -euo pipefail
container="${ORDERFLOW_POSTGRES_CONTAINER:-orderflow-postgres}"
command -v docker >/dev/null || { echo 'ERROR: Docker is required.' >&2; exit 2; }
docker inspect "$container" >/dev/null 2>&1 || { echo "ERROR: PostgreSQL container '$container' is not running." >&2; exit 2; }
for db in orderflow orderflow_test; do
  actual="$(docker exec "$container" psql -X -U orderflow -d "$db" -Atqc 'SELECT current_database()' 2>/dev/null || true)"
  if [[ "$actual" != "$db" ]]; then echo "ERROR: refusing to migrate unverified database '$db'." >&2; exit 2; fi
  for table in payments orders inventory_movements; do
    present="$(docker exec "$container" psql -X -U orderflow -d "$db" -Atqc "SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='$table'" 2>/dev/null || true)"
    if [[ "$present" != "1" ]]; then echo "ERROR: required table '$table' missing in '$db'; apply migrations 000001–000007 first." >&2; exit 2; fi
  done
  present="$(docker exec "$container" psql -X -U orderflow -d "$db" -Atqc "SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='mock_provider_transactions'" 2>/dev/null || true)"
  if [[ "$present" == "1" ]]; then
    echo "Migration 000008 appears already applied to '$db'; skipping."
    continue
  fi
  echo "Applying migration 000008 to verified database '$db'..."
  docker exec -i "$container" psql -X -U orderflow -d "$db" -v ON_ERROR_STOP=1 < migrations/000008_mock_provider_reconciliation.up.sql
done
echo 'Mock-provider reconciliation migration check complete.'
