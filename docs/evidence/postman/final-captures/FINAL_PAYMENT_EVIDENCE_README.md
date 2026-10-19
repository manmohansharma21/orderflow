# Final Postman evidence captures — 19 October 2026

These files are genuine screenshots supplied from the local Postman run. They are added as documentation evidence, not application source changes.

| File | What it shows | Evidentiary scope |
|---|---|---|
| `14-order-history-200.png` | `GET /orders?page=1&limit=20` returned HTTP 200 with a pending order in the list. | Order history list/read success; not proof of settlement. |
| `16-payment-success-callback-200.png` | Signed mock SUCCESS callback returned HTTP 200 and payment status `SUCCESS`. | Successful callback for the displayed payment/order IDs. |
| `17-recovery-payment-initiated-201.png` | Recovery-path payment initiation returned HTTP 201 and payment status `PENDING`. | Shows recovery demonstration setup only; not proof the missing callback was recovered. |

Keep the screenshot set alongside `docs/postman/MOCK_PAYMENT_WEBHOOK_SECRET_GUIDE.md`. The application’s automated PostgreSQL integration/reconciliation tests provide additional evidence for order confirmation and missed-callback recovery. No live AWS deployment or real payment-provider transaction is claimed.
