#!/usr/bin/env python3
"""Apply final review edits to the current OrderFlow tree with guarded anchors."""
from pathlib import Path
import json
import re

ROOT = Path('.')

def replace_once(path: str, old: str, new: str) -> None:
    p = ROOT / path
    if not p.exists():
        raise SystemExit(f"ERROR: expected existing file {path} was not found")
    text = p.read_text()
    count = text.count(old)
    if count != 1:
        raise SystemExit(f"ERROR: expected exactly one patch anchor in {path}, found {count}: {old[:100]!r}")
    p.write_text(text.replace(old, new, 1))

# HTTP request IDs, latency histogram metrics, and demo-only provider outcome controls.
replace_once(
    'internal/http/router.go',
    '"github.com/manmohansharma21/orderflow/internal/repository"\n\t"github.com/manmohansharma21/orderflow/internal/service"',
    '"github.com/manmohansharma21/orderflow/internal/observability"\n\t"github.com/manmohansharma21/orderflow/internal/repository"\n\t"github.com/manmohansharma21/orderflow/internal/service"',
)
replace_once(
    'internal/http/router.go',
    'router := gin.New()\n\trouter.Use(gin.Logger(), gin.Recovery())\n\trouter.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })',
    'router := gin.New()\n\tmetrics := observability.NewHTTPMetrics()\n\trouter.Use(gin.Logger(), gin.Recovery(), metrics.Middleware())\n\trouter.GET("/metrics", metrics.Handler())\n\trouter.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })',
)
replace_once(
    'internal/http/router.go',
    'router.GET("/admin/payments/reconciliation", AuthRequired(), RequireRole("ADMIN"), payments.reconciliation)',
    'router.GET("/admin/payments/reconciliation", AuthRequired(), RequireRole("ADMIN"), payments.reconciliation)\n\t\tmockProvider := newMockProviderHandler(paymentService)\n\t\trouter.POST("/admin/mock-provider/payments/:id/outcome", AuthRequired(), RequireRole("ADMIN"), mockProvider.setOutcome)\n\t\trouter.POST("/admin/payments/reconciliation/run", AuthRequired(), RequireRole("ADMIN"), mockProvider.recoverKnownOutcomes)',
)

# Enable recovery of known provider-side outcomes within the periodic worker.
replace_once(
    'internal/worker/payment_reconciliation.go',
    'type ReconciliationReporter interface {\n\tReconciliationReport(context.Context, int) (domain.PaymentReconciliationReport, error)\n}\n',
    'type ReconciliationReporter interface {\n\tReconciliationReport(context.Context, int) (domain.PaymentReconciliationReport, error)\n}\n\ntype MockProviderOutcomeRecoverer interface {\n\tReconcileKnownMockProviderOutcomes(context.Context, int, int) (int, error)\n}\n',
)
replace_once(
    'internal/worker/payment_reconciliation.go',
    '\t\t\treport, err := reporter.ReconciliationReport(runCtx, olderThanMinutes)',
    '\t\t\tif recoverer, ok := reporter.(MockProviderOutcomeRecoverer); ok {\n\t\t\t\trecovered, recoverErr := recoverer.ReconcileKnownMockProviderOutcomes(runCtx, olderThanMinutes, 100)\n\t\t\t\tif recoverErr != nil {\n\t\t\t\t\tlogger.Printf("ERROR mock-provider outcome recovery failed: %v", recoverErr)\n\t\t\t\t} else if recovered > 0 {\n\t\t\t\t\tlogger.Printf("mock-provider reconciliation recovered_count=%d older_than_minutes=%d", recovered, olderThanMinutes)\n\t\t\t\t}\n\t\t\t}\n\t\t\treport, err := reporter.ReconciliationReport(runCtx, olderThanMinutes)',
)

# Disable Redis by default on the measured workload; it can still be explicitly enabled.
replace_once('internal/config/config.go', 'getEnv("REDIS_ENABLED", "true")', 'getEnv("REDIS_ENABLED", "false")')
example = ROOT / '.env.example'
if example.exists():
    text = example.read_text()
    if 'REDIS_ENABLED=' in text:
        text2, count = re.subn(r'^REDIS_ENABLED=.*$', 'REDIS_ENABLED=false', text, count=1, flags=re.M)
        if count != 1:
            raise SystemExit('ERROR: could not update REDIS_ENABLED in .env.example')
        example.write_text(text2)
    else:
        with example.open('a') as f:
            f.write('\n# Optional product cache; disabled by default because the current local benchmark was slower.\nREDIS_ENABLED=false\nREDIS_ADDR=localhost:6379\nREDIS_DB=0\nPRODUCT_CACHE_TTL_SECONDS=60\n')

