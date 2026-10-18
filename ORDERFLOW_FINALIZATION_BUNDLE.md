# OrderFlow finalization bundle — install and validate

This ZIP is an **overlay**, not a complete copy of the repository. Extract it at the repository root so every file lands at its intended path. It adds documentation/tooling and updates the example environment template; it does not change existing service/repository Go code.

## Files added/updated

- `README.md` (new root README)
- `docs/openapi.yaml`, OpenAPI/Swagger guide, Postman collection
- HLD, LLD, order/payment state diagrams, design pattern guide
- Project report, screenshot checklist, AWS target-deployment plan, residual-work register
- `.github/workflows/ci.yml` (CI incl. dedicated PostgreSQL service and integration tests)
- `Dockerfile`, `.dockerignore`
- `.env.example` with `JWT_SECRET` documented alongside the existing settings

## Extract

From the repository root:

```bash
unzip -o ~/Downloads/ORDERFLOW_DOCUMENTATION_API_CI_ARCHITECTURE_BUNDLE.zip -d .
```

Check this report of files before committing:

```bash
git status --short
git diff --check
```

## Validate content

```bash
go test ./...
go test -race ./internal/integration/...
go vet ./...
git diff --check
```

Validate the OpenAPI YAML by opening `docs/openapi.yaml` in the Swagger UI command documented in README. Import `docs/postman/OrderFlow.postman_collection.json` in Postman. The Postman signature script requires `mockPaymentWebhookSecret` to equal the local `.env` value; do not share or export a collection populated with tokens or secrets.

For the CI migration check, the workflow creates its own disposable database service; it does not use your local database. GitHub Actions status can only be confirmed after pushing the workflow.

## Review before commit

The bundle intentionally does not add duplicate Go models, duplicate DTOs or repository wrappers: those already exist at the reviewed remote HEAD. It also does not claim Redis, Stripe, AWS, cron or screenshots are complete. Those are listed as explicit residuals in `docs/REMOTE_REVIEW_AND_REMAINING_WORK.md`.

After validation, stage the intended additions/updates:

```bash
git add README.md .env.example Dockerfile .dockerignore .github/workflows/ci.yml \
  docs/openapi.yaml docs/OPENAPI_GUIDE.md docs/DESIGN_PATTERNS.md \
  docs/PROJECT_REPORT.md docs/REMOTE_REVIEW_AND_REMAINING_WORK.md \
  docs/SCREENSHOT_CHECKLIST.md docs/AWS_DEPLOYMENT_PLAN.md docs/diagrams \
  docs/postman/OrderFlow.postman_collection.json ORDERFLOW_FINALIZATION_BUNDLE.md

git diff --cached --check
git diff --cached --stat
git status --short
```

Review that no `.env`, token, generated secret, SQL backup, or unrelated local file is staged. Then commit and push only after you are satisfied with the local validation:

```bash
git commit -m "docs: add OpenAPI architecture report and CI setup"
git push origin main
git log -1 --oneline
git status --short
```
