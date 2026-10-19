# OrderFlow — final callback and evidence steps

This overlay preserves the latest evidence. The order-history screenshot is a successful `200 OK`, but shows a single-item pending order; the payment-initiation screenshot is `201 Created` with `PENDING`; the callback screenshot is troubleshooting evidence only. The callback did not reach the API because a Postman pre-request guard stopped it when `mockPaymentWebhookSecret` was empty.

## 1. Set the callback secret in Postman without printing it

In macOS Terminal:

```bash
cd ~/Projects/orderflow
grep -c '^MOCK_PAYMENT_WEBHOOK_SECRET=' .env
```

Expected output: `1`. Then copy the value to your clipboard without printing it:

```bash
sed -n 's/^MOCK_PAYMENT_WEBHOOK_SECRET=//p' .env | tr -d '\r' | pbcopy
```

Now in Postman select the **same collection shown in the callback request breadcrumb**. Your screenshot shows the older `OrderFlow — Guided Local API Evidence Run` collection. Click its root, open **Variables**, find `mockPaymentWebhookSecret`, paste the clipboard into **Current Value** (not Initial Value), and Save. The top-right `No environment` state is okay; this is a collection variable.

Never paste the secret into chat, screenshots, shared values, or an exported Postman collection. If the grep count is not `1`, stop and inspect the key name in `.env` locally.

## 2. Use the IDs from one collection only

Postman variables are collection-specific. Stay in the collection that created the current payment. Before sending the callback, ensure these Current Values are non-empty:
- `paymentId`
- `paymentAmountPaise`
- `paymentCurrency`
- `paymentOrderId`

If any is empty or from another collection, re-run in this SAME collection:
1. `01 — Create a fresh order for SUCCESS payment` (201)
2. `02 — Initiate payment` (201; stores `paymentId`, amount, currency)

Do not manually change the callback request body. The script signs the body using HMAC-SHA256 and adds the `X-Mock-Signature` header.

## 3. Send signed SUCCESS, then verify order status

Run `04 — Signed SUCCESS callback` in the same payment folder. Expected `200 OK` and payment status `SUCCESS`.

Immediately run `05 — Verify successful order is CONFIRMED`. It requests `GET /orders/{{paymentOrderId}}`; expected `200 OK` and `"status": "CONFIRMED"`.

If the pre-request guard still blocks, the variable was not saved in the active collection. If it reaches the API but responds HTTP `401`, the secret in Postman does not match the one loaded by the running server. Verify locally that both come from the same `.env`; if you changed `.env` after starting `go run ./cmd/api`, restart the server, log in again, and retry. Don't share the secret/token.

Capture:
- `docs/evidence/postman/final-captures/16-payment-success-callback-200.png`
- `docs/evidence/postman/final-captures/17-payment-order-confirmed-200.png`

The supplied order-history screenshot proves the endpoint returns 200, but to specifically demonstrate history containing a multi-product order, use the same collection that created the multi-product order, rerun Order history after that order, and capture the list showing that order.
