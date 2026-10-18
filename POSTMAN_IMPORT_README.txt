OrderFlow Postman guided pack
=============================

1. Import docs/postman/OrderFlow_Guided_Evidence.postman_collection.json into Postman.
2. Follow docs/postman/ORDERFLOW_POSTMAN_ATOMIC_STEPS.md exactly, one request at a time.
3. Use only localhost and the local development database.
4. Set mockPaymentWebhookSecret only in Postman's local Current Value. Never commit or share populated secrets.
5. Don't run the whole collection automatically on first use: requests have dependencies and create local demo data.
