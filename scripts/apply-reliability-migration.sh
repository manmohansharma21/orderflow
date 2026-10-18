#!/usr/bin/env bash
set -euo pipefail

container="${ORDERFLOW_POSTGRES_CONTAINER:-orderflow-postgres}"
command -v docker >/dev/null || { echo "ERROR: Docker is required." >&2; exit 2; }
docker inspect "$container" >/dev/null 2>&1 || { echo "ERROR: PostgreSQL container '$container' does not exist or is not running." >&2; exit 2; }
ready=0
for _ in $(seq 1 30); do
  if docker exec "$container" pg_isready -U orderflow -d orderflow >/dev/null 2>&1; then ready=1; break; fi
  sleep 1
done
if [[ "$ready" -ne 1 ]]; then echo "ERROR: PostgreSQL in '$container' did not become ready." >&2; exit 2; fi
for db in orderflow orderflow_test; do
  actual="$(docker exec "$container" psql -X -U orderflow -d "$db" -Atqc 'SELECT current_database()' 2>/dev/null || true)"
  if [[ "$actual" != "$db" ]]; then echo "ERROR: expected database '$db'; unable to verify current database." >&2; exit 2; fi
  for table in orders payments inventory_movements; do
    present="$(docker exec "$container" psql -X -U orderflow -d "$db" -Atqc "SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='$table'" 2>/dev/null || true)"
    if [[ "$present" != "1" ]]; then
      echo "ERROR: required table '$table' is missing in '$db'. Apply and verify migrations 000001–000006 before migration 000007." >&2
      exit 2
    fi
  done
  if docker exec "$container" psql -X -U orderflow -d "$db" -Atqc "SELECT 1 FROM information_schema.columns WHERE table_name='orders' AND column_name='idempotency_request_hash'" | grep -q 1; then
    echo "Migration 000007 appears already applied to '$db'; skipping."
    continue
  fi
  echo "Applying migration 000007 to verified database '$db'..."
  docker exec -i "$container" psql -X -U orderflow -d "$db" -v ON_ERROR_STOP=1 < migrations/000007_order_payment_reliability.up.sql
done

echo "Reliability migration application complete. Verify both databases before testing."
