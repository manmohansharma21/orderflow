# Postman evidence supplement (19 October 2026)

This supplement registers new local captures while distinguishing successful HTTP operations from test setup failures.

- Catalog and inventory: ADMIN login, product A/B creation, public list, inventory availability and movement audit return the successful statuses shown in each screenshot.
- Order: the supplied multi-product `POST /orders` response is `201 Created`, total `74900` paise. It is evidence of order creation only; the order history `GET /orders` success screenshot is still outstanding.
- Payment: payment initiation returned `201 Created` with status `PENDING`, which is expected before callback settlement. The invalid callback signature returned `401`, the expected negative security case.
- Diagnostic-only captures: a status-verification request returned `200` but still showed `PENDING`, with one of two Postman tests failing; another callback request never ran because the local `mockPaymentWebhookSecret` collection variable was unset. Neither is treated as successful settlement.
- Benchmark: a separate API run on 18 October 2026 showed warm Redis improving p50/p95/mean for that run, while a different run showed a regression. This is mixed local evidence, not a robust claim of general speed-up.

The report Word review draft includes the new success and diagnostic screenshots with captions. Before final PDF, add the two genuinely missing screenshots if possible: order history showing the created order, and a signed SUCCESS callback with the order retrieved as `CONFIRMED`. Do not include any screenshot exposing JWTs, passwords, webhook secrets or `.env` values.
