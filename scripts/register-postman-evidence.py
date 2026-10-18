#!/usr/bin/env python3
"""Register the new public-product screenshot in the evidence index."""
from pathlib import Path

path = Path("docs/evidence/README.md")
if not path.exists():
    raise SystemExit(f"Missing {path}; run from repository root")
text = path.read_text(encoding="utf-8")
marker = "postman/initial/08-product-list-public-200.png"
if marker in text:
    print("Product-list evidence is already registered")
    raise SystemExit(0)
row = "| Public product listing (HTTP 200) | `postman/initial/08-product-list-public-200.png` | `GET /products?page=1&limit=20` returned 200 with one active product; endpoint is public. |\n"
needle = "## Remaining screenshots: take these after final local test and green CI"
if needle not in text:
    raise SystemExit(f"Could not find expected insertion point in {path}; no files changed")
text = text.replace(needle, row + "\n" + needle, 1)
path.write_text(text.rstrip() + "\n", encoding="utf-8")
print(f"Updated {path}")
