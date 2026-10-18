#!/usr/bin/env bash
set -euo pipefail

if [[ ! -f .env ]]; then
  echo "ERROR: .env not found; configure it from .env.example first." >&2
  exit 2
fi
command -v docker >/dev/null || { echo "ERROR: Docker is required." >&2; exit 2; }
command -v python3 >/dev/null || { echo "ERROR: python3 is required." >&2; exit 2; }

# Read local configuration into this process only; do not print secret values.
set -a
# shellcheck disable=SC1091
source .env
set +a
: "${DATABASE_URL:?DATABASE_URL is required in .env}"
: "${JWT_SECRET:?JWT_SECRET is required in .env}"
: "${MOCK_PAYMENT_WEBHOOK_SECRET:?MOCK_PAYMENT_WEBHOOK_SECRET is required in .env}"

NO_CACHE_CONTAINER=orderflow-api-benchmark-nocache
CACHE_CONTAINER=orderflow-api-benchmark-cache
IMAGE=orderflow:local-benchmark
BASE_PORT="${ORDERFLOW_BENCH_NO_CACHE_PORT:-18081}"
CACHE_PORT="${ORDERFLOW_BENCH_CACHE_PORT:-18082}"
REQUESTS="${ORDERFLOW_BENCH_REQUESTS:-300}"
CONCURRENCY="${ORDERFLOW_BENCH_CONCURRENCY:-10}"

# Convert host-published Compose PostgreSQL address for a Mac/Linux Desktop container.
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

cleanup() {
  docker rm -f "$NO_CACHE_CONTAINER" "$CACHE_CONTAINER" >/dev/null 2>&1 || true
}
trap cleanup EXIT

docker compose up -d postgres redis
./scripts/apply-reliability-migration.sh
redis_ready=0
for _ in $(seq 1 30); do
  if docker exec orderflow-redis redis-cli ping 2>/dev/null | grep -q PONG; then redis_ready=1; break; fi
  sleep 1
done
if [[ "$redis_ready" -ne 1 ]]; then echo "ERROR: local Redis did not become ready." >&2; exit 2; fi
docker build -t "$IMAGE" .
cleanup

run_api() {
  local name="$1" port="$2" redis_enabled="$3"
  docker run --detach --name "$name" -p "${port}:8080" \
    -e PORT=8080 -e APP_ENV=benchmark \
    -e DATABASE_URL="$DOCKER_DATABASE_URL" \
    -e JWT_SECRET="$JWT_SECRET" \
    -e MOCK_PAYMENT_WEBHOOK_SECRET="$MOCK_PAYMENT_WEBHOOK_SECRET" \
    -e REDIS_ENABLED="$redis_enabled" -e REDIS_ADDR=host.docker.internal:6379 \
    -e REDIS_PASSWORD="${REDIS_PASSWORD:-}" -e REDIS_DB=0 -e PRODUCT_CACHE_TTL_SECONDS=60 \
    -e RECONCILIATION_INTERVAL_SECONDS=0 -e ORDER_EXPIRY_INTERVAL_SECONDS=0 \
    "$IMAGE" >/dev/null
  local ready=0
  for _ in $(seq 1 30); do
    if curl --silent --fail "http://127.0.0.1:${port}/ready" >/dev/null; then ready=1; break; fi
    sleep 1
  done
  if [[ "$ready" -ne 1 ]]; then
    docker logs "$name" >&2
    echo "ERROR: API container '$name' did not become ready." >&2
    return 1
  fi
}

run_api "$NO_CACHE_CONTAINER" "$BASE_PORT" false
run_api "$CACHE_CONTAINER" "$CACHE_PORT" true

