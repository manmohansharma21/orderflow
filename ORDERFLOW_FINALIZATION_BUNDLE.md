# OrderFlow next milestone and report working draft

This ZIP is an **overlay** for the existing `manmohansharma21/orderflow` repository, not a complete duplicate of its source. It includes Redis product-read caching, a scheduled read-only reconciliation reporter, Docker/CI smoke checks, updated developer docs, an exact screenshot runbook, and a formatted 52-page Applied Software Project Report working draft in `submission/` as DOCX and PDF.

**Do not submit the report as-is.** It contains highlighted identity/date placeholders and intentional evidence gaps. Paste actual screenshots, update results to match the final commit, and export the final report only after all rubric-required behavior is verified.

## 1. Extract at the repository root

```bash
unzip -o ~/Downloads/ORDERFLOW_FINALIZATION_PACK_DRAFT.zip -d .
```

See `docs/RESIDUAL_MILESTONE.md` and `docs/evidence/README.md` for the work sequence and precise screenshot filenames/chapters. Existing request DTOs, domain models, services and repository files are already in the repo; this bundle does not create duplicate layers.

## 2. Verify the residual code

Use Go 1.26+ and Docker Desktop. Keep `ORDERFLOW_TEST_DATABASE_URL` set to the isolated database named exactly `orderflow_test`, per `INTEGRATION_TESTING_MILESTONE.md`. Do not rerun the migration setup against an already initialized local test DB.

```bash
chmod +x scripts/verify-residual-milestone.sh scripts/docker-smoke-test.sh
./scripts/verify-residual-milestone.sh
./scripts/docker-smoke-test.sh
```

The first script runs `go mod tidy`, formatting, full tests, race-enabled PostgreSQL integration tests, `go vet`, and whitespace validation, using a dedicated local Redis instance (test code uses Redis DBs 14 and 15). The second builds and starts the API container and probes `/health` and `/ready`; it uses `.env` without printing the secret values and does not drop databases. Review any `go.mod`/`go.sum` changes caused by `go mod tidy`.

## 3. Review then commit

After the checks pass:

```bash
git status --short
git diff --check
git diff --stat
```

Review staged paths before committing. Include intended source/config/doc changes plus `go.mod` and `go.sum` if updated by `go mod tidy`. Do **not** stage `.env`, local DB dumps, credentials, tokens, generated secrets, or the `submission/` report draft until it has been personalized and evidence-filled.

```bash
git add .env.example .dockerignore Dockerfile docker-compose.yml .github/workflows/ci.yml \
  README.md ORDERFLOW_FINALIZATION_BUNDLE.md \
  cmd/api/main.go internal/cache internal/config/config.go internal/http/router.go \
  internal/integration/product_cache_test.go \
  internal/repository/product_repository_adapter.go \
  internal/service/product_cache.go internal/service/product_service.go \
  internal/worker scripts/docker-smoke-test.sh scripts/verify-residual-milestone.sh \
  docs
# Add go.mod and go.sum if modified by go mod tidy.
git diff --cached --check
git diff --cached --stat
git status --short
```

If staged files are correct and **all local verification passed**, commit and push:

```bash
git commit -m "feat: add Redis product cache and reconciliation worker"
git push origin main
git log -1 --oneline
git status --short
```

Then open GitHub Actions and confirm a green run for the exact latest HEAD SHA. The previously green run at `60c255b` predates this overlay, so it does not validate it.

## 4. Remaining rubric work (not claimed complete here)

Before calling the capstone final, compare `docs/REMOTE_REVIEW_AND_REMAINING_WORK.md` against the original specification. Gaps include multi-product orders, order-create idempotency, callback amount/currency and failure-reason validation, successful stock-consumption semantics, order expiry, required admin reporting and performance benchmarks. The scheduled worker added here is **read-only**; it reports mismatches but does not automatically charge/refund or transition order/payment states.

AWS is only a deployment plan until provisioned and verified with explicit account/budget approval. Capture API/CI/Docker evidence using `docs/evidence/README.md`; do not fabricate screenshots or benchmark metrics. Learn `.github/workflows/ci.yml` as a separate tracker item.
