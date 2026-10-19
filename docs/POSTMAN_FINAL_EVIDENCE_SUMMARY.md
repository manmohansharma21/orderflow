# Final payment and order evidence supplement

Captured 19 October 2026 and intended for the final evidence commit.

- Order history: `docs/evidence/postman/final-captures/14-order-history-200.png`
- Signed SUCCESS callback: `docs/evidence/postman/final-captures/16-payment-success-callback-200.png`
- Recovery-path payment initiation: `docs/evidence/postman/final-captures/17-recovery-payment-initiated-201.png`

Read `docs/postman/MOCK_PAYMENT_WEBHOOK_SECRET_GUIDE.md` for the purpose and safe handling of `MOCK_PAYMENT_WEBHOOK_SECRET`. The success callback has status `SUCCESS`; the recovery-path payment screenshot remains `PENDING` at initiation. It must not be represented as a completed recovery. The manual order history screenshot proves a successful list/read response, not that the particular callback's order was retrieved in the same screenshot.
