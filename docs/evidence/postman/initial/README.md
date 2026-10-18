# Initial Postman captures — October 2026

These captures were supplied by the project author. They are preserved for traceability and may support the report selectively.

| File | What it shows | How to use it |
|---|---|---|
| `01-health-200.png` | Postman `GET /health` → `200 OK`, `status=ok` | Suitable initial health evidence |
| `02-readiness-200.png` | Postman `GET /ready` → `200 OK`, `status=ready` | Suitable initial readiness evidence |
| `03-metrics-200.png` | Postman `GET /metrics` → `200 OK` and HTTP counter/histogram exposition | Suitable observability evidence |
| `04-register-customer-201.png` | Customer registration → `201 Created` | Suitable initial authentication-flow evidence |
| `05-login-customer-200-redacted.png` | Customer login → `200 OK`; bearer token and account identifiers redacted | Safe supporting evidence; original was not included because it displayed a live JWT |
| `troubleshooting/06-product-create-401-empty-admin-token.png` | Product POST got `401 bearer token required` with empty/unresolved admin variable | Troubleshooting history only, not successful CRUD evidence |
| `troubleshooting/07-product-create-403-non-admin-token.png` | Product POST got `403 insufficient permissions` | Confirms role enforcement for a non-admin token; troubleshooting history only |

The login screenshot in this folder is a redacted derivative. The raw image with the live JWT is deliberately omitted. Rotate the local `JWT_SECRET` before the next test session to invalidate the token previously shown in the raw screenshot; then log in again. Never commit `.env`, tokens or populated Postman environment exports.

## What remains for final report evidence

Capture corrected successful product administration (`201`), multi-product order creation (`201`), order history (`200`), exact idempotency replay (`200`), changed-payload key conflict (`409`), cancellation (`200` then `409`), payment success/failure/recovery and report endpoints. The failed product-create screenshots above should not be presented as successful create evidence.
