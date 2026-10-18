# Final repository cleanup, benchmark evidence and Postman guide

## What this overlay adds

- Adds the user's benchmark screenshot and an adjacent text transcript.
- Preserves historical Mermaid HLD/LLD/state variants under `docs/diagrams/archive/`. The current rendered PNG/SVG plus DOT sources remain canonical for the README and report.
- Adds the guided Postman collection (49 requests) and atomic runbook. It covers 28 distinct registered method/path patterns; 49 means individual requests/scenarios, not 49 different route definitions.
- Adds an idempotent script to remove the personal Scaler portal checklist from the public README, remove the repeated state diagram reference, refresh stale HEAD references, and update the report-source performance section with both measured runs.

## Extraction and apply

From the repo root:

```bash
unzip -o ~/Downloads/ORDERFLOW_FINAL_EVIDENCE_AND_CLEANUP.zip -d .
python3 scripts/apply-final-doc-cleanup.py
python3 -m json.tool docs/postman/OrderFlow_Guided_Evidence.postman_collection.json >/dev/null

git diff --check
go test ./...
go test -race ./internal/integration/...
go vet ./...
```

Review `git status --short` and `git diff --stat` before staging. Do not stage `.env`, secrets, database dumps, local report draft binaries, or populated Postman environment exports.

## Postman import

Import only `docs/postman/OrderFlow_Guided_Evidence.postman_collection.json`. The ZIP includes an atomic request-by-request runbook. Protected routes require JWTs; the collection automates fresh customer registration/login and logs into an existing local admin account once. If the admin login fails, update local Postman Current Value to an existing ADMIN account; do not disable authorization to obtain screenshots.

## Monolith vs microservices

OrderFlow is a modular monolith because one deployable Go/Gin process hosts the HTTP handlers, services, repositories and workers. REST APIs are interfaces, not deployment boundaries. PostgreSQL and optional Redis are supporting infrastructure services, not separate OrderFlow business microservices.
