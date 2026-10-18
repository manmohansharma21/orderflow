#!/usr/bin/env python3
"""Rename the demo collection to a neutral, submission-appropriate name."""
import json
from pathlib import Path

repo = Path.cwd()
old = repo / "docs/postman/OrderFlow_Guided_Evidence.postman_collection.json"
new = repo / "docs/postman/OrderFlow_API_Verification.postman_collection.json"
if new.exists():
    target = new
elif old.exists():
    target = old
else:
    raise SystemExit("Run from repo root; guided Postman collection JSON not found")

obj = json.loads(target.read_text(encoding="utf-8"))
obj.setdefault("info", {})["name"] = "OrderFlow API Verification — Local"
new.write_text(json.dumps(obj, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
if old.exists() and old != new:
    old.unlink()

replacements = {
    "OrderFlow_Guided_Evidence.postman_collection.json": "OrderFlow_API_Verification.postman_collection.json",
    "OrderFlow — Guided Local API Evidence Run (Admin Token Fix)": "OrderFlow API Verification — Local",
    "OrderFlow — Guided Local API Evidence Run": "OrderFlow API Verification — Local",
}
for base in (repo / "README.md", repo / "docs", repo / "POSTMAN_IMPORT_README.txt", repo / "POSTMAN_AUTH_FIX_IMPORT_README.txt"):
    if not base.exists():
        continue
    files = [base] if base.is_file() else list(base.rglob("*.md")) + list(base.rglob("*.txt"))
    for path in files:
        if path.name == Path(__file__).name or "__pycache__" in path.parts or path.suffix.lower() not in {".md", ".txt"}:
            continue
        try:
            text = path.read_text(encoding="utf-8")
        except UnicodeDecodeError:
            continue
        updated = text
        for before, after in replacements.items():
            updated = updated.replace(before, after)
        if updated != text:
            path.write_text(updated.rstrip() + "\n", encoding="utf-8")
print("Postman collection renamed to 'OrderFlow API Verification — Local'; source file references updated.")
