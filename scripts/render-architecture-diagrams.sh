#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")/.."
for src in docs/diagrams/*.dot; do
  name="$(basename "$src" .dot)"
  dot -Tpng -Gdpi=200 "$src" -o "docs/diagrams/${name}.png"
  dot -Tsvg "$src" -o "docs/diagrams/${name}.svg"
done
echo "Rendered architecture PNG/SVG assets from docs/diagrams/*.dot"
