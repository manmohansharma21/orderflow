# OrderFlow evidence: exactly when to capture screenshots and where to paste them

Create these folders under `docs/evidence/`. Capture real results only; do not fabricate screenshots or include tokens, credentials, `.env` values, webhook signatures, or AWS account-sensitive information. PNG is preferred. Use a demo account and redact any bearer token before saving.

| When to capture | Exact suggested filename | Where to paste in the local final Word report copy (the PDF is submitted separately) |
|---|---|---|
| After the residual bundle passes locally and GitHub Actions is green on the same commit SHA | `docs/evidence/ci/01-github-actions-green.png` | Chapter 14, Docker/CI verification; cite the commit SHA below the figure |
| Immediately after the local Docker smoke script succeeds | `docs/evidence/docker/01-health-readiness.png` | Chapter 14; show successful image/container and `/health`, `/ready` responses only |
| After registering a demo customer and logging in via Postman | `docs/evidence/postman/01-auth-profile.png` | Chapter 7, REST API; redact the access token |
| After product create/update and duplicate-SKU test | `docs/evidence/postman/02-product-admin.png` | Chapter 7, REST API / catalog |
| After stock receipt or adjustment and movement listing | `docs/evidence/postman/03-inventory-audit.png` | Chapter 8, database/inventory workflow |
| After order creation, list/detail and cancellation | `docs/evidence/postman/04-order-lifecycle.png` | Chapters 8–9, transaction and lifecycle |
| After signed payment success, failure, invalid signature and callback replay | `docs/evidence/postman/05-payment-callbacks.png` | Chapter 10; make sure the displayed result matches the final code and specification |
| After running the configured reconciliation process and endpoint | `docs/evidence/postman/06-reconciliation.png` | Chapter 10, payment recovery |
| Only after running a repeatable load test both before and after cache/index changes | `docs/evidence/performance/01-baseline.png` and `02-redis-result.png` (the Redis result may be slower; report the actual result) | Chapter 13; report environment, data size, request count, concurrency, duration, p50/p95, throughput and errors |
| After confirming Mermaid diagrams render on the pushed README | `docs/evidence/architecture/01-hld-lld-rendered.png` | Chapters 4–6; diagram sources remain in `docs/diagrams/` |
| Only after an actual AWS deployment has been approved, launched and verified | `docs/evidence/aws/01-live-health.png` | Chapter 15; redact account IDs, public/private secrets and security-sensitive details |

## Workflow
1. Run the command/API request; confirm the response code and payload are real.
2. Capture only the relevant terminal or Postman pane, not a whole desktop containing unrelated personal data.
3. Save the PNG with the filename above.
4. In Word, replace the matching highlighted `[PASTE ACTUAL SCREENSHOT HERE]` / evidence note, insert the image, and add a numbered caption (for example, `Figure 14.01: GitHub Actions passed on commit <SHA>`).
5. Update the report's List of Figures and pagination after all inserts. Never state a benchmark, cloud deployment, Redis verification, or workflow outcome that was not observed.

## Final-stage accuracy gate
This completion overlay changes successful payment to consume on-hand and reserved stock atomically and adds a regression test for replay. Do not capture the payment/stock-consumption evidence until this overlay's tests pass and the latest pushed commit has a green GitHub Actions run. If the observed result differs from the final code/report text, stop and fix the discrepancy rather than documenting an unsupported claim.

## CI learning checklist
Read `.github/workflows/ci.yml` top to bottom: trigger, permissions, service containers and health checks, environment variables, checkout, Go setup, `go mod tidy` consistency check, migration loop, unit/integration tests, race tests, `go vet`, whitespace check, Docker build, container smoke test and health/readiness curl probes. After the Redis/worker bundle is committed, capture only the green run for that exact HEAD SHA.

## Final reliability additions

- [ ] Multi-product order creates one order with all item snapshots and reserves every line atomically.
- [ ] Same order idempotency key replays the original result; changed payload returns 409.
- [ ] Mock provider outcome is set without delivering a callback, then `POST /admin/payments/reconciliation/run?older_than_minutes=0` recovers it through the callback transition path.
- [ ] `X-Request-ID` appears in HTTP responses and `GET /metrics` returns request counters/histogram.
- [ ] Benchmark outputs for product API, order creation and order-history listing are saved with hardware/workload details; do not omit negative cache results.
