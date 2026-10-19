# `MOCK_PAYMENT_WEBHOOK_SECRET`: purpose and safe local setup

## Purpose

OrderFlow uses `MOCK_PAYMENT_WEBHOOK_SECRET` to authenticate its **simulated payment callback** in local development. The Postman request signs the exact callback JSON request body with HMAC-SHA256 and sends the result in the `X-Mock-Signature` header as `sha256=<hex digest>`. The Go API calculates the expected signature using the configured secret and rejects an invalid signature (HTTP 401).

This is not a Stripe/Razorpay credential, not a JWT signing key, and not evidence that real money moved. It exists to exercise signed webhooks safely without a real payment provider.

## Where does the value come from?

A local value is generated independently and stored in `.env`. If the variable is already present and the callback succeeds, **do not rotate or regenerate it just for the report**. To create a new value for a fresh local setup:

```bash
openssl rand -hex 32
```

Put its output in the ignored local `.env` file:

```dotenv
MOCK_PAYMENT_WEBHOOK_SECRET=<generated-random-hex-value>
```

Keep this separate from `JWT_SECRET`. The API reads configuration at process startup; restart the local API after changing `.env`.

## Configure Postman

1. Use the same Postman collection that contains the signed-callback request (the local verification collection).
2. Open the root collection's **Variables** editor.
3. Put the exact local value in **Current Value** for `mockPaymentWebhookSecret`.
4. Leave **Initial Value** blank; don't export a populated collection and don't commit the value.
5. Run `04 — Signed SUCCESS callback` after running the payment initiation request in the *same collection*, so its `paymentId`, `paymentAmountPaise`, `paymentCurrency`, and `eventSuffix` are populated.
6. A successful response shows HTTP `200` and payment status `SUCCESS`. The supplied capture from 19 October 2026 records this successful result.

If the UI suggests a dedicated secret/Vault feature, note that the current Postman pre-request script reads `pm.collectionVariables.get('mockPaymentWebhookSecret')`. A Vault value alone is not read by that script unless the script is explicitly changed. For this local-only flow, set the active collection's Current Value and keep it out of Initial Value/export. Do not put the value into the HTTP body or manually invent the signature.

## Security checklist

- Never commit `.env`.
- Never share the webhook secret, a populated Postman environment/collection export, or a live bearer JWT.
- Don't put the secret in a screenshot. The HTTP callback screenshot should show the response only.
- Use a separate random secret from `JWT_SECRET`.
- For production, use a managed secret store and a real provider's documented signature verification; this local mock is not production payment integration.
