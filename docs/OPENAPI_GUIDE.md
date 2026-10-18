# OpenAPI / Swagger usage

`openapi.yaml` documents the REST endpoints currently registered by `internal/http/router.go` at remote-review baseline `8ffd04e`.

## Open the spec in a Swagger viewer

From repository root:

```bash
docker run --rm -p 8081:8080 \
  -e SWAGGER_JSON=/spec/openapi.yaml \
  -v "$PWD/docs/openapi.yaml:/spec/openapi.yaml:ro" \
  swaggerapi/swagger-ui
```

Then browse to `http://localhost:8081`. Swagger UI is a separate container; OrderFlow itself does not yet expose `/swagger`.

## Authentication

Login using `/auth/login`, copy `access_token`, and use `Authorization: Bearer <token>` on protected routes. Admin operations need a token whose user role is `ADMIN`. Public registration creates `CUSTOMER` only.

## Mock webhook signing

For `POST /payments/mock-callback`, calculate HMAC-SHA256 over the exact raw JSON bytes using `MOCK_PAYMENT_WEBHOOK_SECRET`. Send `X-Mock-Signature: sha256=<lowercase-hex>`. Do not prettify or reorder the JSON after signing. The Postman collection includes a pre-request script to build the header.
