# Visual evidence and diagram repair patch

Prepared 19 October 2026 for remote `main` HEAD `3df0139`. This ZIP is an overlay for your existing repository, not a complete project ZIP.

From the repository root: `unzip -o ~/Downloads/ORDERFLOW_FINAL_VISUAL_EVIDENCE_PATCH.zip -d .` then `git status --short` and `git diff --check`. Review `README.md`, `docs/PROJECT_REPORT.md`, `docs/diagrams/`, and `docs/evidence/README.md`. Never stage `.env`, tokens, webhook secrets or database backups. The password-bearing raw VS Code screenshot was intentionally excluded; only a safe crop of the Docker Compose portion is included.

The README references pre-rendered diagram PNGs rather than inline Mermaid, while keeping editable DOT/Mermaid source. The screenshot guide separates actual captured evidence from remaining API/Postman evidence. It does not fabricate missing screenshots.

After extracting, run `scripts/render-architecture-diagrams.sh` only if Graphviz is installed and you want to regenerate images; the ZIP already includes the rendered PNGs and SVGs. Then run `go test ./...`, `go test -race ./internal/integration/...`, `go vet ./...`, and `git diff --check`. Stage only intended documentation/evidence paths, commit and push, then confirm a green Actions run for the new SHA.