# Persist report evidence, not generated report drafts, in the source repository.
gitignore = ROOT / '.gitignore'
if gitignore.exists():
    text = gitignore.read_text()
    marker = '# Generated submission report drafts (upload the final PDF separately)'
    if marker not in text:
        gitignore.write_text(text.rstrip() + '\n\n' + marker + '\nsubmission/*DRAFT.*\n')

# Update OpenAPI paths before components so the published contract includes new routes.
openapi = ROOT / 'docs/openapi.yaml'
if not openapi.exists():
    raise SystemExit('ERROR: docs/openapi.yaml is missing')
text = openapi.read_text()
if '/admin/mock-provider/payments/{id}/outcome:' not in text:
    paths = '''  /metrics:
    get:
      tags: [Operations]
      summary: Request counter and latency histogram metrics
      security: []
      responses:
        '200':
          description: Prometheus text exposition format
          content:
            text/plain:
              schema: {type: string}
  /admin/mock-provider/payments/{id}/outcome:
    post:
      tags: [Administration]
      summary: Set a simulated provider outcome without delivering a callback (development only)
      description: Requires ADMIN JWT and is disabled when APP_ENV=production. This is a local test hook, not a real gateway API.
      security: [{BearerAuth: []}]
      parameters:
        - in: path
          name: id
          required: true
          schema: {$ref: '#/components/schemas/UUID'}
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              required: [outcome, amount_paise, currency]
              properties:
                outcome: {type: string, enum: [SUCCESS, FAILED]}
                amount_paise: {type: integer, format: int64, minimum: 0}
                currency: {type: string, minLength: 3, maxLength: 3, example: INR}
                failure_reason: {type: string, minLength: 3}
      responses:
        '200':
          description: Simulated provider-side outcome stored; OrderFlow payment remains pending until reconciliation
          content:
            application/json:
              schema: {$ref: '#/components/schemas/MockProviderOutcome'}
        '400': {$ref: '#/components/responses/BadRequest'}
        '401': {$ref: '#/components/responses/Unauthorized'}
        '403': {$ref: '#/components/responses/Forbidden'}
        '404': {$ref: '#/components/responses/NotFound'}
        '409': {$ref: '#/components/responses/Conflict'}
  /admin/payments/reconciliation/run:
    post:
      tags: [Administration]
      summary: Recover known terminal simulated-provider outcomes
      description: Uses the same validated payment transition path as a signed callback; not a live gateway integration.
      security: [{BearerAuth: []}]
      parameters:
        - in: query
          name: older_than_minutes
          schema: {type: integer, minimum: 0, maximum: 10080, default: 0}
        - in: query
          name: limit
          schema: {type: integer, minimum: 1, maximum: 1000, default: 100}
      responses:
        '200':
          description: Number of provider outcomes applied to still-pending OrderFlow payments
          content:
            application/json:
              schema:
                type: object
                required: [recovered_count, older_than_minutes, limit]
                properties:
                  recovered_count: {type: integer, minimum: 0}
                  older_than_minutes: {type: integer}
                  limit: {type: integer}
        '400': {$ref: '#/components/responses/BadRequest'}
        '401': {$ref: '#/components/responses/Unauthorized'}
        '403': {$ref: '#/components/responses/Forbidden'}
'''
    anchor = 'components:\n'
    if text.count(anchor) != 1:
        raise SystemExit('ERROR: expected exactly one components anchor in docs/openapi.yaml')
    text = text.replace(anchor, paths + anchor, 1)
    text = text.replace('    UUID:\n', '''    MockProviderOutcome:
      type: object
      required: [payment_id, provider_payment_id, outcome, amount_paise, currency, updated_at]
      properties:
        payment_id: {$ref: '#/components/schemas/UUID'}
        provider_payment_id: {type: string}
        outcome: {type: string, enum: [PENDING, SUCCESS, FAILED]}
        amount_paise: {type: integer, format: int64, minimum: 0}
        currency: {type: string, minLength: 3, maxLength: 3}
        failure_reason: {type: string}
        updated_at: {type: string, format: date-time}
    UUID:
''', 1)
    openapi.write_text(text)

