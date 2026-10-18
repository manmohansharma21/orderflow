#!/usr/bin/env python3
from pathlib import Path

root_readme = Path("README.md")
if root_readme.exists():
    s = root_readme.read_text(encoding="utf-8")
    s = s.replace("`docs/postman/OrderFlow.postman_collection.json`", "`docs/postman/OrderFlow_Guided_Evidence.postman_collection.json`")
    s = s.replace("import `docs/postman/OrderFlow.postman_collection.json` into Postman", "import `docs/postman/OrderFlow_Guided_Evidence.postman_collection.json` into Postman")
    s = "\n".join(line.rstrip() for line in s.splitlines()) + "\n"
    root_readme.write_text(s, encoding="utf-8")

p = Path("docs/evidence/README.md")
if p.exists():
    s = p.read_text(encoding="utf-8")
    heading = "## Initial Postman captures and authorization diagnosis"
    supplement = "\n\n## Initial Postman captures and authorization diagnosis\n\nUser-supplied initial Postman screenshots are preserved under `docs/evidence/postman/initial/`. The login screenshot is redacted; the original displayed an active JWT and is intentionally not included. The 401/403 product-create screenshots are troubleshooting history, not successful CRUD evidence. See `docs/postman/POSTMAN_ADMIN_TOKEN_TROUBLESHOOTING.md`.\n\nImportant distinction: `GET /products` is public and expected to return `200`. `POST/PATCH/DELETE /products` require a valid ADMIN token. An empty `adminToken` produces `401 bearer token required`; a valid CUSTOMER token produces `403 insufficient permissions`. Run the collection's existing-ADMIN login request and verify `user.role == ADMIN` before admin operations.\n"
    if heading not in s:
        s = s.rstrip() + supplement + "\n"
    s = "\n".join(line.rstrip() for line in s.splitlines()) + "\n"
    p.write_text(s, encoding="utf-8")
print("README/evidence links updated")
