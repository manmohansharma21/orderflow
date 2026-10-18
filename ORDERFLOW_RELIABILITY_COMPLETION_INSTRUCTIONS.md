# OrderFlow reliability and submission instructions — current revision

The reliability and mock-provider recovery source code is already pushed at `1515cb0`, and GitHub Actions for that exact commit completed successfully. Do not reapply `ORDERFLOW_FINAL_REVIEW_PATCH.zip`; that was an incremental overlay used to develop this revision.

## Verified locally in the supplied terminal log

```bash
./scripts/verify-residual-milestone.sh
./scripts/docker-smoke-test.sh
```

Tests/race integration checks/static checks passed; Docker image built and `/health` and `/ready` returned successful responses. The matching remote run is https://github.com/manmohansharma21/orderflow/actions/runs/39160539244.

## Outstanding finalization tasks

- Apply `ORDERFLOW_README_SCREENSHOT_GUIDE_PATCH.zip` and commit the docs cleanup.
- Complete the final Postman/API evidence session; capture genuine redacted PNGs into `docs/evidence/` and paste those same files into the Word report.
- The product-list API benchmark is complete and committed. It found Redis slower for the supplied local workload; retain the negative result. The order-history benchmark was interrupted before completion; use the bounded command in `docs/RELIABILITY_MILESTONE.md` if a completed value is required.
- Refresh the Word report, insert screenshots, verify references and dates, export the required PDF and inspect it.
- After the final push, confirm CI is green for the final SHA, then download the full source ZIP from GitHub `main` for the portal.

Never commit `.env`, credentials, DB backups or generated report draft binaries. AWS remains conditional on the official rubric and budget approval; the payment integration remains a mock provider.
