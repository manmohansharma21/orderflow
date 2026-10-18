# OrderFlow architecture diagram sources and browser-safe images

Use the PNGs in the README and Word report because they render consistently in GitHub and document viewers. Editable source files are retained as Graphviz DOT; HLD, LLD and state also have Mermaid `.mmd` sources.

| Diagram | Image | Editable source |
|---|---|---|
| HLD — modular monolith | `HLD.png` | `HLD.dot`, `HLD.mmd` |
| LLD — order creation/idempotency | `LLD.png` | `LLD.dot`, `LLD.mmd` |
| Domain class diagram | `CLASS_DIAGRAM.png` | `CLASS_DIAGRAM.dot` |
| ERD/schema | `ERD.png` | `ERD.dot` |
| Order/payment state | `ORDER_PAYMENT_STATE.png` | `ORDER_PAYMENT_STATE.dot`, `.mmd` |
| Multi-item order request flow | `ORDER_REQUEST_FLOW.png` | `ORDER_REQUEST_FLOW.dot` |
| Concurrent inventory reservation | `ORDER_CONCURRENCY.png` | `ORDER_CONCURRENCY.dot` |
| Mock-provider recovery | `PAYMENT_RECOVERY.png` | `PAYMENT_RECOVERY.dot` |
| GitHub Actions CI | `CI_PIPELINE.png` | `CI_PIPELINE.dot` |
| AWS target design (not deployed) | `AWS_TARGET.png` | `AWS_TARGET.dot` |
| Project lifecycle | `PROJECT_LIFECYCLE.png` | `PROJECT_LIFECYCLE.dot` |

To regenerate the PNG/SVG from DOT with Graphviz installed, run `scripts/render-architecture-diagrams.sh`. The HLD labels one deployable Go application and makes clear that internal packages/workers are not microservices. PostgreSQL is the durable source of truth; Redis is optional and only used for product reads.
