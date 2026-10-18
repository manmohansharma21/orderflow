#!/usr/bin/env bash
set -euo pipefail
if [[ -z "${ORDERFLOW_TEST_DATABASE_URL:-}" ]]; then
  echo "ERROR: set ORDERFLOW_TEST_DATABASE_URL to the dedicated orderflow_test database; see INTEGRATION_TESTING_MILESTONE.md" >&2
  exit 2
fi
command -v go >/dev/null || { echo "ERROR: Go is required" >&2; exit 2; }
command -v docker >/dev/null || { echo "ERROR: Docker is required to start test Redis" >&2; exit 2; }
docker compose up -d redis
export ORDERFLOW_REDIS_TEST_ADDR="${ORDERFLOW_REDIS_TEST_ADDR:-localhost:6379}"
printf 'Benchmark environment: database=orderflow_test redis=%s iterations=50\n' "$ORDERFLOW_REDIS_TEST_ADDR"
go test -run '^$' -bench '^BenchmarkProductListDatabaseVsRedis$' -benchmem -benchtime=50x ./internal/integration/...
