#!/usr/bin/env bash
set -euo pipefail
if [[ -z "${ORDERFLOW_TEST_DATABASE_URL:-}" ]]; then
  echo 'ERROR: set ORDERFLOW_TEST_DATABASE_URL to the dedicated orderflow_test database.' >&2
  exit 2
fi
command -v go >/dev/null || { echo 'ERROR: Go is required.' >&2; exit 2; }
mkdir -p docs/evidence/performance
out="docs/evidence/performance/order-flow-benchmarks-$(date -u +%Y%m%dT%H%M%SZ).txt"
{
  echo 'OrderFlow service-level order creation/history benchmarks (actual run)'
  echo "UTC timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "Environment: $(uname -a)"
  echo "Go: $(go version)"
  echo 'Database: dedicated orderflow_test; benchmarks create isolated fixtures and clean them up.'
  echo 'These are service/repository-path ns/op benchmarks, not HTTP P95 measurements.'
  echo 'Run configuration: 200 ms per benchmark, one sample per benchmark to keep the demonstration bounded.'
  echo
  go test -run='^$' -bench='BenchmarkOrder(Creation|HistoryList)$' -benchtime=200ms -benchmem -count=1 ./internal/integration/...
} 2>&1 | tee "$out"
echo "Saved order-flow benchmark output to $out"
