#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${ORDERFLOW_TEST_DATABASE_URL:-}" ]]; then
  echo "ERROR: ORDERFLOW_TEST_DATABASE_URL must target the existing orderflow_test database." >&2
  echo "See INTEGRATION_TESTING_MILESTONE.md; do not point it at orderflow or production." >&2
  exit 2
fi

command -v go >/dev/null || { echo "ERROR: Go 1.26+ is required by go.mod." >&2; exit 2; }
command -v docker >/dev/null || { echo "ERROR: Docker is required for the Redis-backed checks." >&2; exit 2; }

docker compose up -d postgres redis
./scripts/apply-reliability-migration.sh
./scripts/apply-mock-provider-migration.sh
export ORDERFLOW_REDIS_TEST_ADDR="${ORDERFLOW_REDIS_TEST_ADDR:-localhost:6379}"

go mod tidy
find cmd internal -name '*.go' -print0 | xargs -0 gofmt -w

go test -count=1 ./...
go test -race -count=1 ./internal/integration/...
go vet ./...
git diff --check

echo "Residual Go/Redis/reconciliation verification passed. Review git status; go.mod/go.sum may have changed due go mod tidy."
