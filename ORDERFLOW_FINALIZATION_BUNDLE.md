# OrderFlow completion bundle — overlay instructions

This ZIP overlays the existing `manmohansharma21/orderflow` repository. It does not create duplicate DTO/model/service/repository layers. Existing code is extended to address the remaining core requirements in the original OrderFlow specification.

## Extraction

Run this from the repository root on your Mac:

```bash
unzip -o ~/Downloads/ORDERFLOW_CAPSTONE_RELIABILITY_COMPLETION.zip -d .
```

## Verify

1. Keep `ORDERFLOW_TEST_DATABASE_URL` pointed at the dedicated `orderflow_test` database. Do not paste or print the URL.
2. Run `chmod +x scripts/*.sh`.
3. Run `./scripts/verify-residual-milestone.sh`; it starts PostgreSQL/Redis, applies migration 000007 only if needed, runs formatting, `go mod tidy`, unit/integration tests, race checks, vet and whitespace validation.
4. Run `./scripts/docker-smoke-test.sh` to build and run the API container and validate `/health` and `/ready`.
5. Run `./scripts/benchmark-product-cache.sh` and retain actual before/after output. Do not invent response-time figures.

## Commit and push

Unstage any earlier `submission/*_DRAFT.*` files before staging. Review `git status --short`, `git diff --cached --check` and `git diff --cached --stat`. Keep `.env`, all secrets and database backups out of Git. Only commit if all local checks pass. Push `main` and confirm that GitHub Actions succeeds for the exact new HEAD.

## Completion boundaries

The code overlay includes multi-item orders/order idempotency, payment amount/currency validation and failure reasons, stock consumption on successful payment, safe unpaid-order expiry, admin reports, Redis cache/invalidation and a scheduled read-only reconciliation reporter. The report markdown and diagrams are updated, but the final Word/PDF report is intentionally deferred until actual benchmark results and evidence are available. AWS deployment and real-provider integration are not claimed; they require external configuration/approval. The CI-learning walkthrough is in `docs/CI_CONFIGURATION_WALKTHROUGH.md`. The screenshot runbook remains for the final evidence session only.