# Add a Postman scenario that demonstrates a lost callback recovered from the mock ledger.
collection_path = ROOT / 'docs/postman/OrderFlow.postman_collection.json'
if not collection_path.exists():
    raise SystemExit('ERROR: docs/postman/OrderFlow.postman_collection.json is missing')
collection = json.loads(collection_path.read_text())
vars_ = collection.setdefault('variable', [])
existing_vars = {v.get('key') for v in vars_}
for key, value in [
    ('recoveryOrderId', ''), ('recoveryPaymentId', ''),
    ('recoveryPaymentAmountPaise', '0'), ('recoveryPaymentCurrency', 'INR'),
]:
    if key not in existing_vars:
        vars_.append({'key': key, 'value': value, 'type': 'string'})

def req_item(name, method, url, headers=None, body=None, tests=None, description=''):
    request = {'method': method, 'header': headers or [], 'url': url, 'description': description}
    if body is not None:
        request['body'] = {'mode': 'raw', 'raw': body, 'options': {'raw': {'language': 'json'}}}
    item = {'name': name, 'request': request}
    if tests:
        item['event'] = [{'listen': 'test', 'script': {'type': 'text/javascript', 'exec': tests}}]
    return item

folder_name = 'Missed callback reconciliation demo'
if not any(x.get('name') == folder_name for x in collection.get('item', [])):
    auth_customer = [{'key': 'Authorization', 'value': 'Bearer {{customerToken}}'}, {'key': 'Content-Type', 'value': 'application/json'}]
    auth_admin = [{'key': 'Authorization', 'value': 'Bearer {{adminToken}}'}, {'key': 'Content-Type', 'value': 'application/json'}]
    folder = {
      'name': folder_name,
      'description': 'Run after login as a customer and admin, and after a product with stock exists. This deliberately simulates a provider outcome without sending the webhook; the recovery endpoint applies the same transition path.',
      'item': [
        req_item('1 - Create recovery order', 'POST', '{{baseUrl}}/orders', auth_customer + [{'key':'Idempotency-Key','value':'{{$guid}}'}],
          '{\n  "items": [{"product_id": "{{productId}}", "quantity": 1}]\n}',
          ['pm.test("order created", () => pm.response.to.have.status(201));', 'const body = pm.response.json();', 'pm.collectionVariables.set("recoveryOrderId", body.order_id);', 'pm.collectionVariables.set("recoveryPaymentAmountPaise", String(body.total_amount_paise));'],
          'Creates a fresh one-product order and stores its ID and exact server-calculated total.'),
        req_item('2 - Initiate recovery payment', 'POST', '{{baseUrl}}/orders/{{recoveryOrderId}}/payments', auth_customer + [{'key':'Idempotency-Key','value':'{{$guid}}'}], None,
          ['pm.test("payment attempt created", () => pm.response.to.have.status(201));', 'const body = pm.response.json();', 'pm.collectionVariables.set("recoveryPaymentId", body.id);', 'pm.collectionVariables.set("recoveryPaymentAmountPaise", String(body.amount_paise));', 'pm.collectionVariables.set("recoveryPaymentCurrency", body.currency || "INR");'],
          'Creates a pending mock payment; do not send its regular callback in this recovery scenario.'),
        req_item('3 - Set provider outcome without callback', 'POST', '{{baseUrl}}/admin/mock-provider/payments/{{recoveryPaymentId}}/outcome', auth_admin,
          '{\n  "outcome": "SUCCESS",\n  "amount_paise": {{recoveryPaymentAmountPaise}},\n  "currency": "{{recoveryPaymentCurrency}}"\n}',
          ['pm.test("provider outcome stored", () => pm.response.to.have.status(200));', 'pm.test("this only changes the simulated provider ledger", () => pm.expect(pm.response.json().outcome).to.eql("SUCCESS"));'],
          'Development-only ADMIN route; it updates simulated provider state, not OrderFlow payment state.'),
        req_item('4 - Run reconciliation immediately', 'POST', '{{baseUrl}}/admin/payments/reconciliation/run?older_than_minutes=0&limit=100', auth_admin, None,
          ['pm.test("reconciliation succeeded", () => pm.response.to.have.status(200));', 'pm.test("one payment recovered", () => pm.expect(pm.response.json().recovered_count).to.be.at.least(1));'],
          'Uses the same processing path as a signed callback to confirm the order and consume reserved stock.'),
        req_item('5 - Verify recovered order', 'GET', '{{baseUrl}}/orders/{{recoveryOrderId}}', [{'key':'Authorization','value':'Bearer {{customerToken}}'}], None,
          ['pm.test("recovered order is confirmed", () => { pm.response.to.have.status(200); pm.expect(pm.response.json().status).to.eql("CONFIRMED"); });'],
          'Confirms that the missed successful callback was recovered.'),
      ]
    }
    collection['item'].append(folder)
