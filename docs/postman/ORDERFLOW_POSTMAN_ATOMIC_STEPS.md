# OrderFlow Postman: atomic local evidence runbook

This is a **local development/demo collection** for `http://localhost:8080`. It is not for production. It targets the current OrderFlow routes on GitHub `main` and uses response tests to check expected HTTP codes.

## 0. Start prerequisites

From the repository root in Terminal A:

```bash
docker compose up -d postgres redis
go run ./cmd/api
```

Keep this terminal open. In Terminal B confirm:

```bash
curl -i http://localhost:8080/health
curl -i http://localhost:8080/ready
```

Both should return `HTTP/1.1 200 OK`. If port 8080 is already in use, do not start a second API process. Diagnose with `lsof -nP -iTCP:8080 -sTCP:LISTEN`.

## 1. Import collection

1. Download `OrderFlow_Guided_Evidence.postman_collection.json` from this folder.
2. Open Postman → **Import** → select the JSON file.
3. Expand the collection **OrderFlow — Guided Local API Evidence Run**.
4. Open the collection's **Variables**. The `baseUrl` should be `http://localhost:8080`.
5. The default admin credentials are the local test user previously used in this repo (`of.admin.test.20261018@example.com`). If the login request fails, update `adminEmail` and `adminPassword` in Postman's **Current value** to the existing ADMIN account in your local `orderflow` database. Do not paste credentials into chat or commit populated values.
6. For signed webhook requests, set `mockPaymentWebhookSecret` in Postman's **Current value** to the exact local `MOCK_PAYMENT_WEBHOOK_SECRET` from `.env`. Keep it local; never share screenshots showing it, export populated secrets, or commit it. It must be at least 32 characters.

## 2. Run manually in order

**Run one request at a time, top to bottom. Do not click Run collection for all requests on the first attempt**: requests create data and several later requests depend on values captured by tests.

### Folder 00 — Health and observability

Run the three requests. Expected `200` for `/health`, `/ready`, `/metrics`. Screenshot: `docs/evidence/browser/01-health-readiness-http-200.png` and `docs/evidence/observability/01-request-metrics.png` (existing ones are already captured; only replace with newer clean screenshots if useful).

### Folder 01 — Authentication and RBAC

1. Register fresh CUSTOMER → `201`; test auto-generates a unique email.
2. Login CUSTOMER → `200`; token is saved in `customerToken` automatically.
3. Profile without token → `401`.
4. Profile with CUSTOMER token → `200`.
5. Login existing ADMIN → `200`; saves `adminToken`.
6. ADMIN route using customer token → `403`.
7. ADMIN route with admin token → `200`.

If step 5 fails, stop and correct the local admin variable values. The collection will not promote a user to ADMIN automatically.

### Folder 02 — Product catalog/admin

Run in order. Product A and B create with 50 initial units each; tests save `productId`, `productId2`, `sku`, and `sku2`. Expected statuses: list `200`, create A/B `201`, duplicate SKU `409`, search `200`, details `200`, patch `200`.

### Folder 03 — Inventory and audit

Run in order. Expected `200` for availability, stock receipt, stock adjustment, movement list. Receipts and adjustments modify local test data, which is intended for this local demo.

### Folder 04 — Multi-product orders/idempotency

Run in order. Expected statuses: unauthenticated order `401`; create multi-product order `201`; exact same body/key retry `200` with same order ID; reuse that key with a changed quantity `409`; order history/detail `200`; cancel once `200`; repeat cancellation `409`.

### Folder 05 — Payment success and failure

Run in order. Each payment path creates its own order. Expected statuses: create order `201`; initiate payment `201`; intentionally bad signature `401`; signed success callback `200`; confirmed order `200`; second order/payment `201`; signed failed callback `200`; reconciliation report `200`.

Signed callback requests create a fresh event ID and HMAC-sign the exact resolved raw request body using Postman's `CryptoJS`. If the signed callback returns `401`, verify the secret in Postman Current Value matches `.env` exactly; don't remove signature validation.

### Folder 06 — Missed callback recovery

Run all five requests in the order displayed, manually one-by-one. This is distinct from a normal webhook callback:

1. Customer creates an order (`201`).
2. Customer starts a payment (`201`).
3. Admin sets the simulated provider outcome to SUCCESS without calling the webhook (`200`).
4. Admin runs immediate reconciliation (`200`, `recovered_count` at least 1).
5. Customer verifies the order (`200`, status `CONFIRMED`).

This uses a development-only simulator and is not a real payment gateway. Do not use in a production environment.

### Folder 07 — Reports and deactivation

Admin orders, sales and inventory reports each return `200`. Then a disposable product is created (`201`), deactivated (`200`), and then fetch returns `404`. Do not use Product A or B for deactivation because they were used in order demonstrations.

## 3. Screenshots and repository paths

Use macOS **Command + Shift + 4** to capture a rectangular area. Capture the request name/URL, response status and useful JSON body. Do not show Authorization headers, tokens, passwords, webhook signature, secret variable values, or any `.env` terminal output.

Save final captures under:

- `docs/evidence/postman/01-auth-profile-and-authorization.png`
- `docs/evidence/postman/02-product-admin.png`
- `docs/evidence/postman/03-inventory-audit.png`
- `docs/evidence/postman/04-order-lifecycle.png`
- `docs/evidence/postman/05-payment-callbacks.png`
- `docs/evidence/postman/06-reconciliation-and-reports.png`
- `docs/evidence/postman/07-admin-reports-and-deactivation.png`

One image per flow is enough if the status + representative response can be read. For a flow that needs multiple response codes, capture the relevant calls separately or use the Postman Runner summary, but do not fabricate a composite claiming a single response represented multiple requests.

After saving screenshots, inspect the staged diff before commit:

```bash
git status --short
git diff --check
git add docs/evidence/postman docs/postman/OrderFlow_Guided_Evidence.postman_collection.json docs/postman/ORDERFLOW_POSTMAN_ATOMIC_STEPS.md
git diff --cached --check
git diff --cached --stat
```

Never stage `.env`, database dumps, Postman populated environment exports, tokens or webhook secrets. When reviewed:

```bash
git commit -m "docs: add guided Postman API evidence"
git push origin main
git log -1 --oneline
git status --short
```

Then verify a green GitHub Actions run for the new commit.

## Expected status code quick reference

| Scenario | Expected code |
|---|---:|
| Health/readiness/metrics | `200` |
| Register customer | `201` |
| Login | `200` |
| Profile without token | `401` |
| Customer token on admin route | `403` |
| Admin creates product | `201` |
| Duplicate SKU | `409` |
| Create order | `201` |
| Same idempotency key + identical body | `200` |
| Same idempotency key + different body | `409` |
| Cancel order first time / replay | `200` / `409` |
| Initiate payment | `201` |
| Invalid mock callback signature | `401` |
| Signed success/failure callback | `200` |
| Callback amount/currency mismatch | `422` (use separate manual request if capturing this negative case) |
| Admin report | `200` |

The default architecture is a **modular monolith**: a single Go/Gin process with modules and REST endpoints. It is not a microservices deployment. PostgreSQL and Redis are backing services, not separate OrderFlow business microservices.
