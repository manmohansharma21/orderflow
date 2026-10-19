# OrderFlow final Postman evidence supplement

These PNGs were supplied from the local Postman runs on 19 October 2026 and are retained here for the repository and report. The ADMIN login screenshot is the redacted copy; do not replace it with a raw JWT-bearing screenshot.

## Success and expected-negative evidence

| File | Observed response | Report treatment |
|---|---|---|
| `01-admin-login-200-token-redacted.png` | ADMIN login 200; token redacted | Success evidence |
| `02-product-create-201.png` | Product A create 201 | Success evidence |
| `03-inventory-availability-200.png` | Inventory availability 200 | Success evidence |
| `04-inventory-movement-audit-200.png` | Inventory movement audit 200 | Success evidence |
| `05-product-b-created-201.png` | Product B create 201 | Success evidence |
| `06-product-list-200.png` | Public product list 200 | Success evidence |
| `07-multi-product-order-created-201.png` | Multi-product order create 201; total 74,900 paise | Success evidence for order creation only; it does not prove order-history retrieval |
| `08-order-created-for-payment-201.png` | New payment-test order returned 201 | Success evidence for order creation |
| `09-payment-initiated-matched-order-201.png` | Payment initiation 201 for that order; status PENDING | Matched initiation example; not settlement evidence |
| `10-payment-initiation-independent-run-201.png` | Another payment initiation returned 201, status PENDING | Separate run; keep distinct from the matched pair above |
| `11-invalid-callback-signature-401.png` | Invalid callback signature rejected with 401 | Expected-negative security evidence |
| `troubleshooting/12-order-confirmation-still-pending.png` | GET order returned 200 but status remained PENDING; tests 1/2 | Diagnostic only, not evidence of successful settlement |
| `troubleshooting/13-callback-pre-request-secret-missing.png` | Postman pre-request script stopped because local mock secret was not configured | Client setup issue; callback was not sent to the application |

`docs/evidence/performance/benchmark-results-20261018T230232Z.png` preserves a separate real cache benchmark run. Its warm-Redis values were better than PostgreSQL-only for that run; another run showed the opposite. The report therefore describes results as variable and does not claim a consistent cache speed-up.

## Still missing for the ideal final evidence set

1. A successful `GET /orders?page=1&limit=20` showing the newly created order in its history list.
2. A successfully signed SUCCESS callback followed by a GET order response with `status: CONFIRMED`.

Automated integration tests do cover successful and failed payment transitions, but these two screenshots have not been supplied in this capture set. Do not fabricate them or relabel the pending screenshot as success.
