# OrderFlow final Postman evidence plan

## Why these screenshots are enough

Do not screenshot every API. The report needs a small set of representative, successful end-to-end flows, plus a few security/error examples. Existing health/readiness/metrics, customer registration/login (use the redacted login capture), and 401/403 troubleshooting evidence is already preserved. New evidence in this overlay includes the successful public product-list request (HTTP 200). The collection-folder screenshot is only a navigation reference, not a report claim.

## Current successful public product list

`postman/initial/08-product-list-public-200.png` shows `GET /products?page=1&limit=20` returning `200 OK` and one active product. This endpoint is public; it does not require an ADMIN token.

Product ID shown in the result: `308e809b-1ace-455a-b089-2c4da6be8e23`.

## Fix why the ADMIN promotion returned UPDATE 0

`UPDATE 0` means the email in the SQL query does not exist in the local `orderflow` database. First list users and roles (no passwords are returned):

```bash
docker exec -i orderflow-postgres psql -U orderflow -d orderflow \
  -c "SELECT email, role FROM users ORDER BY role DESC, email;"
```

Use an existing ADMIN if available. If none exists, run the collection's `01 — Register a fresh CUSTOMER` request, copy the exact email returned with `201`, then promote that specific existing email in the *local disposable demo database*:

```bash
docker exec -i orderflow-postgres psql -U orderflow -d orderflow -v ON_ERROR_STOP=1 \
  -c "UPDATE users SET role='ADMIN' WHERE email='PASTE_THE_REGISTERED_EMAIL_HERE';"
```

It must return `UPDATE 1`. Set the Postman Current Value variables `adminEmail` and `adminPassword` to that exact registered account, then run `Login existing ADMIN` again. Re-login after promotion because the JWT includes a role claim. Do not do this in production.

## Create order history now, without needing product creation

The current product list has one active product. Verify stock first:

`GET http://localhost:8080/inventory/308e809b-1ace-455a-b089-2c4da6be8e23`

If `available` is greater than zero, run `POST /orders` in Postman using the valid `customerToken`, Content-Type `application/json`, and header `Idempotency-Key: orderflow-demo-20261019-001`. Use this body:

```json
{
  "product_id": "308e809b-1ace-455a-b089-2c4da6be8e23",
  "quantity": 1
}
```

Expected `201 Created`. Save the `id` from the response. Then run `GET /orders?page=1&limit=20` and `GET /orders/{id}` using the same customer token; both should return `200 OK`. If stock is zero, use the ADMIN inventory receipt request first or select an active product with stock.

Capture two screenshots: `postman/02-order-created-201.png` and `postman/03-order-history-200.png`. Show request path, status and relevant response JSON; keep Authorization hidden.

## Minimum final report screenshot set

1. Health/readiness and metrics (already captured).
2. Authentication/RBAC: login/profile success plus an intentional 401 and 403 (existing troubleshooting screenshots can be referenced; the login image must be redacted).
3. Product public list 200 (new screenshot in this overlay), and one successful ADMIN product creation 201 if easy.
4. Order create 201 and order history 200 (capture after steps above).
5. One payment SUCCESS/order-confirmed screenshot, and one missed-callback recovery/report screenshot because payments are a key feature.
6. Current green GitHub Actions run for the final evidence commit. `ci/historical-green-2654070.png` only documents the earlier green run for SHA `2654070`; recapture the latest SHA after the final evidence commit.
7. Redis benchmark (already captured). Report the actual measurements without claiming a guaranteed speedup; the two local runs differed.

No need to capture every route or every error case. Keep only readable evidence that supports a report claim.

## Token safety

The previously shared login screenshot exposed a JWT. Keep only the already-redacted login image; do not stage screenshots with a token. If that token may still be valid, rotate the local `JWT_SECRET` in `.env`, restart the local API, then log in again. Never commit `.env`, credentials or populated Postman environment exports.

## Swagger/OpenAPI

The repository includes `docs/openapi.yaml`, an OpenAPI 3.0 contract. The Go application does not expose a built-in `/swagger` route. To view interactive Swagger UI, import this YAML in Swagger Editor. This is optional evidence; the API contract and collection are already in the repository.
