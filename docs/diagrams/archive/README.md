# Archived Mermaid diagram variants

These files preserve earlier Mermaid-only views from prior commits for learning and comparison. They are historical source variants, not the canonical diagrams used in the README/report.

Canonical rendered diagrams are the PNGs in `docs/diagrams/` (HLD.png, LLD.png, ERD.png, CLASS_DIAGRAM.png, ORDER_PAYMENT_STATE.png, ORDER_CONCURRENCY.png, PAYMENT_RECOVERY.png, CI_PIPELINE.png, AWS_TARGET.png and PROJECT_LIFECYCLE.png). Editable Graphviz DOT sources sit alongside them. The current Mermaid sources remain in `docs/diagrams/*.mmd`.

- `HLD_detailed_1515cb0.mmd`: includes mock-provider ledger, HMAC validation, cache, workers and metrics.
- `HLD_initial_60c255b.mmd`: early request/callback architecture.
- `LLD_multi_product_1515cb0.mmd`: multi-product order idempotency and transaction flow.
- `LLD_single_product_60c255b.mmd`: early single-product reservation path.
- `ORDER_PAYMENT_STATE_detailed_1515cb0.mmd`: current domain states including expiry.
- `ORDER_PAYMENT_STATE_initial_60c255b.mmd`: earlier lifecycle without expiry.
