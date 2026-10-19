# OrderFlow — final local Postman steps

## Current state confirmed by the supplied logs/screenshots

- Local database already has `of.admin.test.20261018@example.com` with role `ADMIN`.
- That ADMIN login returned `200 OK`; `user.role` was `ADMIN`.
- Admin `POST /products` returned `201 Created` and the collection has `productId` from the response.
- Inventory availability and admin inventory movement-list requests returned `200 OK`.

No extra role change is required. If you want to explicitly re-apply the role in the **local development database only**, the known existing account is:

```bash
docker exec -i orderflow-postgres psql \
  -U orderflow -d orderflow -v ON_ERROR_STOP=1 \
  -c "UPDATE users SET role='ADMIN' WHERE email='of.admin.test.20261018@example.com';"
```

It should report `UPDATE 1` because the email exists. However, the latest query already showed this account is ADMIN, so this command is redundant. Do not run a role update against production or a shared database.

## Postman variables / why request ordering matters

Use the collection `docs/postman/OrderFlow_API_Verification.postman_collection.json`; it is a 49-request collection with 28 distinct HTTP method/path patterns. It uses collection variables, so an Environment is not required. The collection scripts automatically save these IDs after successful responses:
- customer login saves `customerToken`
- admin login saves `adminToken` only when login returns 200 and `user.role == ADMIN`
- product A/B creation saves `productId` and `productId2`
- multi-product order creation saves `orderId`
- payment initiation saves `paymentId`, `failurePaymentId`, or `recoveryPaymentId` in the corresponding path

A Postman request references the saved variables. Don't manually paste IDs or tokens if the prior request succeeded. Don't run dependent requests in parallel or out of order.

## Exact next sequence

1. `01 — Authentication and RBAC / 01 — Register a fresh CUSTOMER`
   - Expected `201`; creates a unique email automatically.
2. `01 — Authentication and RBAC / 02 — Login CUSTOMER`
   - Expected `200`; test saves `customerToken`.
3. `01 — Authentication and RBAC / 05 — Login existing ADMIN`
   - Expected `200`; user role must be `ADMIN`; test saves `adminToken`.
4. `02 — Product catalog/admin / 01 — List products`
   - Expected `200`; public endpoint, no token required.
5. The Product A create request was already successful (`201`) in the screenshot. Do not repeat unnecessarily. If needing Product B, run `03 — Create product B for multi-product order` as ADMIN; expected `201`, saves `productId2`.
6. `03 — Inventory and audit / 01 — Availability`
   - Already captured as `200`; it uses `productId`.
7. `03 — Inventory and audit / 04 — Inventory movements audit`
   - Already captured as `200`; it requires `adminToken`.
8. `04 — Multi-product orders and idempotency / 01 — Create order WITHOUT token`
   - Intentional negative test, expected `401`; you can skip recapturing it.
9. `04 — Multi-product orders and idempotency / 02 — Create MULTI-PRODUCT order`
   - Ensure `customerToken`, `productId`, and `productId2` are set. Expected `201`; response test saves `orderId`.
10. `04 — Multi-product orders and idempotency / 03 — Retry identical order and key`
    - Expected `200`; same order ID.
11. `04 — Multi-product orders and idempotency / 05 — Order history`
    - Expected `200`; uses `customerToken`.
12. `04 — Multi-product orders and idempotency / 06 — Order details`
    - Expected `200`; uses saved `orderId`.
13. Payment success path, in order:
    - `05 / 01 — Create a fresh order for SUCCESS payment` (`201`, saves `paymentOrderId`)
    - `05 / 02 — Initiate payment` (`201`, saves `paymentId`, amount and currency)
    - `05 / 03 — Invalid callback signature` (`401`, intentional)
    - `05 / 04 — Signed SUCCESS callback` (`200`, uses local `MOCK_PAYMENT_WEBHOOK_SECRET`)
    - `05 / 05 — Verify successful order is CONFIRMED` (`200`)
14. Optional missed-callback demo:
    - Run folder `06 — Missed callback recovery` sequentially; the outcome test requires `adminToken`, while the customer order requests require `customerToken`. Expected final order status is `CONFIRMED`.

## Important if section 04 gets 401/403

- `401 bearer token required`: `customerToken` is empty/not attached. Re-run Register then Login CUSTOMER; confirm the login test is green.
- `403 insufficient permissions`: the protected route received a valid non-admin token. Use the admin login only for admin routes; use customer token for customer order APIs.
- `GET /products` is public. `POST/PATCH/DELETE /products` require ADMIN.
- The `01 — Create order WITHOUT token` request in folder 04 is intentionally expected to return 401. That failure is correct evidence, not a collection error.

Never export populated collection variables or show JWTs, passwords, webhook signatures, `.env`, or database credentials in screenshots.
