#!/usr/bin/env python3
"""Idempotently refresh stale status text and remove portal-specific checklist from README."""
from pathlib import Path


def replace_once_or_report(text: str, old: str, new: str, label: str) -> str:
    if old in text:
        return text.replace(old, new, 1)
    if new in text:
        return text
    print(f"NOTE: expected source text not found for {label}; left unchanged")
    return text

readme = Path("README.md")
if readme.exists():
    text = readme.read_text(encoding="utf-8")
    text = text.replace(
        "**Latest repository HEAD reviewed:** `3df0139` — documentation and evidence update.  ",
        "**Source of truth:** the public `main` branch. Always verify the GitHub Actions run for the exact commit being submitted.  ",
    )
    text = text.replace(
        "At the latest repository HEAD `3df0139`, the GitHub Actions run is green.",
        "The last reviewed baseline had a successful GitHub Actions run; verify the run for the exact commit being submitted.",
    )
    text = text.replace(
        "- Latest reviewed HEAD: `3df0139`; last behavior-changing implementation: `1515cb0`\n- Successful run: <https://github.com/manmohansharma21/orderflow/actions/runs/39165514715>",
        "- The latest implementation and docs are on `main`. Check <https://github.com/manmohansharma21/orderflow/actions> for the run associated with the exact commit you intend to submit.",
    )
    # Remove the repeated second Order/payment state diagram block, preserving the first.
    duplicate = "### Order/payment states\n\n![OrderFlow order and payment state transitions](docs/diagrams/ORDER_PAYMENT_STATE.png)\n\n"
    if text.count(duplicate) > 1:
        first = text.find(duplicate)
        before = text[:first + len(duplicate)]
        after = text[first + len(duplicate):].replace(duplicate, "", 1)
        text = before + after
    # Replace portal-specific instructions with a concise release-verification section.
    start_marker = "## Scaler submission checklist\n"
    end_marker = "## Scope boundaries and limitations\n"
    start = text.find(start_marker)
    end = text.find(end_marker, start + len(start_marker)) if start >= 0 else -1
    if start >= 0 and end > start:
        replacement = (
            "## Release verification\n\n"
            "Before producing a release archive, run `go test ./...`, `go test -race ./internal/integration/...`, `go vet ./...`, and `git diff --check`; verify the GitHub Actions run for the exact commit. Keep generated evidence and editable diagrams in the repository, and never commit `.env`, credentials or database backups. Personal portal-upload steps are maintained separately in `docs/FINAL_SUBMISSION_STATUS.md`.\n\n"
        )
        text = text[:start] + replacement + text[end:]
    # Update status paragraph with current benchmark screenshot and avoid stale SHAs.
    text = text.replace(
        "Latest reviewed HEAD is `3df0139`, whose GitHub Actions run succeeded: https://github.com/manmohansharma21/orderflow/actions/runs/39165514715.",
        "The latest reviewed commit at the time of this note passed CI; verify the current `main` run before submission: https://github.com/manmohansharma21/orderflow/actions.",
    )
    if "benchmark-results-20261018T230232Z.png" not in text:
        text += ("\n\n## Additional benchmark evidence\n\n"
                 "The run captured at `docs/evidence/performance/benchmark-results-20261018T230232Z.png` showed 300/300 successful requests in both variants. This run favored warm Redis (P50 2.429 ms vs 2.829 ms; P95 4.140 ms vs 5.203 ms), while an earlier run showed the opposite. Treat results as variable local measurements, not a guaranteed cache speedup. See `docs/evidence/performance/README.md`.\n")
    readme.write_text(text, encoding="utf-8")
    print("README cleanup applied")

report = Path("docs/PROJECT_REPORT.md")
if report.exists():
    text = report.read_text(encoding="utf-8")
    text = text.replace("**Last verified code baseline:** `3df0139`", "**Source of truth:** public `main`; verify CI for the exact submitted HEAD")
    text = text.replace("https://github.com/manmohansharma21/orderflow/actions/runs/39163364836 (success)", "https://github.com/manmohansharma21/orderflow/actions (verify exact submitted HEAD)")
    text = text.replace("Code baseline `3df0139` has a successful [GitHub Actions run](https://github.com/manmohansharma21/orderflow/actions/runs/39163364836)", "The reviewed implementation baseline passed [GitHub Actions](https://github.com/manmohansharma21/orderflow/actions); verify the exact final HEAD before submission")
    start = text.find("## Performance results\n")
    end = text.find("## Scope limitations and next steps\n", start + 1) if start >= 0 else -1
    performance = '''## Performance results\n\nTwo actual local HTTP benchmark runs were captured for `GET /products?page=1&limit=20`, each with 300 requests per variant and 10 concurrent workers on an Apple M2 (macOS arm64), Go 1.27.2 and Docker Desktop 29.8.2. Both runs reported 300 successful responses and zero errors in each variant.\n\n| Run | Variant | P50 | P95 | Mean | Throughput |\n|---|---|---:|---:|---:|---:|\n| 2026-10-18 19:13:09 UTC | Redis disabled | 2.480 ms | 4.270 ms | 3.315 ms | 2,821.81 req/s |\n| 2026-10-18 19:13:09 UTC | Redis warm | 4.447 ms | 13.966 ms | 5.702 ms | 1,697.92 req/s |\n| 2026-10-18 23:02:32 UTC | Redis disabled | 2.829 ms | 5.203 ms | 3.748 ms | 2,190.99 req/s |\n| 2026-10-18 23:02:32 UTC | Redis warm | 2.429 ms | 4.140 ms | 2.937 ms | 3,297.18 req/s |\n\n**Interpretation:** the first run favored PostgreSQL without cache; the second favored warm Redis. This inconsistency means the available data does not support a stable general claim that caching improves performance. Results depend on dataset, cache warmth and local runtime noise. Redis remains optional and disabled by default in the project configuration. A stronger conclusion would require repeat runs, a representative dataset, warm-up/control, and reporting distribution/variance.\n\n![Screenshot of the second completed API cache benchmark](evidence/performance/benchmark-results-20261018T230232Z.png)\n\nThe separate bounded service/repository benchmark completed at the previously reviewed baseline: order creation 1,848,705 ns/op, 5,892 B/op, 112 allocs/op; order history 551,748 ns/op, 19,837 B/op, 312 allocs/op. These one-sample Apple M2 service-path measurements are not HTTP latency percentiles. The corresponding actual output is stored in `docs/evidence/performance/order-flow-benchmarks-20261018T202650Z.txt`.\n\n'''
    if start >= 0 and end > start:
        text = text[:start] + performance + text[end:]
    else:
        print("NOTE: Performance section markers not found; left section unchanged")
    if "`3df0139`" in text:
        text = text.replace("`3df0139`", "the previously reviewed baseline")
    report.write_text(text, encoding="utf-8")
    print("Project report source refreshed")
