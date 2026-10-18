#!/usr/bin/env bash
set -euo pipefail

# Run from the root of the existing OrderFlow repository after unzipping this overlay.
python3 scripts/patch-final-review-files.py

chmod +x scripts/apply-mock-provider-migration.sh scripts/verify-residual-milestone.sh scripts/benchmark-api-cache.sh
./scripts/apply-reliability-migration.sh
./scripts/apply-mock-provider-migration.sh

# The report is uploaded separately; remove stale generated drafts from source control.
if git ls-files --error-unmatch submission/OrderFlow_Applied_Software_Project_Report_DRAFT.docx >/dev/null 2>&1; then
  git rm -f submission/OrderFlow_Applied_Software_Project_Report_DRAFT.docx submission/OrderFlow_Applied_Software_Project_Report_DRAFT.pdf
fi

find cmd internal -name '*.go' -print0 | xargs -0 gofmt -w
echo 'Final review patch applied. Run the verification checklist before staging or committing.'
