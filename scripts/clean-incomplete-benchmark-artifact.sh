#!/usr/bin/env bash
set -euo pipefail
partial='docs/evidence/performance/order-history-benchmark-20261018T201742Z.txt'
if [[ -f "$partial" ]]; then
  rm "$partial"
  echo "Removed incomplete benchmark output: $partial"
else
  echo "No known incomplete benchmark output found; nothing removed."
fi
# Remove trailing horizontal whitespace from the documentation lines flagged by git diff --check.
for file in README.md docs/FINALIZATION_STATUS.md docs/PROJECT_REPORT.md docs/REMOTE_REVIEW_AND_REMAINING_WORK.md; do
  if [[ -f "$file" ]]; then
    perl -pi -e 's/[ \t]+$//' "$file"
  fi
done
echo 'Cleanup complete. Run git diff --check and inspect git status before committing.'
