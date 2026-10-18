#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${ORDERFLOW_TEST_DATABASE_URL:-}" ]]; then
  echo "ERROR: ORDERFLOW_TEST_DATABASE_URL is required." >&2
  echo "It must connect to the dedicated database named exactly orderflow_test." >&2
  echo "See INTEGRATION_TESTING_MILESTONE.md; this runner refuses to silently skip the integration suite." >&2
  exit 2
fi

# The Go test helper independently verifies the parsed database name is exactly
# orderflow_test before creating fixtures or running any mutations.
go test -count=1 ./internal/integration/...
