# OrderFlow Final Reliability and Submission Instructions

The current reliability baseline is `d891056` and its GitHub Actions run passed. The code in that commit covers multi-product order creation, request idempotency, callback amount/currency validation, payment failure reasons, successful-payment stock consumption, unpaid-order expiry and admin reports.

This final-review overlay closes the remaining mock-provider missed-callback gap using a separate simulated-provider ledger and the same callback processing path; it adds request IDs, latency histogram metrics, an optional Postman recovery walkthrough, updated OpenAPI, diagrams and report/evidence docs. It is **not verified until your local checks and the new remote CI run pass**.

## Apply / verify

```bash
unzip -o ~/Downloads/ORDERFLOW_FINAL_REVIEW_PATCH.zip -d .
chmod +x scripts/*.sh
./scripts/apply-final-review-patch.sh
./scripts/verify-residual-milestone.sh
./scripts/docker-smoke-test.sh
./scripts/benchmark-api-cache.sh
```

The migration script checks the database names `orderflow` and `orderflow_test` and skips migration `000008` if its table already exists. Do not run the test migrations against production data. The benchmark writes actual results under `docs/evidence/performance/`.

Before committing, inspect `git status --short`, `git diff --check` and the staged file list. Never stage `.env`, credentials, backups, or report `DRAFT` binaries. Run the tests, commit, push, and confirm CI for that exact SHA is green.

## Honest performance interpretation

The captured API benchmark reports direct PostgreSQL at P95 4.270 ms and warm Redis at P95 13.966 ms (300 successful requests per case, local Apple M2). It is not a cache speedup. Keep Redis disabled by default until representative larger-data testing shows benefit.

## Scaler upload

Provide the public repository link, download the **full source archive from the final pushed GitHub `main`**, and upload the final PDF report in its separate field. Do not upload this overlay ZIP as the project archive. Screenshots must be captured from actual requests/CI runs and added only after final testing.
