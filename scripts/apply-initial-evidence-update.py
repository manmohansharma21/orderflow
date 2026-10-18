#!/usr/bin/env python3
"""Apply OrderFlow evidence/status corrections to existing Markdown docs. Idempotent."""
from pathlib import Path

root = Path(__file__).resolve().parents[1]
replacements = {
    "The order-create service benchmark returned partial observations between about 1.90 and 3.47 ms/op across interrupted runs; there is no valid order-history result yet.":
    "The bounded service-level benchmark completed locally: order creation 1,848,705 ns/op (5,892 B/op, 112 allocs/op) and order history 551,748 ns/op (19,837 B/op, 312 allocs/op), using 200 ms per benchmark and one sample. These are service/repository-path results, not HTTP P95 measurements.",
    "The `BenchmarkOrderHistoryList` service-level benchmark did not complete in the captured terminal session, so no order-history performance result is claimed.":
    "The bounded `BenchmarkOrderHistoryList` service-level benchmark completed at 551,748 ns/op, 19,837 B/op and 312 allocs/op; these are local service/repository-path measurements, not HTTP P95.",
    "The current payment integration is deliberately a mock provider; a real provider lookup is needed before reconciliation can automatically repair a missed callback.":
    "The payment integration is deliberately a mock provider. The separate mock-provider ledger and reconciliation path recover known terminal outcomes through the validated payment state-transition path; real gateway status lookup remains out of scope.",
    "A configurable background worker runs the existing reconciliation report periodically and logs each issue. It is read-only; it does not mutate payments/orders or replace a real provider reconciliation process.":
    "A configurable background worker expires safe unpaid orders and reconciles known terminal results from the separate simulated-provider ledger through the shared callback path, then reports remaining inconsistent states. Real-gateway reconciliation remains out of scope.",
    "Scheduled payment reconciliation remains read-only for the mock provider until a provider status-lookup contract exists.":
    "Scheduled mock-provider reconciliation recovers known terminal outcomes from the separate simulator ledger; real provider status lookup remains out of scope.",
    "the current order-history run was interrupted, so use a bounded `go test -run=^$ -bench=BenchmarkOrderHistoryList -benchtime=100ms -benchmem -count=1 ./internal/integration/...` if a repeat is needed.":
    "the bounded service benchmark completed with order creation at 1,848,705 ns/op and order history at 551,748 ns/op; repeat only for a more stable sample if useful.",
    "a completed order-history benchmark if required":
    "a final Postman/CI/Docker evidence session and report completion",
}
for rel in ("docs/PROJECT_REPORT.md", "README.md", "docs/REDIS_AND_RECONCILIATION.md"):
    path = root / rel
    if not path.exists():
        continue
    text = path.read_text(encoding="utf-8")
    original = text
    for old, new in replacements.items():
        text = text.replace(old, new)
    if rel == "docs/PROJECT_REPORT.md":
        marker = "## Evidence refresh — 19 October 2026"
        if marker not in text:
            text += """

## Evidence refresh — 19 October 2026

This section supersedes earlier status statements above where they refer to stale verification or benchmark state.

- Verified implementation commit at this review: `3c708ab`; GitHub Actions passed at https://github.com/manmohansharma21/orderflow/actions/runs/39163364836.
- Local `/health`, `/ready`, `/metrics` browser captures and an initial development server log are saved under `docs/evidence/`. The server log is supporting evidence only.
- Product-list API benchmark: no cache P50 2.480 ms / P95 4.270 ms / 2,821.81 req/s; warm Redis P50 4.447 ms / P95 13.966 ms / 1,697.92 req/s. Redis was slower for this workload; no speedup is claimed.
- Bounded service-level benchmark: order creation 1,848,705 ns/op (5,892 B/op, 112 allocs/op); order history 551,748 ns/op (19,837 B/op, 312 allocs/op). These are single-sample service/repository-path measurements, not HTTP latency percentiles.
- Still needed for submission: final green-CI screenshot for the submitted SHA, Docker smoke/build screenshot, Postman business-flow screenshots, benchmark screenshot(s), metadata/reference checks, and Word-to-PDF export. AWS evidence is conditional on the rubric; the diagram is target architecture, not proof of deployment.
"""
    if text != original:
        path.write_text(text, encoding="utf-8")
        print(f"Updated {rel}")
    else:
        print(f"No matching stale phrases in {rel}; left unchanged")
