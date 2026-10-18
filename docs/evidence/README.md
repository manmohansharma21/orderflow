# OrderFlow evidence and screenshot guide

Use this folder to hold evidence from the final end-to-end session. The same real PNGs saved here should be pasted into the corresponding Word report figure placeholders before PDF export. Do not create synthetic screenshots or include bearer tokens, passwords, webhook signatures, database URLs, secrets or unrelated personal/account details.

| Capture after | Exact filename | Report placement |
|---|---|---|
| Final pushed HEAD has a green Actions run | `ci/01-github-actions-green.png` | Chapter 14; show final SHA |
| Local container smoke passes | `docker/01-health-readiness.png` | Chapter 14; image build plus `/health` and `/ready` |
| Customer login/profile succeeds | `postman/01-auth-profile.png` | Chapter 7; show profile result, not token |
| Product create/update and duplicate SKU test run | `postman/02-product-admin.png` | Chapter 7 |
| Inventory availability, receipt/adjustment, movements run | `postman/03-inventory-audit.png` | Chapters 8–9 |
| Multi-product order, idempotent replay/conflict and cancellation run | `postman/04-order-lifecycle.png` | Chapters 8–9 |
| Signed payment success/failure, invalid signature and replay run | `postman/05-payment-callbacks.png` | Chapter 10 |
| Missed callback recovered and admin reports run | `postman/06-reconciliation-and-reports.png` | Chapter 10 / reporting section |
| API cache benchmark run | `performance/01-api-cache-benchmark.png` | Chapter 13; include workload and negative result |
| Bounded order-history benchmark completes | `performance/02-order-benchmarks.png` | Chapter 13; only after successful completion |
| `/metrics` and `X-Request-ID` verified | `observability/01-request-metrics.png` | Chapters 11–12 |
| Actual approved AWS deployment, only if required | `aws/01-live-health.png` | Chapter 15; otherwise leave absent |

## macOS screenshot steps

Create folders with `mkdir -p docs/evidence/{ci,docker,postman,performance,observability,aws}`. Arrange only the relevant result on screen, press **Command + Shift + 4**, drag to crop, and move the saved PNG to the filename above. For Postman, keep method/path, status and response body visible; do not show Authorization token values. Redact permanently in Preview before committing if any sensitive value is visible. Capture the exact green Actions run for final submitted SHA. For AWS, do not capture/claim deployment unless it actually exists.

The product-list benchmark file `performance/benchmark-results-20261018T191309Z.txt` is already committed with actual measurements. The warm Redis path was slower for this small workload, so the screenshot/report must retain that result. The most recent order-history benchmark attempt was interrupted; only save a success screenshot after the bounded standalone benchmark exits normally.
