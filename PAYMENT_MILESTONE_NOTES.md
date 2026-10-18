# OrderFlow Payment Milestone

This bundle adds simulated payment initiation, an HMAC-authenticated mock callback,
idempotency-key handling, and a read-only reconciliation report.

## Routes

- `POST /orders/:id/payments` — authenticated customer; requires `Idempotency-Key`.
- `POST /payments/mock-callback` — mock provider callback; requires
  `X-Mock-Signature: sha256=<hex HMAC-SHA256 of the exact request body>`.
- `GET /admin/payments/reconciliation?older_than_minutes=15` — ADMIN-only report.

## Payment behavior

- Starting a payment creates a `MOCK` payment with status `PENDING`.
- Repeating the same key for the same order returns the existing payment.
- A `SUCCESS` callback moves payment to `SUCCESS` and order to `CONFIRMED`.
- A `FAILED` callback moves payment to `FAILED`, order to `FAILED`, releases
  reserved stock, and records `ORDER_RELEASED` movements in the same transaction.
- Repeating a callback with the same terminal outcome is idempotent.
- Reconciliation reports stale pending payments and inconsistent order/payment states.
- No real payment provider or money movement is connected.

## Setup

Add a local-only secret to `.env` (do not commit `.env`):
`MOCK_PAYMENT_WEBHOOK_SECRET=<random value of at least 32 bytes>`

The migration `000006_payment_workflow.up.sql` is additive and supports the legacy
local schema with `provider_reference` as well as the corrected clean-install schema.
Review database state and back up before applying it.
