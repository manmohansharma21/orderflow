# Submission evidence / screenshots checklist

Save evidence to `docs/evidence/` (create the directory locally) and reference filenames in the report. Avoid capturing JWTs, passwords, database URLs, webhook secrets, or other credentials in screenshots.

## API behavior (Postman)

- [ ] Health response: `GET /health` → 200.
- [ ] Ready response with database connected: `GET /ready` → 200.
- [ ] Register/login response; redact the bearer token before saving a screenshot.
- [ ] Public products search with `search`, `page`, and `limit`.
- [ ] Admin creates a product and initial inventory; confirm duplicate SKU receives 409.
- [ ] Inventory available quantity, admin receipt, adjustment and movement history.
- [ ] Customer creates an order; detail/list show `PENDING` and total in paise.
- [ ] Customer cancellation releases reserved stock and leaves an `ORDER_RELEASED` movement.
- [ ] Payment initiation with `Idempotency-Key`; repeating same key returns the same payment ID.
- [ ] Valid signed mock callback success → order `CONFIRMED`. IMPORTANT: current code retains reservation on SUCCESS, while the original capstone spec says successful reserved stock becomes consumed; fix and test this mismatch before claiming stock consumption in the report.
- [ ] Valid signed mock callback failure → order `FAILED` and stock released once.
- [ ] Invalid mock signature → 401.
- [ ] Admin payment reconciliation response.

## Automated test evidence

- [ ] Terminal or CI log for `go test ./...`.
- [ ] Terminal or CI log for `go test -race ./internal/integration/...`.
- [ ] Terminal or CI log for `go vet ./...`.
- [ ] GitHub Actions workflow run after the CI file is pushed (the presence of a workflow file is not evidence that it ran).
- [ ] GitHub commit page showing the final commit SHA.

## Architecture and documentation evidence

- [ ] README HLD and LLD diagrams render on GitHub.
- [ ] OpenAPI file loads in Swagger UI without parse errors.
- [ ] Postman collection imported and at least one full test path has passed.
- [ ] Project report updated with actual results, screenshot references, and rubric-specific additions.
- [ ] Final submission ZIP/PDF reviewed; `.env`, access tokens, and credentials excluded.

## When to capture and where to paste

See [`docs/evidence/README.md`](evidence/README.md) for the exact capture timing, filenames, and report chapters. Create the named subfolders as evidence is captured.

## CI learning checklist

Add time to learn `.github/workflows/ci.yml`: triggers, job permissions, service containers, migration loop, Go version setup, tests/race, vet/whitespace, Docker build and runtime probes. Do not call the residual workflow verified until the exact pushed commit has a green run.

## New residual milestone evidence

- [ ] GitHub Actions run after Redis/worker changes: Redis TTL and version tests pass.
- [ ] GitHub Actions Docker image build passes.
- [ ] GitHub Actions container smoke check shows `/health` and `/ready` succeed.
- [ ] Product list/detail cache invalidation after an admin update passes in the PostgreSQL+Redis test.
- [ ] Optional local `scripts/docker-smoke-test.sh` output saved without environment values.
- [ ] Reconciliation worker log example captured with synthetic/test data only; redact IDs if using real data.
