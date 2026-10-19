#!/usr/bin/env python3
"""Append the final Postman screenshot section to docs/evidence/README.md idempotently."""
from pathlib import Path

p = Path("docs/evidence/README.md")
marker = "## Final Postman captures supplied on 19 October 2026"
if not p.exists():
    raise SystemExit("docs/evidence/README.md not found; run from repository root")
text = p.read_text()
if marker in text:
    print("Final Postman capture section already present")
    raise SystemExit(0)
addition = """
## Final Postman captures supplied on 19 October 2026

| Capture | File | Meaning |
|---|---|---|
| ADMIN login, token redacted | `postman/final-captures/01-admin-login-200-token-redacted.png` | Login returned 200 and role was ADMIN; JWT value is redacted. |
| Product create | `postman/final-captures/02-product-create-201.png` | Admin product create returned 201 Created. |
| Inventory availability | `postman/final-captures/03-inventory-availability-200.png` | Availability returned 200; quantities reflect the local test run. |
| Inventory movement audit | `postman/final-captures/04-inventory-movement-audit-200.png` | Admin movement audit returned 200 and includes opening-balance history. |

The next final captures should be one multi-product order (201), order history (200), and one successful payment/missed-callback recovery result (order CONFIRMED). See `../postman/FINAL_LOCAL_POSTMAN_STEPS.md`. Do not capture every endpoint.
"""
p.write_text(text.rstrip() + "\n\n" + addition.strip() + "\n")
print("Updated docs/evidence/README.md while preserving existing sections")
