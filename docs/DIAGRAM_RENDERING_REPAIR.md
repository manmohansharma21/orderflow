# Diagram rendering repair

The previous inline Mermaid blocks could render differently across browser previews. The root README now embeds committed PNGs (with source files retained under `docs/diagrams/`), so the displayed architecture is no longer dependent on a client-side Mermaid parser.

The HLD labels a single deployable Go application (modular monolith). The LLD shows order-level idempotency, transaction boundaries, conditional stock reservation and rollback. The ERD shows the mock-provider transaction ledger introduced by migration 000008. The AWS image is explicitly marked as a target architecture, not a live deployment.

After extraction, inspect `docs/diagrams/*.png` locally and in the GitHub README preview. If the image link is broken, verify case-sensitive repository paths and ensure the PNG is committed. Do not replace the PNG embed with raw Mermaid until the target renderer is confirmed to support it.

The root README does not use inline Mermaid code blocks. HLD, LLD, ERD, state, concurrency, payment recovery, CI and AWS target drawings are linked as repository-relative PNGs, with editable DOT and selected Mermaid sources retained in `docs/diagrams/`.