collection_path.write_text(json.dumps(collection, indent=2, ensure_ascii=False) + '\n')

# Keep the Markdown project report synchronized with the current code and actual benchmark.
project_report = ROOT / 'docs/PROJECT_REPORT.md'
if project_report.exists():
    text = project_report.read_text()
    substitutions = [
        ('**Report scope:** implementation present at remote-review commit `8ffd04e`, plus documentation and CI artifacts proposed in this bundle.', '**Report scope:** OrderFlow implementation in the current repository, including reliability migration `000007`, multi-product/idempotent order creation, inventory/payment transitions, Redis catalog cache, admin reports, Docker and CI. Replace the final commit SHA with the exact submitted SHA after the last verification run.'),
        ('The current integration tests exercise core order/payment flows, rollback behavior and concurrent requests against a dedicated PostgreSQL test database. Live Stripe processing, Redis caching and cloud deployment are future work and are not represented as completed features in this report.', 'The integration tests exercise order/payment workflows, rollback behavior, concurrent requests, Redis TTL/invalidation, and periodic worker behavior against disposable dependencies. The provider is simulated, not Stripe. AWS resources are not represented as deployed unless provisioned and verified separately.'),
        ('The prior integration milestone was green at commit `8ffd04e`. This bundle adds Redis and worker tests; the updated CI must pass before reporting these changes as verified. The CI workflow is configured to provide disposable PostgreSQL and Redis services, perform migrations, run tests and race checks, build the image, and probe health/readiness inside the running container.', 'Commit `d891056` has a successful GitHub Actions run including the migration chain, unit/database/Redis tests, integration race tests, vet, Docker build and runtime health/readiness probes. Any additional final-review patch must receive a new green run before it is described as verified.'),
        ('1. **Redis:** optional product list/detail cache with TTL and versioned invalidation is included in this milestone. Treat it as verified only after updated CI reports green.', '1. **Redis:** product list/detail cache with TTL and generation-based invalidation is implemented. The latest measured small local API workload was slower with the warm Redis cache, so Redis is disabled by default pending a representative benchmark that demonstrates benefit.'),
        ('OrderFlow implements the central backend path from product and inventory through transactional order creation and mock payment state changes, with a layered Go design and PostgreSQL-backed correctness controls. The next steps are to validate the added OpenAPI/CI/container/documentation files, capture submission evidence, and decide from the rubric whether Redis, live payment processing, scheduled reconciliation and AWS deployment must be completed before submission.', 'OrderFlow implements the core product, inventory, multi-product order, idempotency, mock-payment, cancellation, expiry and reporting workflows with PostgreSQL-backed correctness controls. The local provider-recovery demo and HTTP observability additions must pass the final local test suite and a CI run on the submitted SHA. Live Stripe is outside the initial simulated-provider scope; AWS deployment depends on the evaluation rubric and actual resource approval.'),
        ('`BenchmarkProductListDatabaseVsRedis` compares uncached PostgreSQL product listing to a warmed Redis hit. The final results section must include the actual `go test -bench` output and hardware/database details; no numerical improvement is claimed before a repeatable benchmark is run. AWS architecture is documented but a real deployment must be separately approved, performed and verified.', '`BenchmarkProductListDatabaseVsRedis` compares PostgreSQL reads to warm Redis hits. In the supplied local API run (300 requests per variant, 10 concurrent workers, Apple M2, Go 1.27.2, Docker 29.8.2), cache disabled produced P50 2.480 ms, P95 4.270 ms, mean 3.315 ms and 2821.81 requests/s; warm Redis produced P50 4.447 ms, P95 13.966 ms, mean 5.702 ms and 1697.92 requests/s. Redis was slower for this workload; no improvement is claimed. The AWS diagram is a target design, not evidence of live deployment.'),
    ]
    for old, new in substitutions:
        if old in text:
            text = text.replace(old, new, 1)
    report_recon_old = 'The reconciliation report currently detects stale pending payments, confirmed orders lacking successful payments and failed payments whose orders remain pending. The new worker periodically emits issues to logs but does not automatically query a live gateway or mutate payment state. Therefore it is a read-only monitor,'
    report_recon_new = 'The reconciliation report detects stale pending payments, confirmed orders lacking successful payments and failed payments whose orders remain pending. The mock-provider reconciliation worker can also look up a known terminal result in the separate simulated-provider ledger and pass it through the same callback processing path; it does not query a live gateway.'
    if report_recon_old in text:
        text = text.replace(report_recon_old, report_recon_new, 1)
    project_report.write_text(text)

