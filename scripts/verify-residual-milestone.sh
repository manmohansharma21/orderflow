#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${ORDERFLOW_TEST_DATABASE_URL:-}" ]]; then
  echo "ERROR: ORDERFLOW_TEST_DATABASE_URL must target the existing orderflow_test database." >&2
  echo "See INTEGRATION_TESTING_MILESTONE.md; do not point it at orderflow or production." >&2
  exit 2
fi

command -v go >/dev/null || { echo "ERROR: Go 1.26+ is required by go.mod." >&2; exit 2; }
command -v docker >/dev/null || { echo "ERROR: Docker is required for the Redis-backed checks." >&2; exit 2; }

docker compose up -d redis
export ORDERFLOW_REDIS_TEST_ADDR="${ORDERFLOW_REDIS_TEST_ADDR:-localhost:6379}"

go mod tidy
gofmt -w cmd/api/main.go internal/cache/*.go internal/config/config.go \
  internal/http/router.go internal/integration/product_cache_test.go \
  internal/repository/product_repository_adapter.go internal/service/product_cache.go \
  internal/service/product_service.go internal/worker/*.go

go test -count=1 ./...
go test -race -count=1 ./internal/integration/...
go vet ./...
git diff --check

echo "Residual Go/Redis/reconciliation verification passed. Review git status; go.mod/go.sum may have changed due go mod tidy."