OUTFILE="docs/evidence/performance/benchmark-results-$(date -u +%Y%m%dT%H%M%SZ).txt"
mkdir -p docs/evidence/performance
{
  echo "OrderFlow product-list API benchmark (actual local measurements)"
  echo "UTC timestamp: $(date -u +%Y-%m-%dT%H:%M:%SZ)"
  echo "Environment: $(uname -a)"
  echo "Go: $(go version 2>/dev/null || echo 'not installed on host')"
  echo "Docker: $(docker --version)"
  echo "Requests per variant: $REQUESTS; concurrent workers: $CONCURRENCY"
  echo "Endpoint: GET /products?page=1&limit=20; test dataset is the current local DB"
  echo "Cache TTL: 60 seconds; Redis DB: 0; workers disabled during measurement"
  echo "No claim that caching improves performance is made until the two measured results are compared."
  echo
} | tee "$OUTFILE"
python3 - "$BASE_PORT" "$CACHE_PORT" "$REQUESTS" "$CONCURRENCY" <<'PY' | tee -a "$OUTFILE"
import concurrent.futures
import http.client
import math
import sys
import time
from urllib.error import URLError
from urllib.request import urlopen

base_port, cache_port, requests_total, concurrency = map(int, sys.argv[1:])
if requests_total < 50 or requests_total > 100000 or concurrency < 1 or concurrency > 100:
    raise SystemExit("ORDERFLOW_BENCH_REQUESTS must be 50..100000 and ORDERFLOW_BENCH_CONCURRENCY 1..100")
path = "/products?page=1&limit=20"

def warm(port):
    for _ in range(10):
        with urlopen(f"http://127.0.0.1:{port}{path}", timeout=10) as resp:
            body = resp.read()
            if resp.status != 200:
                raise RuntimeError(f"warm-up got HTTP {resp.status}")

def worker(port, count):
    conn = http.client.HTTPConnection("127.0.0.1", port, timeout=10)
    timings, errors = [], []
    try:
        for _ in range(count):
            begin = time.perf_counter_ns()
            try:
                conn.request("GET", path, headers={"Accept": "application/json"})
                response = conn.getresponse()
                response.read()
                elapsed = (time.perf_counter_ns() - begin) / 1e6
                if response.status != 200:
                    errors.append(f"HTTP {response.status}")
                else:
                    timings.append(elapsed)
            except Exception as exc:
                errors.append(type(exc).__name__ + ": " + str(exc))
                try:
                    conn.close()
                except Exception:
                    pass
                conn = http.client.HTTPConnection("127.0.0.1", port, timeout=10)
    finally:
        conn.close()
    return timings, errors

def measure(label, port):
    warm(port)
    counts = [requests_total // concurrency + (1 if i < requests_total % concurrency else 0) for i in range(concurrency)]
    t0 = time.perf_counter()
    latencies, errors = [], []
    with concurrent.futures.ThreadPoolExecutor(max_workers=concurrency) as executor:
        futures = [executor.submit(worker, port, n) for n in counts if n]
        for future in concurrent.futures.as_completed(futures):
            values, failures = future.result()
            latencies.extend(values)
            errors.extend(failures)
    elapsed = time.perf_counter() - t0
    if not latencies:
        raise SystemExit(f"{label}: zero successful responses; first errors={errors[:5]}")
    latencies.sort()
    def percentile(frac):
        return latencies[max(0, math.ceil(frac * len(latencies)) - 1)]
    print(f"[{label}]")
    print(f"  attempted={requests_total} successful={len(latencies)} errors={len(errors)} elapsed_seconds={elapsed:.3f}")
    print(f"  p50_ms={percentile(0.50):.3f} p95_ms={percentile(0.95):.3f} mean_ms={sum(latencies)/len(latencies):.3f} throughput_rps={len(latencies)/elapsed:.2f}")
    if errors:
        print(f"  first_errors={errors[:5]}")
    print()
    return len(errors)

errors = 0
errors += measure("PostgreSQL / Redis cache disabled", base_port)
errors += measure("PostgreSQL / Redis cache enabled (warm cache)", cache_port)
print("Interpret these as local measurements for this workload and dataset, not universal production benchmarks.")
if errors:
    raise SystemExit(1)
PY

printf 'Saved actual benchmark output to %s\n' "$OUTFILE"
