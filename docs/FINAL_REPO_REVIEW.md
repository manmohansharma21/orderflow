## Source of truth

Use `main` as the source of truth. Do not copy an old incremental ZIP over the repository. The source archive for submission must be downloaded from the final pushed `main` after all evidence files are committed and CI is green.

## Diagram decision

Keep current browser-safe rendered PNGs and their editable DOT sources as the canonical README/report figures. Keep current Mermaid `.mmd` sources as editable Mermaid alternatives. The `docs/diagrams/archive/` copies preserve useful earlier Mermaid variants from commits `60c255b` and `1515cb0`; these are historical references and must not replace the current rendered images.

The current application is a modular monolith: one Go executable (`cmd/api`) hosts Gin routes, handlers, services and workers in one process. REST endpoints are API interfaces, not proof of microservices. PostgreSQL and optional Redis are supporting data services.

## Guided Postman collection

`docs/postman/OrderFlow_Guided_Evidence.postman_collection.json` contains 49 requests grouped into scenarios and covers 28 distinct registered method/path route patterns (some requests repeat a route to demonstrate different outcomes such as 201/200/409/401/403/404). It is broad coverage, not every possible edge case: callback amount/currency mismatch (422), cross-owner order access, every stock-conflict variant and every expiry timing case are not all demonstrated in this 49-request collection. Run requests sequentially because IDs/tokens are saved by tests. Use local demo credentials and a local webhook secret only in Postman's Current Value; never export populated secrets.

Postman is recommended because it is fastest for request/response screenshots. Login is required for protected order/payment/admin APIs; the collection automates customer registration/login token handling and performs one admin login using your existing local ADMIN account. If the admin login fails, stop to fix local test credentials rather than removing the security middleware.

## Screenshot checklist

Already have useful evidence for `/health` and `/ready` HTTP 200, `/metrics`, a development server log with a `POST /orders` 201, and two product-list API benchmark runs. Still capture actual Postman flows for auth/RBAC, product administration, inventory, order/idempotency/cancellation, payment success/failure/recovery and reports; latest green GitHub Actions run; and fresh Docker smoke output. Save PNGs under `docs/evidence/` using `docs/evidence/POSTMAN_SCREENSHOT_RUNBOOK.md` and `docs/SCREENSHOT_CHECKLIST.md`.

## README

The README should describe the software, architecture, endpoints, setup, tests, benchmarks, limits and links to docs. User-specific Scaler submission instructions should live in `docs/FINAL_SUBMISSION_STATUS.md`, not in the general public README. Keep content that helps a reviewer run the project; remove only stale HEAD references, duplicate diagram blocks, and portal-specific checklist text.
