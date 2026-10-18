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
- [ ] Valid signed mock callback success → order `CONFIRMED` and reservation retained.
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
