# Postman admin-token troubleshooting (local OrderFlow only)

## What the two errors mean

- `401 Unauthorized` with `{"error":"bearer token required"}` means Postman did not send a usable Bearer token. The screenshot showed `Authorization: Bearer {{adminToken}}`, but the variable was empty or unresolved.
- `403 Forbidden` with `{"error":"insufficient permissions"}` means a token was present and valid, but its role was not `ADMIN` (typically the CUSTOMER token was used).
- `GET /products` is public and should return `200` without any token. The protected operation is `POST /products` (create), as well as `PATCH` and `DELETE`; those need a valid ADMIN token.

## Correct sequence

1. Open **OrderFlow API Verification — Local** in Postman.
2. Open folder **01 — Authentication and RBAC** and run requests top-to-bottom.
3. Run **01 Register CUSTOMER**, then **02 Login CUSTOMER**. The collection saves `customerToken`.
4. Run **05 Login existing ADMIN**. Do not continue until the request returns `200 OK`, the response's `user.role` is `ADMIN`, and the test `admin token saved only after role verification` passes.
5. The collection clears any stale admin token before admin login. It stores `adminToken` only if role `ADMIN` is confirmed. Admin requests now fail fast before sending if that check has not succeeded.
6. In **02 — Product catalog/admin**, run **01 List products (expect 200)** first. This is public and does not need a token. Then run **02 Create product A** and **03 Create product B** only after the admin login passes.
7. After Product A and B return `201`, open folder **04 — Multi-product orders and idempotency**. Run **02 Create MULTI-PRODUCT order** (`201`), then **05 Order history** (`200`) and **06 Order details** (`200`). This creates genuine local order history; no separate seed is necessary.

## If the existing admin login fails

First check the collection's `adminEmail` and `adminPassword` Current Values. Defaults are historical local test credentials and might not exist in a recreated database. Do not post passwords/tokens to chat or commit them.

If there is no usable local ADMIN test account, create one intentionally in your disposable local database:

1. In Postman, send `POST http://localhost:8080/auth/register` with JSON body such as:
   `{"email":"of.local.admin.20261019@example.com","password":"OrderFlowAdminDemo2026"}`
   Use a new email if this one already exists. Registration correctly creates a CUSTOMER, not an ADMIN.
2. Promote only this local demo account in the local Compose database:
   ```bash
   docker exec -i orderflow-postgres psql -U orderflow -d orderflow -v ON_ERROR_STOP=1 \
     -c "UPDATE users SET role='ADMIN' WHERE email='of.local.admin.20261019@example.com';"
   ```
   Confirm the command updates exactly one row. If your local Compose container/database names differ, inspect `docker compose ps` and adapt them. Never run this against production or any shared database.
3. Set the collection Current Values `adminEmail` and `adminPassword` to that local account. Rerun **05 Login existing ADMIN**, and confirm response role `ADMIN` before proceeding.

## Token screenshot security

A previously submitted login screenshot showed a bearer JWT. The copy preserved in `docs/evidence/postman/initial/` is redacted. Treat the visible token as exposed: rotate the local `JWT_SECRET` in `.env` and restart the API before continuing; this invalidates existing tokens and means you must log in again. Do not commit `.env` or the original unredacted screenshot.

## Capture policy

The captured 401 and 403 product-create screenshots are **troubleshooting history**, not successful product-creation evidence. Keep them for traceability, but don't use them to claim admin product creation succeeded. For the final report, capture the corrected flow: public list `200`, admin login `200` with token hidden, admin product create `201`, duplicate SKU `409`, update `200`, deactivate `200` and inactive detail `404`.
