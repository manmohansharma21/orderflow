# OrderFlow remote review and remaining-work tracker

**Reviewed remote HEAD:** `1515cb0` (`feat: complete provider recovery and HTTP observability`).
**CI:** successful: https://github.com/manmohansharma21/orderflow/actions/runs/39160539244

## Verified at this revision

- Go/Gin handlers and request DTOs, domain models, services, repositories, config, `pgxpool`, and migrations 000001–000008.
- Authentication and role checks, product CRUD/search, inventory receipt/adjustment/audit history.
- Multi-product transactional orders, user-scoped order list/detail/cancel, order idempotency and concurrent reservation protections.
- Mock payment initiation, HMAC-signed callback, amount/currency checks, failure reason, payment idempotency, replay safety, success inventory consumption and failed/cancelled/expired reservation release.
- Separate mock-provider ledger; known terminal outcome recovery after a missed callback; scheduled expiry and reconciliation workers.
- Admin order/sales/inventory reports, Redis product list/detail caching, request IDs, HTTP metrics, OpenAPI, Postman and editable HLD/LLD/state diagrams.
- CI passes migration application, unit/integration tests, Redis tests, race integration checks, vet, whitespace check, Docker image build and `/health` / `/ready` smoke test. Local reliability verification and Docker smoke also passed in the supplied terminal logs.

## Still required before submission

| Task | Status | Action |
|---|---|---|
| README/documentation cleanup | Documentation overlay prepared | Apply `ORDERFLOW_README_SCREENSHOT_GUIDE_PATCH.zip`, review, commit/push and verify the next CI run |
| Postman walkthrough/screenshots | Pending | Use `docs/SCREENSHOT_CHECKLIST.md`; redact tokens/secrets |
| Docker/CI screenshots | Pending | Capture successful local smoke and green run for the final submitted SHA |
| Product-list API performance | Measured | Redis warm-cache run was slower; retain negative result honestly |
| Order-history benchmark | Incomplete because the previous run was interrupted | Run bounded `BenchmarkOrderHistoryList` independently and save output if the report requires it |
| Word/PDF report | Word review draft available; placeholders remain | Insert screenshots, refresh TOC/list-of-figures/page numbers, complete real dates, review references and export/verify final PDF |
| Full source ZIP for portal | Pending final docs/evidence commit | Download from GitHub `main` after final push and green CI; do not upload incremental patches |
| Live AWS deployment | Not deployed; rubric-dependent | Check assignment/contest rubric and obtain cost approval before provisioning |
| Previous Lambda submission | Pending review | Revisit when the earlier deployment repository/artifact is available |
| Understand CI config | Guide committed | Work through `.github/workflows/ci.yml` with `docs/CI_CONFIGURATION_WALKTHROUGH.md` |
| Post-submission architecture walkthrough | Pending | Explain request flow, transactions, repositories, idempotency, payment recovery, cache and metrics |

## Performance facts that must be preserved

The committed 300-request API benchmark (10 concurrent workers on Apple M2 / Go 1.27.2 / Docker Desktop 29.8.2) measured Redis-disabled P50/P95/mean/throughput of 2.480 ms / 4.270 ms / 3.315 ms / 2,821.81 req/s. Warm Redis measured 4.447 ms / 13.966 ms / 5.702 ms / 1,697.92 req/s. Both variants returned 300/300 successes; Redis was slower for this workload. Service-level order creation values were captured, but the order-history benchmark attempt was interrupted and must not be reported as completed.

## Scope guardrails

The payment provider is a deterministic simulator, not Stripe. The AWS architecture is a target design, not evidence of deployed resources. The portal screenshot shows a repo link, project ZIP and separate report PDF; no live AWS URL field is displayed. Confirm a separate rubric requirement before incurring AWS costs. The submission report must distinguish verified code from pending screenshots/administrative dates.
