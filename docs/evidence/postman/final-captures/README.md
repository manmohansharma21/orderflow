# Final Postman captures (local test environment)

These are real screenshots supplied by the project author from local OrderFlow/Postman requests.

| File | Evidence | Interpretation |
|---|---|---|
| `01-admin-login-200-token-redacted.png` | `POST /auth/login` returned `200 OK`; account role is `ADMIN` | Access token redacted before committing. |
| `02-product-create-201.png` | Admin `POST /products` returned `201 Created` for `OrderFlow Postman Product A` | Positive product-create case. Product ID shown in response is captured by the Postman collection test. |
| `03-inventory-availability-200.png` | `GET /inventory/{productId}` returned `200 OK`; on-hand and available quantities were 50 in this run | Local demo data; values can change. |
| `04-inventory-movement-audit-200.png` | Admin inventory movement list returned `200 OK` with an `OPENING_BALANCE` record | Demonstrates the stock movement audit trail. |

These are local test screenshots, not evidence of a live AWS deployment.
