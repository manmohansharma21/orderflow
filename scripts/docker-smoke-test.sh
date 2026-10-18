#!/usr/bin/env bash
set -euo pipefail

if [[ ! -f .env ]]; then echo "ERROR: .env not found; copy .env.example and fill local values first." >&2; exit 2; fi
command -v docker >/dev/null || { echo "ERROR: Docker is required." >&2; exit 2; }
command -v python3 >/dev/null || { echo "ERROR: python3 is required." >&2; exit 2; }
command -v curl >/dev/null || { echo "ERROR: curl is required." >&2; exit 2; }

# Load local config into this process only; never print environment values.
set -a
# shellcheck disable=SC1091
source .env
set +a
: "${DATABASE_URL:?DATABASE_URL is required in .env}"
: "${JWT_SECRET:?JWT_SECRET is required in .env}"
: "${MOCK_PAYMENT_WEBHOOK_SECRET:?MOCK_PAYMENT_WEBHOOK_SECRET is required in .env}"

CONTAINER=orderflow-api-smoke-local
IMAGE=orderflow:local-smoke
PORT_HOST="${ORDERFLOW_SMOKE_PORT:-18080}"
DOCKER_DATABASE_URL="$(python3 - "$DATABASE_URL" <<'PY'
import sys
from urllib.parse import urlsplit, urlunsplit
p = urlsplit(sys.argv[1])
host = p.hostname or ""
if host in ("localhost", "127.0.0.1", "::1"):
    host = "host.docker.internal"
netloc = host
if p.port:
    netloc += f":{p.port}"
if p.username is not None:
    auth = p.username
    if p.password is not None:
        auth += ":" + p.password
    netloc = auth + "@" + netloc
print(urlunsplit((p.scheme, netloc, p.path, p.query, p.fragment)))
PY
)"

cleanup() { docker rm -f "$CONTAINER" >/dev/null 2>&1 || true; }
trap cleanup EXIT

docker compose up -d postgres redis
./scripts/apply-reliability-migration.sh
# Wait for both local dependencies without displaying connection credentials.
for _ in $(seq 1 30); do
  if docker exec orderflow-postgres pg_isready -U orderflow -d orderflow >/dev/null 2>&1; then break; fi
  sleep 1
done
docker exec orderflow-postgres pg_isready -U orderflow -d orderflow >/dev/null
for _ in $(seq 1 30); do
  if docker exec orderflow-redis redis-cli ping 2>/dev/null | grep -q PONG; then break; fi
  sleep 1
done
docker exec orderflow-redis redis-cli ping | grep -q PONG
docker build -t "$IMAGE" .
docker rm -f "$CONTAINER" >/dev/null 2>&1 || true
docker run -d --name "$CONTAINER" -p "${PORT_HOST}:8080"   -e PORT=8080 -e APP_ENV=smoke -e DATABASE_URL="$DOCKER_DATABASE_URL"   -e JWT_SECRET="$JWT_SECRET" -e MOCK_PAYMENT_WEBHOOK_SECRET="$MOCK_PAYMENT_WEBHOOK_SECRET"   -e REDIS_ENABLED=true -e REDIS_ADDR=host.docker.internal:6379   -e REDIS_DB=0 -e PRODUCT_CACHE_TTL_SECONDS=60   -e RECONCILIATION_INTERVAL_SECONDS=0 "$IMAGE" >/dev/null

ready=0
for _ in $(seq 1 30); do
  if curl --silent --fail "http://127.0.0.1:${PORT_HOST}/health" >/dev/null; then ready=1; break; fi
  sleep 1
done
if [[ "$ready" -ne 1 ]]; then docker logs "$CONTAINER"; echo "ERROR: health endpoint did not become ready." >&2; exit 1; fi
curl --silent --show-error --fail "http://127.0.0.1:${PORT_HOST}/health"
echo
curl --silent --show-error --fail "http://127.0.0.1:${PORT_HOST}/ready"
echo
echo "Docker smoke test passed. Secrets were not printed."