# Update the cache/reconciliation guide to avoid describing recovery as report-only.
redis_doc = ROOT / 'docs/REDIS_AND_RECONCILIATION.md'
if redis_doc.exists():
    text = redis_doc.read_text()
    text = text.replace(
        '`internal/worker/payment_reconciliation.go` periodically calls the already-existing `PaymentService.ReconciliationReport` and logs each stale/inconsistent record. `RECONCILIATION_INTERVAL_SECONDS` defaults to 300; setting it to 0 disables the worker. `RECONCILIATION_OLDER_THAN_MINUTES` defaults to 15.\n\nThe worker is intentionally **read-only**. It does not decide whether a gateway payment succeeded, retry charges, refund customers, or mutate status automatically. A production integration should rely on authoritative provider APIs/webhooks and an explicit remediation workflow.',
        '`internal/worker/payment_reconciliation.go` periodically checks the simulated-provider ledger for known terminal outcomes on still-pending payments, applies them through the same callback transition path, then logs remaining stale/inconsistent records from `PaymentService.ReconciliationReport`. `RECONCILIATION_INTERVAL_SECONDS` defaults to 300; setting it to 0 disables the worker. `RECONCILIATION_OLDER_THAN_MINUTES` defaults to 15.\n\nThe recovery logic is for the simulated provider only. It does not query a live gateway, initiate charges, or refund money. Production provider reconciliation must rely on the authoritative provider API/webhooks and an explicit remediation workflow. The development-only ADMIN endpoint can simulate a provider result without delivering the callback; it returns 404 when `APP_ENV=production`.'
    )
    redis_doc.write_text(text)

# Make the final screenshot guide reflect the latest code and honest benchmark result.
evidence_readme = ROOT / 'docs/evidence/README.md'
if evidence_readme.exists():
    text = evidence_readme.read_text()
    text = text.replace('`docs/evidence/performance/01-baseline.png` and `02-cached.png`', '`docs/evidence/performance/01-baseline.png` and `02-redis-result.png` (the Redis result may be slower; report the actual result)')
    text = text.replace('`submission/OrderFlow_Applied_Software_Project_Report_DRAFT.docx`', 'the local final Word report copy (the PDF is submitted separately)')
    text = text.replace('IMPORTANT: current code retains reservation on SUCCESS, while the original capstone spec says successful reserved stock becomes consumed; fix and test this mismatch before claiming stock consumption in the report.', 'The latest reliability code should atomically decrement on-hand and reserved quantities and write one ORDER_FULFILLED movement; verify this in the final run before capturing evidence.')
    if 'Final reliability additions' not in text:
        text += '''
## Final reliability additions

- [ ] Multi-product order creates one order with all item snapshots and reserves every line atomically.
- [ ] Same order idempotency key replays the original result; changed payload returns 409.
- [ ] Mock provider outcome is set without delivering a callback, then `POST /admin/payments/reconciliation/run?older_than_minutes=0` recovers it through the callback transition path.
- [ ] `X-Request-ID` appears in HTTP responses and `GET /metrics` returns request counters/histogram.
- [ ] Benchmark outputs for product API, order creation and order-history listing are saved with hardware/workload details; do not omit negative cache results.
'''
    evidence_readme.write_text(text)

