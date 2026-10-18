# Mock-provider recovery and HTTP observability

## Why provider-side state is separate

The normal callback path depends on a notification reaching OrderFlow. A successful provider transaction can therefore be left internally `PENDING` if the callback is delayed or lost. Migration `000008_mock_provider_reconciliation` creates a separate `mock_provider_transactions` ledger for the simulated provider's state. It is a coursework simulation, not an external payment gateway.

A local demo sequence is:

1. Create an order and initiate a payment with the normal authenticated APIs.
2. As an ADMIN in a non-production environment, call `POST /admin/mock-provider/payments/{id}/outcome` with the provider's simulated terminal result. This updates only the simulated provider ledger; the OrderFlow payment remains pending.
3. Call `POST /admin/payments/reconciliation/run?older_than_minutes=0&limit=100`, or wait for the configured periodic worker.
4. The reconciler reads terminal provider outcomes for still-pending payments and sends each one through `PaymentService.ProcessMockCallback`. That shared path validates amount/currency, guards state transitions and performs the same stock consumption/release and movement audit as an ordinary callback.
5. Verify the order through `GET /orders/{id}` and inspect inventory movements as ADMIN.

The simulator endpoint returns 404 when `APP_ENV=production`, and both simulator/recovery routes require an ADMIN JWT. Keep the application environment set explicitly in cloud deployments. Do not treat this local mock ledger as a substitute for a real provider API, signature verification or settlement reconciliation.

A provider amount/currency mismatch is surfaced as an error; the order/payment remain unchanged and the reconciliation worker logs the failure. If a user callback or cancellation already won the race, the reconciler does not repeat the state transition. PostgreSQL remains the source of truth for OrderFlow's order/payment states and stock.

## HTTP observability

- Every HTTP response includes `X-Request-ID`. A safe incoming ID is echoed; otherwise a random ID is generated.
- The application logs the request ID, method, registered Gin route, response status and duration. Registered route patterns are used instead of raw paths to avoid high-cardinality path IDs.
- `GET /metrics` returns Prometheus text exposition for `orderflow_http_requests_total` and `orderflow_http_request_duration_seconds` histogram buckets, sums and counts.
- The histogram contains fixed latency buckets up to 5 seconds. It does not publish calculated P95 directly; compute percentiles from a metrics backend or use the repeatable benchmark script.

## Redis performance result

The local API benchmark supplied during finalization tested 300 requests per variant with 10 concurrent workers on an Apple M2 Mac, Go 1.27.2, Docker Desktop 29.8.2, and the existing local dataset:

| Metric | Redis disabled | Warm Redis cache |
|---|---:|---:|
| Successful requests | 300/300 | 300/300 |
| P50 | 2.480 ms | 4.447 ms |
| P95 | 4.270 ms | 13.966 ms |
| Mean | 3.315 ms | 5.702 ms |
| Throughput | 2821.81 requests/sec | 1697.92 requests/sec |

For this workload, the cache was slower. The project keeps Redis optional but disables it by default until a representative larger-data workload demonstrates a benefit. The result is a valid negative experiment, not an optimization win. Re-run with the benchmark script after changing the workload and record all results without omitting unfavorable measurements.

## Validation

After adding migration 000008, run:

```bash
./scripts/verify-residual-milestone.sh
./scripts/docker-smoke-test.sh
./scripts/benchmark-api-cache.sh
```

Then run the optional Postman folder called **Missed callback reconciliation demo**. Use only disposable development data and redact tokens/secrets before taking final screenshots.