# Bring older report/docs wording in line with the actual committed implementation and measured result.
readme = ROOT / 'README.md'
if readme.exists():
    text = readme.read_text()
    milestone_old = '> **Implementation milestone:** Redis product-read caching, TTL-based expiration, versioned invalidation, a periodic reconciliation reporter, and a Docker runtime smoke test are included in this change bundle. They must pass the repository workflow after the bundle is committed; do not treat a local ZIP as proof of a green CI run.'
    milestone_new = '> **Verified baseline:** commit `d891056` passed GitHub Actions, including migrations, tests, race checks, vet, Docker build and runtime health/readiness. The final-review overlay adds mock-provider recovery and HTTP observability; it must pass a new local test run and CI before submission.'
    if milestone_old in text:
        text = text.replace(milestone_old, milestone_new, 1)
    text = text.replace(
        '- Configurable read-only reconciliation worker that periodically logs stale pending payments or inconsistent order/payment states.',
        '- Scheduled mock-provider reconciliation that recovers known terminal outcomes through the shared callback path and logs remaining inconsistent states; real gateway recovery is out of scope.'
    )
    text = text.replace(
        '| `internal/worker` | Periodic read-only payment reconciliation reporting |',
        '| `internal/worker` | Scheduled provider-outcome recovery, unpaid-order expiry and reconciliation reporting |'
    )
    text = text.replace(
        '| `internal/cache` | Redis adapter for product read caching |',
        '| `internal/cache` | Redis adapter for product read caching |\n| `internal/observability` | Request IDs, request counters and HTTP latency histogram at `/metrics` |'
    )
    old = 'The scheduled reconciliation worker is read-only and reports issues to application logs; it does not automatically settle payments.'
    new = 'The scheduled worker recovers known terminal outcomes from the separate simulated-provider ledger through the same validated callback transition path, then reports remaining stale/inconsistent states to application logs. It does not integrate with a live gateway.'
    if old in text:
        text = text.replace(old, new, 1)
    section = '''\n## Final verification and submission\n\n- [Mock-provider recovery and HTTP observability](docs/MOCK_PROVIDER_RECOVERY_AND_OBSERVABILITY.md)\n- [Submission portal checklist](docs/FINAL_SUBMISSION_GUIDE.md)\n- [Evidence and screenshot checklist](docs/evidence/README.md)\n- HLD and LLD source diagrams are maintained in `docs/diagrams/`; the rendered diagrams are also included below in this README.\n\n`GET /metrics` exposes request counters and a duration histogram. `X-Request-ID` is returned on every HTTP response. Redis remains optional and defaults off because the latest local product-list benchmark was slower with the warm cache; do not claim a cache speedup without a new representative benchmark.\n'''
    current_diagrams = '''\n## Final HLD/LLD after reliability update\n\nThe editable diagram sources are in `docs/diagrams/HLD.mmd`, `docs/diagrams/LLD.mmd`, and `docs/diagrams/ORDER_PAYMENT_STATE.mmd`. These diagrams include the simulated-provider ledger, request metrics, and the callback-recovery path.\n\n### Current HLD\n```mermaid\nflowchart LR\n    Client[Customer / ADMIN / Postman] --> API[Gin REST API]\n    API --> Request[Request ID, DTO validation, JWT/roles]\n    Request --> Services[Product / Order / Inventory / Payment services]\n    Services --> PostgreSQL[(PostgreSQL source of truth)]\n    Services -. optional catalog cache .-> Redis[(Redis, TTL + invalidation)]\n    PostgreSQL --> ProviderLedger[(Mock provider outcome ledger)]\n    Worker[Expiry + provider reconciliation workers] --> Services\n    API --> Metrics[/metrics: request counts + duration histogram/]\n    CI[GitHub Actions] -. tests and container smoke .-> API\n```\n\n### Current LLD: missed callback recovery\n```mermaid\nsequenceDiagram\n    participant P as Mock provider ledger\n    participant W as Reconciliation worker\n    participant S as Payment service\n    participant R as Payment repository\n    participant DB as PostgreSQL\n    P-->>W: Known SUCCESS/FAILED outcome\n    W->>S: ReconcileKnownMockProviderOutcomes\n    S->>S: Reuse ProcessMockCallback validation/state path\n    S->>R: Lock and apply payment/order/stock transition\n    R->>DB: Commit status + movement audit atomically\n    S-->>W: Recovered count / error\n```\n'''
    if '## Final HLD/LLD after reliability update' not in text:
        text = text.rstrip() + '\n' + current_diagrams
    if '## Final verification and submission' not in text:
        text = text.rstrip() + '\n' + section
    readme.write_text(text)

# Ensure migration 000008 is applied by the normal residual/benchmark preflights too.
for script_path in ('scripts/verify-residual-milestone.sh', 'scripts/benchmark-api-cache.sh'):
    p = ROOT / script_path
    if p.exists():
        s = p.read_text()
        anchor = './scripts/apply-reliability-migration.sh\n'
        if './scripts/apply-mock-provider-migration.sh\n' not in s:
            if s.count(anchor) != 1:
                raise SystemExit(f'ERROR: expected reliability migration anchor in {script_path}')
            p.write_text(s.replace(anchor, anchor + './scripts/apply-mock-provider-migration.sh\n', 1))

print('Final review edits applied to router, worker, config, OpenAPI, Postman and README.')
